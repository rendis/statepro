package piece

import "fmt"

type EventNotFountError struct {
	EventName string
}

func (e *EventNotFountError) Error() string {
	return fmt.Sprintf("Event '%s' not found", e.EventName)
}

type EventNotDefinedError struct {
	EventName string
	StateName string
}

func (e *EventNotDefinedError) Error() string {
	return fmt.Sprintf("Event '%s' not defined in state '%s'", e.EventName, e.StateName)
}
