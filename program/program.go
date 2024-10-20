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

type MetaProgram struct {
	Ctx      context.Context
	EventBus *event.EventBus
}

func NewMetaProgram(ctx context.Context) *MetaProgram {
	meta := &MetaProgram{
		Ctx:      ctx,
		EventBus: event.NewEventBus(),
	}
	return meta
}

type IProgram interface {
	Run()
	ToMeta() *MetaProgram
	ReceivePacket(conn net.Conn) (packet.IPacket, bool)
	SendPacket(conn net.Conn, p packet.IPacket) bool
	EmitEventReceivePacket(conn net.Conn, connCtx context.Context)
}

func (meta *MetaProgram) ToMeta() *MetaProgram { return meta }

func (meta *MetaProgram) ReceivePacket(conn net.Conn) (packet.IPacket, bool) {
	r, err := comm.ReceivePacket(conn)
	if err != nil {
		if meta.Ctx.Err() != nil || util.IsNetClose(err) {
			return nil, false
		}
		r = &packet.PacketUnknown{Err: err}
	}
	return r, true
}

func (meta *MetaProgram) SendPacket(conn net.Conn, p packet.IPacket) bool {
	_, err := comm.SendPacket(conn, p)
	if err != nil {
		if meta.Ctx.Err() == nil {
			logger.Error("Error sending packet: ", err)
		}
		return false
	}
	return true
}

func (meta *MetaProgram) EmitEventReceivePacket(conn net.Conn, connCtx context.Context) {
	pattern.NewConfigSelectContextAndChannel[struct{}]().
		WithCtx(meta.Ctx).
		WithGoroutine(func(ch chan struct{}) {
			for {
				r, ok := meta.ReceivePacket(conn)
				if !ok {
					return
				}

				switch r := r.(type) {
				case *packet.PacketHello:
					ok = event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
				case *packet.PacketProxyNegotiationRequest:
					ok = event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
				case *packet.PacketProxyNegotiationResponse:
					ok = event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
				case *packet.PacketProxyData:
					ok = event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
				case *packet.PacketNewEndConnection:
					ok = event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
				case *packet.PacketEndConnectionClosed:
					ok = event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
				case *packet.PacketEnd:
					ok = event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
				case *packet.PacketUnknown:
					ok = event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
				}

				if !ok {
					return
				}
			}
		}).
		Run()
}
