package structure

import (
	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
	"github.com/google/uuid"
)

type IndexMap[T comparable] struct {
	m map[string]T
}

func NewIndexMap[T comparable]() *IndexMap[T] {
	return &IndexMap[T]{
		m: make(map[string]T),
	}
}

func (self *IndexMap[T]) Load(key string) (value T, ok bool) {
	value, ok = self.m[key]
	return
}

func (self *IndexMap[T]) Store(value T) (index string) {
	for {
		i := uuid.NewString()
		_, ok := self.m[i]
		if ok {
			continue
		} else {
			self.m[i] = value
			index = i
			break
		}
	}
	return
}

func (self *IndexMap[T]) LoadOrStore(key string, value T) (actual T, loaded bool) {
	actual, loaded = self.m[key]
	if !loaded {
		self.m[key] = value
		actual = value
	}
	return
}

func (self *IndexMap[T]) Delete(key string) {
	delete(self.m, key)
}

func (s *IndexMap[T]) LoadAndDelete(key string) (value T, loaded bool) {
	value, loaded = s.m[key]
	if loaded {
		delete(s.m, key)
	}
	return
}

func (self *IndexMap[T]) CompareAndDelete(key string, value T) (deleted bool) {
	if v, ok := self.m[key]; ok && v == value {
		delete(self.m, key)
		return true
	}
	return false
}

func (self *IndexMap[T]) Swap(key string, value T) (previous T, loaded bool) {
	previous, loaded = self.m[key]
	self.m[key] = value
	return
}
func (self *IndexMap[T]) CompareAndSwap(key string, old, new T) (swapped bool) {
	if v, ok := self.m[key]; ok && v == old {
		self.m[key] = new
		return true
	}
	return false
}
func (self *IndexMap[T]) Range(f func(key string, value T) bool) {
	for k, v := range self.m {
		if !f(k, v) {
			break
		}
	}
}

func (self *IndexMap[T]) Len() int {
	return len(self.m)
}

func (self *IndexMap[T]) Stream() *hof.Stream[container.Entry[string, T]] {
	return hof.NewStreamWithMap(self.m)
}
