package experimental

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// ===================== System-specific Helpers =====================

// buildThreeUniverseQMFromStrings is a convenience variant that creates UniverseModels
// from initial strings and reality maps. Different signature from the shared buildThreeUniverseQM.
func buildThreeUniverseQMFromStrings(
	t *testing.T,
	u1Initial string, u1Realities map[string]*theoretical.RealityModel,
	u2Initial string, u2Realities map[string]*theoretical.RealityModel,
	u3Initial string, u3Realities map[string]*theoretical.RealityModel,
	initials []string,
) (*ExQuantumMachine, *ExUniverse, *ExUniverse, *ExUniverse) {
	t.Helper()
	um1 := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       &u1Initial,
		Realities:     u1Realities,
	}
	um2 := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Initial:       &u2Initial,
		Realities:     u2Realities,
	}
	um3 := &theoretical.UniverseModel{
		ID:            "u3",
		CanonicalName: "Universe3",
		Initial:       &u3Initial,
		Realities:     u3Realities,
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": um1,
			"u2": um2,
			"u3": um3,
		},
		Initials: initials,
	}
	u1 := NewExUniverse(um1)
	u2 := NewExUniverse(um2)
	u3 := NewExUniverse(um3)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2, u3})
	if err != nil {
		t.Fatalf("failed to build three-universe QM: %v", err)
	}
	return qm.(*ExQuantumMachine), u1, u2, u3
}

// ===================== J. Snapshot & Serialization (14 tests) =====================

func TestSnapshot_AfterInit(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snap := qm.GetSnapshot()
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Resume.ActiveUniverses == nil {
		t.Fatal("expected ActiveUniverses to be non-nil")
	}
	reality, ok := snap.Resume.ActiveUniverses["TestUniverse"]
	if !ok {
		t.Fatal("expected TestUniverse in ActiveUniverses")
	}
	if reality != "stateA" {
		t.Fatalf("expected ActiveUniverses[TestUniverse]=stateA, got %s", reality)
	}
}

func TestSnapshot_DuringSuperposition(t *testing.T) {
	// u1 transitions to external universe, enters superposition
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"U:u2:stateX"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	exQM := qm.(*ExQuantumMachine)
	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Send event to trigger transition to external universe
	evt := NewEventBuilder("go").Build()
	if _, err := exQM.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	snap := exQM.GetSnapshot()
	if snap.Resume.SuperpositionUniverses == nil {
		t.Fatal("expected SuperpositionUniverses to be non-nil")
	}
	if _, ok := snap.Resume.SuperpositionUniverses["Universe1"]; !ok {
		t.Fatal("expected Universe1 in SuperpositionUniverses")
	}
}

func TestSnapshot_AfterTransition(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	assertReality(t, u, "stateB")
	snap := qm.GetSnapshot()
	reality, ok := snap.Resume.ActiveUniverses["TestUniverse"]
	if !ok {
		t.Fatal("expected TestUniverse in ActiveUniverses after transition")
	}
	if reality != "stateB" {
		t.Fatalf("expected reality=stateB in snapshot, got %s", reality)
	}
}

func TestSnapshot_WithMetadata(t *testing.T) {
	actionName := "test:sys:snap-metadata"
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.AddToUniverseMetadata("myKey", "myValue")
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snap := qm.GetSnapshot()
	uSnap := snap.Snapshots["u1"]
	if uSnap == nil {
		t.Fatal("expected universe snapshot for u1")
	}
	md, ok := uSnap["metadata"]
	if !ok {
		t.Fatal("expected metadata key in universe snapshot")
	}
	mdMap, ok := md.(map[string]any)
	if !ok {
		t.Fatalf("expected metadata to be map[string]any, got %T", md)
	}
	if mdMap["myKey"] != "myValue" {
		t.Fatalf("expected metadata[myKey]=myValue, got %v", mdMap["myKey"])
	}
}

func TestSnapshot_LoadResume(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}

	// Build first machine, init, transition to B
	qm1, _ := buildQM(t, "stateA", realities)
	if err := qm1.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm1.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	snap := qm1.GetSnapshot()

	// Build second machine, load snapshot
	qm2, u2 := buildQM(t, "stateA", realities)
	if err := qm2.LoadSnapshot(snap, nil); err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}

	assertReality(t, u2, "stateB")
}

func TestSnapshot_LoadNil(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	err := qm.LoadSnapshot(nil, nil)
	if err != nil {
		t.Fatalf("expected no error for nil snapshot, got: %v", err)
	}
}

func TestSnapshot_LoadInvalid(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	qm, _ := buildQM(t, "stateA", realities)

	// Create a snapshot with a universe that references a reality that does not exist
	invalidSnap := &instrumentation.MachineSnapshot{
		Snapshots: map[string]instrumentation.SerializedUniverseSnapshot{
			"u1": {
				"initialized":    true,
				"currentReality": "nonExistentReality",
			},
		},
	}
	err := qm.LoadSnapshot(invalidSnap, nil)
	if err == nil {
		t.Fatal("expected error for invalid snapshot, got nil")
	}
}

