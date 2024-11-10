package program

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/base/pattern"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
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
	pattern.NewConfigSelectContextAndChannel[packet.IPacket]().
		WithCtx(meta.Ctx).
		WithGoroutine(func(ch chan packet.IPacket) {
			for {
				r, ok := meta.ReceivePacket(conn)
				if !ok {
					return
				}
				ch <- r
			}
		}).
		WithChannelHandlerWithInterruption(func(r packet.IPacket) bool {
			switch r := r.(type) {
			case *packet.PacketHello:
				return event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
			case *packet.PacketProxyNegotiationRequest:
				return event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
			case *packet.PacketProxyNegotiationResponse:
				return event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
			case *packet.PacketProxyData:
				return event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
			case *packet.PacketNewEndSideConnection:
				return event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
			case *packet.PacketEndSideConnectionClosed:
				return event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
			case *packet.PacketEnd:
				return event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
			case *packet.PacketUnknown:
				return event.Publish(meta.EventBus, event.NewEventReceivedPacket(conn, connCtx, r))
			default:
				logger.Error("Unknown packet type: ", r)
				return false
			}
		}).
		Run()
}
