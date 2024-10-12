package program

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
)

func (client *Client) initEventManager() {
	event.Subscribe(client.EventManager, func(event event.EventHello) bool {
		logger.Info("Hello from: ", event.Conn.RemoteAddr())
		client.SendPacket(event.Conn, &packet.PacketHello{})
		return false
	})
	event.Subscribe(client.EventManager, func(event event.EventProxyConfirm) bool {
		logger.Info("Proxy confirmed: ", event.Packet.Name)
		return true
	})
	event.Subscribe(client.EventManager, func(event event.EventNewProxyConnection) bool {
		client.consume(func(conn net.Conn) bool {
			if address, ok := client.Sessions.Load(event.Packet.Name); ok {
				if ok := client.SendPacket(conn, &packet.PacketProxy{Uuid: event.Packet.Uuid}); ok {
					backendConn, err := net.Dial("tcp", address)
					if err != nil {
						logger.Error("Error connecting to backend server: ", err)
						return false
					}
					defer backendConn.Close()
					comm.Proxy(client.Ctx, conn, backendConn)
				}
			}
			return false
		})
		return true
	})
	event.Subscribe(client.EventManager, func(event event.EventEnd) bool {
		return false
	})
	event.Subscribe(client.EventManager, func(event event.EventUnknown) bool {
		event.Conn.Write([]byte("Hello, world!"))
		return false
	})
}

func (client *Client) emitEvent(conn net.Conn, connCtx context.Context) {
	pattern.SelectContextAndChannel(
		client.Ctx,
		make(chan struct{}),
		func() {},
		func(struct{}) bool { return false },
		func(ch chan struct{}) {
			for {
				r, ok := client.ReceivePacket(conn)
				if !ok {
					close(ch)
					return
				}

				switch r.Type() {
				case packet.NetPacketTypeHello:
					ok = event.Publish(client.EventManager, event.EventHello{
						Conn: conn,
					})
				case packet.NetPacketTypeProxyConfirm:
					ok = event.Publish(client.EventManager, event.EventProxyConfirm{
						Conn:   conn,
						Packet: *r.(*packet.PacketProxyConfirm),
					})
				case packet.NetPacketTypeNewProxyConnection:
					ok = event.Publish(client.EventManager, event.EventNewProxyConnection{
						Conn:   conn,
						Packet: *r.(*packet.PacketNewProxyConnection),
					})
				case packet.NetPacketTypeEnd:
					ok = event.Publish(client.EventManager, event.EventEnd{})
				case packet.NetPacketTypeUnknown:
					ok = event.Publish(client.EventManager, event.EventUnknown{Conn: conn})
				}

				if !ok {
					close(ch)
					return
				}
			}
		},
	)
}

func (client *Client) consume(consumer func(net.Conn) bool) {
	conn, err := net.Dial("tcp", client.RemoteAddress)
	if err != nil {
		logger.Error("Error connecting to remote server: ", err)
		return
	}
	defer conn.Close()
	connCtx, cancel := context.WithCancel(context.Background())
	if ok := consumer(conn); ok {
		client.emitEvent(conn, connCtx)
	}
	cancel()
}

func (client *Client) Hello() {
	client.consume(func(conn net.Conn) bool {
		return client.SendPacket(conn, &packet.PacketHello{})
	})
}

func (client *Client) StartProxy(name string, frontendAddress string, backendAddress string) {
	client.consume(func(conn net.Conn) bool {
		ok := client.SendPacket(conn, &packet.PacketProxyNegotiate{
			Name:            name,
			FrontendAddress: frontendAddress,
		})
		if ok {
			client.Sessions.Store(name, backendAddress)
		}
		return ok
	})
}
