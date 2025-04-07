package session

import (
	"context"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/arch/acceptors"
	"club.asynclab/asrp/pkg/arch/dispatchers"
	"club.asynclab/asrp/pkg/base/channel"
)

type ServerSession struct {
	ctx        context.Context
	ctxCancel  context.CancelFunc
	dispatcher arch.IDispatcher
	acceptor   arch.IAcceptor
}

func NewServerSession(parentCtx context.Context, frontendAddr string) (*ServerSession, error) {
	ctx, cancel := context.WithCancel(parentCtx)
	acceptor, err := acceptors.NewAcceptorTCP(ctx, frontendAddr)
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
