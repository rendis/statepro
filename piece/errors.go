package piece

import "fmt"

// StateNotFountError is an error type for state not found in machine
type StateNotFountError struct {
	EventName string
}

func (e *StateNotFountError) Error() string {
	return fmt.Sprintf("Event '%s' not found", e.EventName)
}

// EventNotDefinedError is an error type for event not defined on state
type EventNotDefinedError struct {
	EventName string
	StateName string
}

func (e *EventNotDefinedError) Error() string {
	return fmt.Sprintf("Event '%s' not defined in state '%s'", e.EventName, e.StateName)
}
