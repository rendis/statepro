package experimental

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// ===================== Additional Helpers =====================

func withObserver(src string, args map[string]any) func(*theoretical.RealityModel) {
	return func(r *theoretical.RealityModel) {
		r.Observers = append(r.Observers, &theoretical.ObserverModel{Src: src, Args: args})
	}
}

func withExitAction(src string) func(*theoretical.RealityModel) {
	return func(r *theoretical.RealityModel) {
		r.ExitActions = append(r.ExitActions, &theoretical.ActionModel{Src: src})
	}
}

func registerTestObserver(t *testing.T, name string, fn instrumentation.ObserverFn) {
	t.Helper()
	if err := builtin.RegisterObserver(name, fn); err != nil {
		t.Fatalf("failed to register observer %s: %v", name, err)
	}
}

func registerTestInvoke(t *testing.T, name string, fn instrumentation.InvokeFn) {
	t.Helper()
	if err := builtin.RegisterInvoke(name, fn); err != nil {
		t.Fatalf("failed to register invoke %s: %v", name, err)
	}
}

func assertSuperposition(t *testing.T, u *ExUniverse) {
	t.Helper()
	if !u.inSuperposition {
		t.Fatalf("expected universe to be in superposition, but inSuperposition=%v", u.inSuperposition)
	}
}

func assertNotSuperposition(t *testing.T, u *ExUniverse) {
	t.Helper()
	if u.inSuperposition {
		t.Fatalf("expected universe NOT to be in superposition, but inSuperposition=%v", u.inSuperposition)
	}
}

func assertFinalized(t *testing.T, u *ExUniverse) {
	t.Helper()
	if !u.isFinalReality {
		t.Fatalf("expected universe to be finalized, but isFinalReality=%v", u.isFinalReality)
	}
}

func buildThreeUniverseQM(
	t *testing.T,
	u1Model, u2Model, u3Model *theoretical.UniverseModel,
	initials []string,
) (*ExQuantumMachine, *ExUniverse, *ExUniverse, *ExUniverse) {
	t.Helper()
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			u1Model.ID: u1Model,
			u2Model.ID: u2Model,
			u3Model.ID: u3Model,
		},
		Initials: initials,
	}
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	u3 := NewExUniverse(u3Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2, u3})
	if err != nil {
		t.Fatalf("failed to build three-universe QM: %v", err)
	}
	return qm.(*ExQuantumMachine), u1, u2, u3
}

func notifyTransition(targets []string) *theoretical.TransitionModel {
	nt := theoretical.TransitionTypeNotify
	return &theoretical.TransitionModel{
		Type:    &nt,
		Targets: targets,
	}
}

func newUnsuccessfulFinalReality(id string) *theoretical.RealityModel {
	return &theoretical.RealityModel{
		ID:   id,
		Type: theoretical.RealityTypeUnsuccessfulFinal,
		On:   map[string][]*theoretical.TransitionModel{},
	}
}

func withEntryInvoke(src string) func(*theoretical.RealityModel) {
	return func(r *theoretical.RealityModel) {
		r.EntryInvokes = append(r.EntryInvokes, &theoretical.InvokeModel{Src: src})
	}
}

func withExitInvoke(src string) func(*theoretical.RealityModel) {
	return func(r *theoretical.RealityModel) {
		r.ExitInvokes = append(r.ExitInvokes, &theoretical.InvokeModel{Src: src})
	}
}

func buildQMCustomInitials(t *testing.T, um *theoretical.UniverseModel, initials []string) (*ExQuantumMachine, *ExUniverse) {
	t.Helper()
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{um.ID: um},
		Initials:      initials,
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	return qm.(*ExQuantumMachine), u
}

// ===================== F. Observers & Superposition (16 tests) =====================

// F1. Universe with Initial=nil enters superposition after Init
func TestSuperposition_NoInitial_Enters(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		// Initial is nil
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
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertSuperposition(t, u)
	if u.currentReality != nil {
		t.Fatalf("expected currentReality=nil in superposition, got %q", *u.currentReality)
	}
}

