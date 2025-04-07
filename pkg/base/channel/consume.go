package channel

import "context"

func ConsumeWithCtx[T any](ctx context.Context, ch <-chan T, handler func(T) bool) {
	for {
		select {
		case result, ok := <-ch:
			if !ok || !handler(result) {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func Consume[T any](ch <-chan T, handler func(T) bool) {
	ConsumeWithCtx(context.Background(), ch, handler)
}
