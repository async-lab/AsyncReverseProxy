package event

import (
	"net"

	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
)

var logger = logging.GetLogger()

var Manager = NewEventManager()

type EventHello struct {
	Conn net.Conn
}

type EventProxyNegotiate struct {
	Conn   net.Conn
	Packet packet.PacketProxyNegotiate
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
