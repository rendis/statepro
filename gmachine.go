package statepro

import (
	"context"
	"sync"
)

type gMachine[ContextType any] struct {
	id      string
	xm      *XMachine
	context *ContextType

	// states
	states          map[string]*gState[ContextType]
	entryState      *gState[ContextType]
	currentState    *gState[ContextType]
	prevState       *gState[ContextType]
	inEventsByState map[string]map[string]struct{}

	// for processing
	pmMtx      sync.Mutex
	processing bool

	// for event
	evtMtx       sync.Mutex
	currentEvent *gEvent
	eventChanged bool

	// source handlers
	contextToSourceFn     ProMachineToSourceHandler[ContextType]
	toSourceExecutionMode ExecutionType

	// final states
	finalStates []string
}

// PlaceOn initialize the machine with the stateName as current state, without executing onEntry
func (pm *gMachine[ContextType]) PlaceOn(stateName string) error {
	// lock processing
	pm.pmMtx.Lock()
	defer func() {
		pm.processing = false
		pm.pmMtx.Unlock()
	}()
	pm.processing = true

	// get new state from machine and set it as current state
	if s, ok := pm.states[stateName]; ok {
		pm.currentState = s
		return nil
	}
	return &StateNotFountError{EventName: stateName}
}

func (pm *gMachine[ContextType]) StartOn(ctx context.Context, stateName string) TransitionResponse {
	// build event
	var evtFilled = &gEvent{
		evtType: EventTypeStartOn,
	}

	return pm.StartOnWithEvent(ctx, stateName, evtFilled)
}

func (pm *gMachine[ContextType]) Start(ctx context.Context) TransitionResponse {
	// build event
	var evtFilled = &gEvent{
		evtType: EventTypeStart,
	}

	return pm.StartOnWithEvent(ctx, *pm.entryState.Name, evtFilled)
}

func (pm *gMachine[ContextType]) StartOnWithEvent(ctx context.Context, stateName string, event Event) TransitionResponse {
	// lock processing
	pm.pmMtx.Lock()
	defer func() {
		pm.processing = false
		pm.pmMtx.Unlock()
	}()
	pm.processing = true
	pm.eventChanged = false

	// get new state from machine and set it as current state
	if s, ok := pm.states[stateName]; ok {
		pm.currentState = s
	} else {
		return &transitionResponse{lastEvent: event, err: &StateNotFountError{EventName: stateName}}
	}

	// set current event from value
	var newEvent = event.(*gEvent)
	newEvent.machineInfo = pm
	pm.currentEvent = newEvent

	// call contextToSource on entry
	if err := pm.callContextToSourceOnEntry(ctx); err != nil {
		return &transitionResponse{lastEvent: event, err: err}
	}

	// running first onEntry
	target, _, err := pm.currentState.onEntry(ctx, pm.context, newEvent, pm)
	if err != nil {
		ce, _ := pm.getCurrentEvent()
		return &transitionResponse{lastEvent: ce, err: err}
	}

	// do transitions until target == nil
	err = pm.doTransitions(ctx, target)

	// build response
	ce, _ := pm.getCurrentEvent()
	return &transitionResponse{
		lastEvent: ce,
		err:       err,
	}
}

func (pm *gMachine[ContextType]) SendEvent(ctx context.Context, event Event) TransitionResponse {
	// lock processing
	pm.pmMtx.Lock()
	defer func() {
		pm.processing = false
		pm.pmMtx.Unlock()
	}()
	pm.processing = true
	pm.eventChanged = false

	// set event from value
	var newEvent = event.(*gEvent)
	newEvent.machineInfo = pm
	pm.currentEvent = newEvent

	// running first onEvent
	target, err := pm.currentState.onEvent(ctx, pm.context, newEvent, pm)
	if err != nil {
		ce, _ := pm.getCurrentEvent()
		return &transitionResponse{lastEvent: ce, err: err}
	}

	// do transitions until target == nil
	err = pm.doTransitions(ctx, target)

	// build response
	ce, _ := pm.getCurrentEvent()
	return &transitionResponse{
		lastEvent: ce,
		err:       err,
	}
}

func (pm *gMachine[ContextType]) GetNextEvents() []string {
	return pm.currentState.getNextEvents()
}

func (pm *gMachine[ContextType]) GetNextTargets() []string {
	return pm.currentState.getNextTargets()
}

func (pm *gMachine[ContextType]) ContainsTarget(target string) bool {
	return pm.currentState.containsTarget(target)
}

func (pm *gMachine[ContextType]) GetState() string {
	return *pm.currentState.Name
}

func (pm *gMachine[ContextType]) IsFinalState() bool {
	return pm.currentState.isFinalState()
}

func (pm *gMachine[ContextType]) GetContext() *ContextType {
	return pm.context
}

func (pm *gMachine[ContextType]) SetContext(machineCtx *ContextType) {
	pm.pmMtx.Lock()
	defer pm.pmMtx.Unlock()
	pm.context = machineCtx
}

func (pm *gMachine[ContextType]) CallContextToSource(ctx context.Context) error {
	if pm.contextToSourceFn == nil {
		return ContextToSourceNotImplementedError
	}
	stateName := *pm.currentState.Name
	return pm.contextToSourceFn.ContextToSource(ctx, stateName, pm.context)
}

