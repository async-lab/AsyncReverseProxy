package structure

type SyncMap[T1 comparable, T2 comparable] struct {
	MetaSyncStructure[SyncMap[T1, T2]]
	m map[T1]T2
}

func NewSyncMap[T1 comparable, T2 comparable]() *SyncMap[T1, T2] {
	return &SyncMap[T1, T2]{
		MetaSyncStructure: *NewMetaSyncStructure[SyncMap[T1, T2]](),
		m:                 make(map[T1]T2),
	}
}

func (s *SyncMap[T1, T2]) Load(key T1) (value T2, ok bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	value, ok = s.m[key]
	return
}

func (s *SyncMap[T1, T2]) Store(key T1, value T2) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.m[key] = value
}

func (s *SyncMap[T1, T2]) LoadOrStore(key T1, value T2) (actual T2, loaded bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	actual, loaded = s.m[key]
	if !loaded {
		s.m[key] = value
		actual = value
	}
	return
}

func (s *SyncMap[T1, T2]) Delete(key T1) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	delete(s.m, key)
}

func (s *SyncMap[T1, T2]) LoadAndDelete(key T1) (value T2, loaded bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	value, loaded = s.m[key]
	if loaded {
		delete(s.m, key)
	}
	return
}

func (s *SyncMap[T1, T2]) CompareAndDelete(key T1, value T2) (deleted bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	if v, ok := s.m[key]; ok && v == value {
		delete(s.m, key)
		return true
	}
	return false
}

func (s *SyncMap[T1, T2]) Swap(key T1, value T2) (previous T2, loaded bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	previous, loaded = s.m[key]
	s.m[key] = value
	return
}
func (s *SyncMap[T1, T2]) CompareAndSwap(key T1, old, new T2) (swapped bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	if v, ok := s.m[key]; ok && v == old {
		s.m[key] = new
		return true
	}
	return false
}
func (s *SyncMap[T1, T2]) Range(f func(key T1, value T2) bool) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	for k, v := range s.m {
		if !f(k, v) {
			break
		}
	}
}

func (s *SyncMap[T1, T2]) Len() int {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	return len(s.m)
}
