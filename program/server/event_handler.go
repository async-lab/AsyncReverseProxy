package server

import (
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/pkg/util"
	"github.com/google/uuid"
)

func EventHandlerReceivedPacketProxyNegotiationRequest(e *event.EventReceivedPacket[*packet.PacketProxyNegotiationRequest]) bool {
	server := GetServer()
	proxyUuid := uuid.NewString()

	sendResponse := func(success bool, reason string) {
		server.SendPacket(e.Conn, &packet.PacketProxyNegotiationResponse{
			Name:    e.Packet.Name,
			Success: success,
			Reason:  reason,
		})
	}

	if e.Packet.Token != server.Config.Server.Token {
		sendResponse(false, "Invalid token")
		return false
	}

	addToProxiesConnections := func() {
		server.ProxyConnections.Compute(func(s *structure.SyncMap[string, *structure.SyncMap[string, net.Conn]]) {
			conns, _ := s.LoadOrStore(e.Packet.Name, structure.NewSyncMap[string, net.Conn]())
			conns.Store(proxyUuid, e.Conn)
		})
	}

	go func() {
		defer func() {
			if conns, ok := server.ProxyConnections.Load(e.Packet.Name); ok {
				conns.Delete(proxyUuid)
			}
		}()

		<-e.ConnCtx.Done()
	}()

	frontendAddress, loaded := server.Sessions.LoadOrStore(e.Packet.Name, e.Packet.FrontendAddress)
	if loaded {
		addToProxiesConnections()
		sendResponse(true, "")
		return true
	}

	listener, err := net.Listen("tcp", frontendAddress)
	if err != nil {
		sendResponse(false, err.Error())
		return false
	}

	go func() {
		defer listener.Close()
		<-e.ConnCtx.Done()
	}()

	addToProxiesConnections()
	sendResponse(true, "")

	logger.Info("New proxy [", e.Packet.Name, "] negotiation success")

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

					if !server.SendPacket(e.Conn, &packet.PacketNewEndConnection{Name: e.Packet.Name, Uuid: connUuid}) {
						return
					}

					server.FrontendConnections.Store(connUuid, conn)
					defer server.FrontendConnections.Delete(connUuid)
					defer server.SendPacket(e.Conn, &packet.PacketEndConnectionClosed{Uuid: connUuid})

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
					event.Publish(server.EventBus, &event.EventPacketProxyDataQueue{Packet: p})
				}).
				Run()
		}).
		Run()

	return true
}

func EventHandlerReceivedPacketProxyData(e *event.EventReceivedPacket[*packet.PacketProxyData]) bool {
	server := GetServer()
	conn, ok := server.FrontendConnections.Load(e.Packet.Uuid)
	if !ok {
		return true
	}
	if _, err := conn.Write(e.Packet.Data); err != nil {
		server.FrontendConnections.Delete(e.Packet.Uuid)
	}
	return true
}

func EventHandlerReceivedPacketEndConnectionClosed(e *event.EventReceivedPacket[*packet.PacketEndConnectionClosed]) bool {
	server := GetServer()
	if conn, ok := server.FrontendConnections.Load(e.Packet.Uuid); ok {
		conn.Close()
	}
	server.FrontendConnections.Delete(e.Packet.Uuid)
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

	if conns.Len() == 0 {
		if conn, ok := server.FrontendConnections.Load(e.Packet.Uuid); ok {
			conn.Close()
		}
		server.FrontendConnections.Delete(e.Packet.Uuid)
		server.ProxyConnections.Delete(e.Packet.Name)
		server.Sessions.Delete(e.Packet.Name)
	}

	return true
}

func AddServerEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerReceivedPacketProxyNegotiationRequest)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyData)
	event.Subscribe(bus, EventHandlerPacketProxyDataQueue)
	event.Subscribe(bus, EventHandlerReceivedPacketEndConnectionClosed)
}
