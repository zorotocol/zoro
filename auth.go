package oracle

import (
	"context"
	"errors"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/zorotocol/oracle/pkg/db"
	"time"
)

type Authenticator struct {
	DB    *db.DB
	Cache *expirable.LRU[string, time.Time]
}

func (a *Authenticator) get(ctx context.Context, token string) (time.Time, error) {
	deadline, ok := a.Cache.Get(token)
	if ok {
		return deadline, nil
	}
	deadline, err := a.DB.GetLogDeadlineByPasswordHash(ctx, token)
	if err != nil {
		return time.Time{}, err
	}
	a.Cache.Add(token, deadline)
	return deadline, err
}
func (a *Authenticator) Authenticate(ctx context.Context, token string) (deadline time.Time, err error) {
	deadline, err = a.get(ctx, token)
	if err != nil {
		return deadline, err
	}
	now := time.Now()
	if deadline.Before(now) {
		return deadline, errors.New("expired")
	}
	return deadline, nil
}
