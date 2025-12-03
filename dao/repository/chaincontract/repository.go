package chaincontract

import (
	"fmt"

	"github.com/dijiacoder/MetaNodeStakeSync/dao/model"
	"github.com/dijiacoder/MetaNodeStakeSync/dao/query"
)

type ChainContract model.ChainContract

type ChainEndpoint model.ChainEndpoint

type ContractWithEndpoint struct {
	ChainContract ChainContract
	ChainEndpoint ChainEndpoint
}

func GetContractWithEndPoint() ([]ContractWithEndpoint, error) {
	// 从数据库查询合约和端点信息
	contracts, err := query.ChainContract.Find()
	if err != nil {
		return nil, fmt.Errorf("failed to query chain contracts: %v", err)
	}

	// 获取所有端点信息
	endpoints, err := query.ChainEndpoint.Find()
	if err != nil {
		return nil, fmt.Errorf("failed to query chain endpoints: %v", err)
	}

	// 创建端点映射
	endpointMap := make(map[int32]ChainEndpoint)
	for _, endpoint := range endpoints {
		endpointMap[endpoint.ChainID] = ChainEndpoint(*endpoint)
	}

	// 构建结果
	var result []ContractWithEndpoint
	for _, contract := range contracts {
		contractWithEndpoint := ContractWithEndpoint{
			ChainContract: ChainContract(*contract),
			ChainEndpoint: endpointMap[contract.ChainID],
		}
		result = append(result, contractWithEndpoint)
	}

	return result, nil
}
