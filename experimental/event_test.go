package experimental

import (
	"testing"

	"github.com/rendis/statepro/instrumentation"
)

func TestNewEventBuilder(t *testing.T) {
	eb := NewEventBuilder("testEvent")
	if eb == nil {
		t.Fatal("Expected non-nil EventBuilder")
	}
}

func TestEventBuilder_Build(t *testing.T) {
	eb := NewEventBuilder("testEvent")
	event := eb.Build()
	if event == nil {
		t.Fatal("Expected non-nil Event")
	}
	if event.GetEventName() != "testEvent" {
		t.Fatalf("Expected name 'testEvent', got %s", event.GetEventName())
	}
	if event.GetEvtType() != instrumentation.EventTypeOn {
		t.Fatalf("Expected type On, got %v", event.GetEvtType())
	}
}

func TestEventBuilder_SetData(t *testing.T) {
	eb := NewEventBuilder("testEvent")
	data := map[string]any{"key": "value"}
	eb.SetData(data)
	event := eb.Build()
	if event.GetData()["key"] != "value" {
		t.Fatal("Expected data to be set")
	}
}

func TestEventBuilder_SetEvtType(t *testing.T) {
	eb := NewEventBuilder("testEvent")
	eb.SetEvtType(instrumentation.EventTypeOnEntry)
	event := eb.Build()
	if event.GetEvtType() != instrumentation.EventTypeOnEntry {
		t.Fatalf("Expected type OnEntry, got %v", event.GetEvtType())
	}
}

func TestEventBuilder_SetFlags(t *testing.T) {
	eb := NewEventBuilder("testEvent")
	flags := instrumentation.EventFlags{ReplayOnEntry: true}
	eb.SetFlags(flags)
	event := eb.Build()
	if !event.GetFlags().ReplayOnEntry {
		t.Fatal("Expected flags to be set")
	}
}

func TestEvent_GetEventName(t *testing.T) {
	event := &Event{Name: "test"}
	if event.GetEventName() != "test" {
		t.Fatalf("Expected 'test', got %s", event.GetEventName())
	}
}

func TestEvent_GetData(t *testing.T) {
	data := map[string]any{"key": "value"}
	event := &Event{Data: data}
	if event.GetData()["key"] != "value" {
		t.Fatal("Expected data")
	}
}

func TestEvent_DataContainsKey(t *testing.T) {
	data := map[string]any{"key": "value"}
	event := &Event{Data: data}
	if !event.DataContainsKey("key") {
		t.Fatal("Expected true for existing key")
	}
	if event.DataContainsKey("missing") {
		t.Fatal("Expected false for missing key")
	}
}

func TestEvent_DataContainsKey_NilData(t *testing.T) {
	event := &Event{}
	if event.DataContainsKey("key") {
		t.Fatal("Expected false for nil data")
	}
}

func TestEvent_GetEvtType(t *testing.T) {
	event := &Event{EvtType: instrumentation.EventTypeOn}
	if event.GetEvtType() != instrumentation.EventTypeOn {
		t.Fatalf("Expected On, got %v", event.GetEvtType())
	}
}

func TestEvent_GetFlags(t *testing.T) {
	flags := instrumentation.EventFlags{ReplayOnEntry: true}
	event := &Event{Flags: flags}
	if !event.GetFlags().ReplayOnEntry {
		t.Fatal("Expected flags")
	}
}

func TestEvent_String(t *testing.T) {
	event := &Event{Name: "test"}
	if event.String() != "test" {
		t.Fatalf("Expected 'test', got %s", event.String())
	}
}
