package oracle

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"time"
)

type Block struct {
	Logs   []Log `bson:"Logs,omitempty"`
	Number int64 `bson:"Number"`
}
type Log struct {
	Index        uint64         `bson:"Index"`
	Tx           common.Hash    `bson:"Tx"`
	Email        string         `bson:"Email"`
	Amount       *big.Int       `bson:"Amount"`
	Deadline     time.Time      `bson:"Deadline"`
	Payer        common.Address `bson:"Payer"`
	Token        string         `bson:"Token"`
	Raw          string         `bson:"Raw"`
	BlockedUntil time.Time      `bson:"BlockedUntil"`
}

func createLogs(salt []byte, logs ...types.Log) []Log {
	if len(logs) == 0 {
		panic(errors.New("empty logs list"))
	}
	result := make([]Log, 0, len(logs))
	for _, log := range logs {
		purchase, err := ParsePurchase(log)
		if err != nil {
			continue
		}
		if log.BlockNumber != logs[0].BlockNumber {
			panic(errors.New("inconsistent block number"))
		}
		raw := GenerateToken(salt, log.BlockHash, uint64(log.Index))
		result = append(result, Log{
			Index:        uint64(log.Index),
			Tx:           log.TxHash,
			Raw:          raw,
			Token:        Hash(raw),
			Amount:       purchase.Amount,
			Deadline:     purchase.Deadline,
			Payer:        purchase.Payer,
			Email:        purchase.Email,
			BlockedUntil: time.Time{},
		})
	}
	return result
}
func Hash(str string) string {
	b := sha256.Sum224([]byte(str))
	return hex.EncodeToString(b[:])
}
func GenerateToken(salt []byte, blockHash common.Hash, logIndex uint64) (raw string) {
	hasher := sha256.New()
	hasher.Write(salt)
	hasher.Write(blockHash[:])
	{
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], logIndex)
		hasher.Write(b[:])
	}
	return base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
}
