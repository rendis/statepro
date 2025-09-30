package experimental

import (
	"context"
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

func TestExQuantumMachine_PositionMachine_StaticPositioning(t *testing.T) {
	// Create test model with realities
	realityModel := &theoretical.RealityModel{
		ID:   "state1",
		Type: theoretical.RealityTypeTransition,
	}

	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Realities: map[string]*theoretical.RealityModel{
			"state1": realityModel,
		},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	universes := []*ExUniverse{
		NewExUniverse(universeModel),
	}

	qm, err := NewExQuantumMachine(qmModel, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Test static positioning (no flow execution)
	err = qm.PositionMachine(ctx, machineContext, "u1", "state1", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify universe state
	exQm := qm.(*ExQuantumMachine)
	universe := exQm.universes["u1"]
	if !universe.initialized {
		t.Error("Universe should be initialized")
	}
	if universe.currentReality == nil || *universe.currentReality != "state1" {
		t.Errorf("Expected current reality 'state1', got %v", universe.currentReality)
	}
	if !universe.realityInitialized {
		t.Error("Reality should be initialized")
	}
	if universe.inSuperposition {
		t.Error("Universe should not be in superposition")
	}
}

func TestExQuantumMachine_PositionMachine_FlowExecution(t *testing.T) {
	// Create test model with entry actions
	realityModel := &theoretical.RealityModel{
		ID:           "state1",
		Type:         theoretical.RealityTypeTransition,
		EntryActions: []*theoretical.ActionModel{
			{Src: "test-action"},
		},
		EntryInvokes: []*theoretical.InvokeModel{
			{Src: "test-invoke"},
		},
	}

	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Realities: map[string]*theoretical.RealityModel{
			"state1": realityModel,
		},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	universes := []*ExUniverse{
		NewExUniverse(universeModel),
	}

	qm, err := NewExQuantumMachine(qmModel, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Test with full flow execution
	err = qm.PositionMachine(ctx, machineContext, "u1", "state1", true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify universe state
	exQm := qm.(*ExQuantumMachine)
	universe := exQm.universes["u1"]
	if !universe.initialized {
		t.Error("Universe should be initialized")
	}
	if universe.currentReality == nil || *universe.currentReality != "state1" {
		t.Errorf("Expected current reality 'state1', got %v", universe.currentReality)
	}
	if !universe.realityInitialized {
		t.Error("Reality should be initialized")
	}
}

func TestExQuantumMachine_PositionMachine_InvalidUniverse(t *testing.T) {
	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{},
		Initials:      []string{},
	}

	qm, err := NewExQuantumMachine(qmModel, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	err = qm.PositionMachine(ctx, machineContext, "nonexistent", "state1", false)
	if err == nil {
		t.Fatal("Expected error for nonexistent universe")
	}
}

func TestExQuantumMachine_PositionMachine_InvalidReality(t *testing.T) {
	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Realities:     map[string]*theoretical.RealityModel{},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	universes := []*ExUniverse{
		NewExUniverse(universeModel),
	}

	qm, err := NewExQuantumMachine(qmModel, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	err = qm.PositionMachine(ctx, machineContext, "u1", "nonexistent", false)
	if err == nil {
		t.Fatal("Expected error for nonexistent reality")
	}
}

func TestExQuantumMachine_PositionMachine_EmptyParameters(t *testing.T) {
	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{},
		Initials:      []string{},
	}

	qm, err := NewExQuantumMachine(qmModel, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Test empty universe ID
	err = qm.PositionMachine(ctx, machineContext, "", "state1", false)
	if err == nil {
		t.Fatal("Expected error for empty universeID")
	}

	// Test empty reality ID
	err = qm.PositionMachine(ctx, machineContext, "u1", "", false)
	if err == nil {
		t.Fatal("Expected error for empty realityID")
	}
}

func TestExQuantumMachine_PositionMachine_CompareStaticVsFlow(t *testing.T) {
	// This test compares static positioning vs flow execution to ensure both work correctly
	realityModel := &theoretical.RealityModel{
		ID:   "state1",
		Type: theoretical.RealityTypeTransition,
		EntryActions: []*theoretical.ActionModel{
			{Src: "test-action"},
		},
	}

	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Realities: map[string]*theoretical.RealityModel{
			"state1": realityModel,
		},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	// Test static positioning
	universes1 := []*ExUniverse{NewExUniverse(universeModel)}
	qm1, err := NewExQuantumMachine(qmModel, universes1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	err = qm1.PositionMachine(ctx, "test", "u1", "state1", false)
	if err != nil {
		t.Fatalf("Expected no error for static positioning, got %v", err)
	}

	// Test flow execution
	universes2 := []*ExUniverse{NewExUniverse(universeModel)}
	qm2, err := NewExQuantumMachine(qmModel, universes2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = qm2.PositionMachine(ctx, "test", "u1", "state1", true)
	if err != nil {
		t.Fatalf("Expected no error for flow execution, got %v", err)
	}

	// Both should result in properly positioned universes
	exQm1 := qm1.(*ExQuantumMachine)
	exQm2 := qm2.(*ExQuantumMachine)

	universe1 := exQm1.universes["u1"]
	universe2 := exQm2.universes["u1"]

	// Both universes should be in same final state (positioned correctly)
	if !universe1.initialized || !universe2.initialized {
		t.Error("Both universes should be initialized")
	}

	if !universe1.realityInitialized || !universe2.realityInitialized {
		t.Error("Both universes should have reality initialized")
	}

	if universe1.currentReality == nil || universe2.currentReality == nil {
		t.Error("Both universes should have current reality set")
	}

	if *universe1.currentReality != "state1" || *universe2.currentReality != "state1" {
		t.Error("Both universes should be in state1")
	}

	// Both should have tracking history
	if len(universe1.tracking) == 0 || len(universe2.tracking) == 0 {
		t.Error("Both universes should have tracking history")
	}
}

func TestExQuantumMachine_PositionMachineOnInitial_Success(t *testing.T) {
	// Create test model with initial state
	initialRealityModel := &theoretical.RealityModel{
		ID:   "initial-state",
		Type: theoretical.RealityTypeTransition,
	}

	initialState := "initial-state"
	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       &initialState,
		Realities: map[string]*theoretical.RealityModel{
			"initial-state": initialRealityModel,
		},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	universes := []*ExUniverse{
		NewExUniverse(universeModel),
	}

	qm, err := NewExQuantumMachine(qmModel, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Test positioning on initial state
	err = qm.PositionMachineOnInitial(ctx, machineContext, "u1", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify universe is positioned on initial state
	exQm := qm.(*ExQuantumMachine)
	universe := exQm.universes["u1"]
	if universe.currentReality == nil || *universe.currentReality != "initial-state" {
		t.Errorf("Expected current reality 'initial-state', got %v", universe.currentReality)
	}
	if !universe.initialized {
		t.Error("Universe should be initialized")
	}
}

func TestExQuantumMachine_PositionMachineOnInitial_NoInitialState(t *testing.T) {
	// Create test model WITHOUT initial state
	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       nil, // No initial state
		Realities:     map[string]*theoretical.RealityModel{},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	universes := []*ExUniverse{
		NewExUniverse(universeModel),
	}

	qm, err := NewExQuantumMachine(qmModel, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Should fail because universe has no initial state
	err = qm.PositionMachineOnInitial(ctx, machineContext, "u1", false)
	if err == nil {
		t.Fatal("Expected error for universe without initial state")
	}
}

func TestExQuantumMachine_PositionMachineOnInitial_InvalidUniverse(t *testing.T) {
	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{},
		Initials:      []string{},
	}

	qm, err := NewExQuantumMachine(qmModel, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Should fail for nonexistent universe
	err = qm.PositionMachineOnInitial(ctx, machineContext, "nonexistent", false)
	if err == nil {
		t.Fatal("Expected error for nonexistent universe")
	}
}

func TestExQuantumMachine_PositionMachineOnInitial_EmptyUniverseID(t *testing.T) {
	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{},
		Initials:      []string{},
	}

	qm, err := NewExQuantumMachine(qmModel, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Should fail for empty universeID
	err = qm.PositionMachineOnInitial(ctx, machineContext, "", false)
	if err == nil {
		t.Fatal("Expected error for empty universeID")
	}
}

func TestExQuantumMachine_PositionMachineOnInitial_WithFlowExecution(t *testing.T) {
	// Create test model with initial state and entry actions
	initialRealityModel := &theoretical.RealityModel{
		ID:   "initial-state",
		Type: theoretical.RealityTypeTransition,
		EntryActions: []*theoretical.ActionModel{
			{Src: "test-action"},
		},
	}

	initialState := "initial-state"
	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       &initialState,
		Realities: map[string]*theoretical.RealityModel{
			"initial-state": initialRealityModel,
		},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	universes := []*ExUniverse{
		NewExUniverse(universeModel),
	}

	qm, err := NewExQuantumMachine(qmModel, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Test with flow execution enabled
	err = qm.PositionMachineOnInitial(ctx, machineContext, "u1", true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify universe is positioned and initialized
	exQm := qm.(*ExQuantumMachine)
	universe := exQm.universes["u1"]
	if universe.currentReality == nil || *universe.currentReality != "initial-state" {
		t.Errorf("Expected current reality 'initial-state', got %v", universe.currentReality)
	}
	if !universe.initialized || !universe.realityInitialized {
		t.Error("Universe and reality should be initialized")
	}
}

func TestExQuantumMachine_PositionMachineByCanonicalName_Success(t *testing.T) {
	// Create test model with canonical name
	realityModel := &theoretical.RealityModel{
		ID:   "state1",
		Type: theoretical.RealityTypeTransition,
	}

	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Realities: map[string]*theoretical.RealityModel{
			"state1": realityModel,
		},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	universes := []*ExUniverse{
		NewExUniverse(universeModel),
	}

	qm, err := NewExQuantumMachine(qmModel, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Test positioning using canonical name
	err = qm.PositionMachineByCanonicalName(ctx, machineContext, "TestUniverse", "state1", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify universe is positioned correctly
	exQm := qm.(*ExQuantumMachine)
	universe := exQm.universes["u1"]
	if universe.currentReality == nil || *universe.currentReality != "state1" {
		t.Errorf("Expected current reality 'state1', got %v", universe.currentReality)
	}
	if !universe.initialized {
		t.Error("Universe should be initialized")
	}
}

func TestExQuantumMachine_PositionMachineByCanonicalName_InvalidCanonicalName(t *testing.T) {
	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{},
		Initials:      []string{},
	}

	qm, err := NewExQuantumMachine(qmModel, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Should fail for nonexistent canonical name
	err = qm.PositionMachineByCanonicalName(ctx, machineContext, "NonexistentUniverse", "state1", false)
	if err == nil {
		t.Fatal("Expected error for nonexistent canonical name")
	}
}

func TestExQuantumMachine_PositionMachineByCanonicalName_EmptyCanonicalName(t *testing.T) {
	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{},
		Initials:      []string{},
	}

	qm, err := NewExQuantumMachine(qmModel, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Should fail for empty canonical name
	err = qm.PositionMachineByCanonicalName(ctx, machineContext, "", "state1", false)
	if err == nil {
		t.Fatal("Expected error for empty canonical name")
	}
}

func TestExQuantumMachine_PositionMachineOnInitialByCanonicalName_Success(t *testing.T) {
	// Create test model with canonical name and initial state
	initialRealityModel := &theoretical.RealityModel{
		ID:   "initial-state",
		Type: theoretical.RealityTypeTransition,
	}

	initialState := "initial-state"
	universeModel := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       &initialState,
		Realities: map[string]*theoretical.RealityModel{
			"initial-state": initialRealityModel,
		},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": universeModel,
		},
		Initials: []string{"u1"},
	}

	universes := []*ExUniverse{
		NewExUniverse(universeModel),
	}

	qm, err := NewExQuantumMachine(qmModel, universes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Test positioning on initial using canonical name
	err = qm.PositionMachineOnInitialByCanonicalName(ctx, machineContext, "TestUniverse", false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify universe is positioned on initial state
	exQm := qm.(*ExQuantumMachine)
	universe := exQm.universes["u1"]
	if universe.currentReality == nil || *universe.currentReality != "initial-state" {
		t.Errorf("Expected current reality 'initial-state', got %v", universe.currentReality)
	}
	if !universe.initialized {
		t.Error("Universe should be initialized")
	}
}

func TestExQuantumMachine_PositionMachineOnInitialByCanonicalName_InvalidCanonicalName(t *testing.T) {
	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{},
		Initials:      []string{},
	}

	qm, err := NewExQuantumMachine(qmModel, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Should fail for nonexistent canonical name
	err = qm.PositionMachineOnInitialByCanonicalName(ctx, machineContext, "NonexistentUniverse", false)
	if err == nil {
		t.Fatal("Expected error for nonexistent canonical name")
	}
}

func TestExQuantumMachine_PositionMachineOnInitialByCanonicalName_EmptyCanonicalName(t *testing.T) {
	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "QuantumMachine1",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{},
		Initials:      []string{},
	}

	qm, err := NewExQuantumMachine(qmModel, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	ctx := context.Background()
	machineContext := "test-context"

	// Should fail for empty canonical name
	err = qm.PositionMachineOnInitialByCanonicalName(ctx, machineContext, "", false)
	if err == nil {
		t.Fatal("Expected error for empty canonical name")
	}
}