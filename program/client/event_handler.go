package client

import (
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/program"
)

func EventHandlerProxyConfirm(event event.EventProxyConfirm) bool {
	client, ok := program.Program.(*Client)
	if !ok {
		return true
	}

	if event.Packet.Success {
		logger.Info("Proxy confirmed: ", event.Packet.Name)
	} else {
		logger.Info("Proxy confirmation failed: ", event.Packet.Name)
		client.Sessions.Delete(event.Packet.Name)
	}
	return true
}

func EventHandlerNewProxyConnection(event event.EventNewProxyConnection) bool {
	client, ok := program.Program.(*Client)
	if !ok {
		return true
	}

	client.consume(func(conn net.Conn) bool {
		if address, ok := client.Sessions.Load(event.Packet.Name); ok {
			if ok := program.SendPacket(conn, &packet.PacketProxy{Uuid: event.Packet.Uuid}); ok {
				backendConn, err := net.Dial("tcp", address)
				if err != nil {
					logger.Error("Error connecting to backend server: ", err)
					return false
				}
				defer backendConn.Close()
				comm.Proxy(client.Meta.Ctx, conn, backendConn)
			}
		}
		return false
	})
	return true
}

func AddClientEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerProxyConfirm)
	event.Subscribe(bus, EventHandlerNewProxyConnection)
}
