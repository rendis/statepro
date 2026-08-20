package experimental

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// Este archivo cubre gaps y defectos del runtime demostrados con comportamiento observable.
// Cada test nombra el contrato que se espera (docs, comentarios de código o schema).

// ---------------------------------------------------------------------------
// Snapshots
// ---------------------------------------------------------------------------

func TestGap_LoadSnapshot_RestauraTracking(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("goB", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withOnTransition("goC", []string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC"),
	}

	qm1, _ := buildQM(t, "stateA", realities)
	if err := qm1.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm1.SendEvent(context.Background(), NewEventBuilder("goB").Build()); err != nil {
		t.Fatalf("SendEvent(goB) failed: %v", err)
	}
	if _, err := qm1.SendEvent(context.Background(), NewEventBuilder("goC").Build()); err != nil {
		t.Fatalf("SendEvent(goC) failed: %v", err)
	}

	snap := qm1.GetSnapshot()
	want := []string{"stateA", "stateB", "stateC"}
	if got := snap.Tracking["u1"]; len(got) != len(want) {
		t.Fatalf("snapshot tracking before load: got %v, want %v", got, want)
	}

	qm2, u2 := buildQM(t, "stateA", realities)
	if err := qm2.LoadSnapshot(snap, nil); err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}

	got := qm2.GetSnapshot().Tracking["u1"]
	if len(got) != len(want) {
		t.Fatalf("tracking after LoadSnapshot: got %v, want %v (docs: LoadSnapshot restaura el historial)", got, want)
	}
	for i, exp := range want {
		if got[i] != exp {
			t.Fatalf("tracking[%d] after LoadSnapshot: got %q, want %q", i, got[i], exp)
		}
	}
	if len(u2.tracking) != len(want) {
		t.Fatalf("universe.tracking after LoadSnapshot: got %v, want %v", u2.tracking, want)
	}
}

func TestGap_GetSnapshot_TrackingEsCopiaIndependiente(t *testing.T) {
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
	before := append([]string(nil), snap.Tracking["u1"]...)

	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	afterSnap := snap.Tracking["u1"]
	if len(afterSnap) != len(before) {
		t.Fatalf("snapshot tracking mutated after later SendEvent: before %v, after %v — GetSnapshot must copy the slice", before, afterSnap)
	}
	for i := range before {
		if afterSnap[i] != before[i] {
			t.Fatalf("snapshot tracking[%d] mutated: got %q, want %q", i, afterSnap[i], before[i])
		}
	}
}

func TestGap_GetSnapshot_ConcurrentWithSendEvent(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("ping", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withOnTransition("pong", []string{"stateA"}, nil),
		),
	}

	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = qm.SendEvent(context.Background(), NewEventBuilder("ping").Build())
			_, _ = qm.SendEvent(context.Background(), NewEventBuilder("pong").Build())
		}()
		go func() {
			defer wg.Done()
			_ = qm.GetSnapshot()
		}()
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// Estados finales y ReplayOnEntry
// ---------------------------------------------------------------------------

func TestGap_FinalReality_IgnoraHandlersOn(t *testing.T) {
	// RealityTypeFinal documenta que On se ignora. canHandleEvent también lo declara.
	done := newFinalReality("DONE")
	done.On = map[string][]*theoretical.TransitionModel{
		"revive": {{Targets: []string{"stateA"}}},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"DONE"}, nil),
		),
		"DONE": done,
	}

	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent(go) failed: %v", err)
	}
	assertReality(t, u, "DONE")
	assertFinalized(t, u)

	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("revive").Build())
	if err != nil {
		t.Fatalf("SendEvent(revive) failed: %v", err)
	}
	if handled {
		t.Fatal("expected final reality On handlers to be ignored")
	}
	assertReality(t, u, "DONE")
}

func TestGap_ReplayOnEntry_NoEjecutaUniversosFinalizados(t *testing.T) {
	actionName := "test:gap:replay-final-entry"
	var calls int
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		calls++
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"DONE"}, nil),
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
	assertFinalized(t, u)
	if calls != 1 {
		t.Fatalf("expected 1 entry call on init, got %d", calls)
	}

	if err := qm.ReplayOnEntry(context.Background()); err != nil {
		t.Fatalf("ReplayOnEntry failed: %v", err)
	}
	if calls != 1 {
		t.Fatalf("ReplayOnEntry must skip finalized universes, got %d calls (want 1)", calls)
	}
}

