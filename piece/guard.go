package piece

import "fmt"

type TPredicate[T any] func(T, Event) (bool, error)

type GGuard[T any] struct {
	Condition *string
	Target    *string // Mandatory
	Actions   []*GAction[T]
	Predicate TPredicate[T]
}

func (g *GGuard[T]) check(c T, e Event) (string, bool, error) {
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

func (g *GGuard[T]) doActions(c T, e Event) error {
	for _, a := range g.Actions {
		if err := a.do(c, e); err != nil {
			return err
		}
	}
	return nil
}

func CastPredicate[T any](i any) (TPredicate[T], error) {
	if f, ok := i.(func(T, Event) (bool, error)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("predicate '%s' with wrong type", i)
}

/*
func (g *GGuard[T]) check(c T, e Event, supplier GSupplier[T]) (string, bool, error) {
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

func (g *GGuard[T]) doActions(c T, e Event, supplier GSupplier[T]) error {
	for _, a := range g.Actions {
		if err := a.do(c, e, supplier); err != nil {
			return err
		}
	}
	return nil
}
*/
