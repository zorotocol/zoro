package auth

import (
	"context"
	"errors"
	"github.com/zorotocol/zoro/pkg/db"
	"io"
	"net/http"
	"strconv"
	"time"
)

var _ http.Handler = &Authenticator{}

type Authenticator struct {
	DB *db.DB
}

func (a *Authenticator) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.ContentLength != 56 {
		http.Error(writer, "invalid content-length", http.StatusBadRequest)
		return
	}
	hash := make([]byte, 56)
	if _, err := io.ReadFull(request.Body, hash); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	deadline, err := a.Authenticate(request.Context(), string(hash))
	if err != nil {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
		return
	}
	_, _ = writer.Write([]byte(strconv.FormatInt(deadline.Unix(), 10)))
}

func (a *Authenticator) Authenticate(ctx context.Context, hash string) (deadline time.Time, err error) {
	if db.ValidatePasswordHash(hash) {
		return time.Time{}, errors.New("invalid hash")
	}
	deadline, err = a.DB.GetLogDeadlineByPasswordHash(ctx, hash)
	if err != nil {
		return time.Time{}, err
	}
	now := time.Now()
	if deadline.Before(now) {
		return deadline, errors.New("expired")
	}
	return deadline, nil
}
