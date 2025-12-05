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
		g.GenerateModel("contract_events"),
		g.GenerateModel("pool_info"),
	)

	g.Execute()
}
