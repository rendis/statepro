package piece

import (
	"sync"
)

type GMachine[ContextType any] struct {
	Id         string
	EntryState *GState[ContextType]
	States     map[string]*GState[ContextType]
}

type ProMachine[ContextType any] interface {
	PlaceOn(stateName string) error
	SendEvent(event Event) TransitionResponse
	GetNextEvents() []string
	GetState() string
	IsFinalState() bool
	GetContext() ContextType
}

func NewProMachine[ContextType any](machine *GMachine[ContextType], context *ContextType) ProMachine[ContextType] {
	return &proMachineImpl[ContextType]{
		context:      context,
		gMachine:     machine,
		currentState: machine.EntryState,
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
}

func (pm *proMachineImpl[ContextType]) PlaceOn(stateName string) error {
	pm.pmMtx.Lock()
	defer pm.pmMtx.Unlock()

	if s, ok := pm.gMachine.States[stateName]; ok {
		pm.currentState = s
		return nil
	}
	return &EventNotFountError{EventName: stateName}
}

func (pm *proMachineImpl[ContextType]) SendEvent(event Event) TransitionResponse {
	pm.pmMtx.Lock()
	defer func() {
		pm.processing = false
		pm.pmMtx.Unlock()
	}()
	pm.processing = true

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

	// get current event to check if an action in onEvent has been changed the event
	evtFilled, doOnEvent := pm.getCurrentEvent()

	// while target != nil
	for target != nil && *target != *pm.currentState.Name {

		// if doOnEvent == true => an action has been changed the event
		if doOnEvent {
			target, err = pm.currentState.onEvent(*pm.context, evtFilled, pm)
			if err != nil {
				ce, _ := pm.getCurrentEvent()
				return &transitionResponse{lastEvent: ce, err: err}
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
			return &transitionResponse{lastEvent: evtFilled, err: err}
		}

		evtFilled, doOnEvent = pm.getCurrentEvent()
		if doOnEvent {
			continue
		}
		target, _, err = pm.currentState.onEntry(*pm.context, evtFilled, pm)

		if err != nil {
			ce, _ := pm.getCurrentEvent()
			return &transitionResponse{lastEvent: ce, err: err}
		}
	}

	ce, _ := pm.getCurrentEvent()
	return &transitionResponse{
		lastEvent: ce,
	}
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

func (pm *proMachineImpl[ContextType]) setCurrenEvent(event Event) {
	pm.evtMtx.Lock()
	defer pm.evtMtx.Unlock()
	evtCasted, _ := event.(*GEvent)
	evtCasted.from = *pm.currentState.Name
	pm.currentEvent = evtCasted
	pm.eventChanged = true
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
	Assign(context ContextType)
	Send(event Event)
}

func (pm *proMachineImpl[ContextType]) Assign(context ContextType) {
	pm.ctxMtx.Lock()
	defer pm.ctxMtx.Unlock()
	pm.context = &context
}

func (pm *proMachineImpl[ContextType]) Send(event Event) {
	pm.setCurrenEvent(event)
}

type TransitionResponse interface {
	LastEvent() Event
	Error() error
}

type transitionResponse struct {
	respCh    chan Event
	err       error
	lastEvent Event
}

func (t *transitionResponse) LastEvent() Event {
	return t.lastEvent
}

func (t *transitionResponse) Error() error {
	return t.err
}
