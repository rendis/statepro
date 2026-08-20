package experimental

import (
	"testing"
)

func TestProcessReference_UniverseReality(t *testing.T) {
	ref := "U:testUniverse:testReality"
	refType, parts, err := processReference(ref)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if refType != RefTypeUniverseReality {
		t.Fatalf("Expected RefTypeUniverseReality, got %v", refType)
	}
	if len(parts) != 2 {
		t.Fatalf("Expected 2 parts, got %d", len(parts))
	}
	if parts[0] != "testUniverse" {
		t.Fatalf("Expected universe 'testUniverse', got %s", parts[0])
	}
	if parts[1] != "testReality" {
		t.Fatalf("Expected reality 'testReality', got %s", parts[1])
	}
}

func TestProcessReference_Universe(t *testing.T) {
	ref := "U:testUniverse"
	refType, parts, err := processReference(ref)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if refType != RefTypeUniverse {
		t.Fatalf("Expected RefTypeUniverse, got %v", refType)
	}
	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}
	if parts[0] != "testUniverse" {
		t.Fatalf("Expected universe 'testUniverse', got %s", parts[0])
	}
}

func TestProcessReference_Reality(t *testing.T) {
	ref := "testReality"
	refType, parts, err := processReference(ref)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if refType != RefTypeReality {
		t.Fatalf("Expected RefTypeReality, got %v", refType)
	}
	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}
	if parts[0] != "testReality" {
		t.Fatalf("Expected reality 'testReality', got %s", parts[0])
	}
}

func TestProcessReference_Invalid(t *testing.T) {
	invalidRefs := []string{
		"",
		"U:",
		"U:universe:",
		"U::reality",
		"U:universe:reality:extra",
		"u:universe:reality",    // lowercase u
		"U_universe reality",    // contains space
		"123invalid",            // starts with number
		"invalid!",              // contains invalid character
		"U:123universe:reality", // universe starts with number
	}

	for _, ref := range invalidRefs {
		_, _, err := processReference(ref)
		if err == nil {
			t.Fatalf("Expected error for invalid ref '%s', got nil", ref)
		}
	}
}

func TestProcessReference_SingleCharacterIDs(t *testing.T) {
	tests := []struct {
		ref     string
		refType refType
		parts   []string
	}{
		{ref: "U:a", refType: RefTypeUniverse, parts: []string{"a"}},
		{ref: "U:a:b", refType: RefTypeUniverseReality, parts: []string{"a", "b"}},
		{ref: "a", refType: RefTypeReality, parts: []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			gotType, parts, err := processReference(tt.ref)
			if err != nil {
				t.Fatalf("processReference(%q) error: %v — schema allows min length 1", tt.ref, err)
			}
			if gotType != tt.refType {
				t.Fatalf("ref type: got %v, want %v", gotType, tt.refType)
			}
			if len(parts) != len(tt.parts) {
				t.Fatalf("parts: got %v, want %v", parts, tt.parts)
			}
			for i, p := range tt.parts {
				if parts[i] != p {
					t.Fatalf("parts[%d]: got %q, want %q", i, parts[i], p)
				}
			}
		})
	}
}

func TestProcessReference_EdgeCases(t *testing.T) {
	// Test with special characters in names (allowed by regex)
	ref := "U:universe_123-test:reality_456-test"
	refType, parts, err := processReference(ref)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if refType != RefTypeUniverseReality {
		t.Fatalf("Expected RefTypeUniverseReality, got %v", refType)
	}
	if len(parts) != 2 {
		t.Fatalf("Expected 2 parts, got %d", len(parts))
	}
	if parts[0] != "universe_123-test" {
		t.Fatalf("Expected universe 'universe_123-test', got %s", parts[0])
	}
	if parts[1] != "reality_456-test" {
		t.Fatalf("Expected reality 'reality_456-test', got %s", parts[1])
	}
}
