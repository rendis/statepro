package builtin

import (
	"context"
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// Mock implementations for testing observers

type mockEvent struct {
	name    string
	data    map[string]any
	evtType instrumentation.EventType
	flags   instrumentation.EventFlags
}

func (e *mockEvent) GetEventName() string    { return e.name }
func (e *mockEvent) GetData() map[string]any { return e.data }
func (e *mockEvent) DataContainsKey(key string) bool {
	if e.data == nil {
		return false
	}
	_, ok := e.data[key]
	return ok
}
func (e *mockEvent) GetEvtType() instrumentation.EventType { return e.evtType }
func (e *mockEvent) GetFlags() instrumentation.EventFlags  { return e.flags }

type mockAccumulatorStatistics struct {
	events map[string][]instrumentation.Event
}

func (m *mockAccumulatorStatistics) GetRealitiesEvents() map[string][]instrumentation.Event {
	return m.events
}

func (m *mockAccumulatorStatistics) GetRealityEvents(realityName string) map[string]instrumentation.Event {
	result := make(map[string]instrumentation.Event)
	if events, ok := m.events[realityName]; ok {
		for _, event := range events {
			result[event.GetEventName()] = event
		}
	}
	return result
}

func (m *mockAccumulatorStatistics) GetAllRealityEvents(realityName string) map[string][]instrumentation.Event {
	result := make(map[string][]instrumentation.Event)
	if events, ok := m.events[realityName]; ok {
		eventMap := make(map[string][]instrumentation.Event)
		for _, event := range events {
			eventMap[event.GetEventName()] = append(eventMap[event.GetEventName()], event)
		}
		result = eventMap
	}
	return result
}

func (m *mockAccumulatorStatistics) GetAllEventsNames() []string {
	names := make(map[string]bool)
	for _, events := range m.events {
		for _, event := range events {
			names[event.GetEventName()] = true
		}
	}
	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	return result
}

func (m *mockAccumulatorStatistics) CountAllEventsNames() int {
	return len(m.GetAllEventsNames())
}

func (m *mockAccumulatorStatistics) CountAllEvents() int {
	count := 0
	for _, events := range m.events {
		count += len(events)
	}
	return count
}

type mockObserverExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeId            string
	accumulatorStats      instrumentation.AccumulatorStatistics
	event                 instrumentation.Event
	observer              theoretical.ObserverModel
	universeMetadata      map[string]any
}

func (m *mockObserverExecutorArgs) GetContext() any                  { return m.context }
func (m *mockObserverExecutorArgs) GetRealityName() string           { return m.realityName }
func (m *mockObserverExecutorArgs) GetUniverseCanonicalName() string { return m.universeCanonicalName }
func (m *mockObserverExecutorArgs) GetUniverseId() string            { return m.universeId }
func (m *mockObserverExecutorArgs) GetAccumulatorStatistics() instrumentation.AccumulatorStatistics {
	return m.accumulatorStats
}
func (m *mockObserverExecutorArgs) GetEvent() instrumentation.Event        { return m.event }
func (m *mockObserverExecutorArgs) GetObserver() theoretical.ObserverModel { return m.observer }
func (m *mockObserverExecutorArgs) GetUniverseMetadata() map[string]any    { return m.universeMetadata }
func (m *mockObserverExecutorArgs) AddToUniverseMetadata(key string, value any) {
	if m.universeMetadata == nil {
		m.universeMetadata = make(map[string]any)
	}
	m.universeMetadata[key] = value
}
func (m *mockObserverExecutorArgs) DeleteFromUniverseMetadata(key string) (any, bool) {
	if m.universeMetadata == nil {
		return nil, false
	}
	value, ok := m.universeMetadata[key]
	if ok {
		delete(m.universeMetadata, key)
	}
	return value, ok
}
func (m *mockObserverExecutorArgs) UpdateUniverseMetadata(md map[string]any) {
	m.universeMetadata = md
}

func TestAlwaysTrue(t *testing.T) {
	ctx := context.Background()
	args := &mockObserverExecutorArgs{}

	result, err := AlwaysTrue(ctx, args)
	if err != nil {
		t.Fatalf("AlwaysTrue should not return error, got %v", err)
	}
	if !result {
		t.Fatal("AlwaysTrue should return true")
	}
}

