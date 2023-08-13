package statepro

import (
	"context"
	"errors"
)

// ProMachine interface to interact with the machine
type ProMachine[ContextType any] interface {
	PlaceOn(stateName string) error
	Start(ctx context.Context) TransitionResponse
	StartOn(ctx context.Context, stateName string) TransitionResponse
	StartOnWithEvent(ctx context.Context, stateName string, event Event) TransitionResponse
	SendEvent(ctx context.Context, event Event) TransitionResponse
	GetNextEvents() []string
	GetNextTargets() []string
	ContainsTarget(target string) bool
	GetState() string
	IsFinalState() bool
	GetContext() *ContextType
	SetContext(machineCtx *ContextType)
	CallContextToSource(ctx context.Context) error
	GetSuccessFlow() []string
	GetInEventsForCurrentState() []string
	InEventOnCurrentState(event string) bool
}

// ProMachineToSourceHandler interface to handle context to source
type ProMachineToSourceHandler[ContextType any] interface {
	ContextToSource(ctx context.Context, stateName string, machineContext *ContextType) error
}

// MachineLinks interface to link machine methods to the machine definition
type MachineLinks[ContextType any] interface {
	ApplyId(id, version string) bool
}

// GetProMachineOptions options to get a ProMachine instance
type GetProMachineOptions[ContextType any] struct {
	Definition []byte
	Context    *ContextType
}

func (o *GetProMachineOptions[ContextType]) validate() error {
	// definition must be provided
	if len(o.Definition) == 0 {
		return errors.New("definition must be provided")
	}
	return nil
}

// ExecutionType type of action execution, always, on entry or on exit
type ExecutionType string

const (
	ExecutionTypeNone    ExecutionType = "none"
	ExecutionTypeAlways  ExecutionType = "always"
	ExecutionTypeOnEntry ExecutionType = "onEntry"
	ExecutionTypeOnExit  ExecutionType = "onExit"
)

// ActionOption option to register an action
type ActionOption struct {
	ActionName    string
	Required      bool
	ExecutionType ExecutionType
}

// StateProOptions options to initialize statepro
type StateProOptions struct {
	ToSourceExecutionMode ExecutionType
	Actions               []ActionOption
}
