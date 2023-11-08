package builtin

import (
	"github.com/rendis/abslog/v3"
	"github.com/rendis/statepro/v3/instrumentation"
)

func SetLogger(logger abslog.AbsLog) {
	abslog.SetLogger(logger)
}

var builtinObserverRegistry = map[string]instrumentation.ObserverFn{
	"builtin:observer:containsAllEvents":       ContainsAllEvents,
	"builtin:observer:containsAtLeastOneEvent": ContainsAtLeastOneEvent,
	"builtin:observer:alwaysTrue":              AlwaysTrue,
	"builtin:observer:greaterThanEqualCounter": GreaterThanEqualCounter,
}

var builtinActionRegistry = map[string]instrumentation.ActionFn{
	"builtin:action:logBasicInfo":       LogBasicInfo,
	"builtin:action:logArgs":            LogArgs,
	"builtin:action:logArgsWithoutKeys": LogArgsWithoutKeys,
	"builtin:action:logJustArgsValues":  LogJustArgsValues,
}

var builtinInvokeRegistry = map[string]instrumentation.InvokeFn{}

var builtinConditionRegistry = map[string]instrumentation.ConditionFn{}

func GetObserver(src string) instrumentation.ObserverFn {
	if fn := builtinObserverRegistry[src]; fn != nil {
		return fn
	}
	return getObserver(src)
}

func GetAction(src string) instrumentation.ActionFn {
	if fn := builtinActionRegistry[src]; fn != nil {
		return fn
	}
	return getAction(src)
}

func GetInvoke(src string) instrumentation.InvokeFn {
	if fn := builtinInvokeRegistry[src]; fn != nil {
		return fn
	}
	return getInvoke(src)
}

func GetCondition(src string) instrumentation.ConditionFn {
	if fn := builtinConditionRegistry[src]; fn != nil {
		return fn
	}
	return getCondition(src)
}
