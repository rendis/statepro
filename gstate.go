package statepro

import (
	"context"
)

type StateType string

const (
	StateTypeInitial StateType = "initial"
	StateTypeNormal  StateType = "normal"
	StateTypeFinal   StateType = "final"
	StateTypeHistory StateType = "history"
	StateTypeShared  StateType = "shared"
)

type gState[ContextType any] struct {
	Name      *string // Mandatory
	Always    *gTransition[ContextType]
	Entry     []*gAction[ContextType]
	Exit      []*gAction[ContextType]
	On        map[string]*gTransition[ContextType]
	Services  []*gService[ContextType]
	StateType StateType
}

func (s *gState[ContextType]) onEntry(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) (*string, bool, error) {
	target, err := s.always(ctx, c, e, at)
	if err != nil {
		return nil, false, err
	}

	// if always guard returns a target state, then send
	if target != nil {
		return target, false, nil
	}

	err = s.execEntry(ctx, c, e, at)
	if err != nil {
		return nil, false, err
	}

	if !s.isFinalState() && len(s.Services) > 0 {
		s.invokeServices(ctx, c, e)
		return nil, true, nil
	}

	return nil, false, nil
}

func (s *gState[ContextType]) always(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) (*string, error) {
	if s.Always != nil {
		return s.Always.resolve(ctx, c, e, at)
	}
	return nil, nil
}

func (s *gState[ContextType]) execEntry(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) error {
	for _, a := range s.Entry {
		err := a.do(ctx, c, e, at)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *gState[ContextType]) invokeServices(ctx context.Context, c *ContextType, e Event) {
	for _, srv := range s.Services {
		go srv.invoke(ctx, c, e)
	}
}

func (s *gState[ContextType]) onEvent(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) (*string, error) {
	// check if the event is defined in the state
	if s.On == nil || s.On[e.GetEventName()] == nil {
		return nil, &EventNotDefinedError{EventName: e.GetEventName(), StateName: *s.Name}
	}

	// on event actions
	return s.On[e.GetEventName()].resolve(ctx, c, e, at)
}

func (s *gState[ContextType]) execExit(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) error {
	if s.StateType == StateTypeFinal {
		return nil
	}
	for _, a := range s.Exit {
		err := a.do(ctx, c, e, at)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *gState[ContextType]) isFinalState() bool {
	return s.StateType == StateTypeFinal
}

func (s *gState[ContextType]) getNextEvents() []string {
	var events []string
	for k := range s.On {
		events = append(events, k)
	}
	return events
}

func (s *gState[ContextType]) getNextTargets() []string {
	var notRepeated = make(map[string]bool)
	for _, v := range s.On {
		for _, g := range v.Guards {
			if g.Target == nil || notRepeated[*g.Target] {
				continue
			}
			notRepeated[*g.Target] = true
		}
	}

	var targets []string
	for k := range notRepeated {
		targets = append(targets, k)
	}
	return targets
}

func (s *gState[ContextType]) containsTarget(target string) bool {
	for _, t := range s.getNextTargets() {
		if t == target {
			return true
		}
	}
	return false
}
