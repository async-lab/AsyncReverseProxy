package comm

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/base/lang"
)

type Listener struct {
	net.Listener
	Ctx       context.Context
	CtxCancel context.CancelFunc
	closed    bool
}

func NewListenerWithCtx(ctx context.Context, cancel context.CancelFunc, listener net.Listener) *Listener {
	if ctx == nil {
		panic("ctx is nil")
	}
	if cancel == nil {
		panic("cancel is nil")
	}
	if listener == nil {
		panic("listener is nil")
	}

	ret := &Listener{
		Listener:  listener,
		Ctx:       ctx,
		CtxCancel: cancel,
		closed:    false,
	}
	go func() {
		defer ret.Close()
		<-ctx.Done()
	}()
	return ret
}

func NewListenerWithParentCtx(parentCtx context.Context, listener net.Listener) *Listener {
	if parentCtx == nil {
		panic("parentCtx is nil")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	return NewListenerWithCtx(ctx, cancel, listener)
}

func NewListener(listener net.Listener) *Listener {
	return NewListenerWithParentCtx(context.Background(), listener)
}

func (l *Listener) Accept() (*Conn, error) {
	c, err := l.Listener.Accept()
	if lang.IsNetLost(err) {
		l.Close()
	}
	if c == nil {
		return nil, err
	}
	return NewConnWithParentCtx(l.Ctx, c), err
}

func (l *Listener) Close() error {
	l.CtxCancel()
	l.closed = true
	return l.Listener.Close()
}

func (c *Listener) IsClosed() bool {
	return c.closed
}
