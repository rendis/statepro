package experimental

import (
	"context"
	"errors"
	"fmt"
	"github.com/rendis/devtoolkit"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

const (
	startEventName   = "start"
	startOnEventName = "startOn"
)

type universeInfoSnapshot struct {
	ID                         string            `json:"id"`
	CanonicalName              string            `json:"canonicalName"`
	Version                    string            `json:"version"`
	Initialized                bool              `json:"initialized"`
	CurrentReality             *string           `json:"currentReality,omitempty"`
	RealityInitialized         bool              `json:"realityInitialized"`
	InSuperposition            bool              `json:"inSuperposition"`
	RealityBeforeSuperposition *string           `json:"realityBeforeSuperposition,omitempty"`
	Accumulator                *eventAccumulator `json:"accumulator,omitempty"`
}

func NewExUniverse(
	model *theoretical.UniverseModel,
	laws instrumentation.UniverseLaws,
) *ExUniverse {
	return &ExUniverse{
		model:             model,
		observerExecutor:  getUniverseObserverExecutor(laws),
		actionExecutor:    getUniverseActionExecutor(laws),
		invokeExecutor:    getUniverseInvokeExecutor(laws),
		conditionExecutor: getUniverseConditionExecutor(laws),
	}
}

type ExUniverse struct {
	// initialized true when the universe is initialized
	// the universe is initialized when the first operation is executed
	initialized bool

	// universeContext is the universe context
	universeContext any

	// model of the ExUniverse
	model *theoretical.UniverseModel

	// machineLawsExecutor is the machine laws executor
	constantsLawsExecutor instrumentation.ConstantsLawsExecutor

	// laws executors
	observerExecutor  instrumentation.ObserverExecutor
	actionExecutor    instrumentation.ActionExecutor
	invokeExecutor    instrumentation.InvokeExecutor
	conditionExecutor instrumentation.ConditionExecutor

	// currentReality is the current reality of the ExUniverse
	currentReality *string

	// realityBeforeSuperposition is the reality before the ExUniverse entered in superposition
	realityBeforeSuperposition *string

	// isFinalReality true when the current reality type belongs to the final states
	// used to know when the ExUniverse has been exited and finalized
	isFinalReality bool

	// realityInitialized true when the current reality always transitions are executed
	realityInitialized bool

	// inSuperposition is true if the ExUniverse is in superposition
	inSuperposition bool

	// externalTargets is the list of external targets
	// used for ExQuantumMachine to send events to external targets
	// externalTargets are cleared on each interaction with the ExUniverse
	externalTargets []string

	// eventAccumulator universe Event accumulator
	// used to accumulate events for each reality when the ExUniverse is in superposition (inSuperposition == true)
	eventAccumulator instrumentation.Accumulator
}

// handleEvent handles an Event where depending on the state of the universe
func (u *ExUniverse) handleEvent(ctx context.Context, realityName *string, evt instrumentation.Event, universeContext any) ([]string, instrumentation.Event, error) {
	var handleEventFn func() error
	u.universeContext = universeContext

	if realityName != nil {
		handleEventFn = func() error { return u.receiveEventToReality(ctx, *realityName, evt) }
	} else {
		handleEventFn = func() error { return u.receiveEvent(ctx, evt) }
	}

	targets, err := u.universeDecorator(handleEventFn)
	return targets, evt, err
}

// start starts the universe on the default reality (initial reality)
// start set initial reality as the current reality and execute:
// - always operations
// - initial operations
func (u *ExUniverse) start(ctx context.Context, universeContext any) ([]string, instrumentation.Event, error) {
	u.universeContext = universeContext
	evt := NewEventBuilder(startEventName).
		SetEvtType(instrumentation.EventTypeStart).
		Build()

	var initFn = func() error {
		if err := u.initializeUniverseOn(ctx, u.model.Initial, evt); err != nil {
			return errors.Join(fmt.Errorf("error initializing universe '%s'", u.model.ID), err)
		}
		return nil
	}

	targets, err := u.universeDecorator(initFn)
	return targets, evt, err
}

// startOnReality starts the universe on the given reality
// startOnReality set the given reality as the current reality and execute:
// - always operations
// - initial operations
func (u *ExUniverse) startOnReality(ctx context.Context, realityName string, universeContext any) ([]string, instrumentation.Event, error) {
	u.universeContext = universeContext
	evt := NewEventBuilder(startOnEventName).
		SetEvtType(instrumentation.EventTypeStartOn).
		Build()

	var initFn = func() error {
		if err := u.initializeUniverseOn(ctx, realityName, evt); err != nil {
			return errors.Join(fmt.Errorf("error initializing universe '%s'", u.model.ID), err)
		}
		return nil
	}

	targets, err := u.universeDecorator(initFn)
	return targets, evt, err
}

// placeOn sets the given reality as the current reality
// placeOn not execute always, initial or exit operations, only set the current reality
func (u *ExUniverse) placeOn(realityName string) error {
	realityModel, ok := u.model.Realities[realityName]
	if !ok {
		return fmt.Errorf("reality '%s' does not exist", realityName)
	}

	u.initialized = true
	u.currentReality = &realityName
	u.realityInitialized = true
	u.inSuperposition = false
	u.realityBeforeSuperposition = nil
	u.externalTargets = nil
	u.eventAccumulator = newEventAccumulator()
	u.isFinalReality = theoretical.IsFinalState(realityModel.Type)
	return nil
}

// getSnapshot returns a snapshot of the universe
func (u *ExUniverse) getSnapshot() instrumentation.SerializedUniverseSnapshot {
	var infoSnapshot = universeInfoSnapshot{
		ID:                         u.model.ID,
		CanonicalName:              u.model.CanonicalName,
		Version:                    u.model.Version,
		Initialized:                u.initialized,
		CurrentReality:             u.currentReality,
		RealityInitialized:         u.realityInitialized,
		InSuperposition:            u.inSuperposition,
		RealityBeforeSuperposition: u.realityBeforeSuperposition,
	}

	if u.eventAccumulator != nil {
		if accumulator, ok := u.eventAccumulator.(*eventAccumulator); ok {
			infoSnapshot.Accumulator = accumulator
		}
	}

	m, _ := devtoolkit.StructToMap(infoSnapshot)
	return m
}

// loadSnapshot loads a snapshot of the universe
func (u *ExUniverse) loadSnapshot(universeSnapshot instrumentation.SerializedUniverseSnapshot) error {
	snapshot, err := devtoolkit.MapToStruct[universeInfoSnapshot](universeSnapshot)
	if err != nil {
		return errors.Join(fmt.Errorf("error loading snapshot for universe '%s'", u.model.ID), err)
	}

	u.initialized = snapshot.Initialized
	u.currentReality = snapshot.CurrentReality
	u.realityInitialized = snapshot.RealityInitialized
	u.inSuperposition = snapshot.InSuperposition
	u.realityBeforeSuperposition = snapshot.RealityBeforeSuperposition
	u.eventAccumulator = snapshot.Accumulator

	if u.currentReality != nil {
		realityModel := u.model.GetReality(*u.currentReality)
		if realityModel == nil {
			return fmt.Errorf("reality '%s' does not exist", *u.currentReality)
		}
		u.isFinalReality = theoretical.IsFinalState(realityModel.Type)
	}

	return nil
}

//--------------------------------------------------------//

// canHandleEvent returns true if the universe can handle the given Event
// A universe can handle an Event if:
// - the universe is initialized &&
// - || in superposition
// - || current reality not final and can handle the event
func (u *ExUniverse) canHandleEvent(evt instrumentation.Event) bool {
	// if not initialized -> false
	if !u.initialized {
		return false
	}

	// if in superposition -> true
	if u.inSuperposition {
		return true
	}

	// if current reality not final and can handle the event -> true
	if u.currentReality != nil && !u.isFinalReality && u.canRealityHandleEvent(*u.currentReality, evt) {
		return true
	}

	// otherwise -> false
	return false
}

// isActive returns true if the universe is active
// A universe is active if:
// - has been initialized &&
// - || it is in superposition state
// - || the current reality is established and it is not final
func (u *ExUniverse) isActive() bool {
	return u.initialized && (u.inSuperposition || !u.isFinalReality)
}

func (u *ExUniverse) universeDecorator(operation func() error) ([]string, error) {
	// clear externalTargets
	u.externalTargets = nil

	// execute operation
	if err := operation(); err != nil {
		return nil, err
	}

	// return externalTargets
	return u.externalTargets, nil
}

// receiveEventToReality receives an Event only if one of the following conditions is met:
//   - the ExUniverse is in superposition
//   - not in superposition but the current reality is the target reality and not a final reality
func (u *ExUniverse) receiveEventToReality(ctx context.Context, reality string, event instrumentation.Event) error {
	// handling superposition
	if u.inSuperposition {
		isNewReality, err := u.accumulateEventForReality(ctx, reality, event)
		if err != nil {
			return errors.Join(fmt.Errorf("error accumulating Event for reality '%s'", reality), err)
		}

		if isNewReality {
			// establish new reality
			if err = u.establishNewReality(ctx, reality, event); err != nil {
				return errors.Join(fmt.Errorf("error establishing new reality '%s'", reality), err)
			}
		}

		return nil
	}

	// handling not final current reality
	if !u.isFinalReality && *u.currentReality == reality {
		return u.onEvent(ctx, event)
	}

	// return error
	var currentRealityName string
	if u.currentReality != nil {
		currentRealityName = *u.currentReality
	}
	return fmt.Errorf(
		"universe '%s' can't receive Event '%s' to reality '%s'. inSuperposition: '%t', currentRealityName: '%s', IsFinalReality: '%t'",
		u.model.ID, event.GetEventName(), reality, u.inSuperposition, currentRealityName, u.isFinalReality,
	)
}

// receiveEvent receives an Event only if one of the following conditions is met:
//   - the ExUniverse is in superposition (the Event will be accumulated for each reality)
//   - universe is not initialized, the Event initializes the universe and will be received by the initial reality
//   - not in superposition and the current reality not a final reality
func (u *ExUniverse) receiveEvent(ctx context.Context, event instrumentation.Event) error {
	// handling superposition
	if u.inSuperposition {
		isNewReality, realityName, err := u.accumulateEventForAllRealities(ctx, event)
		if err != nil {
			return errors.Join(fmt.Errorf("error accumulating Event for all realities"), err)
		}

		if !isNewReality {
			return nil
		}

		// establish new reality
		if err = u.establishNewReality(ctx, realityName, event); err != nil {
			return errors.Join(fmt.Errorf("error establishing new reality '%s'", realityName), err)
		}
	}

	// handling not initialized universe
	if !u.initialized {
		if err := u.initializeUniverseOn(ctx, u.model.Initial, event); err != nil {
			return errors.Join(fmt.Errorf("error initializing universe '%s'", u.model.ID), err)
		}
		return nil
	}

	// handling not final current reality
	if !u.isFinalReality {
		return u.onEvent(ctx, event)
	}

	// return error
	var currentRealityName string
	if u.currentReality != nil {
		currentRealityName = *u.currentReality
	}

	return fmt.Errorf(
		"universe '%s' can't receive external Event. inSuperposition: '%t', currentRealityName: '%s', IsFinalReality: '%t'",
		u.model.ID, u.inSuperposition, currentRealityName, u.isFinalReality,
	)
}

func (u *ExUniverse) onEvent(ctx context.Context, event instrumentation.Event) error {
	realityModel := u.model.GetReality(*u.currentReality)
	transitions, ok := realityModel.On[event.GetEventName()]
	if !ok {
		return fmt.Errorf("reality '%s' does not have transitions for Event '%s'", realityModel.ID, event.GetEventName())
	}

	approvedTransition, err := u.getApprovedTransition(ctx, transitions, event)
	if err != nil {
		return errors.Join(fmt.Errorf("error executing on transitions for reality '%s'", realityModel.ID), err)
	}

	if err = u.doCyclicTransition(ctx, approvedTransition, event); err != nil {
		return errors.Join(fmt.Errorf("error executing on transitions for reality '%s'", realityModel.ID), err)
	}

	return nil
}

func (u *ExUniverse) initializeUniverseOn(ctx context.Context, realityName string, event instrumentation.Event) error {
	// mark universe as initialized
	u.initialized = true

	// establish initial reality
	if err := u.establishNewReality(ctx, realityName, event); err != nil {
		u.initialized = false
		return errors.Join(fmt.Errorf("error establishing initial reality '%s'", realityName), err)
	}

	return nil
}

func (u *ExUniverse) establishNewReality(ctx context.Context, reality string, event instrumentation.Event) error {
	// set current reality
	u.currentReality = &reality

	// clear Event accumulator
	u.eventAccumulator = nil

	// quit superposition
	u.inSuperposition = false
	u.realityBeforeSuperposition = nil

	// get if the current reality is a final reality
	realityModel := u.model.GetReality(reality)
	u.isFinalReality = theoretical.IsFinalState(realityModel.Type)

	// execute always
	if err := u.executeAlways(ctx, event); err != nil {
		return errors.Join(fmt.Errorf("error executing always transitions for universe '%s'", u.model.ID), err)
	}

	// check if the always process left the system in superposition
	if u.inSuperposition {
		return nil
	}

	// mark current reality as initialized
	u.realityInitialized = true

	// execute on entry process
	if err := u.executeOnEntryProcess(ctx, event); err != nil {
		return errors.Join(fmt.Errorf("error executing on entry process for universe '%s'", u.model.ID), err)
	}

	return nil
}

func (u *ExUniverse) executeAlways(ctx context.Context, event instrumentation.Event) error {
	// execute current reality always transitions
	realityModel := u.model.GetReality(*u.currentReality)
	approvedTransition, err := u.getApprovedTransition(ctx, realityModel.Always, event)
	if err != nil {
		return errors.Join(fmt.Errorf("error executing always transitions for reality '%s'", realityModel.ID), err)
	}

	return u.doCyclicTransition(ctx, approvedTransition, event)
}

// doCyclicTransition executes transition and always transitions of the current reality and goes to the next reality if necessary
// doCyclicTransition ends when:
//   - there are no always transitions
//   - there are always transitions but none of them are approved
//   - there are always transitions but all of them point to another universe (superposition)
func (u *ExUniverse) doCyclicTransition(ctx context.Context, approvedTransition *theoretical.TransitionModel, event instrumentation.Event) error {
	for {
		// if no approved transition or no targets -> return nil
		if approvedTransition == nil || len(approvedTransition.Targets) == 0 {
			return nil
		}

		// execute constants transition actions
		args := instrumentation.QuantumMachineExecutorArgs{
			Context:               u.universeContext,
			RealityName:           *u.currentReality,
			UniverseID:            u.model.ID,
			UniverseCanonicalName: u.model.CanonicalName,
			Event:                 event,
		}
		if err := u.constantsLawsExecutor.ExecuteTransitionAction(ctx, &args); err != nil {
			return errors.Join(fmt.Errorf("error executing constants transition actions for reality '%s'", *u.currentReality), err)
		}

		// execute universe transition actions
		if err := u.executeActions(ctx, approvedTransition.Actions, event); err != nil {
			return errors.Join(fmt.Errorf("error executing transition actions for reality '%s'", *u.currentReality), err)
		}

		// execute constants transition invokes, invokes are executed asynchronously
		u.constantsLawsExecutor.ExecuteTransitionInvokes(ctx, &args)

		// execute universe transition invokes
		u.executeInvokes(ctx, approvedTransition.Invokes, event)

		// get is transition is of type notify, save external targets and return nil
		if approvedTransition.IsNotification() {
			u.externalTargets = approvedTransition.Targets
			return nil
		}

		// len(targets) > 1 -> set superposition
		if len(approvedTransition.Targets) > 1 {
			return u.initSuperposition(ctx, approvedTransition.Targets, event)
		}

		// len(targets) == 1 && target points to another universe -> set superposition
		refTyp, _, err := processReference(approvedTransition.Targets[0])
		if err != nil {
			return errors.Join(fmt.Errorf("error processing reference '%s'", approvedTransition.Targets[0]), err)
		}
		if refTyp != RefTypeReality {
			return u.initSuperposition(ctx, approvedTransition.Targets, event)
		}

		// set current reality
		u.currentReality = &approvedTransition.Targets[0]
		realityModel := u.model.GetReality(*u.currentReality)
		u.isFinalReality = theoretical.IsFinalState(realityModel.Type)

		// execute current reality always transitions
		approvedTransition, err = u.getApprovedTransition(ctx, realityModel.Always, event)
		if err != nil {
			return errors.Join(fmt.Errorf("error executing always transitions for reality '%s'", realityModel.ID), err)
		}
	}
}

func (u *ExUniverse) getApprovedTransition(ctx context.Context, transitionModels []*theoretical.TransitionModel, event instrumentation.Event) (*theoretical.TransitionModel, error) {
	if transitionModels == nil || len(transitionModels) == 0 {
		return nil, nil
	}

	for _, transition := range transitionModels {
		doTransition, err := u.executeCondition(ctx, transition.Condition, event)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error executing always transition condition '%s'", transition.Condition.Src), err)
		}

		if doTransition {
			return transition, nil
		}
	}

	return nil, nil
}

