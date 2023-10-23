package experimental

import (
	"context"
	"errors"
	"fmt"
	"github.com/rendis/devtoolkit"
	"github.com/rendis/statepro/v3/theoretical"
	"sync"
)

const (
	startEventName   = "start"
	startOnEventName = "startOn"
)

type ExUniverseSnapshot map[string]any

type universeInfoSnapshot struct {
	ID                         string      `json:"id" bson:"id" xml:"id"`
	Initialized                bool        `json:"initialized" bson:"initialized" xml:"initialized"`
	CurrentReality             *string     `json:"currentReality,omitempty" bson:"currentReality,omitempty" xml:"currentReality,omitempty"`
	RealityInitialized         bool        `json:"realityInitialized" bson:"realityInitialized" xml:"realityInitialized"`
	InSuperposition            bool        `json:"inSuperposition" bson:"inSuperposition" xml:"inSuperposition"`
	RealityBeforeSuperposition *string     `json:"realityBeforeSuperposition,omitempty" bson:"realityBeforeSuperposition,omitempty" xml:"realityBeforeSuperposition,omitempty"`
	Accumulator                Accumulator `json:"accumulator,omitempty" bson:"accumulator,omitempty" xml:"accumulator,omitempty"`
}

func NewExUniverse(id string, model *theoretical.UniverseModel, laws UniverseLaws) *ExUniverse {
	return &ExUniverse{
		id:    id,
		model: model,
		laws:  laws,
	}
}

type ExUniverse struct {
	id string

	// initialized true when the universe is initialized
	// the universe is initialized when the first operation is executed
	initialized bool

	// universeContext is the universe context
	universeContext any

	// model of the ExUniverse
	model *theoretical.UniverseModel

	// laws of the ExUniverse
	laws UniverseLaws

	// universeMtx is the mutex for the ExUniverse
	universeMtx sync.Mutex

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

	// eventAccumulator universe event accumulator
	// used to accumulate events for each reality when the ExUniverse is in superposition (inSuperposition == true)
	eventAccumulator Accumulator
}

func (u *ExUniverse) IsActive() bool {
	return u.initialized && (u.inSuperposition || (u.currentReality != nil && !u.isFinalReality))
}

func (u *ExUniverse) IsInitialized() bool {
	return u.initialized
}

func (u *ExUniverse) HandleEvent(ctx context.Context, realityName *string, evt Event) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	var handleEventFn func() error

	if realityName != nil {
		handleEventFn = func() error { return u.receiveEventToReality(ctx, *realityName, evt) }
	} else {
		handleEventFn = func() error { return u.receiveEvent(ctx, evt) }
	}

	targets, err := u.universeDecorator(handleEventFn)
	return targets, evt, err
}

func (u *ExUniverse) PlaceOn(realityName string) error {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

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

func (u *ExUniverse) PlaceOnInitial() error {
	return u.PlaceOn(u.model.Initial)
}

func (u *ExUniverse) Start(ctx context.Context) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	evt := &event{EvtType: startEventName}

	var initFn = func() error {
		if err := u.initializeUniverseOn(ctx, u.model.Initial, evt); err != nil {
			return errors.Join(fmt.Errorf("error initializing universe '%s'", u.id), err)
		}
		return nil
	}

	targets, err := u.universeDecorator(initFn)
	return targets, evt, err
}

func (u *ExUniverse) StartOnReality(ctx context.Context, realityName string) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	evt := &event{EvtType: startEventName}

	var initFn = func() error {
		if err := u.initializeUniverseOn(ctx, realityName, evt); err != nil {
			return errors.Join(fmt.Errorf("error initializing universe '%s'", u.id), err)
		}
		return nil
	}

	targets, err := u.universeDecorator(initFn)
	return targets, evt, err
}

func (u *ExUniverse) SendEvent(ctx context.Context, evt Event) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	var sendEventFn = func() error {
		return u.receiveEvent(ctx, evt)
	}

	targets, err := u.universeDecorator(sendEventFn)
	return targets, evt, err
}

func (u *ExUniverse) SendEventToReality(ctx context.Context, realityName string, evt Event) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	var sendEventFn = func() error {
		return u.receiveEventToReality(ctx, realityName, evt)
	}

	targets, err := u.universeDecorator(sendEventFn)
	return targets, evt, err
}

// GetSnapshot returns a snapshot of the universe
func (u *ExUniverse) GetSnapshot() ExUniverseSnapshot {
	var infoSnapshot = universeInfoSnapshot{
		ID:                         u.id,
		Initialized:                u.initialized,
		CurrentReality:             u.currentReality,
		RealityInitialized:         u.realityInitialized,
		InSuperposition:            u.inSuperposition,
		RealityBeforeSuperposition: u.realityBeforeSuperposition,
		Accumulator:                u.eventAccumulator,
	}

	m, _ := devtoolkit.StructToMap(infoSnapshot)
	return m
}

