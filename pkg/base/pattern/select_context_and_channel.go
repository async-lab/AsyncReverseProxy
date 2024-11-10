package pattern

import "context"

func SelectContextAndChannel[CH any](
	ctx context.Context,
	ctxDoneHandler func(),
	channelHandler func(CH) bool,
	channelBufferSize int,
	goroutine func(chan CH),
) {
	ch := make(chan CH, channelBufferSize)

	go func() {
		defer close(ch)
		goroutine(ch)
	}()

	for {
		select {
		case <-ctx.Done():
			ctxDoneHandler()
			return
		case result, ok := <-ch:
			if !ok || !channelHandler(result) {
				return
			}
		}
	}
}

type ConfigSelectContextAndChannel[CH any] struct {
	Ctx               context.Context
	Ch                chan CH
	CtxDoneHandler    func()
	ChannelHandler    func(CH) bool
	channelBufferSize int
	Goroutine         func(chan CH)
}

func NewConfigSelectContextAndChannel[CH any]() *ConfigSelectContextAndChannel[CH] {
	return &ConfigSelectContextAndChannel[CH]{
		Ctx:               context.TODO(),
		CtxDoneHandler:    func() {},
		ChannelHandler:    func(CH) bool { return true },
		channelBufferSize: 16,
		Goroutine:         func(chan CH) {},
	}
}

func (c *ConfigSelectContextAndChannel[T]) WithCtx(ctx context.Context) *ConfigSelectContextAndChannel[T] {
	c.Ctx = ctx
	return c
}

func (c *ConfigSelectContextAndChannel[T]) WithCtxDoneHandler(ctxDoneHandler func()) *ConfigSelectContextAndChannel[T] {
	c.CtxDoneHandler = ctxDoneHandler
	return c
}

func (c *ConfigSelectContextAndChannel[T]) WithChannelHandler(channelHandler func(T)) *ConfigSelectContextAndChannel[T] {
	c.ChannelHandler = func(t T) bool {
		channelHandler(t)
		return true
	}
	return c
}

func (c *ConfigSelectContextAndChannel[T]) WithChannelHandlerWithInterruption(channelHandler func(T) bool) *ConfigSelectContextAndChannel[T] {
	c.ChannelHandler = channelHandler
	return c
}

func (c *ConfigSelectContextAndChannel[T]) WithChannelBufferSize(channelBufferSize int) *ConfigSelectContextAndChannel[T] {
	c.channelBufferSize = channelBufferSize
	return c
}

func (c *ConfigSelectContextAndChannel[T]) WithGoroutine(goroutine func(chan T)) *ConfigSelectContextAndChannel[T] {
	c.Goroutine = goroutine
	return c
}

func (c *ConfigSelectContextAndChannel[T]) Run() {
	SelectContextAndChannel[T](c.Ctx, c.CtxDoneHandler, c.ChannelHandler, c.channelBufferSize, c.Goroutine)
}
