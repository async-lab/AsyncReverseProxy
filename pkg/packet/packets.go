package packet

// both
//
// 打招呼
type PacketHello struct{}

// c to s
//
// 请求一个新的代理
type PacketProxyNegotiationRequest struct {
	Name            string
	FrontendAddress string
}

// s to c
//
// 代理请求确认
type PacketProxyNegotiationResponse struct {
	Name    string
	Success bool
}

// s to c
//
// 请求新的代理连接
type PacketProxyConnectionRequest struct {
	Name string
	Uuid string
	Num  int
}

// c to s
//
// 代理连接标识
type PacketProxyConnectionResponse struct {
	Name string
}

// s to c
//
// 新终端连接
type PacketNewEndConnection struct {
	Name string
	Uuid string
}

// both
//
// 终端连接关闭
type PacketEndConnectionClosed struct {
	Uuid string
}

// both
//
// 代理数据包
type PacketProxyData struct {
	Name string
	Uuid string
	Data []byte
}

// c to s
//
// 心跳
// type PacketHeartbeat struct {
// 	Name string
// }

// both
//
// 结束
type PacketEnd struct{}

func init() {
	RegisterPacket[PacketHello]()
	RegisterPacket[PacketProxyNegotiationRequest]()
	RegisterPacket[PacketProxyNegotiationResponse]()
	RegisterPacket[PacketProxyConnectionRequest]()
	RegisterPacket[PacketProxyConnectionResponse]()
	RegisterPacket[PacketNewEndConnection]()
	RegisterPacket[PacketEndConnectionClosed]()
	RegisterPacket[PacketProxyData]()
	RegisterPacket[PacketEnd]()
}
