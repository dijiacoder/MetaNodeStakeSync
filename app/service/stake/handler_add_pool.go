package stake

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/dijiacoder/MetaNodeStakeSync/dao/model"
	"github.com/dijiacoder/MetaNodeStakeSync/dao/repository/poolinfo"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
)

func (t *TaskStake) HandleAddPoolEvent(l ethereumTypes.Log) error {
	// 校验topics长度（topic0签名 + 三个indexed参数）
	if len(l.Topics) < 4 {
		return fmt.Errorf("HandleAddPoolEvent: invalid topics length, tx=%s", l.TxHash.Hex())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先查重：若已按创建交易写入过pool_info，则跳过
	exists, err := poolinfo.ExistsByCreatedTx(ctx, t.DB, l.TxHash.Hex())
	if err != nil {
		return fmt.Errorf("HandleAddPoolEvent: ExistsByCreatedTx error: %w", err)
	}
	if exists {
		return nil
	}

	// 解析indexed参数
	stTokenAddress := l.Topics[1].Hex()  // 质押代币地址
	poolWeight := l.Topics[2].Big()      // 池子权重
	lastRewardBlock := l.Topics[3].Big() // 最后奖励区块

	// 解析非indexed参数（data中）
	params, err := t.ABI.Events["AddPool"].Inputs.UnpackValues(l.Data)
	if err != nil {
		return fmt.Errorf("HandleAddPoolEvent: unpack data error: %w", err)
	}
	if len(params) < 2 {
		return fmt.Errorf("HandleAddPoolEvent: invalid params length")
	}
	minDepositAmount := params[0].(*big.Int)    // 最小质押数量
	unstakeLockedBlocks := params[1].(*big.Int) // 解锁区块数

	// 获取区块时间
	block, err := t.Client.BlockByNumber(ctx, big.NewInt(int64(l.BlockNumber)))
	if err != nil {
		return fmt.Errorf("HandleAddPoolEvent: get block error: %w", err)
	}

	// 计算下一个PoolID（同一合约地址下：如果当前是0条，则NextPoolID=0）
	poolID, err := poolinfo.GetNextPoolID(ctx, t.DB, t.Address)
	if err != nil {
		return fmt.Errorf("HandleAddPoolEvent: GetNextPoolID error: %w", err)
	}

	// 类型转换
	poolWeightFloat, _ := new(big.Float).SetInt(poolWeight).Float64()
	minDepositAmountFloat, _ := new(big.Float).
		Quo(new(big.Float).SetInt(minDepositAmount), big.NewFloat(1e18)).
		Float64()

	isActive := true
	createdBlock := l.BlockNumber
	createdTx := l.TxHash.Hex()
	createdAt := time.Unix(int64(block.Time()), 0)

	// 构建并写入 pool_info
	item := &model.PoolInfo{
		PoolID:              poolID,
		ContractAddress:     t.Address,
		StTokenAddress:      stTokenAddress,
		PoolWeight:          poolWeightFloat,
		LastRewardBlock:     lastRewardBlock.Uint64(),
		MinDepositAmount:    minDepositAmountFloat,
		UnstakeLockedBlocks: int32(unstakeLockedBlocks.Int64()),
		IsActive:            &isActive,
		CreatedBlock:        &createdBlock,
		CreatedTx:           &createdTx,
		CreatedAt:           &createdAt,
	}

	if err := poolinfo.Create(ctx, t.DB, item); err != nil {
		return fmt.Errorf("HandleAddPoolEvent: create pool_info error: %w", err)
	}

	return nil
}
