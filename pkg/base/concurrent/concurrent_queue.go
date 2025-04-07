package concurrent

import (
	"math"
)

type ConcurrentQueue[T any] struct {
	*MetaConcurrentStructure[ConcurrentQueue[T]]
	l    []T
	head int
	tail int
	size int
	max  int
	// cond *sync.Cond
}

func NewSyncQueueWithInitialCapacityAndMaxCapacity[T any](initialCapacity int, maxCapacity int) *ConcurrentQueue[T] {
	meta := NewMetaSyncStructure[ConcurrentQueue[T]]()
	return &ConcurrentQueue[T]{
		MetaConcurrentStructure: meta,
		l:                       make([]T, initialCapacity),
		head:                    0,
		tail:                    0,
		size:                    0,
		max:                     maxCapacity,
		// cond:              sync.NewCond(meta.Lock),
	}
}

func NewSyncQueueWithInitialCapacity[T any](initialCapacity int) *ConcurrentQueue[T] {
	return NewSyncQueueWithInitialCapacityAndMaxCapacity[T](initialCapacity, math.MaxInt)
}

func NewSyncQueueWithMaxCapacity[T any](maxCapacity int) *ConcurrentQueue[T] {
	return NewSyncQueueWithInitialCapacityAndMaxCapacity[T](16, maxCapacity)
}

func NewSyncQueue[T any]() *ConcurrentQueue[T] {
	return NewSyncQueueWithInitialCapacity[T](16)
}

func (cq *ConcurrentQueue[T]) resize(newCapacity int) {
	newQueue := make([]T, newCapacity)
	if cq.head < cq.tail {
		copy(newQueue, cq.l[cq.head:cq.tail])
	} else {
		n := copy(newQueue, cq.l[cq.head:])
		copy(newQueue[n:], cq.l[:cq.tail])
	}
	cq.l = newQueue
	cq.head = 0
	cq.tail = cq.size
}

func (cq *ConcurrentQueue[T]) Push(data T) bool {
	cq.Lock.Lock()
	defer cq.Lock.Unlock()

	if cq.size == cq.max {
		return false
	}

	if cq.size == len(cq.l) {
		cq.resize(2 * len(cq.l))
	}

	cq.l[cq.tail] = data
	cq.tail = (cq.tail + 1) % len(cq.l)
	cq.size++
	// cq.cond.Signal()
	return true
}

func (cq *ConcurrentQueue[T]) Pop() (value T, ok bool) {
	cq.Lock.Lock()
	defer cq.Lock.Unlock()

	if cq.size == 0 {
		ok = false
		return
	}

	value = cq.l[cq.head]
	cq.head = (cq.head + 1) % len(cq.l)
	cq.size--
	ok = true
	return
}

// func (cq *SyncQueue[T]) PopWithWait() (value T) {
// 	cq.Lock.Lock()
// 	defer cq.Lock.Unlock()

// 	for cq.size == 0 {
// 		cq.cond.Wait()
// 	}

// 	value = cq.l[cq.head]
// 	cq.head = (cq.head + 1) % len(cq.l)
// 	cq.size--
// 	return
// }

func (cq *ConcurrentQueue[T]) GetMax() int {
	cq.Lock.Lock()
	defer cq.Lock.Unlock()

	return cq.max
}

func (cq *ConcurrentQueue[T]) SetMax(max int) {
	cq.Lock.Lock()
	defer cq.Lock.Unlock()

	cq.max = max
}

func (cq *ConcurrentQueue[T]) Len() int {
	cq.Lock.Lock()
	defer cq.Lock.Unlock()

	return cq.size
}

func (cq *ConcurrentQueue[T]) Empty() bool {
	cq.Lock.Lock()
	defer cq.Lock.Unlock()

	return cq.size == 0
}

func (cq *ConcurrentQueue[T]) Compute(f func(v *ConcurrentQueue[T])) {
	cq.Lock.Lock()
	defer cq.Lock.Unlock()
	f(cq)
}
