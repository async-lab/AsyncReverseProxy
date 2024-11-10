package structure

import (
	"unsafe"

	"club.asynclab/asrp/pkg/base/concurrent"
)

type MetaSyncStructure[T any] struct {
	Lock *concurrent.ReentrantLock
}

func NewMetaSyncStructure[T any]() *MetaSyncStructure[T] {
	return &MetaSyncStructure[T]{
		Lock: concurrent.NewReentrantLock(),
	}
}

func (s *MetaSyncStructure[T]) Compute(f func(v *T)) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	f((*T)(unsafe.Pointer(s)))
}
