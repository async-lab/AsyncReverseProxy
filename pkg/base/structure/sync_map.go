package structure

import (
	"club.asynclab/asrp/pkg/base/container"
	"club.asynclab/asrp/pkg/base/hof"
)

type SyncMap[T1 comparable, T2 comparable] struct {
	*MetaSyncStructure[SyncMap[T1, T2]]
	m map[T1]T2
}

func NewSyncMap[T1 comparable, T2 comparable]() *SyncMap[T1, T2] {
	return &SyncMap[T1, T2]{
		MetaSyncStructure: NewMetaSyncStructure[SyncMap[T1, T2]](),
		m:                 make(map[T1]T2),
	}
}

func (m *SyncMap[T1, T2]) Stream() *hof.Stream[container.Entry[T1, T2]] {
	return hof.NewStreamWithMapWithLocker(m.m, m.Lock)
}

func (m *SyncMap[T1, T2]) Load(key T1) (value T2, ok bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	value, ok = m.m[key]
	return
}

func (m *SyncMap[T1, T2]) Store(key T1, value T2) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	m.m[key] = value
}

func (m *SyncMap[T1, T2]) LoadOrStore(key T1, value T2) (actual T2, loaded bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	actual, loaded = m.m[key]
	if !loaded {
		m.m[key] = value
		actual = value
	}
	return
}

func (m *SyncMap[T1, T2]) Delete(key T1) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	delete(m.m, key)
}

func (m *SyncMap[T1, T2]) LoadAndDelete(key T1) (value T2, loaded bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	value, loaded = m.m[key]
	if loaded {
		delete(m.m, key)
	}
	return
}

func (m *SyncMap[T1, T2]) CompareAndDelete(key T1, value T2) (deleted bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	if v, ok := m.m[key]; ok && v == value {
		delete(m.m, key)
		return true
	}
	return false
}

func (m *SyncMap[T1, T2]) Swap(key T1, value T2) (previous T2, loaded bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	previous, loaded = m.m[key]
	m.m[key] = value
	return
}
func (m *SyncMap[T1, T2]) CompareAndSwap(key T1, old, new T2) (swapped bool) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	if v, ok := m.m[key]; ok && v == old {
		m.m[key] = new
		return true
	}
	return false
}

func (m *SyncMap[T1, T2]) Len() int {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	return len(m.m)
}
