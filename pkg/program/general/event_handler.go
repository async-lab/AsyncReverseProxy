package general

import (
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/program"
)

var logger = logging.GetLogger()

func EventHandlerHello(event *event.EventReceivedPacket[*packet.PacketHello]) bool {
	logger.Info("Hello from: ", event.Conn.RemoteAddr())
	program.Program.SendPacket(event.Conn, &packet.PacketHello{})
	return false
}

func EventHandlerPacketEnd(event *event.EventReceivedPacket[*packet.PacketEnd]) bool {
	return false
}

func EventHandlerPacketUnknown(event *event.EventReceivedPacket[*packet.PacketUnknown]) bool {
	event.Conn.Write([]byte("HTTP/1.1 200 OK\nContent-Type: text/plain; charset=utf-8\nContent-Length: 13\n\nHello, world!"))
	logger.Error("Unknown packet received")
	return false
}

func AddGeneralEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerHello)
	event.Subscribe(bus, EventHandlerPacketEnd)
	event.Subscribe(bus, EventHandlerPacketUnknown)
}
