package piece

import "fmt"

type TPredicate[ContextType any] func(ContextType, Event) (bool, error)

type GGuard[ContextType any] struct {
	Condition *string
	Target    *string // Mandatory
	Actions   []*GAction[ContextType]
	Predicate TPredicate[ContextType]
}

func (g *GGuard[ContextType]) check(c ContextType, e Event) (string, bool, error) {
	// (Else || Directly) GGuard
	if g.Condition == nil {
		err := g.doActions(c, e)
		if err != nil {
			return "", false, err
		}
		return *g.Target, true, nil
	}

	// (If || ElseIf) GGuard
	if g.Predicate == nil {
		return "", false, fmt.Errorf("guard '%s' not found", *g.Condition)
	}

	ok, err := (g.Predicate)(c, e)
	if err != nil {
		return "", false, err
	}

	if ok {
		err = g.doActions(c, e)
		if err != nil {
			return "", false, err
		}

		return *g.Target, true, nil
	}
	return "", false, nil
}

func (g *GGuard[ContextType]) doActions(c ContextType, e Event) error {
	for _, a := range g.Actions {
		if err := a.do(c, e); err != nil {
			return err
		}
	}
	return nil
}

func CastPredicate[ContextType any](i any) (TPredicate[ContextType], error) {
	if f, ok := i.(func(ContextType, Event) (bool, error)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("predicate '%s' with wrong type", i)
}

/*
func (g *GGuard[ContextType]) check(c ContextType, e Event, supplier GSupplier[ContextType]) (string, bool, error) {
	// (Else || Directly) GGuard
	if g.Condition == nil {
		err := g.doActions(c, e, supplier)
		if err != nil {
			return "", false, err
		}
		return *g.Target, true, nil
	}

	// (If || ElseIf) GGuard
	pred := supplier.getGuard(*g.Condition)
	if pred == nil {
		return "", false, fmt.Errorf("guard '%s' not found", *g.Condition)
	}

	ok, err := (pred)(c, e)
	if err != nil {
		return "", false, err
	}

	if ok {
		err = g.doActions(c, e, supplier)
		if err != nil {
			return "", false, err
		}

		return *g.Target, true, nil
	}
	return "", false, nil
}

func (g *GGuard[ContextType]) doActions(c ContextType, e Event, supplier GSupplier[ContextType]) error {
	for _, a := range g.Actions {
		if err := a.do(c, e, supplier); err != nil {
			return err
		}
	}
	return nil
}
*/
