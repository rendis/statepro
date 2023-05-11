package piece

import (
	"errors"
	"fmt"
)

// StateNotFountError used when state not found
type StateNotFountError struct {
	EventName string
}

func (e *StateNotFountError) Error() string {
	return fmt.Sprintf("Event '%s' not found", e.EventName)
}

// EventNotDefinedError used when event not defined in state
type EventNotDefinedError struct {
	EventName string
	StateName string
}

func (e *EventNotDefinedError) Error() string {
	return fmt.Sprintf("Event '%s' not defined in state '%s'", e.EventName, e.StateName)
}

// ContextToSourceNotImplementedError used when contextToSourceFn is called but not implemented
var ContextToSourceNotImplementedError = errors.New("ContextToSource function not implemented")
