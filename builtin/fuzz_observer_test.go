package builtin

import (
	"context"
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// FuzzBuiltinObserverArgs feeds hostile arg maps into builtin observers.
func FuzzBuiltinObserverArgs(f *testing.F) {
	f.Add("event1", int64(1), int64(10), true)
	f.Add("", int64(0), int64(0), false)
	f.Add("x", int64(-1), int64(1000000), true)

	f.Fuzz(func(t *testing.T, eventName string, minV, maxV int64, includeCounter bool) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("observer panicked: %v", r)
			}
		}()

		ev := &mockEvent{name: eventName}
		stats := &mockAccumulatorStatistics{
			events: map[string][]instrumentation.Event{
				"r": {ev},
			},
		}
		args := map[string]any{
			"0":       eventName,
			"minimum": minV,
			"maximum": maxV,
		}
		if includeCounter {
			args[eventName] = minV
		}
		obs := &mockObserverExecutorArgs{
			realityName:      "r",
			accumulatorStats: stats,
			observer:         theoretical.ObserverModel{Src: "fuzz", Args: args},
			event:            ev,
		}
		ctx := context.Background()
		_, _ = ContainsAllEvents(ctx, obs)
		_, _ = ContainsAtLeastOneEvent(ctx, obs)
		_, _ = GreaterThanEqualCounter(ctx, obs)
		_, _ = TotalEventsBetweenLimits(ctx, obs)
		_, _ = AlwaysTrue(ctx, obs)
	})
}
