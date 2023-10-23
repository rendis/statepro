package experimental

import (
	"context"
	"github.com/rendis/statepro/v3/theoretical"
)

// QuantumMachineLaws is the interface that must be implemented by a quantum machine.
// The quantum machine laws are the laws that may be applied to each universe.
type QuantumMachineLaws interface {
	// GetQuantumMachineId returns the quantum machine id.
	GetQuantumMachineId() string

	// GetQuantumMachineDescription returns the quantum machine description.
	GetQuantumMachineDescription() string

	// ExecuteObserver executes an observer in the quantum machine.
	// Parameters:
	// 	- ctx: the context
	// 	- quantumMachineContext: the quantum machine context
	// 	- accumulatorStatistics: the accumulator statistics
	// 	- event: last event triggered (contained in the accumulator statistics)
	// 	- observer: the observer to execute
	// Returns:
	// 	- bool: observer result
	// 	- error: if an error occurs
	ExecuteObserver(
		ctx context.Context,
		quantumMachineContext any,
		accumulatorStatistics AccumulatorStatistics,
		event Event,
		observer theoretical.ObserverModel,
	) (bool, error)

	// ExecuteAction executes an action in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- quantumMachineContext: the quantum machine context
	// 	- event: last event triggered
	// 	- action: the action to execute
	// Returns:
	// 	- error: if an error occurs
	ExecuteAction(
		ctx context.Context,
		quantumMachineContext any,
		event Event,
		action theoretical.ActionModel,
	) error

	// ExecuteInvoke executes an invoke in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- quantumMachineContext: the quantum machine context
	// 	- event: last event triggered
	// 	- invoke: the invoke to execute
	ExecuteInvoke(
		ctx context.Context,
		quantumMachineContext any,
		event Event,
		invoke theoretical.InvokeModel,
	)
}

// UniverseLaws is the interface that must be implemented by a universe.
// The universe laws are the laws that may be applied only to the universe.
type UniverseLaws interface {
	// GetUniverseId returns the universe id.
	// Used to link universe with the universe json definition
	GetUniverseId() string

	// GetUniverseDescription returns the universe description.
	GetUniverseDescription() string

	// ExtractObservableKnowledge extracts the universe knowledge from any quantum machine context.
	// This method allows the context segmentation, so the universe only knows and can access to the knowledge that is relevant to it.
	// universeContext can be equal to quantumMachineContext, but it is not recommended.
	// Parameters:
	// 	- quantumMachineContext: the quantum machine context (global context)
	// Returns:
	// 	- universeContext: the universe context (local context)
	// 	- error: if an error occurs
	ExtractObservableKnowledge(quantumMachineContext any) (universeContext any, err error)

	// ExecuteObserver executes an observer in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- universeContext: the universe context
	// 	- accumulatorStatistics: the accumulator statistics
	// 	- event: last event triggered (contained in the accumulator statistics)
	// 	- observer: the observer to execute
	// Returns:
	// 	- bool: observer result
	// 	- error: if an error occurs
	ExecuteObserver(
		ctx context.Context,
		universeContext any,
		accumulatorStatistics AccumulatorStatistics,
		event Event,
		observer theoretical.ObserverModel,
	) (bool, error)

	// ExecuteAction executes an action in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- universeContext: the universe context
	// 	- event: last event triggered
	// 	- action: the action to execute
	// Returns:
	// 	- error: if an error occurs
	ExecuteAction(
		ctx context.Context,
		universeContext any,
		event Event,
		action theoretical.ActionModel,
	) error

	// ExecuteInvoke executes an invoke in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- universeContext: the universe context
	// 	- event: last event triggered
	// 	- invoke: the invoke to execute
	ExecuteInvoke(
		ctx context.Context,
		universeContext any,
		event Event,
		invoke theoretical.InvokeModel,
	)

	// ExecuteCondition executes a condition in the universe.
	// Parameters:
	// 	- conditionName: the condition Name to execute
	// 	- args: the condition arguments
	// 	- universeContext: the universe context
	// 	- event: last event triggered
	// Returns:
	// 	- bool: condition result
	// 	- error: if an error occurs
	ExecuteCondition(
		ctx context.Context,
		conditionName string,
		args map[string]any,
		universeContext any,
		event Event,
	) (bool, error)
}
