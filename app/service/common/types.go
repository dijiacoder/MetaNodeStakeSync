package common

import (
	"context"

	"github.com/dijiacoder/MetaNodeStakeSync/app/service/config"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Context         context.Context
	Config          *config.Config
	DB              *gorm.DB
	RedisClient     *redis.Client
	ContractInfoMap map[int32]*ContractInfo
}

type ContractInfo struct {
	ChainID      int32
	ContractName string
	ABIStr       string
	Address      string
	CreatedHash  *string
	Client       *ethclient.Client
}
