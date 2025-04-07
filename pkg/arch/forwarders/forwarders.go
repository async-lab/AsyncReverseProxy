package forwarders

import (
	"context"

	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/base/pattern"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
)

var logger = logging.GetLogger()

type Forwarder struct {
	Tip          string
	conn         *comm.Conn
	senderPacket *channel.SafeSender[packet.IPacket]
}

func NewForwarder(conn *comm.Conn) *Forwarder {
	forwarder := &Forwarder{
		Tip:          "[Common Forwarder]: ",
		conn:         conn,
		senderPacket: channel.NewSafeSenderWithSize[packet.IPacket](16),
	}

	forwarder.init()
	return forwarder
}

func (forwarder *Forwarder) HandlePacket(p packet.IPacket) bool {
	_, err := comm.SendPacket(forwarder.conn, p)
	if err != nil {
		if forwarder.GetCtx().Err() != nil {
			return false
		}
		logger.Debug(forwarder.Tip, "Error sending packet: ", err)
	}
	return true
}

func (forwarder *Forwarder) receivePacket() packet.IPacket {
	r, err := comm.ReceivePacket(forwarder.conn)
	if err != nil {
		if forwarder.GetCtx().Err() != nil {
			return nil
		}
		r = &packet.PacketUnknown{Err: err}
	}
	return r
}

func (forwarder *Forwarder) GetCtx() context.Context {
	return forwarder.conn.GetCtx()
}

func (forwarder *Forwarder) Close() error {
	return forwarder.conn.Close()
}

func (forwarder *Forwarder) IsClosed() bool {
	return forwarder.conn.IsClosed()
}

func (forwarder *Forwarder) GetChanSendPacket() <-chan packet.IPacket {
	return forwarder.senderPacket.GetChan()
}

// ---------------------------------------------------------------------

func (forwarder *Forwarder) init() {
	go func() {
		pattern.NewConfigSelectContextAndChannel[packet.IPacket]().
			WithCtx(forwarder.GetCtx()).
			WithGoroutine(func(ch chan packet.IPacket) {
				for {
					pkt := forwarder.receivePacket()
					if pkt == nil {
						return
					}
					ch <- pkt
				}
			}).
			WithChannelHandler(func(pkt packet.IPacket) { forwarder.senderPacket.Push(pkt) }).
			Run()
	}()
}
