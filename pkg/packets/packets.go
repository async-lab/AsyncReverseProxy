package packets

import (
	"club.asynclab/asrp/pkg/api"
	"club.asynclab/asrp/pkg/logging"
)

var logger = logging.GetLogger()

const (
	PacketTypeHello api.PacketType = iota
	PacketTypeProxyStart
	PacketTypeProxyConfirm
	PacketTypeProxy
	PacketTypeUnknown
)

type PacketHello struct{ Data api.PackData }
type PacketMessage struct{ Data api.PackData }
type PacketProxyStart struct{ Data api.PackData }
type PacketProxyConfirm struct{ Data api.PackData }
type PacketProxy struct{ Data api.PackData }

func GetPacketFromData(d api.PackData) api.IPacket {
	switch d.Type {
	case PacketTypeHello:
		return &PacketHello{Data: d}
	case PacketTypeProxyStart:
		return &PacketProxyStart{Data: d}
	case PacketTypeProxyConfirm:
		return &PacketProxyConfirm{Data: d}
	case PacketTypeProxy:
		return &PacketProxy{Data: d}
	default:
		return nil
	}
}

func (p *PacketHello) GetData() api.PackData        { return p.Data }
func (p *PacketProxyStart) GetData() api.PackData   { return p.Data }
func (p *PacketProxyConfirm) GetData() api.PackData { return p.Data }
func (p *PacketProxy) GetData() api.PackData        { return p.Data }
