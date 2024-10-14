package event

import (
	"fmt"
	"reflect"

	"club.asynclab/asrp/pkg/util"
)

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
	tType := util.GetStructType((*T)(nil))
	eventManager.subscribe(tType, func(event IEvent) bool {
		if event, ok := event.(T); ok {
			return handler(event)
		} else {
			fmt.Printf("Type assertion failed for event type: %s\n", tType)
			return false
		}
	})
}

func Publish(eventManager *EventBus, event IEvent) bool {
	return eventManager.publish(util.GetStructType(event), event)
}