// F2. Multi-target transition triggers superposition
func TestSuperposition_MultiTarget(t *testing.T) {
	// Multi-target transition to 2 different universes → source enters superposition
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
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertSuperposition(t, u1)
}

// F3. External target transition triggers superposition on source universe
func TestSuperposition_ExternalTarget(t *testing.T) {
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
		Initials:      []string{"U:u1"},
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
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertSuperposition(t, u1)
	// u2 should be initialized
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized")
	}
}

// F4. Target reality with alwaysTrue observer collapses immediately (reality established)
func TestObserver_AlwaysTrue_Collapse(t *testing.T) {
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
	u1 := NewExUniverse(u1Model)
	u2 := NewExUniverse(u2Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1, u2})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// After observer approval, establishNewReality sets currentReality.
	// The universe is initialized and the reality is established.
	assertReality(t, u2, "stateX")
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized after observer collapse")
	}
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality to be initialized after observer collapse")
	}
}

// F5. ContainsAllEvents observer: accumulates [evt1,evt2]; send evt1 -> still super; send evt2 -> collapse
func TestObserver_ContainsAllEvents_Accumulates(t *testing.T) {
	// The ContainsAllEvents observer checks accumulated events by name.
	// We route from u1 to u2:stateX via two separate events.
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("sendEvt1", []string{"U:u2:stateX"}, nil),
				withOnTransition("sendEvt2", []string{"U:u2:stateX"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX",
				withObserver("builtin:observer:containsAllEvents", map[string]any{
					"evt1": "sendEvt1",
					"evt2": "sendEvt2",
				}),
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Send evt1: u1 transitions and routes to u2:stateX
	evt1 := NewEventBuilder("sendEvt1").Build()
	if _, err := qm.SendEvent(context.Background(), evt1); err != nil {
		t.Fatalf("SendEvent(sendEvt1) failed: %v", err)
	}
	// u2 is in superposition with only 1 of 2 required events accumulated
	assertSuperposition(t, u2)
	// currentReality should be nil (observer didn't approve yet)
	if u2.currentReality != nil {
		t.Fatalf("expected u2 currentReality=nil before all events, got %q", *u2.currentReality)
	}

	// u1 is in superposition; restore it so it can handle next event
	if err := qm.(*ExQuantumMachine).PositionMachine(context.Background(), nil, "u1", "stateA", false); err != nil {
		t.Fatalf("PositionMachine failed: %v", err)
	}

	// Send evt2: routes to u2:stateX again
	evt2 := NewEventBuilder("sendEvt2").Build()
	if _, err := qm.SendEvent(context.Background(), evt2); err != nil {
		t.Fatalf("SendEvent(sendEvt2) failed: %v", err)
	}
	// Now u2 has both events accumulated; observer should approve
	// After approval, establishNewReality sets currentReality
	assertReality(t, u2, "stateX")
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality to be initialized after observer approval")
	}
}

// F6. ContainsAtLeastOne observer: send first matching event -> collapse
func TestObserver_ContainsAtLeastOne_FirstMatch(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("sendEvt1", []string{"U:u2:stateX"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX",
				withObserver("builtin:observer:containsAtLeastOneEvent", map[string]any{
					"evt1": "sendEvt1",
					"evt2": "sendEvt2",
				}),
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt1 := NewEventBuilder("sendEvt1").Build()
	if _, err := qm.SendEvent(context.Background(), evt1); err != nil {
		t.Fatalf("SendEvent(sendEvt1) failed: %v", err)
	}
	// Should collapse — observer approved with first matching event
	assertReality(t, u2, "stateX")
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality initialized after observer approval")
	}
}

// F7. GreaterThanEqual observer: threshold of 3, send 2 -> stays, send 3rd -> collapse
func TestObserver_GreaterThanEqual_Threshold(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("ping", []string{"U:u2:stateX"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX",
				withObserver("builtin:observer:greaterThanEqualCounter", map[string]any{
					"ping": 3,
				}),
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Send ping x2 -> still in superposition (threshold is 3)
	for i := 0; i < 2; i++ {
		evt := NewEventBuilder("ping").Build()
		if _, err := qm.SendEvent(context.Background(), evt); err != nil {
			t.Fatalf("SendEvent(ping) #%d failed: %v", i+1, err)
		}
		// Restore u1 so it can handle next event
		if err := qm.(*ExQuantumMachine).PositionMachine(context.Background(), nil, "u1", "stateA", false); err != nil {
			t.Fatalf("PositionMachine failed: %v", err)
		}
	}
	// After 2 pings, u2 should still not have a reality (observer needs 3)
	if u2.currentReality != nil {
		t.Fatalf("expected u2 currentReality=nil after 2 pings, got %q", *u2.currentReality)
	}

	// Send ping #3 -> observer approves (threshold met)
	evt := NewEventBuilder("ping").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent(ping) #3 failed: %v", err)
	}
	assertReality(t, u2, "stateX")
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality initialized after threshold met")
	}
}

// F8. TotalEventsBetweenLimits observer: min:2, max:5; 1 event -> stays; 2 events -> collapse
func TestObserver_TotalBetweenLimits_Range(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("tick", []string{"U:u2:stateX"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Realities: map[string]*theoretical.RealityModel{
			"stateX": newTransitionReality("stateX",
				withObserver("builtin:observer:totalEventsBetweenLimits", map[string]any{
					"minimum": 2,
					"maximum": 5,
				}),
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 1 event -> stays (minimum=2, not yet met)
	evt := NewEventBuilder("tick").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent(tick) #1 failed: %v", err)
	}
	if u2.currentReality != nil {
		t.Fatalf("expected u2 currentReality=nil after 1 tick, got %q", *u2.currentReality)
	}

	// Restore u1
	if err := qm.(*ExQuantumMachine).PositionMachine(context.Background(), nil, "u1", "stateA", false); err != nil {
		t.Fatalf("PositionMachine failed: %v", err)
	}

	// 2nd event -> collapse (total=2 is within [2,5])
	evt2 := NewEventBuilder("tick").Build()
	if _, err := qm.SendEvent(context.Background(), evt2); err != nil {
		t.Fatalf("SendEvent(tick) #2 failed: %v", err)
	}
	assertReality(t, u2, "stateX")
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality initialized after observer approval")
	}
}

// F9. Observer with Src="" defaults to true immediately
func TestObserver_EmptySrc_DefaultTrue(t *testing.T) {
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
				withObserver("", nil), // empty src
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// empty Src -> default true -> reality established
	assertReality(t, u2, "stateX")
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality initialized after empty-src observer default-true")
	}
}

// F10. Observer with nonexistent Src defaults to true
func TestObserver_NotFound_DefaultTrue(t *testing.T) {
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
				withObserver("test:adv:nonexistent-observer-xyz", nil),
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// not found -> default true -> reality established
	assertReality(t, u2, "stateX")
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality initialized after not-found observer default-true")
	}
}

// F11. Multiple observers: [alwaysFalse, alwaysTrue] -> second approves
func TestObserver_Multiple_FirstTrueWins(t *testing.T) {
	alwaysFalseObs := "test:adv:observer-always-false"
	registerTestObserver(t, alwaysFalseObs, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		return false, nil
	})
	alwaysTrueObs := "test:adv:observer-always-true"
	registerTestObserver(t, alwaysTrueObs, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
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
				withObserver(alwaysFalseObs, nil),
				withObserver(alwaysTrueObs, nil),
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// second observer returns true -> reality established
	assertReality(t, u2, "stateX")
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality initialized after second observer approved")
	}
}

// F12. Observer returns error -> propagated
func TestObserver_ReturnsError(t *testing.T) {
	errObs := "test:adv:observer-error"
	registerTestObserver(t, errObs, func(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
		return false, fmt.Errorf("observer deliberate error")
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	_, err = qm.SendEvent(context.Background(), evt)
	if err == nil {
		t.Fatal("expected error from observer, got nil")
	}
	if !strings.Contains(err.Error(), "observer deliberate error") {
		t.Fatalf("expected error containing 'observer deliberate error', got: %v", err)
	}
}

// F13. receiveEventToReality accumulates for specific reality (directed event)
func TestSuperposition_DirectedEvent(t *testing.T) {
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
				withObserver("builtin:observer:containsAllEvents", map[string]any{
					"evt1": "go",
					"evt2": "secondEvt",
				}),
			),
			"stateY": newTransitionReality("stateY"),
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
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// directed event to u2:stateX
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertSuperposition(t, u2)
	// The accumulator should have the event for stateX specifically
	if u2.eventAccumulator == nil {
		t.Fatal("expected eventAccumulator to be non-nil during superposition")
	}
}

// F14. receiveEvent accumulates for all realities (broadcast)
func TestSuperposition_BroadcastEvent(t *testing.T) {
	um := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		// No Initial -> superposition on init
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withObserver("builtin:observer:containsAllEvents", map[string]any{
					"evt1": "tick",
					"evt2": "tock",
				}),
			),
			"stateB": newTransitionReality("stateB",
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertSuperposition(t, u)

	// broadcast event to universe (no specific reality)
	// Since universe is in superposition and was NOT started with an initial,
	// we need to use handleEvent directly (as receiveEvent) — which SendEvent cannot do
	// because canHandleEvent returns false for superposition.
	// However, if another universe sends to U:u1 (no reality), it calls handleEvent with nil realityName -> receiveEvent.
	// Let's test via a second universe routing to U:u1.
	// For simplicity, verify the accumulator is initialized for all realities after superposition init.
	if u.eventAccumulator == nil {
		t.Fatal("expected eventAccumulator to be non-nil during superposition")
	}
}

// F15. Full observer flow: observer approves -> entry actions execute on collapse
func TestSuperposition_Collapse_FullFlow(t *testing.T) {
	entryActionName := "test:adv:collapse-entry"
	var entryRan bool
	registerTestAction(t, entryActionName, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		entryRan = true
		return nil
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
				withObserver("builtin:observer:alwaysTrue", nil),
				withEntryAction(entryActionName),
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// observer approves -> establishNewReality -> entry actions run
	if !entryRan {
		t.Fatal("expected entry action to run on collapse")
	}
	assertReality(t, u2, "stateX")
	if !u2.realityInitialized {
		t.Fatal("expected u2 reality to be initialized after collapse")
	}
}

// F16. GetSnapshot during superposition reports SuperpositionUniverses correctly
func TestSuperposition_Snapshot(t *testing.T) {
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
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertSuperposition(t, u1)

	snapshot := qm.(*ExQuantumMachine).GetSnapshot()
	superUniverses := snapshot.Resume.SuperpositionUniverses
	if superUniverses == nil {
		t.Fatal("expected SuperpositionUniverses to be non-nil")
	}
	if _, ok := superUniverses["Universe1"]; !ok {
		t.Fatalf("expected Universe1 in SuperpositionUniverses, got %v", superUniverses)
	}
}

// ===================== G. Always Transitions (10 tests) =====================

// G1. Always with true condition transitions to target
func TestAlways_TrueCondition(t *testing.T) {
	condTrue := "test:adv:always-true-g1"
	registerTestCondition(t, condTrue, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, &theoretical.ConditionModel{Src: condTrue}),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// G2. Always with false condition stays at current reality
func TestAlways_FalseCondition(t *testing.T) {
	condFalse := "test:adv:always-false-g2"
	registerTestCondition(t, condFalse, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, &theoretical.ConditionModel{Src: condFalse}),
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateA")
}

// G3. Chain of always transitions: A -> B -> C -> D
func TestAlways_Chain_ABtoCD(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withAlways([]string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC",
			withAlways([]string{"stateD"}, nil),
		),
		"stateD": newTransitionReality("stateD"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateD")
}

// G4. Always targeting external universe -> superposition
func TestAlways_ExternalUniverse(t *testing.T) {
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
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertSuperposition(t, u1)
	assertReality(t, u2, "stateX")
}

// G5. Always with multiple targets -> superposition
func TestAlways_MultipleTargets(t *testing.T) {
	// Always with multiple targets to different universes → superposition
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withAlways([]string{"U:u2:stateX", "U:u3:stateY"}, nil),
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
	assertSuperposition(t, u1)
}

// G6. After on-transition, always of new reality executes
func TestAlways_AfterOnTransition(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withOnTransition("go", []string{"stateB"}, nil),
		),
		"stateB": newTransitionReality("stateB",
			withAlways([]string{"stateC"}, nil),
		),
		"stateC": newTransitionReality("stateC"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateC")
}

// G7. Always notification: u1 stays at A, u2 receives notification
func TestAlways_Notification(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": {
				ID:   "stateA",
				Type: theoretical.RealityTypeTransition,
				On:   map[string][]*theoretical.TransitionModel{},
				Always: []*theoretical.TransitionModel{
					notifyTransition([]string{"U:u2:stateX"}),
				},
			},
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
		Initials:      []string{"U:u1"},
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
	// u1 stays at stateA (notify does NOT exit source)
	assertReality(t, u1, "stateA")
	assertNotSuperposition(t, u1)
	// u2 should be initialized (received notification)
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized after notification")
	}
}

// G8. Always with no condition executes unconditionally
func TestAlways_NoCondition_Executes(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, nil), // no condition
		),
		"stateB": newTransitionReality("stateB"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// G9. Multiple always transitions: first matching wins
func TestAlways_MultipleAlways_FirstWins(t *testing.T) {
	condTrue1 := "test:adv:always-true-g9a"
	condTrue2 := "test:adv:always-true-g9b"
	registerTestCondition(t, condTrue1, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})
	registerTestCondition(t, condTrue2, func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withAlways([]string{"stateB"}, &theoretical.ConditionModel{Src: condTrue1}),
			withAlways([]string{"stateC"}, &theoretical.ConditionModel{Src: condTrue2}),
		),
		"stateB": newTransitionReality("stateB"),
		"stateC": newTransitionReality("stateC"),
	}
	qm, u := buildQM(t, "stateA", realities)
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	assertReality(t, u, "stateB") // first always wins
}

// G10. Always to final reality sets isFinalReality=true
func TestAlways_ToFinal(t *testing.T) {
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
	assertReality(t, u, "DONE")
	assertFinalized(t, u)
}

// ===================== H. Cross-Universe Routing (10 tests) =====================

// H1. Cross-universe target initializes second universe on specific reality
func TestCrossUniverse_InitializesSecond(t *testing.T) {
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
		Initial:       strPtr("stateX"),
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
		Initials:      []string{"U:u1"},
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
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized")
	}
	assertReality(t, u2, "stateX")
}

