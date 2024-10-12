package pattern

import "context"

func SelectContextAndChannel[T any](
	ctx context.Context,
	ch chan T,
	ctxDoneHandler func(),
	channelHandler func(T) bool,
	goroutines ...func(chan T),
) {
	for _, goroutine := range goroutines {
		go goroutine(ch)
	}
	for {
		select {
		case <-ctx.Done():
			ctxDoneHandler()
			return
		case result := <-ch:
			if !channelHandler(result) {
				return
			}
		}
	}
}