// LoadSnapshot loads a snapshot of the universe
func (u *ExUniverse) LoadSnapshot(universeSnapshot ExUniverseSnapshot) error {
	snapshot, err := devtoolkit.MapToStruct[universeInfoSnapshot](universeSnapshot)
	if err != nil {
		return errors.Join(fmt.Errorf("error loading snapshot for universe '%s'", u.id), err)
	}

	u.id = snapshot.ID
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

//------------------------------------------------------------------------------------------------------//

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

// receiveEventToReality receives an event only if one of the following conditions is met:
//   - the ExUniverse is in superposition
//   - not in superposition but the current reality is the target reality and not a final reality
func (u *ExUniverse) receiveEventToReality(ctx context.Context, reality string, event Event) error {
	// handling superposition
	if u.inSuperposition {
		isNewReality, err := u.accumulateEventForReality(ctx, reality, event)
		if err != nil {
			return errors.Join(fmt.Errorf("error accumulating event for reality '%s'", reality), err)
		}

		if isNewReality {
			// establish new reality
			if err := u.establishNewReality(ctx, reality, event); err != nil {
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
		"universe '%s' can't receive external event to reality '%s'. inSuperposition: '%t', currentRealityName: '%s', IsFinalReality: '%t'",
		u.id, reality, u.inSuperposition, currentRealityName, u.isFinalReality,
	)
}

// receiveEvent receives an event only if one of the following conditions is met:
//   - the ExUniverse is in superposition (the event will be accumulated for each reality)
//   - universe is not initialized, the event initializes the universe and will be received by the initial reality
//   - not in superposition and the current reality not a final reality
func (u *ExUniverse) receiveEvent(ctx context.Context, event Event) error {
	// handling superposition
	if u.inSuperposition {
		isNewReality, realityName, err := u.accumulateEventForAllRealities(ctx, event)
		if err != nil {
			return errors.Join(fmt.Errorf("error accumulating event for all realities"), err)
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
			return errors.Join(fmt.Errorf("error initializing universe '%s'", u.id), err)
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
		"universe '%s' can't receive external event. inSuperposition: '%t', currentRealityName: '%s', IsFinalReality: '%t'",
		u.id, u.inSuperposition, currentRealityName, u.isFinalReality,
	)
}

func (u *ExUniverse) onEvent(ctx context.Context, event Event) error {
	realityModel := u.model.GetReality(*u.currentReality)
	transitions, ok := realityModel.On[event.GetEventName()]
	if !ok {
		return fmt.Errorf("reality '%s' does not have transitions for event '%s'", realityModel.ID, event.GetEventName())
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

func (u *ExUniverse) initializeUniverseOn(ctx context.Context, realityName string, event Event) error {
	// establish initial reality
	if err := u.establishNewReality(ctx, realityName, event); err != nil {
		return errors.Join(fmt.Errorf("error establishing initial reality '%s'", realityName), err)
	}

	// mark universe as initialized
	u.initialized = true
	return nil
}

func (u *ExUniverse) establishNewReality(ctx context.Context, reality string, event Event) error {
	// set current reality
	u.currentReality = &reality

	// quit superposition
	u.inSuperposition = false
	u.realityBeforeSuperposition = nil

	// get if the current reality is a final reality
	realityModel := u.model.GetReality(reality)
	u.isFinalReality = theoretical.IsFinalState(realityModel.Type)

	// execute always
	if err := u.executeAlways(ctx, event); err != nil {
		return errors.Join(fmt.Errorf("error executing always transitions for universe '%s'", u.id), err)
	}

	// check if the always process left the system in superposition
	if u.inSuperposition {
		return nil
	}

	// mark current reality as initialized
	u.realityInitialized = true

	// execute on entry process
	if err := u.executeOnEntryProcess(ctx, event); err != nil {
		return errors.Join(fmt.Errorf("error executing on entry process for universe '%s'", u.id), err)
	}

	return nil
}

func (u *ExUniverse) executeAlways(ctx context.Context, event Event) error {
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
func (u *ExUniverse) doCyclicTransition(ctx context.Context, approvedTransition *theoretical.TransitionModel, event Event) error {
	for {
		// if no approved transition or no targets -> return nil
		if approvedTransition == nil || len(approvedTransition.Targets) == 0 {
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

		// execute current reality always transitions
		realityModel := u.model.GetReality(*u.currentReality)
		approvedTransition, err = u.getApprovedTransition(ctx, realityModel.Always, event)
		if err != nil {
			return errors.Join(fmt.Errorf("error executing always transitions for reality '%s'", realityModel.ID), err)
		}
	}
}

func (u *ExUniverse) getApprovedTransition(ctx context.Context, transitionModels []*theoretical.TransitionModel, evt Event) (*theoretical.TransitionModel, error) {
	if transitionModels == nil || len(transitionModels) == 0 {
		return nil, nil
	}

	for _, transition := range transitionModels {
		doTransition, err := u.executeCondition(ctx, transition.Condition, evt)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("error executing always transition condition '%s'", transition.Condition.Src), err)
		}

		if doTransition {
			return transition, nil
		}
	}

	return nil, nil
}

func (u *ExUniverse) executeCondition(ctx context.Context, conditionModel *theoretical.ConditionModel, evt Event) (bool, error) {
	// if conditionModel is nil then return true, the transition is always executed
	if conditionModel == nil {
		return true, nil
	}

	return u.laws.ExecuteCondition(
		ctx,
		conditionModel.Src,
		conditionModel.Args,
		u.universeContext,
		evt,
	)
}

func (u *ExUniverse) initSuperposition(ctx context.Context, targets []string, event Event) error {
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

func (u *ExUniverse) executeOnExitProcess(ctx context.Context, event Event) error {
	realityModel := u.model.GetReality(*u.currentReality)

	// execute on exit actions
	if err := u.executeActions(ctx, realityModel.ExitActions, event); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on exit actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	// execute on exit invokes, invokes are executed asynchronously
	u.executeInvokes(ctx, realityModel.ExitInvokes, event)

	return nil
}

func (u *ExUniverse) executeOnEntryProcess(ctx context.Context, event Event) error {
	realityModel := u.model.GetReality(*u.currentReality)

	// execute on entry actions
	if err := u.executeActions(ctx, realityModel.EntryActions, event); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on entry actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	// execute on entry invokes, invokes are executed asynchronously
	u.executeInvokes(ctx, realityModel.EntryInvokes, event)

	return nil
}

func (u *ExUniverse) executeActions(ctx context.Context, actionModels []*theoretical.ActionModel, event Event) error {
	if actionModels == nil || len(actionModels) == 0 {
		return nil
	}

	// execute actions
	for _, action := range actionModels {
		if err := u.laws.ExecuteAction(ctx, u.universeContext, event, *action); err != nil {
			return errors.Join(fmt.Errorf("error executing action '%s'", action.Src), err)
		}
	}

	return nil
}

func (u *ExUniverse) executeInvokes(ctx context.Context, invokeModels []*theoretical.InvokeModel, event Event) {
	if invokeModels == nil || len(invokeModels) == 0 {
		return
	}

	// execute invokes
	for _, invoke := range invokeModels {
		go u.laws.ExecuteInvoke(
			ctx,
			u.universeContext,
			event,
			*invoke,
		)
	}
}

func (u *ExUniverse) accumulateEventForReality(ctx context.Context, realityName string, event Event) (bool, error) {
	// accumulate event
	u.eventAccumulator.Accumulate(realityName, event)

	// execute observers
	realityModel := u.model.GetReality(realityName)
	isNewReality, err := u.executeObservers(ctx, realityModel, event)
	if err != nil {
		return false, errors.Join(fmt.Errorf("error executing observers for reality '%s'", realityModel.ID), err)
	}
	return isNewReality, nil
}

func (u *ExUniverse) accumulateEventForAllRealities(ctx context.Context, event Event) (bool, string, error) {
	priorityRealities := u.eventAccumulator.GetActiveRealities()
	var priorityRealitiesMap = make(map[string]bool)

	// accumulate event for priority realities
	for _, reality := range priorityRealities {
		priorityRealitiesMap[reality] = true
		isNewReality, err := u.accumulateEventForReality(ctx, reality, event)
		if err != nil {
			return false, "", errors.Join(fmt.Errorf("error accumulating event for reality '%s'", reality), err)
		}

		if isNewReality {
			return true, reality, nil
		}
	}

	// accumulate event for other realities
	for reality := range u.model.Realities {
		if _, ok := priorityRealitiesMap[reality]; !ok {
			isNewReality, err := u.accumulateEventForReality(ctx, reality, event)
			if err != nil {
				return false, "", errors.Join(fmt.Errorf("error accumulating event for reality '%s'", reality), err)
			}

			if isNewReality {
				return true, reality, nil
			}
		}
	}

	return false, "", nil
}

func (u *ExUniverse) executeObservers(ctx context.Context, realityModel *theoretical.RealityModel, evt Event) (bool, error) {
	if realityModel.Observers == nil || len(realityModel.Observers) == 0 {
		return false, nil
	}

	for _, observer := range realityModel.Observers {
		isApproved, err := u.laws.ExecuteObserver(ctx, u.universeContext, u.eventAccumulator.GetStatistics(), evt, *observer)
		if err != nil {
			return false, errors.Join(fmt.Errorf("error executing observer '%s'", observer.Src), err)
		}

		if isApproved {
			return true, nil
		}
	}

	return false, nil
}