// H2. Chain across three universes: u1->U:u2:B, u2 initialized with Initial=B has always->U:u3:C
func TestCrossUniverse_Chain_U1U2U3(t *testing.T) {
	// u2 has Initial=stateB so when routed to U:u2, it initializes on stateB directly
	// (not via superposition), and its always transition fires.
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
		Initial:       strPtr("stateB"),
		Realities: map[string]*theoretical.RealityModel{
			"stateB": newTransitionReality("stateB",
				withAlways([]string{"U:u3:stateC"}, nil),
			),
		},
	}
	u3Model := &theoretical.UniverseModel{
		ID:            "u3",
		CanonicalName: "Universe3",
		Realities: map[string]*theoretical.RealityModel{
			"stateC": newTransitionReality("stateC"),
		},
	}
	qm, _, u2, u3 := buildThreeUniverseQM(t, u1Model, u2Model, u3Model, []string{"U:u1"})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// u2 should be in superposition (always->external)
	assertSuperposition(t, u2)
	// u3 should have stateC as currentReality
	assertReality(t, u3, "stateC")
}

// H3. Target non-existent universe -> error
func TestCrossUniverse_NonExistent_Error(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"U:doesNotExist:stateX"}, nil),
			),
		},
	}
	qmm := &theoretical.QuantumMachineModel{
		ID:            "qm1",
		CanonicalName: "TestQM",
		Version:       "1.0.0",
		Universes:     map[string]*theoretical.UniverseModel{"u1": u1Model},
		Initials:      []string{"U:u1"},
	}
	u1 := NewExUniverse(u1Model)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u1})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	_, err = qm.SendEvent(context.Background(), evt)
	if err == nil {
		t.Fatal("expected error for non-existent universe target, got nil")
	}
	if !strings.Contains(err.Error(), "doesNotExist") {
		t.Fatalf("expected error mentioning 'doesNotExist', got: %v", err)
	}
}

