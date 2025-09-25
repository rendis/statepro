package builtin

import (
	"context"
	"testing"

	"github.com/rendis/statepro/instrumentation"
)

func TestRegisterObserver_Valid(t *testing.T) {
	fn := func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) { return true, nil }
	err := RegisterObserver("custom:observer:test", fn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Check if registered
	retrieved := GetObserver("custom:observer:test")
	if retrieved == nil {
		t.Fatal("Expected function to be registered")
	}
}

func TestRegisterObserver_InvalidSrc(t *testing.T) {
	fn := func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) { return true, nil }
	err := RegisterObserver("invalid src", fn)
	if err == nil {
		t.Fatal("Expected error for invalid src")
	}
}

func TestRegisterAction_Valid(t *testing.T) {
	fn := func(ctx context.Context, args instrumentation.ActionExecutorArgs) error { return nil }
	err := RegisterAction("custom:action:test", fn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	retrieved := GetAction("custom:action:test")
	if retrieved == nil {
		t.Fatal("Expected function to be registered")
	}
}

func TestRegisterAction_InvalidSrc(t *testing.T) {
	fn := func(ctx context.Context, args instrumentation.ActionExecutorArgs) error { return nil }
	err := RegisterAction("invalid src", fn)
	if err == nil {
		t.Fatal("Expected error for invalid src")
	}
}

func TestRegisterInvoke_Valid(t *testing.T) {
	fn := func(ctx context.Context, args instrumentation.InvokeExecutorArgs) {}
	err := RegisterInvoke("custom:invoke:test", fn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	retrieved := GetInvoke("custom:invoke:test")
	if retrieved == nil {
		t.Fatal("Expected function to be registered")
	}
}

func TestRegisterInvoke_InvalidSrc(t *testing.T) {
	fn := func(ctx context.Context, args instrumentation.InvokeExecutorArgs) {}
	err := RegisterInvoke("invalid src", fn)
	if err == nil {
		t.Fatal("Expected error for invalid src")
	}
}

func TestRegisterCondition_Valid(t *testing.T) {
	fn := func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) { return true, nil }
	err := RegisterCondition("custom:condition:test", fn)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	retrieved := GetCondition("custom:condition:test")
	if retrieved == nil {
		t.Fatal("Expected function to be registered")
	}
}

func TestRegisterCondition_InvalidSrc(t *testing.T) {
	fn := func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) { return true, nil }
	err := RegisterCondition("invalid src", fn)
	if err == nil {
		t.Fatal("Expected error for invalid src")
	}
}

func TestNormalizeSrc_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"valid:name", "valid:name"},
		{"a1.b-2_c:3", "a1.b-2_c:3"},
		{"  spaced  ", "spaced"},
		{"Test123", "Test123"},
	}

	for _, tt := range tests {
		result, err := normalizeSrc(tt.input)
		if err != nil {
			t.Errorf("normalizeSrc(%q) returned error: %v", tt.input, err)
		}
		if result != tt.expected {
			t.Errorf("normalizeSrc(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeSrc_Invalid(t *testing.T) {
	tests := []string{
		"",           // empty
		" ",          // only space
		"_invalid",   // starts with underscore
		"invalid-",   // ends with dash
		"123invalid", // starts with number
		"invalid.",   // ends with dot
		"inv@lid",    // contains invalid character
	}

	for _, input := range tests {
		_, err := normalizeSrc(input)
		if err == nil {
			t.Errorf("normalizeSrc(%q) should have returned error", input)
		}
	}
}