func (u *ExUniverse) executeCondition(ctx context.Context, conditionModel *theoretical.ConditionModel, event instrumentation.Event) (bool, error) {
	// if conditionModel is nil then return true, the transition is always executed
	if conditionModel == nil {
		return true, nil
	}

	args := &conditionExecutorArgs{
		context:               u.universeContext,
		realityName:           *u.currentReality,
		universeCanonicalName: u.model.CanonicalName,
		universeID:            u.model.ID,
		event:                 event,
		condition:             *conditionModel,
	}
	return u.runConditionExecutor(ctx, args)
}

func (u *ExUniverse) initSuperposition(ctx context.Context, targets []string, event instrumentation.Event) error {
	// execute on exit process
	if err := u.executeOnExitProcess(ctx, event); err != nil {
		return err
	}

	// set superposition
	u.realityBeforeSuperposition = u.currentReality
	u.currentReality = nil
	u.inSuperposition = true
	u.externalTargets = targets
	u.eventAccumulator = newEventAccumulator()
	return nil
}

func (u *ExUniverse) executeOnEntryProcess(ctx context.Context, event instrumentation.Event) error {
	realityModel := u.model.GetReality(*u.currentReality)

	// execute on constants entry actions
	args := &instrumentation.QuantumMachineExecutorArgs{
		Context:               u.universeContext,
		RealityName:           realityModel.ID,
		UniverseCanonicalName: u.model.CanonicalName,
		UniverseID:            u.model.ID,
		Event:                 event,
	}
	if err := u.constantsLawsExecutor.ExecuteEntryAction(ctx, args); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on entry machine actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	// execute on entry universe actions
	if err := u.executeActions(ctx, realityModel.EntryActions, event); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on entry actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	// execute on entry constants invokes, invokes are executed asynchronously
	u.constantsLawsExecutor.ExecuteEntryInvokes(ctx, args)

	// execute on entry universe invokes, invokes are executed asynchronously
	u.executeInvokes(ctx, realityModel.EntryInvokes, event)

	return nil
}

