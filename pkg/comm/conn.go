package comm

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/base/lang"
)

type Conn struct {
	net.Conn
	Ctx       context.Context
	CtxCancel context.CancelFunc
	Closed    bool
}

func NewConnWithParentCtx(parentCtx context.Context, conn net.Conn) *Conn {
	if parentCtx == nil {
		panic("parentCtx is nil")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	ret := &Conn{
		Conn:      conn,
		Ctx:       ctx,
		CtxCancel: cancel,
		Closed:    false,
	}
	go func() {
		defer ret.Close()
		switch conn := conn.(type) {
		case *Conn:
			select {
			case <-conn.Ctx.Done():
				break
			case <-ctx.Done():
				break
			}
		default:
			<-ctx.Done()
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
		c.CtxCancel()
		c.Closed = true
	}()
	return c.Conn.Close()
}
