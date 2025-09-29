package instrumentation

import (
	"context"
)


type QuantumMachine interface {
	// Init initializes the quantum machine.
	// Initialize the machine with the given machine context.
	Init(ctx context.Context, machineContext any) error

	// InitWithEvent initializes the quantum machine with an event.
	// Initialize the machine with the given machine context and event.
	InitWithEvent(ctx context.Context, machineContext any, event Event) error

	// SendEvent sends an event to all universes that can handle it.
	// Returns true if the event was handled by at least one universe.
	SendEvent(ctx context.Context, event Event) (bool, error)

	// LoadSnapshot loads a snapshot into the quantum machine.
	LoadSnapshot(snapshot *MachineSnapshot, machineContext any) error

	// GetSnapshot returns the current snapshot of the quantum machine.
	GetSnapshot() *MachineSnapshot

	// ReplayOnEntry replays the entry actions for the current realities.
	ReplayOnEntry(ctx context.Context) error

	// PositionMachine positions the quantum machine in a specific universe and reality.
	// Parameters:
	//   - ctx: Context for execution
	//   - machineContext: Machine context to use
	//   - universeID: Target universe identifier
	//   - realityID: Target reality (state) identifier
	//   - executeFlow: If true, executes full entry flow (entry actions, always transitions).
	//                  If false, only positions the machine without executing any actions.
	// Returns error if universe/reality doesn't exist or positioning fails
	PositionMachine(ctx context.Context, machineContext any, universeID string, realityID string, executeFlow bool) error
}
