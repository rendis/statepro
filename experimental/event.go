package experimental

import (
	"github.com/rendis/statepro/v3/instrumentation"
)

type Event struct {
	Name    string                     `json:"name" bson:"name" xml:"name"`
	Data    map[string]any             `json:"data,omitempty" bson:"data,omitempty" xml:"data,omitempty"`
	EvtType instrumentation.EventType  `json:"type" bson:"type" xml:"type"`
	Flags   instrumentation.EventFlags `json:"flags,omitempty" bson:"flags,omitempty" xml:"flags,omitempty"`
}

func (e *Event) GetEventName() string {
	return e.Name
}

func (e *Event) GetData() map[string]any {
	return e.Data
}

func (e *Event) DataContainsKey(key string) bool {
	if e.Data == nil {
		return false
	}
	_, ok := e.Data[key]
	return ok
}

func (e *Event) GetEvtType() instrumentation.EventType {
	return e.EvtType
}

func (e *Event) GetFlags() instrumentation.EventFlags {
	return e.Flags
}

func (e *Event) String() string {
	return e.Name
}

// ---------  Event Builder --------- //

func NewEventBuilder(name string) instrumentation.EventBuilder {
	return &EventBuilder{
		data:    map[string]any{},
		name:    name,
		evtType: instrumentation.EventTypeOn,
	}
}

type EventBuilder struct {
	name    string
	data    map[string]any
	evtType instrumentation.EventType
	flags   instrumentation.EventFlags
}

func (eb *EventBuilder) SetData(data map[string]any) instrumentation.EventBuilder {
	eb.data = data
	return eb
}

func (eb *EventBuilder) SetEvtType(evtType instrumentation.EventType) instrumentation.EventBuilder {
	eb.evtType = evtType
	return eb
}

func (eb *EventBuilder) SetFlags(flags instrumentation.EventFlags) instrumentation.EventBuilder {
	eb.flags = flags
	return eb
}

func (eb *EventBuilder) Build() instrumentation.Event {
	return &Event{
		Name:    eb.name,
		Data:    eb.data,
		EvtType: eb.evtType,
		Flags:   eb.flags,
	}
}
