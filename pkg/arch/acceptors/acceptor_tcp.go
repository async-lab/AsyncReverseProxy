package acceptors

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/comm"
)

func NewAcceptorTCP(parentCtx context.Context, addr string) (*Acceptor[comm.TCP], error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	acceptor := NewAcceptor[comm.TCP](comm.NewListenerWithParentCtx(parentCtx, listener))
	return acceptor, nil
}
