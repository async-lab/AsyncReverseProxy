package concurrent

type MetaConcurrentStructure[T any] struct {
	Lock *ReentrantRWLock
}

func NewMetaSyncStructure[T any]() *MetaConcurrentStructure[T] {
	return &MetaConcurrentStructure[T]{
		Lock: NewReentrantRWLock(),
	}
}
