package acceptors

import (
	"context"

	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/base/pattern"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/packet"
)

type Acceptor[T comm.Protocol] struct {
	listener     *comm.Listener
	conns        *concurrent.ConcurrentIndexMap[*comm.Conn]
	senderPacket *channel.SafeSender[packet.IPacket]
}

func NewAcceptor[T comm.Protocol](listener *comm.Listener) *Acceptor[T] {
	acceptor := &Acceptor[T]{
		listener:     listener,
		conns:        concurrent.NewSyncIndexMap[*comm.Conn](),
		senderPacket: channel.NewSafeSenderWithParentCtxAndSize[packet.IPacket](listener.Ctx, 16),
	}

	acceptor.init()
	return acceptor
}

func (acceptor *Acceptor[T]) GetChanSendPacket() <-chan packet.IPacket {
	return acceptor.senderPacket.GetChan()
}

func (acceptor *Acceptor[T]) GetCtx() context.Context {
	return acceptor.listener.Ctx
}

func (acceptor *Acceptor[T]) Close() error {
	return acceptor.listener.Close()
}

// ------------------------------------------------------------------------------------

func (acceptor *Acceptor[T]) writeData(pkt *packet.PacketProxyData) bool {
	conn, ok := acceptor.conns.Load(pkt.Uuid)
	if !ok {
		return false
	}
	_, err := conn.Write(pkt.Data)
	return err == nil
}

func (acceptor *Acceptor[T]) handleConnection(conn *comm.Conn) {
	defer conn.Close()

	uuid := acceptor.conns.Store(conn)
	defer acceptor.conns.Delete(uuid)
	defer acceptor.senderPacket.Push(&packet.PacketEndSideConnectionClosed{MetaPacketForConn: packet.MetaPacketForConn{Uuid: uuid}})

	for {
		bytes, err := comm.ReadForBytes(conn)
		if err != nil {
			if conn.GetCtx().Err() != nil {
				return
			}
			continue
		}

		acceptor.senderPacket.Push(&packet.PacketProxyData{MetaPacketForConn: packet.MetaPacketForConn{Uuid: uuid}, Data: bytes})
	}
}

func (acceptor *Acceptor[T]) HandlePacket(pkt packet.IPacket) bool {
	if acceptor.GetCtx().Err() != nil {
		return false
	}

	switch pkt := pkt.(type) {
	case *packet.PacketProxyData:
		if !acceptor.writeData(pkt) {
			acceptor.senderPacket.Push(&packet.PacketEndSideConnectionClosed{MetaPacketForConn: packet.MetaPacketForConn{Uuid: pkt.Uuid}})
		}
	case *packet.PacketEndSideConnectionClosed:
		if conn, ok := acceptor.conns.Load(pkt.Uuid); ok {
			conn.Close()
		}
	}
	return true
}

func (acceptor *Acceptor[T]) init() {
	go func() {
		pattern.NewConfigSelectContextAndChannel[*comm.Conn]().
			WithCtx(acceptor.GetCtx()).
			WithGoroutine(func(ch chan *comm.Conn) {
				for {
					conn, err := acceptor.listener.Accept()
					if err != nil {
						if acceptor.GetCtx().Err() != nil {
							return
						}
						continue
					}
					ch <- conn
				}
			}).
			WithChannelHandler(func(conn *comm.Conn) { go acceptor.handleConnection(conn) }).
			Run()
	}()
}