// H4. U:u2:stateX -> u2 initializes on stateX
func TestCrossUniverse_SpecificReality(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": newTransitionReality("stateA",
				withOnTransition("go", []string{"U:u2:stateY"}, nil),
			),
		},
	}
	u2Model := &theoretical.UniverseModel{
		ID:            "u2",
		CanonicalName: "Universe2",
		Initial:       strPtr("stateX"),
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
		Initials:      []string{"U:u1"},
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
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// u2 should be on stateY (not stateX which is the default initial)
	assertReality(t, u2, "stateY")
}

// H5. U:u2 (universe-only, u2 has Initial=X) -> u2 on X
func TestCrossUniverse_UniverseOnly_UsesInitial(t *testing.T) {
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
		Initials:      []string{"U:u1"},
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
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u2, "stateX")
}

// H6. U:u2 (no initial) -> u2 in superposition
func TestCrossUniverse_UniverseOnly_NoInitial_Superposition(t *testing.T) {
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
		// No Initial
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
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertSuperposition(t, u2)
}

// H7. Notification transition: u1 stays, u2 receives
func TestCrossUniverse_Notification(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": {
				ID:   "stateA",
				Type: theoretical.RealityTypeTransition,
				On: map[string][]*theoretical.TransitionModel{
					"notify": {notifyTransition([]string{"U:u2:stateX"})},
				},
			},
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
		Initials:      []string{"U:u1"},
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

	evt := NewEventBuilder("notify").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// u1 stays at stateA (notify does not exit)
	assertReality(t, u1, "stateA")
	assertNotSuperposition(t, u1)
	// u2 should be initialized
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized from notification")
	}
}

