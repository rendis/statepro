package experimental

import (
	"log/slog"
	"sync"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

func cloneAnyMap(src map[string]any) map[string]any {
	if src == nil {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func withMetadataLock(mu *sync.Mutex, fn func()) {
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	fn()
}

func metaGet(mu *sync.Mutex, md map[string]any) map[string]any {
	var out map[string]any
	withMetadataLock(mu, func() {
		out = cloneAnyMap(md)
	})
	return out
}

func metaAdd(mu *sync.Mutex, md *map[string]any, key string, value any) {
	withMetadataLock(mu, func() {
		if *md == nil {
			*md = make(map[string]any)
		}
		(*md)[key] = value
	})
}

func metaDelete(mu *sync.Mutex, md map[string]any, key string) (any, bool) {
	var value any
	var ok bool
	withMetadataLock(mu, func() {
		value, ok = md[key]
		if ok {
			delete(md, key)
		}
	})
	return value, ok
}

func metaUpdate(mu *sync.Mutex, md *map[string]any, src map[string]any) {
	withMetadataLock(mu, func() {
		if *md == nil {
			*md = make(map[string]any)
		}
		for k := range *md {
			delete(*md, k)
		}
		for k, v := range src {
			(*md)[k] = v
		}
	})
}

// --------- ObserverExecutorArgs ---------//

type observerExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeID            string
	universeMetadata      map[string]any
	metadataMu            *sync.Mutex
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
	return metaGet(o.metadataMu, o.universeMetadata)
}

func (o *observerExecutorArgs) AddToUniverseMetadata(key string, value any) {
	metaAdd(o.metadataMu, &o.universeMetadata, key, value)
}

func (o *observerExecutorArgs) DeleteFromUniverseMetadata(key string) (any, bool) {
	return metaDelete(o.metadataMu, o.universeMetadata, key)
}

func (o *observerExecutorArgs) UpdateUniverseMetadata(md map[string]any) {
	metaUpdate(o.metadataMu, &o.universeMetadata, md)
}

// --------- ActionExecutorArgs ---------//

type actionExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeID            string
	universeMetadata      map[string]any
	metadataMu            *sync.Mutex
	event                 instrumentation.Event
	action                theoretical.ActionModel
	actionType            instrumentation.ActionType
	getSnapshotFn         func() *instrumentation.MachineSnapshot
	emittedEvents         *[]instrumentation.EmittedEvent
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
	return metaGet(a.metadataMu, a.universeMetadata)
}

func (a *actionExecutorArgs) AddToUniverseMetadata(key string, value any) {
	metaAdd(a.metadataMu, &a.universeMetadata, key, value)
}

func (a *actionExecutorArgs) DeleteFromUniverseMetadata(key string) (any, bool) {
	return metaDelete(a.metadataMu, a.universeMetadata, key)
}

func (a *actionExecutorArgs) UpdateUniverseMetadata(md map[string]any) {
	metaUpdate(a.metadataMu, &a.universeMetadata, md)
}

func (a *actionExecutorArgs) EmitEvent(eventName string, data map[string]any) {
	if eventName == "" {
		return
	}
	if a.emittedEvents == nil {
		slog.Warn("EmitEvent called outside entry action context — event ignored",
			"actionType", a.actionType,
			"reality", a.realityName,
			"event", eventName,
		)
		return
	}
	*a.emittedEvents = append(*a.emittedEvents, instrumentation.EmittedEvent{Name: eventName, Data: data})
}

//--------- InvokeExecutorArgs ---------//

type invokeExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeID            string
	universeMetadata      map[string]any
	metadataMu            *sync.Mutex
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
	return metaGet(i.metadataMu, i.universeMetadata)
}

func (i *invokeExecutorArgs) AddToUniverseMetadata(key string, value any) {
	metaAdd(i.metadataMu, &i.universeMetadata, key, value)
}

func (i *invokeExecutorArgs) DeleteFromUniverseMetadata(key string) (any, bool) {
	return metaDelete(i.metadataMu, i.universeMetadata, key)
}

func (i *invokeExecutorArgs) UpdateUniverseMetadata(md map[string]any) {
	metaUpdate(i.metadataMu, &i.universeMetadata, md)
}

//--------- ConditionExecutorArgs ---------//

type conditionExecutorArgs struct {
	context               any
	realityName           string
	universeCanonicalName string
	universeID            string
	universeMetadata      map[string]any
	metadataMu            *sync.Mutex
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
	return metaGet(c.metadataMu, c.universeMetadata)
}

func (c *conditionExecutorArgs) AddToUniverseMetadata(key string, value any) {
	metaAdd(c.metadataMu, &c.universeMetadata, key, value)
}

func (c *conditionExecutorArgs) DeleteFromUniverseMetadata(key string) (any, bool) {
	return metaDelete(c.metadataMu, c.universeMetadata, key)
}

func (c *conditionExecutorArgs) UpdateUniverseMetadata(md map[string]any) {
	metaUpdate(c.metadataMu, &c.universeMetadata, md)
}
