package experimental

import (
	"context"
	"fmt"
	"github.com/rendis/devtoolkit"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
	"sync"
)

type initFunc func(context.Context, any, *ExUniverse, []string) ([]string, instrumentation.Event, error)

var qmInitFunctions = map[refType]initFunc{
	RefTypeUniverse: func(ctx context.Context, uCtx any, u *ExUniverse, _ []string) ([]string, instrumentation.Event, error) {
		return u.start(ctx, uCtx)
	},
	RefTypeUniverseReality: func(ctx context.Context, uCtx any, u *ExUniverse, p []string) ([]string, instrumentation.Event, error) {
		return u.startOnReality(ctx, p[1], uCtx)
	},
}

func NewExQuantumMachine(qmm *theoretical.QuantumMachineModel, laws instrumentation.QuantumMachineLaws, universes []*ExUniverse) (instrumentation.QuantumMachine, error) {

	qm := &ExQuantumMachine{
		model:             qmm,
		observerExecutor:  getUniverseObserverExecutor(laws),
		actionExecutor:    getUniverseActionExecutor(laws),
		invokeExecutor:    getUniverseInvokeExecutor(laws),
		conditionExecutor: getUniverseConditionExecutor(laws),
		universes:         map[string]*ExUniverse{},
	}

	for _, u := range universes {
		if u == nil {
			continue
		}

		// check if universe already exists
		if _, ok := qm.universes[u.model.ID]; ok {
			return nil, fmt.Errorf("universe '%s' already exists", u.model.ID)
		}

		u.constantsLawsExecutor = qm
		qm.universes[u.model.ID] = u
	}

	return qm, nil
}

type ExQuantumMachine struct {
	// model is the quantum machine model
	model *theoretical.QuantumMachineModel

	// machineContext is the quantum machine context
	machineContext any

	// laws executors
	observerExecutor  instrumentation.ObserverExecutor
	actionExecutor    instrumentation.ActionExecutor
	invokeExecutor    instrumentation.InvokeExecutor
	conditionExecutor instrumentation.ConditionExecutor

	// universes is the map of the quantum machine universes
	// key: theoretical.UniverseModel.ID
	universes map[string]*ExUniverse

	// quantumMachineMtx is the mutex for the quantum machine
	quantumMachineMtx sync.Mutex
}

//--------- QuantumMachine interface implementation ---------

func (qm *ExQuantumMachine) Init(ctx context.Context, machineContext any) error {
	qm.quantumMachineMtx.Lock()
	defer qm.quantumMachineMtx.Unlock()

	qm.machineContext = machineContext

	var pairs []devtoolkit.Pair[instrumentation.Event, []string]

	for _, ref := range qm.model.Initials {
		// get reference type and parts
		refT, parts, err := processReference(ref)
		if err != nil {
			return err
		}

		// check if universe exists
		universe, ok := qm.universes[parts[0]]
		if !ok {
			return fmt.Errorf("universe '%s' not found on ref universes", parts[0])
		}

		// get init function
		initFn, ok := qmInitFunctions[refT]
		if !ok {
			return fmt.Errorf("invalid ref type '%d'", refT)
		}

		// execute init function
		transitions, evt, err := initFn(ctx, machineContext, universe, parts)
		if err != nil {
			return err
		}

		pair := devtoolkit.NewPair[instrumentation.Event, []string](evt, transitions)
		pairs = append(pairs, pair)
	}

	return qm.executeTargetPairs(ctx, pairs)
}

func (qm *ExQuantumMachine) SendEvent(ctx context.Context, event instrumentation.Event) error {
	qm.quantumMachineMtx.Lock()
	defer qm.quantumMachineMtx.Unlock()

	var pairs []devtoolkit.Pair[instrumentation.Event, []string]

	for _, u := range qm.getActiveUniverses() {
		externalTargets, _, err := u.handleEvent(ctx, nil, event, qm.machineContext)
		if err != nil {
			return err
		}

		if len(externalTargets) == 0 {
			continue
		}

		pair := devtoolkit.NewPair[instrumentation.Event, []string](event, externalTargets)
		pairs = append(pairs, pair)
	}

	return qm.executeTargetPairs(ctx, pairs)
}

