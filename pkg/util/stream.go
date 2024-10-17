package util

type Stream[T any] struct {
	source <-chan T
}

func NewStream[T any](source <-chan T) *Stream[T] {
	return &Stream[T]{
		source: source,
	}
}

func NewStreamWithSlice[T any](slice []T) *Stream[T] {
	source := make(chan T)
	go func() {
		defer close(source)
		for _, item := range slice {
			source <- item
		}
	}()

	return NewStream(source)
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

	return NewStream(out)
}

func (s *Stream[T]) Map(f func(T) T) *Stream[T] {
	out := make(chan T)
	go func() {
		defer close(out)
		for item := range s.source {
			out <- f(item)
		}
	}()
	return NewStream(out)
}

func (s *Stream[T]) ForEach(action func(T)) {
	for item := range s.source {
		action(item)
	}
}

func (s *Stream[T]) Collect() []T {
	var result []T
	for item := range s.source {
		result = append(result, item)
	}
	return result
}

func (s *Stream[T]) First() (T, bool) {
	value, ok := <-s.source
	return value, ok
}
