package oracle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mr-tron/base58"
	"time"
)

type Block struct {
	Logs   []Log `bson:"Logs,omitempty"`
	Number int64 `bson:"Number"`
}
type Log struct {
	Index        int64     `bson:"Index"`
	Tx           string    `bson:"Tx"`
	Email        string    `bson:"Email"`
	Hours        int64     `bson:"Hours"`
	Deadline     time.Time `bson:"Deadline"`
	Payer        string    `bson:"Payer"`
	Token        string    `bson:"Token"`
	Raw          string    `bson:"Raw"`
	BlockedUntil time.Time `bson:"BlockedUntil,omitempty"`
}

func createLogs(salt []byte, logs ...types.Log) []Log {
	if len(logs) == 0 {
		return []Log{}
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
		raw := GenerateToken(salt, log.TxHash)
		result = append(result, Log{
			Index:        int64(log.Index),
			Tx:           log.TxHash.String(),
			Raw:          raw,
			Token:        Hash(raw),
			Hours:        purchase.Hours,
			Deadline:     purchase.Deadline,
			Payer:        purchase.Payer.String(),
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
func GenerateToken(salt []byte, tx common.Hash) (raw string) {
	hasher := sha256.New()
	hasher.Write(salt)
	hasher.Write(tx[:])
	return base58.Encode(hasher.Sum(nil))
}
