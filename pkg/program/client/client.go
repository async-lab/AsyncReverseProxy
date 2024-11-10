package client

import (
	"context"

	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
	"club.asynclab/asrp/pkg/base/lang"
	"club.asynclab/asrp/pkg/base/structure"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/program"
	"club.asynclab/asrp/pkg/program/general"
)

var logger = logging.GetLogger()

type Client struct {
	program.MetaProgram
	Config             *config.ConfigClient
	Sessions           *structure.SyncMap[container.Entry[string, string], string] // (name, server) -> backend_address
	ProxyConnections   *structure.SyncMap[string, *comm.Conn]                      // uuid -> conn
	BackendConnections *structure.SyncMap[string, *comm.Conn]                      // uuid -> conn
}

func NewClient(ctx context.Context, config *config.ConfigClient) *Client {
	client := &Client{
		MetaProgram:        *program.NewMetaProgram(ctx),
		Config:             config,
		Sessions:           structure.NewSyncMap[container.Entry[string, string], string](),
		ProxyConnections:   structure.NewSyncMap[string, *comm.Conn](),
		BackendConnections: structure.NewSyncMap[string, *comm.Conn](),
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
		p := client.Config.RemoteServers
		lang.Useless(p)
		_, ok := hof.NewStreamWithSlice(client.Config.RemoteServers).Filter(func(c container.Wrapper[*config.ConfigItemRemoteServer]) bool {
			return (*c.Get()).Name == proxy.RemoteServerName
		}).First()
		if !ok {
			logger.Error("RemoteServer not found: ", proxy.RemoteServerName)
			return false
		}

		if proxy.Weight <= 0 {
			proxy.Weight = 1
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