func (pm *gMachine[ContextType]) GetSuccessFlow() []string {
	successFlow := make([]string, len(pm.xm.SuccessFlow))
	for i, s := range pm.xm.SuccessFlow {
		successFlow[i] = s
	}
	return successFlow
}

func (pm *gMachine[ContextType]) setNextState(ctx context.Context, target string) error {
	// call contextToSource on exit, if needed
	if err := pm.callContextToSourceOnExit(ctx); err != nil {
		return err
	}

	// switch to new state
	pm.prevState = pm.currentState
	pm.currentState = pm.states[target]

	// call contextToSource on entry, if needed
	return pm.callContextToSourceOnEntry(ctx)
}

func (pm *gMachine[ContextType]) callContextToSourceOnEntry(ctx context.Context) error {
	if pm.toSourceExecutionMode != ExecutionTypeOnEntry && pm.toSourceExecutionMode != ExecutionTypeAlways {
		return nil
	}
	stateName := *pm.currentState.Name
	return pm.contextToSourceFn.ContextToSource(ctx, stateName, pm.context)
}

func (pm *gMachine[ContextType]) callContextToSourceOnExit(ctx context.Context) error {
	if pm.toSourceExecutionMode != ExecutionTypeOnExit && pm.toSourceExecutionMode != ExecutionTypeAlways {
		return nil
	}
	stateName := *pm.currentState.Name
	return pm.contextToSourceFn.ContextToSource(ctx, stateName, pm.context)
}

func (pm *gMachine[ContextType]) doTransitions(ctx context.Context, target *string) (err error) {
	// check if an action in [onEvent, onEntry] has been changed the event
	var currentEvent, eventChanged = pm.getCurrentEvent()

	// while target != nil
	for target != nil {

		// if eventChanged, call onEvent over current state
		if eventChanged {
			target, err = pm.currentState.onEvent(ctx, pm.context, currentEvent, pm)
			if err != nil {
				return err
			}

			// check if an action in [onEvent] has been changed the event
			currentEvent, eventChanged = pm.getCurrentEvent()
			continue
		}

		// running onExit
		if err = pm.currentState.execExit(ctx, pm.context, currentEvent, pm); err != nil {
			return err
		}

		// check if an action in [execExit] has been changed the event
		currentEvent, eventChanged = pm.getCurrentEvent()
		if eventChanged {
			continue
		}

		// update previous and current state
		if err := pm.setNextState(ctx, *target); err != nil {
			return err
		}

		// running onEntry
		target, _, err = pm.currentState.onEntry(ctx, pm.context, currentEvent, pm)
		if err != nil {
			return err
		}

		// get current event to check if an action in [onEntry] has been changed the event
		currentEvent, eventChanged = pm.getCurrentEvent()
	}

	return nil
}

func (pm *gMachine[ContextType]) setCurrenEvent(event Event, eventChanged bool) {
	pm.evtMtx.Lock()
	defer pm.evtMtx.Unlock()
	evtCasted, _ := event.(*gEvent)
	*pm.currentEvent = *evtCasted
	pm.eventChanged = eventChanged
}

func (pm *gMachine[ContextType]) getCurrentEvent() (*gEvent, bool) {
	pm.evtMtx.Lock()
	defer func() {
		pm.eventChanged = false
		pm.evtMtx.Unlock()
	}()
	return pm.currentEvent, pm.eventChanged
}

func (pm *gMachine[ContextType]) GetInEventsForCurrentState() []string {
	var inEvents []string
	for event := range pm.inEventsByState[*pm.currentState.Name] {
		inEvents = append(inEvents, event)
	}
	return inEvents
}

func (pm *gMachine[ContextType]) InEventOnCurrentState(event string) bool {
	_, ok := pm.inEventsByState[*pm.currentState.Name][event]
	return ok
}

// MachineBasicInfo interface to get basic info from a machine
type MachineBasicInfo interface {
	GetFromState() string     // get state name from which event is triggered
	GetCurrentState() string  // get state name to which event is triggered
	GetFinalStates() []string // get final states
}

func (pm *gMachine[ContextType]) GetFromState() string {
	return *pm.prevState.Name
}

func (pm *gMachine[ContextType]) GetCurrentState() string {
	return *pm.currentState.Name
}

func (pm *gMachine[ContextType]) GetFinalStates() []string {
	newFinalStates := make([]string, len(pm.finalStates))
	copy(newFinalStates, pm.finalStates)
	return newFinalStates
}

// ActionTool is the interface to send and propagate events
type ActionTool[ContextType any] interface {
	Send(event Event) // send new event to current state
	AssignContext(context ContextType)
}

func (pm *gMachine[ContextType]) Send(event Event) {
	var newEvent = event.(*gEvent)
	newEvent.machineInfo = pm
	pm.setCurrenEvent(event, true)
}

func (pm *gMachine[ContextType]) AssignContext(context ContextType) {
	pm.context = &context
}

// TransitionResponse is the response of a transition
type TransitionResponse interface {
	GetLastEvent() Event
	Error() error
}

type transitionResponse struct {
	respCh    chan Event
	err       error
	lastEvent Event
}

func (t *transitionResponse) GetLastEvent() Event {
	return t.lastEvent
}

func (t *transitionResponse) Error() error {
	return t.err
}
