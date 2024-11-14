package client

import (
	"net"

	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/lang"
	"club.asynclab/asrp/pkg/base/pattern"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
)

func EventHandlerReceivedPacketProxyNegotiationResponse(e *event.EventReceivedPacket[*packet.PacketProxyNegotiationResponse]) bool {
	client := GetClient()

	if e.Packet.Success {
		logger.Info("[", e.Packet.Name, "] -> [", e.Packet.RemoteServerName, "] successfully negotiated")
		client.Sessions.Store(*container.NewEntry(e.Packet.Name, e.Packet.RemoteServerName), e.Packet.BackendAddress)
	} else {
		logger.Error("[", e.Packet.Name, "] -> [", e.Packet.RemoteServerName, "] negotiation failed")
		logger.Error("Reason: ", e.Packet.Reason)
		client.Sessions.Delete(*container.NewEntry(e.Packet.Name, e.Packet.RemoteServerName))
	}

	return e.Packet.Success
}

func EventHandlerReceivedPacketProxyData(e *event.EventReceivedPacket[*packet.PacketProxyData]) bool {
	client := GetClient()
	client.ProxyConnections.LoadOrStore(e.Packet.Uuid, e.Conn)

	conn, loaded := client.BackendConnections.Load(e.Packet.Uuid)
	if !loaded {
		return client.SendPacket(e.Conn, &packet.PacketEndSideConnectionClosed{Uuid: e.Packet.Uuid})
	}

	if _, err := conn.Write(e.Packet.Data); err != nil {
		client.BackendConnections.Delete(e.Packet.Uuid)
		return client.SendPacket(e.Conn, &packet.PacketEndSideConnectionClosed{Uuid: e.Packet.Uuid})
	}

	return true
}

func EventHandlerReceivedPacketNewEndSideConnection(e *event.EventReceivedPacket[*packet.PacketNewEndSideConnection]) bool {
	client := GetClient()
	session, loaded := client.Sessions.Stream().Filter(func(t container.Entry[container.Entry[string, string], string]) bool {
		return t.Key.Key == e.Packet.Name
	}).First()
	if !loaded {
		logger.Error("No session found for ", e.Packet.Name)
		return client.SendPacket(e.Conn, &packet.PacketEndSideConnectionClosed{Uuid: e.Packet.Uuid})
	}

	conn, err := net.Dial("tcp", session.Value)
	if err != nil {
		return client.SendPacket(e.Conn, &packet.PacketEndSideConnectionClosed{Uuid: e.Packet.Uuid})
	}
	client.BackendConnections.Store(e.Packet.Uuid, comm.NewConnWithParentCtx(client.Ctx, conn))
	go pattern.NewConfigSelectContextAndChannel[*packet.PacketProxyData]().
		WithCtx(client.Ctx).
		WithGoroutine(func(packetCh chan *packet.PacketProxyData) {
			defer func() {
				client.SendPacket(e.Conn, &packet.PacketEndSideConnectionClosed{Uuid: e.Packet.Uuid})
				conn.Close()
			}()
			for {
				bytes, err := comm.ReadForBytes(conn)
				if err != nil {
					if client.Ctx.Err() != nil || lang.IsNetClose(err) {
						client.ProxyConnections.Delete(e.Packet.Uuid)
						client.BackendConnections.Delete(e.Packet.Uuid)
						return
					}
					continue
				}

				packetCh <- &packet.PacketProxyData{
					Name: e.Packet.Name,
					Uuid: e.Packet.Uuid,
					Data: bytes,
				}
			}
		}).
		WithChannelHandler(func(p *packet.PacketProxyData) {
			event.Publish(client.EventBus, &event.EventPacketProxyDataQueue{Packet: p})
		}).
		Run()
	return true
}

func EventHandlerReceivedPacketEndSideConnectionClosed(e *event.EventReceivedPacket[*packet.PacketEndSideConnectionClosed]) bool {
	client := GetClient()
	if conn, ok := client.BackendConnections.Load(e.Packet.Uuid); ok {
		conn.Close()
	}
	client.BackendConnections.Delete(e.Packet.Uuid)
	return true
}

func EventHandlerPacketProxyDataQueue(e *event.EventPacketProxyDataQueue) bool {
	client := GetClient()
	conn, ok := client.ProxyConnections.Load(e.Packet.Uuid)
	if !ok {
		return true
	}
	if !client.SendPacket(conn, e.Packet) {
		client.ProxyConnections.Delete(e.Packet.Uuid)
	}
	return true
}

func AddClientEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerReceivedPacketProxyNegotiationResponse)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyData)
	event.Subscribe(bus, EventHandlerPacketProxyDataQueue)
	event.Subscribe(bus, EventHandlerReceivedPacketNewEndSideConnection)
	event.Subscribe(bus, EventHandlerReceivedPacketEndSideConnectionClosed)
}
