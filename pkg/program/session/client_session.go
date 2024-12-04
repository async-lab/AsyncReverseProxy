package session

import (
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/program"
)

type ClientSession struct {
	*MetaSessionForConnection
	BackendAddress string
	// ProxyConns     *structure.SyncMap[string, *comm.Conn]
}

func NewClientSession(name string, backendAddress string) *ClientSession {
	return &ClientSession{
		MetaSessionForConnection: NewMetaSessionForConnection(name),
		BackendAddress:           backendAddress,
		// ProxyConns:               structure.NewSyncMap[string, *comm.Conn](), // connUuid -> proxyConn
	}
}

// func (s *ClientSession) AddProxyConnection(uuid string, conn *comm.Conn) {
// 	s.ProxyConns.Store(uuid, conn)

// 	go func() {
// 		defer s.ProxyConns.Delete(uuid)
// 		<-conn.Ctx.Done()
// 	}()
// }

func (s *ClientSession) AcceptFrontendConnection(uuid string, proxyConn *comm.Conn) bool {
	_conn, err := net.Dial("tcp", s.BackendAddress)
	if err != nil {
		return program.Program.SendPacket(proxyConn, &packet.PacketEndSideConnectionClosed{Name: s.Name, Uuid: uuid})
	}
	conn := comm.NewConnWithParentCtx(proxyConn.Ctx, _conn)
	s.EndConns.Store(uuid, conn)

	logger.Debug("[", s.Name, "] (", s.EndConns.Len(), ") end connection established")

	go func() {
		defer s.EndConns.Delete(uuid)
		<-conn.Ctx.Done()
	}()

	go func() {
		defer program.Program.SendPacket(proxyConn, &packet.PacketEndSideConnectionClosed{Name: s.Name, Uuid: uuid})
		defer conn.Close()

		for {
			bytes, err := comm.ReadForBytes(conn)
			if err != nil {
				if conn.Ctx.Err() != nil {
					return
				}
				continue
			}

			program.Program.SendPacket(proxyConn, &packet.PacketProxyData{
				Name: s.Name,
				Uuid: uuid,
				Data: bytes,
			})
		}
	}()

	return true
}
