package experimental

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

// constantsLawsExecutor is the interface that must be implemented by a quantum machine executor.
type constantsLawsExecutor interface {
	ExecuteEntryInvokes(ctx context.Context, args *quantumMachineExecutorArgs)
	ExecuteExitInvokes(ctx context.Context, args *quantumMachineExecutorArgs)
	ExecuteEntryAction(ctx context.Context, args *quantumMachineExecutorArgs) error
	ExecuteExitAction(ctx context.Context, args *quantumMachineExecutorArgs) error
}

type quantumMachineExecutorArgs struct {
	context               any
	realityName           string
	universeName          string
	event                 Event
	accumulatorStatistics AccumulatorStatistics
}

// ObserverExecutor executes an observer in the quantum machine.
// Parameters:
//   - ctx: the context
//   - args: the observer executor arguments
//
// Returns:
//   - bool: observer result
//   - error: if an error occurs
type ObserverExecutor interface {
	ExecuteObserver(ctx context.Context, args ObserverExecutorArgs) (bool, error)
}

// ActionExecutor executes an action in the quantum machine.
// Parameters:
//   - ctx: the context
//   - args: the action executor arguments
//
// Returns:
//   - error: if an error occurs
type ActionExecutor interface {
	ExecuteAction(ctx context.Context, args ActionExecutorArgs) error
}

// InvokeExecutor executes an invoke in the quantum machine.
// Parameters:
//   - ctx: the context
//   - args: the invoke executor arguments
type InvokeExecutor interface {
	ExecuteInvoke(ctx context.Context, args InvokeExecutorArgs)
}

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

// ObservableKnowledgeExtractorExecutor extracts the universe knowledge from any quantum machine context.
// This method allows the context segmentation, so the universe only knows and can access to the knowledge that is relevant to it.
// universeContext can be equal to quantumMachineContext, but it is not recommended.
// Parameters:
//   - quantumMachineContext: the quantum machine context (global context)
//
// Returns:
//   - universeContext: the universe context (local context)
//   - error: if an error occurs
type ObservableKnowledgeExtractorExecutor interface {
	ExtractObservableKnowledge(quantumMachineContext any) (universeContext any, err error)
}

//---------- ExecutorArgs ----------//

type ObserverExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseName() string
	GetAccumulatorStatistics() AccumulatorStatistics
	GetEvent() Event
	GetObserver() theoretical.ObserverModel
}

type ActionExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseName() string
	GetEvent() Event
	GetAction() theoretical.ActionModel
}

type InvokeExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseName() string
	GetEvent() Event
	GetInvoke() theoretical.InvokeModel
}

type ConditionExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetUniverseName() string
	GetEvent() Event
	GetCondition() theoretical.ConditionModel
}
