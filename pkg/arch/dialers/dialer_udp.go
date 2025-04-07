package dialers

import (
	"context"
	"net"
	"time"

	"club.asynclab/asrp/pkg/base/network"
	"club.asynclab/asrp/pkg/comm"
)

type DialerImplUDP struct{}

var implUDP = &DialerImplUDP{}

func NewDialerUDP(parentCtx context.Context, addr string) (*Dialer[comm.UDP], error) {
	dialer := NewDialer[comm.UDP](parentCtx, implUDP, addr)
	return dialer, nil
}

func (impl *DialerImplUDP) Dial(ctx context.Context, addr string) (*comm.Conn, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, err
	}
	return comm.NewConnWithParentCtx(ctx, network.NewConnWithTimeout(conn, 60*time.Second, 0)), nil
}
