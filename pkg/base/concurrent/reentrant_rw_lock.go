package concurrent

import "sync"

//TODO 写得有问题，
//TODO 一是没有写好读锁不占
//TODO 二是没有锁升级
type ReentrantRWLock struct {
	rLock      *sync.Mutex
	rCond      *sync.Cond
	rOwner     int
	rHoldCount int
	wLock      *sync.Mutex
	wCond      *sync.Cond
	wOwner     int
	wHoldCount int
}

func NewReentrantRWLock() *ReentrantRWLock {
	rLock := &sync.Mutex{}
	wLock := &sync.Mutex{}
	return &ReentrantRWLock{
		rLock: rLock,
		rCond: sync.NewCond(rLock),
		wLock: wLock,
		wCond: sync.NewCond(wLock),
	}
}

func (rw *ReentrantRWLock) RLock() {

}

func (rw *ReentrantRWLock) RLocker() *sync.Mutex {
	return rw.rLock
}

func (rw *ReentrantRWLock) RUnlock() {
}

func (rw *ReentrantRWLock) Lock() {
}

func (rw *ReentrantRWLock) Unlock() {
}
