package server

import (
	"context"
	"net"

	"club.asynclab/asrp/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/program"
	"club.asynclab/asrp/program/general"
)

var logger = logging.GetLogger()

type Server struct {
	program.MetaProgram
	Config              *config.ConfigServer
	Sessions            *structure.SyncMap[string, string]                               // name -> frontend_address
	ProxyConnections    *structure.SyncMap[string, *structure.SyncMap[string, net.Conn]] // name -> conn
	FrontendConnections *structure.SyncMap[string, net.Conn]                             // uuid -> conn
}

func NewServer(ctx context.Context, config *config.ConfigServer) *Server {
	server := &Server{
		MetaProgram:         *program.NewMetaProgram(ctx),
		Config:              config,
		Sessions:            structure.NewSyncMap[string, string](),
		ProxyConnections:    structure.NewSyncMap[string, *structure.SyncMap[string, net.Conn]](),
		FrontendConnections: structure.NewSyncMap[string, net.Conn](),
	}
	general.AddGeneralEventHandler(server.EventBus)
	AddServerEventHandler(server.EventBus)
	return server
}

func GetServer() *Server {
	server, ok := program.Program.(*Server)
	if !ok {
		panic("Program is not a server")
	}
	return server
}

func (server *Server) CheckConfig() bool {
	if server.Config == nil {
		logger.Error("Config is nil")
		return false
	}
	if server.Config.Server == nil {
		logger.Error("Server config is nil")
		return false
	}
	return true
}

func (server *Server) Run() {
	if !server.CheckConfig() {
		return
	}

	logger.Info("Server starting...")

	server.Listen()

	<-server.Ctx.Done()
}
