package statepro

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSON custom unmarshalers

// XMachineRaw is the raw representation of XMachine
type XMachineRaw struct {
	Id          *string            `json:"id"`
	Initial     *string            `json:"initial"`
	States      map[string]*XState `json:"states"`
	SuccessFlow []*string          `json:"successFlow"` // Not part of the XState, but used to define the success flow (Optional)
	Version     string             `json:"version"`     // Not part of the XState, but used to define the machine version (Required)
}

func (x *XMachine) UnmarshalJSON(data []byte) error {
	var raw XMachineRaw
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	x.Id = raw.Id
	x.Initial = raw.Initial
	x.States = raw.States
	x.SuccessFlow = raw.SuccessFlow
	x.Version = raw.Version

	return validateXMachine(x)
}

// XStateRaw is the raw representation of XState
type XStateRaw struct {
	Always      json.RawMessage `json:"always,omitempty"` // []any 			-> any = XTransition | []XTransition
	On          json.RawMessage `json:"on,omitempty"`     // map[string]any -> any = XTransition | []XTransition
	After       json.RawMessage `json:"after,omitempty"`  // map[int]any 	-> any = XTransition | []XTransition
	Invoke      json.RawMessage `json:"invoke,omitempty"` // []any 			-> any = XInvoke | []XInvoke
	Entry       json.RawMessage `json:"entry,omitempty"`  // []any 			-> any = string | []string -> string = ActionName
	Exit        json.RawMessage `json:"exit,omitempty"`   // []any 			-> any = string | []string -> string = ActionName
	Type        *string         `json:"type,omitempty"`
	Description *string         `json:"description,omitempty"`
}

func (x *XState) UnmarshalJSON(data []byte) error {
	var raw XStateRaw
	_ = json.Unmarshal(data, &raw)

	// Always
	x.Always = unmarshalToArr[XTransition](raw.Always)

	// On
	x.On = unmarshalToStrMap[XTransition](raw.On)

	// After
	x.After = unmarshalToIntMap[XTransition](raw.After)

	// Invoke
	x.Invoke = unmarshalToArr[XInvoke](raw.Invoke)

	// Entry
	x.Entry = unmarshalToArr[string](raw.Entry)

	// Exit
	x.Exit = unmarshalToArr[string](raw.Exit)

	x.Type = raw.Type
	x.Description = raw.Description

	return nil
}

// XTransitionRaw is the raw representation of XTransition
type XTransitionRaw struct {
	Condition   *string         `json:"cond"`
	Target      *string         `json:"target"`
	Actions     json.RawMessage `json:"actions"` // []any -> any = string | []string -> string = ActionName
	Description *string         `json:"description"`
}

func (x *XTransition) UnmarshalJSON(data []byte) error {
	var raw XTransitionRaw
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	x.Condition = raw.Condition
	x.Target = raw.Target
	x.Actions = unmarshalToArr[string](raw.Actions)
	x.Description = raw.Description
	return nil
}

// validateXMachine validates the XMachine
func validateXMachine(x *XMachine) error {
	// id
	if x.Id == nil || len(*x.Id) == 0 {
		return fmt.Errorf("machine id is required")
	}

	// initial
	if x.Initial == nil || len(*x.Initial) == 0 {
		return fmt.Errorf("initial state is required")
	}

	// states
	if len(x.States) == 0 {
		return fmt.Errorf("machine definition must have at least one state")
	}

	// version
	version := strings.TrimSpace(x.Version)
	if len(version) == 0 {
		return fmt.Errorf("machine version is required")
	}
	x.Version = version

	// check if success flow states exist
	if x.SuccessFlow != nil && len(x.SuccessFlow) > 0 {
		var errs []string
		for _, stateName := range x.SuccessFlow {
			if _, ok := x.States[*stateName]; !ok {
				errs = append(errs, fmt.Sprintf("success flow state '%s' does not exist in states", stateName))
			}
		}
		if len(errs) > 0 {
			return fmt.Errorf(strings.Join(errs, "\n"))
		}
	}

	// check if initial state exists
	if _, ok := x.States[*x.Initial]; !ok {
		return fmt.Errorf("initial state '%s' does not exist in states", *x.Initial)
	}

	return nil
}

func unmarshalToStrMap[V any](bArr json.RawMessage) map[string][]*V {
	if bArr == nil {
		return map[string][]*V{}
	}

	// []*V
	var mapArr map[string][]*V
	if err := json.Unmarshal(bArr, &mapArr); err == nil {
		return mapArr
	}

	// *V
	var mapSimp map[string]*V
	if err := json.Unmarshal(bArr, &mapSimp); err == nil {
		var resp = map[string][]*V{}
		for k, v := range mapSimp {
			resp[k] = []*V{v}
		}
		return resp
	}

	return map[string][]*V{}
}

func unmarshalToIntMap[V any](bArr json.RawMessage) map[int][]*V {
	if bArr == nil {
		return map[int][]*V{}
	}

	// []*V
	var mapArr map[int][]*V
	if err := json.Unmarshal(bArr, &mapArr); err == nil {
		return mapArr
	}

	// *V
	var mapSimp map[int]*V
	if err := json.Unmarshal(bArr, &mapSimp); err == nil {
		var resp = map[int][]*V{}
		for k, v := range mapSimp {
			resp[k] = []*V{v}
		}
		return resp
	}

	return map[int][]*V{}
}

func unmarshalToArr[V any](bArr json.RawMessage) []*V {
	if bArr == nil {
		return []*V{}
	}

	// []*V
	var arr []*V
	if err := json.Unmarshal(bArr, &arr); err == nil {
		return arr
	}

	// *V
	var simp *V
	if err := json.Unmarshal(bArr, &simp); err == nil {
		return []*V{simp}
	}

	return []*V{}
}