// ---------------------------------------------------------------------------
// Universal constants: ActionType y nivel universo
// ---------------------------------------------------------------------------

func TestGap_Constants_ActionTypeCorrecto(t *testing.T) {
	var entryType, exitType, transType instrumentation.ActionType
	entrySrc := "test:gap:const-entry-type"
	exitSrc := "test:gap:const-exit-type"
	transSrc := "test:gap:const-trans-type"

	registerTestAction(t, entrySrc, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		entryType = args.GetActionType()
		return nil
	})
	registerTestAction(t, exitSrc, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		exitType = args.GetActionType()
		return nil
	})
	registerTestAction(t, transSrc, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		transType = args.GetActionType()
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
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
			EntryActions:        []*theoretical.ActionModel{{Src: entrySrc}},
			ExitActions:         []*theoretical.ActionModel{{Src: exitSrc}},
			ActionsOnTransition: []*theoretical.ActionModel{{Src: transSrc}},
		},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	if entryType != instrumentation.ActionTypeEntry {
		t.Fatalf("constants entry ActionType=%q, want %q", entryType, instrumentation.ActionTypeEntry)
	}
	if exitType != instrumentation.ActionTypeExit {
		t.Fatalf("constants exit ActionType=%q, want %q", exitType, instrumentation.ActionTypeExit)
	}
	if transType != instrumentation.ActionTypeTransition {
		t.Fatalf("constants transition ActionType=%q, want %q", transType, instrumentation.ActionTypeTransition)
	}
}

func TestGap_UniverseConstants_SeEjecutan(t *testing.T) {
	var order []string
	machineEntry := "test:gap:uconst-machine"
	universeEntry := "test:gap:uconst-universe"
	realityEntry := "test:gap:uconst-reality"

	registerTestAction(t, machineEntry, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "machine")
		return nil
	})
	registerTestAction(t, universeEntry, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "universe")
		return nil
	})
	registerTestAction(t, realityEntry, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "reality")
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(realityEntry)),
	}
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Initial:       strPtr("stateA"),
		Realities:     realities,
		UniversalConstants: &theoretical.UniversalConstantsModel{
			EntryActions: []*theoretical.ActionModel{{Src: universeEntry}},
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": um},
		Initials:      []string{"U:u1"},
		UniversalConstants: &theoretical.UniversalConstantsModel{
			EntryActions: []*theoretical.ActionModel{{Src: machineEntry}},
		},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	want := []string{"machine", "universe", "reality"}
	if len(order) != len(want) {
		t.Fatalf("universe-level UniversalConstants not executed. order=%v, want %v", order, want)
	}
	for i, exp := range want {
		if order[i] != exp {
			t.Fatalf("order[%d]=%q, want %q (docs: machine → universe → reality)", i, order[i], exp)
		}
	}
}

// ---------------------------------------------------------------------------
// Rollback y fallos de acciones
// ---------------------------------------------------------------------------

func TestGap_EntryActionFailure_InitNoDejaRealityAMedias(t *testing.T) {
	actionName := "test:gap:entry-fail-init"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("entry boom")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)
	err := qm.Init(context.Background(), nil)
	if err == nil {
		t.Fatal("expected Init error")
	}

	if u.initialized {
		t.Fatal("expected initialized=false after failed entry")
	}
	if u.currentReality != nil {
		t.Fatalf("expected currentReality=nil after failed entry, got %s (docs: last reality is restored)", *u.currentReality)
	}
}

func TestGap_EntryActionFailure_TransicionRestauraRealityAnterior(t *testing.T) {
	actionName := "test:gap:entry-fail-transition"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("entry boom on B")
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB", withEntryAction(actionName)),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	_, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err == nil {
		t.Fatal("expected SendEvent error from failed entry of stateB")
	}

	assertReality(t, u, "stateA")
}

func TestGap_TransitionActionFailure_PermaneceEnRealityOrigen(t *testing.T) {
	actionName := "test:gap:trans-fail-stay"
	registerTestAction(t, actionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return fmt.Errorf("transition boom")
	})

	stateA := newTransitionReality("stateA")
	stateA.On = map[string][]*theoretical.TransitionModel{
		"go": {{
			Targets: []string{"stateB"},
			Actions: []*theoretical.ActionModel{{Src: actionName}},
		}},
	}
	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err == nil {
		t.Fatal("expected transition action error")
	}
	assertReality(t, u, "stateA")
}

// ---------------------------------------------------------------------------
// Observers: semántica documentada
// ---------------------------------------------------------------------------

