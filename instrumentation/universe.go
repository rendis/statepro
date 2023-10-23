package instrumentation

import (
	"context"
)

type Universe interface {
	// HandleEvent handles an Event where depending on the state of the universe
	HandleEvent(ctx context.Context, realityName *string, event Event) ([]string, Event, error)

	// CanHandleEvent returns true if the universe can handle the given Event
	// A universe can handle an Event if all the following conditions are true:
	// - not in superposition state
	// - current reality is established and not final
	// - the current reality can handle the Event
	CanHandleEvent(evt Event) bool

	// IsActive returns true if the universe is active
	// A universe is active if:
	// - has been initialized &&
	// - || it is in superposition state
	// - || the current reality is established and it is not final
	IsActive() bool

	// IsInitialized returns true if the universe has been initialized
	IsInitialized() bool

	// PlaceOn sets the given reality as the current reality
	// PlaceOn not execute always, initial or exit operations, only set the current reality
	PlaceOn(realityName string) error

	// PlaceOnInitial sets the initial reality as the current reality
	// PlaceOnInitial not execute initials operations, only set the current reality
	PlaceOnInitial() error

	// Start starts the universe on the default reality (initial reality)
	// Start set initial reality as the current reality and execute:
	// - always operations
	// - initial operations
	Start(ctx context.Context) ([]string, Event, error)

	// StartOnReality starts the universe on the given reality
	// StartOnReality set the given reality as the current reality and execute:
	// - always operations
	// - initial operations
	StartOnReality(ctx context.Context, realityName string) ([]string, Event, error)

	// GetSnapshot returns a snapshot of the universe
	GetSnapshot() SerializedUniverseSnapshot

	// LoadSnapshot loads a snapshot of the universe
	LoadSnapshot(snapshot SerializedUniverseSnapshot) error
}
