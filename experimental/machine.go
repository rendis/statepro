package experimental

import (
	"context"
	"fmt"
	"github.com/rendis/devtoolkit"
	"github.com/rendis/statepro/v3/theoretical"
	"sync"
)

type initFunc func(context.Context, *ExUniverse, []string) ([]string, Event, error)

var qmInitFunctions = map[refType]initFunc{
	RefTypeUniverse: func(ctx context.Context, u *ExUniverse, _ []string) ([]string, Event, error) {
		return u.Start(ctx)
	},
	RefTypeUniverseReality: func(ctx context.Context, u *ExUniverse, p []string) ([]string, Event, error) {
		return u.StartOnReality(ctx, p[1])
	},
}

type ExQuantumMachineSnapshot map[string]ExUniverseSnapshot

func NewExQuantumMachine(qmm *theoretical.QuantumMachineModel, qml QuantumMachineLaws, universes []*ExUniverse) (*ExQuantumMachine, error) {
	qm := &ExQuantumMachine{
		model:     qmm,
		laws:      qml,
		universes: map[string]*ExUniverse{},
	}

	for _, u := range universes {
		if u == nil {
			continue
		}

		// check if universe already exists
		if _, ok := qm.universes[u.id]; ok {
			return nil, fmt.Errorf("universe '%s' already exists", u.id)
		}

		qm.universes[u.id] = u
	}

	return qm, nil
}

type ExQuantumMachine struct {
	// model is the quantum machine model
	model *theoretical.QuantumMachineModel

	// machineContext is the quantum machine context
	machineContext any

	// laws of the quantum machine
	laws QuantumMachineLaws

	// universes is the map of the quantum machine universes
	// key: theoretical.UniverseModel.ID@theoretical.UniverseModel.Version
	universes map[string]*ExUniverse

	// quantumMachineMtx is the mutex for the quantum machine
	quantumMachineMtx sync.Mutex
}

func (qm *ExQuantumMachine) Init(ctx context.Context, machineContext any) error {
	qm.machineContext = machineContext

	var pairs []devtoolkit.Pair[Event, []string]

	for _, initial := range qm.model.Initials {
		refT, parts, err := getReferenceType(initial)
		if err != nil {
			return err
		}

		// check if universe exists
		universe, ok := qm.universes[parts[0]]
		if !ok {
			return fmt.Errorf("universe '%s' not found on initial universes", parts[0])
		}

		// get init function
		f, ok := qmInitFunctions[refT]
		if !ok {
			return fmt.Errorf("invalid ref type '%d'", refT)
		}

		// execute init function
		transitions, evt, err := f(ctx, universe, parts)
		if err != nil {
			return err
		}

		pair := devtoolkit.NewPair[Event, []string](evt, transitions)
		pairs = append(pairs, pair)
	}

	return qm.executeTargetPairs(ctx, pairs)
}

func (qm *ExQuantumMachine) SendEvent(ctx context.Context, event Event) error {
	var pairs []devtoolkit.Pair[Event, []string]

	for _, u := range qm.getActiveUniverses() {
		externalTargets, _, err := u.HandleExternalEvent(ctx, nil, event)
		if err != nil {
			return err
		}

		if len(externalTargets) == 0 {
			continue
		}

		pair := devtoolkit.NewPair[Event, []string](event, externalTargets)
		pairs = append(pairs, pair)
	}

	return qm.executeTargetPairs(ctx, pairs)
}

func (qm *ExQuantumMachine) GetSnapshot() ExQuantumMachineSnapshot {
	var snapshot = make(ExQuantumMachineSnapshot)
	for _, u := range qm.universes {
		snapshot[u.id] = u.GetSnapshot()
	}
	return snapshot
}

func (qm *ExQuantumMachine) LoadSnapshot(snapshot ExQuantumMachineSnapshot) error {
	for _, u := range qm.universes {
		if _, ok := snapshot[u.id]; !ok {
			continue
		}

		err := u.LoadSnapshot(snapshot[u.id])
		if err != nil {
			return err
		}
	}
	return nil
}

func (qm *ExQuantumMachine) ModelToMap() (map[string]any, error) {
	return qm.model.ToMap()
}

//----------------------------------

func (qm *ExQuantumMachine) getActiveUniverses() []*ExUniverse {
	var activeUniverses []*ExUniverse
	for _, u := range qm.universes {
		if u.IsActive() {
			activeUniverses = append(activeUniverses, u)
		}
	}
	return activeUniverses
}

func (qm *ExQuantumMachine) executeTargetPairs(ctx context.Context, pairs []devtoolkit.Pair[Event, []string]) error {

	// while there are pairs to execute
	for len(pairs) > 0 {
		pair := pairs[0]
		pairs = pairs[1:]

		// execute transition
		e, t := pair.GetAll()
		newTargets, err := qm.executeTransitions(ctx, e, t)
		if err != nil {
			return err
		}

		// add new targets to the queue
		newPair := devtoolkit.NewPair[Event, []string](pair.GetFirst(), newTargets)
		pairs = append(pairs, newPair)
	}

	return nil
}

func (qm *ExQuantumMachine) executeTransitions(ctx context.Context, evt Event, targets []string) ([]string, error) {

	var newTargets []string

	for _, target := range targets {
		refT, parts, _ := getReferenceType(target)
		exUniverse := qm.universes[parts[0]]

		var realityName *string = nil
		if refT == RefTypeUniverseReality {
			realityName = &parts[1]
		}

		newTransitions, _, err := exUniverse.HandleExternalEvent(ctx, realityName, evt)
		if err != nil {
			return nil, err
		}

		newTargets = append(newTargets, newTransitions...)
	}

	return newTargets, nil
}
