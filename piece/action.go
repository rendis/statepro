package piece

import "fmt"

type TAction[ContextType any] func(*ContextType, Event, ActionTool) error

type GAction[ContextType any] struct {
	Name string
	Act  TAction[ContextType]
}

func (a *GAction[ContextType]) do(c *ContextType, e Event, at ActionTool) error {
	if a.Act == nil {
		return fmt.Errorf("action '%s' not found", a.Name)
	}
	return a.Act(c, e, at)
}

func CastToAct[ContextType any](i any) (TAction[ContextType], error) {
	if f, ok := i.(func(*ContextType, Event, ActionTool) error); ok {
		return f, nil
	}
	return nil, fmt.Errorf("action '%s' with wrong type", i)
}
