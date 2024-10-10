package packet

import (
	"encoding/json"
	"reflect"

	"club.asynclab/asrp/pkg/util"
)

// type Side int
// const (
//
//	SideServer Side = iota
//	SideClient
//

type IPacket interface {
	Type() NetPacketType
}
type NetPacketType int
type NetPackData map[string]interface{}

var TypeMap = map[NetPacketType]IPacket{}

type NetPacket struct {
	Type NetPacketType `json:"type"`
	Data NetPackData   `json:"data"`
}

func (networkPacket NetPacket) Serialize() ([]byte, error) { return json.Marshal(networkPacket) }
func Deserialize(bytes []byte) (NetPacket, error) {
	netPacket := NetPacket{}
	return netPacket, json.Unmarshal(bytes, &netPacket)
}

func NewNetPacket(p IPacket) NetPacket {
	return NetPacket{
		Type: p.Type(),
		Data: util.StructToMap(p),
	}
}

func FromNetPacket(netPacket NetPacket) IPacket {
	if p, ok := TypeMap[netPacket.Type]; ok {
		c := reflect.New(util.GetStructType(p)).Interface().(IPacket)
		return util.MapToStruct(netPacket.Data, c)
	}

	return &PacketUnknown{}
}