func (u *ExUniverse) executeOnExitProcess(ctx context.Context, event instrumentation.Event) error {
	realityModel := u.model.GetReality(*u.currentReality)

	// execute on exit constants actions
	args := &instrumentation.QuantumMachineExecutorArgs{
		Context:               u.universeContext,
		RealityName:           realityModel.ID,
		UniverseCanonicalName: u.model.CanonicalName,
		UniverseID:            u.model.ID,
		Event:                 event,
	}

	if err := u.constantsLawsExecutor.ExecuteExitAction(ctx, args); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on exit machine actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	// execute on exit universe actions
	if err := u.executeActions(ctx, realityModel.ExitActions, event); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on exit actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	// execute on exit constants invokes, invokes are executed asynchronously
	u.constantsLawsExecutor.ExecuteExitInvokes(ctx, args)

	// execute on exit universe invokes, invokes are executed asynchronously
	u.executeInvokes(ctx, realityModel.ExitInvokes, event)

	return nil
}

func (u *ExUniverse) executeActions(ctx context.Context, actionModels []*theoretical.ActionModel, event instrumentation.Event) error {
	if actionModels == nil || len(actionModels) == 0 {
		return nil
	}

	// execute actions
	for _, action := range actionModels {
		args := &actionExecutorArgs{
			context:               u.universeContext,
			realityName:           *u.currentReality,
			universeCanonicalName: u.model.CanonicalName,
			universeID:            u.model.ID,
			event:                 event,
			action:                *action,
			getSnapshotFn:         u.constantsLawsExecutor.GetSnapshot,
		}
		if err := u.runActionExecutor(ctx, args); err != nil {
			return errors.Join(fmt.Errorf("error executing action '%s'", action.Src), err)
		}
	}

	return nil
}

