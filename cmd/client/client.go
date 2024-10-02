package main

import (
	"net"
	"os"
	"time"

	"club.asynclab/asrp/pkg/api"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/events"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packets"
)

var logger = logging.GetLogger()

var idChan = make(chan string, 1)
var proxyChan = make(chan struct{})

var id string

func startProxy(conn net.Conn, frontendAddress string) {
	defer conn.Close()
	p := &packets.PacketProxyStart{
		Data: api.PackData{
			Type: packets.PacketTypeProxyStart,
			Content: map[string]interface{}{
				"address": frontendAddress,
			},
		},
	}
	comm.SendPacketData(conn, p)

	idChan := make(chan string, 1)

	go func() {
		for {
			r, err := comm.ReceivePacketData(conn)
			if err != nil {
				logger.Error("Error receiving packet: ", err)
				return
			}

			e := api.Event{}
			switch r.GetData().Type {
			case packets.PacketTypeProxyConfirm:
				e = api.Event{
					Type: events.EventTypeProxyConfirm,
					Payload: events.EventPayloadProxyConfirm{
						Conn:   conn,
						IdChan: idChan,
						Packet: r,
					},
				}
			}

			events.Manager.Publish(e)
		}
	}()

	id = <-idChan
	close(proxyChan)

	for {
		hello := &packets.PacketHello{Data: api.PackData{
			Type: packets.PacketTypeHello,
			Content: api.PacketContent{
				"uuid": id,
			},
		}}
		comm.SendPacketData(conn, hello)
		time.Sleep(1 * time.Second)
	}
}

func handleProxy(conn net.Conn, backendAddress string) {
	defer conn.Close()
	p := &packets.PacketProxy{
		Data: api.PackData{
			Type:    packets.PacketTypeProxy,
			Content: api.PacketContent{"uuid": id},
		},
	}
	comm.SendPacketData(conn, p)

	for {
		func() {
			backendConn, err := net.Dial("tcp", backendAddress)
			if err != nil {
				logger.Error("Error connecting to local server: ", err)
				return
			}
			defer backendConn.Close()

			comm.Proxy(conn, backendConn)
		}()
	}
}

func newConnection(listenAddress string) net.Conn {
	conn, err := net.Dial("tcp", listenAddress)
	if err != nil {
		logger.Error("Error connecting to server: ", err)
		return nil
	}
	return conn
}

func RunClient(listenAddress string, frontendAddress string, backendAddress string) {
	controlConn := newConnection(listenAddress)
	if controlConn == nil {
		return
	}
	go startProxy(controlConn, frontendAddress)
	<-proxyChan
	streamConn := newConnection(listenAddress)
	if streamConn == nil {
		return
	}
	go handleProxy(streamConn, backendAddress)
	select {}
}

func main() {
	RunClient(os.Args[1], os.Args[2], os.Args[3])
}
