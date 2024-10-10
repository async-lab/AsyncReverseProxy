package structure

import "sync"

type SyncMap[T1 any, T2 any] struct {
	SyncMap sync.Map
}

func (s *SyncMap[T1, T2]) Load(key T1) (value T2, ok bool) {
	v, ok := s.SyncMap.Load(key)
	if ok {
		value = v.(T2)
	}
	return
}

func (s *SyncMap[T1, T2]) Store(key T1, value T2) {
	s.SyncMap.Store(key, value)
}

func (s *SyncMap[T1, T2]) Delete(key T1) {
	s.SyncMap.Delete(key)
}
