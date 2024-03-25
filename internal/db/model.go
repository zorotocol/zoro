package db

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mr-tron/base58"
	"github.com/zorotocol/oracle/internal/abi"
	"time"
)

type Block struct {
	Logs   []Log `bson:"Logs,omitempty"`
	Number int64 `bson:"Number"`
}
type Log struct {
	Index        int64
	Tx           string
	Email        string
	Deadline     time.Time
	PasswordHash string
	Password     string
	BlockedUntil time.Time
	NextRetry    time.Time
	Hours        int64
}

func CreateLogs(salt []byte, logs ...types.Log) []Log {
	if len(logs) == 0 {
		return []Log{}
	}
	result := make([]Log, 0, len(logs))
	for _, log := range logs {
		purchase, err := abi.ParsePurchase(log)
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
			Password:     raw,
			PasswordHash: Hash(raw),
			Hours:        purchase.Hours,
			Deadline:     purchase.Deadline,
			Email:        purchase.Email,
			BlockedUntil: time.Time{},
			NextRetry:    time.Time{},
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
