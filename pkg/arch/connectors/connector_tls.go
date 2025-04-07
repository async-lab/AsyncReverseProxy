package connectors

import (
	"context"
	"crypto/tls"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/config"
)

type ConnectorImplTLS struct{}

var implTLS = &ConnectorImplTLS{}

func NewConnectorTLS(parentCtx context.Context, remoteConfig config.ConfigItemRemote, proxyConfig config.ConfigItemProxy) (*Connector[comm.TLS], error) {
	connector := NewConnector[comm.TLS](parentCtx, implTLS, remoteConfig, proxyConfig)
	return connector, nil
}

func (impl *ConnectorImplTLS) Connect(ctx context.Context, addr string) (*comm.Conn, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	return comm.NewConnWithParentCtx(ctx, conn), nil
}
