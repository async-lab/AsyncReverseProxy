package event

import (
	"reflect"

	"club.asynclab/asrp/pkg/base/lang"
	"club.asynclab/asrp/pkg/logging"
)

var logger = logging.GetLogger()

type IEvent interface{}
type EventHandler[T IEvent] func(event T) bool
type EventBus struct {
	listeners map[reflect.Type][]EventHandler[IEvent]
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[reflect.Type][]EventHandler[IEvent]),
	}
}

func (b *EventBus) subscribe(t reflect.Type, h EventHandler[IEvent]) {
	b.listeners[t] = append(b.listeners[t], h)
}

func (b *EventBus) publish(t reflect.Type, e IEvent) bool {
	for _, handler := range b.listeners[t] {
		if ok := handler(e); !ok {
			return false
		}
	}
	return true
}

func Subscribe[T IEvent](b *EventBus, h EventHandler[T]) {
	t := lang.GetActualTypeWithGeneric[T]()
	b.subscribe(t, func(e IEvent) bool {
		if e, ok := e.(T); ok {
			return h(e)
		} else {
			logger.Error("Type assertion failed for event type: ", t)
			return false
		}
	})
}

func Publish(b *EventBus, e IEvent) bool {
	return b.publish(lang.GetActualType(e), e)
}
