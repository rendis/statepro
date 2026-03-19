---
name: statepro
description: "Quantum state machine library (Go) with visual studio editor. Use when: working with statepro imports (github.com/rendis/statepro), writing or editing state machine JSON definitions, creating universes/realities/transitions, registering observers/actions/invokes/conditions, handling superposition/collapse mechanics, using the debugger CLI/TUI or automation bot, integrating the studio visual editor (@rendis/statepro-studio-react or statepro-studio web component), validating machine definitions against JSON schema, managing machine snapshots, or any task involving quantum machine models, event routing, accumulator patterns, or the statepro ecosystem."
---

# StatePro — Quantum State Machine Library

Go library for multi-universe state machines with superposition. Three layers: `theoretical/` (models), `instrumentation/` (interfaces), `experimental/` (runtime). Visual editor via `studio/`.

## Terminology

| Quantum Metaphor | Traditional FSM | Description |
|---|---|---|
| **QuantumMachine** | State Machine | Top-level container, multiple universes |
| **Universe** | Independent FSM | Self-contained state machine with own current state |
| **Reality** | State | A state within a universe |
| **Transition** | Edge | Movement between realities, triggered by events |
| **Superposition** | — | Universe has no concrete reality; accumulates events until observer collapses |
| **Observer** | — | Guard that watches accumulated events, triggers collapse |
| **Action** | Side Effect (sync) | Synchronous operation on entry/exit/transition |
| **Invoke** | Side Effect (async) | Fire-and-forget async operation |
| **Condition** | Guard | Boolean check on a transition |
| **Universal Constants** | Global Hooks | Actions/invokes that execute on ALL entries/exits/transitions |
| **Accumulator** | — | Collects events during superposition for observer evaluation |

## Machine Definition Quick Reference

Minimal valid machine (one universe, two realities):

```json
{
  "id": "my-machine",
  "canonicalName": "my-machine",
  "version": "1.0.0",
  "initials": ["U:my-universe"],
  "universes": {
    "my-universe": {
      "id": "my-universe",
      "canonicalName": "my-universe",
      "version": "1.0.0",
      "initial": "IDLE",
      "realities": {
        "IDLE": {
          "id": "IDLE",
          "type": "transition",
          "on": {
            "start": [{ "targets": ["RUNNING"] }]
          }
        },
        "RUNNING": {
          "id": "RUNNING",
          "type": "final"
        }
      }
    }
  }
}
```

**Reality types**: `transition` (intermediate), `final` (successful end), `unsuccessfulFinal` (failed end).

**Transition types**: `default` (abandon current reality), `notify` (send to target WITHOUT leaving current reality).

**Target formats**: `"RealityID"` (internal), `"U:UniverseID"` (external universe), `"U:UniverseID:RealityID"` (external universe at specific reality).

For complete schema, all fields, and validation rules see [references/machine-definition.md](references/machine-definition.md).

## Go API Quick Reference

```go
import (
    "github.com/rendis/statepro/v3"
    "github.com/rendis/statepro/v3/builtin"
    "github.com/rendis/statepro/v3/instrumentation"
)

// 1. Register executors BEFORE creating machine
builtin.RegisterAction("myAction", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    machineCtx := args.GetContext().(*MyContext)
    return nil
})

// 2. Load & create machine
jsonBytes, _ := os.ReadFile("state_machine.json")
model, _ := statepro.DeserializeQuantumMachineFromBinary(jsonBytes)
qm, _ := statepro.NewQuantumMachine(model)

// 3. Initialize with context
ctx := context.Background()
machineCtx := &MyContext{}
qm.Init(ctx, machineCtx)

// 4. Send events
event := statepro.NewEventBuilder("start").
    SetData(map[string]any{"key": "value"}).
    Build()
handled, err := qm.SendEvent(ctx, event)

// 5. Snapshots
snapshot := qm.GetSnapshot()
qm.LoadSnapshot(snapshot, machineCtx)
```

For full API reference, all interfaces, and executor signatures see [references/go-api.md](references/go-api.md).

## Executor Registration

```go
builtin.RegisterObserver(src string, fn instrumentation.ObserverFn) error
builtin.RegisterAction(src string, fn instrumentation.ActionFn) error
builtin.RegisterInvoke(src string, fn instrumentation.InvokeFn) error
builtin.RegisterCondition(src string, fn instrumentation.ConditionFn) error
```

**Src pattern**: `^\S+$` (no whitespace). Convention: `"myObserver"` or `"namespace:name"`.

### Function Signatures

```go
// Observer — returns true to approve superposition collapse
type ObserverFn func(ctx context.Context, args ObserverExecutorArgs) (bool, error)

// Action — synchronous side effect (entry, exit, or transition)
type ActionFn func(ctx context.Context, args ActionExecutorArgs) error

// Invoke — fire-and-forget async (no return value)
type InvokeFn func(ctx context.Context, args InvokeExecutorArgs)

// Condition — guard on transition (must return true for transition to fire)
type ConditionFn func(ctx context.Context, args ConditionExecutorArgs) (bool, error)
```