func TestSnapshot_RoundTrip(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}

	// Build, init, transition
	qm1, _ := buildQM(t, "stateA", realities)
	if err := qm1.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm1.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	snap1 := qm1.GetSnapshot()

	// Load into new machine, get snapshot again
	qm2, _ := buildQM(t, "stateA", realities)
	if err := qm2.LoadSnapshot(snap1, nil); err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}
	snap2 := qm2.GetSnapshot()

	// Compare Resume fields
	j1, _ := json.Marshal(snap1.Resume)
	j2, _ := json.Marshal(snap2.Resume)
	if string(j1) != string(j2) {
		t.Fatalf("round-trip Resume mismatch:\nsnap1: %s\nsnap2: %s", j1, j2)
	}
}

func TestSnapshot_FinalizedUniverses(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"DONE"}, nil),
		),
		"DONE": newFinalReality("DONE"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snap := qm.GetSnapshot()
	if snap.Resume.FinalizedUniverses == nil {
		t.Fatal("expected FinalizedUniverses to be non-nil")
	}
	reality, ok := snap.Resume.FinalizedUniverses["TestUniverse"]
	if !ok {
		t.Fatal("expected TestUniverse in FinalizedUniverses")
	}
	if reality != "DONE" {
		t.Fatalf("expected FinalizedUniverses[TestUniverse]=DONE, got %s", reality)
	}
}

func TestSnapshot_ActiveUniverses(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snap := qm.GetSnapshot()
	if snap.Resume.ActiveUniverses == nil {
		t.Fatal("expected ActiveUniverses to be non-nil")
	}
	reality, ok := snap.Resume.ActiveUniverses["TestUniverse"]
	if !ok {
		t.Fatal("expected TestUniverse in ActiveUniverses")
	}
	if reality != "stateA" {
		t.Fatalf("expected ActiveUniverses[TestUniverse]=stateA, got %s", reality)
	}
}

func TestSnapshot_SuperpositionUniverses(t *testing.T) {
	// Universe with no initial -> starts in superposition
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		// No Initial -> starts in superposition
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withObserver("builtin:observer:alwaysTrue", nil),
			),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
		Initials:  []string{"U:u1"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	exQM := qm.(*ExQuantumMachine)
	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	assertSuperposition(t, u)
	snap := exQM.GetSnapshot()
	if snap.Resume.SuperpositionUniverses == nil {
		t.Fatal("expected SuperpositionUniverses to be non-nil")
	}
	if _, ok := snap.Resume.SuperpositionUniverses["TestUniverse"]; !ok {
		t.Fatal("expected TestUniverse in SuperpositionUniverses")
	}
}

func TestSnapshot_Tracking(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("goB", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withOnTransition("goC", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evtB := NewEventBuilder("goB").Build()
	if _, err := qm.SendEvent(context.Background(), evtB); err != nil {
		t.Fatalf("SendEvent(goB) failed: %v", err)
	}
	evtC := NewEventBuilder("goC").Build()
	if _, err := qm.SendEvent(context.Background(), evtC); err != nil {
		t.Fatalf("SendEvent(goC) failed: %v", err)
	}

	snap := qm.GetSnapshot()
	tracking := snap.Tracking["u1"]
	expected := []string{"stateA", "stateB", "stateC"}
	if len(tracking) != len(expected) {
		t.Fatalf("tracking length: got %d, want %d. tracking=%v", len(tracking), len(expected), tracking)
	}
	for i, exp := range expected {
		if tracking[i] != exp {
			t.Fatalf("tracking[%d]: got %q, want %q", i, tracking[i], exp)
		}
	}
}

func TestSnapshot_MultipleUniverses(t *testing.T) {
	// u1: active, u2: final, u3: no initial -> superposition
	u1Realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	u2Realities := map[string]*theoretical.RealityModel{
		"stateX": newTransitionReality("stateX",
			withAlways([]string{"DONE"}, nil),
		),
		"DONE": newFinalReality("DONE"),
	}
	u3Realities := map[string]*theoretical.RealityModel{
		"stateP": newTransitionReality("stateP",
			withObserver("builtin:observer:alwaysTrue", nil),
		),
	}

	um1 := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Universe1", Initial: strPtr("stateA"), Realities: u1Realities,
	}
	um2 := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Universe2", Initial: strPtr("stateX"), Realities: u2Realities,
	}
	um3 := &theoretical.UniverseModel{
		ID: "u3", CanonicalName: "Universe3",
		// No Initial -> superposition
		Realities: u3Realities,
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um1, "u2": um2, "u3": um3},
		Initials:  []string{"U:u1", "U:u2", "U:u3"},
	}
	u1 := NewExUniverse(um1)
	u2 := NewExUniverse(um2)
	u3 := NewExUniverse(um3)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2, u3})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	exQM := qm.(*ExQuantumMachine)
	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	snap := exQM.GetSnapshot()

	// u1 active
	if _, ok := snap.Resume.ActiveUniverses["Universe1"]; !ok {
		t.Fatal("expected Universe1 in ActiveUniverses")
	}
	// u2 finalized
	if _, ok := snap.Resume.FinalizedUniverses["Universe2"]; !ok {
		t.Fatal("expected Universe2 in FinalizedUniverses")
	}
	// u3 superposition
	if _, ok := snap.Resume.SuperpositionUniverses["Universe3"]; !ok {
		t.Fatal("expected Universe3 in SuperpositionUniverses")
	}
}

