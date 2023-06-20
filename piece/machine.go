package piece

import (
	"sync"
)

type ContextFromSourceFnDefinition[ContextType any] func(params ...any) (*ContextType, error)
type ContextToSourceFnDefinition[ContextType any] func(ContextType) error

type GMachine[ContextType any] struct {
	Id          string
	EntryState  *GState[ContextType]
	States      map[string]*GState[ContextType]
	SuccessFlow []string
	Version     string
}

type ProMachine[ContextType any] interface {
	PlaceOn(stateName string) error
	StartOn(stateName string) TransitionResponse
	StartOnWithEvent(stateName string, event Event) TransitionResponse
	SendEvent(event Event) TransitionResponse
	GetNextEvents() []string
	GetState() string
	IsFinalState() bool
	GetContext() ContextType
	CallContextToSource() error
}

func NewProMachine[ContextType any](
	machine *GMachine[ContextType],
	context *ContextType,
	contextFromSourceFn ContextFromSourceFnDefinition[ContextType],
	contextToSourceFn ContextToSourceFnDefinition[ContextType],
) ProMachine[ContextType] {
	return &proMachineImpl[ContextType]{
		context:             context,
		gMachine:            machine,
		currentState:        machine.EntryState,
		contextFromSourceFn: contextFromSourceFn,
		contextToSourceFn:   contextToSourceFn,
	}
}

type proMachineImpl[ContextType any] struct {
	pmMtx      sync.Mutex
	gMachine   *GMachine[ContextType]
	processing bool

	ctxMtx  sync.RWMutex
	context *ContextType

	evtMtx       sync.Mutex
	currentEvent *GEvent
	eventChanged bool

	currentState *GState[ContextType]
	prevState    *GState[ContextType]

	contextFromSourceFn ContextFromSourceFnDefinition[ContextType]
	contextToSourceFn   ContextToSourceFnDefinition[ContextType]
}

func (pm *proMachineImpl[ContextType]) PlaceOn(stateName string) error {
	pm.pmMtx.Lock()
	defer func() {
		pm.processing = false
		pm.pmMtx.Unlock()
	}()
	pm.processing = true

	// get new state from machine and set it as current state
	if s, ok := pm.gMachine.States[stateName]; ok {
		pm.currentState = s
		return nil
	}
	return &StateNotFountError{EventName: stateName}
}

func (pm *proMachineImpl[ContextType]) StartOn(stateName string) TransitionResponse {
	// build event
	var evtFilled = &GEvent{
		from:    *pm.currentState.Name,
		evtType: EventTypeStartOn,
	}

	return pm.StartOnWithEvent(stateName, evtFilled)
}

func (pm *proMachineImpl[ContextType]) StartOnWithEvent(stateName string, event Event) TransitionResponse {
	// lock processing
	pm.pmMtx.Lock()
	defer func() {
		pm.processing = false
		pm.pmMtx.Unlock()
	}()
	pm.processing = true
	pm.eventChanged = false

	// get new state from machine and set it as current state
	if s, ok := pm.gMachine.States[stateName]; ok {
		pm.currentState = s
	} else {
		return &transitionResponse{lastEvent: event, err: &StateNotFountError{EventName: stateName}}
	}

	// set current event from value
	var evtFilled = event.(*GEvent)
	evtFilled.from = *pm.currentState.Name
	pm.currentEvent = evtFilled

	// running first onEntry
	target, _, err := pm.currentState.onEntry(*pm.context, evtFilled, pm)
	if err != nil {
		ce, _ := pm.getCurrentEvent()
		return &transitionResponse{lastEvent: ce, err: err}
	}

	// do transitions until target == nil
	err = pm.doTransitions(target)

	// build response
	ce, _ := pm.getCurrentEvent()
	return &transitionResponse{
		lastEvent: ce,
		err:       err,
	}
}

