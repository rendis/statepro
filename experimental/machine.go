package experimental

import (
	"context"
	"fmt"
	"sync"

	"github.com/rendis/abslog/v3"
	"github.com/rendis/devtoolkit"
	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

const (
	universeNotFoundErrorTemplate = "universe '%s' not found"
)

type initFunc func(context.Context, any, *ExUniverse, []string, instrumentation.Event) ([]string, instrumentation.Event, error)

var qmInitFunctions = map[refType]initFunc{
	RefTypeUniverse: func(ctx context.Context, uCtx any, u *ExUniverse, _ []string, event instrumentation.Event) ([]string, instrumentation.Event, error) {
		return u.start(ctx, uCtx, event)
	},
	RefTypeUniverseReality: func(ctx context.Context, uCtx any, u *ExUniverse, realities []string, event instrumentation.Event) ([]string, instrumentation.Event, error) {
		return u.startOnReality(ctx, realities[1], uCtx, event)
	},
}

func NewExQuantumMachine(qmm *theoretical.QuantumMachineModel, universes []*ExUniverse) (instrumentation.QuantumMachine, error) {

	qm := &ExQuantumMachine{
		model:     qmm,
		universes: map[string]*ExUniverse{},
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

	// universes is the map of the quantum machine universes
	// key: theoretical.UniverseModel.ID
	universes map[string]*ExUniverse

	// quantumMachineMtx is the mutex for the quantum machine
	quantumMachineMtx sync.Mutex
}

//--------- QuantumMachine interface implementation ---------

func (qm *ExQuantumMachine) Init(ctx context.Context, machineContext any) error {
	return qm.init(ctx, machineContext, nil)
}

func (qm *ExQuantumMachine) InitWithEvent(ctx context.Context, machineContext any, event instrumentation.Event) error {
	return qm.init(ctx, machineContext, event)
}

func (qm *ExQuantumMachine) SendEvent(ctx context.Context, event instrumentation.Event) (bool, error) {
	qm.quantumMachineMtx.Lock()
	defer qm.quantumMachineMtx.Unlock()

	var pairs []devtoolkit.Pair[instrumentation.Event, []string]

	activeUniverses := qm.getLazyActiveUniverses(event)
	if len(activeUniverses) == 0 {
		return false, nil
	}

	for _, u := range activeUniverses {
		externalTargets, err := u.handleEvent(ctx, nil, event, qm.machineContext)
		if err != nil {
			return true, err
		}

		if len(externalTargets) == 0 {
			continue
		}

		pair := devtoolkit.NewPair(event, externalTargets)
		pairs = append(pairs, pair)
	}

	return true, qm.executeExternalTargetPairs(ctx, pairs)
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
		qm.setSnapshotFromUniverseInSuperposition(u, machineSnapshot)

		// add tracking
		machineSnapshot.AddTracking(u.model.ID, u.tracking)
	}

	return machineSnapshot
}

func (qm *ExQuantumMachine) ReplayOnEntry(ctx context.Context) error {
	qm.quantumMachineMtx.Lock()
	defer qm.quantumMachineMtx.Unlock()

	var evt = NewEventBuilder("replayOnEntry").
		SetEvtType(instrumentation.EventTypeOnEntry).
		SetFlags(instrumentation.EventFlags{
			ReplayOnEntry: true,
		}).
		Build()

	for _, u := range qm.getActiveUniverses() {
		err := u.replayOnEntry(ctx, evt, qm.machineContext)
		if err != nil {
			return err
		}
	}

	return nil
}

func (qm *ExQuantumMachine) PositionMachine(ctx context.Context, machineContext any, universeID string, realityID string, executeFlow bool) error {
	qm.quantumMachineMtx.Lock()
	defer qm.quantumMachineMtx.Unlock()

	// Validate parameters
	if universeID == "" {
		return fmt.Errorf("universeID cannot be empty")
	}

	if realityID == "" {
		return fmt.Errorf("realityID cannot be empty")
	}

	// Get target universe
	universe, ok := qm.universes[universeID]
	if !ok {
		return fmt.Errorf(universeNotFoundErrorTemplate, universeID)
	}

	// Validate reality exists in universe model
	_, ok = universe.model.Realities[realityID]
	if !ok {
		return fmt.Errorf("reality '%s' does not exist in universe '%s'", realityID, universeID)
	}

	// Set machine context
	qm.machineContext = machineContext

	if !executeFlow {
		// Path 1: Static positioning - only set flags, no execution
		return universe.positionStatic(realityID, machineContext)
	}

	// Path 2: Execute full flow - reuse existing startOnReality with proper external target handling

	// reset reality initialized flag, to force entry actions to be executed
	universe.realityInitialized = false

	// execute startOnReality
	externalTargets, originalEvent, err := universe.startOnReality(ctx, realityID, machineContext, nil)
	if err != nil {
		return err
	}

	// Process any external targets generated (cross-universe transitions)
	if len(externalTargets) > 0 {
		// Use the original event from startOnReality to preserve metadata, type, and flags
		// This ensures downstream universes receive the correct StartOn event context
		var pairs []devtoolkit.Pair[instrumentation.Event, []string]
		pair := devtoolkit.NewPair(originalEvent, externalTargets)
		pairs = append(pairs, pair)

		// Process external targets using existing machinery
		return qm.executeExternalTargetPairs(ctx, pairs)
	}

	return nil
}

func (qm *ExQuantumMachine) PositionMachineOnInitial(ctx context.Context, machineContext any, universeID string, executeFlow bool) error {
	// Validate universeID
	if universeID == "" {
		return fmt.Errorf("universeID cannot be empty")
	}

	// Get target universe (without lock - PositionMachine will handle locking)
	qm.quantumMachineMtx.Lock()
	universe, ok := qm.universes[universeID]
	qm.quantumMachineMtx.Unlock()

	if !ok {
		return fmt.Errorf(universeNotFoundErrorTemplate, universeID)
	}

	// Get initial state from universe model
	initialState := universe.model.Initial
	if initialState == nil || *initialState == "" {
		return fmt.Errorf("universe '%s' has no initial state configured", universeID)
	}

	// Delegate to PositionMachine with the initial state
	return qm.PositionMachine(ctx, machineContext, universeID, *initialState, executeFlow)
}

func (qm *ExQuantumMachine) PositionMachineByCanonicalName(ctx context.Context, machineContext any, universeCanonicalName string, realityID string, executeFlow bool) error {
	// Validate canonical name
	if universeCanonicalName == "" {
		return fmt.Errorf("universeCanonicalName cannot be empty")
	}

	// Find universe by canonical name
	qm.quantumMachineMtx.Lock()
	var universeID string
	for id, universe := range qm.universes {
		if universe.model.CanonicalName == universeCanonicalName {
			universeID = id
			break
		}
	}
	qm.quantumMachineMtx.Unlock()

	if universeID == "" {
		return fmt.Errorf("universe with canonical name '%s' not found", universeCanonicalName)
	}

	// Delegate to PositionMachine with the resolved universe ID
	return qm.PositionMachine(ctx, machineContext, universeID, realityID, executeFlow)
}

func (qm *ExQuantumMachine) PositionMachineOnInitialByCanonicalName(ctx context.Context, machineContext any, universeCanonicalName string, executeFlow bool) error {
	// Validate canonical name
	if universeCanonicalName == "" {
		return fmt.Errorf("universeCanonicalName cannot be empty")
	}

	// Find universe by canonical name
	qm.quantumMachineMtx.Lock()
	var universeID string
	for id, universe := range qm.universes {
		if universe.model.CanonicalName == universeCanonicalName {
			universeID = id
			break
		}
	}
	qm.quantumMachineMtx.Unlock()

	if universeID == "" {
		return fmt.Errorf("universe with canonical name '%s' not found", universeCanonicalName)
	}

	// Delegate to PositionMachineOnInitial with the resolved universe ID
	return qm.PositionMachineOnInitial(ctx, machineContext, universeID, executeFlow)
}

//--------- ConstantsLawsExecutor interface implementation ---------

func (qm *ExQuantumMachine) ExecuteEntryInvokes(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.EntryInvokes) == 0 {
		return
	}

	for _, invoke := range qm.model.UniversalConstants.EntryInvokes {
		qm.executeInvoke(ctx, *invoke, args)
	}
}

