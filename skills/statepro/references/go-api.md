# Go API Reference

## Table of Contents

1. [Package Imports](#package-imports)
2. [Creating a Machine](#creating-a-machine)
3. [Serialization](#serialization)
4. [QuantumMachine Interface](#quantummachine-interface)
5. [Event System](#event-system)
6. [Accumulator](#accumulator)
7. [MachineSnapshot](#machinesnapshot)
8. [Executor Registration](#executor-registration)
9. [Executor Args Interfaces](#executor-args-interfaces)
10. [EmitEvent Mechanics](#emitevent-mechanics)
11. [Complete Usage Example](#complete-usage-example)

## Package Imports

```go
import (
    "github.com/rendis/statepro/v3"                   // NewQuantumMachine, NewEventBuilder, serde
    "github.com/rendis/statepro/v3/builtin"            // RegisterObserver/Action/Invoke/Condition
    "github.com/rendis/statepro/v3/instrumentation"    // Interfaces, Event, MachineSnapshot
    "github.com/rendis/statepro/v3/theoretical"        // Model structs (rarely used directly)
    "github.com/rendis/statepro/v3/debugger/cli"       // TUI debugger
    "github.com/rendis/statepro/v3/debugger/bot"       // Automation bot
)
```

## Creating a Machine

```go
func statepro.NewQuantumMachine(qmModel *theoretical.QuantumMachineModel) (instrumentation.QuantumMachine, error)
func statepro.NewEventBuilder(eventName string) instrumentation.EventBuilder
```

## Serialization

```go
// From JSON bytes
func statepro.DeserializeQuantumMachineFromBinary(b []byte) (*theoretical.QuantumMachineModel, error)

// From map[string]any
func statepro.DeserializeQuantumMachineFromMap(source map[string]any) (*theoretical.QuantumMachineModel, error)

// To JSON bytes
func statepro.SerializeQuantumMachineToBinary(source *theoretical.QuantumMachineModel) ([]byte, error)

// To map[string]any
func statepro.SerializeQuantumMachineToMap(source *theoretical.QuantumMachineModel) (map[string]any, error)
```

## QuantumMachine Interface

```go
type QuantumMachine interface {
    // Initialize machine with initials. machineContext accessible in all executors via GetContext().
    Init(ctx context.Context, machineContext any) error

    // Init with custom event propagated to activated universes.
    InitWithEvent(ctx context.Context, machineContext any, event Event) error

    // Send event to all active universes. Returns (handled bool, error).
    SendEvent(ctx context.Context, event Event) (bool, error)

    // Restore machine state from snapshot.
    LoadSnapshot(snapshot *MachineSnapshot, machineContext any) error

    // Capture complete machine state.
    GetSnapshot() *MachineSnapshot

    // Re-execute entry actions for current realities of active universes.
    ReplayOnEntry(ctx context.Context) error

    // Position universe at specific reality. executeFlow=true runs full entry logic.
    PositionMachine(ctx context.Context, machineContext any, universeID string, realityID string, executeFlow bool) error

    // Position universe at its initial reality.
    PositionMachineOnInitial(ctx context.Context, machineContext any, universeID string, executeFlow bool) error

    // Same as PositionMachine but by canonical name.
    PositionMachineByCanonicalName(ctx context.Context, machineContext any, universeCanonicalName string, realityID string, executeFlow bool) error

    // Same as PositionMachineOnInitial but by canonical name.
    PositionMachineOnInitialByCanonicalName(ctx context.Context, machineContext any, universeCanonicalName string, executeFlow bool) error
}
```

**Notes**:
- `Init()` has double-init guard — errors if called twice without `LoadSnapshot()` reset.
- `SendEvent()` routes to all active (non-finalized, non-superposition) universes.
- `PositionMachine(..., executeFlow=true)` runs full entry flow (actions, invokes, always transitions).
- `PositionMachine(..., executeFlow=false)` static positioning only, no side effects.

## Event System

### EventType

```go
type EventType string
const (
    EventTypeStart   EventType = "Start"    // Universe initial start
    EventTypeStartOn EventType = "StartOn"  // Universe starts on specific reality
    EventTypeOn      EventType = "On"       // From reality.on transitions
    EventTypeOnEntry EventType = "OnEntry"  // Force current reality to re-execute entry
    EventTypeEmitted EventType = "Emitted"  // Emitted internally by entry actions
)
```

### Event Interface

```go
type Event interface {
    GetEventName() string
    GetData() map[string]any
    DataContainsKey(key string) bool
    GetEvtType() EventType
    GetFlags() EventFlags
}

type EventFlags struct {
    ReplayOnEntry bool `json:"replayOnEntry"`
}
```

### EventBuilder

```go
type EventBuilder interface {
    SetData(data map[string]any) EventBuilder
    SetEvtType(evtType EventType) EventBuilder
    SetFlags(flags EventFlags) EventBuilder
    Build() Event
}
```

Usage:
```go
event := statepro.NewEventBuilder("confirm").
    SetData(map[string]any{"userId": "123"}).
    Build()
```

## Accumulator

Active during superposition. Collects events per reality.

```go
type Accumulator interface {
    Accumulate(realityName string, event Event)
    GetStatistics() AccumulatorStatistics
    GetActiveRealities() []string
}

type AccumulatorStatistics interface {
    // All accumulated events grouped by reality (with repetitions)
    GetRealitiesEvents() map[string][]Event

    // Events for a reality, deduplicated by name (latest occurrence)
    GetRealityEvents(realityName string) map[string]Event

    // Events for a reality grouped by name (with repetitions)
    GetAllRealityEvents(realityName string) map[string][]Event

    // All unique event names across all realities
    GetAllEventsNames() []string

    // Count unique event names (without repetitions)
    CountAllEventsNames() int

    // Count total events (with repetitions)
    CountAllEvents() int
}
```

## MachineSnapshot

```go
type MachineSnapshot struct {
    Resume    UniversesResume                          `json:"resume"`
    Snapshots map[string]SerializedUniverseSnapshot    `json:"snapshots,omitempty"`
    Tracking  map[string][]string                      `json:"tracking,omitempty"`
}

type UniversesResume struct {
    ActiveUniverses                map[string]string `json:"activeUniverses,omitempty"`
    FinalizedUniverses             map[string]string `json:"finalizedUniverses,omitempty"`
    SuperpositionUniverses         map[string]string `json:"superpositionUniverses,omitempty"`
    SuperpositionUniversesFinalized map[string]string `json:"superpositionUniversesFinalized,omitempty"`
}
```

**Resume maps**: canonicalName → current reality name.

**Tracking**: universeId → ordered list of visited reality names.

Helper methods: `AddActiveUniverse()`, `AddFinalizedUniverse()`, `AddSuperpositionUniverse()`, `AddUniverseSnapshot()`, `AddTracking()`, `ToJson()`.

## Executor Registration

```go
// Src pattern: ^[a-zA-Z][a-zA-Z0-9_:.-]*[a-zA-Z0-9]$
func builtin.RegisterObserver(src string, fn instrumentation.ObserverFn) error
func builtin.RegisterAction(src string, fn instrumentation.ActionFn) error
func builtin.RegisterInvoke(src string, fn instrumentation.InvokeFn) error
func builtin.RegisterCondition(src string, fn instrumentation.ConditionFn) error

// Retrieve (falls back to built-in if custom not found)
func builtin.GetObserver(src string) instrumentation.ObserverFn
func builtin.GetAction(src string) instrumentation.ActionFn
func builtin.GetInvoke(src string) instrumentation.InvokeFn
func builtin.GetCondition(src string) instrumentation.ConditionFn
```

**Registration must happen before `NewQuantumMachine()`** — the machine resolves executors at creation time.

## Executor Args Interfaces

### Common Methods (all executor args)

```go
GetContext() any                          // machineContext passed to Init()
GetRealityName() string
GetUniverseCanonicalName() string
GetUniverseId() string
GetEvent() Event
GetUniverseMetadata() map[string]any     // mutable, persists across transitions
AddToUniverseMetadata(key string, value any)
DeleteFromUniverseMetadata(key string) (any, bool)
UpdateUniverseMetadata(md map[string]any) // replaces all metadata
```

### ObserverExecutorArgs (additional)

```go
GetAccumulatorStatistics() AccumulatorStatistics
GetObserver() theoretical.ObserverModel
```

### ActionExecutorArgs (additional)

```go
GetAction() theoretical.ActionModel
GetActionType() ActionType        // "entry" | "exit" | "transition"
GetSnapshot() *MachineSnapshot
EmitEvent(eventName string, data map[string]any)  // entry actions only
```

### InvokeExecutorArgs (additional)

```go
GetInvoke() theoretical.InvokeModel
```

### ConditionExecutorArgs (additional)

```go
GetCondition() theoretical.ConditionModel
```

## EmitEvent Mechanics

Only available in **entry actions** (ActionType = "entry"). Calling from transition/exit actions is a no-op with warning.

1. Entry action calls `args.EmitEvent("eventName", data)`
2. Event queued in FIFO order
3. After ALL entry actions complete, emitted events process sequentially
4. Each emitted event has `EvtType = EventTypeEmitted`
5. First event that triggers an approved transition wins
6. Remaining queued events discarded after a transition fires
7. Max nesting depth: **10** (prevents infinite A→B→A loops)

## Complete Usage Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/rendis/statepro/v3"
    "github.com/rendis/statepro/v3/builtin"
    "github.com/rendis/statepro/v3/instrumentation"
)

type AppContext struct {
    UserID string
    Active bool
}

func main() {
    // 1. Register executors
    builtin.RegisterAction("disableAdmission", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
        appCtx := args.GetContext().(*AppContext)
        appCtx.Active = false
        return nil
    })

    builtin.RegisterCondition("checkTemplate", func(ctx context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
        expected := args.GetCondition().Args["templateId"]
        actual := args.GetEvent().GetData()["templateId"]
        return expected == actual, nil
    })

    // 2. Load definition
    jsonBytes, _ := os.ReadFile("state_machine.json")
    model, _ := statepro.DeserializeQuantumMachineFromBinary(jsonBytes)

    // 3. Create machine
    qm, _ := statepro.NewQuantumMachine(model)

    // 4. Initialize
    ctx := context.Background()
    appCtx := &AppContext{UserID: "user-1", Active: true}
    qm.Init(ctx, appCtx)

    // 5. Send events
    event := statepro.NewEventBuilder("confirm").Build()
    handled, _ := qm.SendEvent(ctx, event)
    fmt.Printf("Event handled: %v\n", handled)

    // 6. Snapshot
    snapshot := qm.GetSnapshot()
    jsonStr, _ := snapshot.ToJson()
    fmt.Println(jsonStr)

    // 7. Load snapshot (restore state)
    qm.LoadSnapshot(snapshot, appCtx)
}
```
