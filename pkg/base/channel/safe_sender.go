package channel

import (
	"context"
	"sync"
)

type SafeSender[T any] struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	ch        chan T
	wg        *sync.WaitGroup
}

func NewSafeSenderWithParentCtxAndSize[T any](parentCtx context.Context, size int) *SafeSender[T] {
	ctx, cancel := context.WithCancel(parentCtx)
	ss := &SafeSender[T]{
		ctx:       ctx,
		ctxCancel: cancel,
		ch:        make(chan T, size),
		wg:        &sync.WaitGroup{},
	}

	ss.wg.Add(1)
	go func() {
		<-ss.GetCtx().Done()
		ss.wg.Done()
	}()

	go func() {
		ss.wg.Wait()
		close(ss.ch)
	}()

	return ss
}

func NewSafeSenderWithSize[T any](size int) *SafeSender[T] {
	return NewSafeSenderWithParentCtxAndSize[T](context.Background(), size)
}

func NewSafeSender[T any]() *SafeSender[T] {
	return NewSafeSenderWithSize[T](0)
}

func (ss *SafeSender[T]) GetCtx() context.Context {
	return ss.ctx
}

func (ss *SafeSender[T]) Close() error {
	ss.ctxCancel()
	return nil
}

func (ss *SafeSender[T]) Push(it T) bool {
	ss.wg.Add(1)
	defer ss.wg.Done()

	if ss.GetCtx().Err() != nil {
		return false
	}
	select {
	case ss.ch <- it:
		return true
	case <-ss.GetCtx().Done():
		return false
	}
}

func (ss *SafeSender[T]) TryPush(it T) bool {
	ss.wg.Add(1)
	defer ss.wg.Done()

	if ss.GetCtx().Err() != nil {
		return false
	}
	select {
	case ss.ch <- it:
		return true
	default:
		return false
	}
}

func (ss *SafeSender[T]) GetChan() <-chan T {
	return ss.ch
}
