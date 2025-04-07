package network

import (
	"net"
	"time"
)

type ConnWithTimeout struct {
	net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewConnWithTimeout(conn net.Conn, readTimeout time.Duration, writeTimeout time.Duration) *ConnWithTimeout {
	return &ConnWithTimeout{
		Conn:         conn,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

func (c *ConnWithTimeout) Read(b []byte) (n int, err error) {
	defer c.Conn.SetReadDeadline(time.Time{})
	if c.readTimeout > 0 {
		c.Conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}
	return c.Conn.Read(b)
}

func (c *ConnWithTimeout) Write(b []byte) (n int, err error) {
	defer c.Conn.SetWriteDeadline(time.Time{})
	if c.writeTimeout > 0 {
		c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	return c.Conn.Write(b)
}
