package session

import (
	"net"

	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
	"club.asynclab/asrp/pkg/base/structure"
	"club.asynclab/asrp/pkg/comm"
)

type ProxyConnection struct {
	*comm.Conn
	Priority int
	Weight   int
}

type LoadBalancer struct {
	structure.MetaSyncStructure[LoadBalancer]
	connections  *structure.IndexMap[ProxyConnection]
	totalWeights map[int]int // priority -> totalWeight
	currentIndex int
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		connections:       structure.NewIndexMap[ProxyConnection](),
		totalWeights:      make(map[int]int),
		currentIndex:      0,
		MetaSyncStructure: *structure.NewMetaSyncStructure[LoadBalancer](),
	}
}

func (lb *LoadBalancer) AddConn(conn *comm.Conn, priority int, weight int) (uuid string) {
	lb.Lock.Lock()
	defer lb.Lock.Unlock()
	uuid = lb.connections.Store(ProxyConnection{
		Conn:     conn,
		Priority: priority,
		Weight:   weight,
	})
	lb.totalWeights[priority] += weight
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
	if lb.totalWeights[conn.Priority] == 0 {
		delete(lb.totalWeights, conn.Priority)
	}
	lb.connections.Delete(uuid)
}

func (lb *LoadBalancer) Next() (uuid string, conn *comm.Conn, ok bool) {
	lb.Lock.Lock()
	defer lb.Lock.Unlock()

	if lb.connections.Len() == 0 {
		return
	}

	totalWeight, _ok := hof.NewStreamWithMap(lb.totalWeights).Max(func(bigger container.Entry[int, int], smaller container.Entry[int, int]) bool {
		return bigger.GetKey() > smaller.GetKey()
	})

	if !_ok || totalWeight.GetValue() == 0 {
		return
	}

	lb.currentIndex = (lb.currentIndex + 1) % totalWeight.GetValue() // TODO 这里不知道为什么value有时候会为0，先加个判断
	i := lb.currentIndex

	lb.connections.Stream().Filter(func(t container.Entry[string, ProxyConnection]) bool {
		return t.GetValue().Priority == totalWeight.GetKey()
	}).Range(func(t container.Entry[string, ProxyConnection]) bool {
		i -= t.GetValue().Weight
		if i < 0 {
			uuid, conn, ok = t.GetKey(), t.GetValue().Conn, true
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
