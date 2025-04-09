package ushers

import (
	"context"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/arch/forwarders"
	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/base/pattern"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
)

var logger = logging.GetLogger()

type Usher[T comm.Protocol] struct {
	listener        *comm.Listener
	token           string
	senderForwarder *channel.SafeSender[*arch.ForwarderWithValues]
}

func NewUsher[T comm.Protocol](listener *comm.Listener, token string) *Usher[T] {
	usher := &Usher[T]{
		listener:        listener,
		token:           token,
		senderForwarder: channel.NewSafeSenderWithParentCtxAndSize[*arch.ForwarderWithValues](listener.Ctx, 16),
	}
	usher.init()
	return usher
}

func (usher *Usher[T]) GetChanSendForwarder() <-chan *arch.ForwarderWithValues {
	return usher.senderForwarder.GetChan()
}

func (usher *Usher[T]) GetCtx() context.Context {
	return usher.listener.Ctx
}

func (usher *Usher[T]) Close() error {
	return usher.listener.Close()
}

// ----------------------------------------------------------------------

func (usher *Usher[T]) handleConnection(conn *comm.Conn) {
	f := forwarders.NewForwarder(conn)
	p, ok := (<-f.GetChanSendPacket()).(*packet.PacketProxyNegotiationRequest)
	if !ok {
		f.HandlePacket(&packet.PacketEnd{})
		f.Close()
		logger.Debug("Invalid packet")
		return
	}

	sendResponse := func(success bool, reason string) {
		f.HandlePacket(&packet.PacketProxyNegotiationResponse{
			Success: success,
			Reason:  reason,
		})
	}

	// 检查token
	if p.Token != usher.token {
		sendResponse(false, "Invalid token")
		f.Close()
		return
	}

	sendResponse(true, "")

	if p.Weight <= 0 {
		p.Weight = 1
	}

	usher.senderForwarder.Push(&arch.ForwarderWithValues{
		IForwarder: f,
		InitPacket: p,
	})
}

// -------------------------------------------

func (usher *Usher[T]) routineRead() {
	pattern.NewConfigSelectContextAndChannel[*comm.Conn]().
		WithCtx(usher.GetCtx()).
		WithGoroutine(func(ch chan *comm.Conn) {
			for {
				conn, err := usher.listener.Accept()
				if err != nil {
					if usher.GetCtx().Err() != nil {
						return
					}
					logger.Error("Error accepting connection: ", err)
					continue
				}
				ch <- conn
			}
		}).
		WithChannelHandler(func(conn *comm.Conn) { go usher.handleConnection(conn) }).
		Run()
}

func (usher *Usher[T]) init() {
	go usher.routineRead()
}
