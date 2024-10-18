package client

import (
	"context"
	"net"

	"club.asynclab/asrp/config"
	"club.asynclab/asrp/pkg/logging"
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
