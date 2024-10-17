package client

import (
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/pattern"
	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/pkg/util"
)

func EventHandlerReceivedPacketProxyNegotiationResponse(e *event.EventReceivedPacket[*packet.PacketProxyNegotiationResponse]) bool {
	client := GetClient()

	if !e.Packet.Success {
		client.Sessions.Delete(e.Packet.Name)
	}

	logger.Info("Proxy negotiation ", e.Packet.Name, " response: ", e.Packet.Success)

	return e.Packet.Success
}

func EventHandlerReceivedPacketProxyConnectionRequest(e *event.EventReceivedPacket[*packet.PacketProxyConnectionRequest]) bool {
	return true
}

func EventHandlerReceivedPacketProxyData(e *event.EventReceivedPacket[*packet.PacketProxyData]) bool {
	client := GetClient()

	client.ProxyConnections.LoadOrStore(e.Packet.Uuid, e.Conn)

	client.BackendConnections.Compute(func(s *structure.SyncMap[string, net.Conn]) {
		conn, loaded := s.Load(e.Packet.Uuid)
		if !loaded {
			backendAddress, loaded := client.Sessions.Load(e.Packet.Name)
			if !loaded {
				logger.Error("No session found for ", e.Packet.Name)
				return
			}

			bConn, err := net.Dial("tcp", backendAddress)
			if err != nil {
				logger.Error("Error connecting to remote server: ", err)
				return
			}
			conn = bConn
			go pattern.NewConfigSelectContextAndChannel[*packet.PacketProxyData]().
				WithCtx(client.Ctx).
				WithGoroutine(func(packetCh chan *packet.PacketProxyData) {
					defer conn.Close()
					for {
						bytes, err := comm.ReadForBytes(conn)
						if err != nil {
							if client.Ctx.Err() != nil || util.IsNetClose(err) {
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
					go event.Publish(client.EventBus, &event.EventPacketProxyDataQueue{Packet: p})
				}).
				Run()
			s.Store(e.Packet.Uuid, bConn)
		}

		go func() {
			if _, err := conn.Write(e.Packet.Data); err != nil {
				client.BackendConnections.Delete(e.Packet.Uuid)
			}
		}()
	})

	return true
}

func EventHandlerPacketProxyDataQueue(e *event.EventPacketProxyDataQueue) bool {
	client := GetClient()

	conn, ok := client.ProxyConnections.Load(e.Packet.Uuid)
	if !ok {
		return false
	}

	if !client.SendPacket(conn, e.Packet) {
		client.ProxyConnections.Delete(e.Packet.Uuid)
	}

	return true
}

func AddClientEventHandler(bus *event.EventBus) {
	event.Subscribe(bus, EventHandlerReceivedPacketProxyNegotiationResponse)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyConnectionRequest)
	event.Subscribe(bus, EventHandlerReceivedPacketProxyData)
	event.Subscribe(bus, EventHandlerPacketProxyDataQueue)
}
