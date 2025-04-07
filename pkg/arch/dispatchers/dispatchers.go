package dispatchers

import (
	"context"

	"club.asynclab/asrp/pkg/arch"
	"club.asynclab/asrp/pkg/base/channel"
	"club.asynclab/asrp/pkg/base/concurrent"
	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
	"club.asynclab/asrp/pkg/base/structure"
	"club.asynclab/asrp/pkg/logging"
	"club.asynclab/asrp/pkg/packet"
)

var logger = logging.GetLogger()

type Dispatcher struct {
	concurrent.MetaConcurrentStructure[Dispatcher]
	ctx          context.Context
	ctxCancel    context.CancelFunc
	forwarders   *structure.IndexMap[*arch.ForwarderWithValues]
	totalWeights map[int]int // priority -> totalWeight
	currentIndex int
	senderPacket *channel.SafeSender[packet.IPacket]
	connsMap     *concurrent.ConcurrentMap[string, string]
}

func NewDispatcher(parentCtx context.Context) *Dispatcher {
	ctx, cancel := context.WithCancel(parentCtx)
	dispatcher := &Dispatcher{
		ctx:                     ctx,
		ctxCancel:               cancel,
		forwarders:              structure.NewIndexMap[*arch.ForwarderWithValues](),
		totalWeights:            make(map[int]int),
		currentIndex:            0,
		MetaConcurrentStructure: *concurrent.NewMetaSyncStructure[Dispatcher](),
		senderPacket:            channel.NewSafeSenderWithParentCtxAndSize[packet.IPacket](ctx, 16),
		connsMap:                concurrent.NewSyncMap[string, string](),
	}

	return dispatcher
}

func (dispatcher *Dispatcher) GetChanSendPacket() <-chan packet.IPacket {
	return dispatcher.senderPacket.GetChan()
}

func (dispatcher *Dispatcher) GetCtx() context.Context {
	return dispatcher.ctx
}

func (dispatcher *Dispatcher) Close() error {
	dispatcher.ctxCancel()
	return nil
}

// ---------------------------------------------------------------------

func (dispatcher *Dispatcher) packetHandlerMiddleware(forwarderUuid string, f func(packet.IPacket) bool) func(packet.IPacket) bool {
	return func(pkt packet.IPacket) bool {
		switch pkt := pkt.(type) {
		case *packet.PacketProxyData:
			dispatcher.connsMap.LoadOrStore(pkt.Uuid, forwarderUuid)
		}

		return f(pkt)
	}
}

func (dispatcher *Dispatcher) AddForwarder(fwv *arch.ForwarderWithValues) (uuid string) {
	dispatcher.Lock.Lock()
	defer dispatcher.Lock.Unlock()

	uuid = dispatcher.forwarders.Store(fwv)
	dispatcher.totalWeights[fwv.Priority] += fwv.Weight

	go channel.ConsumeWithCtx(dispatcher.GetCtx(), fwv.GetChanSendPacket(), dispatcher.senderPacket.Push)

	return
}

func (dispatcher *Dispatcher) RemoveForwarder(uuid string) {
	dispatcher.Lock.Lock()
	defer dispatcher.Lock.Unlock()

	conn, ok := dispatcher.forwarders.Load(uuid)
	if !ok {
		return
	}
	dispatcher.totalWeights[conn.Priority] -= conn.Weight
	if dispatcher.totalWeights[conn.Priority] == 0 {
		delete(dispatcher.totalWeights, conn.Priority)
	}
	dispatcher.forwarders.Delete(uuid)
	dispatcher.connsMap.Delete(uuid)
}

func (dispatcher *Dispatcher) Next() (uuid string, forwarder arch.IForwarder, ok bool) {
	dispatcher.Lock.Lock()
	defer dispatcher.Lock.Unlock()

	if dispatcher.forwarders.Len() == 0 {
		return
	}

	totalWeight, _ok := hof.NewStreamWithMap(dispatcher.totalWeights).Max(func(bigger container.Entry[int, int], smaller container.Entry[int, int]) bool {
		return bigger.GetKey() > smaller.GetKey()
	})

	if !_ok || totalWeight.GetValue() == 0 {
		return
	}

	dispatcher.currentIndex = (dispatcher.currentIndex + 1) % totalWeight.GetValue()
	i := dispatcher.currentIndex

	dispatcher.forwarders.Stream().Filter(func(t container.Entry[string, *arch.ForwarderWithValues]) bool {
		return t.GetValue().Priority == totalWeight.GetKey()
	}).Range(func(t container.Entry[string, *arch.ForwarderWithValues]) bool {
		i -= t.GetValue().Weight
		if i < 0 {
			uuid, forwarder, ok = t.GetKey(), t.GetValue(), true
			return false
		}
		return true
	})

	return
}

func (dispatcher *Dispatcher) ConsumeNext(f func(uuid string, forwarder arch.IForwarder) bool) (ok bool) {
	uuid, forwarder, _ok := dispatcher.Next()
	if _ok {
		ok = f(uuid, forwarder)
	}
	return ok
}

func (dispatcher *Dispatcher) Len() int {
	dispatcher.Lock.Lock()
	defer dispatcher.Lock.Unlock()

	return dispatcher.forwarders.Len()
}

func (dispatcher *Dispatcher) HandlePacket(pkt packet.IPacket) bool {
	if dispatcher.GetCtx().Err() != nil {
		return false
	}

	switch pkt := pkt.(type) {
	case packet.IPacketForConn:
		if uuid, ok := dispatcher.connsMap.Load(pkt.GetUuid()); ok {
			if v, ok := dispatcher.forwarders.Load(uuid); ok {
				v.HandlePacket(pkt)
				return true
			}
		}
	}

	if uuid, forwarder, ok := dispatcher.Next(); ok {
		dispatcher.packetHandlerMiddleware(uuid, forwarder.HandlePacket)(pkt)
	}

	return true
}
