package piece

import (
	"fmt"
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

type GState[CTX any] struct {
	Name      string
	Always    *GTransition[CTX]
	Entry     []*GAction[CTX]
	Exit      []*GAction[CTX]
	On        map[string]*GTransition[CTX]
	Services  []*Service[CTX]
	StateType StateType

	srvChMtx sync.Mutex
	srvCh    chan *GInvocationResponse
}

func (s *GState[CTX]) onEntry(c CTX, e GEvent) {
	for _, a := range s.Entry {
		a.execute(c, e)
	}
}

func (s *GState[CTX]) onExit(c CTX, e GEvent) {
	if s.StateType == StateTypeFinal {
		return
	}
	s.srvChMtx.Lock()
	close(s.srvCh)
	s.srvChMtx.Unlock()
	for _, a := range s.Exit {
		a.execute(c, e)
	}
}

func (s *GState[CTX]) onEvent(c CTX, e GEvent) (string, error) {
	if err := s.checkFinalState(); err != nil {
		return "", err
	}

	if s.On != nil && s.On[e.Name] != nil {
		return s.On[e.Name].resolve(c, e), nil
	}
	return "", nil
}

func (s *GState[CTX]) invokeServices(c CTX, e GEvent) {
	if err := s.checkFinalState(); err != nil {
		//return "", Err
		return
	}

	s.srvCh = make(chan *GInvocationResponse, 1)
	for _, srv := range s.Services {
		go srv.invoke(c, e, s.srvCh)
	}
	select {
	case resp, open := <-s.srvCh:
		if open {
			close(s.srvCh)
			fmt.Printf("Service invoked: %+v\n", resp)
			return
		}
		fmt.Printf("Service Invocation response channel was closed: %+v\n", resp)
	}
}

func (s *GState[CTX]) checkFinalState() error {
	if s.StateType == StateTypeFinal {
		return fmt.Errorf("state '%s' is final", s.Name)
	}
	return nil
}
