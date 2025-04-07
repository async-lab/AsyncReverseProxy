package concurrent

import "github.com/google/uuid"

type ConcurrentIndexMap[T any] struct {
	*ConcurrentMap[string, T]
}

func NewSyncIndexMap[T any]() *ConcurrentIndexMap[T] {
	return &ConcurrentIndexMap[T]{
		ConcurrentMap: NewSyncMap[string, T](),
	}
}

func (cim *ConcurrentIndexMap[T]) Store(value T) (index string) {
	cim.Lock.Lock()
	defer cim.Lock.Unlock()
	for {
		i := uuid.NewString()
		_, ok := cim.m[i]
		if ok {
			continue
		} else {
			cim.m[i] = value
			index = i
			break
		}
	}
	return
}

func (cim *ConcurrentIndexMap[T]) Update(index string, value T) {
	cim.ConcurrentMap.Store(index, value)
}

func (sim *ConcurrentIndexMap[T]) Compute(f func(v *ConcurrentIndexMap[T])) {
	sim.Lock.Lock()
	defer sim.Lock.Unlock()
	f(sim)
}
