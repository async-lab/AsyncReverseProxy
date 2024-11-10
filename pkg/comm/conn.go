package comm

import (
	"context"
	"net"
	"time"

	"club.asynclab/asrp/pkg/base/lang"
)

type Conn struct {
	Conn      net.Conn
	Ctx       context.Context
	CtxCancel context.CancelFunc
	Closed    bool
}

func NewConn(conn net.Conn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	return &Conn{Conn: conn, Ctx: ctx, CtxCancel: cancel}
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

func (c *Conn) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.Conn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.Conn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.Conn.SetWriteDeadline(t)
}
