package piece

import "encoding/json"

// Event Builder

func BuildEvent(name string) EventBuilder {
	return &GEventBuilder{
		name: name,
	}
}

type EventBuilder interface {
	WithData(data any) EventBuilder
	WithErr(err error) EventBuilder
	WithType(evtType EventType) EventBuilder
	Build() Event
}

type GEventBuilder struct {
	name    string
	data    any
	err     error
	evtType EventType
}

func (b *GEventBuilder) WithData(data any) EventBuilder {
	b.data = data
	return b
}

func (b *GEventBuilder) WithErr(err error) EventBuilder {
	b.err = err
	return b
}

func (b *GEventBuilder) WithType(evtType EventType) EventBuilder {
	b.evtType = evtType
	return b
}

func (b *GEventBuilder) Build() Event {
	return &GEvent{
		name:    b.name,
		data:    b.data,
		err:     b.err,
		evtType: b.evtType,
	}
}

// Event

type EventType string

const (
	EventTypeTransitional EventType = "Transitional"
	EventTypeOnEntry      EventType = "OnEntry"
	EventTypeDoAfter      EventType = "DoAfter"
	EventTypeDoAlways     EventType = "DoAlways"
	EventTypeOnDone       EventType = "OnDone"
	EventTypeOnError      EventType = "OnError"
)

type Event interface {
	GetName() string
	GetData() any
	GetFrom() string
	GetDataAsMap() (map[string]any, error)
	GetErr() error
	GetEvtType() EventType
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

func (e *GEvent) GetData() any {
	return e.data
}

func (e *GEvent) GetFrom() string {
	return e.from
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
