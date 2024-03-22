package oracle

import (
	"context"
	"errors"
	"github.com/zorotocol/oracle/pkg/multirun"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	"runtime"
	"time"
)

type Sender func(ctx context.Context, mail *Mail) error
type Mailer struct {
	Collection *mongo.Collection
	Delay      time.Duration
	Sender     Sender
}
type Mail struct {
	Key      string    `bson:"Key"`
	Email    string    `bson:"Email"`
	Tx       string    `bson:"Tx"`
	LogIndex int64     `bson:"LogIndex"`
	Deadline time.Time `bson:"Deadline"`
}

type record struct {
	Mail  Mail  `bson:"Mail"`
	Start int64 `bson:"Start"`
	End   int64 `bson:"End"`
}

func (m *Mailer) sendMail(ctx context.Context, mail *Mail) error {
	return m.Sender(ctx, mail)
}
func ensureIndexes(ctx context.Context, col *mongo.Collection) error {
	t := true
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"Mail.Key": 1},
		Options: &options.IndexOptions{
			Unique: &t,
		},
	})
	_, err = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{bson.E{"Start", 1}, bson.E{"End", 1}},
		Options: nil,
	})
	return err
}
func removeMail(col *mongo.Collection, key string) {
	_, _ = col.DeleteOne(context.Background(), bson.M{"Mail.Key": key})
}
func pullMail(ctx context.Context, col *mongo.Collection, lockDuration time.Duration) (*Mail, error) {
	now := time.Now().Unix()
	newStart := now + (lockDuration.Milliseconds() / 1000)
	var rec record
	if err := col.FindOneAndUpdate(ctx, bson.M{
		"Start": bson.M{"$lt": now},
		"End":   bson.M{"$gt": newStart},
	}, bson.M{
		"$set": bson.M{
			"Start": newStart,
		},
	}, &options.FindOneAndUpdateOptions{
		Sort: bson.M{"Number": 1},
	}).Decode(&rec); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &rec.Mail, nil
}
func insertMail(ctx context.Context, col *mongo.Collection, start time.Time, deadline time.Time, mail *Mail) error {
	if start.After(deadline) {
		return errors.New("start after deadline")
	}
	if deadline.Before(time.Now()) {
		return errors.New("past deadline")
	}
	if mail.Key == "" {
		return errors.New("empty mail key")
	}
	_, err := col.InsertOne(ctx, record{
		Mail:  *mail,
		Start: start.Unix(),
		End:   deadline.Unix(),
	})
	return err
}

func (m *Mailer) Enqueue(ctx context.Context, waitUntil, deadline time.Time, mail *Mail) error {
	return insertMail(ctx, m.Collection, waitUntil, deadline, mail)
}

func (m *Mailer) Start(parentCtx context.Context) error {
	if err := ensureIndexes(parentCtx, m.Collection); err != nil {
		return err
	}
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
		mail, err := pullMail(ctx, m.Collection, m.Delay)
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
			removeMail(m.Collection, mail.Key)
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
