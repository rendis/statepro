package piece

import "fmt"

type TAction[ContextType any] func(ContextType, Event, ActionTool[ContextType]) error

type GAction[ContextType any] struct {
	Name string
	Act  TAction[ContextType]
	Tool ActionTool[ContextType]
}

func (a *GAction[ContextType]) do(c ContextType, e Event) error {
	if a.Act == nil {
		return fmt.Errorf("action '%s' not found", a.Name)
	}
	return a.Act(c, e, a.Tool)
}

func CastToAct[ContextType any](i any) (TAction[ContextType], error) {
	if f, ok := i.(func(ContextType, Event, ActionTool[ContextType]) error); ok {
		return f, nil
	}
	return nil, fmt.Errorf("action '%s' with wrong type", i)
}
