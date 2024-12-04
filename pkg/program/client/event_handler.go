package client

import (
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
)

func EventHandlerReceivedPacketProxyNegotiationResponse(e *event.EventReceivedPacket[*packet.PacketProxyNegotiationResponse]) bool {
	if e.Packet.Success {
		logger.Info("[", e.Packet.Name, "] -> [", e.Packet.RemoteServerName, "] successfully negotiated")
	} else {
		logger.Error("[", e.Packet.Name, "] -> [", e.Packet.RemoteServerName, "] negotiation failed")
		logger.Error("Reason: ", e.Packet.Reason)
	}

	return e.Packet.Success
}

func EventHandlerReceivedPacketProxyData(e *event.EventReceivedPacket[*packet.PacketProxyData]) bool {
	client := GetClient()

	s, ok := client.Sessions.Load(e.Packet.Name)
	if !ok {
		return false
	}

	conn, loaded := s.EndConns.Load(e.Packet.Uuid)
	if !loaded {
		return client.SendPacket(e.Conn, &packet.PacketEndSideConnectionClosed{Name: e.Packet.Name, Uuid: e.Packet.Uuid})
	}

	if _, err := conn.Write(e.Packet.Data); err != nil {
		return client.SendPacket(e.Conn, &packet.PacketEndSideConnectionClosed{Name: e.Packet.Name, Uuid: e.Packet.Uuid})
	}

	return true
}

func EventHandlerReceivedPacketNewEndSideConnection(e *event.EventReceivedPacket[*packet.PacketNewEndSideConnection]) bool {
	client := GetClient()
	s, ok := client.Sessions.Load(e.Packet.Name)
	if !ok {
		return false
	}

	return s.AcceptFrontendConnection(e.Packet.Uuid, e.Conn)
}

func EventHandlerReceivedPacketEndSideConnectionClosed(e *event.EventReceivedPacket[*packet.PacketEndSideConnectionClosed]) bool {
	client := GetClient()

	s, ok := client.Sessions.Load(e.Packet.Name)
	if !ok {
		return false
	}

	if conn, ok := s.EndConns.Load(e.Packet.Uuid); ok {
		conn.Close()
	}
	return true
}

func AddClientEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerReceivedPacketProxyNegotiationResponse)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyData)
	event.Subscribe(bus, EventHandlerReceivedPacketNewEndSideConnection)
	event.Subscribe(bus, EventHandlerReceivedPacketEndSideConnectionClosed)
}
