package server

import (
	"context"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/arch/ushers"
	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/program"
	"club.asynclab/asrp/pkg/program/session"
)

var logger = logging.GetLogger()

type Server struct {
	*program.MetaProgram
	Config   config.ConfigServer
	Sessions *concurrent.ConcurrentMap[string, *session.ServerSession]
	Usher    arch.IUsher
}

func NewServer(parentCtx context.Context, config config.ConfigServer) (*Server, error) {
	meta := program.NewMetaProgram(parentCtx)
	if config.Server == nil {
		panic("Config is nil")
	}
	usher, err := ushers.NewUsherTLS(meta.Ctx, config.Server.Listen, config.Server.Token)
	if err != nil {
		return nil, err
	}
	server := &Server{
		MetaProgram: meta,
		Config:      config,
		Sessions:    concurrent.NewSyncMap[string, *session.ServerSession](),
		Usher:       usher,
	}
	return server, nil
}

func GetServer() *Server {
	server, ok := program.Program.(*Server)
	if !ok {
		panic("Program is not a server")
	}
	return server
}

func (server *Server) CheckConfig() bool {
	if server.Config.Server == nil {
		logger.Error("Server config is nil")
		return false
	}
	if server.Config.Server.Listen == "" {
		logger.Error("Server listen address is empty")
		return false
	}
	return true
}

func (server *Server) handleForwarder(fwv *arch.ForwarderWithValues) bool {
	s, ok := server.Sessions.Load(fwv.InitPacket.Name)
	if !ok {
		_s, err := session.NewServerSession(server.Ctx, fwv.InitPacket)
		if err != nil {
			logger.Error("Error creating session: ", err)
			fwv.Close()
			return true
		}
		server.Sessions.Store(fwv.InitPacket.Name, _s)
		s = _s
	}
	uuid := s.GetDispatcher().AddForwarder(fwv)
	logger.Info("["+fwv.InitPacket.Name+"] (", s.GetDispatcher().Len(), ") established")
	go func() {
		<-fwv.GetCtx().Done()
		s.GetDispatcher().RemoveForwarder(uuid)

		remaining := s.GetDispatcher().Len()

		logger.Info("["+fwv.InitPacket.Name+"] (", remaining, ") closed")

		if remaining == 0 {
			server.Sessions.Delete(fwv.InitPacket.Name)
			s.Close()
		}
	}()
	return true
}

func (server *Server) Run() {
	if !server.CheckConfig() {
		return
	}

	logger.Info("Server is running on ", server.Config.Server.Listen)

	channel.ConsumeWithCtx(server.Ctx, server.Usher.GetChanSendForwarder(), server.handleForwarder)
}
