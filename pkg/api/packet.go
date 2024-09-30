package api

import "net"

type PacketType int

const (
	Hello PacketType = iota
	Message
)

type IPacketData interface {
	cts(conn *net.Conn)
	stc(conn *net.Conn)
}

type Packet struct {
	Type PacketType
	Data IPacketData
}
