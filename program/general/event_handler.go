package general

import (
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/program"
)

var logger = logging.GetLogger()

func EventHandlerHello(event event.EventHello) bool {
	logger.Info("Hello from: ", event.Conn.RemoteAddr())
	program.SendPacket(event.Conn, &packet.PacketHello{})
	return false
}

func EventHandlerEnd(event event.EventEnd) bool {
	return false
}

func EventHandlerUnknown(event event.EventUnknown) bool {
	event.Conn.Write([]byte("Hello, world!"))
	return false
}

func AddGeneralEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerHello)
	event.Subscribe(bus, EventHandlerEnd)
	event.Subscribe(bus, EventHandlerUnknown)
}
