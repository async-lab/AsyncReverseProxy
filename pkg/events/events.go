package events

import (
	"net"
	"sync"

	"club.asynclab/asrp/pkg/api"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packets"
	"club.asynclab/asrp/pkg/util"
	"github.com/google/uuid"
)

var logger = logging.GetLogger()

var Manager = api.NewEventManager()

const (
	EventTypeHello api.EventType = iota
	EventTypeProxyStart
	EventTypeProxyConfirm
	EventTypeProxy
)

type EventPayloadProxyStart struct {
	Conn     net.Conn
	Sessions *sync.Map
	Packet   api.IPacket
}

type EventPayloadProxyConfirm struct {
	Conn   net.Conn
	IdChan chan string
	Packet api.IPacket
}
type EventPayloadProxy struct {
	Conn     net.Conn
	Sessions *sync.Map
	Packet   api.IPacket
}

func (payload EventPayloadProxyStart) ToMap() map[string]interface{} {
	return util.StructToMap(payload)
}
func (payload EventPayloadProxyConfirm) ToMap() map[string]interface{} {
	return util.StructToMap(payload)
}
func (payload EventPayloadProxy) ToMap() map[string]interface{} {
	return util.StructToMap(payload)
}

func init() {
	Manager.Subscribe(EventTypeHello, func(event api.Event) {})
	Manager.Subscribe(EventTypeProxyStart, func(event api.Event) {
		if payload, ok := event.Payload.(EventPayloadProxyStart); ok {
			id := uuid.NewString()
			if address, ok := payload.Packet.GetData().Content["address"].(string); ok {
				payload.Sessions.Store(id, address)
				p := &packets.PacketProxyConfirm{
					Data: api.PackData{
						Type: packets.PacketTypeProxyConfirm,
						Content: api.PacketContent{
							"uuid": id,
						},
					},
				}
				comm.SendPacketData(payload.Conn, p)
			}
		}
	})
	Manager.Subscribe(EventTypeProxyConfirm, func(event api.Event) {
		if payload, ok := event.Payload.(EventPayloadProxyConfirm); ok {
			if id, ok := payload.Packet.GetData().Content["uuid"].(string); ok {
				logger.Info("Proxy confirmed: ", id)
				payload.IdChan <- id
			}
		}
	})
	Manager.Subscribe(EventTypeProxy, func(event api.Event) {
		if payload, ok := event.Payload.(EventPayloadProxy); ok {
			if id, ok := payload.Packet.GetData().Content["uuid"].(string); ok {
				if address, ok := payload.Sessions.Load(id); ok {
					for {
						frontendListener, err := net.Listen("tcp", address.(string))
						if err != nil {
							logger.Error("Error connecting to remote server: ", err)
							return
						}
						defer frontendListener.Close()

						logger.Info("Started proxy on: ", address)

						for {
							func() {
								frontendConn, err := frontendListener.Accept()
								if err != nil {
									logger.Error("Error accepting connection: ", err)
								}
								defer frontendConn.Close()

								comm.Proxy(frontendConn, payload.Conn)
							}()
						}
					}
				}
			}
		}
	})
}
