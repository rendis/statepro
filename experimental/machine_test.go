package experimental

import (
	"testing"

	"github.com/rendis/statepro/v3/theoretical"
)

func TestNewExQuantumMachine_Valid(t *testing.T) {
	model := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": {ID: "u1", CanonicalName: "Universe1"},
		},
		Initials: []string{"u1"},
	}
	universes := []*ExUniverse{
		NewExUniverse(model.Universes["u1"]),
	}

	qm, err := NewExQuantumMachine(model, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if qm == nil {
		t.Fatal("Expected non-nil QuantumMachine")
	}
}

func TestNewExQuantumMachine_DuplicateUniverse(t *testing.T) {
	model := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": {ID: "u1", CanonicalName: "Universe1"},
		},
		Initials: []string{"u1"},
	}
	universes := []*ExUniverse{
		NewExUniverse(model.Universes["u1"]),
		NewExUniverse(model.Universes["u1"]), // duplicate
	}

	_, err := NewExQuantumMachine(model, universes)
	if err == nil {
		t.Fatal("Expected error for duplicate universe")
	}
}
