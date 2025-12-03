# GORM Gen 优化与扩展建议

## 1. 配置优化

### 1.1 生成模式优化

根据项目需求选择合适的生成模式：

```go
// 针对高并发场景，使用WithContext模式
Mode: gen.WithDefaultQuery | gen.WithQueryInterface,

// 对于简单应用，可以使用WithoutContext提高性能
Mode: gen.WithoutContext | gen.WithDefaultQuery,
```

### 1.2 字段选项优化

根据数据特点调整字段配置：

```go
FieldNullable:     true,   // 允许空值字段（推荐用于区块链数据，可能有未知值）
FieldCoverable:    true,   // 允许覆盖字段值（用于更新操作）
FieldSignable:     true,   // 处理有符号整数（推荐用于区块链数据）
FieldWithIndexTag: true,   // 生成索引标签（提高查询性能）
FieldWithTypeTag:  true,   // 生成类型标签（增强类型安全性）
```

### 1.3 测试配置

根据需要调整测试生成：

```go
WithUnitTest: false,  // 生产环境可以禁用测试生成，减小代码体积
WithMock:     true,   // 开发环境可以启用mock生成，便于单元测试
```

## 2. 高级功能扩展

### 2.1 自定义类型映射

为区块链特有的数据类型（如地址、哈希）配置自定义映射：

```go
// 在gen/main.go中添加自定义类型映射
g.WithDataTypeMap(map[string]func(gorm.ColumnType) (dataType string){
	"varchar(42)": func(columnType gorm.ColumnType) (dataType string) {
		// 将地址字段映射为特定的类型
		if strings.Contains(strings.ToLower(columnType.Name()), "address") {
			return "string"  // 或者返回自定义的Address类型
		}
		return "string"
	},
	"varchar(66)": func(columnType gorm.ColumnType) (dataType string) {
		// 将哈希字段映射为特定的类型
		if strings.Contains(strings.ToLower(columnType.Name()), "tx") || 
		   strings.Contains(strings.ToLower(columnType.Name()), "hash") {
			return "string"  // 或者返回自定义的Hash类型
		}
		return "string"
	},
})
```

### 2.2 生成自定义方法

为生成的模型添加自定义方法：

```go
// 在gen/main.go中添加自定义方法生成
type Method struct{}

func (m Method) AddAddressValidation(table gen.Table) []string {
	return []string{
		fmt.Sprintf(`// ValidateAddress 验证地址格式
func (m *%s) ValidateAddress() bool {
	// 实现以太坊地址验证逻辑
	return len(m.%s) == 42 && strings.HasPrefix(m.%s, "0x")
}`, table.Name(), "ContractAddress", "ContractAddress"),
	}
}

// 应用自定义方法
g.WithMethod(Method{})
```

### 2.3 批量操作优化

为大量区块链数据处理添加批量操作支持：

```go
// 在gen/main.go中添加批量操作配置
g.WithOpts(gen.Opt{
	// 启用批量插入优化
	PrepareStmt: true,
})
```

## 3. 性能优化建议

### 3.1 连接池配置

针对高吞吐量的区块链数据同步，优化数据库连接池：

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	PrepareStmt: true,  // 启用预编译语句缓存
})

// 获取底层sql.DB实例优化连接池
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
```

### 3.2 索引优化

根据查询模式确保数据库表有适当的索引：

```sql
-- 为频繁查询的字段创建复合索引
CREATE INDEX idx_event_address_block ON event_claim(contract_address, block_number);
CREATE INDEX idx_user_pool ON user_pool_stats(user_address, pool_id);
```

### 3.3 查询优化

使用生成的查询时的性能建议：

- 避免全表扫描，总是使用索引字段进行过滤
- 对于大量事件数据，使用分页查询
- 只选择需要的字段，使用 `Select()` 方法
- 对于统计查询，优先使用数据库的聚合函数

## 4. 错误处理增强

### 4.1 自定义错误处理

为生成的查询添加统一的错误处理：

```go
// 创建错误处理包装函数
func handleDBError(err error) error {
	if err == nil {
		return nil
	}
	
	// 处理特定错误类型
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("数据未找到: %w", err)
	}
	
	// 处理数据库错误
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// 处理MySQL特定错误
		return fmt.Errorf("数据库错误[%d]: %s", mysqlErr.Number, mysqlErr.Message)
	}
	
	return fmt.Errorf("查询错误: %w", err)
}

