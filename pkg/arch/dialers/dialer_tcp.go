package dialers

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/comm"
)

type DialerImplTCP struct{}

var implTCP = &DialerImplTCP{}

func NewDialerTCP(parentCtx context.Context, addr string) (*Dialer[comm.TCP], error) {
	dialer := NewDialer[comm.TCP](parentCtx, implTCP, addr)
	return dialer, nil
}

func (impl *DialerImplTCP) Dial(ctx context.Context, addr string) (*comm.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return comm.NewConnWithParentCtx(ctx, conn), nil
}
