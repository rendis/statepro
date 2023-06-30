package statepro

import (
	"context"
	"fmt"
)

// Action

type tAction[ContextType any] func(context.Context, *ContextType, Event, ActionTool[ContextType]) error

type gAction[ContextType any] struct {
	Name string
	Act  tAction[ContextType]
}

func (a *gAction[ContextType]) do(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) error {
	if a.Act == nil {
		return fmt.Errorf("action '%s' not found", a.Name)
	}
	return a.Act(ctx, c, e, at)
}

func castToAct[ContextType any](i any) (tAction[ContextType], error) {
	if f, ok := i.(func(context.Context, *ContextType, Event, ActionTool[ContextType]) error); ok {
		return f, nil
	}
	return nil, fmt.Errorf("action '%s' with wrong type. Expected: func(context.Context, ContextType, Event, ActionTool)", i)
}

// Guard

type tPredicate[ContextType any] func(context.Context, *ContextType, Event) (bool, error)

type gGuard[ContextType any] struct {
	Condition *string
	Target    *string // Mandatory
	Actions   []*gAction[ContextType]
	Predicate tPredicate[ContextType]
}

func (g *gGuard[ContextType]) check(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) (string, bool, error) {
	// if Condition is nil, then it is an Else or Directly gGuard
	if g.Condition == nil {
		err := g.doActions(ctx, c, e, at)
		if err != nil {
			return "", false, err
		}
		return *g.Target, true, nil
	}

	// if Predicate is nil, then it is an If or ElseIf gGuard
	if g.Predicate == nil {
		return "", false, fmt.Errorf("guard '%s' not found", *g.Condition)
	}

	ok, err := (g.Predicate)(ctx, c, e)
	if err != nil {
		return "", false, err
	}

	if ok {
		err = g.doActions(ctx, c, e, at)
		if err != nil {
			return "", false, err
		}

		return *g.Target, true, nil
	}
	return "", false, nil
}

func (g *gGuard[ContextType]) doActions(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) error {
	for _, a := range g.Actions {
		if err := a.do(ctx, c, e, at); err != nil {
			return err
		}
	}
	return nil
}

func castPredicate[ContextType any](i any) (tPredicate[ContextType], error) {
	if f, ok := i.(func(context.Context, *ContextType, Event) (bool, error)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("predicate '%s' with wrong type", i)
}

// Service

type tInvocation[ContextType any] func(context.Context, ContextType, Event)

type invocationResponse struct {
	Target *string
	Event  Event
	Err    error
}

type gService[ContextType any] struct {
	Id      *string // Mandatory
	Src     *string // Mandatory
	Inv     tInvocation[ContextType]
	OnDone  *gTransition[ContextType]
	OnError *gTransition[ContextType]
}

func (s *gService[ContextType]) invoke(ctx context.Context, c *ContextType, e Event) {
	s.Inv(ctx, *c, e)
}

func castToSrv[ContextType any](i any) (tInvocation[ContextType], error) {
	if f, ok := i.(func(context.Context, ContextType, Event)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("service '%s' with wrong type. Expected: func(context.Context, ContextType, Event)", i)
}

// Transition

type gTransition[ContextType any] struct {
	Guards []*gGuard[ContextType] // At least one guard is required
}

func (t *gTransition[ContextType]) resolve(ctx context.Context, c *ContextType, e Event, at ActionTool[ContextType]) (*string, error) {
	for _, g := range t.Guards {
		target, ok, err := g.check(ctx, c, e, at)
		if err != nil {
			return nil, err
		}
		if ok {
			return &target, nil
		}
	}
	return nil, nil
}