func TestSnapshot_AccumulatorPreserved(t *testing.T) {
	// Build a universe that enters superposition via external target, with an accumulator
	u1Model := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Universe1", Initial: strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"U:u2"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX",
				withObserver("builtin:observer:alwaysTrue", nil),
			),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	exQM := qm.(*ExQuantumMachine)
	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Transition to superposition
	evt := NewEventBuilder("go").Build()
	if _, err := exQM.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	assertSuperposition(t, u1)

	snap := exQM.GetSnapshot()
	uSnap := snap.Snapshots["u1"]
	if uSnap == nil {
		t.Fatal("expected u1 snapshot")
	}
	// Verify the snapshot captured the superposition state
	inSup, ok := uSnap["inSuperposition"]
	if !ok {
		t.Fatal("expected inSuperposition key in snapshot")
	}
	if inSup != true {
		t.Fatalf("expected inSuperposition=true, got %v", inSup)
	}
}

// ===================== K. Positioning (8 tests) =====================

func TestPosition_ToFinal_Static(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
		"DONE":   newFinalReality("DONE"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.PositionMachine(context.Background(), nil, "u1", "DONE", false); err != nil {
		t.Fatalf("PositionMachine failed: %v", err)
	}

	assertFinalized(t, u)
	assertReality(t, u, "DONE")
}

func TestPosition_WithAlways_Chain(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withAlways([]string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.PositionMachine(context.Background(), nil, "u1", "stateA", true); err != nil {
		t.Fatalf("PositionMachine (flow) failed: %v", err)
	}
	assertReality(t, u, "stateC")
}

func TestPosition_WithEntryActions(t *testing.T) {
	actionName := "test:sys:pos-entry"
	ran := false
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		ran = true
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.PositionMachine(context.Background(), nil, "u1", "stateA", true); err != nil {
		t.Fatalf("PositionMachine (flow) failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if !ran {
		t.Fatal("expected entry action to run with executeFlow=true")
	}
}

func TestPosition_Static_NoActions(t *testing.T) {
	actionName := "test:sys:pos-static-no"
	ran := false
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		ran = true
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.PositionMachine(context.Background(), nil, "u1", "stateA", false); err != nil {
		t.Fatalf("PositionMachine (static) failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if ran {
		t.Fatal("expected entry action NOT to run with executeFlow=false")
	}
}

func TestPosition_Flow_CrossUniverse(t *testing.T) {
	u1Realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"U:u2:stateX"}, nil),
		),
	}
	u2Realities := map[string]*theoretical.RealityModel{
		"stateX": newTransitionReality("stateX"),
	}

	qm, u1, u2 := buildMultiUniverseQM(t, "stateA", u1Realities, "stateX", u2Realities)
	if err := qm.PositionMachine(context.Background(), nil, "u1", "stateA", true); err != nil {
		t.Fatalf("PositionMachine (flow, cross-universe) failed: %v", err)
	}

	assertSuperposition(t, u1)
	// u2 should have been initialized via external target
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized after cross-universe positioning")
	}
}