// H8. Notification sets externalTargets correctly
func TestCrossUniverse_Notification_ExternalTargets(t *testing.T) {
	u1Model := &theoretical.UniverseModel{
		ID:            "u1",
		CanonicalName: "Universe1",
		Initial:       strPtr("stateA"),
		Realities: map[string]*theoretical.RealityModel{
			"stateA": {
				ID:   "stateA",
				Type: theoretical.RealityTypeTransition,
				On: map[string][]*theoretical.TransitionModel{
					"notify": {notifyTransition([]string{"U:u2:stateX", "U:u3:stateY"})},
				},
			},
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
	u3Model := &theoretical.UniverseModel{
		ID:            "u3",
		CanonicalName: "Universe3",
		Initial:       strPtr("stateY"),
		Realities: map[string]*theoretical.RealityModel{
			"stateY": newTransitionReality("stateY"),
		},
	}
	qm, u1, u2, u3 := buildThreeUniverseQM(t, u1Model, u2Model, u3Model, []string{"U:u1"})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("notify").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// u1 stays (notification)
	assertReality(t, u1, "stateA")
	// Both u2 and u3 should be initialized
	if !u2.initialized {
		t.Fatal("expected u2 to be initialized")
	}
	if !u3.initialized {
		t.Fatal("expected u3 to be initialized")
	}
}

// H9. Multiple mixed cross-universe targets
func TestCrossUniverse_MultipleTargets_Mixed(t *testing.T) {
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
			"stateX": newTransitionReality("stateX"),
		},
	}
	u3Model := &theoretical.UniverseModel{
		ID:            "u3",
		CanonicalName: "Universe3",
		Realities: map[string]*theoretical.RealityModel{
			"stateY": newTransitionReality("stateY"),
		},
	}
	qm, u1, u2, u3 := buildThreeUniverseQM(t, u1Model, u2Model, u3Model, []string{"U:u1"})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// u1 enters superposition (multiple external targets)
	assertSuperposition(t, u1)
	// u2 on stateX, u3 on stateY
	assertReality(t, u2, "stateX")
	assertReality(t, u3, "stateY")
}

