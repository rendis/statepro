# Runtime & Execution

The `experimental` package provides the reference implementation of `instrumentation.QuantumMachine`.
This document explains how it interprets a model and processes events.

## Initialization

- Call `qm.Init(ctx, machineContext)` to boot every universe referenced in `QuantumMachineModel.Initials`.
- Use `qm.InitWithEvent(ctx, machineContext, event)` to inject a custom event during initialization.
- A machine can only be initialized once. Calling `Init()` or `InitWithEvent()` again returns an error. To restore a previously initialized machine, use `LoadSnapshot()` instead.
- Each reference in `initials` is validated and mapped to an `ExUniverse` instance.
- Universes without an `initial` reality enter superposition. Universes with an initial reality execute:
  1. Machine-level entry constants.
  2. Universe-level entry constants.
  3. Reality `always` transitions (in order).
  4. Reality entry actions/invokes.

### Manual Positioning

Instead of using standard initialization, you can manually position the machine in a specific state:

- **`PositionMachine(ctx, machineContext, universeID, realityID, executeFlow)`** - Position in a specific universe and reality by ID.
- **`PositionMachineOnInitial(ctx, machineContext, universeID, executeFlow)`** - Position in a universe's configured initial state.
- **`PositionMachineByCanonicalName(ctx, machineContext, universeCanonicalName, realityID, executeFlow)`** - Position using the universe's canonical name instead of ID.
- **`PositionMachineOnInitialByCanonicalName(ctx, machineContext, universeCanonicalName, executeFlow)`** - Combine canonical name lookup with initial state positioning.

The `executeFlow` parameter controls whether to run the full entry flow (actions, transitions) or simply position without execution. This is useful for:

- Restoring machines to a known state for testing
- Skipping initialization logic when loading from persistent storage
- Debugging specific states without executing the full workflow

## Event Routing

1. `SendEvent` locks the machine and finds universes that can handle the event.
2. For each active universe:
   - Superposition universes accumulate the event per reality.
   - Non-superposition universes forward the event to the current reality’s transition list (`on`).
3. Transitions that target other universes are queued as “external targets” so the machine can fan out
   events to multiple destinations.

If no universe handles the event, `SendEvent` returns `(false, nil)`.

## Superposition Lifecycle

- A universe is in superposition when `currentReality` is `nil`.
- Events are stored in an accumulator until a transition is ready to fire.
- Observers and conditions consult `AccumulatorStatistics` to decide whether to collapse into a
  concrete reality.
- When a new reality becomes active, `establishNewReality` runs entry logic, updates tracking, and
  may emit further transitions.

Tip: design observers/conditions to eventually return `true`; otherwise the universe will remain in
superposition indefinitely.

## Actions & Invokes

- **Actions** run synchronously. Any error stops the transition and the machine remains in the previous
  state.
- **Invokes** run asynchronously on separate goroutines. Their function does not return an error
  to the transition. With configured runtime limits, an admission failure does return an error
  before the rejected invoke is started.
- Both receive `instrumentation` executor arguments including the machine context, universe metadata,
  event payload, and snapshot accessors.

### EmitEvent — Internal Event Emission

Entry actions can emit internal events via `args.EmitEvent(eventName, data)`. After **all** entry actions complete, emitted events are processed against the current reality's `On` handlers. If a transition is approved (conditions pass), the machine advances automatically — no external `SendEvent` needed.

This is similar to XState's `raise`: an internal event processed within the same operation.

**Execution flow:**

1. Reality entry actions run (synchronously, in order).
2. Emitted events are collected in a FIFO queue.
3. After all entry actions + invokes complete, each emitted event is matched against `On` handlers.
4. The first event that triggers an approved transition wins. Remaining events are discarded.
5. If the new reality also has entry actions that emit, the process chains recursively (max depth: 10).

**Where EmitEvent works:**

| Context | EmitEvent | Reason |
|---------|-----------|--------|
| Entry actions | Yes | Primary use case |
| Constants entry actions | Yes | Same semantics as reality entry |
| Exit actions | No-op + warning | Reality is being left, no `On` handlers apply |
| Transition actions | No-op + warning | Target already determined |

**Example:**

```go
builtin.RegisterAction("action:createForm", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    templateId := args.GetAction().Args["templateId"].(string)
    // ... create the form ...

    // Emit event to auto-advance. The existing On handler for "create-form"
    // will evaluate its conditions and transition if approved.
    args.EmitEvent("create-form", map[string]any{"templateId": templateId})
    return nil
})
```