func (qm *ExQuantumMachine) LazySendEvent(ctx context.Context, event instrumentation.Event) error {
	qm.quantumMachineMtx.Lock()
	defer qm.quantumMachineMtx.Unlock()

	var pairs []devtoolkit.Pair[instrumentation.Event, []string]

	for _, u := range qm.getLazyActiveUniverses(event) {
		externalTargets, _, err := u.handleEvent(ctx, nil, event, qm.machineContext)
		if err != nil {
			return err
		}

		if len(externalTargets) == 0 {
			continue
		}

		pair := devtoolkit.NewPair[instrumentation.Event, []string](event, externalTargets)
		pairs = append(pairs, pair)
	}

	return qm.executeTargetPairs(ctx, pairs)
}

func (qm *ExQuantumMachine) LoadSnapshot(snapshot *instrumentation.MachineSnapshot, machineContext any) error {
	qm.quantumMachineMtx.Lock()
	defer qm.quantumMachineMtx.Unlock()

	if snapshot == nil {
		return nil
	}

	for _, u := range qm.universes {
		universeSnapshot, ok := snapshot.Snapshots[u.model.ID]

		if !ok {
			continue
		}

		err := u.loadSnapshot(universeSnapshot)
		if err != nil {
			return err
		}
	}

	qm.machineContext = machineContext
	return nil
}

func (qm *ExQuantumMachine) GetSnapshot() *instrumentation.MachineSnapshot {
	var machineSnapshot = &instrumentation.MachineSnapshot{}

	for _, u := range qm.universes {
		universeSnapshot := u.getSnapshot()

		// add snapshot
		machineSnapshot.AddUniverseSnapshot(u.model.ID, universeSnapshot)

		// resume only if initialized
		if !u.initialized {
			continue
		}

		// active universe resume
		if !u.inSuperposition && !u.isFinalReality {
			machineSnapshot.AddActiveUniverse(u.model.CanonicalName, *u.currentReality)
		}

		// finalized universe resume
		if !u.inSuperposition && u.isFinalReality {
			machineSnapshot.AddFinalizedUniverse(u.model.CanonicalName, *u.currentReality)
		}

		// superposition universe resume
		if u.inSuperposition {
			machineSnapshot.AddSuperpositionUniverse(u.model.CanonicalName, *u.realityBeforeSuperposition)
		}
	}

	return machineSnapshot
}

//--------- constantsLawsExecutor interface implementation ---------

func (qm *ExQuantumMachine) ExecuteEntryInvokes(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.EntryInvokes) == 0 {
		return
	}

	if qm.invokeExecutor == nil {
		return
	}

	for _, invoke := range qm.model.UniversalConstants.EntryInvokes {
		a := &invokeExecutorArgs{
			context:               args.Context,
			realityName:           args.RealityName,
			universeCanonicalName: args.UniverseCanonicalName,
			universeID:            args.UniverseID,
			event:                 args.Event,
			invoke:                *invoke,
		}
		go qm.invokeExecutor.ExecuteInvoke(ctx, a)
	}
}

func (qm *ExQuantumMachine) ExecuteExitInvokes(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.ExitInvokes) == 0 {
		return
	}

	if qm.invokeExecutor == nil {
		return
	}

	for _, invoke := range qm.model.UniversalConstants.ExitInvokes {
		a := &invokeExecutorArgs{
			context:               args.Context,
			realityName:           args.RealityName,
			universeCanonicalName: args.UniverseCanonicalName,
			universeID:            args.UniverseID,
			event:                 args.Event,
			invoke:                *invoke,
		}
		go qm.invokeExecutor.ExecuteInvoke(ctx, a)
	}
}

