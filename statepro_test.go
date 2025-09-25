package statepro

import (
	"testing"

	"github.com/rendis/statepro/v3/theoretical"
)

func TestNewQuantumMachine_ValidInput(t *testing.T) {
	model := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": {ID: "u1", CanonicalName: "Universe1"},
		},
		Initials: []string{"u1"},
	}

	qm, err := NewQuantumMachine(model)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if qm == nil {
		t.Fatal("Expected non-nil QuantumMachine")
	}
}

func TestNewQuantumMachine_NilModel(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for nil model")
		}
	}()
	NewQuantumMachine(nil)
}

func TestNewEventBuilder(t *testing.T) {
	eb := NewEventBuilder("testEvent")
	if eb == nil {
		t.Fatal("Expected non-nil EventBuilder")
	}
}

func TestDeserializeQuantumMachineFromMap_Valid(t *testing.T) {
	input := map[string]any{
		"id":            "qm1",
		"canonicalName": "QuantumMachine1",
		"version":       "1.0.0",
		"universes": map[string]any{
			"u1": map[string]any{
				"id":            "u1",
				"canonicalName": "Universe1",
			},
		},
		"initials": []any{"u1"},
	}
	model, err := DeserializeQuantumMachineFromMap(input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Fatal("Expected non-nil model")
	}
	if len(model.Universes) != 1 {
		t.Fatalf("Expected 1 universe, got %d", len(model.Universes))
	}
}

func TestDeserializeQuantumMachineFromMap_Nil(t *testing.T) {
	model, err := DeserializeQuantumMachineFromMap(nil)
	if err != nil {
		t.Fatalf("Expected no error for nil input, got %v", err)
	}
	if model != nil {
		t.Fatal("Expected nil model for nil input")
	}
}

func TestDeserializeQuantumMachineFromMap_Invalid(t *testing.T) {
	input := map[string]any{
		"universes": "invalid_type",
	}
	_, err := DeserializeQuantumMachineFromMap(input)
	if err == nil {
		t.Fatal("Expected error for invalid input")
	}
}

func TestDeserializeQuantumMachineFromBinary_Valid(t *testing.T) {
	input := `{"id":"qm1","canonicalName":"QuantumMachine1","version":"1.0.0","universes":{"u1":{"id":"u1","canonicalName":"Universe1"}},"initials":["u1"]}`
	model, err := DeserializeQuantumMachineFromBinary([]byte(input))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if model == nil {
		t.Fatal("Expected non-nil model")
	}
}

func TestDeserializeQuantumMachineFromBinary_Invalid(t *testing.T) {
	input := `invalid json`
	_, err := DeserializeQuantumMachineFromBinary([]byte(input))
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

func TestSerializeQuantumMachineToMap_Valid(t *testing.T) {
	model := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": {ID: "u1", CanonicalName: "Universe1"},
		},
		Initials: []string{"u1"},
	}
	m, err := SerializeQuantumMachineToMap(model)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if m == nil {
		t.Fatal("Expected non-nil map")
	}
}

func TestSerializeQuantumMachineToMap_Nil(t *testing.T) {
	m, err := SerializeQuantumMachineToMap(nil)
	if err != nil {
		t.Fatalf("Expected no error for nil model, got %v", err)
	}
	if m != nil {
		t.Fatal("Expected nil map for nil model")
	}
}

func TestSerializeQuantumMachineToBinary_Valid(t *testing.T) {
	model := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": {ID: "u1", CanonicalName: "Universe1"},
		},
		Initials: []string{"u1"},
	}
	b, err := SerializeQuantumMachineToBinary(model)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(b) == 0 {
		t.Fatal("Expected non-empty bytes")
	}
}

func TestSerializeQuantumMachineToBinary_Nil(t *testing.T) {
	b, err := SerializeQuantumMachineToBinary(nil)
	if err != nil {
		t.Fatalf("Expected no error for nil model, got %v", err)
	}
	if b != nil {
		t.Fatal("Expected nil bytes for nil model")
	}
}
