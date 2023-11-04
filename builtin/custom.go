package builtin

import (
	"fmt"
	"github.com/rendis/statepro/v3/instrumentation"
	"regexp"
	"strings"
)

const observerOperationType = "observer"
const customObserverPattern = `^custom:observer:[a-zA-Z][a-zA-Z0-9]*$`

var customObserverRegistry = map[string]instrumentation.ObserverFn{}

const actionOperationType = "action"
const customActionPattern = `^custom:action:[a-zA-Z][a-zA-Z0-9]*$`

var customActionRegistry = map[string]instrumentation.ActionFn{}

const invokeOperationType = "invoke"
const customInvokePattern = `^custom:invoke:[a-zA-Z][a-zA-Z0-9]*$`

var customInvokeRegistry = map[string]instrumentation.InvokeFn{}

const conditionOperationType = "condition"
const customConditionPattern = `^custom:condition:[a-zA-Z][a-zA-Z0-9]*$`

var customConditionRegistry = map[string]instrumentation.ConditionFn{}

var patterns = map[string]*regexp.Regexp{
	observerOperationType:  regexp.MustCompile(customObserverPattern),
	actionOperationType:    regexp.MustCompile(customActionPattern),
	invokeOperationType:    regexp.MustCompile(customInvokePattern),
	conditionOperationType: regexp.MustCompile(customConditionPattern),
}

func RegisterCustomObserver(src string, fn instrumentation.ObserverFn) error {
	normalizedSrc := normalizeName(src, observerOperationType)
	pattern := patterns[observerOperationType]
	if !pattern.MatchString(normalizedSrc) {
		return fmt.Errorf("invalid custom observer name '%s', must match pattern '%s'", src, customObserverPattern)
	}
	customObserverRegistry[src] = fn
	return nil
}

func RegisterCustomAction(src string, fn instrumentation.ActionFn) error {
	normalizedSrc := normalizeName(src, actionOperationType)
	pattern := patterns["action"]
	if !pattern.MatchString(normalizedSrc) {
		return fmt.Errorf("invalid custom action name '%s', must match pattern '%s'", src, customActionPattern)
	}
	customActionRegistry[src] = fn
	return nil
}

func RegisterCustomInvoke(src string, fn instrumentation.InvokeFn) error {
	normalizedSrc := normalizeName(src, invokeOperationType)
	pattern := patterns["invoke"]
	if !pattern.MatchString(normalizedSrc) {
		return fmt.Errorf("invalid custom invoke name '%s', must match pattern '%s'", src, customInvokePattern)
	}
	customInvokeRegistry[src] = fn
	return nil
}

func RegisterCustomCondition(src string, fn instrumentation.ConditionFn) error {
	normalizedSrc := normalizeName(src, conditionOperationType)
	pattern := patterns["condition"]
	if !pattern.MatchString(normalizedSrc) {
		return fmt.Errorf("invalid custom condition name '%s', must match pattern '%s'", src, customConditionPattern)
	}
	customConditionRegistry[normalizedSrc] = fn
	return nil
}

func getCustomObserver(src string) instrumentation.ObserverFn {
	return customObserverRegistry[src]
}

func getCustomAction(src string) instrumentation.ActionFn {
	return customActionRegistry[src]
}

func getCustomInvoke(src string) instrumentation.InvokeFn {
	return customInvokeRegistry[src]
}

func getCustomCondition(src string) instrumentation.ConditionFn {
	return customConditionRegistry[src]
}

func normalizeName(name string, operationType string) string {
	fullPrefix := fmt.Sprintf("custom:%s:", operationType)
	if !strings.HasPrefix(name, fullPrefix) {
		name = fullPrefix + name
	}
	return name
}
