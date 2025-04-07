package ushers

import (
	"context"
	"crypto/tls"

	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/util"
)

func NewUsherTLS(parentCtx context.Context, addr string, token string) (*Usher[comm.TLS], error) {
	cert, err := util.GenerateSelfSignedCert()
	if err != nil {
		return nil, err
	}

	listener, err := tls.Listen("tcp", addr, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		return nil, err
	}

	usher := NewUsher[comm.TLS](comm.NewListenerWithParentCtx(parentCtx, listener), token)
	return usher, nil
}
