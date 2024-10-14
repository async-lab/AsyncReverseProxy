package program

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
	"club.asynclab/asrp/pkg/util"
)

var logger = logging.GetLogger()

var Program IProgram = nil

type ProgramMeta struct {
	Ctx      context.Context
	EventBus *event.EventBus
}

type IProgram interface {
	GetMeta() *ProgramMeta
}

func NewProgramMeta(ctx context.Context) *ProgramMeta {
	meta := &ProgramMeta{
		Ctx:      ctx,
		EventBus: event.NewEventBus(),
	}
	return meta
}

func ReceivePacket(conn net.Conn) (packet.IPacket, bool) {
	r, err := comm.ReceivePacket(conn)
	if err != nil {
		if Program.GetMeta().Ctx.Err() != nil {
			return nil, false
		}
		if util.IsConnectionClose(err) {
			logger.Error("Connection closed: ", err)
			return nil, false
		}
		r = &packet.PacketUnknown{Err: err}
	}
	return r, true
}

func SendPacket(conn net.Conn, p packet.IPacket) bool {
	_, error := comm.SendPacket(conn, p)
	if error != nil {
		logger.Error("Error sending packet: ", error)
		return false
	}
	return true
}

func EmitEvent(conn net.Conn, connCtx context.Context) {
	meta := Program.GetMeta()
	pattern.SelectContextAndChannel(
		meta.Ctx,
		make(chan struct{}),
		func() {},
		func(struct{}) bool { return false },
		func(ch chan struct{}) {
			for {
				r, ok := ReceivePacket(conn)
				if !ok {
					close(ch)
					comm.SendPacket(conn, &packet.PacketEnd{})
					return
				}

				switch r.Type() {
				case packet.NetPacketTypeHello:
					ok = event.Publish(meta.EventBus, event.EventHello{
						Conn: conn,
					})
				case packet.NetPacketTypeProxyNegotiate:
					ok = event.Publish(meta.EventBus, event.EventProxyNegotiate{
						Conn:    conn,
						ConnCtx: connCtx,
						Packet:  *r.(*packet.PacketProxyNegotiate),
					})
				case packet.NetPacketTypeProxy:
					ok = event.Publish(meta.EventBus, event.EventProxy{
						Conn:   conn,
						Packet: *r.(*packet.PacketProxy),
					})
				case packet.NetPacketTypeEnd:
					ok = event.Publish(meta.EventBus, event.EventEnd{})
				case packet.NetPacketTypeUnknown:
					ok = event.Publish(meta.EventBus, event.EventUnknown{Conn: conn})

				case packet.NetPacketTypeProxyConfirm:
					ok = event.Publish(meta.EventBus, event.EventProxyConfirm{
						Conn:   conn,
						Packet: *r.(*packet.PacketProxyConfirm),
					})
				case packet.NetPacketTypeNewProxyConnection:
					ok = event.Publish(meta.EventBus, event.EventNewProxyConnection{
						Conn:   conn,
						Packet: *r.(*packet.PacketNewProxyConnection),
					})
				}

				if !ok {
					close(ch)
					comm.SendPacket(conn, &packet.PacketEnd{})
					return
				}
			}
		})
}
