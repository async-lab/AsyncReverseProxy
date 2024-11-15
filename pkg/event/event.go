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

func (e *EventBus) subscribe(tType reflect.Type, handler EventHandler[IEvent]) {
	e.listeners[tType] = append(e.listeners[tType], handler)
}

func (e *EventBus) publish(tType reflect.Type, event IEvent) bool {
	for _, handler := range e.listeners[tType] {
		if ok := handler(event); !ok {
			return false
		}
	}
	return true
}

func Subscribe[T IEvent](eventManager *EventBus, handler EventHandler[T]) {
	tType := lang.GetActualTypeWithGeneric[T]()
	eventManager.subscribe(tType, func(event IEvent) bool {
		if event, ok := event.(T); ok {
			return handler(event)
		} else {
			logger.Error("Type assertion failed for event type: ", tType)
			return false
		}
	})
}

func Publish(eventManager *EventBus, event IEvent) bool {
	return eventManager.publish(lang.GetActualType(event), event)
}
