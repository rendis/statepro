package experimental

import (
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// --------- laws executor extractors ---------//
func getUniverseObserverExecutor(laws any) instrumentation.ObserverExecutor {
	if executor, ok := laws.(instrumentation.ObserverExecutor); ok {
		return executor
	}
	return nil
}

func getUniverseActionExecutor(laws any) instrumentation.ActionExecutor {
	if executor, ok := laws.(instrumentation.ActionExecutor); ok {
		return executor
	}
	return nil
}

func getUniverseInvokeExecutor(laws any) instrumentation.InvokeExecutor {
	if executor, ok := laws.(instrumentation.InvokeExecutor); ok {
		return executor
	}
	return nil
}

func getUniverseConditionExecutor(laws any) instrumentation.ConditionExecutor {
	if executor, ok := laws.(instrumentation.ConditionExecutor); ok {
		return executor
	}
	return nil
}

func getObservableKnowledgeExtractor(laws any) instrumentation.ObservableKnowledgeExtractorExecutor {
	if executor, ok := laws.(instrumentation.ObservableKnowledgeExtractorExecutor); ok {
		return executor
	}
	return nil
}

// --------- ObserverExecutorArgs ---------//
type observerExecutorArgs struct {
	context               any
	realityName           string
	universeName          string
	accumulatorStatistics instrumentation.AccumulatorStatistics
	event                 instrumentation.Event
	observer              theoretical.ObserverModel
}

func (o *observerExecutorArgs) GetContext() any {
	return o.context
}

func (o *observerExecutorArgs) GetRealityName() string {
	return o.realityName
}

func (o *observerExecutorArgs) GetUniverseName() string {
	return o.universeName
}

func (o *observerExecutorArgs) GetAccumulatorStatistics() instrumentation.AccumulatorStatistics {
	return o.accumulatorStatistics
}

func (o *observerExecutorArgs) GetEvent() instrumentation.Event {
	return o.event
}

func (o *observerExecutorArgs) GetObserver() theoretical.ObserverModel {
	return o.observer
}

// --------- ActionExecutorArgs ---------//
type actionExecutorArgs struct {
	context       any
	realityName   string
	universeName  string
	event         instrumentation.Event
	action        theoretical.ActionModel
	getSnapshotFn func() *instrumentation.MachineSnapshot
}

func (a *actionExecutorArgs) GetContext() any {
	return a.context
}

func (a *actionExecutorArgs) GetRealityName() string {
	return a.realityName
}

func (a *actionExecutorArgs) GetUniverseName() string {
	return a.universeName
}

func (a *actionExecutorArgs) GetEvent() instrumentation.Event {
	return a.event
}

func (a *actionExecutorArgs) GetAction() theoretical.ActionModel {
	return a.action
}

func (a *actionExecutorArgs) GetSnapshot() *instrumentation.MachineSnapshot {
	return a.getSnapshotFn()
}

//--------- InvokeExecutorArgs ---------//

type invokeExecutorArgs struct {
	context      any
	realityName  string
	universeName string
	event        instrumentation.Event
	invoke       theoretical.InvokeModel
}

func (i *invokeExecutorArgs) GetContext() any {
	return i.context
}

func (i *invokeExecutorArgs) GetRealityName() string {
	return i.realityName
}

func (i *invokeExecutorArgs) GetUniverseName() string {
	return i.universeName
}

func (i *invokeExecutorArgs) GetEvent() instrumentation.Event {
	return i.event
}

func (i *invokeExecutorArgs) GetInvoke() theoretical.InvokeModel {
	return i.invoke
}

//--------- ConditionExecutorArgs ---------//

type conditionExecutorArgs struct {
	context      any
	realityName  string
	universeName string
	event        instrumentation.Event
	condition    theoretical.ConditionModel
}

func (c *conditionExecutorArgs) GetContext() any {
	return c.context
}

func (c *conditionExecutorArgs) GetRealityName() string {
	return c.realityName
}

func (c *conditionExecutorArgs) GetUniverseName() string {
	return c.universeName
}

func (c *conditionExecutorArgs) GetEvent() instrumentation.Event {
	return c.event
}

func (c *conditionExecutorArgs) GetCondition() theoretical.ConditionModel {
	return c.condition
}
