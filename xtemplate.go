package statepro

import (
	"encoding/json"
	"fmt"
)

type XMachine struct {
	Id      *string            `json:"id"`
	Initial *string            `json:"initial"`
	States  *map[string]XState `json:"states"`
}

type XState struct {
	Always any             `json:"always"` // any = XTransition | []XTransition
	On     *map[string]any `json:"on"`     // any = XTransition | []XTransition
	After  *map[int]any    `json:"after"`  // any = XTransition | []XTransition
	Type   *string         `json:"type"`
	Invoke any             `json:"invoke"` // any = XInvoke | []XInvoke
	Entry  any             `json:"entry"`  // any = XActionName | []XActionName
	Exit   any             `json:"exit"`   // any = XActionName | []XActionName
}

type XTransition struct {
	Condition *string `json:"cond"`
	Target    *string `json:"target"`
	Actions   any     `json:"actions"` // any = XActionName | []XActionName  (X)
}

type XInvoke struct {
	Id      *string        `json:"id"`
	Src     *string        `json:"src"`
	OnDone  *[]XTransition `json:"onDone"`
	OnError *[]XTransition `json:"onError"`
}

func (x *XState) GetOn() (map[string][]XTransition, error) {
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

func (x *XState) GetEntryActions() ([]string, error) {
	if x.Entry == nil {
		return []string{}, nil
	}
	return parseFromAny[string](x.Entry)
}

func (x *XState) GetExitActions() ([]string, error) {
	if x.Exit == nil {
		return []string{}, nil
	}
	return parseFromAny[string](x.Exit)
}

func (x *XState) GetInvoke() ([]XInvoke, error) {
	if x.Invoke == nil {
		return []XInvoke{}, nil
	}
	return parseFromAny[XInvoke](x.Invoke)
}

func (x *XState) GetAfter() (map[int][]XTransition, error) {
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

func (x *XState) GetAlways() ([]XTransition, error) {
	if x.Always == nil {
		return []XTransition{}, nil
	}
	return parseFromAny[XTransition](x.Always)
}

func (x *XTransition) GetActions() ([]string, error) {
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

	return nil, fmt.Errorf("invalid type: %T", a)
}

func parseFromMap[T any](m map[string]any) (*T, error) {
	jsonStr, _ := json.Marshal(m)
	var t = new(T)
	if err := json.Unmarshal(jsonStr, t); err != nil {
		return nil, err
	}
	return t, nil
}
