package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:           "./dao/query",
		Mode:              gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldCoverable:    true,
		FieldSignable:     true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		WithUnitTest:      true,
	})

	dsn := "root:st123456@tcp(localhost:3306)/stake_db?charset=utf8mb4&parseTime=True&loc=Local"
	gormdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	g.UseDB(gormdb)

	// 已有的表模型生成
	g.ApplyBasic(
		g.GenerateModel("chain_contracts"),
		g.GenerateModel("chain_endpoints"),
		g.GenerateModel("event_claim"),
		g.GenerateModel("event_withdraw"),
		g.GenerateModel("event_request_unstake"),
		g.GenerateModel("event_deposit"),
		g.GenerateModel("event_update_pool"),
		g.GenerateModel("event_set_pool_weight"),
		g.GenerateModel("event_update_pool_info"),
		g.GenerateModel("event_add_pool"),
		g.GenerateModel("event_set_metanode_per_block"),
		g.GenerateModel("event_set_end_block"),
		g.GenerateModel("event_set_start_block"),
		g.GenerateModel("event_pause_claim"),
		g.GenerateModel("event_pause_withdraw"),
		g.GenerateModel("event_set_metanode"),
		g.GenerateModel("user_pool_stats"),
		g.GenerateModel("pool_info"),
		g.GenerateModel("sync_status"),
	)

	g.Execute()
}
