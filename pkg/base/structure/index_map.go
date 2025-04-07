package structure

import (
	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
	"github.com/google/uuid"
)

type IndexMap[T any] struct {
	m map[string]T
}

func NewIndexMap[T any]() *IndexMap[T] {
	return &IndexMap[T]{
		m: make(map[string]T),
	}
}

func (im *IndexMap[T]) Load(key string) (value T, ok bool) {
	value, ok = im.m[key]
	return
}

func (im *IndexMap[T]) Store(value T) (index string) {
	for {
		i := uuid.NewString()
		_, ok := im.m[i]
		if ok {
			continue
		} else {
			im.m[i] = value
			index = i
			break
		}
	}
	return
}

func (im *IndexMap[T]) LoadOrStore(key string, value T) (actual T, loaded bool) {
	actual, loaded = im.m[key]
	if !loaded {
		im.m[key] = value
		actual = value
	}
	return
}

func (im *IndexMap[T]) Delete(key string) {
	delete(im.m, key)
}

func (s *IndexMap[T]) LoadAndDelete(key string) (value T, loaded bool) {
	value, loaded = s.m[key]
	if loaded {
		delete(s.m, key)
	}
	return
}

// func (im *IndexMap[T]) CompareAndDelete(key string, value T) (deleted bool) {
// 	if v, ok := im.m[key]; ok && v == value {
// 		delete(im.m, key)
// 		return true
// 	}
// 	return false
// }

func (im *IndexMap[T]) Swap(key string, value T) (previous T, loaded bool) {
	previous, loaded = im.m[key]
	im.m[key] = value
	return
}

// func (im *IndexMap[T]) CompareAndSwap(key string, old, new T) (swapped bool) {
// 	if v, ok := im.m[key]; ok && v == old {
// 		im.m[key] = new
// 		return true
// 	}
// 	return false
// }

func (im *IndexMap[T]) Len() int {
	return len(im.m)
}

func (im *IndexMap[T]) Stream() *hof.Stream[container.Entry[string, T]] {
	return hof.NewStreamWithMap(im.m)
}
