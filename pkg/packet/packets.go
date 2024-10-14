package packet

import (
	"club.asynclab/asrp/pkg/logging"
)

var logger = logging.GetLogger()

const (
	NetPacketTypeHello NetPacketType = iota
	NetPacketTypeMessage
	NetPacketTypeProxyNegotiate
	NetPacketTypeProxyConfirm
	NetPacketTypeProxy
	NetPacketTypeNewProxyConnection
	NetPacketTypeHeart
	NetPacketTypeEnd
	NetPacketTypeUnknown
)

type PacketHello struct{}
type PacketMessage struct{}
type PacketProxyNegotiate struct {
	Name            string
	FrontendAddress string
}
type PacketProxyConfirm struct {
	Name    string
	Success bool
}
type PacketProxy struct {
	Uuid string
}
type PacketNewProxyConnection struct {
	Name string
	Uuid string
}
type PacketHeart struct {
	Uuid string
}
type PacketEnd struct{}
type PacketUnknown struct {
	Err error
}

func (p *PacketHello) Type() NetPacketType              { return NetPacketTypeHello }
func (p *PacketMessage) Type() NetPacketType            { return NetPacketTypeMessage }
func (p *PacketProxyNegotiate) Type() NetPacketType     { return NetPacketTypeProxyNegotiate }
func (p *PacketProxyConfirm) Type() NetPacketType       { return NetPacketTypeProxyConfirm }
func (p *PacketProxy) Type() NetPacketType              { return NetPacketTypeProxy }
func (p *PacketNewProxyConnection) Type() NetPacketType { return NetPacketTypeNewProxyConnection }
func (p *PacketHeart) Type() NetPacketType              { return NetPacketTypeHeart }
func (p *PacketEnd) Type() NetPacketType                { return NetPacketTypeEnd }
func (p *PacketUnknown) Type() NetPacketType            { return NetPacketTypeUnknown }

func init() {
	TypeMap[NetPacketTypeHello] = (*PacketHello)(nil)
	TypeMap[NetPacketTypeMessage] = (*PacketMessage)(nil)
	TypeMap[NetPacketTypeProxyNegotiate] = (*PacketProxyNegotiate)(nil)
	TypeMap[NetPacketTypeProxyConfirm] = (*PacketProxyConfirm)(nil)
	TypeMap[NetPacketTypeProxy] = (*PacketProxy)(nil)
	TypeMap[NetPacketTypeNewProxyConnection] = (*PacketNewProxyConnection)(nil)
	TypeMap[NetPacketTypeHeart] = (*PacketHeart)(nil)
	TypeMap[NetPacketTypeEnd] = (*PacketEnd)(nil)
	TypeMap[NetPacketTypeUnknown] = (*PacketUnknown)(nil)
}
