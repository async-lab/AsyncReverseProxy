package hof

import (
	"sync"

	"club.asynclab/asrp/pkg/base/container"
)

type Stream[T any] struct {
	locker sync.Locker
	source <-chan T
}

func NewStreamWithLockerNotLock[T any](source <-chan T, locker sync.Locker) *Stream[T] {
	return &Stream[T]{
		locker: locker,
		source: source,
	}
}

func NewStreamWithLocker[T any](source <-chan T, locker sync.Locker) *Stream[T] {
	locker.Lock()
	return NewStreamWithLockerNotLock(source, locker)
}

func NewStream[T any](source <-chan T) *Stream[T] {
	return NewStreamWithLocker(source, &sync.Mutex{})
}

func NewStreamWithSliceWithLocker[T any](slice []T, locker sync.Locker) *Stream[container.Wrapper[T]] {
	source := make(chan container.Wrapper[T])
	s := NewStreamWithLocker(source, locker)
	go func() {
		defer close(source)
		for i := range slice {
			source <- *container.NewWrapperWithPtr(&slice[i])
		}
	}()
	return s
}

func NewStreamWithSlice[T any](slice []T) *Stream[container.Wrapper[T]] {
	return NewStreamWithSliceWithLocker(slice, &sync.Mutex{})
}

func NewStreamWithMapWithLocker[T1 comparable, T2 any](m map[T1]T2, locker sync.Locker) *Stream[container.Entry[T1, T2]] {
	source := make(chan container.Entry[T1, T2])
	s := NewStreamWithLocker(source, locker)
	go func() {
		defer close(source)
		for key, value := range m {
			source <- *container.NewEntry(key, value)
		}
	}()
	return s
}

func NewStreamWithMap[T1 comparable, T2 any](m map[T1]T2) *Stream[container.Entry[T1, T2]] {
	return NewStreamWithMapWithLocker(m, &sync.Mutex{})
}

func (s *Stream[T]) Filter(predicate func(T) bool) *Stream[T] {
	out := make(chan T)
	go func() {
		defer close(out)
		for item := range s.source {
			if predicate(item) {
				out <- item
			}
		}
	}()

	return NewStreamWithLockerNotLock(out, s.locker)
}

// 残疾版本的map
func (s *Stream[T]) Map(f func(T) T) *Stream[T] {
	out := make(chan T)
	go func() {
		defer close(out)
		for item := range s.source {
			out <- f(item)
		}
	}()
	return NewStreamWithLockerNotLock(out, s.locker)
}

//----------------------------------------------------------------------------------------------------

func (s *Stream[T]) ForEach(action func(T)) {
	defer s.locker.Unlock()
	for item := range s.source {
		action(item)
	}
}

func (s *Stream[T]) Range(action func(T) bool) {
	defer s.locker.Unlock()
	for item := range s.source {
		if !action(item) {
			break
		}
	}
}

func (s *Stream[T]) Collect() []T {
	defer s.locker.Unlock()
	var result []T
	for item := range s.source {
		result = append(result, item)
	}
	return result
}

func (s *Stream[T]) First() (value T, ok bool) {
	defer s.locker.Unlock()
	value, ok = <-s.source
	return
}

func (s *Stream[T]) Last() (value T, ok bool) {
	defer s.locker.Unlock()
	for item := range s.source {
		value = item
		ok = true
	}
	return value, ok
}

func (s *Stream[T]) Max(comparator func(bigger T, smaller T) bool) (T, bool) {
	defer s.locker.Unlock()
	max, ok := <-s.source
	for item := range s.source {
		if comparator(item, max) {
			max = item
		}
	}
	return max, ok
}

func (s *Stream[T]) Min(comparator func(smaller T, bigger T) bool) (T, bool) {
	defer s.locker.Unlock()
	min, ok := s.First()
	for item := range s.source {
		if comparator(item, min) {
			min = item
		}
	}
	return min, ok
}

func (s *Stream[T]) IsEmpty() bool {
	defer s.locker.Unlock()
	_, ok := <-s.source
	return !ok
}

func (s *Stream[T]) Len() int {
	defer s.locker.Unlock()
	var count int
	for range s.source {
		count++
	}
	return count
}
