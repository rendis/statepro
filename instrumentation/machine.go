package instrumentation

import (
	"context"
)


type QuantumMachine interface {
	// Init initializes the quantum machine by processing all initial universe references.
	// For each reference in the machine model's Initials, it determines whether the reference
	// points to a universe-only or universe+reality, executes the corresponding initialization
	// function, and handles any external targets generated during the process.
	// Parameters:
	//   - ctx: Context for execution
	//   - machineContext: Machine context to be stored and used throughout the machine's lifecycle
	// Returns error if any initial universe reference is invalid or initialization fails
	Init(ctx context.Context, machineContext any) error

	// InitWithEvent initializes the quantum machine with a custom event.
	// Similar to Init, but allows passing a custom event that will be propagated through
	// the initialization process. This is useful when you need to initialize with specific
	// metadata, flags, or data in the event context.
	// Parameters:
	//   - ctx: Context for execution
	//   - machineContext: Machine context to be stored and used throughout the machine's lifecycle
	//   - event: Custom event to propagate during initialization
	// Returns error if any initial universe reference is invalid or initialization fails
	InitWithEvent(ctx context.Context, machineContext any, event Event) error

	// SendEvent sends an event to all active universes that can handle it.
	// The method determines which universes are active and can process the event based on
	// their current state and available transitions. Any external targets generated during
	// event processing are handled automatically.
	// Parameters:
	//   - ctx: Context for execution
	//   - event: Event to send to active universes
	// Returns:
	//   - bool: true if at least one universe processed the event, false if no universes were active or handled it
	//   - error: error if event processing fails
	SendEvent(ctx context.Context, event Event) (bool, error)

	// LoadSnapshot restores the quantum machine state from a snapshot.
	// For each universe in the machine, it loads the corresponding universe snapshot which includes
	// the current reality, superposition state, tracking history, and other universe-specific state.
	// The provided machineContext replaces the current machine context.
	// Parameters:
	//   - snapshot: Snapshot containing the state to restore (nil snapshot is safely ignored)
	//   - machineContext: Machine context to set after loading the snapshot
	// Returns error if any universe fails to load its snapshot
	LoadSnapshot(snapshot *MachineSnapshot, machineContext any) error

	// GetSnapshot captures the current complete state of the quantum machine.
	// The snapshot includes for each universe: current reality, superposition state, tracking history,
	// and categorization into active universes (running), finalized universes (in final state),
	// or superposition universes (awaiting collapse).
	// Returns:
	//   - MachineSnapshot: Complete snapshot of the machine's current state
	GetSnapshot() *MachineSnapshot

	// ReplayOnEntry re-executes entry actions for the current realities of all active universes.
	// Creates a special event with ReplayOnEntry flag set to true, then processes it through all
	// active universes (not finalized or in superposition). This is useful for re-applying entry
	// logic without changing the current state, such as after a code reload or configuration change.
	// Parameters:
	//   - ctx: Context for execution
	// Returns error if entry action replay fails for any universe
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

	// PositionMachineOnInitial positions the quantum machine on the initial state of the specified universe.
	// This is a convenience method that automatically uses the universe's configured initial state.
	// Parameters:
	//   - ctx: Context for execution
	//   - machineContext: Machine context to use
	//   - universeID: Target universe identifier
	//   - executeFlow: If true, executes full entry flow. If false, only positions without executing actions.
	// Returns error if universe doesn't exist, has no initial state, or positioning fails
	PositionMachineOnInitial(ctx context.Context, machineContext any, universeID string, executeFlow bool) error

	// PositionMachineByCanonicalName positions the quantum machine using universe's canonical name.
	// This is a convenience method that resolves the universe by its canonical name.
	// Parameters:
	//   - ctx: Context for execution
	//   - machineContext: Machine context to use
	//   - universeCanonicalName: Target universe canonical name
	//   - realityID: Target reality (state) identifier
	//   - executeFlow: If true, executes full entry flow. If false, only positions without executing actions.
	// Returns error if universe/reality doesn't exist or positioning fails
	PositionMachineByCanonicalName(ctx context.Context, machineContext any, universeCanonicalName string, realityID string, executeFlow bool) error

	// PositionMachineOnInitialByCanonicalName positions the machine on initial state using canonical name.
	// This is a convenience method that resolves the universe by its canonical name and uses its initial state.
	// Parameters:
	//   - ctx: Context for execution
	//   - machineContext: Machine context to use
	//   - universeCanonicalName: Target universe canonical name
	//   - executeFlow: If true, executes full entry flow. If false, only positions without executing actions.
	// Returns error if universe doesn't exist, has no initial state, or positioning fails
	PositionMachineOnInitialByCanonicalName(ctx context.Context, machineContext any, universeCanonicalName string, executeFlow bool) error
}
