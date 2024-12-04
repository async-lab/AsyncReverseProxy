package structure

import (
	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
)

type SyncSlice[T any] struct {
	*MetaSyncStructure[SyncSlice[T]]
	l []T
}

func NewSyncSlice[T any]() *SyncSlice[T] {
	return &SyncSlice[T]{
		MetaSyncStructure: NewMetaSyncStructure[SyncSlice[T]](),
		l:                 make([]T, 0),
	}
}

func (l *SyncSlice[T]) Stream() *hof.Stream[container.Wrapper[T]] {
	return hof.NewStreamWithSliceWithLocker(l.l, l.Lock)
}

func (l *SyncSlice[T]) Append(values ...T) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	l.l = append(l.l, values...)
}

func (l *SyncSlice[T]) Prepend(values ...T) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	l.l = append(values, l.l...)
}

func (l *SyncSlice[T]) Insert(index int, values ...T) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	l.l = append(l.l[:index], append(values, l.l[index:]...)...)
}

func (l *SyncSlice[T]) Remove(index int) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	l.l = append(l.l[:index], l.l[index+1:]...)
}

func (l *SyncSlice[T]) GetPtr(index int) (value *T, ok bool) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	if index < 0 || index >= len(l.l) {
		return nil, false
	}
	return &l.l[index], true
}

func (l *SyncSlice[T]) Get(index int) (value T, ok bool) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	p, ok := l.GetPtr(index)
	return *p, ok
}

func (l *SyncSlice[T]) Set(index int, value T) (ok bool) {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	if index < 0 || index >= len(l.l) {
		return
	}
	l.l[index] = value
	ok = true
	return
}

func (l *SyncSlice[T]) Len() int {
	l.Lock.Lock()
	defer l.Lock.Unlock()
	return len(l.l)
}
