package concurrent

import "sync/atomic"

type Atomic[T any] struct {
	*atomic.Value
}

func NewAtomic[T any]() *Atomic[T] {
	return &Atomic[T]{Value: &atomic.Value{}}
}

func NewAtomicWithValue[T any](val T) *Atomic[T] {
	ret := NewAtomic[T]()
	ret.Store(val)
	return ret
}

func (a *Atomic[T]) Store(val T) {
	a.Value.Store(val)
}

func (a *Atomic[T]) Load() T {
	return a.Value.Load().(T)
}

func (a *Atomic[T]) Swap(val T) T {
	return a.Value.Swap(val).(T)
}

func (a *Atomic[T]) CompareAndSwap(old, new T) bool {
	return a.Value.CompareAndSwap(old, new)
}
