package instrumentation

import (
	"context"
)

type QuantumMachine interface {
	// Init initializes the quantum machine.
	// Initialize the machine based on the values in 'initials'
	Init(ctx context.Context, machineContext any) error

	// SendEvent sends an event to all universes that can handle it.
	SendEvent(ctx context.Context, event Event) error

	// LoadSnapshot loads a snapshot into the quantum machine.
	LoadSnapshot(snapshot *MachineSnapshot, machineContext any) error

	// GetSnapshot returns the current snapshot of the quantum machine.
	GetSnapshot() *MachineSnapshot
}
