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

func NewExUniverse(id string, model *theoretical.UniverseModel, laws UniverseLaws) *ExUniverse {
	return &ExUniverse{
		id:    id,
		model: model,
		laws:  laws,
	}
}

type universeInfoSnapshot struct {
	ID                 string      `json:"id,omitempty" bson:"id,omitempty" xml:"id,omitempty"`
	Initialized        bool        `json:"initialized,omitempty" bson:"initialized,omitempty" xml:"initialized,omitempty"`
	CurrentReality     *string     `json:"currentReality,omitempty" bson:"currentReality,omitempty" xml:"currentReality,omitempty"`
	RealityInitialized bool        `json:"realityInitialized,omitempty" bson:"realityInitialized,omitempty" xml:"realityInitialized,omitempty"`
	InSuperposition    bool        `json:"inSuperposition,omitempty" bson:"inSuperposition,omitempty" xml:"inSuperposition,omitempty"`
	Accumulator        Accumulator `json:"accumulator,omitempty" bson:"accumulator,omitempty" xml:"accumulator,omitempty"`
}

type ExUniverse struct {
	id string

	// initialized true when the universe is initialized
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

	// finalReality true when the current reality type belongs to the final states
	// used to know when the ExUniverse has been exited and finalized
	finalReality bool

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
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()
	return u.initialized && (u.inSuperposition || (u.currentReality != nil && !u.finalReality))
}

func (u *ExUniverse) IsInitialized() bool {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()
	return u.initialized
}

func (u *ExUniverse) HandleExternalEvent(ctx context.Context, realityName *string, evt Event) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	// u.initialized == false && realityName == nil -> PlaceOnInitial & SendEvent
	if !u.initialized && realityName == nil {
		_ = u.PlaceOnInitial(ctx)
		return u.SendEvent(ctx, evt)
	}

	// u.initialized == false && realityName != nil -> PlaceOn & SendEvent
	if !u.initialized && realityName != nil {
		_ = u.PlaceOn(ctx, *realityName)
		return u.SendEvent(ctx, evt)
	}

	// u.initialized == true && realityName == nil -> SendEvent
	if u.initialized && realityName == nil {
		return u.SendEvent(ctx, evt)
	}

	// u.initialized == true && realityName != nil -> SendEventToReality
	if u.initialized && realityName != nil {
		return u.SendEventToReality(ctx, *realityName, evt)
	}

	return nil, nil, nil
}

func (u *ExUniverse) PlaceOn(ctx context.Context, realityName string) error {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	if _, ok := u.model.Realities[realityName]; !ok {
		return fmt.Errorf("reality '%s' does not exist", realityName)
	}

	return u.establishNewReality(ctx, &realityName, nil)
}

func (u *ExUniverse) PlaceOnInitial(ctx context.Context) error {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	return u.establishNewReality(ctx, &u.model.Initial, nil)
}

func (u *ExUniverse) Start(ctx context.Context) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	if err := u.establishNewReality(ctx, &u.model.Initial, nil); err != nil {
		return nil, nil, err
	}

	evt := &event{EvtType: startEventName}

	ts, err := u.externalTransitionsDecorator(func() error { return u.processNewEvent(ctx, evt) })
	return ts, evt, err
}

func (u *ExUniverse) StartOnReality(ctx context.Context, realityName string) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	if _, ok := u.model.Realities[realityName]; !ok {
		return nil, nil, fmt.Errorf("reality '%s' does not exist", realityName)
	}

	if err := u.establishNewReality(ctx, &realityName, nil); err != nil {
		return nil, nil, err
	}

	evt := &event{EvtType: EventTypeBigBangOn}

	ts, err := u.externalTransitionsDecorator(func() error { return u.processNewEvent(ctx, evt) })
	return ts, evt, err
}

func (u *ExUniverse) SendEvent(ctx context.Context, evt Event) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	ts, err := u.externalTransitionsDecorator(func() error { return u.processNewEvent(ctx, evt) })
	return ts, evt, err
}

func (u *ExUniverse) SendEventToReality(ctx context.Context, realityName string, evt Event) ([]string, Event, error) {
	u.universeMtx.Lock()
	defer u.universeMtx.Unlock()

	if _, ok := u.model.Realities[realityName]; !ok {
		return nil, nil, fmt.Errorf("reality '%s' does not exist", realityName)
	}

	ts, err := u.externalTransitionsDecorator(func() error {
		if u.inSuperposition {
			if newReality, err := u.executeEventOnSuperimposedReality(ctx, &realityName, evt); !newReality {
				return err
			}
		}
		if *u.currentReality != realityName {
			return fmt.Errorf("reality '%s' is not the current reality", realityName)
		}
		return u.processEventOnEstablishedReality(ctx, evt)
	})

	return ts, evt, err
}

