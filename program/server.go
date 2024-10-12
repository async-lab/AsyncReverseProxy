package program

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
	"github.com/google/uuid"
)

func (server *Server) initEventManager() {
	event.Subscribe(server.EventManager, func(event event.EventHello) bool {
		logger.Info("Hello from: ", event.Conn.RemoteAddr())
		server.SendPacket(event.Conn, &packet.PacketHello{})
		return false
	})

	event.Subscribe(server.EventManager, func(event event.EventProxyNegotiate) bool {
		server.Sessions.Store(event.Packet.Name, event.Packet.FrontendAddress)
		if ok := server.SendPacket(event.Conn, &packet.PacketProxyConfirm{
			Name: event.Packet.Name,
		}); !ok {
			return false
		}

		if address, ok := server.Sessions.Load(event.Packet.Name); ok {
			frontendListener, err := net.Listen("tcp", address)
			if err != nil {
				logger.Error("Error listening frontend on: ", address)
				return true
			}
			logger.Info("Started proxy on: ", address)

			go pattern.SelectContextAndChannel(
				server.Ctx,
				make(chan net.Conn),
				func() {},
				func(conn net.Conn) bool {
					if conn == nil {
						return false
					}
					id := uuid.NewString()
					if ok := server.SendPacket(event.Conn, &packet.PacketNewProxyConnection{
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
							if server.Ctx.Err() != nil || event.ConnCtx.Err() != nil {
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
		}
		return true
	})
	event.Subscribe(server.EventManager, func(event event.EventProxy) bool {
		if conn, ok := server.ProxyConnections.Load(event.Packet.Uuid); ok {
			defer server.ProxyConnections.Delete(event.Packet.Uuid)
			defer conn.Close()
			comm.Proxy(server.Ctx, event.Conn, conn)
		}
		return false
	})
	event.Subscribe(server.EventManager, func(event event.EventEnd) bool {
		return false
	})
	event.Subscribe(server.EventManager, func(event event.EventUnknown) bool {
		event.Conn.Write([]byte("Hello, world!"))
		return false
	})
}

func (server *Server) emitEvent(conn net.Conn, connCtx context.Context) {
	pattern.SelectContextAndChannel(
		server.Ctx,
		make(chan struct{}),
		func() {},
		func(struct{}) bool { return false },
		func(ch chan struct{}) {
			for {
				r, ok := server.ReceivePacket(conn)
				if !ok {
					close(ch)
					return
				}

				switch r.Type() {
				case packet.NetPacketTypeHello:
					ok = event.Publish(server.EventManager, event.EventHello{
						Conn: conn,
					})
				case packet.NetPacketTypeProxyNegotiate:
					ok = event.Publish(server.EventManager, event.EventProxyNegotiate{
						Conn:    conn,
						ConnCtx: connCtx,
						Packet:  *r.(*packet.PacketProxyNegotiate),
					})
				case packet.NetPacketTypeProxy:
					ok = event.Publish(server.EventManager, event.EventProxy{
						Conn:   conn,
						Packet: *r.(*packet.PacketProxy),
					})
				case packet.NetPacketTypeEnd:
					ok = event.Publish(server.EventManager, event.EventEnd{})
				case packet.NetPacketTypeUnknown:
					ok = event.Publish(server.EventManager, event.EventUnknown{Conn: conn})
				}

				if !ok {
					close(ch)
					return
				}
			}
		})
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	connCtx, cancel := context.WithCancel(context.Background())
	server.emitEvent(conn, connCtx)
	cancel()
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
		server.Ctx,
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
					if server.Ctx.Err() != nil {
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
