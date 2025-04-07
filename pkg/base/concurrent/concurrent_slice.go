package concurrent

import (
	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
)

type ConcurrentSlice[T any] struct {
	*MetaConcurrentStructure[ConcurrentSlice[T]]
	l []T
}

func NewSyncSlice[T any]() *ConcurrentSlice[T] {
	return &ConcurrentSlice[T]{
		MetaConcurrentStructure: NewMetaSyncStructure[ConcurrentSlice[T]](),
		l:                       make([]T, 0),
	}
}

func (cs *ConcurrentSlice[T]) Stream() *hof.Stream[container.Wrapper[T]] {
	return hof.NewStreamFromSliceWithLocker(cs.l, cs.Lock)
}

func (cs *ConcurrentSlice[T]) Append(values ...T) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	cs.l = append(cs.l, values...)
}

func (cs *ConcurrentSlice[T]) Prepend(values ...T) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	cs.l = append(values, cs.l...)
}

func (cs *ConcurrentSlice[T]) Insert(index int, values ...T) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	cs.l = append(cs.l[:index], append(values, cs.l[index:]...)...)
}

func (cs *ConcurrentSlice[T]) Remove(index int) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	cs.l = append(cs.l[:index], cs.l[index+1:]...)
}

func (cs *ConcurrentSlice[T]) GetPtr(index int) (value *T, ok bool) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	if index < 0 || index >= len(cs.l) {
		return nil, false
	}
	return &cs.l[index], true
}

func (cs *ConcurrentSlice[T]) Get(index int) (value T, ok bool) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	p, ok := cs.GetPtr(index)
	return *p, ok
}

func (cs *ConcurrentSlice[T]) Set(index int, value T) (ok bool) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	if index < 0 || index >= len(cs.l) {
		return
	}
	cs.l[index] = value
	ok = true
	return
}

func (cs *ConcurrentSlice[T]) Len() int {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	return len(cs.l)
}

func (cs *ConcurrentSlice[T]) Compute(f func(v *ConcurrentSlice[T])) {
	cs.Lock.Lock()
	defer cs.Lock.Unlock()
	f(cs)
}
