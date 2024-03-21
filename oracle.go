package oracle

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/zorotocol/contract"
	"go.mongodb.org/mongo-driver/mongo"
	"math/big"
)

type Oracle struct {
	EthClient  *ethclient.Client
	Collection *mongo.Collection
	Contracts  []common.Address
	Salt       []byte
	Finality   int64
}

var purchaseEventABI = Must(contract.ContractMetaData.GetAbi()).Events["Purchase"]

func (ora *Oracle) ProcessBlock(ctx context.Context, number int64) error {
	{
		lastNetworkBlockNumber, err := ora.EthClient.BlockNumber(ctx)
		if err != nil {
			return err
		}
		if number > int64(lastNetworkBlockNumber)-ora.Finality {
			return errors.New("vapor or far block number")
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
	block := &Block{
		Logs:   createLogs(ora.Salt, logs...),
		Number: number,
	}
	if err = ensureIndices(ctx, ora.Collection); err != nil {
		return err
	}
	return insertBlock(ctx, ora.Collection, block)
}

func (ora *Oracle) ProcessNextBlock(ctx context.Context) error {
	number, err := getLastBlockNumber(ctx, ora.Collection)
	if err != nil {
		return err
	}
	return ora.ProcessBlock(ctx, number+1)
}