Zero changes to the JSON definition. The existing `on.create-form` transition with its conditions does the rest.

**Error handling:** Chained emits (A emits -> B -> B emits -> C -> ...) are capped at depth 10. Exceeding this limit returns an error and leaves the machine instance in an unrecoverable state. This always indicates a bug in the state machine definition (infinite loop). The caller should discard the machine instance.

**Backward compatibility:** If no action calls `EmitEvent`, behavior is identical to before. Zero overhead when unused.

### Universal Constants Ordering

1. Machine-level entry/exit invocations and actions.
2. Universe-level entry/exit invocations and actions.
3. Reality-specific entry/exit logic.
4. Transition-level actions/invokes (after step 2, before the new reality executes its entry logic).

## Conditions & Observers

- Observers run sequentially; the first success wins. Errors are propagated unless another observer has
  already authorized the transition.
- `TransitionModel.condition` and `conditions` arrays are evaluated sequentially. All must return `true`
  for the transition to proceed.

## Snapshots

`qm.GetSnapshot()` returns an `instrumentation.MachineSnapshot` containing:

- `Resume`: active, finalized, and superposition universes grouped by canonical name.
- `Snapshots`: serialized per-universe state (including accumulators and metadata).
- `Tracking`: ordered history of realities visited per universe.

Use `qm.LoadSnapshot(snapshot, machineContext)` to restore a machine. Snapshots capture the latest
machine context metadata but you must provide any external context objects when reloading.

For persistence, use the optional checked capture capability:

```go
snapshot, err := instrumentation.GetSnapshotWithError(qm)
if err != nil {
    return err
}
encoded, err := snapshot.ToJson()
```

The experimental runtime returns a complete snapshot or an error. Unsupported JSON values in
metadata or event data (such as channels, functions, cycles, or non-finite floats) fail capture.
The legacy `GetSnapshot()` method logs the failure and returns `nil`; check that result before use.
Existing third-party implementations do not need to add a method to the machine interface. The
checked helper returns an unsupported-capability error when a provider lacks `SnapshotProvider`.

Snapshot capture, `json.Unmarshal` into `MachineSnapshot`, and `LoadSnapshot` preserve numeric
values through JSON round-trips. Ordinary values retain the previous `float64` representation;
integers and decimals whose value would change through that conversion retain `json.Number`.
Consumers of untyped metadata must support both representations. This includes integers above
the exact float64 range and long decimal fractions. Captured maps remain detached from live data.

Inside a synchronous action, call `instrumentation.GetSnapshotWithError(args)` using the callback
arguments. Calling the owning machine directly would try to reacquire its execution lock.

`qm.ReplayOnEntry(ctx)` re-executes entry actions and invokes for every active universe without
changing reality assignments. This is useful when you need to re-run side effects after downtime.

### Snapshot Management

**Loading Snapshots:**

Use `qm.LoadSnapshot(snapshot, machineContext)` to restore a machine from a previous snapshot. This replaces:

- All universe states (current realities, superposition states)
- Event accumulators and their statistics
- Tracking history for all universes
- The machine context (provided as parameter)

Snapshots are useful for:

- Persisting state to disk or database
- Recovering from crashes or restarts
- Creating checkpoints for rollback scenarios
- Testing with pre-configured states

## Error Handling

- Most runtime errors originate from actions, invokes, observers, or invalid transitions.
- When an action fails during a transition, the machine rolls back to the previous reality.
- Errors bubble up to the caller of `Init`/`SendEvent`/`ReplayOnEntry`. Handle them at the application
  level (retry, alert, compensating transaction, etc.).

## Extending the Runtime

The experimental runtime implements all instrumentation interfaces. You can build your own runtime by:

1. Implementing `instrumentation.QuantumMachine` (possibly reusing `theoretical` models).
2. Providing your own accumulator or metadata strategy.
3. Re-registering actions/observers/invokes via the `builtin` package or custom registries.

Consult [instrumentation.md](instrumentation.md) for the list of contracts you must satisfy.

## Runtime policies and resource limits

Existing constructors preserve unlimited numeric budgets, permissive unknown observers, and invokes
that continue after their reality exits. Configure policies explicitly when those defaults do not
fit an application's workload:

