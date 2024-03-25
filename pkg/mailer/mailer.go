package mailer

import (
	"context"
	"github.com/zorotocol/oracle/pkg/db"
	"github.com/zorotocol/oracle/pkg/multirun"
	"math/rand"
	"runtime"
	"time"
)

type Sender func(ctx context.Context, mail *db.Log) error
type Mailer struct {
	DB     *db.DB
	Delay  time.Duration
	Sender Sender
}

func (m *Mailer) sendMail(ctx context.Context, mail *db.Log) error {
	return m.Sender(ctx, mail)
}

func (m *Mailer) Start(parentCtx context.Context) error {
	workers := make([]multirun.Job, runtime.NumCPU())
	for i := range workers {
		workers[i] = func(ctx context.Context) error {
			return worker(m, ctx)
		}
	}
	return multirun.Run(parentCtx, workers...)
}
func worker(m *Mailer, ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		mail, err := m.DB.AcquireNextLog(ctx, m.Delay)
		if err != nil {
			return err
		}
		if mail == nil {
			if err := sleepCtx(ctx, randomDuration(time.Second, time.Second*2)); err != nil {
				return err
			}
			continue
		}
		if sendError := m.sendMail(ctx, mail); sendError == nil {
			_ = m.DB.SetLogNextRetry(ctx, mail.PasswordHash, mail.Deadline)
		}
	}
}

func randomDuration(min, max time.Duration) time.Duration {
	return time.Duration(rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(int64(max-min)) + int64(min))
}
func sleepCtx(ctx context.Context, sleep time.Duration) error {
	timer := time.NewTimer(sleep)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
