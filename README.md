# MetaNodeStakeSync

## 项目命令

```
# 
go run .\app\main.go daemon -c .\app\config\config.yaml
```

## 常用命令

```
# 初始化模块
go mod init github.com/dijiacoder/MetaNodeStakeSync

# 整理当前模块的依赖
go mod tidy

# 获取依赖并更新 go.mod 和 go.sum 文件
go get <package>

# 编译 Go 代码，生成可执行文件, <name>: app, app.exe
go build -o <name>
```

## 常用依赖

```
# 为应用程序提供灵活的配置管理解决方案
go get github.com/spf13/viper

# Cobra 框架的官方命令行生成工具
go get github.com/spf13/cobra-cli@latest

# 用于处理用户主目录(home directory)的路径问题
go get github.com/mitchellh/go-homedir

go get github.com/go-redis/redis/v8

go get github.com/ethereum/go-ethereum

go get github.com/ethereum/go-ethereum/ethclient

go get gorm.io/driver/mysql

go get gorm.io/gorm

go get gorm.io/gen
```

## 数据库模型生成

```
go run gen/main.go
```