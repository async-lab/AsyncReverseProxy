package client

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"club.asynclab/asrp/config"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/packet"
	"club.asynclab/asrp/pkg/util"
)

// 包装好的连接消费器
//
// 可以传一个需要操作连接的函数，不用关心连接的建立和关闭
//
// 函数的返回值决定了是否需要继续触发事件
func (client *Client) Consume(remoteAddress string, consumer func(net.Conn) bool) {
	conn, err := tls.Dial("tcp", remoteAddress, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		logger.Error("Error connecting to remote server: ", err)
		return
	}
	defer func() {
		comm.SendPacket(conn, &packet.PacketEnd{})
		conn.Close()
	}()
	connCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if ok := consumer(conn); ok {
		client.EmitEvent(conn, connCtx)
	}
}

func (client *Client) Hello(remoteAddress string) {
	client.Consume(remoteAddress, func(conn net.Conn) bool {
		return client.SendPacket(conn, &packet.PacketHello{})
	})
}

func (client *Client) StartProxy(remoteServer *config.ConfigItemRemoteServer, name string, frontendAddress string, backendAddress string) {
	for {
		select {
		case <-client.Ctx.Done():
			return
		default:
			client.Consume(remoteServer.Address, func(conn net.Conn) bool {
				ok := client.SendPacket(conn, &packet.PacketProxyNegotiationRequest{
					Name:            name,
					FrontendAddress: frontendAddress,
					Token:           remoteServer.Token,
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
