package session

import (
	"context"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/arch/dialers"
	"club.asynclab/asrp/pkg/base/channel"
)

type ClientSession struct {
	Ctx       context.Context
	CtxCancel context.CancelFunc
	Forwarder arch.IForwarder
	Dialer    arch.IDialer
}

func NewClientSession(parentCtx context.Context, forwarder arch.IForwarder, backendAddr string) (*ClientSession, error) {
	ctx, cancel := context.WithCancel(parentCtx)
	dialer, err := dialers.NewDialerTCP(ctx, backendAddr)
	if err != nil {
		cancel()
		return nil, err
	}

	s := &ClientSession{
		Ctx:       ctx,
		CtxCancel: cancel,
		Forwarder: forwarder,
		Dialer:    dialer,
	}

	go channel.ConsumeWithCtx(s.Ctx, s.Dialer.GetChanSendPacket(), s.Forwarder.HandlePacket)
	go channel.ConsumeWithCtx(s.Ctx, s.Forwarder.GetChanSendPacket(), s.Dialer.HandlePacket)

	go func() {
		<-forwarder.GetCtx().Done()
		cancel()
	}()

	return s, nil
}
