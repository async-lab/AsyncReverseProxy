package api

import (
	"encoding/json"
)

type PacketType int
type Side int

// const (
// 	SideServer Side = iota
// 	SideClient
// )

type PacketContent map[string]interface{}

type PackData struct {
	Type    PacketType    `json:"type"`
	Content PacketContent `json:"content"`
}

func (p PackData) Serialize() ([]byte, error) { return json.Marshal(p) }
func Deserialize(data []byte) (PackData, error) {
	p := PackData{}
	return p, json.Unmarshal(data, &p)
}

type IPacket interface {
	GetData() PackData
}
