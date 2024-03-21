package oracle

import (
	"context"
	"crypto/rand"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/zorotocol/contract"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"math/big"
)

type Oracle struct {
	EthClient  *ethclient.Client
	Collection *mongo.Collection
	Contracts  []common.Address
}

var purchaseEventABI = Must(contract.ContractMetaData.GetAbi()).Events["Purchase"]

func (ora *Oracle) ProcessBlock(ctx context.Context, number int64) error {
	logs, err := ora.EthClient.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(number),
		ToBlock:   big.NewInt(number),
		Addresses: ora.Contracts,
		Topics:    [][]common.Hash{{purchaseEventABI.ID}},
	})
	if err != nil {
		return err
	}
	var salt [32]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return err
	}
	block := &Block{
		Logs:   createLogs(salt[:], logs...),
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
