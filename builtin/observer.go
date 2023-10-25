package builtin

import (
	"context"
	"github.com/rendis/statepro/v3/instrumentation"
)

// ContainsAllEvents builtin observer (builtin:observer:containsAllEvents)
// Checks if the accumulated events contains all expected events received as model args.
// Valid args:
//   - map[string]string (key: event name, value: event value)
//
// Any other type will return false.
func ContainsAllEvents(_ context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
	// get accumulated events
	statistics := args.GetAccumulatorStatistics()
	if statistics == nil {
		return false, nil
	}
	accumulatedEvents := statistics.GetRealityEvents(args.GetRealityName())
	if len(accumulatedEvents) == 0 {
		return false, nil
	}

	// iterate over expected events from model args
	for _, expectedEventName := range args.GetObserver().Args {
		name, isString := expectedEventName.(string)
		if !isString {
			return false, nil
		}
		if _, ok := accumulatedEvents[name]; !ok {
			return false, nil
		}
	}

	return true, nil
}

// ContainsAtLeastOneEvent builtin observer (builtin:observer:containsAtLeastOneEvent)
// Checks if the accumulated events contains at least one expected event received as model args.
// Valid args:
//   - map[string]string (key: event name, value: event value)
//
// Any other type will return false.
func ContainsAtLeastOneEvent(_ context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
	// get accumulated events
	statistics := args.GetAccumulatorStatistics()
	if statistics == nil {
		return false, nil
	}
	accumulatedEvents := statistics.GetRealityEvents(args.GetRealityName())
	if len(accumulatedEvents) == 0 {
		return false, nil
	}

	// iterate over expected events from model args
	for _, expectedEventName := range args.GetObserver().Args {
		name, isString := expectedEventName.(string)
		if !isString {
			return false, nil
		}
		if _, ok := accumulatedEvents[name]; ok {
			return true, nil
		}
	}

	return false, nil
}

// AlwaysTrue builtin observer (builtin:observer:alwaysTrue)
// Always returns true, in other words, returns true for any event.
// Valid args:
//   - none
func AlwaysTrue(_ context.Context, _ instrumentation.ObserverExecutorArgs) (bool, error) {
	return true, nil
}

// GreaterThanEqualCounter builtin observer (builtin:observer:greaterThanEqualCounter)
// Checks if the accumulated events appears at least the expected number of times.
// Valid args:
//   - map[string]int (key: event name, value: minimum number of times)
func GreaterThanEqualCounter(_ context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
	// get accumulated events
	statistics := args.GetAccumulatorStatistics()
	if statistics == nil {
		return false, nil
	}
	accumulatedEvents := statistics.GetAllRealityEvents(args.GetRealityName())
	if len(accumulatedEvents) == 0 {
		return false, nil
	}

	// iterate over expected events from model args
	for expectedEventName, expectedEventCounter := range args.GetObserver().Args {
		count, isInt := expectedEventCounter.(int)
		if !isInt {
			return false, nil
		}
		events, ok := accumulatedEvents[expectedEventName]
		if !ok {
			return false, nil
		}

		if len(events) < count {
			return false, nil
		}
	}

	return true, nil
}