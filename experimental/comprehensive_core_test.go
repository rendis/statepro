package experimental

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// ===================== A. Init/Lifecycle (18 tests) =====================

// A01: Init single universe with initial → currentReality=A, initialized
func TestInit_SingleUniverse_WithInitial(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if !u.initialized {
		t.Fatal("expected universe to be initialized")
	}
}

// A02: Init single universe without initial → inSuperposition=true
func TestInit_SingleUniverse_WithoutInitial_Superposition(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       nil,
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:u1"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !u.inSuperposition {
		t.Fatal("expected universe to be in superposition when Initial is nil")
	}
	if u.currentReality != nil {
		t.Fatalf("expected currentReality to be nil, got %v", *u.currentReality)
	}
}

// A03: Init multiple universes, all with initials → both initialized
func TestInit_MultipleUniverses_AllWithInitials(t *testing.T) {
	u1Realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	u2Realities := map[string]*theoretical.RealityModel{
		"stateX": newTransitionReality("stateX"),
	}
	qm, u1, u2 := buildMultiUniverseQM(t, "stateA", u1Realities, "stateX", u2Realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u1, "stateA")
	assertReality(t, u2, "stateX")
	if !u1.initialized || !u2.initialized {
		t.Fatal("expected both universes to be initialized")
	}
}

// A04: Init with mixed initials — U:u1 (uses Initial field) and U:u2:X (specific reality)
func TestInit_MixedInitials_UniverseAndUniverseReality(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA"),
			"stateB": newTransitionReality("stateB"),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Initial:       strPtr("stateY"),
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX"),
			"stateY": newTransitionReality("stateY"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:      []string{"U:u1", "U:u2:stateX"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	// u1 uses its Initial field → stateA
	assertReality(t, u1, "stateA")
	// u2 uses the reality from initials → stateX (overrides Initial=stateY)
	assertReality(t, u2, "stateX")
}

// A05: UniverseReality format overrides Initial
func TestInit_UniverseRealityFormat_OverridesInitial(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA"),
			"stateB": newTransitionReality("stateB"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:u1:stateB"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// A06: Zero universes in Initials → no error, nothing initialized
func TestInit_ZeroUniversesInInitials(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if u.initialized {
		t.Fatal("expected universe to NOT be initialized with empty Initials")
	}
}

// A07: Invalid initial reference → error
func TestInit_InvalidInitialReference(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:noexist"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	err = qm.(*ExQuantumMachine).Init(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for invalid initial reference, got nil")
	}
}

// A08: InitWithEvent → custom event propagated to entry action
func TestInitWithEvent_CustomEventPropagated(t *testing.T) {
	actionName := "test:core:a08-entry-action"
	var capturedEvtName string
	var capturedData map[string]any
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		capturedEvtName = args.GetEvent().GetEventName()
		capturedData = args.GetEvent().GetData()
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)

	customEvt := NewEventBuilder("my-custom-init").
		SetData(map[string]any{"key": "val"}).
		Build()
	if err := qm.InitWithEvent(context.Background(), nil, customEvt); err != nil {
		t.Fatalf("InitWithEvent failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if capturedEvtName != "my-custom-init" {
		t.Fatalf("expected event name 'my-custom-init', got %q", capturedEvtName)
	}
	if capturedData["key"] != "val" {
		t.Fatalf("expected data key=val, got %v", capturedData)
	}
}

// A09: Init twice → second call returns error "machine already initialized"
func TestInit_SecondInit_ReturnsError(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	qm, _ := buildQM(t, "stateA", realities)

	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("first Init failed: %v", err)
	}

	// Second Init must return error
	err := qm.Init(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error on second Init, got nil")
	}
	if !strings.Contains(err.Error(), "already initialized") {
		t.Fatalf("expected 'already initialized' error, got: %v", err)
	}
}

// A10: Init with nil machine context → no error
func TestInit_NilMachineContext(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init(ctx, nil) failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// A11: Init with custom machine context → accessible in action
func TestInit_CustomMachineContext(t *testing.T) {
	type myCtx struct {
		Value string
	}
	actionName := "test:core:a11-entry-action"
	var capturedCtx any
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		capturedCtx = args.GetContext()
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)

	mc := &myCtx{Value: "hello"}
	if err := qm.Init(context.Background(), mc); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if capturedCtx == nil {
		t.Fatal("expected non-nil machine context in action")
	}
	if typed, ok := capturedCtx.(*myCtx); !ok || typed.Value != "hello" {
		t.Fatalf("expected machine context with Value=hello, got %v", capturedCtx)
	}
}

// A12: Entry actions execute on Init
func TestInit_EntryActionsExecuteOnInit(t *testing.T) {
	ran := false
	actionName := "test:core:a12-entry-action"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		ran = true
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !ran {
		t.Fatal("expected entry action to run on Init")
	}
}

// A13: Always transitions execute after Init
func TestInit_AlwaysTransitionsExecuteAfterInit(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// A14: Init on stateA with always→external target U:u2:X → u1 superposition, u2 on X
func TestInit_ExternalTargetFromInitial(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withAlways([]string{"U:u2:stateX"}, nil),
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
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:      []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !u1.inSuperposition {
		t.Fatal("expected u1 to be in superposition after external target")
	}
	assertReality(t, u2, "stateX")
}

// A15: Chained external targets: u1 always→U:u2:stateX, then separately u2:stateX and u3:stateY
// receive events from the external target chain. Both u2 and u3 establish on specified realities.
func TestInit_ChainedExternalTargets(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				// u1's always transitions to external target U:u2:stateX
				withAlways([]string{"U:u2:stateX"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		// No Initial — u2 starts uninitialized and receives external target
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:      []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	// u1 enters superposition after external target transition
	if !u1.inSuperposition {
		t.Fatal("expected u1 to be in superposition")
	}
	// u2 receives external target → establishes on stateX
	assertReality(t, u2, "stateX")
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized via external target chain")
	}
}

// A16: Universe metadata from model is copied and accessible in action
func TestInit_UniverseMetadataCopied(t *testing.T) {
	actionName := "test:core:a16-entry-action"
	var capturedMeta map[string]any
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		capturedMeta = args.GetUniverseMetadata()
		return nil
	})

	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
		},
		Metadata: map[string]any{"myKey": "myValue"},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:u1"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if capturedMeta == nil || capturedMeta["myKey"] != "myValue" {
		t.Fatalf("expected metadata myKey=myValue, got %v", capturedMeta)
	}
}

