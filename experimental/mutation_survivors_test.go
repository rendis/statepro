package experimental

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// Tests aimed at meaningful Gremlins LIVED survivors (boundaries / observable branches).

func TestMutation_ProcessReference_InvalidReturnsSentinel(t *testing.T) {
	rt, parts, err := processReference("!!!not-a-ref")
	if err == nil {
		t.Fatal("expected error")
	}
	if rt != -1 {
		t.Fatalf("invalid ref sentinel type=%d, want -1 (kills invert-negatives / arithmetic on -1)", rt)
	}
	if parts != nil {
		t.Fatalf("parts=%v, want nil", parts)
	}
}

func TestMutation_PositionMachine_ExecuteFlowWithExternalTargets(t *testing.T) {
	notifyType := theoretical.TransitionTypeNotify
	u1Model := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Router", Initial: strPtr("idle"),
		Realities: map[string]*theoretical.RealityModel{
			"idle": {
				ID: "idle", Type: theoretical.RealityTypeTransition,
				Always: []*theoretical.TransitionModel{{
					Type:    &notifyType,
					Targets: []string{"U:u2:ready"},
				}},
			},
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Target", Initial: strPtr("ready"),
		Realities: map[string]*theoretical.RealityModel{
			"ready": newTransitionReality("ready"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-mut-pos", CanonicalName: "Q", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{},
	}
	u1, u2 := NewExUniverse(u1Model), NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	// No Init: PositionMachine with executeFlow must still fan out always→notify
	// external targets (kills skipping the len(externalTargets) > 0 branch).
	if err := qm.PositionMachine(context.Background(), nil, "u1", "idle", true); err != nil {
		t.Fatalf("PositionMachine flow: %v", err)
	}
	assertReality(t, u1, "idle")
	assertReality(t, u2, "ready")
}

func TestMutation_MachineConstantEntryAction_Failure(t *testing.T) {
	actionName := "test:mut-mconst-fail"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("machine constant boom")
	})

	um := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "U", Initial: strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-mut-mconst", CanonicalName: "Q", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
		Initials:  []string{"U:u1"},
		UniversalConstants: &theoretical.UniversalConstantsModel{
			EntryActions: []*theoretical.ActionModel{{Src: actionName}},
		},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err == nil {
		t.Fatal("machine constant entry failure must abort Init (kills err != nil negation)")
	}
}

func TestMutation_LoadSnapshot_EmptyMetadataKeepsNilMap(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	snap := qm.GetSnapshot()
	us := snap.Snapshots["u1"]
	delete(us, "metadata")
	snap.Snapshots["u1"] = us

	u.metadata = nil
	if err := qm.LoadSnapshot(snap, nil); err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if u.metadata != nil {
		t.Fatalf("empty metadata must not allocate map (kills && → ||): %#v", u.metadata)
	}
}

func TestMutation_StartOnReality_NilEventGetsDefault(t *testing.T) {
	var sawType instrumentation.EventType
	actionName := "test:mut-starton-evt"
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		sawType = args.GetEvent().GetEvtType()
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	// PositionMachine passes nil event into startOnReality.
	if err := qm.PositionMachine(context.Background(), nil, "u1", "stateA", true); err != nil {
		t.Fatalf("PositionMachine: %v", err)
	}
	if sawType != instrumentation.EventTypeStartOn {
		t.Fatalf("nil event must become StartOn, got %q", sawType)
	}
}

func TestMutation_EmitDepth_ExactlyAtLimitSucceeds(t *testing.T) {
	// maxEmitDepth=10 → a chain of 10 nested emits must succeed; 11 must fail.
	const depth = maxEmitDepth
	for i := 0; i < depth; i++ {
		idx := i
		name := fmt.Sprintf("test:mut-emit-depth-%d", idx)
		nextEvent := fmt.Sprintf("go-depth-%d", idx+1)
		registerTestAction(t, name, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
			args.EmitEvent(nextEvent, nil)
			return nil
		})
	}

	realities := map[string]*theoretical.RealityModel{}
	for i := 0; i <= depth; i++ {
		id := fmt.Sprintf("d-%d", i)
		if i < depth {
			realities[id] = newTransitionReality(id,
				withEntryAction(fmt.Sprintf("test:mut-emit-depth-%d", i)),
				withOnTransition(fmt.Sprintf("go-depth-%d", i+1), []string{fmt.Sprintf("d-%d", i+1)}, nil),
			)
		} else {
			realities[id] = newTransitionReality(id)
		}
	}

	qm, u := buildQM(t, "d-0", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init at exact maxEmitDepth should succeed: %v", err)
	}
	assertReality(t, u, fmt.Sprintf("d-%d", depth))
}

func TestMutation_ConditionError_UsesConditionsSrcWhenConditionNil(t *testing.T) {
	condName := "test:mut-cond-only-err"
	registerTestCondition(t, condName, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, fmt.Errorf("cond-only boom")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": {
			ID: "stateA", Type: theoretical.RealityTypeTransition,
			On: map[string][]*theoretical.TransitionModel{
				"go": {{
					Targets: []string{"stateB"},
					// Condition intentionally nil; only Conditions populated.
					Conditions: []*theoretical.ConditionModel{{Src: condName}},
				}},
			},
		},
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	_, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err == nil {
		t.Fatal("expected condition error")
	}
	if !strings.Contains(err.Error(), condName) {
		t.Fatalf("error should mention conditions src %q, got: %v", condName, err)
	}
	assertReality(t, u, "stateA")
}

func TestMutation_CanHandleEvent_RequiresInitializedConcrete(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	_, u := buildQM(t, "stateA", realities)
	evt := NewEventBuilder("go").Build()

	if u.canHandleEvent(evt) {
		t.Fatal("uninitialized universe must not handle events")
	}
	u.initialized = true
	u.inSuperposition = false
	u.isFinalReality = false
	u.currentReality = strPtr("stateA")
	if !u.canHandleEvent(evt) {
		t.Fatal("initialized concrete reality with On must handle event")
	}
	u.isFinalReality = true
	if u.canHandleEvent(evt) {
		t.Fatal("final reality must not handle events via canHandleEvent")
	}
}

func TestMutation_PopTrackingIfLast_OnlyPopsMatchingTail(t *testing.T) {
	u := &ExUniverse{tracking: []string{"a", "b", "c"}}
	u.popTrackingIfLast("x")
	if got := strings.Join(u.tracking, ","); got != "a,b,c" {
		t.Fatalf("non-tail pop changed tracking: %s", got)
	}
	u.popTrackingIfLast("c")
	if got := strings.Join(u.tracking, ","); got != "a,b" {
		t.Fatalf("tail pop failed: %s", got)
	}
	u.popTrackingIfLast("a")
	if got := strings.Join(u.tracking, ","); got != "a,b" {
		t.Fatalf("non-last matching id must not pop: %s", got)
	}
}
