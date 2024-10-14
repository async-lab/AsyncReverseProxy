package client

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/program"
	"club.asynclab/asrp/program/general"
)

var logger = logging.GetLogger()

type Client struct {
	Meta             *program.ProgramMeta
	RemoteAddress    string
	Sessions         *structure.SyncMap[string, string]
	ProxyConnections *structure.SyncMap[string, net.Conn]
}

func (client *Client) GetMeta() *program.ProgramMeta { return client.Meta }

func NewClient(ctx context.Context, remoteAddress string) *Client {
	client := &Client{
		Meta:             program.NewProgramMeta(ctx),
		RemoteAddress:    remoteAddress,
		Sessions:         &structure.SyncMap[string, string]{},
		ProxyConnections: &structure.SyncMap[string, net.Conn]{},
	}
	general.AddGeneralEventHandler(client.Meta.EventBus)
	AddClientEventHandler(client.Meta.EventBus)
	return client
}

func (client *Client) consume(consumer func(net.Conn) bool) {
	conn, err := net.Dial("tcp", client.RemoteAddress)
	if err != nil {
		logger.Error("Error connecting to remote server: ", err)
		return
	}
	defer conn.Close()
	connCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if ok := consumer(conn); ok {
		program.EmitEvent(conn, connCtx)
	}
}

func (client *Client) Hello() {
	client.consume(func(conn net.Conn) bool {
		return program.SendPacket(conn, &packet.PacketHello{})
	})
}

func (client *Client) StartProxy(name string, frontendAddress string, backendAddress string) {
	client.consume(func(conn net.Conn) bool {
		ok := program.SendPacket(conn, &packet.PacketProxyNegotiate{
			Name:            name,
			FrontendAddress: frontendAddress,
		})
		if ok {
			client.Sessions.Store(name, backendAddress)
		}
		return ok
	})
}
