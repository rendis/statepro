package experimental

import (
	"fmt"
	"github.com/rendis/statepro/v3/instrumentation"
)

// newEventAccumulator returns a new Event accumulator
func newEventAccumulator() instrumentation.Accumulator {
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

func (ea *eventAccumulator) Accumulate(realityName string, evt instrumentation.Event) {
	if ea.RealitiesEvents == nil {
		ea.RealitiesEvents = map[string][]*event{}
	}

	if _, ok := ea.RealitiesEvents[realityName]; !ok {
		ea.RealitiesEvents[realityName] = []*event{}
	}

	ea.RealitiesEvents[realityName] = append(ea.RealitiesEvents[realityName], evt.(*event))
}

func (ea *eventAccumulator) GetStatistics() instrumentation.AccumulatorStatistics {
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

func (ea *eventAccumulator) GetRealitiesEvents() map[string][]instrumentation.Event {
	var resp = map[string][]instrumentation.Event{}
	for k, evts := range ea.RealitiesEvents {
		resp[k] = []instrumentation.Event{}
		for _, evt := range evts {
			resp[k] = append(resp[k], evt)
		}
	}
	return resp
}

func (ea *eventAccumulator) GetRealityEvents(realityName string) []instrumentation.Event {
	var events []instrumentation.Event
	for _, event := range ea.RealitiesEvents[realityName] {
		events = append(events, event)
	}
	return events
}

func (ea *eventAccumulator) GetAllEventsNames() []string {
	var eventsNames []string
	eventsProcessed := map[string]bool{}

	for _, events := range ea.RealitiesEvents {
		for _, evt := range events {
			if _, ok := eventsProcessed[evt.GetEventName()]; !ok {
				eventsNames = append(eventsNames, evt.GetEventName())
				eventsProcessed[evt.GetEventName()] = true
			}
		}
	}

	return eventsNames
}

func (ea *eventAccumulator) CountAllEventsNames() int {
	return len(ea.GetAllEventsNames())
}
