package statepro

import (
	"errors"
	"fmt"
)

// NoMachineDefinitionAvailableForIdErr used when no machine definition available for id
var NoMachineDefinitionAvailableForIdErr = errors.New("no machine definition available for id")

// ContextToSourceNotImplementedError used when contextToSourceFn is called but not implemented
var ContextToSourceNotImplementedError = errors.New("ContextToSource function not implemented")

// StateProNotInitializedErr used when statepro is not initialized
var StateProNotInitializedErr = errors.New("statepro not initialized. Call InitStatePro() before using any other statepro function")

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
