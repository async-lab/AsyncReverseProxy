package structure

import "club.asynclab/asrp/pkg/base/hof"

type HashSet[T comparable] struct {
	m map[T]struct{}
}

func NewHashSet[T comparable]() *HashSet[T] {
	return &HashSet[T]{
		m: make(map[T]struct{}),
	}
}

func (hs *HashSet[T]) Store(value T) {
	hs.m[value] = struct{}{}
}

func (hs *HashSet[T]) Delete(value T) {
	delete(hs.m, value)
}

func (hs *HashSet[T]) Contains(value T) (ok bool) {
	_, ok = hs.m[value]
	return
}

func (hs *HashSet[T]) Len() int {
	return len(hs.m)
}

func (hs *HashSet[T]) Empty() {
	hs.m = make(map[T]struct{})
}

func (hs *HashSet[T]) Stream() *hof.Stream[T] {
	return hof.NewStreamWithMapKey(hs.m)
}
