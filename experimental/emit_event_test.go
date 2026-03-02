package experimental

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// ---------- helpers ----------

// newTransitionReality creates a transition-type reality with configurable On/Always/Entry.
func newTransitionReality(id string, opts ...func(*theoretical.RealityModel)) *theoretical.RealityModel {
	r := &theoretical.RealityModel{
		ID:   id,
		Type: theoretical.RealityTypeTransition,
		On:   map[string][]*theoretical.TransitionModel{},
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func newFinalReality(id string) *theoretical.RealityModel {
	return &theoretical.RealityModel{
		ID:   id,
		Type: theoretical.RealityTypeFinal,
		On:   map[string][]*theoretical.TransitionModel{},
	}
}

func withOnTransition(eventName string, targets []string, condition *theoretical.ConditionModel) func(*theoretical.RealityModel) {
	return func(r *theoretical.RealityModel) {
		t := &theoretical.TransitionModel{Targets: targets}
		if condition != nil {
			t.Condition = condition
		}
		r.On[eventName] = append(r.On[eventName], t)
	}
}

func withEntryAction(src string) func(*theoretical.RealityModel) {
	return func(r *theoretical.RealityModel) {
		r.EntryActions = append(r.EntryActions, &theoretical.ActionModel{Src: src})
	}
}

func withAlways(targets []string, condition *theoretical.ConditionModel) func(*theoretical.RealityModel) {
	return func(r *theoretical.RealityModel) {
		t := &theoretical.TransitionModel{Targets: targets}
		if condition != nil {
			t.Condition = condition
		}
		r.Always = append(r.Always, t)
	}
}

// buildQM builds a minimal quantum machine for testing with a single universe.
func buildQM(t *testing.T, initial string, realities map[string]*theoretical.RealityModel) (*ExQuantumMachine, *ExUniverse) {
	t.Helper()
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       &initial,
		Realities:     realities,
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
	return qm.(*ExQuantumMachine), u
}

// buildMultiUniverseQM builds a QM with two universes for cross-universe tests.
func buildMultiUniverseQM(
	t *testing.T,
	u1Initial string, u1Realities map[string]*theoretical.RealityModel,
	u2Initial string, u2Realities map[string]*theoretical.RealityModel,
) (*ExQuantumMachine, *ExUniverse, *ExUniverse) {
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
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"u1": um1,
			"u2": um2,
		},
		Initials: []string{"U:u1", "U:u2"},
	}
	u1 := NewExUniverse(um1)
	u2 := NewExUniverse(um2)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build multi-universe QM: %v", err)
	}
	return qm.(*ExQuantumMachine), u1, u2
}

// registerTestAction registers an action. Names must be unique across all tests.
func registerTestAction(t *testing.T, name string, fn instrumentation.ActionFn) {
	t.Helper()
	if err := builtin.RegisterAction(name, fn); err != nil {
		t.Fatalf("failed to register action %s: %v", name, err)
	}
}

// registerTestCondition registers a condition. Names must be unique across all tests.
func registerTestCondition(t *testing.T, name string, fn instrumentation.ConditionFn) {
	t.Helper()
	if err := builtin.RegisterCondition(name, fn); err != nil {
		t.Fatalf("failed to register condition %s: %v", name, err)
	}
}

func assertReality(t *testing.T, u *ExUniverse, expected string) {
	t.Helper()
	if u.currentReality == nil {
		t.Fatalf("expected reality '%s', got nil", expected)
	}
	if *u.currentReality != expected {
		t.Fatalf("expected reality '%s', got '%s'", expected, *u.currentReality)
	}
}

func strPtr(s string) *string {
	return &s
}

// ---------- Tests ----------

// 1. Entry action emits event → transition executes → reality changes
func TestEmitEvent_HappyPath(t *testing.T) {
	actionName := "test:emit-happy"
	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"CREATING": newTransitionReality("CREATING",
			withEntryAction(actionName),
			withOnTransition("advance", []string{"SIGNING"}, nil),
		),
		"SIGNING": newTransitionReality("SIGNING"),
	}

	qm, u := buildQM(t, "CREATING", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "SIGNING")
}

