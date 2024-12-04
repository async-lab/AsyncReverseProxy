package server

import (
	"context"

	"club.asynclab/asrp/pkg/base/structure"
	"club.asynclab/asrp/pkg/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/program"
	"club.asynclab/asrp/pkg/program/general"
	"club.asynclab/asrp/pkg/program/session"
)

var logger = logging.GetLogger()

type Server struct {
	program.MetaProgram
	Config   *config.ConfigServer
	Sessions *structure.SyncMap[string, *session.ServerSession]
}

func NewServer(ctx context.Context, config *config.ConfigServer) *Server {
	server := &Server{
		MetaProgram: *program.NewMetaProgram(ctx),
		Config:      config,
		Sessions:    structure.NewSyncMap[string, *session.ServerSession](),
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
