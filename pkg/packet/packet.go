package packet

import (
	"reflect"

	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/pkg/util"
	"github.com/vmihailenco/msgpack/v5"
)

type IPacket interface{}

type NetPackData map[string]interface{}

type NetPacket struct {
	Type int
	Data NetPackData
}

//----------------------------------------------------------------------------------------------------

var TypeMap = structure.NewBiMap[int, reflect.Type]()

func RegisterPacketWithKey[T IPacket](key int) {
	TypeMap.Put(key, util.GetForStructTypeWithType[T]())
}

func RegisterPacket[T IPacket]() {
	RegisterPacketWithKey[T](TypeMap.Len())
}

func GetNetPacketType(p IPacket) int {
	t, ok := TypeMap.GetKey(util.GetForStructType(p))
	if !ok {
		t = 0
	}
	return t
}

//----------------------------------------------------------------------------------------------------

func (netPacket *NetPacket) Serialize() ([]byte, error) { return msgpack.Marshal(netPacket) }
func Deserialize(bytes []byte) (*NetPacket, error) {
	netPacket := NetPacket{}
	return &netPacket, msgpack.Unmarshal(bytes, &netPacket)
}

func ToNetPacket(p IPacket) *NetPacket {
	return &NetPacket{
		Type: GetNetPacketType(p),
		Data: util.StructToMap(p),
	}
}

func FromNetPacket(netPacket *NetPacket) IPacket {
	if t, ok := TypeMap.GetValue(netPacket.Type); ok && netPacket.Type != 0 {
		p := reflect.New(t).Interface().(IPacket)
		return util.MapToStruct(netPacket.Data, p)
	}

	return &PacketUnknown{}
}

//----------------------------------------------------------------------------------------------------

// both
//
// 未知
type PacketUnknown struct {
	Err error
}

func init() {
	RegisterPacketWithKey[PacketUnknown](0)
}
