package experimental

import (
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
)

// mockEvent implements instrumentation.Event for testing
type mockEvent struct {
	name string
	data map[string]any
}

func (e *mockEvent) GetEventName() string {
	return e.name
}

func (e *mockEvent) GetData() map[string]any {
	return e.data
}

func (e *mockEvent) DataContainsKey(key string) bool {
	if e.data == nil {
		return false
	}
	_, ok := e.data[key]
	return ok
}

func (e *mockEvent) GetEvtType() instrumentation.EventType {
	return instrumentation.EventTypeOn
}

func (e *mockEvent) GetFlags() instrumentation.EventFlags {
	return instrumentation.EventFlags{}
}

func TestNewEventAccumulator(t *testing.T) {
	accumulator := newEventAccumulator()
	if accumulator == nil {
		t.Fatal("Expected non-nil accumulator")
	}

	// Cast to concrete type to access internal fields
	concreteAccumulator, ok := accumulator.(*eventAccumulator)
	if !ok {
		t.Fatal("Expected accumulator to be of type *eventAccumulator")
	}

	// Check that RealitiesEvents is initialized
	if concreteAccumulator.RealitiesEvents == nil {
		t.Fatal("Expected RealitiesEvents to be initialized")
	}
}

func TestEventAccumulator_Accumulate(t *testing.T) {
	accumulator := newEventAccumulator()
	concreteAccumulator, ok := accumulator.(*eventAccumulator)
	if !ok {
		t.Fatal("Expected accumulator to be of type *eventAccumulator")
	}

	event1 := NewEventBuilder("event1").Build()
	event2 := NewEventBuilder("event2").Build()

	// Accumulate events for different realities
	accumulator.Accumulate("reality1", event1)
	accumulator.Accumulate("reality1", event2)
	accumulator.Accumulate("reality2", event1)

	// Check reality1 has 2 events
	if len(concreteAccumulator.RealitiesEvents["reality1"]) != 2 {
		t.Fatalf("Expected 2 events for reality1, got %d", len(concreteAccumulator.RealitiesEvents["reality1"]))
	}

	// Check reality2 has 1 event
	if len(concreteAccumulator.RealitiesEvents["reality2"]) != 1 {
		t.Fatalf("Expected 1 event for reality2, got %d", len(concreteAccumulator.RealitiesEvents["reality2"]))
	}
}

func TestEventAccumulator_GetStatistics(t *testing.T) {
	accumulator := newEventAccumulator()
	stats := accumulator.GetStatistics()

	if stats == nil {
		t.Fatal("Expected non-nil statistics")
	}

	// Should return the accumulator itself (same object)
	concreteAccumulator, ok := accumulator.(*eventAccumulator)
	if !ok {
		t.Fatal("Expected accumulator to be of type *eventAccumulator")
	}

	concreteStats, ok := stats.(*eventAccumulator)
	if !ok {
		t.Fatal("Expected stats to be of type *eventAccumulator")
	}

	if concreteStats != concreteAccumulator {
		t.Fatal("Expected GetStatistics to return the accumulator itself")
	}
}

func TestEventAccumulator_GetActiveRealities(t *testing.T) {
	accumulator := newEventAccumulator()

	// Add events to different realities
	event := NewEventBuilder("event1").Build()
	accumulator.Accumulate("reality1", event)
	accumulator.Accumulate("reality2", event)
	accumulator.Accumulate("reality1", event) // duplicate

	activeRealities := accumulator.GetActiveRealities()

	if len(activeRealities) != 2 {
		t.Fatalf("Expected 2 active realities, got %d", len(activeRealities))
	}

	// Check that both realities are present
	realityMap := make(map[string]bool)
	for _, reality := range activeRealities {
		realityMap[reality] = true
	}

	if !realityMap["reality1"] || !realityMap["reality2"] {
		t.Fatalf("Expected realities 'reality1' and 'reality2', got %v", activeRealities)
	}
}

func TestEventAccumulator_GetActiveRealities_Empty(t *testing.T) {
	accumulator := newEventAccumulator()
	activeRealities := accumulator.GetActiveRealities()

	if len(activeRealities) != 0 {
		t.Fatalf("Expected 0 active realities for empty accumulator, got %d", len(activeRealities))
	}
}

func TestEventAccumulator_GetRealitiesEvents(t *testing.T) {
	accumulator := newEventAccumulator()

	event1 := NewEventBuilder("event1").Build()
	event2 := NewEventBuilder("event2").Build()

	accumulator.Accumulate("reality1", event1)
	accumulator.Accumulate("reality1", event2)
	accumulator.Accumulate("reality2", event1)

	stats := accumulator.GetStatistics()
	realitiesEvents := stats.GetRealitiesEvents()

	if len(realitiesEvents) != 2 {
		t.Fatalf("Expected 2 realities, got %d", len(realitiesEvents))
	}

	// Check reality1 events
	reality1Events := realitiesEvents["reality1"]
	if len(reality1Events) != 2 {
		t.Fatalf("Expected 2 events for reality1, got %d", len(reality1Events))
	}

	// Check reality2 events
	reality2Events := realitiesEvents["reality2"]
	if len(reality2Events) != 1 {
		t.Fatalf("Expected 1 event for reality2, got %d", len(reality2Events))
	}
}

