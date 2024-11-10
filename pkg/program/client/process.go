package client

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/config"
	"club.asynclab/asrp/pkg/packet"
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
		client.EmitEventReceivePacket(conn, connCtx)
	}
}

func (client *Client) Hello(remoteAddress string) {
	client.Consume(remoteAddress, func(conn net.Conn) bool {
		return client.SendPacket(conn, &packet.PacketHello{})
	})
}

func (client *Client) StartProxy(remoteServer config.ConfigItemRemoteServer, proxy config.ConfigItemProxy) {
	for {
		select {
		case <-client.Ctx.Done():
			return
		default:
			client.Consume(remoteServer.Address, func(conn net.Conn) bool {
				ok := client.SendPacket(conn, &packet.PacketProxyNegotiationRequest{
					Name:             proxy.Name,
					FrontendAddress:  proxy.FrontendAddress,
					Priority:         proxy.Priority,
					Weight:           proxy.Weight,
					Token:            remoteServer.Token,
					RemoteServerName: proxy.RemoteServerName,
					BackendAddress:   proxy.BackendAddress,
				})
				return ok
			})
		}

		client.Sessions.Delete(*container.NewEntry(proxy.Name, proxy.RemoteServerName))

		logger.Info("[", proxy.Name, "] -> [", proxy.RemoteServerName, "] closed, retrying in ", config.SleepTime, " seconds...")
		time.Sleep(time.Duration(config.SleepTime) * time.Second)
	}
}

func (client *Client) StartProxyFromConfig() {
	for _, proxy := range client.Config.Proxies {
		remoteServer, _ := hof.NewStreamWithSlice(client.Config.RemoteServers).Filter(func(w container.Wrapper[*config.ConfigItemRemoteServer]) bool {
			return (*w.Get()).Name == proxy.RemoteServerName
		}).First()
		go client.StartProxy(*remoteServer.Get(), *proxy)
	}
}
