package server

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/pattern"
	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/program"
	"club.asynclab/asrp/program/general"
)

var logger = logging.GetLogger()

type Server struct {
	Meta             *program.ProgramMeta
	ListenAddress    string
	Sessions         *structure.SyncMap[string, string]
	ProxyConnections *structure.SyncMap[string, net.Conn]
}

func (server *Server) GetMeta() *program.ProgramMeta { return server.Meta }

func NewServer(ctx context.Context, listenAddress string) *Server {
	server := &Server{
		Meta:             program.NewProgramMeta(ctx),
		ListenAddress:    listenAddress,
		Sessions:         &structure.SyncMap[string, string]{},
		ProxyConnections: &structure.SyncMap[string, net.Conn]{},
	}
	general.AddGeneralEventHandler(server.Meta.EventBus)
	AddServerEventHandler(server.Meta.EventBus)
	return server
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	connCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	program.EmitEvent(conn, connCtx)
}

func (server *Server) Listen() {
	listener, err := net.Listen("tcp", server.ListenAddress)
	if err != nil {
		logger.Error("Error listening: ", err)
		return
	}
	defer listener.Close()

	logger.Info("Listening on: ", server.ListenAddress)

	pattern.SelectContextAndChannel(
		server.Meta.Ctx,
		make(chan net.Conn),
		func() {},
		func(conn net.Conn) bool {
			if conn == nil {
				return false
			}
			go server.handleConnection(conn)
			return true
		},
		func(ch chan net.Conn) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if server.Meta.Ctx.Err() != nil {
						close(ch)
						return
					}
					logger.Error("Error accepting connection: ", err)
					continue
				}
				ch <- conn
			}
		},
	)
}