func TestEventAccumulator_GetAllRealityEvents(t *testing.T) {
	accumulator := newEventAccumulator()

	event1 := NewEventBuilder("event1").Build()
	event2 := NewEventBuilder("event2").Build()

	// Add multiple events with same name
	accumulator.Accumulate("reality1", event1)
	accumulator.Accumulate("reality1", event1) // duplicate event name
	accumulator.Accumulate("reality1", event2)

	stats := accumulator.GetStatistics()
	realityEvents := stats.GetAllRealityEvents("reality1")

	if len(realityEvents) != 2 {
		t.Fatalf("Expected 2 unique event names, got %d", len(realityEvents))
	}

	// Check event1 has 2 occurrences
	event1Events := realityEvents["event1"]
	if len(event1Events) != 2 {
		t.Fatalf("Expected 2 'event1' events, got %d", len(event1Events))
	}

	// Check event2 has 1 occurrence
	event2Events := realityEvents["event2"]
	if len(event2Events) != 1 {
		t.Fatalf("Expected 1 'event2' event, got %d", len(event2Events))
	}
}

func TestEventAccumulator_GetAllRealityEvents_NoEvents(t *testing.T) {
	accumulator := newEventAccumulator()

	stats := accumulator.GetStatistics()
	realityEvents := stats.GetAllRealityEvents("nonexistent")

	if len(realityEvents) != 0 {
		t.Fatalf("Expected empty map for nonexistent reality, got %d events", len(realityEvents))
	}
}

func TestEventAccumulator_GetRealityEvents(t *testing.T) {
	accumulator := newEventAccumulator()

	event1 := NewEventBuilder("event1").Build()
	event2 := NewEventBuilder("event2").Build()

	accumulator.Accumulate("reality1", event1)
	accumulator.Accumulate("reality1", event2)

	stats := accumulator.GetStatistics()
	realityEvents := stats.GetRealityEvents("reality1")

	if len(realityEvents) != 2 {
		t.Fatalf("Expected 2 unique event names, got %d", len(realityEvents))
	}

	// Should contain the most recent event for each name
	if realityEvents["event1"] != event1 {
		t.Fatal("Expected event1 to be the stored event")
	}
	if realityEvents["event2"] != event2 {
		t.Fatal("Expected event2 to be the stored event")
	}
}

func TestEventAccumulator_GetRealityEvents_NoEvents(t *testing.T) {
	accumulator := newEventAccumulator()

	stats := accumulator.GetStatistics()
	realityEvents := stats.GetRealityEvents("nonexistent")

	if len(realityEvents) != 0 {
		t.Fatalf("Expected empty map for nonexistent reality, got %d events", len(realityEvents))
	}
}

func TestEventAccumulator_GetAllEventsNames(t *testing.T) {
	accumulator := newEventAccumulator()

	event1 := NewEventBuilder("event1").Build()
	event2 := NewEventBuilder("event2").Build()
	event3 := NewEventBuilder("event3").Build()

	accumulator.Accumulate("reality1", event1)
	accumulator.Accumulate("reality1", event2)
	accumulator.Accumulate("reality2", event3)
	accumulator.Accumulate("reality2", event1) // duplicate across realities

	stats := accumulator.GetStatistics()
	eventNames := stats.GetAllEventsNames()

	if len(eventNames) != 3 {
		t.Fatalf("Expected 3 unique event names, got %d", len(eventNames))
	}

	// Check all expected names are present
	nameMap := make(map[string]bool)
	for _, name := range eventNames {
		nameMap[name] = true
	}

	if !nameMap["event1"] || !nameMap["event2"] || !nameMap["event3"] {
		t.Fatalf("Expected events 'event1', 'event2', 'event3', got %v", eventNames)
	}
}

func TestEventAccumulator_CountAllEventsNames(t *testing.T) {
	accumulator := newEventAccumulator()

	event1 := NewEventBuilder("event1").Build()
	event2 := NewEventBuilder("event2").Build()

	accumulator.Accumulate("reality1", event1)
	accumulator.Accumulate("reality1", event2)
	accumulator.Accumulate("reality2", event1) // duplicate

	stats := accumulator.GetStatistics()
	count := stats.CountAllEventsNames()

	if count != 2 {
		t.Fatalf("Expected 2 unique event names, got %d", count)
	}
}

func TestEventAccumulator_CountAllEvents(t *testing.T) {
	accumulator := newEventAccumulator()

	event1 := NewEventBuilder("event1").Build()
	event2 := NewEventBuilder("event2").Build()

	accumulator.Accumulate("reality1", event1)
	accumulator.Accumulate("reality1", event2)
	accumulator.Accumulate("reality2", event1)

	stats := accumulator.GetStatistics()
	totalCount := stats.CountAllEvents()

	if totalCount != 3 {
		t.Fatalf("Expected 3 total events, got %d", totalCount)
	}
}

func TestEventAccumulator_CountAllEvents_Empty(t *testing.T) {
	accumulator := newEventAccumulator()

	stats := accumulator.GetStatistics()
	totalCount := stats.CountAllEvents()

	if totalCount != 0 {
		t.Fatalf("Expected 0 total events for empty accumulator, got %d", totalCount)
	}
}

func TestEventAccumulator_String(t *testing.T) {
	accumulator := newEventAccumulator()
	concreteAccumulator, ok := accumulator.(*eventAccumulator)
	if !ok {
		t.Fatal("Expected accumulator to be of type *eventAccumulator")
	}

	event1 := NewEventBuilder("event1").Build()
	event2 := NewEventBuilder("event2").Build()

	accumulator.Accumulate("reality1", event1)
	accumulator.Accumulate("reality2", event2)

	str := concreteAccumulator.String()

	if str == "" {
		t.Fatal("Expected non-empty string representation")
	}

	// Should contain reality names
	if len(str) < 20 { // rough check for meaningful content
		t.Fatalf("Expected meaningful string representation, got: %s", str)
	}
}
