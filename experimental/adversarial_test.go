package experimental

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// Adversarial / hostile scenarios: force complex but possible combinations,
// failure paths, races, and malformed executor behavior.

// ---------------------------------------------------------------------------
// Mixed routing: concrete On handler must NOT also broadcast into superposition
// ---------------------------------------------------------------------------

func TestAdv_SendEvent_ConcreteHandlerNoBroadcastToSuperposition(t *testing.T) {
	obsHits := atomic.Int32{}
	obsName := "test:adv:no-broadcast-obs"
	registerTestObserver(t, obsName, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		obsHits.Add(1)
		return false, nil
	})

	u1Model := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Concrete", Initial: strPtr("idle"),
		Realities: map[string]*theoretical.RealityModel{
			"idle": newTransitionReality("idle",
				withOnTransition("ping", []string{"done"}, nil),
			),
			"done": newFinalReality("done"),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Superposed",
		Realities: map[string]*theoretical.RealityModel{
			"wait": newTransitionReality("wait", withObserver(obsName, nil)),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-adv-broadcast", CanonicalName: "Adv", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{"U:u1", "U:u2"},
	}
	u1, u2 := NewExUniverse(u1Model), NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	assertReality(t, u1, "idle")
	assertSuperposition(t, u2)

	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("ping").Build())
	if err != nil {
		t.Fatalf("SendEvent: %v", err)
	}
	if !handled {
		t.Fatal("expected handled=true via concrete On")
	}
	assertReality(t, u1, "done")
	assertSuperposition(t, u2)
	if hits := obsHits.Load(); hits != 0 {
		t.Fatalf("superposition observer invoked %d times; concrete On must not broadcast", hits)
	}
}

// ---------------------------------------------------------------------------
// Directed external target still reaches superposition without double delivery
// ---------------------------------------------------------------------------

func TestAdv_DirectedExternalTarget_CollapsesSuperpositionOnce(t *testing.T) {
	obsHits := atomic.Int32{}
	obsName := "test:adv:directed-once-obs"
	registerTestObserver(t, obsName, func(_ context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
		obsHits.Add(1)
		return args.GetEvent().GetEventName() == "wake", nil
	})

	u1Model := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "Router", Initial: strPtr("idle"),
		Realities: map[string]*theoretical.RealityModel{
			"idle": {
				ID: "idle", Type: theoretical.RealityTypeTransition,
				On: map[string][]*theoretical.TransitionModel{
					"wake": {notifyTransition([]string{"U:u2"})},
				},
			},
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID: "u2", CanonicalName: "Waiting",
		Realities: map[string]*theoretical.RealityModel{
			"ready": newTransitionReality("ready", withObserver(obsName, nil)),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-adv-directed", CanonicalName: "Adv", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:  []string{"U:u1", "U:u2"},
	}
	u1, u2 := NewExUniverse(u1Model), NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("wake").Build()); err != nil {
		t.Fatalf("SendEvent: %v", err)
	}
	assertReality(t, u2, "ready")
	if hits := obsHits.Load(); hits != 1 {
		t.Fatalf("observer hits=%d, want exactly 1 (no double delivery)", hits)
	}
	_ = u1
}

// ---------------------------------------------------------------------------
// Invoke panic must not crash the process / block the machine
// ---------------------------------------------------------------------------

