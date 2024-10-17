package concurrent

import (
	"sync"

	"club.asynclab/asrp/pkg/util"
)

type ReentrantLock struct {
	mu        *sync.Mutex
	cond      *sync.Cond
	owner     int
	holdCount int
}

func NewReentrantLock() *ReentrantLock {
	mu := &sync.Mutex{}
	return &ReentrantLock{
		mu:   mu,
		cond: sync.NewCond(mu),
	}
}

func (rl *ReentrantLock) Lock() {
	me := util.GetGoroutineId()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.owner == me {
		rl.holdCount++
		return
	}
	for rl.holdCount != 0 {
		rl.cond.Wait()
	}
	rl.owner = me
	rl.holdCount = 1
}

func (rl *ReentrantLock) Unlock() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.holdCount == 0 || rl.owner != util.GetGoroutineId() {
		panic("illegalMonitorStateError")
	}
	rl.holdCount--
	if rl.holdCount == 0 {
		rl.cond.Signal()
	}
}
