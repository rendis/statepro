package experimental

import "fmt"

// Accumulator allows to accumulate events in different realities for a given universe
type Accumulator interface {
	// Accumulate accumulates an event in the given reality
	// Only events that exist in the RealityModel.On will be accumulated
	// Parameters:
	// 	- realityName: the reality name
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
	// The map key is the reality name and the value is the accumulated events
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
		RealitiesEvents: map[string][]*event{},
	}
}

type eventAccumulator struct {
	// RealitiesEvents is the map of realities and their accumulated events
	// The map key is the reality name and the value is the accumulated events
	RealitiesEvents map[string][]*event `json:"realitiesEvents,omitempty"`
}

func (ea *eventAccumulator) String() string {
	var msg string
	for realityName, events := range ea.RealitiesEvents {
		msg += fmt.Sprintf("%s: %v\n", realityName, events)
	}
	return msg
}

// Accumulator implementation

func (ea *eventAccumulator) Accumulate(realityName string, evt Event) {
	if ea.RealitiesEvents == nil {
		ea.RealitiesEvents = map[string][]*event{}
	}

	if _, ok := ea.RealitiesEvents[realityName]; !ok {
		ea.RealitiesEvents[realityName] = []*event{}
	}

	ea.RealitiesEvents[realityName] = append(ea.RealitiesEvents[realityName], evt.(*event))
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
	var resp = map[string][]Event{}
	for k, evts := range ea.RealitiesEvents {
		resp[k] = []Event{}
		for _, evt := range evts {
			resp[k] = append(resp[k], evt)
		}
	}
	return resp
}

func (ea *eventAccumulator) GetRealityEvents(realityName string) []Event {
	var events []Event
	for _, evt := range ea.RealitiesEvents[realityName] {
		events = append(events, evt)
	}
	return events
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
