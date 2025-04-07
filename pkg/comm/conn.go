package comm

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/base/lang"
)

type Conn struct {
	net.Conn
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func NewConnWithCtx(ctx context.Context, cancel context.CancelFunc, conn net.Conn) *Conn {
	if ctx == nil {
		panic("ctx is nil")
	}
	if cancel == nil {
		panic("cancel is nil")
	}
	if conn == nil {
		panic("conn is nil")
	}

	ret := &Conn{
		Conn:      conn,
		ctx:       ctx,
		ctxCancel: cancel,
	}
	go func() {
		defer ret.Close()
		<-ctx.Done()
	}()
	return ret
}

func NewConnWithParentCtx(parentCtx context.Context, conn net.Conn) *Conn {
	if parentCtx == nil {
		panic("parentCtx is nil")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	return NewConnWithCtx(ctx, cancel, conn)
}

func NewConn(conn net.Conn) *Conn {
	return NewConnWithParentCtx(context.Background(), conn)
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if lang.IsNetLost(err) {
		c.Close()
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if lang.IsNetLost(err) {
		c.Close()
	}
	return
}

func (c *Conn) GetCtx() context.Context {
	return c.ctx
}

func (c *Conn) Close() error {
	c.ctxCancel()
	return c.Conn.Close()
}

func (c *Conn) IsClosed() bool {
	return c.ctx.Err() != nil
}
