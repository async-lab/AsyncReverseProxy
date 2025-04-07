package packet

type PacketUnknown struct {
	Err error
}

// c to s
//
// 请求一个新的代理
type PacketProxyNegotiationRequest struct {
	Name         string
	FrontendAddr string
	Priority     uint32
	Weight       uint32
	Token        string
}

// s to c
//
// 代理请求确认
type PacketProxyNegotiationResponse struct {
	Success bool
	Reason  string
}

// -------------------------------------------------------

type IPacketForConn interface {
	GetUuid() string
}

type MetaPacketForConn struct {
	Uuid string
}

func (m *MetaPacketForConn) GetUuid() string {
	return m.Uuid
}

// both
//
// 终端连接关闭
type PacketEndSideConnectionClosed struct {
	MetaPacketForConn
}

// both
//
// 代理数据包
type PacketProxyData struct {
	MetaPacketForConn
	Data []byte
}

// both
//
// 结束
type PacketEnd struct{}

func init() {
	RegisterPacketWithKey[PacketUnknown](0)
	RegisterPacket[PacketProxyNegotiationRequest]()
	RegisterPacket[PacketProxyNegotiationResponse]()
	RegisterPacket[PacketEndSideConnectionClosed]()
	RegisterPacket[PacketProxyData]()
	RegisterPacket[PacketEnd]()
}
