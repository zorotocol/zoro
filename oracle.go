package oracle

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	libABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/zorotocol/contract"
	"github.com/zorotocol/oracle/internal/db"
	"github.com/zorotocol/oracle/internal/mailer"
	"math/big"
)

type Oracle struct {
	EthClient *ethclient.Client
	DB        *db.DB
	Contracts []common.Address
	Salt      []byte
	Finality  int64
	Mailer    *mailer.Mailer
}

var purchaseEventABI libABI.Event

func init() {
	pabi, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	purchaseEventABI = pabi.Events["Purchase"]
}

var ErrVaporBlock = errors.New("vapor block")

func (ora *Oracle) ProcessBlock(ctx context.Context, number int64) error {
	{
		lastNetworkBlockNumber, err := ora.EthClient.BlockNumber(ctx)
		if err != nil {
			return err
		}
		if number > int64(lastNetworkBlockNumber)-ora.Finality {
			return ErrVaporBlock
		}
	}
	logs, err := ora.EthClient.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(number),
		ToBlock:   big.NewInt(number),
		Addresses: ora.Contracts,
		Topics:    [][]common.Hash{{purchaseEventABI.ID}},
	})
	if err != nil {
		return err
	}
	block := &db.Block{
		Logs:   db.CreateLogs(ora.Salt, logs...),
		Number: number,
	}
	return ora.DB.InsertBlock(ctx, block)
}

func (ora *Oracle) ProcessNextBlock(ctx context.Context) error {
	number, err := ora.DB.GetLastBlockNumber(ctx)
	if err != nil {
		return err
	}
	return ora.ProcessBlock(ctx, number+1)
}
