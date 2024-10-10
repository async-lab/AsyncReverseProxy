package program

import (
	"context"
	"net"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/event"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/pkg/util"
)

var logger = logging.GetLogger()

type IProgram interface {
	ReceivePacket(net.Conn) (packet.IPacket, bool)
	SendPacket(net.Conn, packet.IPacket) bool
}

type Server struct {
	Ctx              context.Context
	ListenAddress    string
	EventManager     *event.EventManager
	Sessions         *structure.SyncMap[string, string]
	ProxyConnections *structure.SyncMap[string, net.Conn]
}
type Client struct {
	Ctx              context.Context
	RemoteAddress    string
	EventManager     *event.EventManager
	Sessions         *structure.SyncMap[string, string]
	ProxyConnections *structure.SyncMap[string, net.Conn]
}

func NewServer(ctx context.Context, listenAddress string) *Server {
	server := &Server{
		Ctx:              ctx,
		ListenAddress:    listenAddress,
		EventManager:     event.NewEventManager(),
		Sessions:         &structure.SyncMap[string, string]{},
		ProxyConnections: &structure.SyncMap[string, net.Conn]{},
	}
	server.initEventManager()
	return server
}
func NewClient(ctx context.Context, remoteAddress string) *Client {
	client := &Client{
		Ctx:              ctx,
		RemoteAddress:    remoteAddress,
		EventManager:     event.NewEventManager(),
		Sessions:         &structure.SyncMap[string, string]{},
		ProxyConnections: &structure.SyncMap[string, net.Conn]{},
	}
	client.initEventManager()
	return client
}

func _receivePacket(ctx context.Context, conn net.Conn) (packet.IPacket, bool) {
	r, err := comm.ReceivePacket(conn)
	if err != nil {
		if ctx.Err() != nil {
			return nil, false
		}
		if util.IsConnectionClose(err) {
			logger.Error("Connection closed: ", err)
			return nil, false
		}
		r = &packet.PacketUnknown{Err: err}
	}
	return r, true
}

func _sendPacket(conn net.Conn, p packet.IPacket) bool {
	_, error := comm.SendPacket(conn, p)
	if error != nil {
		logger.Error("Error sending packet: ", error)
		return false
	}
	return true
}

func (server *Server) ReceivePacket(conn net.Conn) (packet.IPacket, bool) {
	return _receivePacket(server.Ctx, conn)
}

func (client *Client) ReceivePacket(conn net.Conn) (packet.IPacket, bool) {
	return _receivePacket(client.Ctx, conn)
}

func (server *Server) SendPacket(conn net.Conn, p packet.IPacket) bool {
	return _sendPacket(conn, p)
}

func (client *Client) SendPacket(conn net.Conn, p packet.IPacket) bool {
	return _sendPacket(conn, p)
}
