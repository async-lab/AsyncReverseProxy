package pattern

import "context"

func SelectContextAndChannel[CH any](
	ctx context.Context,
	// ch chan CH,
	ctxDoneHandler func(),
	channelHandler func(CH),
	goroutine func(chan CH),
) {
	ch := make(chan CH)

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
			if !ok {
				return
			}
			channelHandler(result)
		}
	}
}

type ConfigSelectContextAndChannel[CH any] struct {
	Ctx            context.Context
	Ch             chan CH
	CtxDoneHandler func()
	ChannelHandler func(CH)
	Goroutine      func(chan CH)
}

func NewConfigSelectContextAndChannel[CH any]() *ConfigSelectContextAndChannel[CH] {
	return &ConfigSelectContextAndChannel[CH]{
		Ctx: context.TODO(),
		// Ch:             make(chan CH),
		CtxDoneHandler: func() {},
		ChannelHandler: func(CH) {},
		Goroutine:      func(chan CH) {},
	}
}

func (c *ConfigSelectContextAndChannel[T]) WithCtx(ctx context.Context) *ConfigSelectContextAndChannel[T] {
	c.Ctx = ctx
	return c
}

// func (c *ConfigSelectContextAndChannel[T]) WithCh(ch chan T) *ConfigSelectContextAndChannel[T] {
// 	c.Ch = ch
// 	return c
// }

func (c *ConfigSelectContextAndChannel[T]) WithCtxDoneHandler(ctxDoneHandler func()) *ConfigSelectContextAndChannel[T] {
	c.CtxDoneHandler = ctxDoneHandler
	return c
}

func (c *ConfigSelectContextAndChannel[T]) WithChannelHandler(channelHandler func(T)) *ConfigSelectContextAndChannel[T] {
	c.ChannelHandler = channelHandler
	return c
}

func (c *ConfigSelectContextAndChannel[T]) WithGoroutine(goroutine func(chan T)) *ConfigSelectContextAndChannel[T] {
	c.Goroutine = goroutine
	return c
}

func (c *ConfigSelectContextAndChannel[T]) Run() {
	SelectContextAndChannel[T](c.Ctx, c.CtxDoneHandler, c.ChannelHandler, c.Goroutine)
}