func (u *ExUniverse) executeInvokes(ctx context.Context, invokeModels []*theoretical.InvokeModel, event instrumentation.Event) {
	if invokeModels == nil || len(invokeModels) == 0 {
		return
	}

	// execute invokes
	for _, invoke := range invokeModels {
		args := &invokeExecutorArgs{
			context:               u.universeContext,
			realityName:           *u.currentReality,
			universeCanonicalName: u.model.CanonicalName,
			universeID:            u.model.ID,
			event:                 event,
			invoke:                *invoke,
		}
		u.runInvokeExecutor(ctx, args)
	}
}

func (u *ExUniverse) accumulateEventForReality(ctx context.Context, realityName string, event instrumentation.Event) (bool, error) {
	// accumulate Event
	u.eventAccumulator.Accumulate(realityName, event)

	// execute observers
	realityModel := u.model.GetReality(realityName)
	isNewReality, err := u.executeObservers(ctx, realityModel, event)
	if err != nil {
		return false, errors.Join(fmt.Errorf("error executing observers for reality '%s'", realityModel.ID), err)
	}
	return isNewReality, nil
}

func (u *ExUniverse) accumulateEventForAllRealities(ctx context.Context, event instrumentation.Event) (bool, string, error) {
	priorityRealities := u.eventAccumulator.GetActiveRealities()
	var priorityRealitiesMap = make(map[string]bool)

	// accumulate Event for priority realities
	for _, reality := range priorityRealities {
		priorityRealitiesMap[reality] = true
		isNewReality, err := u.accumulateEventForReality(ctx, reality, event)
		if err != nil {
			return false, "", errors.Join(fmt.Errorf("error accumulating Event for reality '%s'", reality), err)
		}

		if isNewReality {
			return true, reality, nil
		}
	}

	// accumulate Event for other realities
	for reality := range u.model.Realities {
		if _, ok := priorityRealitiesMap[reality]; !ok {
			isNewReality, err := u.accumulateEventForReality(ctx, reality, event)
			if err != nil {
				return false, "", errors.Join(fmt.Errorf("error accumulating Event for reality '%s'", reality), err)
			}

			if isNewReality {
				return true, reality, nil
			}
		}
	}

	return false, "", nil
}

