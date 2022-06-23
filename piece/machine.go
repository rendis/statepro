package piece

import "fmt"

type GMachine[CTX any] struct {
	Id           string
	Context      *CTX
	EntryState   *GState[CTX]
	CurrentState *GState[CTX]
	States       map[string]*GState[CTX]
}

func (m *GMachine[CTX]) Start() {
	m.CurrentState = m.EntryState
	m.CurrentState.onEntry(*m.Context, GEvent{Name: "onEntry", Data: nil, Err: nil, EvtType: EventTypeOnEntry})
}

func (m *GMachine[CTX]) StartOn(target string, c CTX) {
	if s, ok := m.States[target]; ok {
		m.Context = &c
		m.CurrentState = s
		m.CurrentState.onEntry(*m.Context, GEvent{Name: "transitional", Data: nil, Err: nil, EvtType: EventTypeTransitionalEvent})
		return
	}
	fmt.Printf("Error: GState '%s' does not exist\n", target)
}
