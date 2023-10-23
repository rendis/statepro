package experimental

import (
	"context"
)

type Universe interface {
	// HandleEvent handles an event where depending on the state of the universe
	// it will call Start, SendEvent or SendEventToReality
	HandleEvent(ctx context.Context, realityName *string, evt Event) ([]string, Event, error)

	// CanHandleEvent returns true if the universe can handle the given event
	// A universe can handle an event if all the following conditions are true:
	// - not in superposition state
	// - current reality is established and not final
	// - the current reality can handle the event
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

	// SendEvent sends an event to:
	// - the current reality if the universe is in not in superposition state
	// - the realities candidates to be the current reality if the universe is in superposition state
	SendEvent(ctx context.Context, evt Event) ([]string, Event, error)

	// SendEventToReality sends an event to a specific reality in a universe in superposition state
	// Can be used to send events to current reality, preferred use SendEvent method for this purpose
	SendEventToReality(ctx context.Context, realityName string, evt Event) ([]string, Event, error)

	// GetSnapshot returns a snapshot of the universe
	GetSnapshot() ExUniverseSnapshot

	// LoadSnapshot loads a snapshot of the universe
	LoadSnapshot(snapshot ExUniverseSnapshot) error
}
