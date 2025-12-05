package stake

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/dijiacoder/MetaNodeStakeSync/dao/model"
	"github.com/dijiacoder/MetaNodeStakeSync/dao/repository/contractevents"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/dijiacoder/MetaNodeStakeSync/app/service/common"
	"github.com/dijiacoder/MetaNodeStakeSync/app/service/config"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"gorm.io/gorm"
)

// TaskStake stake合约同步任务
type TaskStake struct {
	Context      context.Context
	Config       *config.Config
	DB           *gorm.DB
	RedisClient  *redis.Client
	ChainID      int32
	ContractName string
	ABIStr       string
	Address      string
	CreatedHash  *string
	Client       *ethclient.Client
	ABI          *abi.ABI
}

// NewTaskStake 创建stake合约同步任务
func NewTaskStake(serviceCtx *common.ServiceContext) *TaskStake {
	stakeContract := serviceCtx.ContractInfoMap[1]
	ABI, err := common.GetABI(stakeContract.ABIStr)
	if err != nil {
		logx.Error("Failed to parse ABI: ", err)
	}

	// 解析ABI
	for s := range ABI.Events {
		fmt.Printf("Event: %s, %s\n", s, ABI.Events[s].ID.Hex())
	}

	return &TaskStake{
		Context:      serviceCtx.Context,
		Config:       serviceCtx.Config,
		DB:           serviceCtx.DB,
		RedisClient:  serviceCtx.RedisClient,
		ChainID:      stakeContract.ChainID,
		ContractName: stakeContract.ContractName,
		ABIStr:       stakeContract.ABIStr,
		Address:      stakeContract.Address,
		CreatedHash:  stakeContract.CreatedHash,
		Client:       stakeContract.Client,
		ABI:          ABI,
	}
}

func (t *TaskStake) Start() {
	threading.GoSafe(t.process)
}

func (t *TaskStake) process() {
	for {
		select {
		case <-t.Context.Done():
			logx.Info("stake task stopped")
			return
		default:
			t.queryLogs()
			time.Sleep(1 * time.Second)
		}
	}
}

func (t *TaskStake) queryLogs() {
	startBlock := big.NewInt(0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lastHeigh, err := t.RedisClient.Get(ctx, common.GetKey(t.ChainID, t.Address)).Uint64()
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.Info(err)
	}

	if lastHeigh == 0 {
		receipt, err := t.Client.TransactionReceipt(ctx, ethCommon.HexToHash(*t.CreatedHash))
		if err != nil {
			logx.Info(err)
		}
		startBlock = big.NewInt(0).Add(receipt.BlockNumber, big.NewInt(1))
	} else {
		startBlock = big.NewInt(int64(lastHeigh + 1))
	}

	currentHeight, err := t.Client.BlockNumber(ctx)
	if err != nil {
		logx.Info(err)
		return
	}

	if big.NewInt(int64(currentHeight)).Cmp(startBlock) <= 0 {
		return
	}

	endBlock := big.NewInt(0).Add(startBlock, big.NewInt(int64(9)))
	if endBlock.Cmp(big.NewInt(int64(currentHeight))) > 0 {
		endBlock = big.NewInt(int64(currentHeight))
	}

	logx.Info(fmt.Sprintf("sync stake, start: %d, end: %d, current: %d", startBlock.Int64(), endBlock.Int64(), currentHeight))

	logs, err := t.Client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   endBlock,
		Addresses: []ethCommon.Address{ethCommon.HexToAddress(t.Address)},
	})

	if err != nil {
		logx.Info(err)
		return
	}

	for _, l := range logs {
		errTx := t.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 在事务内使用 tx，确保所有写入共享同一事务上下文
			originalDB := t.DB
			t.DB = tx
			defer func() { t.DB = originalDB }()

			// 判断交易是否已处理
			exists, err := t.HasProcessedTx(ctx, l.TxHash.Hex())
			if err != nil {
				return err
			}
			if exists {
				logx.Info(fmt.Sprintf("queryLogs: transaction already processed, TxHash=%s", l.TxHash.Hex()))
				// 已处理则直接结束事务（无写入，正常提交）
				return nil
			}

			// 保存统一事件记录（与后续事件处理共享同一事务）
			if err := t.SaveContractEvent(ctx, l); err != nil {
				return err
			}

			handlers := map[string]func(ethereumTypes.Log) error{
				t.ABI.Events["AddPool"].ID.Hex(): t.HandleAddPoolEvent,
				//t.ABI.Events["Deposit"].ID.Hex():        t.HandleDepositEvent,
				//t.ABI.Events["Claim"].ID.Hex():          t.HandleClaimEvent,
				//t.ABI.Events["RequestUnstake"].ID.Hex(): t.HandleRequestUnstakeEvent,
				//t.ABI.Events["Withdraw"].ID.Hex():       t.HandleWithdrawEvent,
			}

			eventID := l.Topics[0].Hex()
			if h, ok := handlers[eventID]; ok {
				if err := h(l); err != nil {
					// 出错直接返回错误，触发事务回滚
					return fmt.Errorf("failed to handle event: %s, Block: %d, TxHash: %s, Err: %v", eventID, l.BlockNumber, l.TxHash.Hex(), err)
				}
			} else {
				logx.Info(fmt.Sprintf("Unknown event ID: %s, Block: %d, TxHash: %s", eventID, l.BlockNumber, l.TxHash.Hex()))
			}

			// 正常提交事务
			return nil
		})

		if errTx != nil {
			logx.Error("queryLogs: transaction rollback due to error: ", errTx)
			continue
		}
	}

	err = t.RedisClient.Set(ctx, common.GetKey(t.ChainID, t.Address), endBlock.Uint64(), 0).Err()
	if err != nil {
		logx.Info(err)
		return
	}
}

func (t *TaskStake) HasProcessedTx(ctx context.Context, txHash string) (bool, error) {
	return contractevents.ExistsByTxHash(ctx, t.DB, txHash)
}

func (t *TaskStake) SaveContractEvent(ctx context.Context, l ethereumTypes.Log) error {
	// 解析事件名称
	eventID := l.Topics[0].Hex()
	eventName := ""
	for name, ev := range t.ABI.Events {
		if ev.ID.Hex() == eventID {
			eventName = name
			break
		}
	}

	// 获取区块时间戳
	block, err := t.Client.BlockByNumber(ctx, big.NewInt(int64(l.BlockNumber)))
	if err != nil {
		return err
	}

	// 主题与数据
	var topic1, topic2, topic3 *string
	if len(l.Topics) > 1 {
		s := l.Topics[1].Hex()
		topic1 = &s
	}
	if len(l.Topics) > 2 {
		s := l.Topics[2].Hex()
		topic2 = &s
	}
	if len(l.Topics) > 3 {
		s := l.Topics[3].Hex()
		topic3 = &s
	}
	dataHex := fmt.Sprintf("0x%x", l.Data)

	// 构建并保存
	ev := &model.ContractEvent{
		ContractAddress: t.Address,
		EventName:       eventName,
		Topic0:          eventID,
		Topic1:          topic1,
		Topic2:          topic2,
		Topic3:          topic3,
		Data:            &dataHex,
		BlockNumber:     l.BlockNumber,
		BlockTimestamp:  block.Time(),
		TransactionHash: l.TxHash.Hex(),
		LogIndex:        int32(l.Index),
	}
	return contractevents.Create(ctx, t.DB, ev)
}
