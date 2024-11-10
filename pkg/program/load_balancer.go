package program

import (
	"net"

	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
	"club.asynclab/asrp/pkg/base/structure"
	"club.asynclab/asrp/pkg/comm"
)

type ProxyConnection struct {
	Priority int64
	Weight   int64
	Conn     *comm.Conn
}

type LoadBalancer struct {
	structure.MetaSyncStructure[LoadBalancer]
	connections  *structure.IndexMap[*ProxyConnection]
	totalWeights map[int64]int64
	currentIndex int64
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		connections:       structure.NewIndexMap[*ProxyConnection](),
		totalWeights:      make(map[int64]int64),
		currentIndex:      0,
		MetaSyncStructure: *structure.NewMetaSyncStructure[LoadBalancer](),
	}
}

func (lb *LoadBalancer) AddConn(conn *ProxyConnection) (uuid string) {
	lb.Lock.Lock()
	defer lb.Lock.Unlock()
	uuid = lb.connections.Store(conn)
	lb.totalWeights[conn.Priority] += conn.Weight
	return
}

func (lb *LoadBalancer) RemoveConn(uuid string) {
	lb.Lock.Lock()
	defer lb.Lock.Unlock()
	conn, ok := lb.connections.Load(uuid)
	if !ok {
		return
	}
	lb.totalWeights[conn.Priority] -= conn.Weight
	lb.connections.Delete(uuid)
}

func (lb *LoadBalancer) Next() (uuid string, conn *comm.Conn, ok bool) {
	lb.Lock.Lock()
	defer lb.Lock.Unlock()

	if lb.connections.Len() == 0 {
		return
	}

	totalWeight, _ok := hof.NewStreamWithMap(lb.totalWeights).Max(func(bigger container.Entry[int64, int64], smaller container.Entry[int64, int64]) bool {
		return bigger.Key > smaller.Key
	})

	if !_ok || totalWeight.Value == 0 {
		return
	}

	lb.currentIndex = (lb.currentIndex + 1) % totalWeight.Value // TODO 这里不知道为什么value有时候会为0，先加个判断
	i := lb.currentIndex

	lb.connections.Stream().Filter(func(t container.Entry[string, *ProxyConnection]) bool {
		return (*t.Value).Priority == totalWeight.Key
	}).Range(func(t container.Entry[string, *ProxyConnection]) bool {
		i -= (*t.Value).Weight
		if i < 0 {
			uuid, conn, ok = t.Key, t.Value.Conn, true
			return false
		}
		return true
	})

	return
}

func (lb *LoadBalancer) ConsumeNext(f func(uuid string, conn net.Conn) bool) (ok bool) {
	lb.Compute(func(lb *LoadBalancer) {
		uuid, conn, _ok := lb.Next()
		if _ok {
			ok = f(uuid, conn)
		}
	})
	return ok
}

func (lb *LoadBalancer) Len() int {
	lb.Lock.Lock()
	defer lb.Lock.Unlock()
	return lb.connections.Len()
}
