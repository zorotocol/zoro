package oracle

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	libABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/zorotocol/contract"
	"go.mongodb.org/mongo-driver/mongo"
	"math/big"
	"time"
)

type Oracle struct {
	EthClient  *ethclient.Client
	Collection *mongo.Collection
	Contracts  []common.Address
	Salt       []byte
	Finality   int64
	Mailer     *Mailer
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
	block := &Block{
		Logs:   createLogs(ora.Salt, logs...),
		Number: number,
	}
	if err = ensureIndices(ctx, ora.Collection); err != nil {
		return err
	}
	err = insertBlock(ctx, ora.Collection, block)
	if err != nil {
		return err
	}
	_ = cleanEmptyLogs(context.TODO(), ora.Collection, number)
	_ = enqueueMails(context.TODO(), ora.Mailer, block.Logs...)
	return nil
}
func enqueueMails(ctx context.Context, mail *Mailer, logs ...Log) error {
	for _, log := range logs {
		_ = mail.Enqueue(ctx, time.Time{}, log.Deadline, &Mail{
			Key:      log.Raw,
			Tx:       log.Tx,
			LogIndex: log.Index,
			Deadline: log.Deadline,
			Email:    log.Email,
		})
	}
	return nil
}
func (ora *Oracle) ProcessNextBlock(ctx context.Context) error {
	number, err := getLastBlockNumber(ctx, ora.Collection)
	if err != nil {
		return err
	}
	return ora.ProcessBlock(ctx, number+1)
}