func (pm *proMachineImpl[ContextType]) SendEvent(event Event) TransitionResponse {
	// lock processing
	pm.pmMtx.Lock()
	defer func() {
		pm.processing = false
		pm.pmMtx.Unlock()
	}()
	pm.processing = true
	pm.eventChanged = false

	// set event from value
	var evtFilled = event.(*GEvent)
	evtFilled.from = *pm.currentState.Name
	pm.currentEvent = evtFilled

	// running first onEvent
	target, err := pm.currentState.onEvent(*pm.context, evtFilled, pm)
	if err != nil {
		ce, _ := pm.getCurrentEvent()
		return &transitionResponse{lastEvent: ce, err: err}
	}

	// do transitions until target == nil
	err = pm.doTransitions(target)

	// build response
	ce, _ := pm.getCurrentEvent()
	return &transitionResponse{
		lastEvent: ce,
		err:       err,
	}
}

func (pm *proMachineImpl[ContextType]) doTransitions(target *string) (err error) {
	// get current event to check if an action in onEvent/onEntry has been changed the event
	evtFilled, doOnEvent := pm.getCurrentEvent()

	// while target != nil
	for target != nil && *target != *pm.currentState.Name {

		// if doOnEvent == true => an action has been changed the event
		if doOnEvent {
			target, err = pm.currentState.onEvent(*pm.context, evtFilled, pm)
			if err != nil {
				return err
			}

			// get current event to check if an action in onEvent has been changed the event
			evtFilled, doOnEvent = pm.getCurrentEvent()
			continue
		}

		// update previous and current state
		pm.prevState = pm.currentState
		pm.currentState = pm.gMachine.States[*target]

		evtFilled, doOnEvent = pm.getCurrentEvent()
		// if doOnEvent == true => an action in onEvent has been changed the event
		if doOnEvent {
			continue
		}
		if err = pm.prevState.execExit(*pm.context, evtFilled, pm); err != nil {
			return err
		}

		evtFilled, doOnEvent = pm.getCurrentEvent()
		if doOnEvent {
			continue
		}

		target, _, err = pm.currentState.onEntry(*pm.context, evtFilled, pm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *proMachineImpl[ContextType]) GetNextEvents() []string {
	return pm.currentState.getNextEvents()
}

func (pm *proMachineImpl[ContextType]) GetState() string {
	return *pm.currentState.Name
}

func (pm *proMachineImpl[ContextType]) IsFinalState() bool {
	return pm.currentState.isFinalState()
}

func (pm *proMachineImpl[ContextType]) GetContext() ContextType {
	pm.ctxMtx.RLock()
	defer pm.ctxMtx.RUnlock()
	return *pm.context
}

func (pm *proMachineImpl[ContextType]) CallContextToSource() error {
	if pm.contextToSourceFn == nil {
		return ContextToSourceNotImplementedError
	}
	return pm.contextToSourceFn(*pm.context)
}

func (pm *proMachineImpl[ContextType]) setCurrenEvent(event Event, eventChanged bool) {
	pm.evtMtx.Lock()
	defer pm.evtMtx.Unlock()
	evtCasted, _ := event.(*GEvent)
	evtCasted.from = *pm.currentState.Name
	*pm.currentEvent = *evtCasted
	pm.eventChanged = eventChanged
}

func (pm *proMachineImpl[ContextType]) getCurrentEvent() (*GEvent, bool) {
	pm.evtMtx.Lock()
	defer func() {
		pm.eventChanged = false
		pm.evtMtx.Unlock()
	}()
	return pm.currentEvent, pm.eventChanged
}

type ActionTool[ContextType any] interface {
	Assign(context ContextType) // assign context to machine
	Send(event Event)           // send event to current state
	Propagate(event Event)      // propagate event with new data and error
}

func (pm *proMachineImpl[ContextType]) Assign(context ContextType) {
	pm.ctxMtx.Lock()
	defer pm.ctxMtx.Unlock()
	pm.context = &context
}

func (pm *proMachineImpl[ContextType]) Send(event Event) {
	pm.setCurrenEvent(event, true)
}

func (pm *proMachineImpl[ContextType]) Propagate(event Event) {
	pm.setCurrenEvent(event, false)
}

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