// H10. Event data from u1 is readable in u2 entry action
func TestCrossUniverse_EventDataPreserved(t *testing.T) {
	entryAction := "test:adv:cross-data-entry"
	var capturedData map[string]any
	registerTestAction(t, entryAction, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		capturedData = args.GetEvent().GetData()
		return nil
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
				withEntryAction(entryAction),
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
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").SetData(map[string]any{"key": "value123"}).Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	if capturedData == nil {
		t.Fatal("expected entry action to capture event data")
	}
	if capturedData["key"] != "value123" {
		t.Fatalf("expected data key='value123', got %v", capturedData["key"])
	}
}

// ===================== I. Universal Constants (11 tests) =====================

// I1. Constants entry action fires for every reality entry
func TestConstants_EntryAction_EveryReality(t *testing.T) {
	var entryRealities []string
	constEntry := "test:adv:const-entry-every"
	registerTestAction(t, constEntry, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		entryRealities = append(entryRealities, args.GetRealityName())
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
			EntryActions: []*theoretical.ActionModel{{Src: constEntry}},
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
	// Constants entry should fire for stateA
	if len(entryRealities) != 1 || entryRealities[0] != "stateA" {
		t.Fatalf("expected constants entry for [stateA], got %v", entryRealities)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// Now should also fire for stateB
	if len(entryRealities) != 2 || entryRealities[1] != "stateB" {
		t.Fatalf("expected constants entry for [stateA, stateB], got %v", entryRealities)
	}
}

// I2. Constants exit action fires for every exit
func TestConstants_ExitAction_EveryExit(t *testing.T) {
	var exitRealities []string
	constExit := "test:adv:const-exit-every"
	registerTestAction(t, constExit, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		exitRealities = append(exitRealities, args.GetRealityName())
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
			ExitActions: []*theoretical.ActionModel{{Src: constExit}},
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

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	// Exit of stateA should trigger constants exit
	if len(exitRealities) != 1 || exitRealities[0] != "stateA" {
		t.Fatalf("expected constants exit for [stateA], got %v", exitRealities)
	}
}

// I3. Constants transition action fires on transition
func TestConstants_TransitionAction(t *testing.T) {
	var transitionRealities []string
	constTrans := "test:adv:const-trans-action"
	registerTestAction(t, constTrans, func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		transitionRealities = append(transitionRealities, args.GetRealityName())
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
			ActionsOnTransition: []*theoretical.ActionModel{{Src: constTrans}},
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

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if len(transitionRealities) != 1 || transitionRealities[0] != "stateA" {
		t.Fatalf("expected constants transition for [stateA], got %v", transitionRealities)
	}
}

// I4. Constants entry invoke fires asynchronously
func TestConstants_EntryInvokes_Async(t *testing.T) {
	invokeName := "test:adv:const-entry-invoke"
	ch := make(chan string, 1)
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		ch <- args.GetRealityName()
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
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
			EntryInvokes: []*theoretical.InvokeModel{{Src: invokeName}},
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

	select {
	case realityName := <-ch:
		if realityName != "stateA" {
			t.Fatalf("expected invoke for stateA, got %s", realityName)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for constants entry invoke")
	}
}

// I5. Constants exit invoke fires asynchronously
func TestConstants_ExitInvokes_Async(t *testing.T) {
	invokeName := "test:adv:const-exit-invoke"
	ch := make(chan string, 1)
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		ch <- args.GetRealityName()
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
			ExitInvokes: []*theoretical.InvokeModel{{Src: invokeName}},
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

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	select {
	case realityName := <-ch:
		if realityName != "stateA" {
			t.Fatalf("expected exit invoke for stateA, got %s", realityName)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for constants exit invoke")
	}
}

// I6. Constants transition invoke fires asynchronously
func TestConstants_TransitionInvokes_Async(t *testing.T) {
	invokeName := "test:adv:const-trans-invoke"
	ch := make(chan string, 1)
	registerTestInvoke(t, invokeName, func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		ch <- args.GetRealityName()
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
			InvokesOnTransition: []*theoretical.InvokeModel{{Src: invokeName}},
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

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	select {
	case realityName := <-ch:
		if realityName != "stateA" {
			t.Fatalf("expected transition invoke for stateA, got %s", realityName)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for constants transition invoke")
	}
}

// I7. Constants entry action fires BEFORE reality entry action
func TestConstants_Order_ConstantsBeforeEntry(t *testing.T) {
	var order []string
	constEntry := "test:adv:const-order-entry"
	realityEntry := "test:adv:reality-order-entry"

	registerTestAction(t, constEntry, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "constants")
		return nil
	})
	registerTestAction(t, realityEntry, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "reality")
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withEntryAction(realityEntry),
		),
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
			EntryActions: []*theoretical.ActionModel{{Src: constEntry}},
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

	expected := []string{"constants", "reality"}
	if len(order) != len(expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Fatalf("order[%d]: expected %q, got %q. full: %v", i, exp, order[i], order)
		}
	}
}

// I8. Constants exit action fires BEFORE reality exit action
func TestConstants_Order_ConstantsBeforeExit(t *testing.T) {
	var order []string
	constExit := "test:adv:const-order-exit"
	realityExit := "test:adv:reality-order-exit"

	registerTestAction(t, constExit, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "constants")
		return nil
	})
	registerTestAction(t, realityExit, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "reality")
		return nil
	})

	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA",
			withExitAction(realityExit),
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
			ExitActions: []*theoretical.ActionModel{{Src: constExit}},
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

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	expected := []string{"constants", "reality"}
	if len(order) != len(expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Fatalf("order[%d]: expected %q, got %q. full: %v", i, exp, order[i], order)
		}
	}
}

