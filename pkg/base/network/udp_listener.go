package network

import (
	"context"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/base/goroutine"
)

var bufSize uint32 = 1024

type udpListenerVirtualConn struct {
	ul            *UDPListener
	remoteAddr    net.Addr
	ctx           context.Context
	ctxCancel     context.CancelFunc
	receiver      *channel.SafeSender[[]byte]
	readDeadline  time.Time
	writeDeadline time.Time
}

func NewUDPListenerVirtualConn(ul *UDPListener, remoteAddr net.Addr) *udpListenerVirtualConn {
	ctx, cancel := context.WithCancel(context.Background())
	return &udpListenerVirtualConn{
		ul:            ul,
		remoteAddr:    remoteAddr,
		ctx:           ctx,
		ctxCancel:     cancel,
		receiver:      channel.NewSafeSenderWithParentCtxAndSize[[]byte](ctx, 16),
		readDeadline:  time.Time{},
		writeDeadline: time.Time{},
	}
}

func (c *udpListenerVirtualConn) Read(b []byte) (int, error) {
	if c.ctx.Err() != nil {
		return 0, net.ErrClosed
	}
	if !c.readDeadline.IsZero() && time.Now().After(c.readDeadline) {
		return 0, &net.OpError{Op: "read", Net: "udp", Source: c.LocalAddr(), Addr: c.RemoteAddr(), Err: os.ErrDeadlineExceeded}
	}
	select {
	case data, ok := <-c.receiver.GetChan():
		if !ok {
			return 0, net.ErrClosed
		}
		n := copy(b, data)
		return n, nil
	case <-time.After(time.Until(c.readDeadline)):
		return 0, &net.OpError{Op: "read", Net: "udp", Source: c.LocalAddr(), Addr: c.RemoteAddr(), Err: os.ErrDeadlineExceeded}
	}
}

func (c *udpListenerVirtualConn) Write(b []byte) (int, error) {
	if c.ctx.Err() != nil {
		return 0, net.ErrClosed
	}
	if !c.writeDeadline.IsZero() && time.Now().After(c.writeDeadline) {
		return 0, &net.OpError{Op: "write", Net: "udp", Source: c.LocalAddr(), Addr: c.RemoteAddr(), Err: os.ErrDeadlineExceeded}
	}
	n, err := c.ul.conn.WriteTo(b, c.remoteAddr)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c *udpListenerVirtualConn) LocalAddr() net.Addr {
	return c.ul.conn.LocalAddr()
}

func (c *udpListenerVirtualConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *udpListenerVirtualConn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *udpListenerVirtualConn) SetReadDeadline(t time.Time) error {
	c.readDeadline = t
	return nil
}

func (c *udpListenerVirtualConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

func (c *udpListenerVirtualConn) Close() error {
	c.ctxCancel()
	c.ul.wg.Done()
	c.ul.m.Delete(c.remoteAddr.String())
	return nil
}

type UDPListener struct {
	conn   *net.UDPConn
	m      *concurrent.ConcurrentMap[string, *udpListenerVirtualConn]
	ch     chan *udpListenerVirtualConn
	wg     *sync.WaitGroup
	closed *atomic.Bool
	once   *sync.Once
}

func NewUDPListener(addr string) (*UDPListener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return &UDPListener{
		conn:   conn,
		ch:     make(chan *udpListenerVirtualConn, 16),
		wg:     &sync.WaitGroup{},
		closed: &atomic.Bool{},
		m:      concurrent.NewSyncMap[string, *udpListenerVirtualConn](),
		once:   &sync.Once{},
	}, nil
}

func (ul *UDPListener) Close() error {
	old := ul.closed.Swap(true)
	if !old {
		go func() {
			ul.wg.Wait()
			ul.conn.Close()
			close(ul.ch)
		}()
	}
	return nil
}

func (ul *UDPListener) dispatch() {
	buf := make([]byte, bufSize)
	for {
		n, addr, err := ul.conn.ReadFromUDP(buf)
		if err != nil || ul.closed.Load() {
			return
		}
		var c *udpListenerVirtualConn

		ul.m.Compute(func(v *concurrent.ConcurrentMap[string, *udpListenerVirtualConn]) {
			_c, ok := v.Load(addr.String())
			if !ok {
				ul.wg.Add(1)
				_c := NewUDPListenerVirtualConn(ul, addr)
				v.Store(addr.String(), _c)
				go func() {
					select {
					case ul.ch <- _c:
					default:
						_c.Close()
					}
				}()
			}
			c = _c
		})

		if c == nil {
			panic("c is nil")
		}

		data := make([]byte, n)
		copy(data, buf[:n])
		c.receiver.TryPush(data)
	}
}

func (ul *UDPListener) Accept() (net.Conn, error) {
	if ul.closed.Load() {
		return nil, net.ErrClosed
	}

	ul.once.Do(func() { goroutine.MultiGo(4, ul.dispatch) })
	if c, ok := <-ul.ch; ok {
		return c, nil
	}

	return nil, net.ErrClosed
}