func (qm *ExQuantumMachine) ExecuteExitInvokes(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.ExitInvokes) == 0 {
		return
	}

	for _, invoke := range qm.model.UniversalConstants.ExitInvokes {
		qm.executeInvoke(ctx, *invoke, args)
	}
}

func (qm *ExQuantumMachine) ExecuteEntryAction(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) error {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.EntryActions) == 0 {
		return nil
	}

	for _, action := range qm.model.UniversalConstants.EntryActions {
		if err := qm.executeAction(ctx, action, args); err != nil {
			return err
		}
	}
	return nil
}

func (qm *ExQuantumMachine) ExecuteExitAction(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) error {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.ExitActions) == 0 {
		return nil
	}

	for _, action := range qm.model.UniversalConstants.ExitActions {
		if err := qm.executeAction(ctx, action, args); err != nil {
			return err
		}
	}
	return nil
}

func (qm *ExQuantumMachine) ExecuteTransitionInvokes(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.InvokesOnTransition) == 0 {
		return
	}

	for _, invoke := range qm.model.UniversalConstants.InvokesOnTransition {
		qm.executeInvoke(ctx, *invoke, args)
	}
}

func (qm *ExQuantumMachine) ExecuteTransitionAction(ctx context.Context, args *instrumentation.QuantumMachineExecutorArgs) error {
	if qm.model.UniversalConstants == nil || len(qm.model.UniversalConstants.ActionsOnTransition) == 0 {
		return nil
	}

	for _, action := range qm.model.UniversalConstants.ActionsOnTransition {
		if err := qm.executeAction(ctx, action, args); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------

func (qm *ExQuantumMachine) init(ctx context.Context, machineContext any, event instrumentation.Event) error {
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
		externalTargets, evt, err := initFn(ctx, machineContext, universe, parts, event)
		if err != nil {
			return err
		}

		pair := devtoolkit.NewPair(evt, externalTargets)
		pairs = append(pairs, pair)
	}

	return qm.executeExternalTargetPairs(ctx, pairs)
}

func (qm *ExQuantumMachine) executeInvoke(ctx context.Context, invoke theoretical.InvokeModel, args *instrumentation.QuantumMachineExecutorArgs) {
	if invoke.Src == "" {
		return
	}

	u := qm.universes[args.UniverseID]

	a := &invokeExecutorArgs{
		context:               args.Context,
		realityName:           args.RealityName,
		universeCanonicalName: args.UniverseCanonicalName,
		universeID:            args.UniverseID,
		universeMetadata:      u.metadata,
		event:                 args.Event,
		invoke:                invoke,
	}

	if fn := builtin.GetInvoke(invoke.Src); fn != nil {
		go fn(ctx, a)
		return
	}

	abslog.WarnCtxf(ctx, "invoke '%s' not found", invoke.Src)
}

func (qm *ExQuantumMachine) executeAction(ctx context.Context, model *theoretical.ActionModel, args *instrumentation.QuantumMachineExecutorArgs) error {
	if model.Src == "" {
		return nil
	}

	u := qm.universes[args.UniverseID]

	a := &actionExecutorArgs{
		context:               args.Context,
		realityName:           args.RealityName,
		universeCanonicalName: args.UniverseCanonicalName,
		universeID:            args.UniverseID,
		universeMetadata:      u.metadata,
		event:                 args.Event,
		action:                *model,
		actionType:            instrumentation.ActionTypeEntry,
		getSnapshotFn:         qm.GetSnapshot,
	}

	if fn := builtin.GetAction(model.Src); fn != nil {
		return fn(ctx, a)
	}

	abslog.WarnCtxf(ctx, "action '%s' not found", model.Src)
	return nil
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

func (qm *ExQuantumMachine) getActiveUniverses() []*ExUniverse {
	var activeUniverses []*ExUniverse
	for _, u := range qm.universes {
		if u.isActive() {
			activeUniverses = append(activeUniverses, u)
		}
	}
	return activeUniverses
}

func (qm *ExQuantumMachine) executeExternalTargetPairs(ctx context.Context, pairs []devtoolkit.Pair[instrumentation.Event, []string]) error {
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
		newPair := devtoolkit.NewPair(pair.GetFirst(), newTargets)
		pairs = append(pairs, newPair)
	}

	return nil
}

func (qm *ExQuantumMachine) executeTransitions(ctx context.Context, event instrumentation.Event, targets []string) ([]string, error) {

	var newTargets []string

	for _, target := range targets {
		refT, parts, _ := processReference(target)
		exUniverse := qm.universes[parts[0]]

		if exUniverse == nil {
			return nil, fmt.Errorf(universeNotFoundErrorTemplate, parts[0])
		}

		var realityName *string = nil
		if refT == RefTypeUniverseReality {
			realityName = &parts[1]
		}

		newTransitions, err := exUniverse.handleEvent(ctx, realityName, event, qm.machineContext)
		if err != nil {
			return nil, err
		}

		newTargets = append(newTargets, newTransitions...)
	}

	return newTargets, nil
}

func (qm *ExQuantumMachine) setSnapshotFromUniverseInSuperposition(u *ExUniverse, machineSnapshot *instrumentation.MachineSnapshot) {
	if u.inSuperposition {
		// superposition universe resume
		var realityBeforeSuperposition = "*"
		if u.realityBeforeSuperposition != nil {
			realityBeforeSuperposition = *u.realityBeforeSuperposition
		}

		if !u.isFinalReality {
			machineSnapshot.AddSuperpositionUniverse(u.model.CanonicalName, realityBeforeSuperposition)
		} else {
			machineSnapshot.AddSuperpositionUniverseFinalized(u.model.CanonicalName, realityBeforeSuperposition)
		}
	}
}
