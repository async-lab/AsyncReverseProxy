package container

type Wrapper[T any] struct {
	p *T
}

func NewWrapper[T any](p T) *Wrapper[T] {
	return &Wrapper[T]{p: &p}
}

func NewWrapperWithPointer[T any](p *T) *Wrapper[T] {
	return &Wrapper[T]{p: p}
}

func (w *Wrapper[T]) Get() T {
	return *w.p
}

func (w *Wrapper[T]) GetPointer() *T {
	return w.p
}

func (w *Wrapper[T]) Set(v T) {
	w.p = &v
}

func (w *Wrapper[T]) SetPointer(v *T) {
	w.p = v
}
