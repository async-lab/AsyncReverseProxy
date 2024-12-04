package server

import (
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/program/session"
)

func EventHandlerReceivedPacketProxyNegotiationRequest(e *event.EventReceivedPacket[*packet.PacketProxyNegotiationRequest]) bool {
	server := GetServer()

	sendResponse := func(success bool, reason string) {
		server.SendPacket(e.Conn, &packet.PacketProxyNegotiationResponse{
			Name:             e.Packet.Name,
			Success:          success,
			Reason:           reason,
			RemoteServerName: e.Packet.RemoteServerName,
		})
	}

	// 检查token
	if e.Packet.Token != server.Config.Server.Token {
		sendResponse(false, "Invalid token")
		return false
	}

	// 检查是否已存在该会话，有的话就跳过Listener创建
	if s, ok := server.Sessions.Load(e.Packet.Name); ok {
		if s.Listener.Addr().String() != e.Packet.FrontendAddress {
			sendResponse(false, "Frontend address is different")
			return false
		}
		s.AddProxyConn(e.Conn, int(e.Packet.Priority), int(e.Packet.Weight))
		return true
	}

	listener, err := net.Listen("tcp", e.Packet.FrontendAddress)
	if err != nil {
		sendResponse(false, err.Error())
		return false
	}

	s := session.NewServerSession(e.Packet.Name, comm.NewListener(listener))
	server.Sessions.Store(e.Packet.Name, s)

	closeCtx := s.AddProxyConn(e.Conn, int(e.Packet.Priority), int(e.Packet.Weight))
	go func() {
		defer server.Sessions.Delete(e.Packet.Name)
		<-closeCtx.Done()
	}()
	sendResponse(true, "")
	s.Start()
	return true
}

func EventHandlerReceivedPacketProxyData(e *event.EventReceivedPacket[*packet.PacketProxyData]) bool {
	server := GetServer()
	if s, ok := server.Sessions.Load(e.Packet.Name); ok {
		if conn, ok := s.EndConns.Load(e.Packet.Uuid); ok {
			conn.Write(e.Packet.Data)
		}
	}
	return true
}

func EventHandlerReceivedPacketEndSideConnectionClosed(e *event.EventReceivedPacket[*packet.PacketEndSideConnectionClosed]) bool {
	server := GetServer()
	if s, ok := server.Sessions.Load(e.Packet.Name); ok {
		if conn, ok := s.EndConns.Load(e.Packet.Uuid); ok {
			conn.Close()
		}
	}
	return true
}

func AddServerEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerReceivedPacketProxyNegotiationRequest)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyData)
	event.Subscribe(bus, EventHandlerReceivedPacketEndSideConnectionClosed)
}