func TestAdv_InvokePanic_DoesNotCrashMachine(t *testing.T) {
	panicked := make(chan struct{}, 1)
	invokeName := "test:adv:invoke-panic"
	registerTestInvoke(t, invokeName, func(_ context.Context, _ instrumentation.InvokeExecutorArgs) {
		panicked <- struct{}{}
		panic("intentional adversarial invoke panic")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryInvoke(invokeName),
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	select {
	case <-panicked:
	case <-time.After(2 * time.Second):
		t.Fatal("invoke never started")
	}
	// Give the panic a moment to fire inside the goroutine.
	time.Sleep(20 * time.Millisecond)

	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err != nil {
		t.Fatalf("SendEvent after invoke panic: %v", err)
	}
	if !handled {
		t.Fatal("expected handled after invoke panic")
	}
	assertReality(t, u, "stateB")
}

func TestAdv_MachineConstantInvokePanic_DoesNotCrash(t *testing.T) {
	panicked := make(chan struct{}, 1)
	invokeName := "test:adv:machine-invoke-panic"
	registerTestInvoke(t, invokeName, func(_ context.Context, _ instrumentation.InvokeExecutorArgs) {
		panicked <- struct{}{}
		panic("machine-level invoke panic")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
	}
	um := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "U", Initial: strPtr("stateA"), Realities: realities,
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-adv-mconst", CanonicalName: "Adv", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
		Initials:  []string{"U:u1"},
		UniversalConstants: &theoretical.UniversalConstantsModel{
			EntryInvokes: []*theoretical.InvokeModel{{Src: invokeName}},
		},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	select {
	case <-panicked:
	case <-time.After(2 * time.Second):
		t.Fatal("machine invoke never started")
	}
	time.Sleep(20 * time.Millisecond)
	assertReality(t, u, "stateA")
}

// ---------------------------------------------------------------------------
// Universe-level constant action failure
// ---------------------------------------------------------------------------

func TestAdv_UniverseConstantEntryAction_FailureAbortsInit(t *testing.T) {
	actionName := "test:adv:uconst-fail"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("universe constant boom")
	})

	um := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "U", Initial: strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA"),
		},
		UniversalConstants: &theoretical.UniversalConstantsModel{
			EntryActions: []*theoretical.ActionModel{{Src: actionName}},
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-adv-uconst", CanonicalName: "Adv", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
		Initials:  []string{"U:u1"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err == nil {
		t.Fatal("expected Init to fail when universe constant entry action fails")
	}
	if u.currentReality != nil {
		t.Fatalf("expected no reality after failed universe constants, got %q", *u.currentReality)
	}
}

// ---------------------------------------------------------------------------
// ReplayOnEntry propagates entry action failure
// ---------------------------------------------------------------------------

func TestAdv_ReplayOnEntry_EntryActionFailure(t *testing.T) {
	actionName := "test:adv:replay-fail"
	calls := atomic.Int32{}
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		if calls.Add(1) == 1 {
			return nil // first Init succeeds
		}
		return fmt.Errorf("replay boom")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	assertReality(t, u, "stateA")

	if err := qm.ReplayOnEntry(context.Background()); err == nil {
		t.Fatal("expected ReplayOnEntry to fail when entry action fails")
	}
	assertReality(t, u, "stateA")
}

// ---------------------------------------------------------------------------
// Conditions mutate metadata (covers condition metadata helpers)
// ---------------------------------------------------------------------------

func TestAdv_Condition_MutatesMetadataDuringGuard(t *testing.T) {
	condName := "test:adv:cond-meta"
	registerTestCondition(t, condName, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		args.AddToUniverseMetadata("gate", "open")
		args.UpdateUniverseMetadata(map[string]any{"gate": "open", "stamp": 1})
		md := args.GetUniverseMetadata()
		if md["gate"] != "open" {
			return false, fmt.Errorf("metadata gate=%v", md["gate"])
		}
		prev, ok := args.DeleteFromUniverseMetadata("stamp")
		if !ok || prev != 1 {
			return false, fmt.Errorf("delete stamp: ok=%v prev=%v", ok, prev)
		}
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
		t.Fatalf("Init: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent: %v", err)
	}
	assertReality(t, u, "stateB")
	snap := qm.GetSnapshot()
	raw, ok := snap.Snapshots["u1"]["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("snapshot metadata missing: %#v", snap.Snapshots["u1"])
	}
	if raw["gate"] != "open" {
		t.Fatalf("metadata gate=%v, want open", raw["gate"])
	}
	if _, exists := raw["stamp"]; exists {
		t.Fatal("stamp should have been deleted by condition")
	}
}

// ---------------------------------------------------------------------------
// Invoke metadata delete/update under race with GetSnapshot
// ---------------------------------------------------------------------------

func TestAdv_Invoke_DeleteAndUpdateMetadata_RaceFree(t *testing.T) {
	started := make(chan struct{})
	done := make(chan struct{})
	invokeName := "test:adv:invoke-meta-race"
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		close(started)
		for i := 0; i < 200; i++ {
			args.UpdateUniverseMetadata(map[string]any{"i": i, "k": "v"})
			args.AddToUniverseMetadata("tmp", i)
			args.DeleteFromUniverseMetadata("tmp")
		}
		close(done)
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryInvoke(invokeName)),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	<-started

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = qm.GetSnapshot()
		}
	}()
	<-done
	wg.Wait()
}

// ---------------------------------------------------------------------------
// Concurrent SendEvent + LoadSnapshot + GetSnapshot
// ---------------------------------------------------------------------------

