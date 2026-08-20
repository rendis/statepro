package builtin

import (
	"context"
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// Boundary-focused tests: kill CONDITIONALS_BOUNDARY / INVERT_LOGICAL survivors
// that change observable observer decisions at exact limits.

func TestGreaterThanEqualCounter_ExactCount(t *testing.T) {
	ctx := context.Background()
	event1 := &mockEvent{name: "event1"}
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {event1, event1}, // exactly 2
		},
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer: theoretical.ObserverModel{
			Src:  "test",
			Args: map[string]any{"event1": 2}, // require exactly the accumulated count
		},
	}

	ok, err := GreaterThanEqualCounter(ctx, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("exact count must satisfy >= / kill len(events) < count → <= mutant")
	}
}

func TestTotalEventsBetweenLimits_ExactMinimum(t *testing.T) {
	ctx := context.Background()
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {&mockEvent{name: "a"}}, // total = 1
		},
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer: theoretical.ObserverModel{
			Src: "test",
			Args: map[string]any{
				"minimum": 1,
				"maximum": 10,
			},
		},
	}

	ok, err := TotalEventsBetweenLimits(ctx, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("total==minimum must pass (kills >= → > boundary mutant)")
	}
}

func TestTotalEventsBetweenLimits_ExactMaximum(t *testing.T) {
	ctx := context.Background()
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {
				&mockEvent{name: "a"},
				&mockEvent{name: "b"},
				&mockEvent{name: "c"},
			}, // total = 3
		},
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer: theoretical.ObserverModel{
			Src: "test",
			Args: map[string]any{
				"minimum": 0,
				"maximum": 3,
			},
		},
	}

	ok, err := TotalEventsBetweenLimits(ctx, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("total==maximum must pass (kills <= → < boundary mutant)")
	}
}

func TestTotalEventsBetweenLimits_BelowMinimum(t *testing.T) {
	ctx := context.Background()
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {}, // total = 0
		},
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer: theoretical.ObserverModel{
			Src: "test",
			Args: map[string]any{
				"minimum": 1,
				"maximum": 5,
			},
		},
	}

	ok, err := TotalEventsBetweenLimits(ctx, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("total below minimum must fail (kills && → || logical mutant)")
	}
}

func TestTotalEventsBetweenLimits_AboveMaximum(t *testing.T) {
	ctx := context.Background()
	stats := &mockAccumulatorStatistics{
		events: map[string][]instrumentation.Event{
			"reality1": {
				&mockEvent{name: "a"},
				&mockEvent{name: "b"},
				&mockEvent{name: "c"},
			},
		},
	}
	args := &mockObserverExecutorArgs{
		accumulatorStats: stats,
		realityName:      "reality1",
		observer: theoretical.ObserverModel{
			Src: "test",
			Args: map[string]any{
				"minimum": 0,
				"maximum": 2,
			},
		},
	}

	ok, err := TotalEventsBetweenLimits(ctx, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("total above maximum must fail")
	}
}