func TestPosition_OnInitial_Flow(t *testing.T) {
	actionName := "test:sys:pos-initial-flow"
	ran := false
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		ran = true
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.PositionMachineOnInitial(context.Background(), nil, "u1", true); err != nil {
		t.Fatalf("PositionMachineOnInitial failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if !ran {
		t.Fatal("expected entry action to run with PositionMachineOnInitial flow=true")
	}
}

func TestPosition_ByCanonicalName_Flow(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.PositionMachineByCanonicalName(context.Background(), nil, "TestUniverse", "stateA", true); err != nil {
		t.Fatalf("PositionMachineByCanonicalName failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

func TestPosition_OnInitialByCanonical_Flow(t *testing.T) {
	actionName := "test:sys:pos-canon-initial"
	ran := false
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		ran = true
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.PositionMachineOnInitialByCanonicalName(context.Background(), nil, "TestUniverse", true); err != nil {
		t.Fatalf("PositionMachineOnInitialByCanonicalName failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if !ran {
		t.Fatal("expected entry action to run")
	}
}

// ===================== L. Event Dispatch (10 tests) =====================

func TestDispatch_ActiveUniverse(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !handled {
		t.Fatal("expected handled=true")
	}
	assertReality(t, u, "stateB")
}

func TestDispatch_SuperpositionUniverse_Rejected(t *testing.T) {
	// Build universe that transitions to superposition
	u1Model := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Universe1", Initial: strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"U:u2"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX",
				withObserver("builtin:observer:alwaysTrue", nil),
			),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, _ := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	exQM := qm.(*ExQuantumMachine)
	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Transition u1 to superposition
	evt := NewEventBuilder("go").Build()
	if _, err := exQM.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent(go) failed: %v", err)
	}
	assertSuperposition(t, u1)

	// Now send an event - u1 can't handle (superposition), u2 in superposition too
	evt2 := NewEventBuilder("something").Build()
	handled, err := exQM.SendEvent(context.Background(), evt2)
	if err != nil {
		t.Fatalf("SendEvent(something) failed: %v", err)
	}
	if handled {
		t.Fatal("expected handled=false for superposition universe")
	}
}

func TestDispatch_FinalizedUniverse_Rejected(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"DONE"}, nil),
		),
		"DONE": newFinalReality("DONE"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertFinalized(t, u)

	evt := NewEventBuilder("anything").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if handled {
		t.Fatal("expected handled=false for finalized universe")
	}
}

func TestDispatch_NoMatching_FalseNil(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Send event that has no handler
	evt := NewEventBuilder("nonexistent").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if handled {
		t.Fatal("expected handled=false for unmatched event from getLazy")
	}
}

func TestDispatch_MatchesMultiple(t *testing.T) {
	// Two universes, both active, both handle "go"
	u1Realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	u2Realities := map[string]*theoretical.RealityModel{
		"stateX": newTransitionReality("stateX",
			withOnTransition("go", []string{"stateY"}, nil),
		),
		"stateY": newTransitionReality("stateY"),
	}
	qm, u1, u2 := buildMultiUniverseQM(t, "stateA", u1Realities, "stateX", u2Realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !handled {
		t.Fatal("expected handled=true")
	}
	assertReality(t, u1, "stateB")
	assertReality(t, u2, "stateY")
}

func TestDispatch_DataInConditions(t *testing.T) {
	condName := "test:sys:dispatch-cond-data"
	registerTestCondition(t, condName, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return args.GetEvent().GetData()["allow"] == true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, &theoretical.ConditionModel{Src: condName}),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// First: data with allow=false -> no transition
	evt1 := NewEventBuilder("go").SetData(map[string]any{"allow": false}).Build()
	handled1, err := qm.SendEvent(context.Background(), evt1)
	if err != nil {
		t.Fatalf("SendEvent(1) failed: %v", err)
	}
	// The event matches the handler but condition fails -> still "handled" (has handler) but no transition
	_ = handled1
	assertReality(t, u, "stateA")

	// Second: data with allow=true -> transition
	evt2 := NewEventBuilder("go").SetData(map[string]any{"allow": true}).Build()
	if _, err := qm.SendEvent(context.Background(), evt2); err != nil {
		t.Fatalf("SendEvent(2) failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

func TestDispatch_DataInActions(t *testing.T) {
	actionName := "test:sys:dispatch-action-data"
	var captured string
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		if v, ok := args.GetEvent().GetData()["msg"]; ok {
			captured = v.(string)
		}
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			func(r *theoretical.RealityModel) {
				r.On["go"] = []*theoretical.TransitionModel{{
					Targets: []string{"stateB"},
					Actions: []*theoretical.ActionModel{{Src: actionName}},
				}}
			},
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").SetData(map[string]any{"msg": "hello"}).Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
	if captured != "hello" {
		t.Fatalf("expected captured=hello, got %q", captured)
	}
}

func TestDispatch_ConcurrentSendEvents(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("ping", []string{"stateB"}, nil),
			withOnTransition("pong", []string{"stateA"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withOnTransition("ping", []string{"stateA"}, nil),
			withOnTransition("pong", []string{"stateB"}, nil),
		),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				evtName := "ping"
				if j%2 == 0 {
					evtName = "pong"
				}
				evt := NewEventBuilder(evtName).Build()
				_, _ = qm.SendEvent(context.Background(), evt)
			}
		}(i)
	}
	wg.Wait()
	// If we get here without a race condition, the test passes
}

func TestDispatch_BeforeInit(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	// Do NOT call Init

	evt := NewEventBuilder("go").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent before init failed: %v", err)
	}
	if handled {
		t.Fatal("expected handled=false before init")
	}
}

func TestDispatch_EmptyEventName(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent('') failed: %v", err)
	}
	if handled {
		t.Fatal("expected handled=false for empty event name")
	}
}

// ===================== M. Chaos/Errors (9 tests) =====================

func TestChaos_ConditionError(t *testing.T) {
	condName := "test:sys:chaos-cond-err"
	registerTestCondition(t, condName, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, fmt.Errorf("intentional chaos condition error")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, &theoretical.ConditionModel{Src: condName}),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	_, err := qm.SendEvent(context.Background(), evt)
	if err == nil {
		t.Fatal("expected error from condition")
	}
	if !strings.Contains(err.Error(), "intentional chaos condition error") {
		t.Fatalf("expected condition error in message, got: %v", err)
	}
}

func TestChaos_EntryActionError_Init(t *testing.T) {
	actionName := "test:sys:chaos-entry-err"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("intentional entry action error")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	err := qm.Init(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from entry action during Init")
	}
	if !strings.Contains(err.Error(), "intentional entry action error") {
		t.Fatalf("expected entry action error, got: %v", err)
	}
}

func TestChaos_ExitActionError(t *testing.T) {
	actionName := "test:sys:chaos-exit-err"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("intentional exit action error")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withExitAction(actionName),
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	_, err := qm.SendEvent(context.Background(), evt)
	if err == nil {
		t.Fatal("expected error from exit action")
	}
	if !strings.Contains(err.Error(), "intentional exit action error") {
		t.Fatalf("expected exit action error, got: %v", err)
	}
}

func TestChaos_TransitionActionError(t *testing.T) {
	actionName := "test:sys:chaos-trans-err"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("intentional transition action error")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			func(r *theoretical.RealityModel) {
				r.On["go"] = []*theoretical.TransitionModel{{
					Targets: []string{"stateB"},
					Actions: []*theoretical.ActionModel{{Src: actionName}},
				}}
			},
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	_, err := qm.SendEvent(context.Background(), evt)
	if err == nil {
		t.Fatal("expected error from transition action")
	}
	if !strings.Contains(err.Error(), "intentional transition action error") {
		t.Fatalf("expected transition action error, got: %v", err)
	}
}

func TestChaos_SelfReferenceLoop(t *testing.T) {
	// A→B→A→B should bounce without crash, not infinite (bounded by manual sends)
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withOnTransition("go", []string{"stateA"}, nil),
		),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		evt := NewEventBuilder("go").Build()
		handled, err := qm.SendEvent(context.Background(), evt)
		if err != nil {
			t.Fatalf("SendEvent bounce %d failed: %v", i, err)
		}
		if !handled {
			t.Fatalf("expected handled=true on bounce %d", i)
		}
	}
	// After 20 bounces (even number), should be back at stateA
	assertReality(t, u, "stateA")
}