// 2. Entry action emits event with data → condition evaluates data correctly
func TestEmitEvent_WithDataAndCondition(t *testing.T) {
	actionName := "test:emit-data"
	conditionName := "test:check-template"

	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("create-contract", map[string]any{"templateId": "abc-123"})
		return nil
	})
	registerTestCondition(t, conditionName, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return args.GetEvent().GetData()["templateId"] == "abc-123", nil
	})

	realities := map[string]*theoretical.RealityModel{
		"CREATING": newTransitionReality("CREATING",
			withEntryAction(actionName),
			withOnTransition("create-contract", []string{"SIGNING"}, &theoretical.ConditionModel{Src: conditionName}),
		),
		"SIGNING": newTransitionReality("SIGNING"),
	}

	qm, u := buildQM(t, "CREATING", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "SIGNING")
}

// 3. Chained emits: stateA emits → transition to stateB → stateB emits → transition to stateC
func TestEmitEvent_ChainedEmits(t *testing.T) {
	action1 := "test:emit-chain-1"
	action2 := "test:emit-chain-2"

	registerTestAction(t, action1, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-to-b", nil)
		return nil
	})
	registerTestAction(t, action2, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-to-c", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(action1),
			withOnTransition("go-to-b", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withEntryAction(action2),
			withOnTransition("go-to-c", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateC")
}

// 4. Two actions emit same event → only first is processed
func TestEmitEvent_DuplicateSameEvent(t *testing.T) {
	callCount := 0
	action1 := "test:emit-dup-1"
	action2 := "test:emit-dup-2"

	registerTestAction(t, action1, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		callCount++
		args.EmitEvent("advance", nil)
		return nil
	})
	registerTestAction(t, action2, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		callCount++
		args.EmitEvent("advance", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(action1),
			withEntryAction(action2),
			withOnTransition("advance", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
	if callCount != 2 {
		t.Fatalf("expected both actions to execute, got %d", callCount)
	}
}

// 5. Two actions emit different events → FIFO, first transition wins
func TestEmitEvent_DifferentEvents_FirstWins(t *testing.T) {
	action1 := "test:emit-diff-1"
	action2 := "test:emit-diff-2"

	registerTestAction(t, action1, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-to-b", nil)
		return nil
	})
	registerTestAction(t, action2, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-to-c", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(action1),
			withEntryAction(action2),
			withOnTransition("go-to-b", []string{"stateB"}, nil),
			withOnTransition("go-to-c", []string{"stateC"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
		"stateC": newTransitionReality("stateC"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB") // first emit wins
}

// 6. First emit has no handler, second does → second is processed
func TestEmitEvent_FirstNoHandler_SecondProcessed(t *testing.T) {
	action1 := "test:emit-nohandler-1"
	action2 := "test:emit-nohandler-2"

	registerTestAction(t, action1, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("nonexistent-event", nil)
		return nil
	})
	registerTestAction(t, action2, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(action1),
			withEntryAction(action2),
			withOnTransition("advance", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// 7. First emit has condition that fails, second passes → second wins
func TestEmitEvent_FirstConditionFails_SecondWins(t *testing.T) {
	action1 := "test:emit-condfail-1"
	action2 := "test:emit-condfail-2"
	condFalse := "test:always-false"
	condTrue := "test:always-true"

	registerTestAction(t, action1, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", map[string]any{"path": "first"})
		return nil
	})
	registerTestAction(t, action2, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", map[string]any{"path": "second"})
		return nil
	})
	registerTestCondition(t, condFalse, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, nil
	})
	registerTestCondition(t, condTrue, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(action1),
			withEntryAction(action2),
			// first transition has condition that always fails
			withOnTransition("advance", []string{"stateB"}, &theoretical.ConditionModel{Src: condFalse}),
		),
		"stateB": newTransitionReality("stateB"),
	}
	// add second transition with always-true condition targeting stateC
	realities["stateA"].On["advance"] = append(realities["stateA"].On["advance"],
		&theoretical.TransitionModel{
			Targets:   []string{"stateC"},
			Condition: &theoretical.ConditionModel{Src: condTrue},
		},
	)
	realities["stateC"] = newTransitionReality("stateC")

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	// condition for stateB fails → condition for stateC passes → goes to stateC
	assertReality(t, u, "stateC")
}

// 8. EmitEvent without handler On → no error, reality doesn't change
func TestEmitEvent_NoHandler(t *testing.T) {
	actionName := "test:emit-nohandler"
	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("nonexistent", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
		),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// 9. EmitEvent with condition that fails → no transition
func TestEmitEvent_ConditionFails(t *testing.T) {
	actionName := "test:emit-condfails"
	condName := "test:emit-condfails-cond"

	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})
	registerTestCondition(t, condName, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
			withOnTransition("advance", []string{"stateB"}, &theoretical.ConditionModel{Src: condName}),
		),
		"stateB": newTransitionReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// 10. No action emits → zero new behavior (backward compatibility)
func TestEmitEvent_NoneEmitted(t *testing.T) {
	actionName := "test:no-emit"
	called := false
	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		called = true
		// deliberately does NOT call EmitEvent
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
			withOnTransition("advance", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
	if !called {
		t.Fatal("entry action was not called")
	}
}

// 11. Infinite emit loop → error with maxEmitDepth
func TestEmitEvent_InfiniteLoop(t *testing.T) {
	actionPing := "test:emit-ping"
	actionPong := "test:emit-pong"

	registerTestAction(t, actionPing, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-pong", nil)
		return nil
	})
	registerTestAction(t, actionPong, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-ping", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"PING": newTransitionReality("PING",
			withEntryAction(actionPing),
			withOnTransition("go-pong", []string{"PONG"}, nil),
		),
		"PONG": newTransitionReality("PONG",
			withEntryAction(actionPong),
			withOnTransition("go-ping", []string{"PING"}, nil),
		),
	}

	qm, _ := buildQM(t, "PING", realities)
	err := qm.Init(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for infinite emit loop")
	}
	if !strings.Contains(err.Error(), "depth exceeded") {
		t.Fatalf("expected 'depth exceeded' in error, got: %v", err)
	}
}

// 12. Chain within depth limit → works
func TestEmitEvent_LongChainWithinLimit(t *testing.T) {
	// Chain of 5 transitions via emitted events (well within maxEmitDepth=10)
	for i := 0; i < 5; i++ {
		idx := i
		name := fmt.Sprintf("test:emit-chain-step-%d", idx)
		nextEvent := fmt.Sprintf("go-step-%d", idx+1)
		registerTestAction(t, name, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
			args.EmitEvent(nextEvent, nil)
			return nil
		})
	}

	realities := map[string]*theoretical.RealityModel{}
	for i := 0; i <= 5; i++ {
		id := fmt.Sprintf("step-%d", i)
		if i < 5 {
			action := fmt.Sprintf("test:emit-chain-step-%d", i)
			event := fmt.Sprintf("go-step-%d", i+1)
			next := fmt.Sprintf("step-%d", i+1)
			realities[id] = newTransitionReality(id,
				withEntryAction(action),
				withOnTransition(event, []string{next}, nil),
			)
		} else {
			realities[id] = newTransitionReality(id)
		}
	}

	qm, u := buildQM(t, "step-0", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "step-5")
}

// 13. EmitEvent from exit action → no-op (warning logged)
func TestEmitEvent_FromExitAction_Noop(t *testing.T) {
	emitActionName := "test:emit-from-exit"
	emitCalled := false

	registerTestAction(t, emitActionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		emitCalled = true
		args.EmitEvent("should-be-ignored", nil)
		return nil
	})

	stateA := newTransitionReality("stateA",
		withOnTransition("go-b", []string{"stateB"}, nil),
	)
	stateA.ExitActions = []*theoretical.ActionModel{{Src: emitActionName}}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB",
			withOnTransition("should-be-ignored", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// send event to trigger exit from stateA
	evt := NewEventBuilder("go-b").Build()
	_, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	if !emitCalled {
		t.Fatal("exit action was not called")
	}
	// should be in stateB, not stateC (emit was ignored)
	assertReality(t, u, "stateB")
}

// 14. EmitEvent from transition action → no-op
func TestEmitEvent_FromTransitionAction_Noop(t *testing.T) {
	transActionName := "test:emit-from-transition"
	emitCalled := false

	registerTestAction(t, transActionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		emitCalled = true
		args.EmitEvent("should-be-ignored", nil)
		return nil
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go-b": {
			{
				Targets: []string{"stateB"},
				Actions: []*theoretical.ActionModel{{Src: transActionName}},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB",
			withOnTransition("should-be-ignored", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go-b").Build()
	_, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	if !emitCalled {
		t.Fatal("transition action was not called")
	}
	assertReality(t, u, "stateB")
}

// 15. Emit that targets external universe (U:other-universe) → superposition correct
func TestEmitEvent_ExternalUniverseTarget(t *testing.T) {
	actionName := "test:emit-external"
	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-external", nil)
		return nil
	})

	// Build manually — u2 should NOT be in Initials (it will be initialized via external target)
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withEntryAction(actionName),
				withOnTransition("go-external", []string{"U:u2:stateX"}, nil),
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
		Initials:      []string{"U:u1"}, // only u1
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}

	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// u1 should be in superposition (external target)
	if !u1.inSuperposition {
		t.Fatal("u1 should be in superposition after external target emit")
	}
	// u2 should be established on stateX (from external target processing)
	assertReality(t, u2, "stateX")
}

// 16. Emit that targets multiple targets → superposition
func TestEmitEvent_MultipleTargets_Superposition(t *testing.T) {
	actionName := "test:emit-multi-target"
	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-multi", nil)
		return nil
	})

	// Build manually — u2 should NOT be in Initials
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withEntryAction(actionName),
				withOnTransition("go-multi", []string{"U:u1:stateB", "U:u2:stateX"}, nil),
			),
			"stateB": newTransitionReality("stateB"),
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
		Initials:      []string{"U:u1"}, // only u1
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}

	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// u1 should be in superposition because of multiple targets
	if !u1.inSuperposition {
		t.Fatal("u1 should be in superposition with multiple targets")
	}
	_ = u2 // u2 used for reference only
}

