package builtin

import (
	"context"
	"testing"

	"github.com/rendis/statepro/instrumentation"
)

func TestGetObserver_Builtin(t *testing.T) {
	fn := GetObserver("builtin:observer:alwaysTrue")
	if fn == nil {
		t.Fatal("Expected builtin observer")
	}
}

func TestGetObserver_Custom(t *testing.T) {
	// First register a custom observer
	fn := func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) { return true, nil }
	RegisterObserver("custom:observer:test", fn)

	retrieved := GetObserver("custom:observer:test")
	if retrieved == nil {
		t.Fatal("Expected custom observer")
	}
}

func TestGetObserver_NotFound(t *testing.T) {
	fn := GetObserver("notfound")
	if fn != nil {
		t.Fatal("Expected nil for not found")
	}
}

func TestGetAction_Builtin(t *testing.T) {
	fn := GetAction("builtin:action:logBasicInfo")
	if fn == nil {
		t.Fatal("Expected builtin action")
	}
}

func TestGetAction_Custom(t *testing.T) {
	// First register a custom action
	fn := func(ctx context.Context, args instrumentation.ActionExecutorArgs) error { return nil }
	RegisterAction("custom:action:test", fn)

	retrieved := GetAction("custom:action:test")
	if retrieved == nil {
		t.Fatal("Expected custom action")
	}
}

func TestGetAction_NotFound(t *testing.T) {
	fn := GetAction("notfound")
	if fn != nil {
		t.Fatal("Expected nil for not found")
	}
}

func TestGetInvoke_BuiltinEmpty(t *testing.T) {
	fn := GetInvoke("builtin:invoke:any")
	if fn != nil {
		t.Fatal("Expected nil for empty builtin invoke registry")
	}
}

func TestGetInvoke_Custom(t *testing.T) {
	// First register a custom invoke
	fn := func(ctx context.Context, args instrumentation.InvokeExecutorArgs) {}
	RegisterInvoke("custom:invoke:test", fn)

	retrieved := GetInvoke("custom:invoke:test")
	if retrieved == nil {
		t.Fatal("Expected custom invoke")
	}
}

func TestGetInvoke_NotFound(t *testing.T) {
	fn := GetInvoke("notfound")
	if fn != nil {
		t.Fatal("Expected nil for not found")
	}
}

func TestGetCondition_BuiltinEmpty(t *testing.T) {
	fn := GetCondition("builtin:condition:any")
	if fn != nil {
		t.Fatal("Expected nil for empty builtin condition registry")
	}
}

func TestGetCondition_Custom(t *testing.T) {
	// First register a custom condition
	fn := func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) { return true, nil }
	RegisterCondition("custom:condition:test", fn)

	retrieved := GetCondition("custom:condition:test")
	if retrieved == nil {
		t.Fatal("Expected custom condition")
	}
}

func TestGetCondition_NotFound(t *testing.T) {
	fn := GetCondition("notfound")
	if fn != nil {
		t.Fatal("Expected nil for not found")
	}
}

func TestGetInvoke_BuiltinRegistry(t *testing.T) {
	// Save original registry
	originalRegistry := builtinInvokeRegistry

	// Create a test function
	testFn := func(ctx context.Context, args instrumentation.InvokeExecutorArgs) {}

	// Temporarily modify the builtin registry (accessible since we're in the same package)
	builtinInvokeRegistry = map[string]instrumentation.InvokeFn{
		"test:builtin:invoke": testFn,
	}

	// Test that GetInvoke returns the builtin function
	fn := GetInvoke("test:builtin:invoke")
	if fn == nil {
		t.Fatal("Expected builtin function to be returned")
	}

	// Restore original registry
	builtinInvokeRegistry = originalRegistry
}

func TestGetCondition_BuiltinRegistry(t *testing.T) {
	// Save original registry
	originalRegistry := builtinConditionRegistry

	// Create a test function
	testFn := func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) { return true, nil }

	// Temporarily modify the builtin registry (accessible since we're in the same package)
	builtinConditionRegistry = map[string]instrumentation.ConditionFn{
		"test:builtin:condition": testFn,
	}

	// Test that GetCondition returns the builtin function
	fn := GetCondition("test:builtin:condition")
	if fn == nil {
		t.Fatal("Expected builtin function to be returned")
	}

	// Restore original registry
	builtinConditionRegistry = originalRegistry
}