// 使用示例
func getUserStats(userAddress string) (*query.UserPoolStats, error) {
	stats, err := query.UserPoolStats.Where(
		query.UserPoolStats.UserAddress.Eq(userAddress),
	).FindOne()
	return stats, handleDBError(err)
}
```

### 4.2 事务错误处理

为区块链数据同步中的事务操作添加错误处理：

```go
func processBlockEvents(events []*BlockEvent) error {
	tx := query.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("事务执行过程中发生panic: %v", r)
		}
	}()

	// 在事务中使用查询
	eventQuery := query.EventClaim.WithContext(tx.Statement.Context)

	for _, event := range events {
		// 处理事件...
		if err := eventQuery.Create(convertToEventClaim(event)); err != nil {
			tx.Rollback()
			return fmt.Errorf("创建事件记录失败: %w", err)
		}
	}

	return tx.Commit().Error
}
```

## 5. 扩展性建议

### 5.1 接口抽象

为生成的查询添加接口层，便于扩展和测试：

```go
// 在dao/interfaces.go中定义接口
type SyncStatusRepo interface {
	FindByContractAndChain(contractAddress string, chainID int) (*query.SyncStatus, error)
	UpdateLastSyncedBlock(id int, blockNumber uint64) error
	Create(status *query.SyncStatus) error
}

// 实现接口
type syncStatusRepo struct{}

func (r *syncStatusRepo) FindByContractAndChain(contractAddress string, chainID int) (*query.SyncStatus, error) {
	return query.SyncStatus.Where(
		query.SyncStatus.ContractAddress.Eq(contractAddress),
		query.SyncStatus.ChainID.Eq(chainID),
	).FindOne()
}

// 其他方法实现...
```

### 5.2 多链支持扩展

为支持多链数据同步扩展查询功能：

```go
// 创建多链查询助手
func getChainQuery(chainID int) *query.Query {
	// 根据chainID选择不同的数据库连接
	switch chainID {
	case 1: // Ethereum Mainnet
		return queryETH
	case 11155111: // Sepolia Testnet
		return querySepolia
	default:
		return query.Default
	}
}
```

### 5.3 数据归档扩展

为区块链历史数据添加归档功能：

```go
// 创建归档表生成配置
func generateArchiveModels(g *gen.Generator) {
	// 为需要归档的表添加归档配置
	g.GenerateModel("event_claim_archive")
	g.GenerateModel("event_deposit_archive")
}
```

## 6. 监控与维护

### 6.1 查询性能监控

添加查询性能监控：

```go
// 在gorm配置中添加监控中间件
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	Logger: logger.Default.LogMode(logger.Info),
})

// 添加自定义监控中间件
db.Use(gormlog.New(gin.DefaultWriter, gormlog.Config{
	SlowThreshold: time.Second, // 慢查询阈值
	LogLevel:      logger.Info, // 日志级别
}))
```

### 6.2 定期重新生成

建立定期重新生成代码的流程：

1. 数据库表结构变更后立即重新生成
2. 创建Makefile或脚本简化生成过程

```bash
# 创建scripts/regenerate.sh
#!/bin/bash
set -e
echo "Regenerating GORM models..."
go run gen/main.go
echo "Models regenerated successfully!"
```

### 6.3 版本控制

对生成的代码进行版本控制：

- 将生成的代码纳入版本控制，便于追踪变更
- 使用代码审查确保生成的代码符合项目标准
- 定期更新依赖，确保使用最新版本的gorm.io/gen

## 7. 总结

- 根据项目需求优化gorm.io/gen配置
- 利用高级功能扩展生成的查询能力
- 为区块链数据处理添加性能优化
- 增强错误处理确保数据一致性
- 建立良好的维护和监控机制

通过合理配置和扩展，可以充分发挥gorm.io/gen在区块链数据处理中的优势，提高开发效率和系统性能。