func (qm *ExQuantumMachine) ExecuteEntryAction(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) error {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.EntryActions) == 0 {
		return nil
	}

	if qm.actionExecutor == nil {
		return nil
	}

	for _, action := range qm.model.UniversalConstants.EntryActions {
		a := &actionExecutorArgs{
			context:               args.Context,
			realityName:           args.RealityName,
			universeCanonicalName: args.UniverseCanonicalName,
			universeID:            args.UniverseID,
			event:                 args.Event,
			action:                *action,
			getSnapshotFn:         qm.GetSnapshot,
		}
		if err := qm.actionExecutor.ExecuteAction(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (qm *ExQuantumMachine) ExecuteExitAction(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) error {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.ExitActions) == 0 {
		return nil
	}

	if qm.actionExecutor == nil {
		return nil
	}

	for _, action := range qm.model.UniversalConstants.ExitActions {
		a := &actionExecutorArgs{
			context:               args.Context,
			realityName:           args.RealityName,
			universeCanonicalName: args.UniverseCanonicalName,
			universeID:            args.UniverseID,
			event:                 args.Event,
			action:                *action,
			getSnapshotFn:         qm.GetSnapshot,
		}
		if err := qm.actionExecutor.ExecuteAction(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (qm *ExQuantumMachine) ExecuteTransitionInvokes(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.InvokesOnTransition) == 0 {
		return
	}

	if qm.invokeExecutor == nil {
		return
	}

	for _, invoke := range qm.model.UniversalConstants.InvokesOnTransition {
		a := &invokeExecutorArgs{
			context:               args.Context,
			realityName:           args.RealityName,
			universeCanonicalName: args.UniverseCanonicalName,
			universeID:            args.UniverseID,
			event:                 args.Event,
			invoke:                *invoke,
		}
		go qm.invokeExecutor.ExecuteInvoke(ctx, a)
	}
}

func (qm *ExQuantumMachine) ExecuteTransitionAction(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) error {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.ActionsOnTransition) == 0 {
		return nil
	}

	if qm.actionExecutor == nil {
		return nil
	}

	for _, action := range qm.model.UniversalConstants.ActionsOnTransition {
		a := &actionExecutorArgs{
			context:               args.Context,
			realityName:           args.RealityName,
			universeCanonicalName: args.UniverseCanonicalName,
			universeID:            args.UniverseID,
			event:                 args.Event,
			action:                *action,
			getSnapshotFn:         qm.GetSnapshot,
		}
		if err := qm.actionExecutor.ExecuteAction(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------

func (qm *ExQuantumMachine) getActiveUniverses() []*ExUniverse {
	var activeUniverses []*ExUniverse
	for _, u := range qm.universes {
		if u.isActive() {
			activeUniverses = append(activeUniverses, u)
		}
	}
	return activeUniverses
}

func (qm *ExQuantumMachine) getLazyActiveUniverses(event instrumentation.Event) []*ExUniverse {
	var activeUniverses []*ExUniverse
	for _, u := range qm.universes {
		if u.canHandleEvent(event) {
			activeUniverses = append(activeUniverses, u)
		}
	}
	return activeUniverses
}

func (qm *ExQuantumMachine) executeTargetPairs(ctx context.Context, pairs []devtoolkit.Pair[instrumentation.Event, []string]) error {

	// while there are pairs to execute
	for len(pairs) > 0 {
		pair := pairs[0]
		pairs = pairs[1:]

		// execute transition
		evt, targets := pair.GetAll()
		newTargets, err := qm.executeTransitions(ctx, evt, targets)
		if err != nil {
			return err
		}

		if len(newTargets) == 0 {
			continue
		}

		// add new targets to the queue
		newPair := devtoolkit.NewPair[instrumentation.Event, []string](pair.GetFirst(), newTargets)
		pairs = append(pairs, newPair)
	}

	return nil
}

func (qm *ExQuantumMachine) executeTransitions(ctx context.Context, event instrumentation.Event, targets []string) ([]string, error) {

	var newTargets []string

	for _, target := range targets {
		refT, parts, _ := processReference(target)
		exUniverse := qm.universes[parts[0]]

		var realityName *string = nil
		if refT == RefTypeUniverseReality {
			realityName = &parts[1]
		}

		newTransitions, _, err := exUniverse.handleEvent(ctx, realityName, event, qm.machineContext)
		if err != nil {
			return nil, err
		}

		newTargets = append(newTargets, newTransitions...)
	}

	return newTargets, nil
}
