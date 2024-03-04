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

type QuantumMachineExecutorArgs struct {
	Context               any
	RealityName           string
	UniverseID            string
	UniverseCanonicalName string
	Event                 Event
	AccumulatorStatistics AccumulatorStatistics
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
}
type ObserverFn func(ctx context.Context, args ObserverExecutorArgs) (bool, error)

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
}
type ConditionFn func(ctx context.Context, args ConditionExecutorArgs) (bool, error)