func (u *ExUniverse) GetSnapshot() ExUniverseSnapshot {
	var ss = universeInfoSnapshot{
		ID:                 u.id,
		Initialized:        u.initialized,
		CurrentReality:     u.currentReality,
		RealityInitialized: u.realityInitialized,
		InSuperposition:    u.inSuperposition,
		Accumulator:        u.eventAccumulator,
	}

	m, _ := devtoolkit.StructToMap(ss)
	return m
}

func (u *ExUniverse) LoadSnapshot(snapshotMap ExUniverseSnapshot) error {
	snapshot, err := devtoolkit.MapToStruct[universeInfoSnapshot](snapshotMap)
	if err != nil {
		return errors.Join(fmt.Errorf("error loading snapshot for universe '%s'", u.id), err)
	}

	u.id = snapshot.ID
	u.initialized = snapshot.Initialized
	u.currentReality = snapshot.CurrentReality
	u.realityInitialized = snapshot.RealityInitialized
	u.inSuperposition = snapshot.InSuperposition
	u.eventAccumulator = snapshot.Accumulator

	if u.currentReality != nil {
		realityModel := u.model.GetReality(*u.currentReality)
		if realityModel == nil {
			return fmt.Errorf("reality '%s' does not exist", *u.currentReality)
		}
		u.finalReality = theoretical.IsFinalState(realityModel.Type)
	}

	return nil
}

//----------------------------------

func (u *ExUniverse) externalTransitionsDecorator(operation func() error) ([]string, error) {
	// clear externalTargets
	u.externalTargets = nil

	// execute operation
	if err := operation(); err != nil {
		return nil, err
	}

	return u.externalTargets, nil
}

func (u *ExUniverse) processNewEvent(ctx context.Context, evt Event) error {
	if u.inSuperposition {
		if newReality, err := u.processEventOnSuperposition(ctx, evt); !newReality {
			return err
		}
		// if new reality is established then continue processing the event on the established reality (the new reality)
	}
	return u.processEventOnEstablishedReality(ctx, evt)
}

func (u *ExUniverse) processEventOnSuperposition(ctx context.Context, evt Event) (bool, error) {
	// get candidate realities from accumulator
	candidates := u.eventAccumulator.GetActiveRealities()

	for _, candidate := range candidates {
		newReality, err := u.executeEventOnSuperimposedReality(ctx, &candidate, evt)
		if !newReality && err != nil {
			return false, err
		}

		if newReality {
			return true, nil
		}
	}

	return false, nil
}

func (u *ExUniverse) executeEventOnSuperimposedReality(ctx context.Context, realityName *string, evt Event) (bool, error) {
	// if realityName is not nil send to the given reality
	realityModel := u.model.GetReality(*realityName)

	// accumulate event
	u.eventAccumulator.Accumulate(*realityName, evt)

	// execute observers
	isNewReality, err := u.executeObservers(ctx, realityModel, evt)

	// false && nil -> return nil
	if !isNewReality && err == nil {
		return false, nil
	}

	// false && err -> return err
	if !isNewReality && err != nil {
		return false, errors.Join(fmt.Errorf("error executing observers for reality '%s' in universe '%s'", *realityName, u.id), err)
	}

	// true -> : establish new reality
	// err will be ignored
	if e := u.establishNewReality(ctx, realityName, evt); e != nil {
		return false, errors.Join(fmt.Errorf("error establishing new reality '%s' in universe '%s'", *realityName, u.id), e)
	}

	return true, nil
}

func (u *ExUniverse) processEventOnEstablishedReality(ctx context.Context, evt Event) error {
	realityModel := u.model.GetReality(*u.currentReality)

	// execute always transitions
	approvedTransition, err := u.executeAlways(ctx, realityModel, evt)
	if err != nil {
		return err
	}

	if approvedTransition != nil {
		return u.processApprovedTransition(ctx, approvedTransition, evt)
	}

	// if no always transitions are approved then execute on transitions
	return u.executeOnEntryProcess(ctx, evt)
}

