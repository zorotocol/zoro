package rl

import (
	"goftp.io/server/v2/ratelimit"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

var _ io.ReadWriteCloser = &stream{}

type stream struct {
	user   string
	rl     *RateLimiter
	ent    *entry
	closed atomic.Bool
	conn   io.ReadWriteCloser
	read   io.Reader
	write  io.Writer
}

func (s *stream) Read(p []byte) (n int, err error) {
	return s.read.Read(p)
}

func (s *stream) Write(p []byte) (n int, err error) {
	return s.write.Write(p)
}

func (s *stream) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return os.ErrClosed
	}
	refs := s.ent.rc.Add(-1)
	if refs == 0 {
		s.rl.mutex.Lock()
		delete(s.rl.users, s.user)
		s.rl.mutex.Unlock()
	}
	return s.conn.Close()
}

type entry struct {
	limit *ratelimit.Limiter
	rc    atomic.Int64
}
type RateLimiter struct {
	mutex sync.Mutex
	users map[string]*entry
	limit int64
}

func New(limit int64) *RateLimiter {
	rl := &RateLimiter{
		users: make(map[string]*entry),
		limit: limit,
	}
	return rl
}

func (rl *RateLimiter) GetLimiter(user string, conn io.ReadWriteCloser) io.ReadWriteCloser {
	rl.mutex.Lock()
	ent, _ := rl.users[user]
	if ent == nil {
		ent = &entry{limit: ratelimit.New(rl.limit)}
		rl.users[user] = ent
	}
	ent.rc.Add(1)
	rl.mutex.Unlock()
	reader := ratelimit.Reader(conn, ent.limit)
	writer := ratelimit.Writer(conn, ent.limit)
	return &stream{
		user:  user,
		ent:   ent,
		rl:    rl,
		conn:  conn,
		read:  reader,
		write: writer,
	}
}
