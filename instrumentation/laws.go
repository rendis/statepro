package instrumentation

import (
	"context"
	"github.com/rendis/statepro/v3/theoretical"
)

type ActionType string

const (
	ActionTypeEntry      ActionType = "entry"
	ActionTypeExit       ActionType = "exit"
	ActionTypeTransition ActionType = "transition"
)

// EmittedEvent represents an event emitted internally by an entry action via EmitEvent.
// Emitted events are collected during entry action execution and processed after all entry actions
// complete. They are matched against the current reality's On handlers and, if a transition is
// approved (conditions pass), the machine advances via the existing doCyclicTransition flow.
//
// Events are processed in FIFO order (order of EmitEvent calls). The first event that triggers
// an approved transition wins; remaining events are discarded since the reality has changed.
type EmittedEvent struct {
	// Name is the event name, matched against reality On handler keys.
	Name string

	// Data is optional event data, available to conditions via Event.GetData().
	Data map[string]any
}

type QuantumMachineExecutorArgs struct {
	Context               any
	RealityName           string
	UniverseID            string
	UniverseCanonicalName string
	Event                 Event
	AccumulatorStatistics AccumulatorStatistics

	// EmittedEvents collects events emitted by constants entry actions via EmitEvent.
	// When nil, EmitEvent calls are no-ops (used to restrict emit to entry action context only).
	EmittedEvents *[]EmittedEvent
}
type ConstantsLawsExecutor interface {
	ExecuteEntryInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
	ExecuteExitInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
	ExecuteEntryAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
	ExecuteExitAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
	ExecuteTransitionInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
	ExecuteTransitionAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
	GetSnapshot() *MachineSnapshot
}

type ObserverExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseCanonicalName() string
	GetUniverseId() string
	GetAccumulatorStatistics() AccumulatorStatistics
	GetEvent() Event
	GetObserver() theoretical.ObserverModel
	GetUniverseMetadata() map[string]any
	AddToUniverseMetadata(key string, value any)
	DeleteFromUniverseMetadata(key string) (any, bool)
	UpdateUniverseMetadata(md map[string]any)
}
type ObserverFn func(ctx context.Context, args ObserverExecutorArgs) (bool, error)

// ActionExecutorArgs provides context and operations available to action functions.
// Consumers receive this interface as a parameter in ActionFn — they do not implement it.
// External code that implements this interface (e.g., test mocks) must add the EmitEvent method.
type ActionExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseCanonicalName() string
	GetUniverseId() string
	GetEvent() Event
	GetAction() theoretical.ActionModel
	GetActionType() ActionType
	GetSnapshot() *MachineSnapshot
	GetUniverseMetadata() map[string]any
	AddToUniverseMetadata(key string, value any)
	DeleteFromUniverseMetadata(key string) (any, bool)
	UpdateUniverseMetadata(md map[string]any)

	// EmitEvent queues an internal event to be processed after all entry actions complete.
	// The emitted event is matched against the current reality's On handlers. If a transition
	// is approved (conditions pass), the machine advances automatically.
	//
	// EmitEvent is only effective within entry actions (both reality-level and constants-level).
	// Calling EmitEvent from exit or transition actions is a no-op and logs a warning.
	//
	// Multiple EmitEvent calls accumulate events in FIFO order. The first event that triggers
	// an approved transition wins; subsequent events are discarded (the reality has changed).
	//
	// Emitted events that chain (A's entry emits → transitions to B → B's entry emits → transitions to C)
	// are supported up to a maximum depth of 10 to prevent infinite loops.
	//
	// If the depth limit is exceeded (e.g., A → B → A → B → ...), an error is returned and the
	// universe is left in an unrecoverable state. The caller should treat the machine instance
	// as invalid. This scenario indicates a bug in the state machine definition.
	EmitEvent(eventName string, data map[string]any)
}
type ActionFn func(ctx context.Context, args ActionExecutorArgs) error

type InvokeExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseCanonicalName() string
	GetUniverseId() string
	GetEvent() Event
	GetInvoke() theoretical.InvokeModel
	GetUniverseMetadata() map[string]any
	AddToUniverseMetadata(key string, value any)
	DeleteFromUniverseMetadata(key string) (any, bool)
	UpdateUniverseMetadata(md map[string]any)
}
type InvokeFn func(ctx context.Context, args InvokeExecutorArgs)

type ConditionExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseCanonicalName() string
	GetUniverseId() string
	GetEvent() Event
	GetCondition() theoretical.ConditionModel
	GetUniverseMetadata() map[string]any
	AddToUniverseMetadata(key string, value any)
	DeleteFromUniverseMetadata(key string) (any, bool)
	UpdateUniverseMetadata(md map[string]any)
}
type ConditionFn func(ctx context.Context, args ConditionExecutorArgs) (bool, error)
