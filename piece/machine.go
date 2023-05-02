package piece

import "log"

/*
type MachineBuilder[T any] interface {
	WithContext(context T) MachineBuilder[T]
	WithState(stateName string) MachineBuilder[T]
	Build() *GMachine[T]
}

type MachineBuilderImpl[T any] struct {
	machine *GMachine[T]
}

func (b *MachineBuilderImpl[T]) WithContext(context T) MachineBuilder[T] {
	b.machine.Context = &context
	return b
}

func (b *MachineBuilderImpl[T]) WithState(stateName string) MachineBuilder[T] {
	if s, ok := b.machine.States[stateName]; ok {
		b.machine.CurrentState = s
		return b
	}
	panic(fmt.Sprintf("GState '%s' not found", stateName))
}

func (b *MachineBuilderImpl[T]) Build() *GMachine[T] {
	return b.machine
}
*/

//------------------------------------------------------------------------------

type GMachine[T any] struct {
	Id           string
	Context      *T
	EntryState   *GState[T]
	CurrentState *GState[T]
	States       map[string]*GState[T]
}

type ActionTool[T any] interface {
	Assign(context T)
	Send(event Event)
	Raise(event Event)
}

func (m *GMachine[T]) Assign(context T) {
	// m.Context = &context
	log.Println("Assign")
}

func (m *GMachine[T]) Send(event Event) {
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

func (m *GMachine[T]) Raise(event Event) {
	log.Println("Raise")
}

func (m *GMachine[T]) SendEvent(event Event) TransitionResponse[T] {
	return nil
}

func (m *GMachine[T]) PlaceOn(state ProState, context T) ProMachine[T] {
	return m
}

type ProState interface {
	GetState() string
}

type TransitionResponse[T any] interface {
	GetContext() T
	GetState() ProState
	GetEvent() Event
}

type ProMachine[T any] interface {
	SendEvent(event Event) TransitionResponse[T]
	PlaceOn(state ProState, context T) ProMachine[T]
}

type transitionResp[T any] struct {
	respCh  chan Event
	context *T
}

func (t *transitionResp[T]) GetContext() T {
	select {
	case resp, ok := <-t.respCh:
		if ok {
			close(t.respCh)
			t.processResponse(resp)
		}
	}
	return *t.context
}

func (t *transitionResp[T]) processResponse(evt Event) {

}

/*
type GSupplier[T any] interface {
	getAction(n string) (TAction[T], ActionTool[T])
	getGuard(n string) TPredicate[T]
	getService(n string) TInvocation[T]
}

func (m *GMachine[T]) getAction(n string) (TAction[T], ActionTool[T]) {
	return nil, nil
}

func (m *GMachine[T]) getGuard(n string) TPredicate[T] {
	return nil
}

func (m *GMachine[T]) getService(n string) TInvocation[T] {
	return nil
}
*/

//type Assigner[T any] func(context T)
//type Send func(eventName string)

//func (m *GMachine[T]) WithContext(c T) statepro.MachineBuilder[T] {
//	m.Context = &c
//	return m
//}
//
//func (m *GMachine[T]) WithState(stateName string) statepro.MachineBuilder[T] {
//	if s, ok := m.States[stateName]; ok {
//		m.CurrentState = s
//		return m
//	}
//	panic(fmt.Sprintf("GState '%s' not found", stateName))
//}
//
//func (m *GMachine[T]) Start() statepro.StatePro[T] {
//	return nil
//}

//type ActionTool[T any] struct {
//	Assign Assigner[T]
//	Send   Send
//}

/*
func (m *GMachine[T]) StartOn(target string, c T) {
	if s, ok := m.States[target]; ok {
		m.Context = &c
		m.CurrentState = s
		m.CurrentState.onEntry(*m.Context, Event{Data: nil, Err: nil, EvtType: EventTypeTransitional})
		return
	}
	fmt.Printf("Error: GState '%s' does not exist\n", target)
}
*/
