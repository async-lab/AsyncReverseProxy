package container

type Wrapper[T any] struct {
	p *T
}

func NewWrapper[T any](p T) *Wrapper[T] {
	return &Wrapper[T]{p: &p}
}

func NewWrapperWithPtr[T any](p *T) *Wrapper[T] {
	return &Wrapper[T]{p: p}
}

func (w *Wrapper[T]) Get() T {
	return *w.p
}

func (w *Wrapper[T]) GetPtr() *T {
	return w.p
}

func (w *Wrapper[T]) Set(v T) {
	w.p = &v
}

func (w *Wrapper[T]) SetPtr(v *T) {
	w.p = v
}