func TestAdv_Concurrent_SendEvent_LoadSnapshot_GetSnapshot(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
			withOnTransition("back", []string{"stateA"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withOnTransition("back", []string{"stateA"}, nil),
			withOnTransition("go", []string{"stateB"}, nil),
		),
	}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	base := qm.GetSnapshot()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			name := "go"
			if i%2 == 0 {
				name = "back"
			}
			if _, err := qm.SendEvent(context.Background(), NewEventBuilder(name).Build()); err != nil {
				errCh <- err
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = qm.GetSnapshot()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			if err := qm.LoadSnapshot(base, nil); err != nil {
				errCh <- err
				return
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent adversarial stress timed out")
	case err := <-errCh:
		t.Fatalf("concurrent stress error: %v", err)
	}
	_ = ctx
}

// ---------------------------------------------------------------------------
// Cascade depth exactly at the limit vs over the limit
// ---------------------------------------------------------------------------

func TestAdv_NotifyCascade_DepthAtLimitSucceeds(t *testing.T) {
	// Local SendEvent on u0 + cascade depths 0..10 inclusive (maxExternalTargetDepth=10).
	// Universes: u0..u11. u0..u10 notify next; u11 accepts hop with empty targets.
	const n = 12
	universes := map[string]*theoretical.UniverseModel{}
	ex := make([]*ExUniverse, 0, n)
	initials := make([]string, 0, n)

	for i := 0; i < n; i++ {
		id := fmt.Sprintf("u%d", i)
		var on map[string][]*theoretical.TransitionModel
		if i < n-1 {
			next := fmt.Sprintf("U:u%d", i+1)
			on = map[string][]*theoretical.TransitionModel{
				"hop": {notifyTransition([]string{next})},
			}
		} else {
			on = map[string][]*theoretical.TransitionModel{
				"hop": {{Targets: []string{}}},
			}
		}
		um := &theoretical.UniverseModel{
			ID: id, CanonicalName: id, Initial: strPtr("s"),
			Realities: map[string]*theoretical.RealityModel{
				"s": {ID: "s", Type: theoretical.RealityTypeTransition, On: on},
			},
		}
		universes[id] = um
		ex = append(ex, NewExUniverse(um))
		initials = append(initials, "U:"+id)
	}

	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-adv-depth", CanonicalName: "Adv", Version: "1.0.0",
		Universes: universes, Initials: initials,
	}
	qm, err := NewExQuantumMachine(qmm, ex)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	_, err = qm.SendEvent(context.Background(), NewEventBuilder("hop").Build())
	if err != nil {
		t.Fatalf("cascade at depth limit should succeed, got: %v", err)
	}
}

func TestAdv_NotifyCascade_DepthOverLimitErrors(t *testing.T) {
	// One hop past the limit: cascade tries to schedule depth 11.
	const n = 13 // u0..u12
	universes := map[string]*theoretical.UniverseModel{}
	ex := make([]*ExUniverse, 0, n)
	initials := make([]string, 0, n)

	for i := 0; i < n; i++ {
		id := fmt.Sprintf("v%d", i)
		var on map[string][]*theoretical.TransitionModel
		if i < n-1 {
			next := fmt.Sprintf("U:v%d", i+1)
			on = map[string][]*theoretical.TransitionModel{
				"hop": {notifyTransition([]string{next})},
			}
		} else {
			on = map[string][]*theoretical.TransitionModel{
				"hop": {{Targets: []string{}}},
			}
		}
		um := &theoretical.UniverseModel{
			ID: id, CanonicalName: id, Initial: strPtr("s"),
			Realities: map[string]*theoretical.RealityModel{
				"s": {ID: "s", Type: theoretical.RealityTypeTransition, On: on},
			},
		}
		universes[id] = um
		ex = append(ex, NewExUniverse(um))
		initials = append(initials, "U:"+id)
	}

	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-adv-overdepth", CanonicalName: "Adv", Version: "1.0.0",
		Universes: universes, Initials: initials,
	}
	qm, err := NewExQuantumMachine(qmm, ex)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	_, err = qm.SendEvent(context.Background(), NewEventBuilder("hop").Build())
	if err == nil {
		t.Fatal("expected cascade depth error")
	}
}

// ---------------------------------------------------------------------------
// Many observers all error → error propagates; accumulator still advanced
// ---------------------------------------------------------------------------

func TestAdv_AllObserversError_Propagates(t *testing.T) {
	o1 := "test:adv:all-err-1"
	o2 := "test:adv:all-err-2"
	registerTestObserver(t, o1, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		return false, fmt.Errorf("err1")
	})
	registerTestObserver(t, o2, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		return false, fmt.Errorf("err2")
	})

	um := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "U",
		Realities: map[string]*theoretical.RealityModel{
			"wait": newTransitionReality("wait",
				withObserver(o1, nil),
				withObserver(o2, nil),
			),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-adv-allobs", CanonicalName: "Adv", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
		Initials:  []string{"U:u1"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	_, err = qm.SendEvent(context.Background(), NewEventBuilder("x").Build())
	if err == nil {
		t.Fatal("expected error when every observer fails")
	}
	assertSuperposition(t, u)
}

// ---------------------------------------------------------------------------
// Always + On race-like: always chain then event must still work
// ---------------------------------------------------------------------------

func TestAdv_AlwaysChainThenEvent_Stable(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"a": newTransitionReality("a", withAlways([]string{"b"}, nil)),
		"b": newTransitionReality("b", withAlways([]string{"c"}, nil)),
		"c": newTransitionReality("c",
			withOnTransition("fin", []string{"done"}, nil),
		),
		"done": newFinalReality("done"),
	}
	qm, u := buildQM(t, "a", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
	assertReality(t, u, "c")
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("fin").Build()); err != nil {
		t.Fatalf("SendEvent: %v", err)
	}
	assertReality(t, u, "done")
}

