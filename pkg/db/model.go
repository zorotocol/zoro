package db

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/badoux/checkmail"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mr-tron/base58"
	"github.com/zorotocol/zoro/pkg/abi"
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
		if checkmail.ValidateFormat(purchase.Email) != nil {
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
			PasswordHash: Hash([]byte(raw)),
			Hours:        purchase.Hours,
			Deadline:     purchase.Deadline,
			Email:        purchase.Email,
			BlockedUntil: time.Time{},
			NextRetry:    time.Time{},
		})
	}
	return result
}
func Hash(str []byte) string {
	b := sha256.Sum224(str)
	return hex.EncodeToString(b[:])
}
func GenerateToken(salt []byte, tx common.Hash) (raw string) {
	hasher := sha256.New224()
	hasher.Write(salt)
	hasher.Write(tx[:])
	return base58.Encode(hasher.Sum(nil))
}

func ValidatePasswordHash(hash string) bool {
	if len(hash) != hex.EncodedLen(sha256.Size224) {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}
