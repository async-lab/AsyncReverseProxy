package packet

// both
//
// 打招呼
type PacketHello struct{}

// c to s
//
// 请求一个新的代理
type PacketProxyNegotiationRequest struct {
	// 提交字段
	Name            string
	FrontendAddress string
	Priority        uint32
	Weight          uint32
	Token           string

	// 回显字段
	RemoteServerName string
}

// s to c
//
// 代理请求确认
type PacketProxyNegotiationResponse struct {
	// 响应字段
	Name    string
	Success bool
	Reason  string

	// 原回显字段
	RemoteServerName string
}

// s to c
//
// 新终端连接
type PacketNewEndSideConnection struct {
	Name string
	Uuid string
}

// both
//
// 终端连接关闭
type PacketEndSideConnectionClosed struct {
	Name string
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

// both
//
// 结束
type PacketEnd struct{}

func init() {
	RegisterPacket[PacketHello]()
	RegisterPacket[PacketProxyNegotiationRequest]()
	RegisterPacket[PacketProxyNegotiationResponse]()
	RegisterPacket[PacketNewEndSideConnection]()
	RegisterPacket[PacketEndSideConnectionClosed]()
	RegisterPacket[PacketProxyData]()
	RegisterPacket[PacketEnd]()
}
