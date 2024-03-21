package oracle

import (
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"time"
)

var meta = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_payer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_deadline\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"email\",\"type\":\"string\"}],\"name\":\"Purchase\",\"type\":\"event\"}]",
}

func ABI() *abi.ABI {
	a, err := meta.GetAbi()
	if err != nil {
		panic(err)
	}
	return a
}

type Purchase struct {
	Payer    common.Address
	Amount   *big.Int
	Deadline time.Time
	Email    string
}

var contract = bind.NewBoundContract(common.Address{}, *ABI(), nil, nil, nil)

func ParsePurchase(log types.Log) (*Purchase, error) {
	var event struct {
		Payer    common.Address
		Amount   *big.Int
		Deadline *big.Int
		Email    string
	}
	if err := contract.UnpackLog(&event, "Purchase", log); err != nil {
		return nil, err
	}
	if !event.Deadline.IsInt64() {
		return nil, errors.New("too large deadline")
	}
	return &Purchase{
		Payer:    event.Payer,
		Amount:   event.Amount,
		Deadline: time.Unix(event.Deadline.Int64(), 0),
		Email:    event.Email,
	}, nil
}
