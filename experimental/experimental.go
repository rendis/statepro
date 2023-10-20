package experimental

import (
	"context"
)

type Universe interface {
	// IsActive returns true if the universe is active
	// A universe is active if:
	// - has been initialized &&
	// - || it is in superposition state
	// - || the current reality is established and it is not final
	IsActive() bool

	// IsInitialized returns true if the universe has been initialized
	IsInitialized() bool

	// HandleExternalEvent handles an external event
	// Depending on the state of the universe it will call Start, StartOnReality, SendEvent or SendEventToReality
	HandleExternalEvent(ctx context.Context, realityName *string, evt Event) ([]string, Event, error)

	// PlaceOn sets the given reality as the current reality
	// PlaceOn not execute initials operations, only set the current reality
	PlaceOn(ctx context.Context, realityName string) error

	// PlaceOnInitial sets the initial reality as the current reality
	// PlaceOnInitial not execute initials operations, only set the current reality
	PlaceOnInitial(ctx context.Context) error

	// Start starts the universe on the default reality (initial reality)
	// Start set initial reality as the current reality and execute the initials operations
	Start(ctx context.Context) ([]string, Event, error)

	// StartOnReality starts the universe on the given reality
	// StartOnReality set the given reality as the current reality and execute the initials operations
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
