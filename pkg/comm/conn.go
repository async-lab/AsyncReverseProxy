package comm

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/base/lang"
)

type Conn struct {
	net.Conn
	Ctx       context.Context
	ctxCancel context.CancelFunc
	closed    bool
}

func NewConnWithParentCtx(parentCtx context.Context, conn net.Conn) *Conn {
	if parentCtx == nil {
		panic("parentCtx is nil")
	}

	if conn == nil {
		panic("conn is nil")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	ret := &Conn{
		Conn:      conn,
		Ctx:       ctx,
		ctxCancel: cancel,
		closed:    false,
	}
	go func() {
		defer ret.Close()
		select {
		case <-ret.Ctx.Done():
			break
		case <-ctx.Done():
			break
		}
	}()
	return ret
}

func NewConn(conn net.Conn) *Conn {
	return NewConnWithParentCtx(context.Background(), conn)
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if lang.IsNetClose(err) {
		c.Close()
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if lang.IsNetClose(err) {
		c.Close()
	}
	return
}

func (c *Conn) Close() error {
	defer func() {
		c.ctxCancel()
		c.closed = true
	}()
	return c.Conn.Close()
}

func (c *Conn) isClosed() bool {
	return c.closed
}
