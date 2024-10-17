package structure

import "sync"

type SyncQueue[T any] struct {
	MetaSyncStructure[SyncQueue[T]]
	DataQueue []T
	cond      *sync.Cond
}

func NewSyncQueue[T any]() *SyncQueue[T] {
	meta := NewMetaSyncStructure[SyncQueue[T]]()
	return &SyncQueue[T]{
		MetaSyncStructure: *meta,
		DataQueue:         make([]T, 0),
		cond:              sync.NewCond(meta.Lock),
	}
}

func (q *SyncQueue[T]) Push(data T) {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	q.DataQueue = append(q.DataQueue, data)
	q.cond.Signal()

}

func (q *SyncQueue[T]) Pop() (value T, ok bool) {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	if len(q.DataQueue) == 0 {
		ok = false
		return
	}

	value = q.DataQueue[0]
	q.DataQueue = q.DataQueue[1:]
	ok = true
	return
}

func (q *SyncQueue[T]) PopWithWait() (value T) {
	q.Lock.Lock()
	defer q.Lock.Unlock()
	for len(q.DataQueue) == 0 {
		q.cond.Wait()
	}

	value = q.DataQueue[0]
	q.DataQueue = q.DataQueue[1:]
	return
}

func (q *SyncQueue[T]) Len() int {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	return len(q.DataQueue)
}

func (q *SyncQueue[T]) Empty() bool {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	return len(q.DataQueue) == 0
}
