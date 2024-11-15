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

func (b *EventBus) subscribe(t reflect.Type, handler EventHandler[IEvent]) {
	b.listeners[t] = append(b.listeners[t], handler)
}

func (b *EventBus) publish(t reflect.Type, e IEvent) bool {
	for _, handler := range b.listeners[t] {
		if ok := handler(e); !ok {
			return false
		}
	}
	return true
}

func Subscribe[T IEvent](eventManager *EventBus, handler EventHandler[T]) {
	t := lang.GetActualTypeWithGeneric[T]()
	eventManager.subscribe(t, func(event IEvent) bool {
		if event, ok := event.(T); ok {
			return handler(event)
		} else {
			logger.Error("Type assertion failed for event type: ", t)
			return false
		}
	})
}

func Publish(eventManager *EventBus, e IEvent) bool {
	return eventManager.publish(lang.GetActualType(e), e)
}