func TestChaos_InvalidReference(t *testing.T) {
	_, _, err := processReference("U::bad")
	if err == nil {
		t.Fatal("expected error for invalid reference 'U::bad'")
	}
	if !strings.Contains(err.Error(), "invalid ref") {
		t.Fatalf("expected 'invalid ref' in error, got: %v", err)
	}
}

func TestChaos_ObserverError(t *testing.T) {
	obsName := "test:sys:chaos-obs-err"
	registerTestObserver(t, obsName, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		return false, fmt.Errorf("intentional observer error")
	})

	// u1 is a router that forwards events to u2 via notify transition
	// u2 has no initial (enters superposition) with an observer that errors
	notifyType := theoretical.TransitionTypeNotify
	u1Model := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Universe1", Initial: strPtr("router"),
		Realities: map[string]*theoretical.RealityModel{
			"router": newTransitionReality("router",
				func(r *theoretical.RealityModel) {
					r.On["trigger"] = []*theoretical.TransitionModel{{
						Targets: []string{"U:u2:stateA"},
						Type:    &notifyType,
					}}
				},
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Universe2",
		// No Initial -> starts in superposition
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withObserver(obsName, nil),
			),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, _ := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	exQM := qm.(*ExQuantumMachine)

	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// u1 sends "trigger" via notify to u2:stateA. u2 is not initialized -> enters superposition.
	// The event gets accumulated for stateA -> observer runs -> returns error.
	evt := NewEventBuilder("trigger").Build()
	_, err := exQM.SendEvent(context.Background(), evt)
	if err == nil {
		t.Fatal("expected error from observer")
	}
	if !strings.Contains(err.Error(), "intentional observer error") {
		t.Fatalf("expected observer error in message, got: %v", err)
	}
}

func TestChaos_ConcurrentSendEvents(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("tick", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withOnTransition("tick", []string{"stateA"}, nil),
		),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				evt := NewEventBuilder("tick").Build()
				_, _ = qm.SendEvent(context.Background(), evt)
			}
		}()
	}
	wg.Wait()
}

func TestChaos_NilEventData(t *testing.T) {
	condName := "test:sys:chaos-nil-data"
	registerTestCondition(t, condName, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		data := args.GetEvent().GetData()
		// Access nil map key safely
		_ = data["anything"]
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, &theoretical.ConditionModel{Src: condName}),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Send event with nil data map — the EventBuilder uses empty map by default,
	// but we directly build with nil
	evt := &Event{Name: "go", Data: nil, EvtType: instrumentation.EventTypeOn}
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent with nil data failed: %v", err)
	}
	if !handled {
		t.Fatal("expected handled=true")
	}
	assertReality(t, u, "stateB")
}

// ===================== N. Complex Integration (8 tests) =====================

