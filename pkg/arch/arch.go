package arch

import (
	"context"
	"io"

	"club.asynclab/asrp/pkg/packet"
)

type IElementWithCtx interface {
	io.Closer
	GetCtx() context.Context
}

type IForwarder interface {
	IElementWithCtx
	GetChanSendPacket() <-chan packet.IPacket
	HandlePacket(pkt packet.IPacket) bool
	IsClosed() bool
}

type IDispatcher interface {
	IElementWithCtx
	GetChanSendPacket() <-chan packet.IPacket
	HandlePacket(pkt packet.IPacket) bool
	AddForwarder(fwv *ForwarderWithValues) (uuid string)
	RemoveForwarder(uuid string)
	Len() int
}

type IConnector interface {
	IElementWithCtx
	GetChanSendForwarder() <-chan IForwarder
}

type IUsher interface {
	IElementWithCtx
	GetChanSendForwarder() <-chan *ForwarderWithValues
}

type IDialer interface {
	IElementWithCtx
	GetChanSendPacket() <-chan packet.IPacket
	HandlePacket(pkt packet.IPacket) bool
}

type IAcceptor interface {
	IElementWithCtx
	GetChanSendPacket() <-chan packet.IPacket
	HandlePacket(pkt packet.IPacket) bool
}

type ForwarderWithValues struct {
	IForwarder
	InitPacket *packet.PacketProxyNegotiationRequest
}