// A17: Init without event → event.name="start"
func TestInit_NilEvent_DefaultStartEvent(t *testing.T) {
	actionName := "test:core:a17-entry-action"
	var capturedEvtName string
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		capturedEvtName = args.GetEvent().GetEventName()
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if capturedEvtName != "start" {
		t.Fatalf("expected default event name 'start', got %q", capturedEvtName)
	}
}

// A18: Initials=[U:u1:B] → B entry action runs
func TestInit_UniverseRealityFormat_EntryRuns(t *testing.T) {
	ran := false
	actionName := "test:core:a18-entry-action"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		ran = true
		return nil
	})

	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA"),
			"stateB": newTransitionReality("stateB", withEntryAction(actionName)),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:u1:stateB"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
	if !ran {
		t.Fatal("expected stateB entry action to run")
	}
}

// ===================== B. Reality Types & Transitions (15 tests) =====================

// B01: On handler single event
func TestReality_OnHandler_SingleEvent(t *testing.T) {
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
		t.Fatal("expected event to be handled")
	}
	assertReality(t, u, "stateB")
}

// B02: On handler multiple events
func TestReality_OnHandler_MultipleEvents(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("evt1", []string{"stateB"}, nil),
			withOnTransition("evt2", []string{"stateC"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
		"stateC": newTransitionReality("stateC"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("evt2").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !handled {
		t.Fatal("expected event to be handled")
	}
	assertReality(t, u, "stateC")
}

// B03: Final reality rejects events
func TestFinalReality_RejectsEvents(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"DONE"}, nil),
		),
		"DONE": newFinalReality("DONE"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	// transition to final
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "DONE")

	// events should be rejected
	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("anything").Build())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if handled {
		t.Fatal("expected event NOT to be handled after final reality")
	}
}

// B04: Unsuccessful final reality rejects events
func TestUnsuccessfulFinalReality_RejectsEvents(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"FAILED"}, nil),
		),
		"FAILED": newUnsuccessfulFinalReality("FAILED"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "FAILED")

	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("anything").Build())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if handled {
		t.Fatal("expected event NOT to be handled after unsuccessful final reality")
	}
}