func TestGap_ObserverError_SeIgnoraSiOtroAprueba(t *testing.T) {
	errObs := "test:gap:obs-err-then-ok"
	okObs := "test:gap:obs-ok-after-err"
	registerTestObserver(t, errObs, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		return false, fmt.Errorf("observer failed")
	})
	registerTestObserver(t, okObs, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		return true, nil
	})

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
			"stateX": newTransitionReality("stateX",
				withObserver(errObs, nil),
				withObserver(okObs, nil),
			),
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
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("expected observer error to be ignored when another observer approves, got: %v", err)
	}
	assertReality(t, u2, "stateX")
}

func TestGap_ObserverError_SePropagaSiNadieAprueba(t *testing.T) {
	errObs := "test:gap:obs-err-only"
	registerTestObserver(t, errObs, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		return false, fmt.Errorf("observer failed")
	})

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
			"stateX": newTransitionReality("stateX",
				withObserver(errObs, nil),
			),
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
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	_, err = qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err == nil {
		t.Fatal("expected observer error when no observer approves")
	}
}

func TestGap_SuperpositionCollapse_DeterministaPorID(t *testing.T) {
	// Con dos realities que pueden colapsar, el ganador no debe depender del orden del map.
	const iterations = 30
	seen := map[string]int{}

	for i := 0; i < iterations; i++ {
		u1Model := &theoretical.UniverseModel{
			ID:            "u1",
			CanonicalName: "Universe1",
			Initial:       strPtr("router"),
			Realities: map[string]*theoretical.RealityModel{
				"router": newTransitionReality("router",
					withOnTransition("go", []string{"U:u2"}, nil),
				),
			},
		}
		u2Model := &theoretical.UniverseModel{
			ID:            "u2",
			CanonicalName: "Universe2",
			Realities: map[string]*theoretical.RealityModel{
				"zeta": newTransitionReality("zeta",
					withObserver("builtin:observer:alwaysTrue", nil),
				),
				"alpha": newTransitionReality("alpha",
					withObserver("builtin:observer:alwaysTrue", nil),
				),
			},
		}
		qmm := &theoretical.QuantumMachineModel{
			ID:            "qm1",
			CanonicalName: "TestQM",
			Version:       "1.0.0",
			Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
			Initials:      []string{"U:u1"},
		}
		u2 := NewExUniverse(u2Model)
		qm, err := NewExQuantumMachine(qmm, []*ExUniverse{NewExUniverse(u1Model), u2})
		if err != nil {
			t.Fatalf("build failed: %v", err)
		}
		if err := qm.Init(context.Background(), nil); err != nil {
			t.Fatalf("Init failed: %v", err)
		}
		if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
			t.Fatalf("SendEvent failed: %v", err)
		}
		if u2.currentReality == nil {
			t.Fatal("expected collapse to a reality")
		}
		seen[*u2.currentReality]++
	}

	if len(seen) != 1 {
		t.Fatalf("collapse was non-deterministic across %d runs: %v (expected stable winner by sorted reality id)", iterations, seen)
	}
	if _, ok := seen["alpha"]; !ok {
		t.Fatalf("expected deterministic collapse to 'alpha' (sorted id), got %v", seen)
	}
}

// ---------------------------------------------------------------------------
// Ciclos y referencias inválidas
// ---------------------------------------------------------------------------

func TestGap_AlwaysCycle_RetornaErrorSinColgarse(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withAlways([]string{"stateA"}, nil),
		),
	}
	qm, _ := buildQM(t, "stateA", realities)

	done := make(chan error, 1)
	go func() {
		done <- qm.Init(context.Background(), nil)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error for always-transition cycle")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "cycle") &&
			!strings.Contains(strings.ToLower(err.Error()), "cyclic") &&
			!strings.Contains(err.Error(), "loop") {
			t.Fatalf("expected cycle-related error, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Init hung on always A↔B cycle — missing cycle protection")
	}
}

func TestGap_NotifyCycle_RetornaErrorSinColgarse(t *testing.T) {
	notifyType := theoretical.TransitionTypeNotify
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA", func(r *theoretical.RealityModel) {
				r.On["ping"] = []*theoretical.TransitionModel{{
					Type:    &notifyType,
					Targets: []string{"U:u2"},
				}}
			}),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Initial:       strPtr("stateX"),
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX", func(r *theoretical.RealityModel) {
				r.On["ping"] = []*theoretical.TransitionModel{{
					Type:    &notifyType,
					Targets: []string{"U:u1"},
				}}
			}),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:      []string{"U:u1", "U:u2"},
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, sendErr := qm.SendEvent(context.Background(), NewEventBuilder("ping").Build())
		done <- sendErr
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error for notify cycle U1↔U2")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("SendEvent hung on notify cycle — missing cascade depth limit")
	}
}

