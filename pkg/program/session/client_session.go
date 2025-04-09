package session

import (
	"context"
	"fmt"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/arch/dialers"
	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/config"
)

type ClientSession struct {
	Ctx       context.Context
	CtxCancel context.CancelFunc
	Forwarder arch.IForwarder
	Dialer    arch.IDialer
}

func NewClientSession(parentCtx context.Context, forwarder arch.IForwarder, proxyConfig config.ConfigItemProxy) (*ClientSession, error) {
	ctx, cancel := context.WithCancel(parentCtx)
	var dialer arch.IDialer
	var err error
	switch proxyConfig.Proto {
	case "tcp":
		dialer, err = dialers.NewDialerTCP(ctx, proxyConfig.Backend)
	case "udp":
		dialer, err = dialers.NewDialerUDP(ctx, proxyConfig.Backend)
	default:
		cancel()
		return nil, fmt.Errorf("unsupported protocol: %T", proxyConfig.Proto)
	}
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