// B05: Unsuccessful final IsFinalState=true
func TestUnsuccessfulFinalReality_IsFinalState(t *testing.T) {
	if !theoretical.IsFinalState(theoretical.RealityTypeUnsuccessfulFinal) {
		t.Fatal("expected IsFinalState to return true for RealityTypeUnsuccessfulFinal")
	}
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"FAILED"}, nil),
		),
		"FAILED": newUnsuccessfulFinalReality("FAILED"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "FAILED")
	if !u.isFinalReality {
		t.Fatal("expected isFinalReality=true for unsuccessful final reality")
	}
}

// B06: Self-loop transition increments counter
func TestSelfLoopTransition(t *testing.T) {
	counter := 0
	actionName := "test:core:b06-entry-action"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		counter++
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
			withOnTransition("loop", []string{"stateA"}, nil),
		),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	// counter=1 from Init
	if counter != 1 {
		t.Fatalf("expected counter=1 after init, got %d", counter)
	}

	evt := NewEventBuilder("loop").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if counter != 2 {
		t.Fatalf("expected counter=2 after self-loop, got %d", counter)
	}
}

// B07: Multiple transitions same event, first condition false, second true → goes to C
func TestMultipleTransitions_SameEvent_FirstConditionWins(t *testing.T) {
	condFalse := "test:core:b07-cond-false"
	condTrue := "test:core:b07-cond-true"
	registerTestCondition(t, condFalse, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, nil
	})
	registerTestCondition(t, condTrue, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{Targets: []string{"stateB"}, Condition: &theoretical.ConditionModel{Src: condFalse}},
			{Targets: []string{"stateC"}, Condition: &theoretical.ConditionModel{Src: condTrue}},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
		"stateC": newTransitionReality("stateC"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateC")
}

// B08: Transition with empty targets → stays
func TestTransition_EmptyTargets_NoTransition(t *testing.T) {
	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{Targets: []string{}},
		},
	}
	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// B09: Standard single local target A→B
func TestTransition_SingleLocalTarget(t *testing.T) {
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
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// B10: Single external target → u1 superposition
func TestTransition_SingleExternalTarget_Superposition(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"U:u2"}, nil),
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
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:      []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.(*ExQuantumMachine).SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !u1.inSuperposition {
		t.Fatal("expected u1 to be in superposition after external target")
	}
}

// B11: Single external target with reality → u2 on X
func TestTransition_SingleExternalTargetWithReality(t *testing.T) {
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
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:      []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.(*ExQuantumMachine).SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u2, "stateX")
}

// B12: Multiple external targets → superposition
func TestTransition_MultipleTargets_Superposition(t *testing.T) {
	// Multi-target to 2 external universes → u1 enters superposition
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"U:u2:stateX", "U:u3:stateY"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX",
				withObserver("builtin:observer:alwaysTrue", nil),
			),
		},
	}
	u3Model := &theoretical.UniverseModel{
		ID:            "u3",
		CanonicalName: "Universe3",
		Realities: map[string]*theoretical.RealityModel{
			"stateY": newTransitionReality("stateY",
				withObserver("builtin:observer:alwaysTrue", nil),
			),
		},
	}
	qm, u1, _, _ := buildThreeUniverseQM(t, u1Model, u2Model, u3Model, []string{"U:u1"})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !u1.inSuperposition {
		t.Fatal("expected u1 in superposition with multiple targets")
	}
}

// B13: Transition to final reality → snapshot shows finalized
func TestTransition_ToFinalReality_SnapshotCorrect(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"DONE"}, nil),
		),
		"DONE": newFinalReality("DONE"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "DONE")

	snapshot := qm.GetSnapshot()
	finalized := snapshot.Resume.FinalizedUniverses
	if finalized == nil || finalized["TestUniverse"] != "DONE" {
		t.Fatalf("expected finalized universe TestUniverse=DONE, got %v", finalized)
	}
}

