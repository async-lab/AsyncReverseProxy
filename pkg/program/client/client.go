package client

import (
	"context"

	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
	"club.asynclab/asrp/pkg/base/structure"
	"club.asynclab/asrp/pkg/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/program"
	"club.asynclab/asrp/pkg/program/general"
	"club.asynclab/asrp/pkg/program/session"
)

var logger = logging.GetLogger()

type Client struct {
	program.MetaProgram
	Config   *config.ConfigClient
	Sessions *structure.SyncMap[string, *session.ClientSession]
}

func NewClient(ctx context.Context, config *config.ConfigClient) *Client {
	client := &Client{
		MetaProgram: *program.NewMetaProgram(ctx),
		Config:      config,
		Sessions:    structure.NewSyncMap[string, *session.ClientSession](),
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
		if len(proxy.RemoteServers) == 0 {
			logger.Error("[", proxy.Name, "]'s RemoteServers is empty")
			return false
		}

		ok := !hof.NewStreamWithSlice(proxy.RemoteServers).
			Filter(func(w container.Wrapper[string]) bool {
				ok := !hof.NewStreamWithSlice(client.Config.RemoteServers).
					Filter(func(c container.Wrapper[*config.ConfigItemRemoteServer]) bool { return (*c.Get()).Name == w.Get() }).
					IsEmpty()
				if !ok {
					logger.Error("RemoteServer not found: ", w.Get())
				}
				return ok
			}).IsEmpty()
		if !ok {
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
