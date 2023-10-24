package experimental

import (
	"fmt"
	"github.com/rendis/statepro/v3/instrumentation"
)

// --------- Event --------- //

func NewEvent(name string, data map[string]any, evtType instrumentation.EventType) instrumentation.Event {
	return &event{
		Name:    name,
		Data:    data,
		EvtType: evtType,
	}
}

type event struct {
	Name    string                    `json:"name" bson:"name" xml:"name"`
	Data    map[string]any            `json:"data,omitempty" bson:"data,omitempty" xml:"data,omitempty"`
	EvtType instrumentation.EventType `json:"type" bson:"type" xml:"type"`
}

func (e *event) GetEventName() string {
	return e.Name
}

func (e *event) GetData() map[string]any {
	return e.Data
}

func (e *event) DataContainsKey(key string) bool {
	if e.Data == nil {
		return false
	}
	_, ok := e.Data[key]
	return ok
}

func (e *event) GetEvtType() instrumentation.EventType {
	return e.EvtType
}

func (e *event) String() string {
	return fmt.Sprintf("%s", e.Name)
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
