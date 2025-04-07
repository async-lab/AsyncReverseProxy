package client

import (
	"context"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/arch/connectors"
	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/program"
	"club.asynclab/asrp/pkg/program/session"
)

var logger = logging.GetLogger()

type Client struct {
	program.MetaProgram
	Config   config.ConfigClient
	Sessions *concurrent.ConcurrentMap[string, *session.ClientSession]
}

func NewClient(ctx context.Context, config config.ConfigClient) *Client {
	client := &Client{
		MetaProgram: *program.NewMetaProgram(ctx),
		Config:      config,
		Sessions:    concurrent.NewSyncMap[string, *session.ClientSession](),
	}
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
	if client.Config.Proxies == nil {
		logger.Error("Proxies is nil")
		return false
	}
	if client.Config.Remotes == nil {
		logger.Error("Remotes is nil")
		return false
	}
	for _, proxy := range client.Config.Proxies {
		if len(proxy.Remotes) == 0 {
			logger.Error("[", proxy.Name, "]'s Remotes is empty")
			return false
		}

		for _, remoteName := range proxy.Remotes {
			found := false
			for _, remote := range client.Config.Remotes {
				if remote.Name == remoteName {
					found = true
					break
				}
			}

			if !found {
				logger.Error("Remote not found: ", remoteName)
				return false
			}
		}

		if proxy.Weight <= 0 {
			proxy.Weight = 1
		}
	}
	return true
}

func (client *Client) StartProxySession(remoteConfig config.ConfigItemRemote, proxyConfig config.ConfigItemProxy) {
	connector, err := connectors.NewConnectorTLS(client.Ctx, remoteConfig, proxyConfig)
	if err != nil {
		return
	}

	channel.ConsumeWithCtx(client.Ctx, connector.GetChanSendForwarder(), func(f arch.IForwarder) bool {
		session, err := session.NewClientSession(client.Ctx, f, proxyConfig.Backend)
		if err != nil {
			return false
		}
		client.Sessions.LoadOrStore(proxyConfig.Name, session)
		go func() {
			<-session.Ctx.Done()
			client.Sessions.Delete(proxyConfig.Name)
		}()
		return true
	})
}

func (client *Client) StartProxyFromConfig() {
	for _, proxyConfig := range client.Config.Proxies {
		for _, remoteConfig := range client.Config.Remotes {
			found := false
			for _, remoteName := range proxyConfig.Remotes {
				if remoteConfig.Name == remoteName {
					found = true
					break
				}
			}

			if found {
				go client.StartProxySession(remoteConfig, proxyConfig)
			}
		}
	}
}

func (client *Client) Run() {
	if !client.CheckConfig() {
		return
	}

	logger.Info("Client starting...")

	client.StartProxyFromConfig()

	<-client.Ctx.Done()
}
