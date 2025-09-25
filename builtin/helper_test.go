package builtin

import (
	"testing"
)

func TestTryToCastToInt_Int(t *testing.T) {
	v, ok := TryToCastToInt(42)
	if !ok || v != 42 {
		t.Fatalf("Expected 42, true; got %d, %v", v, ok)
	}
}

func TestTryToCastToInt_String(t *testing.T) {
	v, ok := TryToCastToInt("123")
	if !ok || v != 123 {
		t.Fatalf("Expected 123, true; got %d, %v", v, ok)
	}
}

func TestTryToCastToInt_InvalidString(t *testing.T) {
	_, ok := TryToCastToInt("abc")
	if ok {
		t.Fatal("Expected false for invalid string")
	}
}

func TestTryToCastToInt_Float(t *testing.T) {
	v, ok := TryToCastToInt(3.14)
	if !ok || v != 3 {
		t.Fatalf("Expected 3, true; got %d, %v", v, ok)
	}
}

func TestGetKeyAsInt_Exists(t *testing.T) {
	md := map[string]any{"key": 42}
	v, ok := GetKeyAsInt("key", md)
	if !ok || v != 42 {
		t.Fatalf("Expected 42, true; got %d, %v", v, ok)
	}
}

func TestGetKeyAsInt_NotExists(t *testing.T) {
	md := map[string]any{}
	_, ok := GetKeyAsInt("key", md)
	if ok {
		t.Fatal("Expected false for missing key")
	}
}

func TestTryToCastToString_String(t *testing.T) {
	v, ok := TryToCastToString("test")
	if !ok || v != "test" {
		t.Fatalf("Expected 'test', true; got %s, %v", v, ok)
	}
}

func TestTryToCastToString_Int(t *testing.T) {
	v, ok := TryToCastToString(42)
	if !ok || v != "42" {
		t.Fatalf("Expected '42', true; got %s, %v", v, ok)
	}
}

func TestTryToCastToString_Bool(t *testing.T) {
	v, ok := TryToCastToString(true)
	if !ok || v != "true" {
		t.Fatalf("Expected 'true', true; got %s, %v", v, ok)
	}
}

func TestTryToCastToString_Other(t *testing.T) {
	_, ok := TryToCastToString([]int{1, 2})
	if ok {
		t.Fatal("Expected false for unsupported type")
	}
}

func TestTryToCastToString_Float(t *testing.T) {
	v, ok := TryToCastToString(3.14)
	if !ok || v != "3.140000" {
		t.Fatalf("Expected '3.140000', true; got %s, %v", v, ok)
	}
}

func TestTryToCastToString_Float32(t *testing.T) {
	v, ok := TryToCastToString(float32(2.5))
	if !ok || v != "2.500000" {
		t.Fatalf("Expected '2.500000', true; got %s, %v", v, ok)
	}
}

func TestGetKeyAsString_Exists(t *testing.T) {
	md := map[string]any{"key": "value"}
	v, ok := GetKeyAsString("key", md)
	if !ok || v != "value" {
		t.Fatalf("Expected 'value', true; got %s, %v", v, ok)
	}
}

func TestGetKeyAsString_NotExists(t *testing.T) {
	md := map[string]any{}
	_, ok := GetKeyAsString("key", md)
	if ok {
		t.Fatal("Expected false for missing key")
	}
}
