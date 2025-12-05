package poolinfo

import (
	"context"

	"github.com/dijiacoder/MetaNodeStakeSync/dao/model"
	"gorm.io/gorm"
)

func Create(ctx context.Context, db *gorm.DB, item *model.PoolInfo) error {
	return db.WithContext(ctx).Create(item).Error
}

func GetByPoolIDAndContract(ctx context.Context, db *gorm.DB, poolID int32, contractAddress string) (*model.PoolInfo, error) {
	var res model.PoolInfo
	if err := db.WithContext(ctx).Where("pool_id = ? AND contract_address = ?", poolID, contractAddress).First(&res).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

func UpdateByPoolIDAndContract(ctx context.Context, db *gorm.DB, poolID int32, contractAddress string, updates map[string]interface{}) error {
	return db.WithContext(ctx).Model(&model.PoolInfo{}).Where("pool_id = ? AND contract_address = ?", poolID, contractAddress).Updates(updates).Error
}

func ExistsByPoolIDAndContract(ctx context.Context, db *gorm.DB, poolID int32, contractAddress string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Model(&model.PoolInfo{}).Where("pool_id = ? AND contract_address = ?", poolID, contractAddress).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func ExistsByCreatedTx(ctx context.Context, db *gorm.DB, createdTx string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Model(&model.PoolInfo{}).Where("created_tx = ?", createdTx).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetNextPoolID 根据表中的数据量计算NextPoolID，如果当前是0条，则NextPoolID=0。
func GetNextPoolID(ctx context.Context, db *gorm.DB, contractAddress string) (int32, error) {
	var count int64
	if err := db.WithContext(ctx).
		Model(&model.PoolInfo{}).
		Where("contract_address = ?", contractAddress).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return int32(count), nil
}