func TestContainsAllEvents_AllPresent(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}
	event2 := &mockEvent{name: "event2"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1, event2},
		},
	}

	// Create observer args
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"0": "event1", "1": "event2"}, // expected events
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAllEvents(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAllEvents should not return error, got %v", err)
	}
	if !result {
		t.Fatal("ContainsAllEvents should return true when all events are present")
	}
}

func TestContainsAllEvents_SomeMissing(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"0": "event1", "1": "event2"}, // expected events, event2 missing
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAllEvents(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAllEvents should not return error, got %v", err)
	}
	if result {
		t.Fatal("ContainsAllEvents should return false when some events are missing")
	}
}

func TestContainsAllEvents_NoAccumulatedEvents(t *testing.T) {
	ctx := context.Background()

	// Create mock accumulator statistics with no events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {}, // empty events
		},
	}

	// Create observer args
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"0": "event1"}, // expected events
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAllEvents(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAllEvents should not return error, got %v", err)
	}
	if result {
		t.Fatal("ContainsAllEvents should return false when no events are accumulated")
	}
}

func TestContainsAllEvents_EmptyArgs(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args with empty Args
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{}, // empty args
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAllEvents(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAllEvents should not return error, got %v", err)
	}
	if !result {
		t.Fatal("ContainsAllEvents should return true when no events are required (empty args)")
	}
}

func TestContainsAtLeastOneEvent_Present(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"0": "event1", "1": "event2"}, // event1 is present
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAtLeastOneEvent(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAtLeastOneEvent should not return error, got %v", err)
	}
	if !result {
		t.Fatal("ContainsAtLeastOneEvent should return true when at least one event is present")
	}
}

func TestContainsAtLeastOneEvent_NoAccumulatedEvents(t *testing.T) {
	ctx := context.Background()

	// Create mock accumulator statistics with no events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {}, // empty events
		},
	}

	// Create observer args
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"0": "event1"}, // expected events
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAtLeastOneEvent(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAtLeastOneEvent should not return error, got %v", err)
	}
	if result {
		t.Fatal("ContainsAtLeastOneEvent should return false when no events are accumulated")
	}
}

func TestContainsAtLeastOneEvent_NoneFound(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args with events that don't match
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"0": "event2", "1": "event3"}, // none of these are present
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAtLeastOneEvent(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAtLeastOneEvent should not return error, got %v", err)
	}
	if result {
		t.Fatal("ContainsAtLeastOneEvent should return false when none of the expected events are found")
	}
}

func TestGreaterThanEqualCounter_Sufficient(t *testing.T) {
	ctx := context.Background()

	// Create mock events (multiple occurrences)
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1, event1, event1}, // 3 occurrences
		},
	}

	// Create observer args
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"event1": 2}, // requires at least 2
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := GreaterThanEqualCounter(ctx, args)
	if err != nil {
		t.Fatalf("GreaterThanEqualCounter should not return error, got %v", err)
	}
	if !result {
		t.Fatal("GreaterThanEqualCounter should return true when count is sufficient")
	}
}

func TestGreaterThanEqualCounter_NoAccumulatedEvents(t *testing.T) {
	ctx := context.Background()

	// Create mock accumulator statistics with no events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {}, // empty events
		},
	}

	// Create observer args
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"event1": 1}, // requires at least 1
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := GreaterThanEqualCounter(ctx, args)
	if err != nil {
		t.Fatalf("GreaterThanEqualCounter should not return error, got %v", err)
	}
	if result {
		t.Fatal("GreaterThanEqualCounter should return false when no events are accumulated")
	}
}

func TestGreaterThanEqualCounter_EventNotFound(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1}, // only event1
		},
	}

	// Create observer args looking for event2 (which doesn't exist)
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"event2": 1}, // requires at least 1 of event2
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := GreaterThanEqualCounter(ctx, args)
	if err != nil {
		t.Fatalf("GreaterThanEqualCounter should not return error, got %v", err)
	}
	if result {
		t.Fatal("GreaterThanEqualCounter should return false when the required event is not found")
	}
}

