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

// onEntry is called when the state is entered.
// return
func (s *GState[ContextType]) onEntry(c ContextType, e Event) (*string, bool, error) {
	target, err := s.always(c, e)
	if err != nil {
		return nil, false, err
	}

	// If always guard returns a target state, then send
	if target != nil {
		return target, false, nil
	}

	err = s.execEntry(c, e)
	if err != nil {
		return nil, false, err
	}

	if !s.isFinalState() && len(s.Services) > 0 {
		go s.invokeServices(c, e)
		return nil, true, nil
	}

	return nil, false, nil
}

// onEvent is called
func (s *GState[ContextType]) onEvent(c ContextType, e Event) (*string, error) {
	if s.On != nil && s.On[e.GetName()] != nil {
		return s.On[e.GetName()].resolve(c, e)
	}
	return nil, nil
}

func (s *GState[ContextType]) always(c ContextType, e Event) (*string, error) {
	if s.Always != nil {
		return s.Always.resolve(c, e)
	}
	return nil, nil
}

func (s *GState[ContextType]) execEntry(c ContextType, e Event) error {
	for _, a := range s.Entry {
		err := a.do(c, e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GState[ContextType]) execExit(c ContextType, e Event) error {
	if s.StateType == StateTypeFinal {
		return nil
	}
	//s.srvChMtx.Lock()
	//close(s.srvCh)
	//s.srvChMtx.Unlock()
	for _, a := range s.Exit {
		err := a.do(c, e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GState[ContextType]) invokeServices(c ContextType, e Event) {
	s.srvCh = make(chan *InvocationResponse, 1)
	for _, srv := range s.Services {
		go srv.invoke(c, e, s.srvCh)
	}
	select {
	case resp, open := <-s.srvCh:
		if open {
			close(s.srvCh)
			log.Printf("GService %s returned %s", resp.Target, resp.Event.GetName())
		}
	}
}

func (s *GState[ContextType]) isFinalState() bool {
	return s.StateType == StateTypeFinal
}