func TestInteg_OrderProcessing_FullFlow(t *testing.T) {
	// Order flow: PENDING → VALIDATING → APPROVED → PROCESSING → FULFILLING → SHIPPED → COMPLETED
	var trace []string

	traceAction := func(name string) string {
		actionName := fmt.Sprintf("test:sys:order-%s", name)
		registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
			trace = append(trace, args.GetRealityName())
			return nil
		})
		return actionName
	}

	validateAction := "test:sys:order-validate"
	registerTestAction(t, validateAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		trace = append(trace, "VALIDATING")
		args.EmitEvent("validation-passed", map[string]any{"valid": true})
		return nil
	})

	condValid := "test:sys:order-valid-cond"
	registerTestCondition(t, condValid, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return args.GetEvent().GetData()["valid"] == true, nil
	})

	approveAction := "test:sys:order-approve"
	registerTestAction(t, approveAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		trace = append(trace, "APPROVED")
		return nil
	})

	processAction := traceAction("process")
	fulfillAction := traceAction("fulfill")
	shipAction := traceAction("ship")

	realities := map[string]*theoretical.RealityModel{
		"PENDING": newTransitionReality("PENDING",
			withOnTransition("submit", []string{"VALIDATING"}, nil),
		),
		"VALIDATING": newTransitionReality("VALIDATING",
			withEntryAction(validateAction),
			withOnTransition("validation-passed", []string{"APPROVED"}, &theoretical.ConditionModel{Src: condValid}),
		),
		"APPROVED": newTransitionReality("APPROVED",
			withEntryAction(approveAction),
			withAlways([]string{"PROCESSING"}, nil),
		),
		"PROCESSING": newTransitionReality("PROCESSING",
			withEntryAction(processAction),
			withOnTransition("process-done", []string{"FULFILLING"}, nil),
		),
		"FULFILLING": newTransitionReality("FULFILLING",
			withEntryAction(fulfillAction),
			withOnTransition("fulfill-done", []string{"SHIPPED"}, nil),
		),
		"SHIPPED": newTransitionReality("SHIPPED",
			withEntryAction(shipAction),
			withAlways([]string{"COMPLETED"}, nil),
		),
		"COMPLETED": newFinalReality("COMPLETED"),
	}

	qm, u := buildQM(t, "PENDING", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "PENDING")

	// Submit order
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("submit").Build()); err != nil {
		t.Fatalf("SendEvent(submit) failed: %v", err)
	}
	// VALIDATING entry emits validation-passed → APPROVED → always → PROCESSING
	assertReality(t, u, "PROCESSING")

	// Process done
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("process-done").Build()); err != nil {
		t.Fatalf("SendEvent(process-done) failed: %v", err)
	}
	assertReality(t, u, "FULFILLING")

	// Fulfill done
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("fulfill-done").Build()); err != nil {
		t.Fatalf("SendEvent(fulfill-done) failed: %v", err)
	}
	// SHIPPED → always → COMPLETED
	assertReality(t, u, "COMPLETED")
	assertFinalized(t, u)

	expectedTrace := []string{"VALIDATING", "APPROVED", "PROCESSING", "FULFILLING", "SHIPPED"}
	if len(trace) != len(expectedTrace) {
		t.Fatalf("trace length: got %d, want %d. trace=%v", len(trace), len(expectedTrace), trace)
	}
	for i, exp := range expectedTrace {
		if trace[i] != exp {
			t.Fatalf("trace[%d]: got %q, want %q", i, trace[i], exp)
		}
	}
}

func TestInteg_MultiUniverseWorkflow(t *testing.T) {
	// auth universe: LOGIN → AUTHENTICATED
	// workflow universe: IDLE → ACTIVE → DONE (final)
	// notification universe: WAITING → NOTIFIED

	authAction := "test:sys:integ-auth-entry"
	registerTestAction(t, authAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		return nil
	})

	u1Realities := map[string]*theoretical.RealityModel{
		"LOGIN": newTransitionReality("LOGIN",
			withOnTransition("authenticate", []string{"AUTHENTICATED"}, nil),
		),
		"AUTHENTICATED": newTransitionReality("AUTHENTICATED"),
	}

	u2Realities := map[string]*theoretical.RealityModel{
		"IDLE": newTransitionReality("IDLE",
			withOnTransition("start-work", []string{"ACTIVE"}, nil),
		),
		"ACTIVE": newTransitionReality("ACTIVE",
			withOnTransition("complete", []string{"DONE"}, nil),
		),
		"DONE": newFinalReality("DONE"),
	}

	u3Realities := map[string]*theoretical.RealityModel{
		"WAITING": newTransitionReality("WAITING",
			withOnTransition("notify", []string{"NOTIFIED"}, nil),
		),
		"NOTIFIED": newTransitionReality("NOTIFIED"),
	}

	qm, u1, u2, u3 := buildThreeUniverseQMFromStrings(t,
		"LOGIN", u1Realities,
		"IDLE", u2Realities,
		"WAITING", u3Realities,
		[]string{"U:u1", "U:u2", "U:u3"},
	)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	assertReality(t, u1, "LOGIN")
	assertReality(t, u2, "IDLE")
	assertReality(t, u3, "WAITING")

	// Authenticate
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("authenticate").Build()); err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}
	assertReality(t, u1, "AUTHENTICATED")

	// Start work
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("start-work").Build()); err != nil {
		t.Fatalf("start-work failed: %v", err)
	}
	assertReality(t, u2, "ACTIVE")

	// Complete work
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("complete").Build()); err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	assertReality(t, u2, "DONE")
	assertFinalized(t, u2)

	// Notify
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("notify").Build()); err != nil {
		t.Fatalf("notify failed: %v", err)
	}
	assertReality(t, u3, "NOTIFIED")

	snap := qm.GetSnapshot()
	if _, ok := snap.Resume.ActiveUniverses["Universe1"]; !ok {
		t.Fatal("expected Universe1 in ActiveUniverses")
	}
	if _, ok := snap.Resume.FinalizedUniverses["Universe2"]; !ok {
		t.Fatal("expected Universe2 in FinalizedUniverses")
	}
	if _, ok := snap.Resume.ActiveUniverses["Universe3"]; !ok {
		t.Fatal("expected Universe3 in ActiveUniverses")
	}
}

