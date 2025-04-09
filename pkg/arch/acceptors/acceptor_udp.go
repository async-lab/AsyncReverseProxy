package acceptors

import (
	"context"

	"club.asynclab/asrp/pkg/base/network"
	"club.asynclab/asrp/pkg/comm"
)

func NewAcceptorUDP(parentCtx context.Context, addr string) (*Acceptor[comm.UDP], error) {
	listener, err := network.NewUDPListener(addr)
	if err != nil {
		return nil, err
	}
	acceptor := NewAcceptor[comm.UDP](comm.NewListenerWithParentCtx(parentCtx, listener))
	return acceptor, nil
}
