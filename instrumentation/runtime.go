package instrumentation

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrReentrantCall   = errors.New("synchronous callback cannot reenter its owning machine; use callback arguments")
	ErrMachineClosed   = errors.New("quantum machine is closed")
	ErrUnknownObserver = errors.New("observer is not registered")
)

// RuntimeOptions configures opt-in execution policies. Zero limits mean unlimited.
// The zero value preserves the existing observer and invoke lifecycle contracts.
type RuntimeOptions struct {
	MaxConcurrentInvokes int
	CancelInvokesOnExit  bool
	MaxAccumulatedEvents int // Per universe, including copies accumulated for different realities.
	MaxTrackingEntries   int // Retain the most recent entries per universe.
	StrictObservers      bool
}

func (o RuntimeOptions) Validate() error {
	if o.MaxConcurrentInvokes < 0 || o.MaxAccumulatedEvents < 0 || o.MaxTrackingEntries < 0 {
		return fmt.Errorf("runtime limits must not be negative")
	}
	return nil
}

// ResourceLimitError reports rejected admission without starting another task or event copy.
type ResourceLimitError struct {
	Resource string
	Limit    int
}

func (e *ResourceLimitError) Error() string {
	return fmt.Sprintf("%s limit reached (%d)", e.Resource, e.Limit)
}

// RuntimeLifecycle is an optional capability. Close requests cooperative invoke
// cancellation and rejects new execution; WaitInvokes waits until the pool is idle.
// An invoke that ignores its context cannot be forcibly terminated.
type RuntimeLifecycle interface {
	Close() error
	WaitInvokes(context.Context) error
}

// ContextSnapshotProvider adds cancellable waiting and callback reentry detection.
type ContextSnapshotProvider interface {
	GetSnapshotContext(context.Context) (*MachineSnapshot, error)
	LoadSnapshotContext(context.Context, *MachineSnapshot, any) error
}
