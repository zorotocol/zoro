package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type DB struct {
	PG *sql.DB
}

func (db *DB) InsertBlock(ctx context.Context, block *Block) error {
	tx, err := db.PG.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `insert into "Block" (number) values ($1)`, block.Number); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `delete from "Logs" where deadline < $1`, time.Now()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `delete from "Block" where number < $1`, block.Number); err != nil {
		return err
	}
	for _, log := range block.Logs {
		_, err = tx.ExecContext(ctx, `insert into "Logs" ("passwordHash", "blockNumber", index, tx, password, email, deadline, "nextRetry", "blockedUntil","hours") values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			log.PasswordHash, block.Number, log.Index, log.Tx, log.Password, log.Email, log.Deadline, log.NextRetry, log.BlockedUntil, log.Hours)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) GetLastBlockNumber(ctx context.Context) (int64, error) {
	var blockNumber int64
	err := db.PG.QueryRowContext(ctx, `select number from "Block" order by number desc limit 1`).Scan(&blockNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("DB: GetLastBlockNumber failed: %s", err.Error())
	}
	return blockNumber, nil
}

func (db *DB) GetLogDeadlineByPasswordHash(ctx context.Context, passwordHash string) (time.Time, error) {
	var deadline, blockedUntil time.Time
	err := db.PG.QueryRowContext(ctx, `select deadline,"blockedUntil" from "Logs" where "passwordHash" = $1`, passwordHash).Scan(&deadline, &blockedUntil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("DB: GetLogDeadlineByPasswordHash failed: %s", err.Error())
	}
	//TODO impl firewall
	//if blockedUntil.After(time.Now()) {
	//	return time.Time{}, nil
	//}
	return deadline, nil
}

func (db *DB) AcquireNextLog(ctx context.Context, lockDuration time.Duration) (*Log, error) {
	tx, err := db.PG.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	now := time.Now()
	nextRetry := now.Add(lockDuration)
	var log Log
	err = tx.QueryRowContext(ctx, `select "nextRetry",email,password,deadline,tx,index,hours,"passwordHash" from "Logs" where deadline > $1 and "nextRetry" < $2`, nextRetry, now).Scan(&log.NextRetry, &log.Email, &log.Password, &log.Deadline, &log.Tx, &log.Index, &log.Hours, &log.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("DB: AcquireNextLog failed: %s", err.Error())
	}
	result, err := tx.ExecContext(ctx, `update "Logs" set "nextRetry" = $1 where "passwordHash" = $2 and "nextRetry" = $3`, nextRetry, log.PasswordHash, log.NextRetry)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected != 1 {
		return nil, nil
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &log, nil
}
func (db *DB) SetLogNextRetry(ctx context.Context, passwordHash string, nextRetry time.Time) error {
	_, err := db.PG.ExecContext(ctx, `update "Logs" set "nextRetry" = $1 where "passwordHash" = $2`, nextRetry, passwordHash)
	if err != nil {
		return fmt.Errorf("DB: SetLogNextRetry failed: %s", err.Error())
	}
	return nil
}
