package container

type Entry[T1 any, T2 any] struct {
	key   *T1
	value *T2
}

func NewEntry[T1 any, T2 any](key T1, value T2) *Entry[T1, T2] {
	return &Entry[T1, T2]{key: &key, value: &value}
}

func NewEntryWithPtr[T1 any, T2 any](key *T1, value *T2) *Entry[T1, T2] {
	return &Entry[T1, T2]{key: key, value: value}
}

func (e *Entry[T1, T2]) GetKey() T1 {
	return *e.key
}

func (e *Entry[T1, T2]) GetKeyPtr() *T1 {
	return e.key
}

func (e *Entry[T1, T2]) GetValue() T2 {
	return *e.value
}

func (e *Entry[T1, T2]) GetValuePtr() *T2 {
	return e.value
}
