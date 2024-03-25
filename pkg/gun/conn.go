package gun

import (
	"bytes"
	"context"
	"github.com/zorotocol/zoro/pkg/proto"
	"io"
	"os"
)

var _ io.ReadWriteCloser = &conn{}

type conn struct {
	conn   proto.Gun_TunServer
	buf    bytes.Buffer
	ctx    context.Context
	cancel context.CancelCauseFunc
}

func newConn(c proto.Gun_TunServer) *conn {
	ctx, cancel := context.WithCancelCause(c.Context())
	return &conn{
		conn:   c,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *conn) Read(p []byte) (int, error) {
	if c.buf.Len() > 0 {
		return c.buf.Read(p)
	}
	chunk, err := c.conn.Recv()
	if err != nil {
		return 0, err
	}
	n := copy(p, chunk.Data)
	if len(chunk.Data) > len(p) {
		c.buf.Write(chunk.Data[n:])
	}
	return n, nil
}

func (c *conn) Write(p []byte) (n int, err error) {
	if err := c.conn.Send(&proto.Hunk{Data: p}); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *conn) Close() error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	c.cancel(os.ErrClosed)
	return nil
}