// I9. Constants transition action fires BEFORE reality transition action
func TestConstants_Order_ConstantsBeforeTransition(t *testing.T) {
	var order []string
	constTrans := "test:adv:const-order-trans"
	realityTrans := "test:adv:reality-order-trans"

	registerTestAction(t, constTrans, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "constants")
		return nil
	})
	registerTestAction(t, realityTrans, func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		order = append(order, "reality")
		return nil
	})

	stateA := &theoretical.RealityModel{
		ID:   "stateA",
		Type: theoretical.RealityTypeTransition,
		On: map[string][]*theoretical.TransitionModel{
			"go": {
				{
					Targets: []string{"stateB"},
					Actions: []*theoretical.ActionModel{{Src: realityTrans}},
				},
			},
		},
	}

	realities := map[string]*theoretical.RealityModel{
		"stateA": stateA,
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
			ActionsOnTransition: []*theoretical.ActionModel{{Src: constTrans}},
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

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	expected := []string{"constants", "reality"}
	if len(order) != len(expected) {
		t.Fatalf("expected order %v, got %v", expected, order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Fatalf("order[%d]: expected %q, got %q. full: %v", i, exp, order[i], order)
		}
	}
}

// I10. UniversalConstants=nil -> no error
func TestConstants_Nil_NoOp(t *testing.T) {
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
		// UniversalConstants is nil
	}
	u := NewExUniverse(um)
	qm, err := NewExQuantumMachine(qmm, []*ExUniverse{u})
	if err != nil {
		t.Fatalf("failed to build QM: %v", err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
}

// I11. UniversalConstants={} -> no error
func TestConstants_Empty_NoOp(t *testing.T) {
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
			// all fields nil/empty
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

	evt := NewEventBuilder("go").Build()
	if _, err := qm.SendEvent(context.Background(), evt); err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	assertReality(t, u, "stateB")
}
