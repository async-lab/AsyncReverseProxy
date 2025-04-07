package connectors

import (
	"context"
	"fmt"
	"time"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/arch/forwarders"
	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/config"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
)

var logger = logging.GetLogger()

type IConnectorImpl interface {
	Connect(ctx context.Context, addr string) (*comm.Conn, error)
}

type Connector[T comm.Protocol] struct {
	impl            IConnectorImpl
	remoteConfig    config.ConfigItemRemote
	proxyConfig     config.ConfigItemProxy
	ctx             context.Context
	ctxCancel       context.CancelFunc
	senderForwarder *channel.SafeSender[arch.IForwarder]
}

func NewConnector[T comm.Protocol](parentCtx context.Context, impl IConnectorImpl, remoteConfig config.ConfigItemRemote, proxyConfig config.ConfigItemProxy) *Connector[T] {
	if impl == nil {
		panic("impl is nil")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	connector := &Connector[T]{
		impl:            impl,
		remoteConfig:    remoteConfig,
		proxyConfig:     proxyConfig,
		ctx:             ctx,
		ctxCancel:       cancel,
		senderForwarder: channel.NewSafeSenderWithParentCtxAndSize[arch.IForwarder](ctx, 16),
	}

	connector.init()
	return connector
}

func (connector *Connector[T]) GetChanSendForwarder() <-chan arch.IForwarder {
	return connector.senderForwarder.GetChan()
}

func (connector *Connector[T]) GetCtx() context.Context {
	return connector.ctx
}

func (connector *Connector[T]) Close() error {
	connector.ctxCancel()
	return nil
}

// ---------------------------------------------------------------------

func (connector *Connector[T]) getTip() string {
	return fmt.Sprintf("[%s -> %s] ", connector.proxyConfig.Name, connector.remoteConfig.Name)
}

func (connector *Connector[T]) initConnection(conn *comm.Conn) error {
	f := forwarders.NewForwarder(conn)
	f.Tip = connector.getTip()

	r := &packet.PacketProxyNegotiationRequest{
		Name:         connector.proxyConfig.Name,
		FrontendAddr: connector.proxyConfig.Frontend,
		Token:        connector.remoteConfig.Token,
		Priority:     connector.proxyConfig.Priority,
		Weight:       connector.proxyConfig.Weight,
	}

	if !f.HandlePacket(r) {
		return fmt.Errorf("error sending connection request")
	}

	switch p := (<-f.GetChanSendPacket()).(type) {
	case *packet.PacketProxyNegotiationResponse:
		if !p.Success {
			return fmt.Errorf("error connecting: %s", p.Reason)
		}
		connector.senderForwarder.Push(f)
		return nil
	case *packet.PacketEnd:
		return fmt.Errorf("end by server")
	default:
		return fmt.Errorf("error receiving connection response")
	}
}

func (connector *Connector[T]) init() {
	go func() {
		connCtx := (context.Context)(nil)
		for {
			if connCtx != nil {
				<-connCtx.Done()
				logger.Info(connector.getTip(), "closed.")
				time.Sleep(time.Duration(config.SleepTime) * time.Second)
			}
			if connector.ctx.Err() != nil {
				connector.Close()
				return
			}
			conn, err := connector.impl.Connect(connector.ctx, connector.remoteConfig.Addr)
			if err != nil {
				logger.Error(connector.getTip(), "Error connecting to remote server: ", err)
				time.Sleep(time.Duration(config.SleepTime) * time.Second)
				continue
			}

			err = connector.initConnection(conn)
			if err != nil {
				conn.Close()
				logger.Error(connector.getTip(), "Error init connection: ", err)
				time.Sleep(time.Duration(config.SleepTime) * time.Second)
				continue
			}
			connCtx = conn.GetCtx()
			logger.Info(connector.getTip(), "established.")
		}
	}()
}
