package container

type Entry[T1 any, T2 any] struct {
	Key   T1
	Value T2
}

func NewEntry[T1 any, T2 any](key T1, value T2) *Entry[T1, T2] {
	return &Entry[T1, T2]{Key: key, Value: value}
}
