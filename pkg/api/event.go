package api

type EventType int

type IEventPayload interface {
	ToMap() map[string]interface{}
}

type Event struct {
	Type    EventType
	Payload IEventPayload
}

type EventHandler func(event Event)

type EventManager struct {
	listeners map[EventType][]EventHandler
}

func NewEventManager() *EventManager {
	return &EventManager{
		listeners: make(map[EventType][]EventHandler),
	}
}

func (e *EventManager) Subscribe(eventType EventType, handler EventHandler) {
	e.listeners[eventType] = append(e.listeners[eventType], handler)
}

func (e *EventManager) Publish(event Event) {
	for _, handler := range e.listeners[event.Type] {
		handler(event)
	}
}
