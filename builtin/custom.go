package builtin

import (
	"errors"
	"github.com/rendis/statepro/v3/instrumentation"
	"regexp"
	"strings"
)

const customPattern = `^[a-zA-Z][a-zA-Z0-9_:.-]*[a-zA-Z0-9]$`

var compiledPattern = regexp.MustCompile(customPattern)
var InvalidSrcError = errors.New("invalid src")

var observerRegistry = map[string]instrumentation.ObserverFn{}
var actionRegistry = map[string]instrumentation.ActionFn{}
var invokeRegistry = map[string]instrumentation.InvokeFn{}
var conditionRegistry = map[string]instrumentation.ConditionFn{}

func RegisterObserver(src string, fn instrumentation.ObserverFn) error {
	src, err := normalizeSrc(src)
	if err != nil {
		return err
	}
	observerRegistry[src] = fn
	return nil
}

func RegisterAction(src string, fn instrumentation.ActionFn) error {
	src, err := normalizeSrc(src)
	if err != nil {
		return err
	}
	actionRegistry[src] = fn
	return nil
}

func RegisterInvoke(src string, fn instrumentation.InvokeFn) error {
	src, err := normalizeSrc(src)
	if err != nil {
		return err
	}
	invokeRegistry[src] = fn
	return nil
}

func RegisterCondition(src string, fn instrumentation.ConditionFn) error {
	src, err := normalizeSrc(src)
	if err != nil {
		return err
	}
	conditionRegistry[src] = fn
	return nil
}

func getObserver(src string) instrumentation.ObserverFn {
	return observerRegistry[src]
}

func getAction(src string) instrumentation.ActionFn {
	return actionRegistry[src]
}

func getInvoke(src string) instrumentation.InvokeFn {
	return invokeRegistry[src]
}

func getCondition(src string) instrumentation.ConditionFn {
	return conditionRegistry[src]
}

func normalizeSrc(src string) (string, error) {
	src = strings.TrimSpace(src)

	if !compiledPattern.MatchString(src) {
		return "", InvalidSrcError
	}

	return src, nil
}
