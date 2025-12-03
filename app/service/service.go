package service

import (
	"context"
	"fmt"

	"github.com/dijiacoder/MetaNodeStakeSync/app/service/common"
	"github.com/dijiacoder/MetaNodeStakeSync/app/service/config"
	"github.com/dijiacoder/MetaNodeStakeSync/app/service/stake"
	"github.com/dijiacoder/MetaNodeStakeSync/dao/query"
	"github.com/dijiacoder/MetaNodeStakeSync/dao/repository/chaincontract"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Service struct {
	serviceCtx *common.ServiceContext
}

func New(ctx context.Context, config *config.Config) (*Service, error) {
	logx.Info(fmt.Sprintf("DB connection string: %s", config.DB.DSN))
	db, err := gorm.Open(mysql.Open(config.DB.DSN), &gorm.Config{})
	db = db.Debug()
	if err != nil {
		panic(err)
	}
	query.SetDefault(db)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	contracts, err := chaincontract.GetContractWithEndPoint()
	if err != nil {
		panic(err)
	}

	contractInfoMap := make(map[int32]*common.ContractInfo)

	for _, contract := range contracts {
		logx.Info(fmt.Sprintf("ContractName: %d, ChainID: %d, CreatedTxHash: %s, ContractAddress: %s, ChainEndpointURL: %s", contract.ChainContract.ContractName,
			contract.ChainContract.ChainID, *contract.ChainContract.CreatedTxHash, contract.ChainContract.ContractAddress, contract.ChainEndpoint.URL))
		ethClient, err := ethclient.Dial(contract.ChainEndpoint.URL)
		if err != nil {
			panic(err)
		}

		contractInfo := &common.ContractInfo{
			ChainID:     contract.ChainContract.ChainID,
			ABIStr:      contract.ChainContract.Abi,
			Address:     contract.ChainContract.ContractAddress,
			CreatedHash: contract.ChainContract.CreatedTxHash,
			Client:      ethClient,
		}
		contractInfoMap[contract.ChainContract.ContractName] = contractInfo
	}

	service := &Service{
		serviceCtx: &common.ServiceContext{
			Context:         ctx,
			Config:          config,
			DB:              db,
			RedisClient:     redisClient,
			ContractInfoMap: contractInfoMap,
		},
	}
	return service, nil
}

func (service *Service) Start() {
	//stake contract name: 1
	stake.NewTaskStake(service.serviceCtx).Start()
}
