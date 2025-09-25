package builtin

import (
	"context"
	"testing"

	"github.com/rendis/statepro/instrumentation"
	"github.com/rendis/statepro/theoretical"
)

// Mock implementation for ActionExecutorArgs
type mockActionExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeId            string
	event                 instrumentation.Event
	action                theoretical.ActionModel
	actionType            instrumentation.ActionType
	snapshot              *instrumentation.MachineSnapshot
	universeMetadata      map[string]any
}

func (m *mockActionExecutorArgs) GetContext() any                               { return m.context }
func (m *mockActionExecutorArgs) GetRealityName() string                        { return m.realityName }
func (m *mockActionExecutorArgs) GetUniverseCanonicalName() string              { return m.universeCanonicalName }
func (m *mockActionExecutorArgs) GetUniverseId() string                         { return m.universeId }
func (m *mockActionExecutorArgs) GetEvent() instrumentation.Event               { return m.event }
func (m *mockActionExecutorArgs) GetAction() theoretical.ActionModel            { return m.action }
func (m *mockActionExecutorArgs) GetActionType() instrumentation.ActionType     { return m.actionType }
func (m *mockActionExecutorArgs) GetSnapshot() *instrumentation.MachineSnapshot { return m.snapshot }
func (m *mockActionExecutorArgs) GetUniverseMetadata() map[string]any           { return m.universeMetadata }
func (m *mockActionExecutorArgs) AddToUniverseMetadata(key string, value any) {
	if m.universeMetadata == nil {
		m.universeMetadata = make(map[string]any)
	}
	m.universeMetadata[key] = value
}
func (m *mockActionExecutorArgs) DeleteFromUniverseMetadata(key string) (any, bool) {
	if m.universeMetadata == nil {
		return nil, false
	}
	value, ok := m.universeMetadata[key]
	if ok {
		delete(m.universeMetadata, key)
	}
	return value, ok
}
func (m *mockActionExecutorArgs) UpdateUniverseMetadata(md map[string]any) {
	m.universeMetadata = md
}

func TestLogBasicInfo(t *testing.T) {
	ctx := context.Background()

	action := theoretical.ActionModel{
		Src: "test",
	}
	args := &mockActionExecutorArgs{
		realityName:           "testReality",
		universeCanonicalName: "testUniverse",
		universeId:            "universe1",
		action:                action,
		actionType:            instrumentation.ActionTypeEntry,
	}

	err := LogBasicInfo(ctx, args)
	if err != nil {
		t.Fatalf("LogBasicInfo should not return error, got %v", err)
	}
}

func TestLogArgs(t *testing.T) {
	ctx := context.Background()

	action := theoretical.ActionModel{
		Src:  "test",
		Args: map[string]any{"key": "value", "number": 42},
	}
	args := &mockActionExecutorArgs{
		realityName:           "testReality",
		universeCanonicalName: "testUniverse",
		universeId:            "universe1",
		action:                action,
		actionType:            instrumentation.ActionTypeEntry,
	}

	err := LogArgs(ctx, args)
	if err != nil {
		t.Fatalf("LogArgs should not return error, got %v", err)
	}
}

func TestLogArgsWithoutKeys(t *testing.T) {
	ctx := context.Background()

	action := theoretical.ActionModel{
		Src:  "test",
		Args: map[string]any{"key": "value", "number": 42},
	}
	args := &mockActionExecutorArgs{
		realityName:           "testReality",
		universeCanonicalName: "testUniverse",
		universeId:            "universe1",
		action:                action,
		actionType:            instrumentation.ActionTypeEntry,
	}

	err := LogArgsWithoutKeys(ctx, args)
	if err != nil {
		t.Fatalf("LogArgsWithoutKeys should not return error, got %v", err)
	}
}

func TestLogJustArgsValues(t *testing.T) {
	ctx := context.Background()

	action := theoretical.ActionModel{
		Src:  "test",
		Args: map[string]any{"key": "value", "number": 42},
	}
	args := &mockActionExecutorArgs{
		action: action,
	}

	err := LogJustArgsValues(ctx, args)
	if err != nil {
		t.Fatalf("LogJustArgsValues should not return error, got %v", err)
	}
}
