package piece

import "log"

/*
type MachineBuilder[ContextType any] interface {
	WithContext(context ContextType) MachineBuilder[ContextType]
	WithState(stateName string) MachineBuilder[ContextType]
	Build() *GMachine[ContextType]
}

type MachineBuilderImpl[ContextType any] struct {
	machine *GMachine[ContextType]
}

func (b *MachineBuilderImpl[ContextType]) WithContext(context ContextType) MachineBuilder[ContextType] {
	b.machine.Context = &context
	return b
}

func (b *MachineBuilderImpl[ContextType]) WithState(stateName string) MachineBuilder[ContextType] {
	if s, ok := b.machine.States[stateName]; ok {
		b.machine.CurrentState = s
		return b
	}
	panic(fmt.Sprintf("GState '%s' not found", stateName))
}

func (b *MachineBuilderImpl[ContextType]) Build() *GMachine[ContextType] {
	return b.machine
}
*/

//------------------------------------------------------------------------------

type GMachine[ContextType any] struct {
	Id           string
	Context      *ContextType
	EntryState   *GState[ContextType]
	CurrentState *GState[ContextType]
	States       map[string]*GState[ContextType]
}

type ActionTool[ContextType any] interface {
	Assign(context ContextType)
	Send(event Event)
	Raise(event Event)
}

func (m *GMachine[ContextType]) Assign(context ContextType) {
	// m.Context = &context
	log.Println("Assign")
}

func (m *GMachine[ContextType]) Send(event Event) {
	log.Println("Send")
	// TODO: Prevent multiple sends
	/*
		if m.CurrentState == nil {
			panic("GMachine is not initialized")
		}
		evt := &GEvent{
			name:    eventName,
			evtType: EventTypeTransitional,
		}
		m.CurrentState.onEvent(m.Context, evt, m)
	*/
}

func (m *GMachine[ContextType]) Raise(event Event) {
	log.Println("Raise")
}

func (m *GMachine[ContextType]) SendEvent(event Event) TransitionResponse[ContextType] {
	return nil
}

func (m *GMachine[ContextType]) PlaceOn(state ProState, context ContextType) ProMachine[ContextType] {
	return m
}

type ProState interface {
	GetState() string
}

type TransitionResponse[ContextType any] interface {
	GetContext() ContextType
	GetState() ProState
	GetEvent() Event
}

type ProMachine[ContextType any] interface {
	SendEvent(event Event) TransitionResponse[ContextType]
	PlaceOn(state ProState, context ContextType) ProMachine[ContextType]
}

type transitionResp[ContextType any] struct {
	respCh  chan Event
	context *ContextType
}

func (t *transitionResp[ContextType]) GetContext() ContextType {
	select {
	case resp, ok := <-t.respCh:
		if ok {
			close(t.respCh)
			t.processResponse(resp)
		}
	}
	return *t.context
}

func (t *transitionResp[ContextType]) processResponse(evt Event) {

}

/*
type GSupplier[ContextType any] interface {
	getAction(n string) (TAction[ContextType], ActionTool[ContextType])
	getGuard(n string) TPredicate[ContextType]
	getService(n string) TInvocation[ContextType]
}

func (m *GMachine[ContextType]) getAction(n string) (TAction[ContextType], ActionTool[ContextType]) {
	return nil, nil
}

func (m *GMachine[ContextType]) getGuard(n string) TPredicate[ContextType] {
	return nil
}

func (m *GMachine[ContextType]) getService(n string) TInvocation[ContextType] {
	return nil
}
*/

//type Assigner[ContextType any] func(context ContextType)
//type Send func(eventName string)

//func (m *GMachine[ContextType]) WithContext(c ContextType) statepro.MachineBuilder[ContextType] {
//	m.Context = &c
//	return m
//}
//
//func (m *GMachine[ContextType]) WithState(stateName string) statepro.MachineBuilder[ContextType] {
//	if s, ok := m.States[stateName]; ok {
//		m.CurrentState = s
//		return m
//	}
//	panic(fmt.Sprintf("GState '%s' not found", stateName))
//}
//
//func (m *GMachine[ContextType]) Start() statepro.StatePro[ContextType] {
//	return nil
//}

//type ActionTool[ContextType any] struct {
//	Assign Assigner[ContextType]
//	Send   Send
//}

/*
func (m *GMachine[ContextType]) StartOn(target string, c ContextType) {
	if s, ok := m.States[target]; ok {
		m.Context = &c
		m.CurrentState = s
		m.CurrentState.onEntry(*m.Context, Event{Data: nil, Err: nil, EvtType: EventTypeTransitional})
		return
	}
	fmt.Printf("Error: GState '%s' does not exist\n", target)
}
*/
