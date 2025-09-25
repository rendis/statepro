package experimental

import (
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// mockAccumulatorStatistics implements instrumentation.AccumulatorStatistics for testing
type mockAccumulatorStatistics struct{}

func (m *mockAccumulatorStatistics) GetRealitiesEvents() map[string][]instrumentation.Event {
	return map[string][]instrumentation.Event{}
}

func (m *mockAccumulatorStatistics) GetAllRealityEvents(realityName string) map[string][]instrumentation.Event {
	return map[string][]instrumentation.Event{}
}

func (m *mockAccumulatorStatistics) GetRealityEvents(realityName string) map[string]instrumentation.Event {
	return map[string]instrumentation.Event{}
}

func (m *mockAccumulatorStatistics) GetAllEventsNames() []string {
	return []string{}
}

func (m *mockAccumulatorStatistics) CountAllEventsNames() int {
	return 0
}

func (m *mockAccumulatorStatistics) CountAllEvents() int {
	return 0
}

// mockSnapshot implements *instrumentation.MachineSnapshot for testing
type mockSnapshot struct{}

func (m *mockSnapshot) AddActiveUniverse(canonicalName, realityName string)            {}
func (m *mockSnapshot) AddFinalizedUniverse(canonicalName, realityName string)         {}
func (m *mockSnapshot) AddSuperpositionUniverse(canonicalName, before string)          {}
func (m *mockSnapshot) AddSuperpositionUniverseFinalized(canonicalName, before string) {}
func (m *mockSnapshot) AddUniverseSnapshot(id string, snapshot instrumentation.SerializedUniverseSnapshot) {
}
func (m *mockSnapshot) AddTracking(id string, tracking []string)     {}
func (m *mockSnapshot) GetResume() interface{}                       { return nil }
func (m *mockSnapshot) GetActiveUniverses() []string                 { return nil }
func (m *mockSnapshot) GetFinalizedUniverses() []string              { return nil }
func (m *mockSnapshot) GetSuperpositionUniverses() map[string]string { return nil }
func (m *mockSnapshot) GetTracking() map[string][]string             { return nil }
func (m *mockSnapshot) ToJson() ([]byte, error)                      { return nil, nil }

func TestObserverExecutorArgs_Getters(t *testing.T) {
	event := NewEventBuilder("test").Build()
	stats := &mockAccumulatorStatistics{}
	observer := theoretical.ObserverModel{Src: "test"}

	args := &observerExecutorArgs{
		context:               "testContext",
		realityName:           "testReality",
		universeCanonicalName: "testUniverse",
		universeID:            "universe1",
		universeMetadata:      map[string]any{"key": "value"},
		accumulatorStatistics: stats,
		event:                 event,
		observer:              observer,
	}

	if args.GetContext() != "testContext" {
		t.Error("GetContext failed")
	}
	if args.GetRealityName() != "testReality" {
		t.Error("GetRealityName failed")
	}
	if args.GetUniverseCanonicalName() != "testUniverse" {
		t.Error("GetUniverseCanonicalName failed")
	}
	if args.GetUniverseId() != "universe1" {
		t.Error("GetUniverseId failed")
	}
	if args.GetEvent() != event {
		t.Error("GetEvent failed")
	}
	// Test that GetObserver returns the expected observer
	obs := args.GetObserver()
	if obs.Src != "test" {
		t.Error("GetObserver failed")
	}
	if args.GetUniverseMetadata()["key"] != "value" {
		t.Error("GetUniverseMetadata failed")
	}
}

func TestObserverExecutorArgs_MetadataOperations(t *testing.T) {
	args := &observerExecutorArgs{
		universeMetadata: map[string]any{"existing": "value"},
	}

	// Test AddToUniverseMetadata
	args.AddToUniverseMetadata("newKey", "newValue")
	if args.universeMetadata["newKey"] != "newValue" {
		t.Error("AddToUniverseMetadata failed")
	}

	// Test DeleteFromUniverseMetadata
	value, ok := args.DeleteFromUniverseMetadata("existing")
	if !ok || value != "value" {
		t.Error("DeleteFromUniverseMetadata failed")
	}
	if args.universeMetadata["existing"] != nil {
		t.Error("DeleteFromUniverseMetadata did not remove key")
	}

	// Test UpdateUniverseMetadata
	newMetadata := map[string]any{"updated": "metadata"}
	args.UpdateUniverseMetadata(newMetadata)
	if args.universeMetadata["updated"] != "metadata" {
		t.Error("UpdateUniverseMetadata failed")
	}
}

func TestActionExecutorArgs_Getters(t *testing.T) {
	event := NewEventBuilder("test").Build()
	action := theoretical.ActionModel{Src: "testAction"}

	args := &actionExecutorArgs{
		context:               "testContext",
		realityName:           "testReality",
		universeCanonicalName: "testUniverse",
		universeID:            "universe1",
		universeMetadata:      map[string]any{"key": "value"},
		event:                 event,
		action:                action,
		actionType:            instrumentation.ActionTypeEntry,
		getSnapshotFn:         func() *instrumentation.MachineSnapshot { return nil },
	}

	if args.GetContext() != "testContext" {
		t.Error("GetContext failed")
	}
	if args.GetRealityName() != "testReality" {
		t.Error("GetRealityName failed")
	}
	if args.GetUniverseCanonicalName() != "testUniverse" {
		t.Error("GetUniverseCanonicalName failed")
	}
	if args.GetUniverseId() != "universe1" {
		t.Error("GetUniverseId failed")
	}
	if args.GetEvent() == nil {
		t.Error("GetEvent failed")
	}
	if args.GetAction().Src != "testAction" {
		t.Error("GetAction failed")
	}
	if args.GetActionType() != instrumentation.ActionTypeEntry {
		t.Error("GetActionType failed")
	}
	if args.GetSnapshot() != nil {
		t.Error("GetSnapshot should return nil for test")
	}
}

func TestActionExecutorArgs_MetadataOperations(t *testing.T) {
	args := &actionExecutorArgs{
		universeMetadata: map[string]any{"existing": "value"},
	}

	// Test AddToUniverseMetadata
	args.AddToUniverseMetadata("newKey", "newValue")
	if args.universeMetadata["newKey"] != "newValue" {
		t.Error("AddToUniverseMetadata failed")
	}

	// Test DeleteFromUniverseMetadata
	value, ok := args.DeleteFromUniverseMetadata("existing")
	if !ok || value != "value" {
		t.Error("DeleteFromUniverseMetadata failed")
	}

	// Test UpdateUniverseMetadata
	newMetadata := map[string]any{"updated": "metadata"}
	args.UpdateUniverseMetadata(newMetadata)
	if args.universeMetadata["updated"] != "metadata" {
		t.Error("UpdateUniverseMetadata failed")
	}
}

func TestInvokeExecutorArgs_Getters(t *testing.T) {
	event := NewEventBuilder("test").Build()
	invoke := theoretical.InvokeModel{Src: "testInvoke"}

	args := &invokeExecutorArgs{
		context:               "testContext",
		realityName:           "testReality",
		universeCanonicalName: "testUniverse",
		universeID:            "universe1",
		universeMetadata:      map[string]any{"key": "value"},
		event:                 event,
		invoke:                invoke,
	}

	if args.GetContext() != "testContext" {
		t.Error("GetContext failed")
	}
	if args.GetRealityName() != "testReality" {
		t.Error("GetRealityName failed")
	}
	if args.GetUniverseCanonicalName() != "testUniverse" {
		t.Error("GetUniverseCanonicalName failed")
	}
	if args.GetUniverseId() != "universe1" {
		t.Error("GetUniverseId failed")
	}
	if args.GetEvent() != event {
		t.Error("GetEvent failed")
	}
	if args.GetInvoke().Src != "testInvoke" {
		t.Error("GetInvoke failed")
	}
}

func TestConditionExecutorArgs_Getters(t *testing.T) {
	event := NewEventBuilder("test").Build()
	condition := theoretical.ConditionModel{Src: "testCondition"}

	args := &conditionExecutorArgs{
		context:               "testContext",
		realityName:           "testReality",
		universeCanonicalName: "testUniverse",
		universeID:            "universe1",
		universeMetadata:      map[string]any{"key": "value"},
		event:                 event,
		condition:             condition,
	}

	if args.GetContext() != "testContext" {
		t.Error("GetContext failed")
	}
	if args.GetRealityName() != "testReality" {
		t.Error("GetRealityName failed")
	}
	if args.GetUniverseCanonicalName() != "testUniverse" {
		t.Error("GetUniverseCanonicalName failed")
	}
	if args.GetUniverseId() != "universe1" {
		t.Error("GetUniverseId failed")
	}
	if args.GetEvent() != event {
		t.Error("GetEvent failed")
	}
	if args.GetCondition().Src != "testCondition" {
		t.Error("GetCondition failed")
	}
}
