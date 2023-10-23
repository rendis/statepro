package statepro

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
)

type QuantumMachine interface {
	Init(ctx context.Context, machineContext any) error

	SendEvent(ctx context.Context, event experimental.Event) error

	LazySendEvent(ctx context.Context, event experimental.Event) error

	GetSnapshot() *experimental.ExQuantumMachineSnapshot

	LoadSnapshot(snapshot *experimental.ExQuantumMachineSnapshot) error
}

type Universe interface {
	// HandleEvent handles an event where depending on the state of the universe
	// it will call Start, SendEvent or SendEventToReality
	HandleEvent(ctx context.Context, realityName *string, event experimental.Event) ([]string, experimental.Event, error)

	// CanHandleEvent returns true if the universe can handle the given event
	// A universe can handle an event if all the following conditions are true:
	// - not in superposition state
	// - current reality is established and not final
	// - the current reality can handle the event
	CanHandleEvent(evt experimental.Event) bool

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
	Start(ctx context.Context) ([]string, experimental.Event, error)

	// StartOnReality starts the universe on the given reality
	// StartOnReality set the given reality as the current reality and execute:
	// - always operations
	// - initial operations
	StartOnReality(ctx context.Context, realityName string) ([]string, experimental.Event, error)

	// SendEvent sends an event to:
	// - the current reality if the universe is in not in superposition state
	// - the realities candidates to be the current reality if the universe is in superposition state
	SendEvent(ctx context.Context, event experimental.Event) ([]string, experimental.Event, error)

	// SendEventToReality sends an event to a specific reality in a universe in superposition state
	// Can be used to send events to current reality, preferred use SendEvent method for this purpose
	SendEventToReality(ctx context.Context, realityName string, event experimental.Event) ([]string, experimental.Event, error)

	// GetSnapshot returns a snapshot of the universe
	GetSnapshot() experimental.ExUniverseSnapshot

	// LoadSnapshot loads a snapshot of the universe
	LoadSnapshot(snapshot experimental.ExUniverseSnapshot) error
}

type EventBuilder interface {
	// SetEventName sets the event name
	SetEventName(name string) EventBuilder

	// SetData sets the event data
	SetData(data map[string]any) EventBuilder

	// SetEvtType sets the event type
	SetEvtType(evtType experimental.EventType) EventBuilder

	// Build returns the event
	Build() experimental.Event
}