// 18. Emit triggers transition + new reality has always → both execute
func TestEmitEvent_FollowedByAlways(t *testing.T) {
	actionName := "test:emit-then-always"
	condName := "test:always-true-for-always"

	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})
	registerTestCondition(t, condName, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
			withOnTransition("advance", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withAlways([]string{"stateC"}, &theoretical.ConditionModel{Src: condName}),
		),
		"stateC": newTransitionReality("stateC"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	// stateA → emit → stateB → always → stateC
	assertReality(t, u, "stateC")
}

// 19. Emit doesn't trigger transition → always of current reality executes normally
func TestEmitEvent_NoTransition_AlwaysStillRuns(t *testing.T) {
	actionName := "test:emit-no-trans"
	condAlways := "test:always-true-normal"

	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("nonexistent", nil)
		return nil
	})
	registerTestCondition(t, condAlways, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
			withAlways([]string{"stateB"}, &theoretical.ConditionModel{Src: condAlways}),
		),
		"stateB": newTransitionReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	// emit has no handler → always runs → goes to stateB
	assertReality(t, u, "stateB")
}

// 20. EmitEvent with empty eventName → skip silently
func TestEmitEvent_EmptyName(t *testing.T) {
	actionName := "test:emit-empty-name"
	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
		),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// 21. EmitEvent with nil data → works
