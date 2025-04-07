package concurrent

import (
	"sync"

	"club.asynclab/asrp/pkg/base/lang"
)

type ReentrantRWLock struct {
	mu         *sync.Mutex
	rCond      *sync.Cond
	rOwners    map[int]int
	rHoldCount int
	wCond      *sync.Cond
	wOwner     int
	wHoldCount int
}

func NewReentrantRWLock() *ReentrantRWLock {
	mu := &sync.Mutex{}
	return &ReentrantRWLock{
		mu:         mu,
		rCond:      sync.NewCond(mu),
		rOwners:    make(map[int]int),
		rHoldCount: 0,
		wCond:      sync.NewCond(mu),
		wOwner:     0,
		wHoldCount: 0,
	}
}

func (rw *ReentrantRWLock) RLock() {
	me := lang.GetGoroutineId()

	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.wOwner == me {
		rw.rOwners[me]++
		rw.rHoldCount++
		return
	}

	for rw.wHoldCount > 0 && rw.wOwner != me {
		rw.wCond.Wait()
	}

	rw.rOwners[me]++
	rw.rHoldCount++
}

func (rw *ReentrantRWLock) RUnlock() {
	me := lang.GetGoroutineId()

	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.rHoldCount == 0 || rw.rOwners[me] == 0 {
		panic("unlock of unlocked lock")
	}

	rw.rOwners[me]--
	rw.rHoldCount--
	if rw.rOwners[me] == 0 {
		delete(rw.rOwners, me)
	}
	if rw.rHoldCount == 0 {
		rw.rCond.Signal()
	}
}

func (rw *ReentrantRWLock) Lock() {
	me := lang.GetGoroutineId()

	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.wOwner == me {
		rw.wHoldCount++
		return
	}

	for (rw.rOwners[me] == 0 || rw.rHoldCount > rw.rOwners[me]) && (rw.wHoldCount > 0 || rw.rHoldCount > 0) {
		rw.rCond.Wait()
	}

	rw.wOwner = me
	rw.wHoldCount = 1
}

func (rw *ReentrantRWLock) Unlock() {
	me := lang.GetGoroutineId()

	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.wHoldCount == 0 || rw.wOwner != me {
		panic("unlock of unlocked lock")
	}

	rw.wHoldCount--
	if rw.wHoldCount == 0 {
		rw.wOwner = 0
		rw.rCond.Signal()
		rw.wCond.Broadcast()
	}
}
