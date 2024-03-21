package oracle

import (
	"context"
	"errors"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type cacheEntry struct {
	deadline     time.Time
	blockedUntil time.Time
}
type Authenticator struct {
	Collection *mongo.Collection
	Cache      *expirable.LRU[string, cacheEntry]
}

func (a *Authenticator) get(ctx context.Context, token string) (_ *cacheEntry, _ error) {
	v, ok := a.Cache.Get(token)
	if ok {
		return &v, nil
	}
	log, err := resolveToken(ctx, a.Collection, token)
	if err != nil {
		return nil, err
	}
	ent := cacheEntry{}
	if log != nil {
		ent.blockedUntil = log.BlockedUntil
		ent.deadline = log.Deadline
	}
	a.Cache.Add(token, ent)
	return &ent, nil
}
func (a *Authenticator) Authenticate(ctx context.Context, token string) (deadline time.Time, err error) {
	ent, err := a.get(ctx, token)
	if err != nil {
		return time.Time{}, err
	}
	if ent.blockedUntil.After(time.Now()) {
		return ent.deadline, errors.New("temporary blocked")
	}
	return ent.deadline, nil
}
