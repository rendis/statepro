package bot_test

import (
	"context"
	"testing"

	"github.com/rendis/statepro/v3/debugger/bot"
	"github.com/rendis/statepro/v3/instrumentation"
)

// MockQuantumMachine is a mock implementation of instrumentation.QuantumMachine for testing.
type MockQuantumMachine struct {
	snapshot *instrumentation.MachineSnapshot
}

func (m *MockQuantumMachine) GetSnapshot() *instrumentation.MachineSnapshot {
	return m.snapshot
}

func (m *MockQuantumMachine) LoadSnapshot(s *instrumentation.MachineSnapshot, ctx any) error {
	m.snapshot = s
	return nil
}

func (m *MockQuantumMachine) Init(ctx context.Context, machineContext any) error {
	return nil
}

func (m *MockQuantumMachine) InitWithEvent(ctx context.Context, machineContext any, event instrumentation.Event) error {
	return nil
}

func (m *MockQuantumMachine) SendEvent(ctx context.Context, event instrumentation.Event) (bool, error) {
	if event.GetEventName() == "handled" {
		return true, nil
	}
	return false, nil
}

func (m *MockQuantumMachine) ReplayOnEntry(ctx context.Context) error {
	return nil
}

func (m *MockQuantumMachine) PositionMachine(ctx context.Context, machineContext any, universeID string, realityID string, executeFlow bool) error {
	return nil
}

func (m *MockQuantumMachine) PositionMachineOnInitial(ctx context.Context, machineContext any, universeID string, executeFlow bool) error {
	return nil
}

func (m *MockQuantumMachine) PositionMachineByCanonicalName(ctx context.Context, machineContext any, universeCanonicalName string, realityID string, executeFlow bool) error {
	return nil
}

func (m *MockQuantumMachine) PositionMachineOnInitialByCanonicalName(ctx context.Context, machineContext any, universeCanonicalName string, executeFlow bool) error {
	return nil
}

// MockEvent is a mock implementation of instrumentation.Event
type MockEvent struct {
	name string
}

func (e *MockEvent) GetEventName() string {
	return e.name
}

func (e *MockEvent) GetData() map[string]any {
	return nil
}

func (e *MockEvent) DataContainsKey(key string) bool {
	return false
}

func (e *MockEvent) GetEvtType() instrumentation.EventType {
	return "test"
}

func (e *MockEvent) GetFlags() instrumentation.EventFlags {
	return instrumentation.EventFlags{}
}

func TestBot_IgnoreUnhandledEvents(t *testing.T) {
	qm := &MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}}

	// Test 1: Default behavior (Should error on unhandled event)
	t.Run("Default behavior errors on unhandled event", func(t *testing.T) {
		events := []instrumentation.Event{&MockEvent{name: "unhandled"}}
		provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
			if len(events) > 0 {
				e := events[0]
				events = events[1:]
				return e, nil
			}
			return nil, nil
		}

		b, err := bot.NewBot(qm, provider, false)
		if err != nil {
			t.Fatalf("NewBot failed: %v", err)
		}

		err = b.Run(context.Background(), nil)
		if err == nil {
			t.Error("Expected error for unhandled event, got nil")
		} else if err.Error() != "event 'unhandled' was not handled" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	// Test 2: WithIgnoreUnhandledEvents(true) (Should NOT error)
	t.Run("WithIgnoreUnhandledEvents(true) ignores unhandled event", func(t *testing.T) {
		events := []instrumentation.Event{&MockEvent{name: "unhandled"}, &MockEvent{name: "handled"}}
		provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
			if len(events) > 0 {
				e := events[0]
				events = events[1:]
				return e, nil
			}
			return nil, nil
		}

		b, err := bot.NewBot(qm, provider, false, bot.WithIgnoreUnhandledEvents(true))
		if err != nil {
			t.Fatalf("NewBot failed: %v", err)
		}

		err = b.Run(context.Background(), nil)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		history := b.GetHistory()
		if len(history) != 1 {
			t.Errorf("Expected history length 1, got %d", len(history))
		}
		if history[0].Event.GetEventName() != "handled" {
			t.Errorf("Expected handled event in history, got %s", history[0].Event.GetEventName())
		}
	})
}
