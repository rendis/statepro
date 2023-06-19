package statepro

import (
	"encoding/json"
	"fmt"
)

// XMachine is the json representation of a machine.
// Contains Id, Initial state and the States, these fields are part of the XState.
// The SuccessFlow is not part of the XState, but is used to define the success flow (Optional).
type XMachine struct {
	Id          *string            `json:"id"`
	Initial     *string            `json:"initial"`
	States      *map[string]XState `json:"states"`
	SuccessFlow []string           `json:"successFlow"` // Not part of the XState, but used to define the success flow (Optional)
}

// XState is the json representation of a state.
type XState struct {
	Always      any             `json:"always"` // any = XTransition | []XTransition
	On          *map[string]any `json:"on"`     // any = XTransition | []XTransition
	After       *map[int]any    `json:"after"`  // any = XTransition | []XTransition
	Type        *string         `json:"type"`
	Invoke      any             `json:"invoke"` // any = XInvoke | []XInvoke
	Entry       any             `json:"entry"`  // any = XActionName | []XActionName
	Exit        any             `json:"exit"`   // any = XActionName | []XActionName
	Description *string         `json:"description"`
}

// XTransition is the json representation of a transition.
type XTransition struct {
	Condition   *string `json:"cond"`
	Target      *string `json:"target"`
	Actions     any     `json:"actions"` // any = XActionName | []XActionName  (X)
	Description *string `json:"description"`
}

// XInvoke is the json representation of the invoke property.
type XInvoke struct {
	Id      *string        `json:"id"`
	Src     *string        `json:"src"`
	OnDone  *[]XTransition `json:"onDone"`
	OnError *[]XTransition `json:"onError"`
}

func (x *XState) getOn() (map[string][]XTransition, error) {
	if x.On == nil {
		return nil, nil
	}
	on := make(map[string][]XTransition, len(*x.On))
	if x.On != nil {
		for k, v := range *x.On {

			p, err := parseFromAny[XTransition](v)
			if err != nil {
				return nil, err
			}
			on[k] = p
		}
	}
	return on, nil
}

func (x *XState) getEntryActions() ([]string, error) {
	if x.Entry == nil {
		return []string{}, nil
	}
	return parseFromAny[string](x.Entry)
}

func (x *XState) getExitActions() ([]string, error) {
	if x.Exit == nil {
		return []string{}, nil
	}
	return parseFromAny[string](x.Exit)
}

func (x *XState) getInvoke() ([]XInvoke, error) {
	if x.Invoke == nil {
		return []XInvoke{}, nil
	}
	return parseFromAny[XInvoke](x.Invoke)
}

func (x *XState) getAfter() (map[int][]XTransition, error) {
	if x.After == nil {
		return nil, nil
	}
	after := make(map[int][]XTransition, len(*x.After))
	if x.After != nil {
		for k, v := range *x.After {
			p, err := parseFromAny[XTransition](v)
			if err != nil {
				return nil, err
			}
			after[k] = p
		}
	}
	return after, nil
}

func (x *XState) getAlways() ([]XTransition, error) {
	if x.Always == nil {
		return []XTransition{}, nil
	}
	return parseFromAny[XTransition](x.Always)
}

func (x *XTransition) getActions() ([]string, error) {
	if x.Actions == nil {
		return []string{}, nil
	}
	return parseFromAny[string](x.Actions)
}

func parseFromAny[T any](a any) ([]T, error) {
	if a == nil {
		return []T{}, nil
	}

	if c, ok := a.([]any); ok {
		var r []T
		for _, v := range c {
			if m, ok := v.(map[string]any); ok {
				t, err := parseFromMap[T](m)
				if err != nil {
					return nil, err
				}
				r = append(r, *t)
			} else if m, ok := v.(T); ok {
				r = append(r, m)
			} else {
				return nil, fmt.Errorf("invalid type %T", v)
			}
		}
		return r, nil
	}

	if c, ok := a.(map[string]any); ok {
		t, err := parseFromMap[T](c)
		if err != nil {
			return nil, err
		}
		return []T{*t}, nil
	}

	if c, ok := a.(T); ok {
		return []T{c}, nil
	}

	return nil, fmt.Errorf("invalid type: %T", a)
}

func parseFromMap[ContextType any](m map[string]any) (*ContextType, error) {
	jsonStr, _ := json.Marshal(m)
	var t = new(ContextType)
	if err := json.Unmarshal(jsonStr, t); err != nil {
		return nil, err
	}
	return t, nil
}
