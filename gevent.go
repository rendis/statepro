package statepro

import "encoding/json"

// Event

type EventType string

const (
	EventTypeStart        EventType = "Start"
	EventTypeStartOn      EventType = "StartOn"
	EventTypeTransitional EventType = "Transitional"
	EventTypeOnEntry      EventType = "OnEntry"
	EventTypeDoAlways     EventType = "DoAlways"
)

type Event interface {
	GetEventName() string                  // get event name
	HasData() bool                         // check if event has data
	GetData() any                          // get event data
	GetDataAsMap() (map[string]any, error) // get event data as map <string, any>
	GetErr() error                         // get event error, if any
	GetEvtType() EventType                 // get event type
	ToBuilder() EventBuilder               // get event builder
	GetMachineInfo() MachineBasicInfo      // get machine basic info
}

type gEvent struct {
	name        string
	data        any
	err         error
	evtType     EventType
	machineInfo MachineBasicInfo
}

func (e *gEvent) GetEventName() string {
	return e.name
}

func (e *gEvent) HasData() bool {
	return e.data != nil
}

func (e *gEvent) GetData() any {
	return e.data
}

func (e *gEvent) GetDataAsMap() (map[string]any, error) {
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

func (e *gEvent) GetErr() error {
	return e.err
}

func (e *gEvent) GetEvtType() EventType {
	return e.evtType
}

func (e *gEvent) ToBuilder() EventBuilder {
	return &GEventBuilder{
		name: e.name,
		data: e.data,
		err:  e.err,
	}
}

func (e *gEvent) GetMachineInfo() MachineBasicInfo {
	return e.machineInfo
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
		tmpEvtType := EventTypeTransitional
		b.eventType = &tmpEvtType
	}
	return &gEvent{
		name:    b.name,
		data:    b.data,
		err:     b.err,
		evtType: *b.eventType,
	}
}
