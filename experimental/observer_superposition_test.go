package experimental

import (
	"context"
	"testing"

	"github.com/rendis/statepro/v3/theoretical"
)

// TestObserverCollapseSuperposition reproduces the bug where a universe with an
// observer on its initial reality never collapses from superposition even when
// the observer approves. The root cause is that establishNewReality checks
// inSuperposition after executeOnEntry, but inSuperposition was already true
// from the initOnSuperposition call that preceded it.
//
// Setup:
//   - Universe A: initial reality "START" with an Always transition to U:uB:TARGET
//   - Universe B: reality "TARGET" with a builtin:observer:alwaysTrue observer
//
// Expected: after Init, Universe B should be initialized on "TARGET", NOT in superposition.
func TestObserverCollapseSuperposition(t *testing.T) {
	// Universe A: START → always → U:uB:TARGET
	startReality := &theoretical.RealityModel{
		ID:   "START",
		Type: theoretical.RealityTypeTransition,
		Always: []*theoretical.TransitionModel{
			{Targets: []string{"U:uB:TARGET"}},
		},
	}

	initialA := "START"
	universeA := &theoretical.UniverseModel{
		ID:            "uA",
		CanonicalName: "UniverseA",
		Initial:       &initialA,
		Realities: map[string]*theoretical.RealityModel{
			"START": startReality,
		},
	}

	// Universe B: TARGET (with AlwaysTrue observer)
	targetReality := &theoretical.RealityModel{
		ID:   "TARGET",
		Type: theoretical.RealityTypeTransition,
		On:   map[string][]*theoretical.TransitionModel{},
		Observers: []*theoretical.ObserverModel{
			{Src: "builtin:observer:alwaysTrue"},
		},
	}

	universeB := &theoretical.UniverseModel{
		ID:            "uB",
		CanonicalName: "UniverseB",
		Realities: map[string]*theoretical.RealityModel{
			"TARGET": targetReality,
		},
		// No Initial — universe will enter superposition when targeted
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm-observer-test",
		CanonicalName: "ObserverTestQM",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"uA": universeA,
			"uB": universeB,
		},
		Initials: []string{"U:uA"},
	}

	exA := NewExUniverse(universeA)
	exB := NewExUniverse(universeB)
	qm, err := NewExQuantumMachine(qmModel, []*ExUniverse{exA, exB})
	if err != nil {
		t.Fatalf("NewExQuantumMachine: %v", err)
	}

	// Init: A starts on START → Always fires → targets U:uB:TARGET
	// B enters superposition → AlwaysTrue observer approves → should establish TARGET
	if err = qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Assert Universe B is NOT in superposition
	if exB.inSuperposition {
		t.Fatalf("Universe B should NOT be in superposition (observer approved), but inSuperposition=%t, currentReality=%v",
			exB.inSuperposition, exB.currentReality)
	}

	// Assert Universe B is initialized on TARGET
	if !exB.initialized {
		t.Fatal("Universe B should be initialized")
	}
	if exB.currentReality == nil || *exB.currentReality != "TARGET" {
		t.Fatalf("Universe B should be on reality 'TARGET', got %v", exB.currentReality)
	}
	if !exB.realityInitialized {
		t.Fatal("Universe B reality should be initialized")
	}
}

// TestObserverCollapseSuperposition_SendEvent tests the same scenario but
// triggered by a SendEvent instead of Init. Universe B enters superposition
// from an event-driven transition, observer approves, and should collapse.
func TestObserverCollapseSuperposition_SendEvent(t *testing.T) {
	// Universe A: IDLE → on "go" → U:uB:WAITING
	idleReality := &theoretical.RealityModel{
		ID:   "IDLE",
		Type: theoretical.RealityTypeTransition,
		On: map[string][]*theoretical.TransitionModel{
			"go": {{Targets: []string{"U:uB:WAITING"}}},
		},
	}

	initialA := "IDLE"
	universeA := &theoretical.UniverseModel{
		ID:            "uA",
		CanonicalName: "UniverseA",
		Initial:       &initialA,
		Realities: map[string]*theoretical.RealityModel{
			"IDLE": idleReality,
		},
	}

	// Universe B: WAITING (with AlwaysTrue observer)
	waitingReality := &theoretical.RealityModel{
		ID:   "WAITING",
		Type: theoretical.RealityTypeTransition,
		On:   map[string][]*theoretical.TransitionModel{},
		Observers: []*theoretical.ObserverModel{
			{Src: "builtin:observer:alwaysTrue"},
		},
	}

	universeB := &theoretical.UniverseModel{
		ID:            "uB",
		CanonicalName: "UniverseB",
		Realities: map[string]*theoretical.RealityModel{
			"WAITING": waitingReality,
		},
	}

	qmModel := &theoretical.QuantumMachineModel{
		ID:            "qm-observer-send",
		CanonicalName: "ObserverSendQM",
		Version:       "1.0.0",
		Universes: map[string]*theoretical.UniverseModel{
			"uA": universeA,
			"uB": universeB,
		},
		Initials: []string{"U:uA"},
	}

	exA := NewExUniverse(universeA)
	exB := NewExUniverse(universeB)
	qm, err := NewExQuantumMachine(qmModel, []*ExUniverse{exA, exB})
	if err != nil {
		t.Fatalf("NewExQuantumMachine: %v", err)
	}

	if err = qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// A is on IDLE, B is not initialized
	if exB.initialized {
		t.Fatal("Universe B should not be initialized before event")
	}

	// Send "go" → A transitions to U:uB:WAITING → B superposition → observer approves
	evt := NewEventBuilder("go").Build()
	_, err = qm.SendEvent(context.Background(), evt)
	if err != nil {
		t.Fatalf("SendEvent: %v", err)
	}

	// Assert Universe B collapsed from superposition
	if exB.inSuperposition {
		t.Fatalf("Universe B should NOT be in superposition after observer approved, but inSuperposition=%t, currentReality=%v",
			exB.inSuperposition, exB.currentReality)
	}
	if exB.currentReality == nil || *exB.currentReality != "WAITING" {
		t.Fatalf("Universe B should be on reality 'WAITING', got %v", exB.currentReality)
	}
	if !exB.initialized {
		t.Fatal("Universe B should be initialized")
	}
}
