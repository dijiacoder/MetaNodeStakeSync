# GORM Gen 生成的代码结构与使用方法

## 1. 生成的代码结构

当执行 `go run gen/main.go` 后，在 `./dao/query` 目录下会生成以下结构的代码：

```
./dao/query/
├── gen.go           # 生成的入口文件，包含全局变量和初始化函数
├── user_pool_stats.gen.go  # 用户池统计表的模型和查询
├── pool_info.gen.go        # 资金池信息表的模型和查询
├── sync_status.gen.go      # 同步状态表的模型和查询
├── event_*.gen.go          # 各种事件表的模型和查询
└── example_new_table.gen.go # 新添加表的模型和查询
```

## 2. 生成的文件类型说明

### 2.1 模型文件 (Model)

每个表会生成对应的模型结构体，例如 `sync_status` 表生成的模型：

```go
// SyncStatus 同步状态表 - 记录区块同步进度
type SyncStatus struct {
	ID              int     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContractAddress string  `gorm:"type:varchar(42);not null;comment:合约地址" json:"contract_address"`
	ChainID         int     `gorm:"not null;comment:链ID (如 11155111 for Sepolia)" json:"chain_id"`
	LastSyncedBlock uint64  `gorm:"not null;default:0;comment:最后同步的区块号" json:"last_synced_block"`
	LastSyncTime    time.Time `gorm:"autoUpdateTime;comment:最后同步时间" json:"last_sync_time"`
	SyncError       *string `gorm:"type:text;comment:同步错误信息" json:"sync_error"`
	IsSyncing       bool    `gorm:"default:false;comment:是否正在同步" json:"is_syncing"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

### 2.2 查询文件 (Query)

每个表会生成对应的查询结构体，提供丰富的查询方法：

```go
// SyncStatusQuery 提供同步状态表的查询方法
type SyncStatusQuery struct {
	db *gorm.DB
}

// Find 根据条件查询记录
func (q *SyncStatusQuery) Find(conds ...interface{}) ([]*SyncStatus, error) { ... }

// FindOne 根据条件查询单条记录
func (q *SyncStatusQuery) FindOne(conds ...interface{}) (*SyncStatus, error) { ... }

// First 查询第一条记录
func (q *SyncStatusQuery) First() (*SyncStatus, error) { ... }

// Create 创建记录
func (q *SyncStatusQuery) Create(syncStatus *SyncStatus) error { ... }

// Update 更新记录
func (q *SyncStatusQuery) Update(syncStatus *SyncStatus) error { ... }

// Where 根据条件筛选
func (q *SyncStatusQuery) Where(conds ...interface{}) *SyncStatusQuery { ... }

// OrderBy 排序
func (q *SyncStatusQuery) OrderBy(expression string) *SyncStatusQuery { ... }

// Limit 限制结果数量
func (q *SyncStatusQuery) Limit(limit int) *SyncStatusQuery { ... }

// Offset 偏移量
func (q *SyncStatusQuery) Offset(offset int) *SyncStatusQuery { ... }
```

## 3. 使用方法

### 3.1 初始化查询实例

在应用启动时，需要初始化查询实例：

```go
package main

import (
	"github.com/dijiacoder/MetaNodeStakeSync/dao/query"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 连接数据库
	dsn := "root:st123456@tcp(localhost:3306)/stake_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 初始化查询实例
	query.SetDefault(db)
	
	// 开始使用查询
	// ...
}
```

### 3.2 使用生成的查询方法

#### 查询示例

```go
import (
	"github.com/dijiacoder/MetaNodeStakeSync/dao/query"
)

// 获取特定合约和链的同步状态
func getSyncStatus(contractAddress string, chainID int) (*query.SyncStatus, error) {
	return query.SyncStatus.Where(
		query.SyncStatus.ContractAddress.Eq(contractAddress),
		query.SyncStatus.ChainID.Eq(chainID),
	).FindOne()
}

// 获取所有活跃的资金池
func getActivePools() ([]*query.PoolInfo, error) {
	return query.PoolInfo.Where(
		query.PoolInfo.IsActive.Eq(true),
	).Find()
}

// 查询用户在特定池的统计信息
func getUserPoolStats(userAddress string, poolID int) (*query.UserPoolStats, error) {
	return query.UserPoolStats.Where(
		query.UserPoolStats.UserAddress.Eq(userAddress),
		query.UserPoolStats.PoolID.Eq(poolID),
	).FindOne()
}
```

