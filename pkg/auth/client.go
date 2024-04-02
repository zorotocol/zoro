package auth

import (
	"context"
	"errors"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/zorotocol/zoro/pkg/db"
	"github.com/zorotocol/zoro/pkg/misc"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	URL        string
	Cache      *expirable.LRU[string, time.Time]
}

func (c *Client) req(ctx context.Context, token string) (time.Time, error) {
	req := misc.Must(http.NewRequestWithContext(ctx, http.MethodPost, c.URL, strings.NewReader(token)))
	req.ContentLength = int64(len(token))
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	if resp.StatusCode != 200 {
		return time.Time{}, nil
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 100))
	if err != nil {
		return time.Time{}, err
	}
	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(n, 0), nil
}
func (c *Client) Authenticate(ctx context.Context, hash string) (deadline time.Time, err error) {
	if !db.ValidatePasswordHash(hash) {
		return time.Time{}, errors.New("invalid hash")
	}
	deadline, ok := c.Cache.Get(hash)
	if !ok {
		deadline, err = c.req(ctx, hash)
		if err != nil {
			return time.Time{}, err
		}
		c.Cache.Add(hash, deadline)
	}
	if deadline.Before(time.Now().Add(time.Minute)) {
		return deadline, errors.New("expired")
	}
	return deadline, nil
}
