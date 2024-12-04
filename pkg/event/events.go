package event

import (
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/packet"
)

type EventReceivedPacket[T packet.IPacket] struct {
	Conn   *comm.Conn
	Packet T
}

func NewEventReceivedPacket[T packet.IPacket](conn *comm.Conn, packet T) *EventReceivedPacket[T] {
	return &EventReceivedPacket[T]{Conn: conn, Packet: packet}
}

type EventAcceptedFrontendConnection struct {
	Name string
	Conn *comm.Conn
}
