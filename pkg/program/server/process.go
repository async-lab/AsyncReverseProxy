package server

import (
	"crypto/tls"

	"club.asynclab/asrp/pkg/base/pattern"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/util"
)

func (server *Server) handleConnection(conn *comm.Conn) {
	defer func() {
		defer conn.Close()
		comm.SendPacket(conn, &packet.PacketEnd{})
	}()
	server.EmitEventReceivePacket(conn)
}

func (server *Server) Listen() {
	cert, err := util.GenerateSelfSignedCert()
	if err != nil {
		logger.Error("Error generating cert: ", err)
		return
	}

	_listener, err := tls.Listen("tcp", server.Config.Server.ListenAddress, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		logger.Error("Error listening: ", err)
		return
	}

	listener := comm.NewListenerWithParentCtx(server.Ctx, _listener)

	logger.Info("Listening on: ", server.Config.Server.ListenAddress)

	go pattern.NewConfigSelectContextAndChannel[*comm.Conn]().
		WithCtx(server.Ctx).
		WithGoroutine(func(ch chan *comm.Conn) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if listener.Ctx.Err() != nil {
						return
					}
					logger.Error("Error accepting connection: ", err)
					continue
				}
				ch <- comm.NewConnWithParentCtx(server.Ctx, conn)
			}
		}).
		WithChannelHandler(func(conn *comm.Conn) {
			go server.handleConnection(conn)
		}).
		Run()
}