func TestGreaterThanEqualCounter_InsufficientCount(t *testing.T) {
	ctx := context.Background()

	// Create mock events (only 2 occurrences)
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1, event1}, // 2 occurrences
		},
	}

	// Create observer args requiring 3 occurrences
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"event1": 3}, // requires at least 3
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := GreaterThanEqualCounter(ctx, args)
	if err != nil {
		t.Fatalf("GreaterThanEqualCounter should not return error, got %v", err)
	}
	if result {
		t.Fatal("GreaterThanEqualCounter should return false when count is insufficient")
	}
}

func TestTotalEventsBetweenLimits_WithinRange(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}
	event2 := &mockEvent{name: "event2"}

	// Create mock accumulator statistics with events (2 total events)
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1, event2},
		},
	}

	// Create observer args with limits
	observer := theoretical.ObserverModel{
		Src: "test",
		Args: map[string]any{
			"minimum": 1,
			"maximum": 5,
		},
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		observer:         observer,
	}

	result, err := TotalEventsBetweenLimits(ctx, args)
	if err != nil {
		t.Fatalf("TotalEventsBetweenLimits should not return error, got %v", err)
	}
	if !result {
		t.Fatal("TotalEventsBetweenLimits should return true when total is within range")
	}
}

func TestContainsAllEvents_NonStringArg(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args with non-string arg
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"0": 123}, // non-string arg
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAllEvents(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAllEvents should not return error, got %v", err)
	}
	if result {
		t.Fatal("ContainsAllEvents should return false for non-string args")
	}
}

func TestContainsAtLeastOneEvent_NonStringArg(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args with non-string arg
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"0": 123}, // non-string arg
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := ContainsAtLeastOneEvent(ctx, args)
	if err != nil {
		t.Fatalf("ContainsAtLeastOneEvent should not return error, got %v", err)
	}
	if result {
		t.Fatal("ContainsAtLeastOneEvent should return false for non-string args")
	}
}

func TestGreaterThanEqualCounter_NonIntArg(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args with non-int arg
	observer := theoretical.ObserverModel{
		Src:  "test",
		Args: map[string]any{"event1": "not_an_int"}, // non-int arg
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer:         observer,
	}

	result, err := GreaterThanEqualCounter(ctx, args)
	if err != nil {
		t.Fatalf("GreaterThanEqualCounter should not return error, got %v", err)
	}
	if result {
		t.Fatal("GreaterThanEqualCounter should return false for non-int args")
	}
}

func TestTotalEventsBetweenLimits_InvalidMinCast(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events (1 total event)
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args with invalid minimum (non-castable)
	observer := theoretical.ObserverModel{
		Src: "test",
		Args: map[string]any{
			"minimum": "not_a_number",
			"maximum": 5,
		},
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		observer:         observer,
	}

	result, err := TotalEventsBetweenLimits(ctx, args)
	if err != nil {
		t.Fatalf("TotalEventsBetweenLimits should not return error, got %v", err)
	}
	// Should use default min (math.MinInt) so result should be true
	if !result {
		t.Fatal("TotalEventsBetweenLimits should return true when using default min")
	}
}

func TestTotalEventsBetweenLimits_InvalidMaxCast(t *testing.T) {
	ctx := context.Background()

	// Create mock events
	event1 := &mockEvent{name: "event1"}

	// Create mock accumulator statistics with events (1 total event)
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1},
		},
	}

	// Create observer args with invalid maximum (non-castable)
	observer := theoretical.ObserverModel{
		Src: "test",
		Args: map[string]any{
			"minimum": 0,
			"maximum": []int{1, 2, 3}, // non-castable
		},
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		observer:         observer,
	}

	result, err := TotalEventsBetweenLimits(ctx, args)
	if err != nil {
		t.Fatalf("TotalEventsBetweenLimits should not return error, got %v", err)
	}
	// Should use default max (math.MaxInt) so result should be true
	if !result {
		t.Fatal("TotalEventsBetweenLimits should return true when using default max")
	}
}
