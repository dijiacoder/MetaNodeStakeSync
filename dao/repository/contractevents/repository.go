package contractevents

import (
	"context"

	"github.com/dijiacoder/MetaNodeStakeSync/dao/model"
	"gorm.io/gorm"
)

func Create(ctx context.Context, db *gorm.DB, item *model.ContractEvent) error {
	return db.WithContext(ctx).Create(item).Error
}

func ExistsByTxHash(ctx context.Context, db *gorm.DB, txHash string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&model.ContractEvent{}).
		Where("transaction_hash = ?", txHash).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
