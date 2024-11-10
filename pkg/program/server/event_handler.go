package server

import (
	"net"

	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/lang"
	"club.asynclab/asrp/pkg/base/pattern"
	"club.asynclab/asrp/pkg/base/structure"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/program"
	"github.com/google/uuid"
)

func EventHandlerReceivedPacketProxyNegotiationRequest(e *event.EventReceivedPacket[*packet.PacketProxyNegotiationRequest]) bool {
	server := GetServer()

	sendResponse := func(success bool, reason string) {
		server.SendPacket(e.Conn, &packet.PacketProxyNegotiationResponse{
			Name:             e.Packet.Name,
			Success:          success,
			Reason:           reason,
			RemoteServerName: e.Packet.RemoteServerName,
			BackendAddress:   e.Packet.BackendAddress,
		})
	}

	if e.Packet.Token != server.Config.Server.Token {
		sendResponse(false, "Invalid token")
		return false
	}

	addToLoadBalancers := func() {
		lb, _ := server.LoadBalancers.LoadOrStore(e.Packet.Name, program.NewLoadBalancer())
		proxyUuid := lb.AddConn(&program.ProxyConnection{
			Priority: e.Packet.Priority,
			Weight:   e.Packet.Weight,
			Conn:     e.Conn,
		})

		logger.Info("[", e.Packet.Name, "] (", lb.Len(), ") new negotiation success")

		go func() {
			defer func() {
				if lb != nil {
					lb.Compute(func(lb *program.LoadBalancer) {
						lb.RemoveConn(proxyUuid)
						logger.Info("[", e.Packet.Name, "] (", lb.Len(), ") connection closed")
						if lb.Len() == 0 {
							if entry, ok := server.Sessions.Load(e.Packet.Name); ok {
								entry.Value.Close()
							}
							server.LoadBalancers.Delete(e.Packet.Name)
							server.Sessions.Delete(e.Packet.Name)
							logger.Info("[", e.Packet.Name, "] all connections closed")
						}
					})
				}
			}()
			<-e.Conn.Ctx.Done()
		}()
	}

	skip, isSuccess := false, true

	server.Sessions.Compute(func(s *structure.SyncMap[string, container.Entry[string, net.Listener]]) {
		if _, ok := server.Sessions.Load(e.Packet.Name); ok {
			addToLoadBalancers()
			sendResponse(true, "")
			skip, isSuccess = true, true
			return
		}

		listener, err := net.Listen("tcp", e.Packet.FrontendAddress)
		if err != nil {
			sendResponse(false, err.Error())
			skip, isSuccess = true, true
			return
		}

		server.Sessions.Store(e.Packet.Name, *container.NewEntry(e.Packet.FrontendAddress, listener))
	})

	if skip {
		return isSuccess
	}

	addToLoadBalancers()
	sendResponse(true, "")

	go pattern.NewConfigSelectContextAndChannel[*comm.Conn]().
		WithCtx(server.Ctx).
		WithGoroutine(func(connCh chan *comm.Conn) {
			entry, ok := server.Sessions.Load(e.Packet.Name)
			if !ok {
				return
			}
			listener := entry.Value
			for {
				conn, err := listener.Accept()
				if err != nil {
					if server.Ctx.Err() != nil || lang.IsNetClose(err) {
						return
					}
					continue
				}
				connCh <- comm.NewConn(conn)
			}
		}).
		WithChannelHandler(func(conn *comm.Conn) {
			event.Publish(server.EventBus, &event.EventAcceptedFrontendConnection{
				Name: e.Packet.Name,
				Conn: conn,
			})
		}).
		Run()

	return true
}

func EventHandlerReceivedPacketProxyData(e *event.EventReceivedPacket[*packet.PacketProxyData]) bool {
	server := GetServer()
	conns, ok := server.FrontendConnections.Load(e.Packet.Uuid)
	if !ok {
		return true
	}
	if _, err := conns.Key.Write(e.Packet.Data); err != nil {
		server.FrontendConnections.Delete(e.Packet.Uuid)
	}
	return true
}

func EventHandlerReceivedPacketEndSideConnectionClosed(e *event.EventReceivedPacket[*packet.PacketEndSideConnectionClosed]) bool {
	server := GetServer()
	if conns, ok := server.FrontendConnections.Load(e.Packet.Uuid); ok {
		conns.Key.Close()
	}
	server.FrontendConnections.Delete(e.Packet.Uuid)
	return true
}

func EventHandlerAcceptedFrontendConnection(e *event.EventAcceptedFrontendConnection) bool {
	server := GetServer()
	go pattern.NewConfigSelectContextAndChannel[*packet.PacketProxyData]().
		WithCtx(server.Ctx).
		WithGoroutine(func(packetCh chan *packet.PacketProxyData) {
			defer e.Conn.Close()
			connUuid := uuid.NewString()

			lb, loaded := server.LoadBalancers.LoadOrStore(e.Name, program.NewLoadBalancer())
			if !loaded {
				return
			}
			_, proxyConn, ok := lb.Next()
			if !ok {
				return
			}

			if !server.SendPacket(proxyConn, &packet.PacketNewEndSideConnection{Name: e.Name, Uuid: connUuid}) {
				return
			}

			server.FrontendConnections.Store(connUuid, *container.NewEntry(e.Conn, proxyConn))
			defer server.FrontendConnections.Delete(connUuid)
			defer comm.SendPacket(proxyConn, &packet.PacketEndSideConnectionClosed{Uuid: connUuid})

			go func() {
				defer e.Conn.Close()
				<-proxyConn.Ctx.Done()
			}()

			for {
				bytes, err := comm.ReadForBytes(e.Conn)
				if err != nil {
					if server.Ctx.Err() != nil || lang.IsNetClose(err) {
						return
					}
					continue
				}
				packetCh <- &packet.PacketProxyData{
					Name: e.Name,
					Uuid: connUuid,
					Data: bytes,
				}
			}
		}).
		WithChannelHandlerWithInterruption(func(p *packet.PacketProxyData) bool {
			return event.Publish(server.EventBus, &event.EventPacketProxyDataQueue{Packet: p})
		}).
		Run()
	return true
}

func EventHandlerPacketProxyDataQueue(e *event.EventPacketProxyDataQueue) bool {
	server := GetServer()

	conns, ok := server.FrontendConnections.Load(e.Packet.Uuid)
	if !ok || !server.SendPacket(conns.Value, e.Packet) {
		return false
	}
	return true
}

func AddServerEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerReceivedPacketProxyNegotiationRequest)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyData)
	event.Subscribe(bus, EventHandlerAcceptedFrontendConnection)
	event.Subscribe(bus, EventHandlerPacketProxyDataQueue)
	event.Subscribe(bus, EventHandlerReceivedPacketEndSideConnectionClosed)
}
