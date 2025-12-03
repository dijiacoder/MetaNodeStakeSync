package eventaddpool

import (
	"context"

	"github.com/dijiacoder/MetaNodeStakeSync/dao/model"
	"gorm.io/gorm"
)

// Create 新增一条 EventAddPool 记录
func Create(ctx context.Context, db *gorm.DB, item *model.EventAddPool) error {
	return db.WithContext(ctx).Create(item).Error
}

// GetByID 根据主键ID查询 EventAddPool
func GetByID(ctx context.Context, db *gorm.DB, id int64) (*model.EventAddPool, error) {
	var res model.EventAddPool
	if err := db.WithContext(ctx).First(&res, id).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdateByID 根据主键ID更新指定字段
// updates 示例：map[string]interface{}{"pool_weight": 123.0}
func UpdateByID(ctx context.Context, db *gorm.DB, id int64, updates map[string]interface{}) error {
	return db.WithContext(ctx).Model(&model.EventAddPool{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteByID 根据主键ID删除记录
func DeleteByID(ctx context.Context, db *gorm.DB, id int64) error {
	return db.WithContext(ctx).Delete(&model.EventAddPool{}, id).Error
}
