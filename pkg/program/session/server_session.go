package session

import (
	"context"
	"fmt"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/arch/acceptors"
	"club.asynclab/asrp/pkg/arch/dispatchers"
	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/packet"
)

type ServerSession struct {
	ctx        context.Context
	ctxCancel  context.CancelFunc
	dispatcher arch.IDispatcher
	acceptor   arch.IAcceptor
}

func NewServerSession(parentCtx context.Context, initPacket *packet.PacketProxyNegotiationRequest) (*ServerSession, error) {
	ctx, cancel := context.WithCancel(parentCtx)
	var acceptor arch.IAcceptor
	var err error
	switch initPacket.Proto {
	case "tcp":
		acceptor, err = acceptors.NewAcceptorTCP(ctx, initPacket.FrontendAddr)
	case "udp":
		acceptor, err = acceptors.NewAcceptorUDP(ctx, initPacket.FrontendAddr)
	default:
		cancel()
		return nil, fmt.Errorf("unsupported protocol: %T", initPacket.Proto)
	}
	if err != nil {
		cancel()
		return nil, err
	}
	s := &ServerSession{
		ctx:        ctx,
		ctxCancel:  cancel,
		dispatcher: dispatchers.NewDispatcher(ctx),
		acceptor:   acceptor,
	}

	go channel.ConsumeWithCtx(s.ctx, s.acceptor.GetChanSendPacket(), s.dispatcher.HandlePacket)
	go channel.ConsumeWithCtx(s.ctx, s.dispatcher.GetChanSendPacket(), s.acceptor.HandlePacket)

	return s, nil
}

func (s *ServerSession) Close() {
	s.ctxCancel()
}

func (s *ServerSession) GetCtx() context.Context {
	return s.ctx
}

func (s *ServerSession) GetDispatcher() arch.IDispatcher {
	return s.dispatcher
}

func (s *ServerSession) GetAcceptor() arch.IAcceptor {
	return s.acceptor
}
