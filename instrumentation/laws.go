package instrumentation

import (
	"context"
	"github.com/rendis/statepro/v3/theoretical"
)

// QuantumMachineLaws is the interface that must be implemented by a quantum machine.
// The quantum machine universeLaws are the universeLaws that may be applied to each universe.
type QuantumMachineLaws interface {
	// GetQuantumMachineId returns the quantum machine id.
	GetQuantumMachineId() string

	// GetQuantumMachineDescription returns the quantum machine description.
	GetQuantumMachineDescription() string
}

// UniverseLaws is the interface that must be implemented by a universe.
// The universe universeLaws are the universeLaws that may be applied only to the universe.
type UniverseLaws interface {
	// GetUniverseId returns the universe id.
	// Used to link universe with the universe json definition
	GetUniverseId() string

	// GetUniverseDescription returns the universe description.
	GetUniverseDescription() string
}

// ConstantsLawsExecutor is the interface that must be implemented by a quantum machine executor.
type ConstantsLawsExecutor interface {
	ExecuteEntryInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
	ExecuteExitInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
	ExecuteEntryAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
	ExecuteExitAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
	ExecuteTransitionInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
	ExecuteTransitionAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
	GetSnapshot() *MachineSnapshot
}

type QuantumMachineExecutorArgs struct {
	Context               any
	RealityName           string
	UniverseID            string
	UniverseCanonicalName string
	Event                 Event
	AccumulatorStatistics AccumulatorStatistics
}

// ObserverExecutor executes an observer in the quantum machine.
// Parameters:
//   - ctx: the Context
//   - args: the observer executor arguments
//
// Returns:
//   - bool: observer result
//   - error: if an error occurs
type ObserverExecutor interface {
	ExecuteObserver(ctx context.Context, args ObserverExecutorArgs) (bool, error)
}
type ObserverFn func(ctx context.Context, args ObserverExecutorArgs) (bool, error)

// ActionExecutor executes an action in the quantum machine.
// Parameters:
//   - ctx: the Context
//   - args: the action executor arguments
//
// Returns:
//   - error: if an error occurs
type ActionExecutor interface {
	ExecuteAction(ctx context.Context, args ActionExecutorArgs) error
}
type ActionFn func(ctx context.Context, args ActionExecutorArgs) error

// InvokeExecutor executes an invoke in the quantum machine.
// Parameters:
//   - ctx: the Context
//   - args: the invoke executor arguments
type InvokeExecutor interface {
	ExecuteInvoke(ctx context.Context, args InvokeExecutorArgs)
}
type InvokeFn func(ctx context.Context, args InvokeExecutorArgs)

// ConditionExecutor executes a condition in the universe.
// Parameters:
//   - conditionName: the condition name to execute
//   - args: the condition executor arguments
//
// Returns:
//   - bool: condition result
//   - error: if an error occurs
type ConditionExecutor interface {
	ExecuteCondition(ctx context.Context, args ConditionExecutorArgs) (bool, error)
}
type ConditionFn func(ctx context.Context, args ConditionExecutorArgs) (bool, error)

// ObservableKnowledgeExtractorExecutor extracts the universe knowledge from any quantum machine Context.
// This method allows the Context segmentation, so the universe only knows and can access to the knowledge that is relevant to it.
// universeContext can be equal to quantumMachineContext, but it is not recommended.
// Parameters:
//   - quantumMachineContext: the quantum machine Context (global Context)
//
// Returns:
//   - universeContext: the universe Context (local Context)
//   - error: if an error occurs
type ObservableKnowledgeExtractorExecutor interface {
	ExtractObservableKnowledge(quantumMachineContext any) (universeContext any, err error)
}

//---------- ExecutorArgs ----------//

type ObserverExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseCanonicalName() string
	GetUniverseId() string
	GetAccumulatorStatistics() AccumulatorStatistics
	GetEvent() Event
	GetObserver() theoretical.ObserverModel
}

type ActionExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseCanonicalName() string
	GetUniverseId() string
	GetEvent() Event
	GetAction() theoretical.ActionModel
	GetActionType() ActionType
	GetSnapshot() *MachineSnapshot
}

type InvokeExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseCanonicalName() string
	GetUniverseId() string
	GetEvent() Event
	GetInvoke() theoretical.InvokeModel
}

type ConditionExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseCanonicalName() string
	GetUniverseId() string
	GetEvent() Event
	GetCondition() theoretical.ConditionModel
}

//---------- Enums ----------//

type ActionType string

const (
	ActionTypeEntry      ActionType = "entry"
	ActionTypeExit       ActionType = "exit"
	ActionTypeTransition ActionType = "transition"
)