func (u *ExUniverse) processApprovedTransition(ctx context.Context, approvedTransition *theoretical.TransitionModel, evt Event) error {
	if len(approvedTransition.Targets) == 1 {
		refT, parts, err := getReferenceType(approvedTransition.Targets[0])
		if err != nil {
			return errors.Join(fmt.Errorf("error getting reference type for target '%s' in universe '%s'", approvedTransition.Targets[0], u.id), err)
		}

		if refT == RefTypeReality {
			return u.establishNewReality(ctx, &parts[0], evt)
		}
	}

	return u.establishNewSuperposition(ctx, approvedTransition.Targets, evt)
}

func (u *ExUniverse) establishNewReality(ctx context.Context, realityName *string, evt Event) error {
	// u.currentReality != nil means that the universe has an established reality and now is going to be exited
	if err := u.executeOnExitProcess(ctx, evt); err != nil {
		return errors.Join(fmt.Errorf("error executing on exit process for reality '%s' in universe '%s'", *u.currentReality, u.id), err)
	}

	realityModel := u.model.GetReality(*realityName)

	// update reality
	u.initialized = true
	u.currentReality = realityName
	u.finalReality = theoretical.IsFinalState(realityModel.Type)
	u.inSuperposition = false
	u.realityInitialized = false
	u.externalTargets = nil

	return nil
}

func (u *ExUniverse) establishNewSuperposition(ctx context.Context, targets []string, evt Event) error {
	if !u.finalReality {
		u.currentReality = nil

		// u.currentReality != nil means that the universe has an established reality and now is going to be exited
		if err := u.executeOnExitProcess(ctx, evt); err != nil {
			return errors.Join(fmt.Errorf("error executing on exit process for reality '%s' in universe '%s'", *u.currentReality, u.id), err)
		}
	}

	u.externalTargets = targets
	u.inSuperposition = !u.finalReality
	u.eventAccumulator.Clear()

	return nil
}

//----------------------------------

func (u *ExUniverse) executeOnEntryProcess(ctx context.Context, evt Event) error {
	if u.currentReality == nil {
		return nil
	}

	realityModel := u.model.GetReality(*u.currentReality)

	// execute entry actions
	if err := u.executeEntryActions(ctx, realityModel, evt); err != nil {
		return err
	}

	// launch entry invoke
	u.executeEntryInvokes(ctx, realityModel, evt)

	// execute on transitions
	approvedTransitions, err := u.executeOn(ctx, realityModel, evt)
	if err != nil {
		return err
	}

	// process new targets
	return u.processApprovedTransition(ctx, approvedTransitions, evt)
}

func (u *ExUniverse) executeOnExitProcess(ctx context.Context, evt Event) error {
	if u.currentReality == nil {
		return nil
	}

	realityModel := u.model.GetReality(*u.currentReality)

	// execute exit actions
	if err := u.executeExitActions(ctx, realityModel, evt); err != nil {
		return err
	}

	// launch exit invoke
	u.executeExitInvokes(ctx, realityModel, evt)

	return nil
}

func (u *ExUniverse) executeAlways(ctx context.Context, realityModel *theoretical.RealityModel, evt Event) (*theoretical.TransitionModel, error) {
	approvedTransition, err := u.getApprovedTransition(ctx, realityModel.Always, evt)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error executing always transitions for reality '%s' in universe '%s'", realityModel.ID, u.id), err)
	}
	u.realityInitialized = true
	return approvedTransition, err
}

func (u *ExUniverse) executeObservers(ctx context.Context, realityModel *theoretical.RealityModel, evt Event) (bool, error) {
	// context & cancel function for observers
	var cancelCtx, cancelCtxFunc = context.WithCancel(ctx)
	defer cancelCtxFunc()

	// buffered channels
	var newRealityChan = make(chan bool, 1)
	var errChan = make(chan error, 1)

	// execute observers
	var wg sync.WaitGroup
	for _, observer := range realityModel.Observers {
		wg.Add(1)
		go func(observerModel *theoretical.ObserverModel) {
			defer wg.Done()
			nr, e := u.executeObserver(cancelCtx, observerModel, evt)

			if e != nil && len(errChan) == 0 {
				select {
				case errChan <- e:
				default:
				}
			}

			if nr && len(newRealityChan) == 0 {
				select {
				case newRealityChan <- true:
					cancelCtxFunc()
				default:
				}
			}

		}(observer)
	}

	// wait for all observers to finish
	wg.Wait()

	// close channels
	close(newRealityChan)
	close(errChan)

	// new reality flags
	var newReality bool
	var err error

	// check if there is an error
	select {
	case err = <-errChan:
	default:
	}

	// check if there is a new reality
	select {
	case newReality = <-newRealityChan:
	default:
	}

	return newReality, err
}

