package acceptors

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/comm"
)

func NewAcceptorUDP(parentCtx context.Context, addr string) (*Acceptor[comm.UDP], error) {
	listener, err := net.Listen("udp", addr) // TODO
	if err != nil {
		return nil, err
	}
	acceptor := NewAcceptor[comm.UDP](comm.NewListenerWithParentCtx(parentCtx, listener))
	return acceptor, nil
}
