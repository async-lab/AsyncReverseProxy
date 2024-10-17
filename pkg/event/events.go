package event

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/packet"
)

type EventReceivedPacket[T packet.IPacket] struct {
	Conn    net.Conn
	ConnCtx context.Context
	Packet  T
}

func NewEventReceivedPacket[T packet.IPacket](conn net.Conn, connCtx context.Context, packet T) *EventReceivedPacket[T] {
	return &EventReceivedPacket[T]{Conn: conn, ConnCtx: connCtx, Packet: packet}
}

type EventPacketProxyDataQueue struct {
	Packet *packet.PacketProxyData
}