All executor args provide: `GetContext()`, `GetRealityName()`, `GetUniverseCanonicalName()`, `GetUniverseId()`, `GetEvent()`, `GetUniverseMetadata()`, `AddToUniverseMetadata(key, value)`.

ActionExecutorArgs additionally provides: `GetSnapshot()`, `EmitEvent(name, data)` (entry actions only, FIFO, max depth 10).

## Built-in Behaviors

### Observers

| Src | Args | Behavior |
|---|---|---|
| `builtin:observer:containsAllEvents` | `{"p1":"evt1","p2":"evt2"}` | True if ALL named events accumulated |
| `builtin:observer:containsAtLeastOneEvent` | `{"p1":"evt1","p2":"evt2"}` | True if ANY named event accumulated |
| `builtin:observer:alwaysTrue` | none | Always returns true |
| `builtin:observer:greaterThanEqualCounter` | `{"evt1":3,"evt2":2}` | True if each event appears >= N times |
| `builtin:observer:totalEventsBetweenLimits` | `{"minimum":2,"maximum":5}` | True if total events within range |

### Actions

| Src | Args | Behavior |
|---|---|---|
| `builtin:action:logBasicInfo` | none | Logs actionType, reality, universe |
| `builtin:action:logArgs` | any map | Logs basic info + arg keys and values |
| `builtin:action:logArgsWithoutKeys` | any map | Logs basic info + arg values only |
| `builtin:action:logJustArgsValues` | any map | Logs only arg values |

## Superposition & Collapse

1. Universe enters superposition when targeted externally or with multiple targets
2. `currentReality` = nil; events accumulate in `Accumulator`
3. Observers on each reality evaluate accumulated events
4. First observer returning `true` collapses superposition to that reality
5. Entry actions/invokes execute on the collapsed reality
6. `always` transitions evaluate after entry

**EmitEvent** (entry actions only): queue internal events processed FIFO after all entry actions complete. First event triggering an approved transition wins. Max nesting depth: 10.

## Debugging

### CLI TUI Debugger (Bubble Tea)

```go
debugger := cli.NewStateMachineDebugger()
debugger.SetStateMachinePath("./state_machine.json")
debugger.SetEventsPath("./events.json")       // [{title, name, params?}]
debugger.SetSnapshotsPath("./snapshots.json")
debugger.Run(machineContext)
```

**Views**: Send Event (filterable list, sent/unsent status) | History (rollback support) | Load Snapshot | JSON Viewer.

**Keys**: `enter` send, `esc` back, `v` resume, `m` snapshots, `t` tracking, `s` context, `r` rollback.

### Bot Automation

```go
provider := func(snapshot *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
    // return nil, nil to stop
    return nextEvent, nil
}
b, _ := bot.NewBot(qm, provider, true) // true = call Init
b.Run(ctx, machineContext)
history := b.GetHistory() // []EventHistory{Event, Snapshot}
```

For full debugger docs see [references/debugger.md](references/debugger.md).

## Studio Visual Editor

pnpm monorepo at `studio/` with three packages:
- **`studio/app`** — Standalone Vite dev shell (port 5173)
- **`@rendis/statepro-studio-react`** — Core editor React library (npm)
- **`@rendis/statepro-studio-web-component`** — `<statepro-studio>` custom element (npm)

### Dev Commands

```bash
pnpm -C studio install          # Install deps
pnpm -C studio dev              # Dev with source mode (STUDIO_USE_EDITOR_CORE_SRC=true)
pnpm -C studio dev:dist         # Dev with dist mode
pnpm -C studio build            # Build all packages
pnpm -C studio test             # Run tests
pnpm -C studio typecheck        # TypeScript check
pnpm -C studio dev:doctor       # Diagnose dev environment
```

### React Integration

```tsx
import { StateProEditor } from "@rendis/statepro-studio-react";
import "@rendis/statepro-studio-react/styles.css";

<StateProEditor
  value={{ definition: machineJSON, layout: savedLayout }}
  onChange={(payload) => {
    // payload: { machine, layout, issues, canExport, source, at }
  }}
  features={{ json: { import: true, export: true } }}
  locale="en"
/>
```

### Web Component

```ts
import { defineStateProStudioElement } from "@rendis/statepro-studio-web-component";
import "@rendis/statepro-studio-react/styles.css";

defineStateProStudioElement();
const el = document.querySelector("statepro-studio");
el.value = { definition: machineJSON };
el.addEventListener("studio-change", (e) => console.log(e.detail));
```

For full studio docs (props, Tailwind setup, features, i18n) see [references/studio.md](references/studio.md).

## Build & Test

### Go (core library)

```bash
go build ./...                        # Build
go test ./...                         # Test all
go test ./experimental/...            # Test single package
go test -run TestFuncName ./...       # Test single function
```

### Studio

```bash
pnpm -C studio install                # Install
pnpm -C studio dev                    # Dev (source mode)
pnpm -C studio build                  # Build all
pnpm -C studio test                   # Test all
pnpm -C studio/packages/editor-core test   # Test editor-core only
pnpm -C studio/packages/editor-core build  # Build editor-core only
```

Test descriptions are written in **Spanish**.
