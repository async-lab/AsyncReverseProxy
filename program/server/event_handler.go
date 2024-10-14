package server

import (
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
	"club.asynclab/asrp/program"
	"github.com/google/uuid"
)

func EventHandlerProxyNegotiate(event event.EventProxyNegotiate) bool {
	server, ok := program.Program.(*Server)
	if !ok {
		return true
	}

	confirm := func(success bool) bool {
		return program.SendPacket(event.Conn, &packet.PacketProxyConfirm{
			Name:    event.Packet.Name,
			Success: success,
		})
	}

	frontendListener, err := net.Listen("tcp", event.Packet.FrontendAddress)
	if err != nil {
		logger.Error("Error listening frontend on: ", event.Packet.FrontendAddress)
		return confirm(false)
	}
	logger.Info("Started proxy on: ", event.Packet.FrontendAddress)

	server.Sessions.Store(event.Packet.Name, event.Packet.FrontendAddress)
	if !confirm(true) {
		return false
	}

	go pattern.SelectContextAndChannel(
		server.Meta.Ctx,
		make(chan net.Conn),
		func() {},
		func(conn net.Conn) bool {
			if conn == nil {
				return false
			}
			id := uuid.NewString()
			if ok := program.SendPacket(event.Conn, &packet.PacketNewProxyConnection{
				Name: event.Packet.Name,
				Uuid: id,
			}); ok {
				server.ProxyConnections.Store(id, conn)
			}
			return true
		},
		func(ch chan net.Conn) {
			for {
				frontendConn, err := frontendListener.Accept()
				if err != nil {
					if server.Meta.Ctx.Err() != nil || event.ConnCtx.Err() != nil {
						close(ch)
						return
					}
					logger.Error("Error accepting connection: ", err)
					continue
				}
				ch <- frontendConn
			}
		},
	)

	go func() {
		<-event.ConnCtx.Done()
		server.Sessions.Delete(event.Packet.Name)
		logger.Info("Stopped proxy on: ", event.Packet.Name)
		frontendListener.Close()
	}()

	return true
}

func EventHandlerProxy(event event.EventProxy) bool {
	server, ok := program.Program.(*Server)
	if !ok {
		return true
	}

	if conn, ok := server.ProxyConnections.Load(event.Packet.Uuid); ok {
		defer server.ProxyConnections.Delete(event.Packet.Uuid)
		defer conn.Close()
		comm.Proxy(server.Meta.Ctx, event.Conn, conn)
	}
	return false
}

func AddServerEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerProxyNegotiate)
	event.Subscribe(bus, EventHandlerProxy)
}