func TestEmitEvent_NilData(t *testing.T) {
	actionName := "test:emit-nil-data"
	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
			withOnTransition("advance", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// 22. Final reality with emit → no On handlers, skip
func TestEmitEvent_FinalRealityEntry(t *testing.T) {
	actionName := "test:emit-final"
	condAlways := "test:always-true-final"

	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})
	registerTestCondition(t, condAlways, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"DONE"}, &theoretical.ConditionModel{Src: condAlways}),
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
	assertReality(t, u, "DONE")
}

// 23. Emit during ReplayOnEntry → processed correctly
func TestEmitEvent_ReplayOnEntry(t *testing.T) {
	actionName := "test:emit-replay"
	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
			withOnTransition("advance", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)

	// Position machine on stateA statically (no flow execution)
	if err := qm.PositionMachine(context.Background(), nil, "u1", "stateA", false); err != nil {
		t.Fatalf("PositionMachine failed: %v", err)
	}
	assertReality(t, u, "stateA")

	// ReplayOnEntry should execute entry actions → emit → transition
	if err := qm.ReplayOnEntry(context.Background()); err != nil {
		t.Fatalf("ReplayOnEntry failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// 24. Integration test: Full machine flow CREATING → EmitEvent → SIGNING → sign → SIGNED → always → final
func TestEmitEvent_Integration_FullFlow(t *testing.T) {
	sendContractAction := "test:send-contract"
	condTemplate := "test:if-template"

	registerTestAction(t, sendContractAction, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		templateId := args.GetAction().Args["templateId"]
		args.EmitEvent("create-contract", map[string]any{
			"templateId": templateId,
		})
		return nil
	})
	registerTestCondition(t, condTemplate, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return args.GetEvent().GetData()["templateId"] == "tmpl-001", nil
	})

	alwaysTrue := "test:integration-always-true"
	registerTestCondition(t, alwaysTrue, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"CREATING_CONTRACT": newTransitionReality("CREATING_CONTRACT",
			withEntryAction(sendContractAction),
			withOnTransition("create-contract", []string{"SIGNING_CONTRACT"}, &theoretical.ConditionModel{Src: condTemplate}),
		),
		"SIGNING_CONTRACT": newTransitionReality("SIGNING_CONTRACT",
			withOnTransition("sign", []string{"CONTRACT_SIGNED"}, nil),
		),
		"CONTRACT_SIGNED": newTransitionReality("CONTRACT_SIGNED",
			withAlways([]string{"COMPLETED"}, &theoretical.ConditionModel{Src: alwaysTrue}),
		),
		"COMPLETED": newFinalReality("COMPLETED"),
	}

	// Set the action args on the entry action
	realities["CREATING_CONTRACT"].EntryActions[0].Args = map[string]any{
		"templateId": "tmpl-001",
	}

	qm, u := buildQM(t, "CREATING_CONTRACT", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// After init: CREATING_CONTRACT → emit → SIGNING_CONTRACT
	assertReality(t, u, "SIGNING_CONTRACT")

	// Send external "sign" event
	evt := NewEventBuilder("sign").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !handled {
		t.Fatal("expected event to be handled")
	}

	// After sign: SIGNING_CONTRACT → sign → CONTRACT_SIGNED → always → COMPLETED
	assertReality(t, u, "COMPLETED")
	if !u.isFinalReality {
		t.Fatal("expected final reality")
	}
}

// Test: EmitEvent emitted type is EventTypeEmitted
func TestEmitEvent_EventType(t *testing.T) {
	actionName := "test:emit-type-check"
	condName := "test:check-event-type"

	registerTestAction(t, actionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("typed-event", map[string]any{"key": "val"})
		return nil
	})
	registerTestCondition(t, condName, func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return args.GetEvent().GetEvtType() == instrumentation.EventTypeEmitted, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(actionName),
			withOnTransition("typed-event", []string{"stateB"}, &theoretical.ConditionModel{Src: condName}),
		),
		"stateB": newTransitionReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// Test: Constants entry actions can also emit events
func TestEmitEvent_ConstantsEntryAction(t *testing.T) {
	constantsActionName := "test:constants-emit"
	registerTestAction(t, constantsActionName, func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("constants-advance", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("constants-advance", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}

	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities:     realities,
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:u1"},
		UniversalConstants: &theoretical.UniversalConstantsModel{
			EntryActions: []*theoretical.ActionModel{{Src: constantsActionName}},
		},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}

	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// ---------- Bug regression tests ----------

// Bug #1 regression: transition actions downstream of EmitEvent must receive the emitted event,
// not the original external event.
func TestEmitEvent_CorrectEventInTransitionActions(t *testing.T) {
	entryActionName := "test:emit-correct-evt-entry"
	transActionName := "test:emit-correct-evt-trans"

	registerTestAction(t, entryActionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("my-emitted", map[string]any{"emitKey": "emitVal"})
		return nil
	})

	var capturedEventName string
	var capturedEventType instrumentation.EventType
	var capturedData map[string]any
	registerTestAction(t, transActionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		capturedEventName = args.GetEvent().GetEventName()
		capturedEventType = args.GetEvent().GetEvtType()
		capturedData = args.GetEvent().GetData()
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(entryActionName),
			func(r *theoretical.RealityModel) {
				r.On["my-emitted"] = []*theoretical.TransitionModel{{
					Targets: []string{"stateB"},
					Actions: []*theoretical.ActionModel{{Src: transActionName}},
				}}
			},
		),
		"stateB": newFinalReality("stateB"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	assertReality(t, u, "stateB")

	if capturedEventName != "my-emitted" {
		t.Fatalf("transition action got event name %q, want %q", capturedEventName, "my-emitted")
	}
	if capturedEventType != instrumentation.EventTypeEmitted {
		t.Fatalf("transition action got event type %q, want %q", capturedEventType, instrumentation.EventTypeEmitted)
	}
	if capturedData["emitKey"] != "emitVal" {
		t.Fatalf("transition action got data %v, want emitKey=emitVal", capturedData)
	}
}

// Bug #1 regression: entry actions in the TARGET reality must receive the emitted event.
func TestEmitEvent_CorrectEventInNextEntryActions(t *testing.T) {
	emitterAction := "test:emit-correct-next-entry"
	receiverAction := "test:recv-correct-next-entry"

	registerTestAction(t, emitterAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("my-emitted-2", map[string]any{"payload": "hello"})
		return nil
	})

	var receivedEventName string
	var receivedEventType instrumentation.EventType
	var receivedData map[string]any
	registerTestAction(t, receiverAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		receivedEventName = args.GetEvent().GetEventName()
		receivedEventType = args.GetEvent().GetEvtType()
		receivedData = args.GetEvent().GetData()
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(emitterAction),
			withOnTransition("my-emitted-2", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withEntryAction(receiverAction),
		),
	}

	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if receivedEventName != "my-emitted-2" {
		t.Fatalf("next entry action got event name %q, want %q", receivedEventName, "my-emitted-2")
	}
	if receivedEventType != instrumentation.EventTypeEmitted {
		t.Fatalf("next entry action got event type %q, want %q", receivedEventType, instrumentation.EventTypeEmitted)
	}
	if receivedData["payload"] != "hello" {
		t.Fatalf("next entry action got data %v, want payload=hello", receivedData)
	}
}

// ---------- Intensive tests for doCyclicTransition / establishNewReality optimizations ----------

// Test: transition to a final reality correctly sets isFinalReality and rejects further events.
func TestDoCyclicTransition_TransitionToFinalReality(t *testing.T) {
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
	assertReality(t, u, "stateA")

	// Send "go" → should transition to DONE (final)
	evt := NewEventBuilder("go").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !handled {
		t.Fatal("expected event to be handled")
	}

	assertReality(t, u, "DONE")
	if !u.isFinalReality {
		t.Fatal("expected isFinalReality to be true after transitioning to final reality")
	}

	// Further events should NOT be handled (universe is finalized)
	handled2, err2 := qm.SendEvent(context.Background(), NewEventBuilder("another").Build())
	if err2 != nil {
		t.Fatalf("SendEvent after final should not error, got: %v", err2)
	}
	if handled2 {
		t.Fatal("expected event NOT to be handled after reaching final reality")
	}
}

// Test: chained always transitions ending in final — isFinalReality correct, tracking complete.
func TestDoCyclicTransition_ChainedAlwaysToFinal(t *testing.T) {
	var visited []string
	trackAction := "test:track-chain-always"
	registerTestAction(t, trackAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		visited = append(visited, args.GetRealityName())
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withEntryAction(trackAction),
			withAlways([]string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC",
			withEntryAction(trackAction),
			withAlways([]string{"FINAL"}, nil),
		),
		"FINAL": {
			ID:           "FINAL",
			Type:         theoretical.RealityTypeFinal,
			On:           map[string][]*theoretical.TransitionModel{},
			EntryActions: []*theoretical.ActionModel{{Src: trackAction}},
		},
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")

	evt := NewEventBuilder("go").Build()
	handled, err := qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !handled {
		t.Fatal("expected event to be handled")
	}

	assertReality(t, u, "FINAL")
	if !u.isFinalReality {
		t.Fatal("expected isFinalReality=true after chained always to final")
	}

	// Verify tracking through snapshot
	snapshot := qm.GetSnapshot()
	tracking := snapshot.Tracking["u1"]
	expected := []string{"stateA", "stateB", "stateC", "FINAL"}
	if len(tracking) != len(expected) {
		t.Fatalf("tracking length: got %d, want %d. tracking=%v", len(tracking), len(expected), tracking)
	}
	for i, exp := range expected {
		if tracking[i] != exp {
			t.Fatalf("tracking[%d]: got %q, want %q. full tracking=%v", i, tracking[i], exp, tracking)
		}
	}

	// Verify entry actions ran in order
	expectedVisited := []string{"stateB", "stateC", "FINAL"}
	if len(visited) != len(expectedVisited) {
		t.Fatalf("visited length: got %d, want %d. visited=%v", len(visited), len(expectedVisited), visited)
	}
	for i, exp := range expectedVisited {
		if visited[i] != exp {
			t.Fatalf("visited[%d]: got %q, want %q", i, visited[i], exp)
		}
	}
}

// Test: emit during entry → On transition → always → final. Triple interaction.
func TestDoCyclicTransition_EmitDuringEntryToFinalViaAlways(t *testing.T) {
	emitAction := "test:emit-to-final-always"
	registerTestAction(t, emitAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withEntryAction(emitAction),
			withOnTransition("advance", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC",
			withAlways([]string{"FINAL"}, nil),
		),
		"FINAL": newFinalReality("FINAL"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	assertReality(t, u, "FINAL")
	if !u.isFinalReality {
		t.Fatal("expected isFinalReality=true after emit → transition → always → final")
	}

	// Verify full tracking chain
	tracking := qm.GetSnapshot().Tracking["u1"]
	expected := []string{"stateA", "stateB", "stateC", "FINAL"}
	if len(tracking) != len(expected) {
		t.Fatalf("tracking: got %v, want %v", tracking, expected)
	}
	for i, exp := range expected {
		if tracking[i] != exp {
			t.Fatalf("tracking[%d]: got %q, want %q", i, tracking[i], exp)
		}
	}
}

// Test: always transition targeting another universe → triggers superposition via doCyclicTransition.
func TestDoCyclicTransition_AlwaysTriggersSuperposition(t *testing.T) {
	um1 := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"stateB"}, nil),
			),
			"stateB": newTransitionReality("stateB",
				// always → external universe → initSuperposition at doCyclicTransition line 590-591
				withAlways([]string{"U:u2"}, nil),
			),
		},
	}
	um2 := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		// NO initial — u2 stays uninitialized until routed to
		Realities: map[string]*theoretical.RealityModel{
			"waitState": newTransitionReality("waitState"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um1, "u2": um2},
		Initials:      []string{"U:u1"}, // only u1 in initials
	}
	u1 := NewExUniverse(um1)
	u2 := NewExUniverse(um2)
	qmRaw, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	qm := qmRaw.(*ExQuantumMachine)

	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	// After always → U:u2, u1 enters superposition (external universe target)
	if !u1.inSuperposition {
		t.Fatalf("expected u1 in superposition after always→U:u2, got inSuperposition=%v", u1.inSuperposition)
	}
	if u1.currentReality != nil {
		t.Fatalf("expected u1.currentReality=nil in superposition, got %q", *u1.currentReality)
	}

	// u2 should now be initialized (the machine routed the event to it)
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized after external target routing")
	}
}

