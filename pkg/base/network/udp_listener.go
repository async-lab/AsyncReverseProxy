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
	"club.asynclab/asrp/pkg/base/lang"
	"club.asynclab/asrp/pkg/logging"
)

var logger = logging.GetLogger()

var udpListenerBufSize uint32 = 10 * 1024
var udpListenerBufPool = concurrent.NewPool(func() *[]byte {
	buf := make([]byte, udpListenerBufSize)
	return &buf
})

type udpListenerVirtualConn struct {
	ul            *UDPListener
	remoteAddr    net.Addr
	ctx           context.Context
	ctxCancel     context.CancelFunc
	receiver      *channel.SafeSender[[]byte]
	readDeadline  *concurrent.Atomic[time.Time]
	writeDeadline *concurrent.Atomic[time.Time]
	idleTimer     *time.Timer
}

func NewUDPListenerVirtualConn(ul *UDPListener, remoteAddr net.Addr) *udpListenerVirtualConn {
	ctx, cancel := context.WithCancel(context.Background())
	return &udpListenerVirtualConn{
		ul:            ul,
		remoteAddr:    remoteAddr,
		ctx:           ctx,
		ctxCancel:     cancel,
		receiver:      channel.NewSafeSenderWithParentCtxAndSize[[]byte](ctx, 16),
		readDeadline:  concurrent.NewAtomicWithValue(time.Time{}),
		writeDeadline: concurrent.NewAtomicWithValue(time.Time{}),
		idleTimer:     time.AfterFunc(ul.timeout, cancel),
	}
}

func (c *udpListenerVirtualConn) Read(b []byte) (int, error) {
	if c.ctx.Err() != nil {
		return 0, net.ErrClosed
	}

	ddl := c.readDeadline.Load()
	after := make(<-chan time.Time)
	if !ddl.IsZero() {
		if time.Now().After(ddl) {
			return 0, &net.OpError{Op: "read", Net: "udp", Source: c.LocalAddr(), Addr: c.RemoteAddr(), Err: os.ErrDeadlineExceeded}
		}
		after = time.After(time.Until(ddl))
	}

	select {
	case data, ok := <-c.receiver.GetChan():
		if !ok {
			return 0, net.ErrClosed
		}
		n := copy(b, data)
		return n, nil
	case <-after:
		return 0, &net.OpError{Op: "read", Net: "udp", Source: c.LocalAddr(), Addr: c.RemoteAddr(), Err: os.ErrDeadlineExceeded}
	}
}

func (c *udpListenerVirtualConn) Write(b []byte) (int, error) {
	if c.ctx.Err() != nil {
		return 0, net.ErrClosed
	}
	ddl := c.writeDeadline.Load()
	if !ddl.IsZero() && time.Now().After(ddl) {
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
	c.readDeadline.Store(t)
	return nil
}

func (c *udpListenerVirtualConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline.Store(t)
	return nil
}

func (c *udpListenerVirtualConn) Close() error {
	c.ctxCancel()
	c.idleTimer.Stop()
	return nil
}

func (c *udpListenerVirtualConn) TryPush(data []byte) bool {
	if ok := c.idleTimer.Reset(c.ul.timeout); !ok {
		c.Close()
		return false
	}
	return c.receiver.TryPush(data)
}

type UDPListener struct {
	conn    *net.UDPConn
	m       *concurrent.ConcurrentMap[string, *udpListenerVirtualConn]
	ch      chan *udpListenerVirtualConn
	wg      *sync.WaitGroup
	closed  *atomic.Bool
	once    *sync.Once
	timeout time.Duration
}

func NewUDPListenerWithTimeout(addr string, timeout time.Duration) (*UDPListener, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return &UDPListener{
		conn:    conn,
		ch:      make(chan *udpListenerVirtualConn, 16),
		wg:      &sync.WaitGroup{},
		closed:  &atomic.Bool{},
		m:       concurrent.NewSyncMap[string, *udpListenerVirtualConn](),
		once:    &sync.Once{},
		timeout: timeout,
	}, nil
}

func NewUDPListener(addr string) (*UDPListener, error) {
	return NewUDPListenerWithTimeout(addr, 30*time.Second)
}

func (ul *UDPListener) Addr() net.Addr { return ul.conn.LocalAddr() }

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
	bufPtr := udpListenerBufPool.Get()
	defer udpListenerBufPool.Put(bufPtr)
	buf := *bufPtr

	for {
		n, addr, err := ul.conn.ReadFromUDP(buf)
		if lang.IsNetLost(err) {
			ul.Close()
			return
		}
		if err != nil {
			continue
		}

		var c *udpListenerVirtualConn

		ul.m.Compute(func(v *concurrent.ConcurrentMap[string, *udpListenerVirtualConn]) {
			_c, ok := v.Load(addr.String())
			if !ok {
				if ul.closed.Load() {
					return
				}
				ul.wg.Add(1)
				_c = NewUDPListenerVirtualConn(ul, addr)
				v.Store(addr.String(), _c)
				go func() {
					select {
					case ul.ch <- _c:
					default:
						_c.Close()
					}
					<-_c.ctx.Done()
					ul.wg.Done()
					ul.m.Delete(_c.remoteAddr.String())
				}()
			}
			c = _c
		})

		if c == nil {
			continue
		}

		data := make([]byte, n)
		copy(data, buf[:n])
		ok := c.receiver.TryPush(data)
		if !ok {
			logger.Debug("Error pushing data to receiver")
		}
	}
}

func (ul *UDPListener) Accept() (net.Conn, error) {
	if ul.closed.Load() {
		return nil, net.ErrClosed
	}

	go ul.once.Do(ul.dispatch)

	if c, ok := <-ul.ch; ok {
		return c, nil
	}

	return nil, net.ErrClosed
}
