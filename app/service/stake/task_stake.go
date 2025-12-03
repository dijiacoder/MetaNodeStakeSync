package stake

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/dijiacoder/MetaNodeStakeSync/dao/model"
	"github.com/dijiacoder/MetaNodeStakeSync/dao/repository/eventaddpool"
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
		handlers := map[string]func(ethereumTypes.Log) bool{
			t.ABI.Events["AddPool"].ID.Hex():        t.HandleAddPoolEvent,
			t.ABI.Events["Deposit"].ID.Hex():        t.HandleDepositEvent,
			t.ABI.Events["Claim"].ID.Hex():          t.HandleClaimEvent,
			t.ABI.Events["RequestUnstake"].ID.Hex(): t.HandleRequestUnstakeEvent,
			t.ABI.Events["Withdraw"].ID.Hex():       t.HandleWithdrawEvent,
		}

		eventID := l.Topics[0].Hex()
		if h, ok := handlers[eventID]; ok {
			if !h(l) {
				logx.Error(fmt.Sprintf("Failed to handle event: %s, Block: %d, TxHash: %s", eventID, l.BlockNumber, l.TxHash.Hex()))
				return
			}
		} else {
			logx.Info(fmt.Sprintf("Unknown event ID: %s, Block: %d, TxHash: %s", eventID, l.BlockNumber, l.TxHash.Hex()))
		}
	}

	err = t.RedisClient.Set(ctx, common.GetKey(t.ChainID, t.Address), endBlock.Uint64(), 0).Err()
	if err != nil {
		logx.Info(err)
		return
	}
}

func (t *TaskStake) HandleAddPoolEvent(l ethereumTypes.Log) bool {
	logx.Info(fmt.Sprintf("HandleAddPoolEvent: BlockNumber=%d, TxHash=%s, Address=%s, Topics=%v, Data=%x",
		l.BlockNumber, l.TxHash.Hex(), l.Address.Hex(), l.Topics, l.Data))
	if len(l.Topics) < 4 {
		logx.Error("HandleAddPoolEvent: invalid topics length")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := eventaddpool.ExistsByTxHash(ctx, t.DB, l.TxHash.Hex())
	if err != nil {
		logx.Error("HandleAddPoolEvent: failed to check tx hash existence: ", err)
		return false
	}
	if exists {
		logx.Info(fmt.Sprintf("HandleAddPoolEvent: transaction already processed, TxHash=%s", l.TxHash.Hex()))
		return true
	}

	stTokenAddress := ethCommon.HexToAddress(l.Topics[1].Hex()) //质押代币地址
	poolWeight := l.Topics[2].Big()                             //池子权重
	lastRewardBlock := l.Topics[3].Big()                        //最后奖励区块

	params, err := t.ABI.Events["AddPool"].Inputs.UnpackValues(l.Data)
	if err != nil {
		logx.Error("HandleAddPoolEvent: failed to unpack data: ", err)
		return false
	}

	if len(params) < 2 {
		logx.Error("HandleAddPoolEvent: invalid params length")
		return false
	}

	minDepositAmount := params[0].(*big.Int)    //最小质押数量
	unstakeLockedBlocks := params[1].(*big.Int) //解锁区块数

	logx.Info(fmt.Sprintf("AddPool Event - stTokenAddress: %s, poolWeight: %s, lastRewardBlock: %s, minDepositAmount: %s, unstakeLockedBlocks: %s",
		stTokenAddress.Hex(), poolWeight.String(), lastRewardBlock.String(), minDepositAmount.String(), unstakeLockedBlocks.String()))

	block, err := t.Client.BlockByNumber(ctx, big.NewInt(int64(l.BlockNumber)))
	if err != nil {
		logx.Error("HandleAddPoolEvent: failed to get block: ", err)
		return false
	}

	poolID, err := eventaddpool.GetNextPoolID(ctx, t.DB)

	poolWeightFloat, _ := new(big.Float).SetInt(poolWeight).Float64()
	minDepositAmountStr := new(big.Float).Quo(new(big.Float).SetInt(minDepositAmount), big.NewFloat(1e18))
	minDepositAmountFloat, _ := minDepositAmountStr.Float64()

	eventAddPool := &model.EventAddPool{
		ContractAddress:     t.Address,
		PoolID:              poolID,
		StTokenAddress:      stTokenAddress.Hex(),
		PoolWeight:          poolWeightFloat,
		LastRewardBlock:     lastRewardBlock.Uint64(),
		MinDepositAmount:    minDepositAmountFloat,
		UnstakeLockedBlocks: int32(unstakeLockedBlocks.Int64()),
		BlockNumber:         l.BlockNumber,
		BlockTimestamp:      block.Time(),
		TransactionHash:     l.TxHash.Hex(),
		LogIndex:            int32(l.Index),
	}

	if err := eventaddpool.Create(ctx, t.DB, eventAddPool); err != nil {
		logx.Error("HandleAddPoolEvent: failed to create event: ", err)
		return false
	}

	logx.Info(fmt.Sprintf("HandleAddPoolEvent: saved to database, poolID=%d", poolID))
	return true
}

func (t *TaskStake) HandleDepositEvent(l ethereumTypes.Log) bool {
	logx.Info(fmt.Sprintf("HandleDepositEvent: BlockNumber=%d, TxHash=%s, Address=%s, Topics=%v, Data=%x",
		l.BlockNumber, l.TxHash.Hex(), l.Address.Hex(), l.Topics, l.Data))
	return true
}

func (t *TaskStake) HandleClaimEvent(l ethereumTypes.Log) bool {
	logx.Info(fmt.Sprintf("HandleClaimEvent: BlockNumber=%d, TxHash=%s, Address=%s, Topics=%v, Data=%x",
		l.BlockNumber, l.TxHash.Hex(), l.Address.Hex(), l.Topics, l.Data))
	return true
}

func (t *TaskStake) HandleRequestUnstakeEvent(l ethereumTypes.Log) bool {
	logx.Info(fmt.Sprintf("HandleRequestUnstakeEvent: BlockNumber=%d, TxHash=%s, Address=%s, Topics=%v, Data=%x",
		l.BlockNumber, l.TxHash.Hex(), l.Address.Hex(), l.Topics, l.Data))
	return true
}

func (t *TaskStake) HandleWithdrawEvent(l ethereumTypes.Log) bool {
	logx.Info(fmt.Sprintf("HandleWithdrawEvent: BlockNumber=%d, TxHash=%s, Address=%s, Topics=%v, Data=%x",
		l.BlockNumber, l.TxHash.Hex(), l.Address.Hex(), l.Topics, l.Data))
	return true
}
