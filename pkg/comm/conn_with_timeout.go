package comm

import (
	"time"
)

type ConnWithTimeout struct {
	*Conn
	timeout time.Duration
}

func NewConnWithTimeout(conn *Conn, timeout time.Duration) *ConnWithTimeout {
	return &ConnWithTimeout{
		Conn:    conn,
		timeout: timeout,
	}
}

func (c *ConnWithTimeout) Read(b []byte) (int, error) {
	if c.timeout > 0 {
		if err := c.Conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
			return 0, err
		}
	}
	return c.Conn.Read(b)
}

func (c *ConnWithTimeout) Write(b []byte) (int, error) {
	if c.timeout > 0 {
		if err := c.Conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
			return 0, err
		}
	}
	return c.Conn.Write(b)
}

func (c *ConnWithTimeout) SetReadDeadline(t time.Time) error {
	panic("Deadline not supported")
}

func (c *ConnWithTimeout) SetWriteDeadline(t time.Time) error {
	panic("Deadline not supported")
}