```go
qm, err := statepro.NewQuantumMachineWithOptions(model, instrumentation.RuntimeOptions{
    MaxConcurrentInvokes: 8,
    CancelInvokesOnExit:  true,
    MaxAccumulatedEvents: 1_000,
    MaxTrackingEntries:   100,
    StrictObservers:      true,
})
if err != nil {
    return err
}
```

These are example budgets, not recommended values for every application. Options are copied at
construction; negative numeric limits are rejected. The experimental constructor also exposes
`NewExQuantumMachineWithOptions`. The mandatory machine and executor interfaces are unchanged.

| Policy | Behavior |
| --- | --- |
| `MaxConcurrentInvokes` | One pool for the entire machine, including machine constants, universe constants, and reality/transition invokes. Before a step runs its first callback, reserve room for every invoke it can launch: a transition step counts transition, exit, and target entry invokes; an entry counts its entry invokes. When they do not fit, reject the step with `*instrumentation.ResourceLimitError` before any action runs or any invoke starts; nothing is enqueued. A later step of the same cascade (an `always` transition) is checked again from the stable reality it starts in. Slots release on completion or panic, not merely on a cancellation request. |
| `CancelInvokesOnExit` | Request cancellation of existing invokes associated with a successfully exited reality, before launching exit invokes. Also cancel affected universes on valid snapshot replacement or static positioning. Rejected snapshot restoration does not cancel existing tasks. |
| `MaxAccumulatedEvents` | Per-universe entry budget, counting separate copies for different realities. Reserve worst-case fan-out before callbacks or appending entries, even if an early observer might approve. Reject over-budget snapshots before applying them. This is an entry limit, not a byte limit. |
| `MaxTrackingEntries` | Retain the most recent entries per universe, including restored tracking. Failed entry rollback preserves the previous bounded history. |
| `StrictObservers` | Reject empty or unregistered observer sources with `ErrUnknownObserver` before accumulating the event. Every configured observer in a fan-out must be registered. The default still permits unknown sources. |

The debugger bot independently supports `bot.WithHistoryLimit(n)`, retaining the latest event
snapshots. Zero means unlimited; negative values are rejected. Bot runs check cancellation before
restoration, between events, and after the event provider returns, and propagate checked snapshot
errors. The provider's existing signature has no context argument; a provider that blocks must
arrange its own cancellation. Serialize `Run` and history access at the application level.

### Cancellation, reentry, and shutdown

Operations with a context can cancel while waiting for the machine lock, and are rejected if the
context is already done once the lock is acquired. After that admission point, the operation runs
to completion: cancellation does not abort a transition between callbacks, because earlier actions
may already have produced external side effects and stopping halfway would leave the universe in an
intermediate state. Callbacks receive the context and may observe cancellation themselves; a
callback that returns an error (including `ctx.Err()`) fails the operation like any other callback
error. Invokes launched after cancellation are skipped, since they would only see a done context.
Go cannot safely interrupt arbitrary user code or undo its external side effects, so admission
checks and cancellation do not make execution transactional across callback errors.

Calling the owning machine from a synchronous callback using the callback context (or a derived
context) returns `ErrReentrantCall`. This includes nested synchronous callbacks across machines.
Asynchronous invokes receive a context without that synchronous scope marker, preserving caller
values and cancellation; they can enqueue work on the machine normally. Reusing a callback context
after its callback has completed is supported.

For capture inside an action, use `instrumentation.GetSnapshotWithError(args)` or `args.GetSnapshot()`.
The optional `instrumentation.ContextSnapshotProvider` adds `GetSnapshotContext(ctx)` and
`LoadSnapshotContext(ctx, snapshot, machineContext)`, with cancellable lock waiting and reentry checks.
The legacy context-free machine snapshot methods cannot identify a callback's caller and can still
deadlock if called directly inside its owning synchronous callback. Replacing the callback context
with `context.Background()` also bypasses reentry detection. Preserve the supplied context.

The optional `instrumentation.RuntimeLifecycle` provides:

- `Close()`: idempotently request cancellation of all invokes and reject new execution with
  `ErrMachineClosed`. It returns without waiting for user code to stop. Snapshot reads remain available.
- `WaitInvokes(ctx)`: wait for the pool to become idle or return the context error. Call `Close()` first
  when a stable shutdown barrier is required; concurrent new admissions can otherwise start later.

Always give shutdown waiting a deadline. An invoke that ignores cancellation occupies its slot
until it exits and may outlive `Close()`.
