package oracle

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zorotocol/contract"
	"time"
)

type Purchase struct {
	Payer    common.Address
	Hours    int64
	Deadline time.Time
	Email    string
}

var abi = Must(contract.NewContract(common.Address{}, nil))

func ParsePurchase(log types.Log) (*Purchase, error) {

	parsed, err := abi.ParsePurchase(log)
	if err != nil {
		return nil, err
	}
	if !parsed.Deadline.IsInt64() {
		return nil, errors.New("too large deadline")
	}
	if !parsed.Hours.IsInt64() {
		return nil, errors.New("too large hours")
	}
	return &Purchase{
		Payer:    parsed.Payer,
		Hours:    parsed.Hours.Int64(),
		Deadline: time.Unix(parsed.Deadline.Int64(), 0),
		Email:    parsed.Email,
	}, nil
}
