package experimental

import (
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// --------- ObserverExecutorArgs ---------//
type observerExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeID            string
	universeMetadata      map[string]any
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

func (o *observerExecutorArgs) GetUniverseCanonicalName() string {
	return o.universeCanonicalName
}

func (o *observerExecutorArgs) GetUniverseId() string {
	return o.universeID
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

func (o *observerExecutorArgs) GetUniverseMetadata() map[string]any {
	return o.universeMetadata
}

// --------- ActionExecutorArgs ---------//
type actionExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeID            string
	universeMetadata      map[string]any
	event                 instrumentation.Event
	action                theoretical.ActionModel
	actionType            instrumentation.ActionType
	getSnapshotFn         func() *instrumentation.MachineSnapshot
}

func (a *actionExecutorArgs) GetContext() any {
	return a.context
}

func (a *actionExecutorArgs) GetRealityName() string {
	return a.realityName
}

func (a *actionExecutorArgs) GetUniverseCanonicalName() string {
	return a.universeCanonicalName
}

func (a *actionExecutorArgs) GetUniverseId() string {
	return a.universeID
}

func (a *actionExecutorArgs) GetEvent() instrumentation.Event {
	return a.event
}

func (a *actionExecutorArgs) GetAction() theoretical.ActionModel {
	return a.action
}

func (a *actionExecutorArgs) GetActionType() instrumentation.ActionType {
	return a.actionType
}

func (a *actionExecutorArgs) GetSnapshot() *instrumentation.MachineSnapshot {
	return a.getSnapshotFn()
}

func (a *actionExecutorArgs) GetUniverseMetadata() map[string]any {
	return a.universeMetadata
}

//--------- InvokeExecutorArgs ---------//

type invokeExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeID            string
	universeMetadata      map[string]any
	event                 instrumentation.Event
	invoke                theoretical.InvokeModel
}

func (i *invokeExecutorArgs) GetContext() any {
	return i.context
}

func (i *invokeExecutorArgs) GetRealityName() string {
	return i.realityName
}

func (i *invokeExecutorArgs) GetUniverseCanonicalName() string {
	return i.universeCanonicalName
}

func (i *invokeExecutorArgs) GetUniverseId() string {
	return i.universeID
}

func (i *invokeExecutorArgs) GetEvent() instrumentation.Event {
	return i.event
}

func (i *invokeExecutorArgs) GetInvoke() theoretical.InvokeModel {
	return i.invoke
}

func (i *invokeExecutorArgs) GetUniverseMetadata() map[string]any {
	return i.universeMetadata
}

//--------- ConditionExecutorArgs ---------//

type conditionExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeID            string
	universeMetadata      map[string]any
	event                 instrumentation.Event
	condition             theoretical.ConditionModel
}

func (c *conditionExecutorArgs) GetContext() any {
	return c.context
}

func (c *conditionExecutorArgs) GetRealityName() string {
	return c.realityName
}

func (c *conditionExecutorArgs) GetUniverseCanonicalName() string {
	return c.universeCanonicalName
}

func (c *conditionExecutorArgs) GetUniverseId() string {
	return c.universeID
}

func (c *conditionExecutorArgs) GetEvent() instrumentation.Event {
	return c.event
}

func (c *conditionExecutorArgs) GetCondition() theoretical.ConditionModel {
	return c.condition
}

func (c *conditionExecutorArgs) GetUniverseMetadata() map[string]any {
	return c.universeMetadata
}
