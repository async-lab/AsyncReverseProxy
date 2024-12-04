package structure

import "github.com/google/uuid"

type SyncIndexMap[T comparable] struct {
	*SyncMap[string, T]
}

func NewSyncIndexMap[T comparable]() *SyncIndexMap[T] {
	return &SyncIndexMap[T]{
		SyncMap: NewSyncMap[string, T](),
	}
}

func (self *SyncIndexMap[T]) Store(value T) (index string) {
	self.Lock.Lock()
	defer self.Lock.Unlock()
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