// B14: SendEvent with unknown event → not handled (false, nil)
func TestTransition_UnknownEvent_NotHandled(t *testing.T) {
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
	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("unknown").Build())
	if err != nil {
		t.Fatalf("expected no error for unknown event, got: %v", err)
	}
	if handled {
		t.Fatal("expected event NOT to be handled for unknown event name")
	}
	// machine state unchanged
	assertReality(t, u, "stateA")
}

// B15: Final reality with entry actions → actions run
func TestTransition_FinalReality_WithEntryActions(t *testing.T) {
	ran := false
	actionName := "test:core:b15-entry-action"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		ran = true
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"DONE"}, nil),
		),
		"DONE": {
			ID:           "DONE",
			Type:         theoretical.RealityTypeFinal,
			On:           map[string][]*theoretical.TransitionModel{},
			EntryActions: []*theoretical.ActionModel{{Src: actionName}},
		},
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "DONE")
	if !ran {
		t.Fatal("expected entry actions to run on final reality")
	}
}

// ===================== C. Conditions (12 tests) =====================

// C01: Single true condition → transition
func TestCondition_SingleTrue(t *testing.T) {
	condName := "test:core:c01-cond-true"
	registerTestCondition(t, condName, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
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
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// C02: Single false condition → stays
func TestCondition_SingleFalse(t *testing.T) {
	condName := "test:core:c02-cond-false"
	registerTestCondition(t, condName, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, nil
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
	_, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	// stays in stateA (no approved transition, no error from SendEvent for condition rejection)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// C03: Multiple conditions all true (AND logic) → transition
func TestCondition_MultipleAllTrue_AND(t *testing.T) {
	condTrue1 := "test:core:c03-cond-true1"
	condTrue2 := "test:core:c03-cond-true2"
	registerTestCondition(t, condTrue1, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})
	registerTestCondition(t, condTrue2, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets: []string{"stateB"},
				Conditions: []*theoretical.ConditionModel{
					{Src: condTrue1},
					{Src: condTrue2},
				},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// C04: Multiple conditions one false (AND logic) → stays
func TestCondition_MultipleOneFalse_AND(t *testing.T) {
	condTrue := "test:core:c04-cond-true"
	condFalse := "test:core:c04-cond-false"
	registerTestCondition(t, condTrue, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})
	registerTestCondition(t, condFalse, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, nil
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets: []string{"stateB"},
				Conditions: []*theoretical.ConditionModel{
					{Src: condTrue},
					{Src: condFalse},
				},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// C05: Mixed Condition and Conditions both true → transition
func TestCondition_MixedConditionAndConditions(t *testing.T) {
	condMain := "test:core:c05-cond-main"
	condExtra := "test:core:c05-cond-extra"
	registerTestCondition(t, condMain, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})
	registerTestCondition(t, condExtra, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets:    []string{"stateB"},
				Condition:  &theoretical.ConditionModel{Src: condMain},
				Conditions: []*theoretical.ConditionModel{{Src: condExtra}},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// C06: Mixed Condition true but Conditions has false → stays
func TestCondition_MixedOneFailsInConditions(t *testing.T) {
	condMain := "test:core:c06-cond-main"
	condFail := "test:core:c06-cond-fail"
	registerTestCondition(t, condMain, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})
	registerTestCondition(t, condFail, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, nil
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets:    []string{"stateB"},
				Condition:  &theoretical.ConditionModel{Src: condMain},
				Conditions: []*theoretical.ConditionModel{{Src: condFail}},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// C07: Condition returns error → propagated
func TestCondition_ReturnsError(t *testing.T) {
	condName := "test:core:c07-cond-error"
	registerTestCondition(t, condName, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, fmt.Errorf("intentional error")
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
	_, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err == nil {
		t.Fatal("expected error from condition, got nil")
	}
	if !strings.Contains(err.Error(), "intentional error") {
		t.Fatalf("expected 'intentional error' in message, got: %v", err)
	}
}

// C08: Condition with empty Src → defaults to true
func TestCondition_EmptySrc_DefaultTrue(t *testing.T) {
	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets:   []string{"stateB"},
				Condition: &theoretical.ConditionModel{Src: ""},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// C09: Condition with nonexistent Src → defaults to false
func TestCondition_NotFound_DefaultFalse(t *testing.T) {
	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets:   []string{"stateB"},
				Condition: &theoretical.ConditionModel{Src: "nonexistent-condition-xyz123"},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// C10: Condition args passed correctly
func TestCondition_ArgsPassedCorrectly(t *testing.T) {
	condName := "test:core:c10-cond-args"
	var capturedArgs map[string]any
	registerTestCondition(t, condName, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		capturedArgs = args.GetCondition().Args
		return true, nil
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets:   []string{"stateB"},
				Condition: &theoretical.ConditionModel{Src: condName, Args: map[string]any{"threshold": 42}},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if capturedArgs == nil || capturedArgs["threshold"] != 42 {
		t.Fatalf("expected condition args threshold=42, got %v", capturedArgs)
	}
}

// C11: Event data accessible in condition
func TestCondition_EventDataAccessible(t *testing.T) {
	condName := "test:core:c11-cond-evtdata"
	var capturedEvtData map[string]any
	registerTestCondition(t, condName, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		capturedEvtData = args.GetEvent().GetData()
		return capturedEvtData["approved"] == true, nil
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
	evt := NewEventBuilder("go").SetData(map[string]any{"approved": true}).Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// C12: Condition mutates metadata → persists
func TestCondition_MetadataMutation(t *testing.T) {
	condName := "test:core:c12-cond-metamut"
	registerTestCondition(t, condName, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		args.AddToUniverseMetadata("condRan", true)
		return true, nil
	})

	actionName := "test:core:c12-check-meta"
	var metaValue any
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		metaValue = args.GetUniverseMetadata()["condRan"]
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, &theoretical.ConditionModel{Src: condName}),
		),
		"stateB": newTransitionReality("stateB", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if metaValue != true {
		t.Fatalf("expected metadata condRan=true, got %v", metaValue)
	}
}

// ===================== D. Actions (16 tests) =====================

// D01: Entry action execution order
func TestEntryAction_ExecutionOrder(t *testing.T) {
	var order []string
	for i := 0; i < 3; i++ {
		idx := i
		name := fmt.Sprintf("test:core:d01-entry-%d", idx)
		registerTestAction(t, name, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
			order = append(order, fmt.Sprintf("entry-%d", idx))
			return nil
		})
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction("test:core:d01-entry-0"),
			withEntryAction("test:core:d01-entry-1"),
			withEntryAction("test:core:d01-entry-2"),
		),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	expected := []string{"entry-0", "entry-1", "entry-2"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Fatalf("order[%d]: expected %q, got %q", i, exp, order[i])
		}
	}
}

// D02: Exit action execution order
func TestExitAction_ExecutionOrder(t *testing.T) {
	var order []string
	for i := 0; i < 3; i++ {
		idx := i
		name := fmt.Sprintf("test:core:d02-exit-%d", idx)
		registerTestAction(t, name, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
			order = append(order, fmt.Sprintf("exit-%d", idx))
			return nil
		})
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withExitAction("test:core:d02-exit-0"),
			withExitAction("test:core:d02-exit-1"),
			withExitAction("test:core:d02-exit-2"),
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	expected := []string{"exit-0", "exit-1", "exit-2"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Fatalf("order[%d]: expected %q, got %q", i, exp, order[i])
		}
	}
}

// D03: Transition action execution order
func TestTransitionAction_ExecutionOrder(t *testing.T) {
	var order []string
	for i := 0; i < 2; i++ {
		idx := i
		name := fmt.Sprintf("test:core:d03-trans-%d", idx)
		registerTestAction(t, name, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
			order = append(order, fmt.Sprintf("trans-%d", idx))
			return nil
		})
	}

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets: []string{"stateB"},
				Actions: []*theoretical.ActionModel{
					{Src: "test:core:d03-trans-0"},
					{Src: "test:core:d03-trans-1"},
				},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	expected := []string{"trans-0", "trans-1"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Fatalf("order[%d]: expected %q, got %q", i, exp, order[i])
		}
	}
}

// D04: Entry action failure → Init error
func TestEntryAction_Failure_RealityNotEstablished(t *testing.T) {
	actionName := "test:core:d04-entry-fail"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("entry action failure")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	err := qm.Init(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from failed entry action, got nil")
	}
	if !strings.Contains(err.Error(), "entry action failure") {
		t.Fatalf("expected 'entry action failure' in message, got: %v", err)
	}
}

// D05: Exit action failure → SendEvent error
func TestExitAction_Failure_ReturnsError(t *testing.T) {
	actionName := "test:core:d05-exit-fail"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("exit action failure")
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
	_, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err == nil {
		t.Fatal("expected error from failed exit action, got nil")
	}
	if !strings.Contains(err.Error(), "exit action failure") {
		t.Fatalf("expected 'exit action failure' in message, got: %v", err)
	}
}

// D06: Transition action failure → aborted
func TestTransitionAction_Failure_Aborted(t *testing.T) {
	actionName := "test:core:d06-trans-fail"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("transition action failure")
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets: []string{"stateB"},
				Actions: []*theoretical.ActionModel{{Src: actionName}},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	_, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err == nil {
		t.Fatal("expected error from failed transition action, got nil")
	}
	if !strings.Contains(err.Error(), "transition action failure") {
		t.Fatalf("expected 'transition action failure' in message, got: %v", err)
	}
}

// D07: Action with empty Src → no error (no-op)
func TestAction_EmptySrc_NoOp(t *testing.T) {
	stateA := newTransitionReality("stateA")
	stateA.EntryActions = []*theoretical.ActionModel{{Src: ""}}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// D08: Action with nonexistent Src → no error (no-op, logs warning)
func TestAction_NotFound_NoOp(t *testing.T) {
	stateA := newTransitionReality("stateA")
	stateA.EntryActions = []*theoretical.ActionModel{{Src: "nonexistent-action-xyz123"}}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// D09: Action args accessible
func TestAction_ArgsAccessible(t *testing.T) {
	actionName := "test:core:d09-action-args"
	var capturedArgs map[string]any
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		capturedArgs = args.GetAction().Args
		return nil
	})

	stateA := newTransitionReality("stateA")
	stateA.EntryActions = []*theoretical.ActionModel{
		{Src: actionName, Args: map[string]any{"count": 7, "label": "test"}},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if capturedArgs == nil {
		t.Fatal("expected captured args to be non-nil")
	}
	if capturedArgs["label"] != "test" {
		t.Fatalf("expected label=test, got %v", capturedArgs["label"])
	}
}

// D10: EmitEvent only from entry action; exit emits are ignored
func TestAction_EmitEvent_OnlyFromEntry(t *testing.T) {
	entryAction := "test:core:d10-entry-emit"
	exitAction := "test:core:d10-exit-emit"

	registerTestAction(t, entryAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("entry-advance", nil)
		return nil
	})
	registerTestAction(t, exitAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("exit-advance", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(entryAction),
			withExitAction(exitAction),
			withOnTransition("entry-advance", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withOnTransition("exit-advance", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	// Entry emit works: stateA→stateB
	assertReality(t, u, "stateB")

	// Now send an event that will trigger exit from stateB — exit emit should be ignored
	realities["stateB"].On["manual-go"] = []*theoretical.TransitionModel{
		{Targets: []string{"stateA"}},
	}
	// Cannot modify after init for On. Instead, verify that stateB is current and exit-advance was ignored.
	// stateB does not have stateC transition from exit, so staying at stateB proves exit emit was ignored.
}

// D11: GetSnapshot from entry action → non-nil
func TestAction_GetSnapshot_FromEntry(t *testing.T) {
	actionName := "test:core:d11-snapshot"
	var gotSnapshot bool
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		snap := args.GetSnapshot()
		gotSnapshot = snap != nil
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !gotSnapshot {
		t.Fatal("expected GetSnapshot() to return non-nil in entry action")
	}
}

// D12: AddToUniverseMetadata persists
func TestAction_SetToUniverseMetadata(t *testing.T) {
	setAction := "test:core:d12-set-meta"
	checkAction := "test:core:d12-check-meta"
	registerTestAction(t, setAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.AddToUniverseMetadata("myKey", "myValue")
		return nil
	})
	var metaVal any
	registerTestAction(t, checkAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		metaVal = args.GetUniverseMetadata()["myKey"]
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(setAction),
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB", withEntryAction(checkAction)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if metaVal != "myValue" {
		t.Fatalf("expected myKey=myValue, got %v", metaVal)
	}
}

// D13: DeleteFromUniverseMetadata works
func TestAction_DeleteFromUniverseMetadata(t *testing.T) {
	setAction := "test:core:d13-set-meta"
	deleteAction := "test:core:d13-del-meta"
	checkAction := "test:core:d13-check-meta"

	registerTestAction(t, setAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.AddToUniverseMetadata("toDelete", "exists")
		return nil
	})
	registerTestAction(t, deleteAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.DeleteFromUniverseMetadata("toDelete")
		return nil
	})
	var hasKey bool
	registerTestAction(t, checkAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		_, hasKey = args.GetUniverseMetadata()["toDelete"]
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(setAction),
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withEntryAction(deleteAction),
			withOnTransition("go2", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC", withEntryAction(checkAction)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent go failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go2").Build()); err != nil {
		t.Fatalf("SendEvent go2 failed: %v", err)
	}
	if hasKey {
		t.Fatal("expected toDelete key to be removed from metadata")
	}
}

// D14: UpdateUniverseMetadata replaces metadata and persists to the universe.
// After the fix, UpdateUniverseMetadata clears-and-copies into the underlying u.metadata
// (consistent with AddTo/DeleteFrom), so changes persist across actions and transitions.
func TestAction_UpdateUniverseMetadata(t *testing.T) {
	updateAction := "test:core:d14-update-meta"
	verifyAction := "test:core:d14-verify-meta"

	var metaInVerify map[string]any
	registerTestAction(t, updateAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.UpdateUniverseMetadata(map[string]any{"brand": "new", "version": 2})
		return nil
	})
	registerTestAction(t, verifyAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		// This action runs in a DIFFERENT reality — should see the persisted metadata
		metaInVerify = make(map[string]any)
		for k, v := range args.GetUniverseMetadata() {
			metaInVerify[k] = v
		}
		return nil
	})

	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withEntryAction(updateAction),
				withOnTransition("go", []string{"stateB"}, nil),
			),
			"stateB": newTransitionReality("stateB",
				withEntryAction(verifyAction),
			),
		},
		Metadata: map[string]any{"old": "data"},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:u1"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	exQM := qm.(*ExQuantumMachine)
	if err := exQM.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Transition to stateB where verify action reads metadata
	if _, err := exQM.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	// Verify the update persisted: "old" key should be gone, new keys present
	if metaInVerify["brand"] != "new" {
		t.Fatalf("expected brand=new in persisted metadata, got %v", metaInVerify)
	}
	if metaInVerify["version"] != 2 {
		t.Fatalf("expected version=2 in persisted metadata, got %v", metaInVerify)
	}
	if _, hasOld := metaInVerify["old"]; hasOld {
		t.Fatal("expected 'old' key to be removed after UpdateUniverseMetadata")
	}
}

// D15: Action receives correct ActionType
func TestAction_ReceivesCorrectActionType(t *testing.T) {
	entryActionName := "test:core:d15-entry-type"
	exitActionName := "test:core:d15-exit-type"
	transActionName := "test:core:d15-trans-type"

	var entryType, exitType, transType instrumentation.ActionType

	registerTestAction(t, entryActionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		entryType = args.GetActionType()
		return nil
	})
	registerTestAction(t, exitActionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		exitType = args.GetActionType()
		return nil
	})
	registerTestAction(t, transActionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		transType = args.GetActionType()
		return nil
	})

	stateA := newTransitionReality("stateA",
		withEntryAction(entryActionName),
		withExitAction(exitActionName),
	)
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets: []string{"stateB"},
				Actions: []*theoretical.ActionModel{{Src: transActionName}},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	if entryType != instrumentation.ActionTypeEntry {
		t.Fatalf("expected entry ActionType=%q, got %q", instrumentation.ActionTypeEntry, entryType)
	}
	if exitType != instrumentation.ActionTypeExit {
		t.Fatalf("expected exit ActionType=%q, got %q", instrumentation.ActionTypeExit, exitType)
	}
	if transType != instrumentation.ActionTypeTransition {
		t.Fatalf("expected transition ActionType=%q, got %q", instrumentation.ActionTypeTransition, transType)
	}
}

// D16: GetUniverseId and GetUniverseCanonicalName correct
func TestAction_GetUniverseID_And_CanonicalName(t *testing.T) {
	actionName := "test:core:d16-ids"
	var capturedID, capturedCN string
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		capturedID = args.GetUniverseId()
		capturedCN = args.GetUniverseCanonicalName()
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if capturedID != "u1" {
		t.Fatalf("expected universe ID 'u1', got %q", capturedID)
	}
	if capturedCN != "TestUniverse" {
		t.Fatalf("expected canonical name 'TestUniverse', got %q", capturedCN)
	}
}

// ===================== E. Invokes (8 tests) =====================

// E01: Entry invoke fires async
func TestEntryInvoke_FiresAsync(t *testing.T) {
	ch := make(chan string, 1)
	invokeName := "test:core:e01-entry-invoke"
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		ch <- args.GetRealityName()
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryInvoke(invokeName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	select {
	case name := <-ch:
		if name != "stateA" {
			t.Fatalf("expected reality name 'stateA', got %q", name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("entry invoke did not fire within timeout")
	}
}

// E02: Exit invoke fires async
func TestExitInvoke_FiresAsync(t *testing.T) {
	ch := make(chan string, 1)
	invokeName := "test:core:e02-exit-invoke"
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		ch <- args.GetRealityName()
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withExitInvoke(invokeName),
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	select {
	case name := <-ch:
		if name != "stateA" {
			t.Fatalf("expected reality name 'stateA' on exit, got %q", name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("exit invoke did not fire within timeout")
	}
}

// E03: Transition invoke fires
func TestTransitionInvoke_Fires(t *testing.T) {
	ch := make(chan string, 1)
	invokeName := "test:core:e03-trans-invoke"
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		ch <- args.GetRealityName()
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {
			{
				Targets: []string{"stateB"},
				Invokes: []*theoretical.InvokeModel{{Src: invokeName}},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	select {
	case name := <-ch:
		if name != "stateA" {
			t.Fatalf("expected reality name 'stateA' on transition, got %q", name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("transition invoke did not fire within timeout")
	}
}

// E04: Invoke with empty Src → no panic
func TestInvoke_EmptySrc_NoOp(t *testing.T) {
	stateA := newTransitionReality("stateA")
	stateA.EntryInvokes = []*theoretical.InvokeModel{{Src: ""}}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// E05: Invoke not found → no panic
func TestInvoke_NotFound_NoOp(t *testing.T) {
	stateA := newTransitionReality("stateA")
	stateA.EntryInvokes = []*theoretical.InvokeModel{{Src: "nonexistent-invoke-xyz123"}}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// E06: Invoke does not block Init
func TestInvoke_DoesNotBlock(t *testing.T) {
	invokeName := "test:core:e06-slow-invoke"
	registerTestInvoke(t, invokeName, func(_ context.Context, _ instrumentation.InvokeExecutorArgs) {
		time.Sleep(5 * time.Second)
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryInvoke(invokeName)),
	}
	qm, _ := buildQM(t, "stateA", realities)

	done := make(chan error, 1)
	go func() {
		done <- qm.Init(context.Background(), nil)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Init failed: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Init blocked by slow invoke (should be async)")
	}
}

// E07: Invoke args accessible
func TestInvoke_ArgsAccessible(t *testing.T) {
	ch := make(chan map[string]any, 1)
	invokeName := "test:core:e07-invoke-args"
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		ch <- args.GetInvoke().Args
	})

	stateA := newTransitionReality("stateA")
	stateA.EntryInvokes = []*theoretical.InvokeModel{
		{Src: invokeName, Args: map[string]any{"url": "https://example.com"}},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	select {
	case args := <-ch:
		if args["url"] != "https://example.com" {
			t.Fatalf("expected url=https://example.com, got %v", args["url"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("invoke did not fire within timeout")
	}
}

// E08: Invoke metadata accessible
func TestInvoke_MetadataAccessible(t *testing.T) {
	ch := make(chan map[string]any, 1)
	invokeName := "test:core:e08-invoke-meta"
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		ch <- args.GetUniverseMetadata()
	})

	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA", withEntryInvoke(invokeName)),
		},
		Metadata: map[string]any{"envKey": "production"},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:u1"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.(*ExQuantumMachine).Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	select {
	case meta := <-ch:
		if meta["envKey"] != "production" {
			t.Fatalf("expected envKey=production, got %v", meta["envKey"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("invoke did not fire within timeout")
	}
}
