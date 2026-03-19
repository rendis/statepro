# Debugger Reference

## Table of Contents

1. [CLI TUI Debugger](#cli-tui-debugger)
2. [File Formats](#file-formats)
3. [TUI Views & Keyboard Shortcuts](#tui-views--keyboard-shortcuts)
4. [Bot Automation](#bot-automation)
5. [CLI Example](#cli-example)
6. [Bot Example](#bot-example)

## CLI TUI Debugger

Interactive terminal debugger built with Bubble Tea (charmbracelet). Send events, view snapshots, inspect history, and rollback.

### Setup

```go
import "github.com/rendis/statepro/v3/debugger/cli"

debugger := cli.NewStateMachineDebugger()
debugger.SetStateMachinePath("./state_machine.json")  // default: ./state_machine.json
debugger.SetEventsPath("./events.json")                // default: ./events.json
debugger.SetSnapshotsPath("./snapshots.json")          // default: ./snapshots.json
debugger.SetSMContextPath("./context.json")            // default: ./context.json
debugger.Run(machineContext)                           // machineContext can be nil
```

All paths are relative to the working directory. Files must exist before launching.

### Required Files

| File | Required | Description |
|---|---|---|
| `state_machine.json` | Yes | QuantumMachineModel JSON definition |
| `events.json` | Yes | Array of sendable events |
| `snapshots.json` | No | Array of loadable snapshots |
| `context.json` | No | Machine context (JSON serialized) |

## File Formats

### events.json

Array of event descriptors:

```json
[
  {
    "title": "confirms admission",
    "name": "confirm"
  },
  {
    "title": "complete data consent survey",
    "name": "fill-form",
    "params": {
      "templateId": "data-consent-admission-survey"
    }
  }
]
```

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | Yes | Display label in TUI |
| `name` | string | Yes | Event name (matches `reality.on` keys) |
| `params` | object | No | Event data payload (`map[string]any`) |

### snapshots.json

Array of `MachineSnapshot` objects:

```json
[
  {
    "resume": {
      "activeUniverses": { "my-universe": "IDLE" }
    },
    "snapshots": {},
    "tracking": { "my-universe": ["IDLE"] }
  }
]
```

## TUI Views & Keyboard Shortcuts

### Main Menu (Initial)

| Key | Action |
|---|---|
| `enter` | Select option |
| `q` / `ctrl+c` | Quit |

Options: Send Event, Load Snapshot, View History.

### Send Event View

Filterable list of events from `events.json`. Shows sent/unsent status.

| Key | Action |
|---|---|
| `enter` | Send selected event |
| `esc` | Back to main menu |
| `v` | View resume (active/finalized/superposition universes) |
| `m` | View machine snapshots |
| `t` | View tracking (visited realities history) |
| `s` | View context |
| type | Filter events by name |

### History View

List of previously sent events with their snapshots.

| Key | Action |
|---|---|
| `enter` | View event details |
| `esc` | Back to main menu |
| `v` | View resume at this point |
| `m` | View machine snapshots at this point |
| `t` | View tracking at this point |
| `s` | View context at this point |
| `r` | **Rollback** — restore machine to this snapshot and truncate subsequent history |

### JSON Viewer

Split-view JSON display for inspecting snapshots and context.

| Key | Action |
|---|---|
| `esc` | Back |
| arrows / `j`/`k` | Navigate |

## Bot Automation

Programmatic event-driven testing. Runs events in sequence, captures snapshots at each step.

### API

```go
import "github.com/rendis/statepro/v3/debugger/bot"

// EventProvider returns next event, or (nil, nil) to stop
type EventProvider func(currentSnapshot *instrumentation.MachineSnapshot) (instrumentation.Event, error)

type SMBot interface {
    Run(ctx context.Context, machineContext any) error
    GetHistory() []*EventHistory
    GetQuantumMachine() instrumentation.QuantumMachine
}

type EventHistory struct {
    Event    instrumentation.Event
    Snapshot *instrumentation.MachineSnapshot
}

// Create bot
func NewBot(
    qm instrumentation.QuantumMachine,
    eventProvider EventProvider,
    initQuantumMachine bool,  // true = call Init() before running
    opts ...BotOption,
) (SMBot, error)

// Options
func WithIgnoreUnhandledEvents(ignore bool) BotOption  // continue on unhandled events
```

### Flow

1. If `initQuantumMachine=true`: calls `qm.Init(ctx, machineContext)`
2. Loop: `EventProvider(currentSnapshot)` → `qm.SendEvent(ctx, event)` → record in history
3. Stops when EventProvider returns `nil` event or error
4. Access results via `bot.GetHistory()`

## CLI Example

```go
package main

import (
    "context"
    "github.com/rendis/statepro/v3/builtin"
    "github.com/rendis/statepro/v3/debugger/cli"
    "github.com/rendis/statepro/v3/instrumentation"
)

func main() {
    // Register custom executors
    builtin.RegisterAction("disableAdmission", func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
        return nil
    })
    builtin.RegisterCondition("checkTemplate", func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
        expected := args.GetCondition().Args["templateId"]
        actual := args.GetEvent().GetData()["templateId"]
        return expected == actual, nil
    })

    // Launch TUI
    debugger := cli.NewStateMachineDebugger()
    debugger.SetStateMachinePath("./state_machine.json")
    debugger.SetEventsPath("./events.json")
    debugger.SetSnapshotsPath("./snapshots.json")
    debugger.Run(nil) // nil machineContext OK for simple testing
}
```

## Bot Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/rendis/statepro/v3"
    "github.com/rendis/statepro/v3/builtin"
    "github.com/rendis/statepro/v3/debugger/bot"
    "github.com/rendis/statepro/v3/instrumentation"
)

func main() {
    // Register executors
    builtin.RegisterAction("disableAdmission", func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
        return nil
    })

    // Load machine
    jsonBytes, _ := os.ReadFile("state_machine.json")
    model, _ := statepro.DeserializeQuantumMachineFromBinary(jsonBytes)
    qm, _ := statepro.NewQuantumMachine(model)

    // Define event sequence
    events := []instrumentation.Event{
        statepro.NewEventBuilder("confirm").Build(),
        statepro.NewEventBuilder("fill-form").Build(),
        statepro.NewEventBuilder("sign").Build(),
    }

    count := 0
    provider := func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
        if count >= len(events) {
            return nil, nil // stop
        }
        event := events[count]
        count++
        return event, nil
    }

    // Run bot
    b, _ := bot.NewBot(qm, provider, true, bot.WithIgnoreUnhandledEvents(true))
    b.Run(context.Background(), nil)

    // Inspect history
    for i, h := range b.GetHistory() {
        fmt.Printf("Step %d: event=%s\n", i, h.Event.GetEventName())
    }
}
```
