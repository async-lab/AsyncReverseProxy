package session

import (
	"context"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/program"
	"github.com/google/uuid"
)

type ServerSession struct {
	*MetaSessionForConnection
	Listener *comm.Listener
	Lb       *LoadBalancer
}

func NewServerSession(name string, listener *comm.Listener) *ServerSession {
	return &ServerSession{
		MetaSessionForConnection: NewMetaSessionForConnection(name),
		Listener:                 listener,
		Lb:                       NewLoadBalancer(),
	}
}

func (s *ServerSession) AddProxyConn(proxyConn *comm.Conn, priority int, weight int) context.Context {
	uuid := s.Lb.AddConn(proxyConn, priority, weight)
	logger.Info("[", s.Name, "] (", s.Lb.Len(), ") connection established")

	closeCtx, cancel := context.WithCancel(context.Background())

	go func() {
		defer func() {
			s.Lb.RemoveConn(uuid)
			if s.Lb.Len() == 0 {
				s.Listener.Close()
				logger.Info("[", s.Name, "] (", s.Lb.Len(), ") connection closed, all connections closed")
				cancel()
			} else {
				logger.Info("[", s.Name, "] (", s.Lb.Len(), ") connection closed")
			}
		}()
		<-proxyConn.Ctx.Done()
	}()

	return closeCtx
}

func (s *ServerSession) Start() {
	go func() {
		for {
			conn, err := s.Listener.Accept()
			if err != nil {
				if s.Listener.Ctx.Err() != nil {
					return
				}
				continue
			}
			go s.handleAcceptedFrontendConnection(conn)
		}
	}()
}

func (s *ServerSession) handleAcceptedFrontendConnection(frontConn *comm.Conn) {
	defer frontConn.Close()

	_, proxyConn, ok := s.Lb.Next()
	if !ok {
		return
	}

	connUuid := uuid.NewString()

	if !program.Program.SendPacket(proxyConn, &packet.PacketNewEndSideConnection{Name: s.Name, Uuid: connUuid}) {
		return
	}

	s.EndConns.Store(connUuid, frontConn)
	defer s.EndConns.Delete(connUuid)
	defer comm.SendPacket(proxyConn, &packet.PacketEndSideConnectionClosed{Name: s.Name, Uuid: connUuid})
	defer logger.Debug("[", s.Name, "] (", s.Lb.Len(), ") end connection closed")

	logger.Debug("[", s.Name, "] (", s.Lb.Len(), ") end connection established")

	go func() {
		defer frontConn.Close()
		<-proxyConn.Ctx.Done()
	}()

	for {
		bytes, err := comm.ReadForBytes(frontConn)
		if err != nil {
			if frontConn.Ctx.Err() != nil {
				return
			}
			continue
		}

		if !program.Program.SendPacket(proxyConn, &packet.PacketProxyData{
			Name: s.Name,
			Uuid: connUuid,
			Data: bytes,
		}) {
			return
		}
	}
}
