package experimental

import "fmt"

// --------- Event --------- //

type EventType string

const (
	EventTypeBigBang   EventType = "BigBang"   // event triggered when the universe starts
	EventTypeBigBangOn EventType = "BigBangOn" // event triggered when the universe starts on a reality
	EventTypeAlways    EventType = "Always"    // event triggered from reality from its "always transitions"
	EventTypeOn        EventType = "On"        // event triggered from reality from its "on transitions"
	EventTypeOnEntry   EventType = "OnEntry"   // event used to force the current reality to execute logic on entry
)

type Event interface {
	// GetEventName returns the event name
	GetEventName() string

	// GetData returns the event data
	GetData() map[string]any

	// DataContainsKey returns true if the event data contains the given key
	DataContainsKey(key string) bool

	// GetEvtType returns the event type
	GetEvtType() EventType
}

type event struct {
	name    string
	data    map[string]any
	evtType EventType
}

func (e *event) GetEventName() string {
	return e.name
}

func (e *event) GetData() map[string]any {
	return e.data
}

func (e *event) DataContainsKey(key string) bool {
	if e.data == nil {
		return false
	}
	_, ok := e.data[key]
	return ok
}

func (e *event) GetEvtType() EventType {
	return e.evtType
}

func (e *event) String() string {
	return fmt.Sprintf("%s", e.name)
}

// --------- Event Builder --------- //

// NewEventBuilder returns a new event builder
func NewEventBuilder(name string) *EventBuilder {
	return &EventBuilder{
		data:    map[string]any{},
		name:    name,
		evtType: EventTypeOn,
	}
}

type EventBuilder struct {
	name    string
	data    map[string]any
	evtType EventType
}

func (eb *EventBuilder) SetEventName(name string) *EventBuilder {
	eb.name = name
	return eb
}

func (eb *EventBuilder) SetData(data map[string]any) *EventBuilder {
	eb.data = data
	return eb
}

func (eb *EventBuilder) SetEvtType(evtType EventType) *EventBuilder {
	eb.evtType = evtType
	return eb
}

func (eb *EventBuilder) Build() Event {
	return &event{
		name:    eb.name,
		data:    eb.data,
		evtType: eb.evtType,
	}
}