#### 创建和更新示例

```go
// 更新同步状态
func updateSyncStatus(contractAddress string, chainID int, blockNumber uint64) error {
	// 查找现有记录
	syncStatus, err := query.SyncStatus.Where(
		query.SyncStatus.ContractAddress.Eq(contractAddress),
		query.SyncStatus.ChainID.Eq(chainID),
	).FindOne()
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新记录
			syncStatus = &query.SyncStatus{
				ContractAddress: contractAddress,
				ChainID:         chainID,
				LastSyncedBlock: blockNumber,
				IsSyncing:       false,
			}
			return query.SyncStatus.Create(syncStatus)
		}
		return err
	}
	
	// 更新现有记录
	syncStatus.LastSyncedBlock = blockNumber
	return query.SyncStatus.Update(syncStatus)
}
```

#### 复杂查询示例

```go
// 获取特定时间段内的所有质押事件
func getDepositEventsByTimeRange(startTime, endTime time.Time) ([]*query.EventDeposit, error) {
	return query.EventDeposit.Where(
		query.EventDeposit.CreatedAt.Between(startTime, endTime),
	).OrderBy("created_at desc").Find()
}

// 获取用户的所有事件记录（使用联表查询）
func getUserEvents(userAddress string, limit, offset int) ([]*query.EventClaim, error) {
	return query.EventClaim.Where(
		query.EventClaim.UserAddress.Eq(userAddress),
	).Limit(limit).Offset(offset).OrderBy("block_number desc").Find()
}
```

## 4. 使用技巧

### 4.1 链式调用

生成的查询支持链式调用，可以组合多个条件：

```go
result, err := query.PoolInfo.Where(
	query.PoolInfo.IsActive.Eq(true),
).Where(
	query.PoolInfo.MinDepositAmount.Lt(1000),
).OrderBy("pool_weight desc").Limit(10).Find()
```

### 4.2 使用事务

```go
tx := query.DB().Begin()

// 在事务中使用查询
poolQuery := query.PoolInfo.WithContext(tx.Statement.Context)
userQuery := query.UserPoolStats.WithContext(tx.Statement.Context)

// 执行操作
if err := poolQuery.Create(newPool); err != nil {
	tx.Rollback()
	return err
}

if err := userQuery.Update(userStats); err != nil {
	tx.Rollback()
	return err
}

return tx.Commit().Error
```

### 4.3 预加载关联

```go
// 假设有关联关系的情况下
result, err := query.PoolInfo.Preload(query.PoolInfo.UserPoolStats).Find()
```

### 4.4 批量操作

```go
// 批量创建
var events []*query.EventClaim
// ... 添加事件到 events 切片
if err := query.EventClaim.CreateInBatches(events, 100); err != nil {
	// 处理错误
}

// 批量更新
if err := query.UserPoolStats.Where(
	query.UserPoolStats.PoolID.Eq(poolID),
).Update(
	query.UserPoolStats.StAmount.Add(bonusAmount),
).Error; err != nil {
	// 处理错误
}
```

## 5. 注意事项

1. 每次数据库表结构变更后，需要重新运行 `go run gen/main.go` 生成最新的代码
2. 生成的代码不应手动修改，所有修改应通过修改生成配置或数据库表结构后重新生成
3. 在使用生成的查询时，应优先使用 `query.表名.字段名.Eq(value)` 形式的条件构造，而不是原始字符串条件，这样可以获得更好的类型安全性
4. 对于复杂的查询，可以使用 `query.Raw()` 方法执行原生 SQL

## 6. 代码生成命令

在项目根目录执行以下命令生成代码：

```bash
go run gen/main.go
```

如果是首次运行，需要确保数据库已经创建并包含相应的表结构。