package common

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func GetKey(chainID int32, address string) string {
	return fmt.Sprintf("%d_%s", chainID, address)
}

func GetABI(abiJSON string) (*abi.ABI, error) {
	wrapABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}
	return &wrapABI, nil
}
