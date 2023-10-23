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
	// 	- args: the observer executor arguments
	// Returns:
	// 	- bool: observer result
	// 	- error: if an error occurs
	ExecuteObserver(ctx context.Context, args ObserverExecutorArgs) (bool, error)

	// ExecuteAction executes an action in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- args: the action executor arguments
	// Returns:
	// 	- error: if an error occurs
	ExecuteAction(ctx context.Context, args ActionExecutorArgs) error

	// ExecuteInvoke executes an invoke in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- args: the invoke executor arguments
	ExecuteInvoke(ctx context.Context, args InvokeExecutorArgs)
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
	// 	- args: the observer executor arguments
	// Returns:
	// 	- bool: observer result
	// 	- error: if an error occurs
	ExecuteObserver(ctx context.Context, args ObserverExecutorArgs) (bool, error)

	// ExecuteAction executes an action in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- args: the action executor arguments
	// Returns:
	// 	- error: if an error occurs
	ExecuteAction(ctx context.Context, args ActionExecutorArgs) error

	// ExecuteInvoke executes an invoke in the universe.
	// Parameters:
	// 	- ctx: the context
	// 	- args: the invoke executor arguments
	ExecuteInvoke(ctx context.Context, args InvokeExecutorArgs)

	// ExecuteCondition executes a condition in the universe.
	// Parameters:
	// 	- conditionName: the condition Name to execute
	// 	- args: the condition executor arguments
	// Returns:
	// 	- bool: condition result
	// 	- error: if an error occurs
	ExecuteCondition(ctx context.Context, args ConditionExecutorArgs) (bool, error)
}

//--------- ObserverExecutorArgs ---------//

type ObserverExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetAccumulatorStatistics() AccumulatorStatistics
	GetEvent() Event
	GetObserver() theoretical.ObserverModel
}

type observerExecutorArgs struct {
	context               any
	realityName           string
	accumulatorStatistics AccumulatorStatistics
	event                 Event
	observer              theoretical.ObserverModel
}

func (o observerExecutorArgs) GetContext() any {
	return o.context
}

func (o observerExecutorArgs) GetRealityName() string {
	return o.realityName
}

func (o observerExecutorArgs) GetAccumulatorStatistics() AccumulatorStatistics {
	return o.accumulatorStatistics
}

func (o observerExecutorArgs) GetEvent() Event {
	return o.event
}

func (o observerExecutorArgs) GetObserver() theoretical.ObserverModel {
	return o.observer
}

//--------- ActionExecutorArgs ---------//

type ActionExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetEvent() Event
	GetAction() theoretical.ActionModel
}

type actionExecutorArgs struct {
	context     any
	realityName string
	event       Event
	action      theoretical.ActionModel
}

func (a actionExecutorArgs) GetContext() any {
	return a.context
}

func (a actionExecutorArgs) GetRealityName() string {
	return a.realityName
}

func (a actionExecutorArgs) GetEvent() Event {
	return a.event
}

func (a actionExecutorArgs) GetAction() theoretical.ActionModel {
	return a.action
}

//--------- InvokeExecutorArgs ---------//

type InvokeExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetEvent() Event
	GetInvoke() theoretical.InvokeModel
}

type invokeExecutorArgs struct {
	context     any
	realityName string
	event       Event
	invoke      theoretical.InvokeModel
}

func (i invokeExecutorArgs) GetContext() any {
	return i.context
}

func (i invokeExecutorArgs) GetRealityName() string {
	return i.realityName
}

func (i invokeExecutorArgs) GetEvent() Event {
	return i.event
}

func (i invokeExecutorArgs) GetInvoke() theoretical.InvokeModel {
	return i.invoke
}

//--------- ConditionExecutorArgs ---------//

type ConditionExecutorArgs interface {
	GetContext() any
	GetRealityName() string
	GetEvent() Event
	GetCondition() theoretical.ConditionModel
}

type conditionExecutorArgs struct {
	context     any
	realityName string
	event       Event
	condition   theoretical.ConditionModel
}

func (c conditionExecutorArgs) GetContext() any {
	return c.context
}

func (c conditionExecutorArgs) GetRealityName() string {
	return c.realityName
}

func (c conditionExecutorArgs) GetEvent() Event {
	return c.event
}

func (c conditionExecutorArgs) GetCondition() theoretical.ConditionModel {
	return c.condition
}
