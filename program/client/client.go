package client

import (
	"context"
	"net"
	"time"

	"club.asynclab/asrp/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/structure"
	"club.asynclab/asrp/pkg/util"
	"club.asynclab/asrp/program"
	"club.asynclab/asrp/program/general"
)

var logger = logging.GetLogger()

type Client struct {
	program.MetaProgram
	Config             *config.ConfigClient
	Sessions           *structure.SyncMap[string, string]   // name -> backend_address
	ProxyConnections   *structure.SyncMap[string, net.Conn] // uuid -> conn
	BackendConnections *structure.SyncMap[string, net.Conn] // uuid -> conn
}

func NewClient(ctx context.Context, config *config.ConfigClient) *Client {
	client := &Client{
		MetaProgram:        *program.NewMetaProgram(ctx),
		Config:             config,
		Sessions:           structure.NewSyncMap[string, string](),
		ProxyConnections:   structure.NewSyncMap[string, net.Conn](),
		BackendConnections: structure.NewSyncMap[string, net.Conn](),
	}
	general.AddGeneralEventHandler(client.EventBus)
	AddClientEventHandler(client.EventBus)
	return client
}

func GetClient() *Client {
	client, ok := program.Program.(*Client)
	if !ok {
		panic("Program is not a client")
	}
	return client
}

// 包装好的连接消费器
//
// 可以传一个需要操作连接的函数，不用关心连接的建立和关闭
//
// 函数的返回值决定了是否需要继续触发事件
func (client *Client) consume(remoteAddress string, consumer func(net.Conn) bool) {
	conn, err := net.Dial("tcp", remoteAddress)
	if err != nil {
		logger.Error("Error connecting to remote server: ", err)
		return
	}
	defer conn.Close()
	connCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if ok := consumer(conn); ok {
		client.EmitEvent(conn, connCtx)
	}
}

func (client *Client) Hello(remoteAddress string) {
	client.consume(remoteAddress, func(conn net.Conn) bool {
		return client.SendPacket(conn, &packet.PacketHello{})
	})
}

func (client *Client) StartProxy(remoteServer *config.ConfigItemRemoteServer, name string, frontendAddress string, backendAddress string) {
	for {
		select {
		case <-client.Ctx.Done():
			return
		default:
			client.consume(remoteServer.Address, func(conn net.Conn) bool {
				ok := client.SendPacket(conn, &packet.PacketProxyNegotiationRequest{
					Name:            name,
					FrontendAddress: frontendAddress,
				})
				if ok {
					client.Sessions.Store(name, backendAddress)
				}
				return ok
			})
		}

		logger.Info("Reconnecting ", remoteServer.Name, " in ", config.SleepTime, " seconds")
		time.Sleep(time.Duration(config.SleepTime) * time.Second)
	}
}

func (client *Client) StartProxyFromConfig() {
	for _, proxy := range client.Config.Proxies {
		remoteServer, _ := util.NewStreamWithSlice(client.Config.RemoteServers).Filter(func(c *config.ConfigItemRemoteServer) bool {
			return c.Name == proxy.RemoteServerName
		}).First()
		go client.StartProxy(remoteServer, proxy.Name, proxy.FrontendAddress, proxy.BackendAddress)
	}
}

func (client *Client) CheckConfig() bool {
	if client.Config == nil {
		logger.Error("Config is nil")
		return false
	}
	if client.Config.Proxies == nil {
		logger.Error("Proxies is nil")
		return false
	}
	if client.Config.RemoteServers == nil {
		logger.Error("RemoteServers is nil")
		return false
	}
	for _, proxy := range client.Config.Proxies {
		_, ok := util.NewStreamWithSlice(client.Config.RemoteServers).Filter(func(c *config.ConfigItemRemoteServer) bool {
			return c.Name == proxy.RemoteServerName
		}).First()
		if !ok {
			logger.Error("RemoteServer not found: ", proxy.RemoteServerName)
			return false
		}
	}
	return true
}

func (client *Client) Run() {
	if !client.CheckConfig() {
		return
	}

	logger.Info("Client starting...")

	client.StartProxyFromConfig()

	<-client.Ctx.Done()
}