func (u *ExUniverse) executeEntryInvokes(ctx context.Context, realityModel *theoretical.RealityModel, evt Event) {
	u.executeInvokes(ctx, realityModel.EntryInvokes, evt)
}

func (u *ExUniverse) executeExitInvokes(ctx context.Context, realityModel *theoretical.RealityModel, evt Event) {
	u.executeInvokes(ctx, realityModel.ExitInvokes, evt)
}

func (u *ExUniverse) executeInvokes(ctx context.Context, invokeModels []*theoretical.InvokeModel, evt Event) {
	for _, invokeModel := range invokeModels {
		go u.executeInvoke(ctx, invokeModel, evt)
	}
}

func (u *ExUniverse) executeEntryActions(ctx context.Context, realityModel *theoretical.RealityModel, evt Event) error {
	if err := u.executeActions(ctx, realityModel.EntryActions, evt); err != nil {
		return errors.Join(fmt.Errorf("error executing entry actions for reality '%s' in universe '%s'", realityModel.ID, u.id), err)
	}
	return nil
}

func (u *ExUniverse) executeExitActions(ctx context.Context, realityModel *theoretical.RealityModel, evt Event) error {
	if err := u.executeActions(ctx, realityModel.ExitActions, evt); err != nil {
		return errors.Join(fmt.Errorf("error executing exit actions for reality '%s' in universe '%s'", realityModel.ID, u.id), err)
	}
	return nil
}

func (u *ExUniverse) executeActions(ctx context.Context, actionModels []*theoretical.ActionModel, evt Event) error {
	for _, actionModel := range actionModels {
		if err := u.executeAction(ctx, actionModel, evt); err != nil {
			return errors.Join(fmt.Errorf("error executing action '%s'", actionModel.Src), err)
		}
	}
	return nil
}

func (u *ExUniverse) executeOn(ctx context.Context, realityModel *theoretical.RealityModel, evt Event) (*theoretical.TransitionModel, error) {
	transitions, ok := realityModel.On[evt.GetEventName()]
	if !ok {
		return nil, nil
	}
	approvedTransition, err := u.getApprovedTransition(ctx, transitions, evt)
	if err == nil {
		return nil, errors.Join(fmt.Errorf("error executing on transitions for reality '%s' in universe '%s'", realityModel.ID, u.id), err)
	}

	return approvedTransition, err
}

func (u *ExUniverse) getApprovedTransition(ctx context.Context, transitionModels []*theoretical.TransitionModel, evt Event) (*theoretical.TransitionModel, error) {

	for _, transition := range transitionModels {
		doTransition, e := u.executeCondition(ctx, transition.Condition, evt)
		if e != nil {
			return nil, errors.Join(fmt.Errorf("error executing always transition condition '%s' in universe '%s'", transition.Condition.Src, u.id), e)
		}

		if doTransition {
			return transition, nil
		}
	}

	return nil, nil
}

//----------------------------------

func (u *ExUniverse) executeTransition(ctx context.Context, transitionModel *theoretical.TransitionModel, evt Event) (bool, error) {
	doTransition, err := u.executeCondition(ctx, transitionModel.Condition, evt)

	if err != nil {
		return false, errors.Join(fmt.Errorf("error executing transition condition '%s' in universe '%s'", transitionModel.Condition.Src, u.id), err)
	}

	return doTransition, nil
}

func (u *ExUniverse) executeObserver(ctx context.Context, observerModel *theoretical.ObserverModel, evt Event) (bool, error) {
	return u.laws.ExecuteObserver(
		ctx,
		observerModel.Src,
		observerModel.Args,
		u.universeContext,
		evt,
		u.eventAccumulator.GetStatistics(),
	)
}

func (u *ExUniverse) executeAction(ctx context.Context, actionModel *theoretical.ActionModel, evt Event) error {
	return u.laws.ExecuteAction(
		ctx,
		actionModel.Src,
		actionModel.Args,
		u.universeContext,
		evt,
	)
}

func (u *ExUniverse) executeInvoke(ctx context.Context, invokeModel *theoretical.InvokeModel, evt Event) {
	u.laws.ExecuteInvoke(
		ctx,
		invokeModel.Src,
		invokeModel.Args,
		u.universeContext,
		evt,
	)
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
