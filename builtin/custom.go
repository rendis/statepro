package builtin

import (
	"errors"
	"regexp"
	"strings"
	"sync"

	"github.com/rendis/statepro/v3/instrumentation"
)

const customPattern = `^[a-zA-Z][a-zA-Z0-9_:.-]*[a-zA-Z0-9]$`

var compiledPattern = regexp.MustCompile(customPattern)
var errorInvalidSrc = errors.New("invalid src")

var observerRegistry = map[string]instrumentation.ObserverFn{}
var actionRegistry = map[string]instrumentation.ActionFn{}
var invokeRegistry = map[string]instrumentation.InvokeFn{}
var conditionRegistry = map[string]instrumentation.ConditionFn{}
var registryMu sync.RWMutex

func RegisterObserver(src string, fn instrumentation.ObserverFn) error {
	src, err := normalizeSrc(src)
	if err != nil {
		return err
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	observerRegistry[src] = fn
	return nil
}

func RegisterAction(src string, fn instrumentation.ActionFn) error {
	src, err := normalizeSrc(src)
	if err != nil {
		return err
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	actionRegistry[src] = fn
	return nil
}

func RegisterInvoke(src string, fn instrumentation.InvokeFn) error {
	src, err := normalizeSrc(src)
	if err != nil {
		return err
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	invokeRegistry[src] = fn
	return nil
}

func RegisterCondition(src string, fn instrumentation.ConditionFn) error {
	src, err := normalizeSrc(src)
	if err != nil {
		return err
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	conditionRegistry[src] = fn
	return nil
}

func getObserver(src string) instrumentation.ObserverFn {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return observerRegistry[src]
}

func getAction(src string) instrumentation.ActionFn {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return actionRegistry[src]
}

func getInvoke(src string) instrumentation.InvokeFn {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return invokeRegistry[src]
}

func getCondition(src string) instrumentation.ConditionFn {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return conditionRegistry[src]
}

func normalizeSrc(src string) (string, error) {
	src = strings.TrimSpace(src)

	if !compiledPattern.MatchString(src) {
		return "", errorInvalidSrc
	}

	return src, nil
}