func (u *ExUniverse) executeObservers(ctx context.Context, realityModel *theoretical.RealityModel, event instrumentation.Event) (bool, error) {
	if realityModel.Observers == nil || len(realityModel.Observers) == 0 {
		return false, nil
	}

	for _, observer := range realityModel.Observers {
		args := &observerExecutorArgs{
			context:               u.universeContext,
			realityName:           realityModel.ID,
			universeCanonicalName: u.model.CanonicalName,
			universeID:            u.model.ID,
			accumulatorStatistics: u.eventAccumulator.GetStatistics(),
			event:                 event,
			observer:              *observer,
		}
		isApproved, err := u.runObserverExecutor(ctx, args)
		if err != nil {
			return false, errors.Join(fmt.Errorf("error executing observer '%s'", observer.Src), err)
		}

		if isApproved {
			return true, nil
		}
	}

	return false, nil
}

func (u *ExUniverse) canRealityHandleEvent(realityName string, evt instrumentation.Event) bool {
	realityModel := u.model.GetReality(realityName)
	if realityModel == nil {
		return false
	}

	_, ok := realityModel.On[evt.GetEventName()]
	return ok
}

//------------- executors -------------

func (u *ExUniverse) runObserverExecutor(ctx context.Context, args *observerExecutorArgs) (bool, error) {
	if u.observerExecutor == nil {
		return false, nil
	}
	return u.observerExecutor.ExecuteObserver(ctx, args)
}

func (u *ExUniverse) runActionExecutor(ctx context.Context, args *actionExecutorArgs) error {
	if u.actionExecutor == nil {
		return nil
	}
	return u.actionExecutor.ExecuteAction(ctx, args)
}

func (u *ExUniverse) runInvokeExecutor(ctx context.Context, args *invokeExecutorArgs) {
	if u.invokeExecutor == nil {
		return
	}
	go u.invokeExecutor.ExecuteInvoke(ctx, args)
}

func (u *ExUniverse) runConditionExecutor(ctx context.Context, args *conditionExecutorArgs) (bool, error) {
	if u.conditionExecutor == nil {
		return false, nil
	}
	return u.conditionExecutor.ExecuteCondition(ctx, args)
}
