package server

import (
	"crypto/tls"

	"club.asynclab/asrp/pkg/base/lang"
	"club.asynclab/asrp/pkg/base/pattern"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/util"
)

func (server *Server) handleConnection(conn *comm.Conn) {
	defer func() {
		comm.SendPacket(conn, &packet.PacketEnd{})
		defer conn.Close()
	}()
	server.EmitEventReceivePacket(conn)
}

func (server *Server) Listen() {
	cert, err := util.GenerateSelfSignedCert()
	if err != nil {
		logger.Error("Error generating cert: ", err)
		return
	}

	listener, err := tls.Listen("tcp", server.Config.Server.ListenAddress, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		logger.Error("Error listening: ", err)
		return
	}

	go func() {
		defer listener.Close()
		<-server.Ctx.Done()
	}()

	logger.Info("Listening on: ", server.Config.Server.ListenAddress)

	go pattern.NewConfigSelectContextAndChannel[*comm.Conn]().
		WithCtx(server.Ctx).
		WithGoroutine(func(ch chan *comm.Conn) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if server.Ctx.Err() != nil || lang.IsNetClose(err) {
						return
					}
					logger.Error("Error accepting connection: ", err)
					continue
				}
				ch <- comm.NewConn(conn)
			}
		}).
		WithChannelHandler(func(conn *comm.Conn) {
			go server.handleConnection(conn)
		}).
		Run()
}
