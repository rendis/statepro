# Debugging & Tooling

StatePro ships with utilities to inspect and exercise machines without writing extra scaffolding.

## Bubble Tea CLI Debugger

Located under `debugger/cli`, this TUI loads definitions, events, and snapshots for interactive
exploration.

### Running the Sample

```bash
go run ./example/cli
```

By default the example loads files from the repository root. Use the setters on
`cli.StateMachineDebugger` to point at custom paths:

```go
debugger := cli.NewStateMachineDebugger()
debugger.SetStateMachinePath("./state_machine.json")
debugger.SetEventsPath("./events.json")
debugger.SetSnapshotsPath("./snapshots.json")
debugger.SetSMContextPath("./context.json")
debugger.Run(nil)
```

### Features

- View machine metadata, universes, and realities as formatted JSON (with color support).
- Send events from the loaded list and observe resulting snapshots in real time.
- Inspect history timelines, tracking information, and serialized accumulator state.
- Load machine context fixtures (`context.json`) to simulate stateful execution.

Tips:

- Provide `events.json` as a list of objects with `name`, optional `data`, and optional flags.
- `snapshots.json` can preload earlier runs—use `qm.GetSnapshot()` and `ToJson()` to generate it.

## Automation Bot

The bot under `debugger/bot` automates event playback for repeatable testing.

### Minimal Usage

```go
qm := loadDefinition() // see modeling guide
provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
    if len(events) == 0 {
        return nil, nil
    }
    evt := events[0]
    events = events[1:]
    return evt, nil
}

b, _ := bot.NewBot(qm, provider, true)
_ = b.Run(context.Background(), myContext)
for _, step := range b.GetHistory() {
    fmt.Println(step.Event.GetEventName(), step.Snapshot.GetResume())
}
```

Parameters:

- `qm`: any `instrumentation.QuantumMachine` (experimental runtime or custom implementation).
- `eventProvider`: called after each step, receives the latest snapshot, and returns the next event
  (or `nil` to stop).
- `initQuantumMachine`: when `true`, `Run` calls `qm.Init` before sending events.

### Use Cases

- Deterministic regression tests for complex flows.
- Batch simulations that assert snapshot contents after each event.
- Generating documentation assets by exporting tracking histories.

## Logging

Use `builtin.SetLogger(abslog.New(...))` or register your own logger to capture runtime diagnostics.
Abslog integrates with structured logging backends and honors context cancellation.

## Next Steps

- Review [runtime.md](runtime.md) to understand how events propagate while using these tools.
- Combine the bot with Go testing frameworks to cover edge cases involving superposition and observers.
