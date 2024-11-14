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
	Closed    bool
}

func NewListenerWithParentCtx(parentCtx context.Context, listener net.Listener) *Listener {
	if parentCtx == nil {
		panic("parentCtx is nil")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	ret := &Listener{
		Listener:  listener,
		Ctx:       ctx,
		CtxCancel: cancel,
		Closed:    false,
	}
	go func() {
		defer ret.Close()
		switch listener := listener.(type) {
		case *Listener:
			select {
			case <-listener.Ctx.Done():
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

func NewListener(listener net.Listener) *Listener {
	return NewListenerWithParentCtx(context.Background(), listener)
}

func (l *Listener) Accept() (c net.Conn, err error) {
	c, err = l.Listener.Accept()
	if lang.IsNetClose(err) {
		l.Close()
	}
	return
}

func (l *Listener) Close() error {
	defer func() {
		l.CtxCancel()
		l.Closed = true
	}()
	return l.Listener.Close()
}
