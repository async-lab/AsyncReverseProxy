package server

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
	"club.asynclab/asrp/pkg/util"
)

func (server *Server) handleConnection(conn net.Conn) {
	defer func() {
		comm.SendPacket(conn, &packet.PacketEnd{})
		defer conn.Close()
	}()
	connCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.EmitEvent(conn, connCtx)
}

func (server *Server) Listen() {
	listener, err := net.Listen("tcp", server.Config.Server.ListenAddress)
	if err != nil {
		logger.Error("Error listening: ", err)
		return
	}

	go func() {
		defer listener.Close()
		<-server.Ctx.Done()
	}()

	logger.Info("Listening on: ", server.Config.Server.ListenAddress)

	go pattern.NewConfigSelectContextAndChannel[net.Conn]().
		WithCtx(server.Ctx).
		WithGoroutine(func(ch chan net.Conn) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if server.Ctx.Err() != nil || util.IsNetClose(err) {
						return
					}
					logger.Error("Error accepting connection: ", err)
					continue
				}
				ch <- conn
			}
		}).
		WithChannelHandler(func(conn net.Conn) {
			go server.handleConnection(conn)
		}).
		Run()
}
