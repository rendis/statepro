package piece

import (
	"log"
	"sync"
)

type StateType string

const (
	StateTypeInitial StateType = "initial"
	StateTypeNormal  StateType = "normal"
	StateTypeFinal   StateType = "final"
	StateTypeHistory StateType = "history"
	StateTypeShared  StateType = "shared"
)

type GState[ContextType any] struct {
	Name      *string // Mandatory
	Always    *GTransition[ContextType]
	Entry     []*GAction[ContextType]
	Exit      []*GAction[ContextType]
	On        map[string]*GTransition[ContextType]
	Services  []*GService[ContextType]
	StateType StateType

	srvChMtx sync.Mutex
	srvCh    chan *InvocationResponse
}

func (s *GState[ContextType]) onEntry(c ContextType, e Event, at ActionTool[ContextType]) (*string, bool, error) {
	target, err := s.always(c, e, at)
	if err != nil {
		return nil, false, err
	}

	// if always guard returns a target state, then send
	if target != nil {
		return target, false, nil
	}

	err = s.execEntry(c, e, at)
	if err != nil {
		return nil, false, err
	}

	if !s.isFinalState() && len(s.Services) > 0 {
		go s.invokeServices(c, e, at)
		return nil, true, nil
	}

	return nil, false, nil
}

func (s *GState[ContextType]) always(c ContextType, e Event, at ActionTool[ContextType]) (*string, error) {
	if s.Always != nil {
		return s.Always.resolve(c, e, at)
	}
	return nil, nil
}

func (s *GState[ContextType]) execEntry(c ContextType, e Event, at ActionTool[ContextType]) error {
	for _, a := range s.Entry {
		err := a.do(c, e, at)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GState[ContextType]) invokeServices(c ContextType, e Event, at ActionTool[ContextType]) {
	s.srvCh = make(chan *InvocationResponse, 1)
	for _, srv := range s.Services {
		go srv.invoke(c, e, at, s.srvCh)
	}
	select {
	case resp, open := <-s.srvCh:
		if open {
			close(s.srvCh)
			log.Printf("GService %s returned %s", resp.Target, resp.Event.GetName())
		}
	}
}

func (s *GState[ContextType]) onEvent(c ContextType, e Event, at ActionTool[ContextType]) (*string, error) {
	// check if the event is defined in the state
	if s.On == nil || s.On[e.GetName()] == nil {
		return nil, &EventNotDefinedError{EventName: e.GetName(), StateName: *s.Name}
	}

	// on event actions
	return s.On[e.GetName()].resolve(c, e, at)
}

func (s *GState[ContextType]) execExit(c ContextType, e Event, at ActionTool[ContextType]) error {
	if s.StateType == StateTypeFinal {
		return nil
	}
	for _, a := range s.Exit {
		err := a.do(c, e, at)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GState[ContextType]) isFinalState() bool {
	return s.StateType == StateTypeFinal
}

func (s *GState[ContextType]) getNextEvents() []string {
	var events []string
	for k := range s.On {
		events = append(events, k)
	}
	return events
}