// Test: emit changes currentReality → always transitions of NEW reality execute (not old one).
func TestDoCyclicTransition_ReReadRealityAfterEmit(t *testing.T) {
	emitAction := "test:emit-reread"
	registerTestAction(t, emitAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("jump", nil)
		return nil
	})

	// stateC has always → FINAL. stateB does NOT have always.
	// After emit in stateB transitions to stateC, the always transitions of stateC (not stateB) must fire.
	var stateCEntryRan bool
	stateCEntryAction := "test:stateC-entry-reread"
	registerTestAction(t, stateCEntryAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		stateCEntryRan = true
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withEntryAction(emitAction),
			withOnTransition("jump", []string{"stateC"}, nil),
			// stateB has NO always transitions
		),
		"stateC": newTransitionReality("stateC",
			withEntryAction(stateCEntryAction),
			withAlways([]string{"FINAL"}, nil), // stateC always → FINAL
		),
		"FINAL": newFinalReality("FINAL"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	if !stateCEntryRan {
		t.Fatal("expected stateC entry action to run")
	}
	assertReality(t, u, "FINAL")
	if !u.isFinalReality {
		t.Fatal("expected isFinalReality=true — always of stateC (not stateB) should have fired")
	}
}

// Test: isFinalReality is correctly false through intermediate states and true only at final.
func TestDoCyclicTransition_IsFinalRealityCorrectThroughChain(t *testing.T) {
	var isFinalAtB, isFinalAtC bool

	entryB := "test:check-final-at-b"
	entryC := "test:check-final-at-c"
	registerTestAction(t, entryB, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		// Access snapshot to check finalization status — isFinalReality reflects universe internal state
		// At stateB (transition type), this should be false
		isFinalAtB = args.GetSnapshot().Resume.FinalizedUniverses != nil &&
			len(args.GetSnapshot().Resume.FinalizedUniverses) > 0
		return nil
	})
	registerTestAction(t, entryC, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		isFinalAtC = args.GetSnapshot().Resume.FinalizedUniverses != nil &&
			len(args.GetSnapshot().Resume.FinalizedUniverses) > 0
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withEntryAction(entryB),
			withAlways([]string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC",
			withEntryAction(entryC),
			withAlways([]string{"DONE"}, nil),
		),
		"DONE": newFinalReality("DONE"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	assertReality(t, u, "DONE")
	if !u.isFinalReality {
		t.Fatal("expected isFinalReality=true at DONE")
	}

	// During stateB entry, universe should NOT have been finalized
	if isFinalAtB {
		t.Fatal("isFinalReality should be false when entering stateB (transition type)")
	}
	// During stateC entry, universe should NOT have been finalized
	if isFinalAtC {
		t.Fatal("isFinalReality should be false when entering stateC (transition type)")
	}
}

// Test: executeAlways receives the correct realityModel — validates that after emit changes
// currentReality, the always transitions belong to the NEW reality, not the old one.
func TestEstablishNewReality_AlwaysUsesCorrectRealityModel(t *testing.T) {
	emitAction := "test:emit-always-correct-model"
	registerTestAction(t, emitAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go", nil)
		return nil
	})

	// Track which condition ran
	var conditionCalled bool
	condName := "test:cond-stateB-always"
	registerTestCondition(t, condName, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		conditionCalled = true
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		// stateA: has emit entry action, On["go"] → stateB. NO always transitions.
		"stateA": newTransitionReality("stateA",
			withEntryAction(emitAction),
			withOnTransition("go", []string{"stateB"}, nil),
		),
		// stateB: has always → stateC WITH a condition. This condition MUST be called.
		"stateB": newTransitionReality("stateB",
			withAlways([]string{"stateC"}, &theoretical.ConditionModel{Src: condName}),
		),
		"stateC": newFinalReality("stateC"),
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// stateA entry emits "go" → On["go"] → stateB → always(condition) → stateC
	assertReality(t, u, "stateC")

	if !conditionCalled {
		t.Fatal("expected stateB's always condition to be called, but it was not — executeAlways may be using wrong realityModel")
	}
	if !u.isFinalReality {
		t.Fatal("expected isFinalReality=true at stateC (final)")
	}
}

// Bug #2 regression: getApprovedTransition must not panic when Condition is nil but Conditions has an error.
func TestGetApprovedTransition_NilConditionWithConditionsError(t *testing.T) {
	condName := "test:cond-returns-error"
	registerTestCondition(t, condName, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, fmt.Errorf("intentional condition error")
	})

	entryAction := "test:emit-for-cond-error"
	registerTestAction(t, entryAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("trigger-cond-error", nil)
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(entryAction),
			func(r *theoretical.RealityModel) {
				r.On["trigger-cond-error"] = []*theoretical.TransitionModel{{
					Targets: []string{"stateB"},
					// Condition is nil, but Conditions has an entry that returns error
					Conditions: []*theoretical.ConditionModel{
						{Src: condName},
					},
				}}
			},
		),
		"stateB": newFinalReality("stateB"),
	}

	qm, _ := buildQM(t, "stateA", realities)

	// This should return an error, NOT panic
	err := qm.Init(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from condition, got nil")
	}
	if !strings.Contains(err.Error(), "intentional condition error") {
		t.Fatalf("expected error containing 'intentional condition error', got: %v", err)
	}
}
