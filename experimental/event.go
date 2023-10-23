package experimental

import "fmt"

type EventType string

const (
	EventTypeBigBang   EventType = "BigBang"   // event triggered when the universe starts
	EventTypeBigBangOn EventType = "BigBangOn" // event triggered when the universe starts on a reality
	EventTypeAlways    EventType = "Always"    // event triggered from reality from its "always transitions"
	EventTypeOn        EventType = "On"        // event triggered from reality from its "on transitions"
	EventTypeOnEntry   EventType = "OnEntry"   // event used to force the current reality to execute logic on entry
)

// -------------------------- //
// ---- Event Definition ---- //
// -------------------------- //

// Event is the interface that wraps the basic event methods.
type Event interface {
	// GetEventName returns the event Name
	GetEventName() string

	// GetData returns the event Data
	GetData() map[string]any

	// DataContainsKey returns true if the event Data contains the given key
	DataContainsKey(key string) bool

	// GetEvtType returns the event type
	GetEvtType() EventType

	// ToBuilder returns event in builder format
	ToBuilder() EventBuilder
}

type event struct {
	Name    string
	Data    map[string]any
	EvtType EventType
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

func (e *event) GetEvtType() EventType {
	return e.EvtType
}

func (e *event) ToBuilder() EventBuilder {
	return &eventBuilder{
		name:    e.Name,
		data:    e.Data,
		evtType: e.EvtType,
	}
}

func (e *event) String() string {
	return fmt.Sprintf("%s", e.Name)
}

// -------------------------------------- //
// ---- Event Accumulator Definition ---- //
// -------------------------------------- //

// Accumulator allows to accumulate events in different realities for a given universe
type Accumulator interface {
	// Accumulate accumulates an event in the given reality
	// Only events that exist in the RealityModel.On will be accumulated
	// Parameters:
	// 	- realityName: the reality Name
	// 	- event: the event to accumulate
	Accumulate(realityName string, event Event)

	// GetStatistics returns the event accumulator statistics
	GetStatistics() AccumulatorStatistics

	// GetActiveRealities returns the realities that have accumulated events
	GetActiveRealities() []string
}

// AccumulatorStatistics allows to get statistics from an event accumulator
type AccumulatorStatistics interface {
	// GetRealitiesEvents returns the accumulated events for each reality
	// Events can be repeated if they are accumulated in more than one reality
	// The map key is the reality Name and the value is the accumulated events
	GetRealitiesEvents() map[string][]Event

	// GetRealityEvents returns the accumulated events for the given reality
	// Events can be repeated if they were received more than once
	GetRealityEvents(realityName string) []Event

	// GetAllEventsNames returns the accumulated events names for all realities (without repetitions)
	GetAllEventsNames() []string

	// CountAllEventsNames returns the number of accumulated events names for all realities (without repetitions)
	CountAllEventsNames() int
}

// newEventAccumulator returns a new event accumulator
func newEventAccumulator() Accumulator {
	return &eventAccumulator{
		RealitiesEvents: map[string][]Event{},
	}
}

type eventAccumulator struct {
	// RealitiesEvents is the map of realities and their accumulated events
	// The map key is the reality Name and the value is the accumulated events
	RealitiesEvents map[string][]Event `json:"realities_events,omitempty" bson:"realitiesEvents,omitempty" xml:"realitiesEvents,omitempty" yaml:"realitiesEvents,omitempty"`
}

func (ea *eventAccumulator) String() string {
	var msg string
	for realityName, events := range ea.RealitiesEvents {
		msg += fmt.Sprintf("%s: %v\n", realityName, events)
	}
	return msg
}

// Accumulator implementation

func (ea *eventAccumulator) Accumulate(realityName string, event Event) {
	if _, ok := ea.RealitiesEvents[realityName]; !ok {
		ea.RealitiesEvents[realityName] = []Event{}
	}

	ea.RealitiesEvents[realityName] = append(ea.RealitiesEvents[realityName], event)
}

func (ea *eventAccumulator) GetStatistics() AccumulatorStatistics {
	return ea
}

func (ea *eventAccumulator) GetActiveRealities() []string {
	var activeRealities []string
	for realityName := range ea.RealitiesEvents {
		activeRealities = append(activeRealities, realityName)
	}
	return activeRealities
}

// AccumulatorStatistics implementation

func (ea *eventAccumulator) GetRealitiesEvents() map[string][]Event {
	return ea.RealitiesEvents
}

func (ea *eventAccumulator) GetRealityEvents(realityName string) []Event {
	return ea.RealitiesEvents[realityName]
}

func (ea *eventAccumulator) GetAllEventsNames() []string {
	var eventsNames []string
	eventsProcessed := map[string]bool{}

	for _, events := range ea.RealitiesEvents {
		for _, event := range events {
			if _, ok := eventsProcessed[event.GetEventName()]; !ok {
				eventsNames = append(eventsNames, event.GetEventName())
				eventsProcessed[event.GetEventName()] = true
			}
		}
	}

	return eventsNames
}

func (ea *eventAccumulator) CountAllEventsNames() int {
	return len(ea.GetAllEventsNames())
}

// ---------------------------------- //
// ---- Event Builder Definition ---- //
// ---------------------------------- //

// EventBuilder is the interface that wraps the basic event builder methods.
type EventBuilder interface {
	// SetEventName sets the event Name
	SetEventName(name string) EventBuilder

	// SetData sets the event Data
	SetData(data map[string]any) EventBuilder

	// SetEvtType sets the event type
	SetEvtType(evtType EventType) EventBuilder

	// Build returns the event
	Build() (Event, error)
}

// NewEventBuilder returns a new event builder
func NewEventBuilder() EventBuilder {
	return &eventBuilder{
		data:    map[string]any{},
		evtType: EventTypeOn,
	}
}

type eventBuilder struct {
	name    string
	data    map[string]any
	evtType EventType
}

func (eb *eventBuilder) SetEventName(name string) EventBuilder {
	eb.name = name
	return eb
}

func (eb *eventBuilder) SetData(data map[string]any) EventBuilder {
	eb.data = data
	return eb
}

func (eb *eventBuilder) SetEvtType(evtType EventType) EventBuilder {
	eb.evtType = evtType
	return eb
}

func (eb *eventBuilder) Build() (Event, error) {
	if eb.evtType == "" {
		return nil, fmt.Errorf("event Name is required")
	}

	return &event{
		Name:    eb.name,
		Data:    eb.data,
		EvtType: eb.evtType,
	}, nil
}
