package piece

import "fmt"

type TAction[T any] func(T, Event, ActionTool[T]) error

type GAction[T any] struct {
	Name string
	Act  TAction[T]
	Tool ActionTool[T]
}

func (a *GAction[T]) do(c T, e Event) error {
	if a.Act == nil {
		return fmt.Errorf("action '%s' not found", a.Name)
	}
	return a.Act(c, e, a.Tool)
}

func CastToAct[T any](i any) (TAction[T], error) {
	if f, ok := i.(func(T, Event, ActionTool[T]) error); ok {
		return f, nil
	}
	return nil, fmt.Errorf("action '%s' with wrong type", i)
}
