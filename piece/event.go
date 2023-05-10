package piece

import "encoding/json"

// Event

type EventType string

const (
	EventTypeTransitional EventType = "Transitional"
	EventTypeOnEntry      EventType = "OnEntry"
	EventTypeDoAfter      EventType = "DoAfter"
	EventTypeDoAlways     EventType = "DoAlways"
	EventTypeOnDone       EventType = "OnDone"
	EventTypeOnError      EventType = "OnError"
	EventTypeStartOn      EventType = "StartOn"
)

type Event interface {
	GetName() string                       // get event name
	GetFrom() string                       // get where the event from
	HasData() bool                         // check if event has data
	GetData() any                          // get event data
	GetDataAsMap() (map[string]any, error) // get event data as map <string, any>
	GetErr() error                         // get event error, if any
	GetEvtType() EventType                 // get event type
	ToBuilder() EventBuilder               // get event builder
}

type GEvent struct {
	name    string
	data    any
	from    string
	err     error
	evtType EventType
}

func (e *GEvent) GetName() string {
	return e.name
}

func (e *GEvent) GetFrom() string {
	return e.from
}

func (e *GEvent) HasData() bool {
	return e.data != nil
}

func (e *GEvent) GetData() any {
	return e.data
}

func (e *GEvent) GetDataAsMap() (map[string]any, error) {
	if e.data == nil {
		return nil, nil
	}
	m, err := json.Marshal(e.data)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	err = json.Unmarshal(m, &data)
	return data, err
}

func (e *GEvent) GetErr() error {
	return e.err
}

func (e *GEvent) GetEvtType() EventType {
	return e.evtType
}

func (e *GEvent) ToBuilder() EventBuilder {
	return &GEventBuilder{
		name: e.name,
		data: e.data,
		err:  e.err,
	}
}

// Event Builder

func BuildEvent(name string) EventBuilder {
	return &GEventBuilder{
		name: name,
	}
}

type EventBuilder interface {
	WithData(data any) EventBuilder
	WithErr(err error) EventBuilder
	WithType(eventType EventType) EventBuilder
	Build() Event
}

type GEventBuilder struct {
	name      string
	data      any
	eventType *EventType
	err       error
}

func (b *GEventBuilder) WithData(data any) EventBuilder {
	b.data = data
	return b
}

func (b *GEventBuilder) WithErr(err error) EventBuilder {
	b.err = err
	return b
}

func (b *GEventBuilder) WithType(eventType EventType) EventBuilder {
	b.eventType = &eventType
	return b
}

func (b *GEventBuilder) Build() Event {
	if b.eventType == nil {
		tmpEvtType := EventTypeOnEntry
		b.eventType = &tmpEvtType
	}
	return &GEvent{
		name:    b.name,
		data:    b.data,
		err:     b.err,
		evtType: *b.eventType,
	}
}
