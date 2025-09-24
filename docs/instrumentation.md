# Instrumentation Reference

The `instrumentation` package defines the contracts that runtimes, executors, and tools use to work
with StatePro. This document summarizes every public type.

## QuantumMachine Interface

```go
type QuantumMachine interface {
    Init(ctx context.Context, machineContext any) error
    InitWithEvent(ctx context.Context, machineContext any, event Event) error
    SendEvent(ctx context.Context, event Event) (bool, error)
    LoadSnapshot(snapshot *MachineSnapshot, machineContext any) error
    GetSnapshot() *MachineSnapshot
    ReplayOnEntry(ctx context.Context) error
}
```

- `machineContext` is an arbitrary pointer you manage. Executors can inspect/mutate it via
  `GetContext()` on argument objects.
- `SendEvent` returns `false` when no active universe accepted the event.

## Events

```go
type Event interface {
    GetEventName() string
    GetData() map[string]any
    DataContainsKey(key string) bool
    GetEvtType() EventType
    GetFlags() EventFlags
}
```

- Use `statepro.NewEventBuilder(name)` to create events.
- `EventType` values: `Start`, `StartOn`, `On`, `OnEntry`.
- `EventFlags` currently exposes `ReplayOnEntry` to indicate replays.

### EventBuilder

```go
type EventBuilder interface {
    SetData(map[string]any) EventBuilder
    SetEvtType(EventType) EventBuilder
    SetFlags(EventFlags) EventBuilder
    Build() Event
}
```

## Executor Argument Struct

```go
type QuantumMachineExecutorArgs struct {
    Context               any
    RealityName           string
    UniverseID            string
    UniverseCanonicalName string
    Event                 Event
    AccumulatorStatistics AccumulatorStatistics
}
```

- Passed to universal constants so they can interact with the active universe.

## Observer Contracts

```go
type ObserverExecutorArgs interface {
    GetContext() any
    GetRealityName() string
    GetUniverseCanonicalName() string
    GetUniverseId() string
    GetAccumulatorStatistics() AccumulatorStatistics
    GetEvent() Event
    GetObserver() theoretical.ObserverModel
    GetUniverseMetadata() map[string]any
    AddToUniverseMetadata(key string, value any)
    DeleteFromUniverseMetadata(key string) (any, bool)
    UpdateUniverseMetadata(md map[string]any)
}

type ObserverFn func(ctx context.Context, args ObserverExecutorArgs) (bool, error)
```

Observers return `true` to authorize a transition. Use accumulator statistics to evaluate historical
patterns.

## Action Contracts

```go
type ActionType string
const (
    ActionTypeEntry      ActionType = "entry"
    ActionTypeExit       ActionType = "exit"
    ActionTypeTransition ActionType = "transition"
)

type ActionExecutorArgs interface {
    GetContext() any
    GetRealityName() string
    GetUniverseCanonicalName() string
    GetUniverseId() string
    GetEvent() Event
    GetAction() theoretical.ActionModel
    GetActionType() ActionType
    GetSnapshot() *MachineSnapshot
    GetUniverseMetadata() map[string]any
    AddToUniverseMetadata(key string, value any)
    DeleteFromUniverseMetadata(key string) (any, bool)
    UpdateUniverseMetadata(md map[string]any)
}

type ActionFn func(ctx context.Context, args ActionExecutorArgs) error
```

Return an error to abort the transition (or initialization) that triggered the action.

## Invoke Contracts

```go
type InvokeExecutorArgs interface {
    GetContext() any
    GetRealityName() string
    GetUniverseCanonicalName() string
    GetUniverseId() string
    GetEvent() Event
    GetInvoke() theoretical.InvokeModel
    GetUniverseMetadata() map[string]any
    AddToUniverseMetadata(key string, value any)
    DeleteFromUniverseMetadata(key string) (any, bool)
    UpdateUniverseMetadata(md map[string]any)
}

type InvokeFn func(ctx context.Context, args InvokeExecutorArgs)
```

Invokes never return errors; they execute asynchronously and should manage their own logging.

## Condition Contracts

```go
type ConditionExecutorArgs interface {
    GetContext() any
    GetRealityName() string
    GetUniverseCanonicalName() string
    GetUniverseId() string
    GetEvent() Event
    GetCondition() theoretical.ConditionModel
    GetUniverseMetadata() map[string]any
    AddToUniverseMetadata(key string, value any)
    DeleteFromUniverseMetadata(key string) (any, bool)
    UpdateUniverseMetadata(md map[string]any)
}

type ConditionFn func(ctx context.Context, args ConditionExecutorArgs) (bool, error)
```

Returning `true` allows the transition to proceed. Return `false` to veto without raising an error.

## ConstantsLawsExecutor

```go
type ConstantsLawsExecutor interface {
    ExecuteEntryInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
    ExecuteExitInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
    ExecuteEntryAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
    ExecuteExitAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
    ExecuteTransitionInvokes(ctx context.Context, args *QuantumMachineExecutorArgs)
    ExecuteTransitionAction(ctx context.Context, args *QuantumMachineExecutorArgs) error
    GetSnapshot() *MachineSnapshot
}
```

The experimental runtime implements this interface internally. Custom runtimes or adapters can supply
alternative implementations.

## Accumulators

```go
type Accumulator interface {
    Accumulate(realityName string, event Event)
    GetStatistics() AccumulatorStatistics
    GetActiveRealities() []string
}

type AccumulatorStatistics interface {
    GetRealitiesEvents() map[string][]Event
    GetRealityEvents(realityName string) map[string]Event
    GetAllRealityEvents(realityName string) map[string][]Event
    GetAllEventsNames() []string
    CountAllEventsNames() int
    CountAllEvents() int
}
```

Use custom accumulators to tailor how events are stored or expose additional analytics to observers.

## Snapshots

```go
type SerializedUniverseSnapshot map[string]any

type MachineSnapshot struct {
    Resume   UniversesResume
    Snapshots map[string]SerializedUniverseSnapshot
    Tracking map[string][]string
}

type UniversesResume struct {
    ActiveUniverses                   map[string]string
    FinalizedUniverses                map[string]string
    SuperpositionUniverses            map[string]string
    SuperpositionUniversesFinalized   map[string]string
}
```

### Helper Methods

- `AddActiveUniverse`, `AddFinalizedUniverse`, `AddSuperpositionUniverse`,
  `AddSuperpositionUniverseFinalized`
- `AddUniverseSnapshot`, `AddTracking`
- `GetResume`, `GetActiveUniverses`, `GetFinalizedUniverses`, `GetSuperpositionUniverses`,
  `GetTracking`
- `ToJson()` — human-readable JSON string (for logging or persistence)

Snapshots are safe to serialize and reload using `QuantumMachine.LoadSnapshot`.

## Putting the Interfaces to Work

Register a custom observer:

```go
func init() {
    _ = builtin.RegisterObserver("custom:observer:requiresCounter", requireCounter)
}

func requireCounter(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    stats := args.GetAccumulatorStatistics()
    if stats == nil {
        return false, nil
    }
    events := stats.GetAllRealityEvents(args.GetRealityName())
    return len(events["confirm"]) >= 2, nil
}
```

When writing your own runtime, use these interfaces as boundaries so existing actions/observers remain
compatible. Refer to [runtime.md](runtime.md) for guidance on orchestration details.
