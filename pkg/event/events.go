package event

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/packet"
)

type EventHello struct {
	Conn net.Conn
}

type EventProxyNegotiate struct {
	Conn    net.Conn
	ConnCtx context.Context
	Packet  packet.PacketProxyNegotiate
}

type EventProxyConfirm struct {
	Conn   net.Conn
	Packet packet.PacketProxyConfirm
}

type EventProxy struct {
	Conn   net.Conn
	Packet packet.PacketProxy
}

type EventNewProxyConnection struct {
	Conn   net.Conn
	Packet packet.PacketNewProxyConnection
}

type EventUnknown struct {
	Conn net.Conn
}

type EventEnd struct{}
