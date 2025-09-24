# Runtime & Execution

The `experimental` package provides the reference implementation of `instrumentation.QuantumMachine`.
This document explains how it interprets a model and processes events.

## Initialization

- Call `qm.Init(ctx, machineContext)` to boot every universe referenced in `QuantumMachineModel.Initials`.
- Use `qm.InitWithEvent(ctx, machineContext, event)` to inject a custom event during initialization.
- Each reference in `initials` is validated and mapped to an `ExUniverse` instance.
- Universes without an `initial` reality enter superposition. Universes with an initial reality execute:
  1. Machine-level entry constants.
  2. Universe-level entry constants.
  3. Reality `always` transitions (in order).
  4. Reality entry actions/invokes.

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
- **Invokes** run asynchronously on separate goroutines. They are "fire-and-forget" and do not affect
  control flow.
- Both receive `instrumentation` executor arguments including the machine context, universe metadata,
  event payload, and snapshot accessors.

### Universal Constants Ordering

1. Machine-level entry/exit invocations and actions.
2. Universe-level entry/exit invocations and actions.
3. Reality-specific entry/exit logic.
4. Transition-level actions/invokes (after step 2, before the new reality executes its entry logic).

## Conditions & Observers

- Observers run in parallel; the first success wins. Errors are propagated unless another observer has
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

`qm.ReplayOnEntry(ctx)` re-executes entry actions and invokes for every active universe without
changing reality assignments. This is useful when you need to re-run side effects after downtime.

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
