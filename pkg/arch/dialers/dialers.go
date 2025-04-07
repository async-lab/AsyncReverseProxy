package dialers

import (
	"context"
	"fmt"

	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/comm"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
)

var logger = logging.GetLogger()

type IDialerImpl interface {
	Dial(ctx context.Context, addr string) (*comm.Conn, error)
}

type Dialer[T comm.Protocol] struct {
	impl         IDialerImpl
	addr         string
	ctx          context.Context
	ctxCancel    context.CancelFunc
	conns        *concurrent.ConcurrentIndexMap[*comm.Conn]
	senderPacket *channel.SafeSender[packet.IPacket]
}

func NewDialer[T comm.Protocol](parentCtx context.Context, impl IDialerImpl, addr string) *Dialer[T] {
	if impl == nil {
		panic("impl is nil")
	}

	ctx, cancel := context.WithCancel(parentCtx)
	dialer := &Dialer[T]{
		impl:         impl,
		addr:         addr,
		ctx:          ctx,
		ctxCancel:    cancel,
		conns:        concurrent.NewSyncIndexMap[*comm.Conn](),
		senderPacket: channel.NewSafeSenderWithParentCtxAndSize[packet.IPacket](ctx, 16),
	}

	return dialer
}

func (dialer *Dialer[T]) GetChanSendPacket() <-chan packet.IPacket {
	return dialer.senderPacket.GetChan()
}

func (dialer *Dialer[T]) GetCtx() context.Context {
	return dialer.ctx
}

func (dialer *Dialer[T]) Close() error {
	dialer.ctxCancel()
	return nil
}

// --------------------------------------------------------------------

func (dialer *Dialer[T]) writeData(pkt *packet.PacketProxyData) bool {
	conn, ok := dialer.conns.Load(pkt.Uuid)
	if !ok {
		return false
	}
	_, err := conn.Write(pkt.Data)
	return err == nil
}

func (dialer *Dialer[T]) handleConnection(uuid string, conn *comm.Conn) {
	defer conn.Close()

	defer dialer.conns.Delete(uuid)
	defer dialer.senderPacket.Push(&packet.PacketEndSideConnectionClosed{MetaPacketForConn: packet.MetaPacketForConn{Uuid: uuid}})

	for {
		bytes, err := comm.ReadForBytes(conn)
		if err != nil {
			if conn.GetCtx().Err() != nil {
				return
			}
			continue
		}

		dialer.senderPacket.Push(&packet.PacketProxyData{MetaPacketForConn: packet.MetaPacketForConn{Uuid: uuid}, Data: bytes})
	}
}

func (dialer *Dialer[T]) HandlePacket(pkt packet.IPacket) bool {
	if dialer.ctx.Err() != nil {
		return false
	}

	switch pkt := pkt.(type) { // TODO dialer失败就发送给dispatcher重分配请求
	case *packet.PacketProxyData:
		ok := true

		dialer.conns.Compute(func(v *concurrent.ConcurrentIndexMap[*comm.Conn]) {
			if _, ok := v.Load(pkt.Uuid); !ok {
				conn, err := dialer.impl.Dial(dialer.ctx, dialer.addr)
				if err != nil {
					ok = false
					logger.Error(fmt.Sprintf("Error dialing: %v", err))
					return
				}
				v.Update(pkt.Uuid, conn)
				go dialer.handleConnection(pkt.Uuid, conn)
			}
		})

		if !ok || !dialer.writeData(pkt) {
			dialer.senderPacket.Push(&packet.PacketEndSideConnectionClosed{MetaPacketForConn: packet.MetaPacketForConn{Uuid: pkt.Uuid}})
		}
	case *packet.PacketEndSideConnectionClosed:
		if conn, ok := dialer.conns.LoadAndDelete(pkt.Uuid); ok {
			conn.Close()
		}
	}

	return true
}