// ---------------------------------------------------------------------------
// Duplicate identical events under superposition until collapse
// ---------------------------------------------------------------------------

func TestAdv_Superposition_DuplicateEventsUntilCollapse(t *testing.T) {
	obsName := "test:adv:dup-collapse"
	registerTestObserver(t, obsName, func(_ context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
		return args.GetAccumulatorStatistics().CountAllEvents() >= 5, nil
	})

	um := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "U",
		Realities: map[string]*theoretical.RealityModel{
			"ready": newTransitionReality("ready", withObserver(obsName, nil)),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm-adv-dup", CanonicalName: "Adv", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
		Initials:  []string{"U:u1"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	for i := 0; i < 4; i++ {
		handled, err := qm.SendEvent(context.Background(), NewEventBuilder("tick").Build())
		if err != nil {
			t.Fatalf("tick %d: %v", i, err)
		}
		if !handled {
			t.Fatalf("tick %d not handled", i)
		}
		assertSuperposition(t, u)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("tick").Build()); err != nil {
		t.Fatalf("collapse tick: %v", err)
	}
	assertReality(t, u, "ready")
}

// ---------------------------------------------------------------------------
// Transition with empty event name / nil data / huge payload
// ---------------------------------------------------------------------------

func TestAdv_SendEvent_NilDataAndEmptyName(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
			withOnTransition("", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("").Build())
	if err != nil {
		t.Fatalf("empty name: %v", err)
	}
	if handled {
		assertReality(t, u, "stateB")
		return
	}

	// Empty name may be unhandled; nil-data event with real name must work.
	evt := NewEventBuilder("go").SetData(nil).Build()
	handled, err = qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("nil data: %v", err)
	}
	if !handled {
		t.Fatal("expected handled for go with nil data")
	}
	assertReality(t, u, "stateB")
}

func TestAdv_SendEvent_LargePayload(t *testing.T) {
	condName := "test:adv:large-payload"
	registerTestCondition(t, condName, func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		data := args.GetEvent().GetData()
		blob, ok := data["blob"].(string)
		if !ok || len(blob) < 100_000 {
			return false, fmt.Errorf("blob missing or short")
		}
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
		t.Fatalf("Init: %v", err)
	}

	blob := make([]byte, 128*1024)
	for i := range blob {
		blob[i] = byte('a' + (i % 26))
	}
	evt := NewEventBuilder("go").SetData(map[string]any{"blob": string(blob)}).Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("large payload: %v", err)
	}
	assertReality(t, u, "stateB")
}

// ---------------------------------------------------------------------------
// NewExQuantumMachine rejects duplicate universe IDs / nil universes
// ---------------------------------------------------------------------------

func TestAdv_NewExQuantumMachine_DuplicateUniverseID(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "U", Realities: map[string]*theoretical.RealityModel{
			"a": newTransitionReality("a"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm", CanonicalName: "Q", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
	}
	u1, u2 := NewExUniverse(um), NewExUniverse(um)
	if _, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2}); err == nil {
		t.Fatal("expected error for duplicate universe IDs")
	}
}

func TestAdv_NewExQuantumMachine_NilUniverseSkipped(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID: "u1", CanonicalName: "U", Initial: strPtr("a"),
		Realities: map[string]*theoretical.RealityModel{"a": newTransitionReality("a")},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID: "qm", CanonicalName: "Q", Version: "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{"u1": um},
		Initials:  []string{"U:u1"},
	}
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{nil, NewExUniverse(um)})
	if err != nil {
		t.Fatalf("nil universe should be skipped: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Fuzz refs — adversarial garbage must never panic
// ---------------------------------------------------------------------------

func FuzzProcessReference(f *testing.F) {
	seeds := []string{
		"", "U:", "U:a", "U:a:b", "U:a:b:c", "a", "U::", "U: ", "U:a:",
		"u:a", "U:😀", "U:" + string(make([]byte, 256)),
		"U:1", "U:a:1", "::::::", "U:a/b", "U:a.b:c_d",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, ref string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("processReference(%q) panicked: %v", ref, r)
			}
		}()
		_, _, _ = processReference(ref)
	})
}
