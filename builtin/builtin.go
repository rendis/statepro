package builtin

import (
	"github.com/rendis/abslog/v3"
	"github.com/rendis/statepro/v3/instrumentation"
)

func SetLogger(logger abslog.AbsLog) {
	abslog.SetLogger(logger)
}

var observerRegistry = map[string]instrumentation.ObserverFn{
	"builtin:observer:containsAllEvents":       ContainsAllEvents,
	"builtin:observer:containsAtLeastOneEvent": ContainsAtLeastOneEvent,
	"builtin:observer:alwaysTrue":              AlwaysTrue,
	"builtin:observer:greaterThanEqualCounter": GreaterThanEqualCounter,
}

var actionRegistry = map[string]instrumentation.ActionFn{
	"builtin:action:logBasicInfo":       LogBasicInfo,
	"builtin:action:logArgs":            LogArgs,
	"builtin:action:logArgsWithoutKeys": LogArgsWithoutKeys,
	"builtin:action:logJustArgsValues":  LogJustArgsValues,
}

func GetBuiltinObserver(src string) instrumentation.ObserverFn {
	return observerRegistry[src]
}

func GetBuiltinAction(src string) instrumentation.ActionFn {
	return actionRegistry[src]
}
