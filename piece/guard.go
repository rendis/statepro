package piece

import "fmt"

type TPredicate[ContextType any] func(*ContextType, Event) (bool, error)

type GGuard[ContextType any] struct {
	Condition *string
	Target    *string // Mandatory
	Actions   []*GAction[ContextType]
	Predicate TPredicate[ContextType]
}

func (g *GGuard[ContextType]) check(c *ContextType, e Event, at ActionTool) (string, bool, error) {
	// if Condition is nil, then it is an Else or Directly GGuard
	if g.Condition == nil {
		err := g.doActions(c, e, at)
		if err != nil {
			return "", false, err
		}
		return *g.Target, true, nil
	}

	// if Predicate is nil, then it is an If or ElseIf GGuard
	if g.Predicate == nil {
		return "", false, fmt.Errorf("guard '%s' not found", *g.Condition)
	}

	ok, err := (g.Predicate)(c, e)
	if err != nil {
		return "", false, err
	}

	if ok {
		err = g.doActions(c, e, at)
		if err != nil {
			return "", false, err
		}

		return *g.Target, true, nil
	}
	return "", false, nil
}

func (g *GGuard[ContextType]) doActions(c *ContextType, e Event, at ActionTool) error {
	for _, a := range g.Actions {
		if err := a.do(c, e, at); err != nil {
			return err
		}
	}
	return nil
}

func CastPredicate[ContextType any](i any) (TPredicate[ContextType], error) {
	if f, ok := i.(func(*ContextType, Event) (bool, error)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("predicate '%s' with wrong type", i)
}