func TestInteg_ObserverGatedWorkflow(t *testing.T) {
	// u1 is a router that forwards events to u2 via notify transitions.
	// u2 has an initial with an observer that uses ContainsAllEvents.
	// Events are accumulated until the observer approves, then the reality is established.
	// Note: after observer approval, the universe remains in superposition state
	// with currentReality set (this is the library's designed behavior for observer-gated realities).

	notifyType := theoretical.TransitionTypeNotify
	entryAction := "test:sys:integ-obs-gate-entry"
	var entryRan bool
	registerTestAction(t, entryAction, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		entryRan = true
		return nil
	})

	u1Model := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Universe1", Initial: strPtr("router"),
		Realities: map[string]*theoretical.RealityModel{
			"router": newTransitionReality("router",
				func(r *theoretical.RealityModel) {
					r.On["event-a"] = []*theoretical.TransitionModel{{
						Targets: []string{"U:u2"},
						Type:    &notifyType,
					}}
					r.On["event-b"] = []*theoretical.TransitionModel{{
						Targets: []string{"U:u2"},
						Type:    &notifyType,
					}}
					r.On["event-c"] = []*theoretical.TransitionModel{{
						Targets: []string{"U:u2"},
						Type:    &notifyType,
					}}
				},
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Universe2",
		Initial: strPtr("GATE"),
		Realities: map[string]*theoretical.RealityModel{
			"GATE": newTransitionReality("GATE",
				withEntryAction(entryAction),
				withObserver("builtin:observer:containsAllEvents", map[string]any{
					"evt1": "event-a",
					"evt2": "event-b",
					"evt3": "event-c",
				}),
			),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, _ := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	exQM := qm.(*ExQuantumMachine)

	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Send first event: u1 router notifies U:u2 -> u2 enters superposition, accumulates event-a
	if _, err := exQM.SendEvent(context.Background(), NewEventBuilder("event-a").Build()); err != nil {
		t.Fatalf("SendEvent(event-a) failed: %v", err)
	}
	if !u2.inSuperposition {
		t.Fatal("expected u2 in superposition after first event")
	}
	if entryRan {
		t.Fatal("entry should not have run yet - observer not yet approved")
	}

	// Send second event
	if _, err := exQM.SendEvent(context.Background(), NewEventBuilder("event-b").Build()); err != nil {
		t.Fatalf("SendEvent(event-b) failed: %v", err)
	}
	if !u2.inSuperposition {
		t.Fatal("expected u2 in superposition after second event")
	}
	if entryRan {
		t.Fatal("entry should not have run yet - observer not yet approved")
	}

	// Send third event -> observer approves -> reality is established
	if _, err := exQM.SendEvent(context.Background(), NewEventBuilder("event-c").Build()); err != nil {
		t.Fatalf("SendEvent(event-c) failed: %v", err)
	}

	// After observer approval, the entry actions execute and currentReality is set.
	// The fix for the observer-superposition bug ensures establishNewReality correctly
	// clears superposition — u2 should now be active, NOT in superposition.
	if !entryRan {
		t.Fatal("expected entry action to have run after observer approval")
	}
	if u2.currentReality == nil || *u2.currentReality != "GATE" {
		t.Fatal("expected currentReality=GATE after observer approval")
	}
	if u2.inSuperposition {
		t.Fatal("expected u2 NOT in superposition after observer collapse (bug fix verified)")
	}

	// Verify snapshot shows u2 as active (not in superposition)
	snap := exQM.GetSnapshot()
	if _, ok := snap.Resume.ActiveUniverses["Universe2"]; !ok {
		t.Fatal("expected Universe2 in ActiveUniverses after observer approval")
	}
}

func TestInteg_EmitEventChainCrossUniverse(t *testing.T) {
	// u1: stateA entry emits "advance" → transitions to U:u2:stateX
	// u2: stateX just exists
	emitAction := "test:sys:integ-emit-cross"
	registerTestAction(t, emitAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})

	u1Model := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Universe1", Initial: strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withEntryAction(emitAction),
				withOnTransition("advance", []string{"U:u2:stateX"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, _ := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	exQM := qm.(*ExQuantumMachine)

	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// u1 should be in superposition (external target)
	assertSuperposition(t, u1)
	// u2 should be established on stateX
	assertReality(t, u2, "stateX")
}

func TestInteg_SnapshotRestoreContinue(t *testing.T) {
	// Run half: PENDING → VALIDATING → PROCESSING
	// Snapshot. Restore. Continue: PROCESSING → FULFILLING → COMPLETED

	validateAction := "test:sys:integ-snap-validate"
	registerTestAction(t, validateAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("validated", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"PENDING": newTransitionReality("PENDING",
			withOnTransition("submit", []string{"VALIDATING"}, nil),
		),
		"VALIDATING": newTransitionReality("VALIDATING",
			withEntryAction(validateAction),
			withOnTransition("validated", []string{"PROCESSING"}, nil),
		),
		"PROCESSING": newTransitionReality("PROCESSING",
			withOnTransition("done", []string{"FULFILLING"}, nil),
		),
		"FULFILLING": newTransitionReality("FULFILLING",
			withAlways([]string{"COMPLETED"}, nil),
		),
		"COMPLETED": newFinalReality("COMPLETED"),
	}

	// First machine: run to PROCESSING
	qm1, u1 := buildQM(t, "PENDING", realities)
	if err := qm1.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm1.SendEvent(context.Background(), NewEventBuilder("submit").Build()); err != nil {
		t.Fatalf("SendEvent(submit) failed: %v", err)
	}
	assertReality(t, u1, "PROCESSING")

	// Snapshot
	snap := qm1.GetSnapshot()

	// Second machine: restore and continue
	qm2, u2 := buildQM(t, "PENDING", realities)
	if err := qm2.LoadSnapshot(snap, nil); err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}
	assertReality(t, u2, "PROCESSING")

	// Continue
	if _, err := qm2.SendEvent(context.Background(), NewEventBuilder("done").Build()); err != nil {
		t.Fatalf("SendEvent(done) failed: %v", err)
	}
	assertReality(t, u2, "COMPLETED")
	assertFinalized(t, u2)
}

func TestInteg_MixedFeatures(t *testing.T) {
	// Machine with observers + conditions + always + emit + constants in one flow
	emitAction := "test:sys:integ-mixed-emit"
	registerTestAction(t, emitAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("check", map[string]any{"proceed": true})
		return nil
	})

	condProceed := "test:sys:integ-mixed-cond"
	registerTestCondition(t, condProceed, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return args.GetEvent().GetData()["proceed"] == true, nil
	})

	constantsAction := "test:sys:integ-mixed-constants"
	var constantsCalled int
	registerTestAction(t, constantsAction, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		constantsCalled++
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"START": newTransitionReality("START",
			withEntryAction(emitAction),
			withOnTransition("check", []string{"MIDDLE"}, &theoretical.ConditionModel{Src: condProceed}),
		),
		"MIDDLE": newTransitionReality("MIDDLE",
			withAlways([]string{"END"}, nil),
		),
		"END": newFinalReality("END"),
	}

	um := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "TestUniverse", Initial: strPtr("START"),
		Realities: realities,
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm1", CanonicalName: "TestQM", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
		Initials:  []string{"U:u1"},
		UniversalConstants: &theoretical.UniversalConstantsModel{
			EntryActions: []*theoretical.ActionModel{{Src: constantsAction}},
		},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	exQM := qm.(*ExQuantumMachine)

	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// START emits "check" → condition passes → MIDDLE → always → END
	assertReality(t, u, "END")
	assertFinalized(t, u)

	// Constants entry action should have been called for each reality entry
	if constantsCalled == 0 {
		t.Fatal("expected constants entry action to be called at least once")
	}
}

func TestInteg_ReplayOnEntry(t *testing.T) {
	actionName := "test:sys:integ-replay"
	callCount := 0
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		callCount++
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)

	// Position statically (no entry action execution)
	if err := qm.PositionMachine(context.Background(), nil, "u1", "stateA", false); err != nil {
		t.Fatalf("PositionMachine failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if callCount != 0 {
		t.Fatalf("expected 0 calls after static position, got %d", callCount)
	}

	// ReplayOnEntry → entry action should run
	if err := qm.ReplayOnEntry(context.Background()); err != nil {
		t.Fatalf("ReplayOnEntry failed: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 call after replay, got %d", callCount)
	}

	// Replay again
	if err := qm.ReplayOnEntry(context.Background()); err != nil {
		t.Fatalf("ReplayOnEntry(2) failed: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls after second replay, got %d", callCount)
	}
}

func TestInteg_MetadataPersists(t *testing.T) {
	setAction := "test:sys:integ-md-set"
	registerTestAction(t, setAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.AddToUniverseMetadata("orderID", "ORD-123")
		args.AddToUniverseMetadata("step", "A")
		return nil
	})

	readActionB := "test:sys:integ-md-read-b"
	var mdAtB map[string]any
	registerTestAction(t, readActionB, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		mdAtB = make(map[string]any)
		for k, v := range args.GetUniverseMetadata() {
			mdAtB[k] = v
		}
		args.AddToUniverseMetadata("step", "B")
		return nil
	})

	readActionC := "test:sys:integ-md-read-c"
	var mdAtC map[string]any
	registerTestAction(t, readActionC, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		mdAtC = make(map[string]any)
		for k, v := range args.GetUniverseMetadata() {
			mdAtC[k] = v
		}
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(setAction),
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withEntryAction(readActionB),
			withOnTransition("go", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC",
			withEntryAction(readActionC),
		),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent(A->B) failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent(B->C) failed: %v", err)
	}

	// Metadata at B should have orderID from A
	if mdAtB["orderID"] != "ORD-123" {
		t.Fatalf("expected orderID=ORD-123 at B, got %v", mdAtB["orderID"])
	}
	if mdAtB["step"] != "A" {
		t.Fatalf("expected step=A at B, got %v", mdAtB["step"])
	}

	// Metadata at C should have step=B from B's update
	if mdAtC["orderID"] != "ORD-123" {
		t.Fatalf("expected orderID=ORD-123 at C, got %v", mdAtC["orderID"])
	}
	if mdAtC["step"] != "B" {
		t.Fatalf("expected step=B at C, got %v", mdAtC["step"])
	}
}
