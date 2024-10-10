package program

import (
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
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
			for {
				frontendListener, err := net.Listen("tcp", address)
				if err != nil {
					logger.Error("Error listening frontend on: ", address)
					return false
				}
				defer frontendListener.Close()

				logger.Info("Started proxy on: ", address)

				acceptChan := make(chan net.Conn)
				go func() {
					for {
						frontendConn, err := frontendListener.Accept()
						if err != nil {
							if server.Ctx.Err() != nil {
								return
							}
							logger.Error("Error accepting connection: ", err)
							continue
						}
						acceptChan <- frontendConn
					}
				}()
				for {
					select {
					case <-server.Ctx.Done():
						break
					case conn := <-acceptChan:
						id := uuid.NewString()
						if ok := server.SendPacket(event.Conn, &packet.PacketNewProxyConnection{
							Name: event.Packet.Name,
							Uuid: id,
						}); ok {
							server.ProxyConnections.Store(id, conn)
						}
					}
				}
			}
		}
		return false
	})
	event.Subscribe(server.EventManager, func(event event.EventProxy) bool {
		if conn, ok := server.ProxyConnections.Load(event.Packet.Uuid); ok {
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

func (server *Server) emitEvent(conn net.Conn) {
	for {
		r, ok := server.ReceivePacket(conn)
		if !ok {
			return
		}

		switch r.Type() {
		case packet.NetPacketTypeHello:
			ok = event.Publish(server.EventManager, event.EventHello{
				Conn: conn,
			})
		case packet.NetPacketTypeProxyNegotiate:
			ok = event.Publish(server.EventManager, event.EventProxyNegotiate{
				Conn:   conn,
				Packet: *r.(*packet.PacketProxyNegotiate),
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
			return
		}
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	server.emitEvent(conn)
}

func (server *Server) Listen() {
	listener, err := net.Listen("tcp", server.ListenAddress)
	if err != nil {
		logger.Error("Error listening: ", err)
		return
	}
	defer listener.Close()

	logger.Info("Listening on: ", server.ListenAddress)

	acceptChan := make(chan net.Conn)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if server.Ctx.Err() != nil {
					return
				}
				logger.Error("Error accepting connection: ", err)
				continue
			}
			acceptChan <- conn
		}
	}()

	for {
		select {
		case <-server.Ctx.Done():
			return
		case conn := <-acceptChan:
			go server.handleConnection(conn)
		}
	}
}
