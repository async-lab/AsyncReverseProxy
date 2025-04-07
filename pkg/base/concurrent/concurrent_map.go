package concurrent

import (
	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
)

type ConcurrentMap[T1 comparable, T2 any] struct {
	*MetaConcurrentStructure[ConcurrentMap[T1, T2]]
	m map[T1]T2
}

func NewSyncMap[T1 comparable, T2 any]() *ConcurrentMap[T1, T2] {
	return &ConcurrentMap[T1, T2]{
		MetaConcurrentStructure: NewMetaSyncStructure[ConcurrentMap[T1, T2]](),
		m:                       make(map[T1]T2),
	}
}

func (cm *ConcurrentMap[T1, T2]) Stream() *hof.Stream[container.Entry[T1, T2]] {
	return hof.NewStreamFromMapWithLocker(cm.m, cm.Lock)
}

func (cm *ConcurrentMap[T1, T2]) Load(key T1) (value T2, ok bool) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	value, ok = cm.m[key]
	return
}

func (cm *ConcurrentMap[T1, T2]) Store(key T1, value T2) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	cm.m[key] = value
}

func (cm *ConcurrentMap[T1, T2]) LoadOrStore(key T1, value T2) (actual T2, loaded bool) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	actual, loaded = cm.m[key]
	if !loaded {
		cm.m[key] = value
		actual = value
	}
	return
}

func (cm *ConcurrentMap[T1, T2]) LoadOrStoreWithSupplier(key T1, supplier func() T2) (actual T2, loaded bool) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	actual, loaded = cm.m[key]
	if !loaded {
		actual = supplier()
		cm.m[key] = actual
	}
	return
}

func (cm *ConcurrentMap[T1, T2]) Delete(key T1) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	delete(cm.m, key)
}

func (cm *ConcurrentMap[T1, T2]) LoadAndDelete(key T1) (value T2, loaded bool) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	value, loaded = cm.m[key]
	if loaded {
		delete(cm.m, key)
	}
	return
}

func (cm *ConcurrentMap[T1, T2]) Swap(key T1, value T2) (previous T2, loaded bool) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	previous, loaded = cm.m[key]
	cm.m[key] = value
	return
}

func (cm *ConcurrentMap[T1, T2]) Len() int {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	return len(cm.m)
}

func (cm *ConcurrentMap[T1, T2]) Compute(f func(v *ConcurrentMap[T1, T2])) {
	cm.Lock.Lock()
	defer cm.Lock.Unlock()
	f(cm)
}
