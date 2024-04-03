package instrumentation

import (
	"context"
)

type QuantumMachine interface {
	// Init initializes the quantum machine.
	// Initialize the machine with the given machine context.
	Init(ctx context.Context, machineContext any) error

	// SendEvent sends an event to all universes that can handle it.
	// Returns true if the event was handled by at least one universe.
	SendEvent(ctx context.Context, event Event) (bool, error)

	// LoadSnapshot loads a snapshot into the quantum machine.
	LoadSnapshot(snapshot *MachineSnapshot, machineContext any) error

	// GetSnapshot returns the current snapshot of the quantum machine.
	GetSnapshot() *MachineSnapshot

	// ReplayOnEntry replays the entry actions for the current realities.
	ReplayOnEntry(ctx context.Context) error
}
