package xparse

import (
	"encoding/json"
	"fmt"
)

type XEventName string
type XActionName string
type XAfterValue int

type XMachine struct {
	Id      *string            `json:"id"`
	Initial *string            `json:"initial"`
	States  *map[string]XState `json:"states"`
}

type XState struct {
	Always any                  `json:"always"` // any = XTransition | []XTransition
	On     *map[XEventName]any  `json:"on"`     // any = XTransition | []XTransition
	After  *map[XAfterValue]any `json:"after"`  // any = XTransition | []XTransition
	Type   *string              `json:"type"`
	Invoke any                  `json:"invoke"` // any = XInvoke | []XInvoke
	Entry  any                  `json:"entry"`  // any = XActionName | []XActionName
	Exit   any                  `json:"exit"`   // any = XActionName | []XActionName
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

func (x *XState) GetOn() (map[XEventName][]XTransition, error) {
	if x.On == nil {
		return nil, nil
	}
	on := make(map[XEventName][]XTransition, len(*x.On))
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

func (x *XState) GetEntryActions() ([]XActionName, error) {
	if x.Entry == nil {
		return []XActionName{}, nil
	}
	return parseFromAny[XActionName](x.Entry)
}

func (x *XState) GetExitActions() ([]XActionName, error) {
	if x.Exit == nil {
		return []XActionName{}, nil
	}
	return parseFromAny[XActionName](x.Exit)
}

func (x *XState) GetInvoke() ([]XInvoke, error) {
	if x.Invoke == nil {
		return []XInvoke{}, nil
	}
	return parseFromAny[XInvoke](x.Invoke)
}

func (x *XState) GetAfter() (map[XAfterValue][]XTransition, error) {
	if x.After == nil {
		return nil, nil
	}
	after := make(map[XAfterValue][]XTransition, len(*x.After))
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

func (x *XTransition) GetActions() ([]XActionName, error) {
	if x.Actions == nil {
		return []XActionName{}, nil
	}
	return parseFromAny[XActionName](x.Actions)
}

func parseFromAny[T any](a any) ([]T, error) {
	if a == nil {
		return []T{}, nil
	}
	if c, ok := a.([]T); ok {
		return c, nil
	}
	if c, ok := a.(T); ok {
		return []T{c}, nil
	}

	if c, ok := a.(map[string]any); ok {
		jsonStr, _ := json.Marshal(c)
		var t = new(T)
		if err := json.Unmarshal(jsonStr, t); err != nil {
			return nil, err
		}
		return []T{*t}, nil
	}

	return nil, fmt.Errorf("invalid type: %T", a)
}
