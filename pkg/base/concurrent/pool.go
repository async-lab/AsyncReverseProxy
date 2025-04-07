package concurrent

import "sync"

type Pool[T any] struct {
	*sync.Pool
}

func NewPool[T any](new func() T) *Pool[T] {
	return &Pool[T]{
		Pool: &sync.Pool{New: func() any { return new() }},
	}
}

func (p *Pool[T]) Get() T {
	return p.Pool.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	p.Pool.Put(x)
}
