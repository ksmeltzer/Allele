package core

type EventType string

const (
	ConfigUpdatedEvent EventType = "config_updated"
	SystemAlertEvent   EventType = "system_alert"
)

type Event struct {
	Type    EventType
	Payload interface{}
}

// EventBus is a simple PubSub system for the microkernel
type EventBus struct {
	subscribers map[EventType][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
	}
}

func (eb *EventBus) Subscribe(eventType EventType) chan Event {
	ch := make(chan Event, 100)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

func (eb *EventBus) Publish(event Event) {
	for _, ch := range eb.subscribers[event.Type] {
		select {
		case ch <- event:
		default: // non-blocking if subscriber channel is full
		}
	}
}
