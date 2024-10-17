package server

import (
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/pkg/util"
	"club.asynclab/asrp/program"
	"github.com/google/uuid"
)

func EventHandlerReceivedPacketProxyNegotiationRequest(e *event.EventReceivedPacket[*packet.PacketProxyNegotiationRequest]) bool {
	server := GetServer()

	sendResponse := func(success bool) {
		server.SendPacket(e.Conn, &packet.PacketProxyNegotiationResponse{
			Name:    e.Packet.Name,
			Success: success,
		})
	}

	addToProxiesConnections := func() {
		server.ProxyConnections.Compute(func(s *structure.SyncMap[string, *structure.SyncMap[string, net.Conn]]) {
			conns, _ := s.LoadOrStore(e.Packet.Name, structure.NewSyncMap[string, net.Conn]())
			conns.Store(uuid.NewString(), e.Conn)
		})
	}

	frontendAddress, loaded := server.Sessions.LoadOrStore(e.Packet.Name, e.Packet.FrontendAddress)
	if loaded {
		addToProxiesConnections()
		sendResponse(true)
		return true
	}

	listener, err := net.Listen("tcp", frontendAddress)
	if err != nil {
		sendResponse(false)
		return false
	}

	go func() {
		defer listener.Close()
		<-server.Ctx.Done()
	}()

	addToProxiesConnections()
	sendResponse(true)

	go pattern.NewConfigSelectContextAndChannel[net.Conn]().
		WithCtx(server.Ctx).
		WithGoroutine(func(connCh chan net.Conn) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if server.Ctx.Err() != nil || util.IsNetClose(err) {
						return
					}
					continue
				}
				connCh <- conn
			}
		}).
		WithChannelHandler(func(conn net.Conn) {
			go pattern.NewConfigSelectContextAndChannel[*packet.PacketProxyData]().
				WithCtx(server.Ctx).
				WithGoroutine(func(packetCh chan *packet.PacketProxyData) {
					defer conn.Close()
					connUuid := uuid.NewString()

					server.FrontendConnections.Store(connUuid, conn)
					for {
						bytes, err := comm.ReadForBytes(conn)
						if err != nil {
							if server.Ctx.Err() != nil || util.IsNetClose(err) {
								return
							}
							continue
						}
						packetCh <- &packet.PacketProxyData{
							Name: e.Packet.Name,
							Uuid: connUuid,
							Data: bytes,
						}
					}
				}).
				WithChannelHandler(func(p *packet.PacketProxyData) {
					go event.Publish(server.EventBus, &event.EventPacketProxyDataQueue{Packet: p})
				}).
				Run()
		}).
		Run()

	return true
}

func EventHandlerReceivedPacketProxyConnectionResponse(e *event.EventReceivedPacket[*packet.PacketProxyConnectionResponse]) bool {
	server := GetServer()
	print(server)
	return true
}

func EventHandlerReceivedPacketProxyData(e *event.EventReceivedPacket[*packet.PacketProxyData]) bool {
	server, ok := program.Program.(*Server)
	if !ok {
		return false
	}
	conn, ok := server.FrontendConnections.Load(e.Packet.Uuid)
	if !ok {
		return false
	}
	go func() {
		if _, err := conn.Write(e.Packet.Data); err != nil {
			server.FrontendConnections.Delete(e.Packet.Uuid)
		}
	}()

	return true
}

// TODO: 这里应该改成轮询每个连接
func EventHandlerPacketProxyDataQueue(e *event.EventPacketProxyDataQueue) bool {
	server := GetServer()

	conns, loaded := server.ProxyConnections.LoadOrStore(e.Packet.Name, structure.NewSyncMap[string, net.Conn]())
	if !loaded {
		return false
	}

	errConns := make([]string, 0)

	conns.Range(func(key string, value net.Conn) bool {
		if !server.SendPacket(value, e.Packet) {
			errConns = append(errConns, key)
		}
		return true
	})

	for _, key := range errConns {
		conns.Delete(key)
	}

	return true
}
func AddServerEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerReceivedPacketProxyNegotiationRequest)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyConnectionResponse)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyData)
	event.Subscribe(bus, EventHandlerPacketProxyDataQueue)
}