func TestGap_InvalidExternalTarget_NoPanic(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"U:u2:stateX", "!!!invalid"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Initial:       strPtr("stateX"),
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model, "u2": u2Model},
		Initials:      []string{"U:u1", "U:u2"},
	}
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{NewExUniverse(u1Model), NewExUniverse(u2Model)})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("invalid external target panicked: %v", r)
		}
	}()

	_, err = qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err == nil {
		t.Fatal("expected error for invalid external target, got nil")
	}
}

func TestGap_SingleCharIDs_SonReferenciasValidas(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u",
		CanonicalName: "U",
		Initial:       strPtr("a"),
		Realities: map[string]*theoretical.RealityModel{
			"a": newTransitionReality("a",
				withOnTransition("go", []string{"b"}, nil),
			),
			"b": newTransitionReality("b"),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "m",
		CanonicalName: "M",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u": um},
		Initials:      []string{"U:u"},
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init with single-char IDs failed: %v", err)
	}
	assertReality(t, u, "a")

	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build()); err != nil {
		t.Fatalf("SendEvent with single-char reality IDs failed: %v", err)
	}
	assertReality(t, u, "b")
}

func TestGap_GetSnapshotFromAction_NoDeadlock(t *testing.T) {
	actionName := "test:gap:snapshot-no-deadlock"
	var snapOK atomic.Bool
	registerTestAction(t, actionName, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		if args.GetSnapshot() != nil {
			snapOK.Store(true)
		}
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction(actionName)),
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
	case <-time.After(2 * time.Second):
		t.Fatal("GetSnapshot from entry action deadlocked with machine mutex")
	}

	if !snapOK.Load() {
		t.Fatal("expected GetSnapshot from entry action to return a snapshot")
	}
}

// ---------------------------------------------------------------------------
// TDD follow-up: SendEvent → superposición, e invokes vs metadata
// ---------------------------------------------------------------------------

func TestTDD_SendEvent_ColapsaUniversoEnSuperposicion(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withObserver("builtin:observer:alwaysTrue", nil),
			),
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
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertSuperposition(t, u)

	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("go").Build())
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !handled {
		t.Fatal("docs/runtime.md: SendEvent must accumulate events for superposition universes")
	}
	assertReality(t, u, "stateA")
	if u.inSuperposition {
		t.Fatal("expected observer alwaysTrue to collapse superposition after SendEvent")
	}
}

func TestTDD_SendEvent_AcumulaEventosHastaObserver(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "TestUniverse",
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withObserver("builtin:observer:containsAllEvents", map[string]any{
					"evt1": "tick",
					"evt2": "tock",
				}),
			),
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
		t.Fatalf("build failed: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertSuperposition(t, u)

	handled, err := qm.SendEvent(context.Background(), NewEventBuilder("tick").Build())
	if err != nil {
		t.Fatalf("SendEvent(tick) failed: %v", err)
	}
	if !handled {
		t.Fatal("expected SendEvent to handle tick while in superposition")
	}
	assertSuperposition(t, u)

	handled, err = qm.SendEvent(context.Background(), NewEventBuilder("tock").Build())
	if err != nil {
		t.Fatalf("SendEvent(tock) failed: %v", err)
	}
	if !handled {
		t.Fatal("expected SendEvent to handle tock while in superposition")
	}
	assertReality(t, u, "stateA")
}

func TestTDD_InvokeMetadata_ConcurrentWithGetSnapshot(t *testing.T) {
	started := make(chan struct{})
	invokeName := "test:tdd:invoke-meta-race"
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		close(started)
		for i := 0; i < 2000; i++ {
			args.AddToUniverseMetadata("n", i)
			_ = args.GetUniverseMetadata()
			args.UpdateUniverseMetadata(map[string]any{"n": i, "ok": true})
		}
	})

	stateA := newTransitionReality("stateA")
	stateA.EntryInvokes = []*theoretical.InvokeModel{{Src: invokeName}}
	realities := map[string]*theoretical.RealityModel{"stateA": stateA}
	qm, _ := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("invoke did not start")
	}

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = qm.GetSnapshot()
			}
		}()
	}
	wg.Wait()
}
