package experimental

import (
	"github.com/rendis/statepro/v3/theoretical"
)

// --------- laws executor extractors ---------//
func getUniverseObserverExecutor(laws any) ObserverExecutor {
	if executor, ok := laws.(ObserverExecutor); ok {
		return executor
	}
	return nil
}

func getUniverseActionExecutor(laws any) ActionExecutor {
	if executor, ok := laws.(ActionExecutor); ok {
		return executor
	}
	return nil
}

func getUniverseInvokeExecutor(laws any) InvokeExecutor {
	if executor, ok := laws.(InvokeExecutor); ok {
		return executor
	}
	return nil
}

func getUniverseConditionExecutor(laws any) ConditionExecutor {
	if executor, ok := laws.(ConditionExecutor); ok {
		return executor
	}
	return nil
}

func getObservableKnowledgeExtractor(laws any) ObservableKnowledgeExtractorExecutor {
	if executor, ok := laws.(ObservableKnowledgeExtractorExecutor); ok {
		return executor
	}
	return nil
}

// --------- ObserverExecutorArgs ---------//
type observerExecutorArgs struct {
	context               any
	realityName           string
	universeName          string
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

func (o observerExecutorArgs) GetUniverseName() string {
	return o.universeName
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

// --------- ActionExecutorArgs ---------//
type actionExecutorArgs struct {
	context      any
	realityName  string
	universeName string
	event        Event
	action       theoretical.ActionModel
}

func (a actionExecutorArgs) GetContext() any {
	return a.context
}

func (a actionExecutorArgs) GetRealityName() string {
	return a.realityName
}

func (a actionExecutorArgs) GetUniverseName() string {
	return a.universeName
}

func (a actionExecutorArgs) GetEvent() Event {
	return a.event
}

func (a actionExecutorArgs) GetAction() theoretical.ActionModel {
	return a.action
}

//--------- InvokeExecutorArgs ---------//

type invokeExecutorArgs struct {
	context      any
	realityName  string
	universeName string
	event        Event
	invoke       theoretical.InvokeModel
}

func (i invokeExecutorArgs) GetContext() any {
	return i.context
}

func (i invokeExecutorArgs) GetRealityName() string {
	return i.realityName
}

func (i invokeExecutorArgs) GetUniverseName() string {
	return i.universeName
}

func (i invokeExecutorArgs) GetEvent() Event {
	return i.event
}

func (i invokeExecutorArgs) GetInvoke() theoretical.InvokeModel {
	return i.invoke
}

//--------- ConditionExecutorArgs ---------//

type conditionExecutorArgs struct {
	context      any
	realityName  string
	universeName string
	event        Event
	condition    theoretical.ConditionModel
}

func (c conditionExecutorArgs) GetContext() any {
	return c.context
}

func (c conditionExecutorArgs) GetRealityName() string {
	return c.realityName
}

func (c conditionExecutorArgs) GetUniverseName() string {
	return c.universeName
}

func (c conditionExecutorArgs) GetEvent() Event {
	return c.event
}

func (c conditionExecutorArgs) GetCondition() theoretical.ConditionModel {
	return c.condition
}
