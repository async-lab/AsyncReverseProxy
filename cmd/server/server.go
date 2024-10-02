package main

import (
	"net"
	"os"
	"sync"

	"club.asynclab/asrp/pkg/api"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/events"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packets"
)

var logger = logging.GetLogger()

var sessions = sync.Map{}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		p, err := comm.ReceivePacketData(conn)
		if err != nil {
			logger.Error("Error receiving packet: ", err)
			return
		}

		e := api.Event{}
		switch p.GetData().Type {
		case packets.PacketTypeProxyStart:
			e = api.Event{
				Type: events.EventTypeProxyStart,
				Payload: events.EventPayloadProxyStart{
					Conn:     conn,
					Sessions: &sessions,
					Packet:   p,
				},
			}
		case packets.PacketTypeProxy:
			e = api.Event{
				Type: events.EventTypeProxy,
				Payload: events.EventPayloadProxy{
					Conn:     conn,
					Sessions: &sessions,
					Packet:   p,
				},
			}
		}

		events.Manager.Publish(e)
	}
}

func RunServer(listenAddress string) {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		logger.Error("Error starting server: ", err)
		return
	}
	defer listener.Close()

	logger.Info("Server started, listening on ", listenAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Error accepting connection: ", err)
			continue
		}
		go handleConnection(conn)
	}
}

func main() {
	RunServer(os.Args[1])
}
