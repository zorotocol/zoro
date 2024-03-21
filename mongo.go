package oracle

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ensureIndices(ctx context.Context, col *mongo.Collection) error {
	_true := true
	if _, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"Number": -1},
		Options: &options.IndexOptions{
			Unique:                  &_true,
			PartialFilterExpression: bson.M{"Number": bson.M{"$exists": true}},
		},
	}); err != nil {
		return err
	}
	if _, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"Logs.Token": 1},
		Options: &options.IndexOptions{
			Unique:                  &_true,
			PartialFilterExpression: bson.M{"Logs.Token": bson.M{"$exists": true}},
		},
	}); err != nil {
		return err
	}
	return nil
}
func insertBlock(ctx context.Context, col *mongo.Collection, block *Block) error {
	_, err := col.InsertOne(ctx, *block)
	return err
}
func cleanEmptyLogs(ctx context.Context, col *mongo.Collection, currentBlockNumber int64) error {
	_, err := col.DeleteMany(ctx, bson.M{
		"Number": bson.M{"$lt": currentBlockNumber},
		"Logs":   bson.M{"$exists": false},
	})
	return err
}
func getLastBlockNumber(ctx context.Context, col *mongo.Collection) (int64, error) {
	var block Block
	if err := col.FindOne(ctx, bson.M{}, &options.FindOneOptions{
		Projection: bson.M{"Number": 1},
		Sort:       bson.M{"Number": -1},
	}).Decode(&block); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = nil
		}
		return 0, err
	}
	return block.Number, nil
}
func resolveToken(ctx context.Context, col *mongo.Collection, token string) (*Log, error) {
	var block Block
	if err := col.FindOne(ctx, bson.M{"Logs.Token": token}, nil).Decode(&block); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &block.Logs[0], nil
}
