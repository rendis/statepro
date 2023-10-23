package experimental

import (
	"fmt"
	"github.com/rendis/statepro/v3/instrumentation"
)

// --------- Event --------- //

func NewEvent(name string, data map[string]any, evtType instrumentation.EventType) instrumentation.Event {
	return &event{
		name:    name,
		data:    data,
		evtType: evtType,
	}
}

type event struct {
	name    string
	data    map[string]any
	evtType instrumentation.EventType
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

func (e *event) GetEvtType() instrumentation.EventType {
	return e.evtType
}

func (e *event) String() string {
	return fmt.Sprintf("%s", e.name)
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
}

func (eb *EventBuilder) SetEventName(name string) instrumentation.EventBuilder {
	eb.name = name
	return eb
}

func (eb *EventBuilder) SetData(data map[string]any) instrumentation.EventBuilder {
	eb.data = data
	return eb
}

func (eb *EventBuilder) SetEvtType(evtType instrumentation.EventType) instrumentation.EventBuilder {
	eb.evtType = evtType
	return eb
}

func (eb *EventBuilder) Build() instrumentation.Event {
	return NewEvent(eb.name, eb.data, eb.evtType)
}
