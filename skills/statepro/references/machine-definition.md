# Machine Definition Reference

## Table of Contents

1. [Model Hierarchy](#model-hierarchy)
2. [QuantumMachineModel](#quantummachinemodel)
3. [UniverseModel](#universemodel)
4. [RealityModel](#realitymodel)
5. [TransitionModel](#transitionmodel)
6. [ObserverModel](#observermodel)
7. [ActionModel](#actionmodel)
8. [InvokeModel](#invokemodel)
9. [ConditionModel](#conditionmodel)
10. [UniversalConstantsModel](#universalconstantsmodel)
11. [Validation Rules](#validation-rules)
12. [Complete Example](#complete-example)

## Model Hierarchy

```
QuantumMachineModel
├── id, canonicalName, version (required)
├── initials[] (required, external refs: "U:universeId" or "U:universeId:realityId")
├── description, metadata (optional)
├── universalConstants (optional)
└── universes{} (required, map)
    └── UniverseModel
        ├── id, canonicalName, version (required)
        ├── initial (optional, must exist in realities map)
        ├── description, metadata, tags (optional)
        ├── universalConstants (optional)
        └── realities{} (required, map)
            └── RealityModel
                ├── id, type (required: transition|final|unsuccessfulFinal)
                ├── on{} (optional, event name → TransitionModel[])
                ├── always[] (optional, auto-evaluated transitions)
                ├── observers[] (optional, superposition watchers)
                ├── entryActions[], exitActions[] (optional, sync)
                ├── entryInvokes[], exitInvokes[] (optional, async)
                └── description, metadata (optional)
```

## QuantumMachineModel

```go
type QuantumMachineModel struct {
    ID                 string                    `json:"id"`                          // required
    CanonicalName      string                    `json:"canonicalName"`               // required
    Version            string                    `json:"version"`                     // required
    Universes          map[string]*UniverseModel `json:"universes"`                   // required, size > 0
    Initials           []string                  `json:"initials"`                    // required, external refs
    UniversalConstants *UniversalConstantsModel  `json:"universalConstants,omitempty"`
    Description        *string                   `json:"description,omitempty"`
    Metadata           map[string]any            `json:"metadata,omitempty"`
}
```

**Initials**: List of universe references to activate on `Init()`. Format: `"U:universeId"` (starts at universe's initial reality) or `"U:universeId:realityId"` (starts at specific reality).

## UniverseModel

```go
type UniverseModel struct {
    ID                 string                   `json:"id"`
    CanonicalName      string                   `json:"canonicalName"`
    Version            string                   `json:"version"`
    Initial            *string                  `json:"initial,omitempty"`        // must exist in realities
    Realities          map[string]*RealityModel `json:"realities"`                // required, size > 0
    UniversalConstants *UniversalConstantsModel `json:"universalConstants,omitempty"`
    Description        *string                  `json:"description,omitempty"`
    Metadata           map[string]any           `json:"metadata,omitempty"`
    Tags               []string                 `json:"tags,omitempty"`
}
```

**Initial**: The default reality when the universe starts. If omitted, the universe must be targeted with `U:universeId:realityId` format.

## RealityModel

```go
type RealityType string
const (
    RealityTypeTransition        RealityType = "transition"        // intermediate
    RealityTypeFinal             RealityType = "final"             // successful end
    RealityTypeUnsuccessfulFinal RealityType = "unsuccessfulFinal" // failed end
)

type RealityModel struct {
    ID           string                        `json:"id"`
    Type         RealityType                   `json:"type"`                    // required
    On           map[string][]*TransitionModel `json:"on,omitempty"`            // event handlers
    Observers    []*ObserverModel              `json:"observers,omitempty"`     // superposition watchers
    Always       []*TransitionModel            `json:"always,omitempty"`        // auto transitions
    EntryActions []*ActionModel                `json:"entryActions,omitempty"`  // sync on entry
    ExitActions  []*ActionModel                `json:"exitActions,omitempty"`   // sync on exit
    EntryInvokes []*InvokeModel                `json:"entryInvokes,omitempty"`  // async on entry
    ExitInvokes  []*InvokeModel                `json:"exitInvokes,omitempty"`   // async on exit
    Description  *string                       `json:"description,omitempty"`
    Metadata     map[string]any                `json:"metadata,omitempty"`
}
```

**Execution order on entry**: constants.entryActions → reality.entryActions → constants.entryInvokes → reality.entryInvokes → process emitted events → evaluate `always` transitions.

**Observers**: Execute only during superposition. First observer returning `true` collapses superposition to this reality.

**Always**: Auto-evaluated transitions. First to pass conditions executes. Re-evaluated after each transition.

## TransitionModel

```go
type TransitionType string
const (
    TransitionTypeDefault TransitionType = "default" // abandon current reality
    TransitionTypeNotify  TransitionType = "notify"  // keep current reality
)

type TransitionModel struct {
    Condition  *ConditionModel   `json:"condition,omitempty"`   // single condition
    Conditions []*ConditionModel `json:"conditions,omitempty"`  // multiple (AND logic)
    Type       *TransitionType   `json:"type,omitempty"`        // default if nil
    Targets    []string          `json:"targets"`               // required, size > 0
    Actions    []*ActionModel    `json:"actions,omitempty"`     // sync on transition
    Invokes    []*InvokeModel    `json:"invokes,omitempty"`     // async on transition
    Description *string          `json:"description,omitempty"`
    Metadata   map[string]any    `json:"metadata,omitempty"`
}
```

**Target formats**:

| Format | Example | Meaning |
|---|---|---|
| `RealityID` | `"RUNNING"` | Internal reality in same universe |
| `U:UniverseID` | `"U:payment-flow"` | External universe at its initial reality |
| `U:UniverseID:RealityID` | `"U:payment-flow:PAID"` | External universe at specific reality |

**Transition type**:
- `default` (or omitted): Current reality is abandoned, targets become new realities
- `notify`: Targets receive the event but current reality remains active

**Execution order**: constants.actionsOnTransition → transition.actions → constants.invokesOnTransition → transition.invokes → apply targets.

**Conditions**: If both `condition` and `conditions` specified, all must pass (AND). If any returns false, transition is skipped.

## ObserverModel

```go
type ObserverModel struct {
    Src         string         `json:"src"`                   // required
    Args        map[string]any `json:"args,omitempty"`
    Description *string        `json:"description,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
}
```

Observers only execute during superposition. They receive `AccumulatorStatistics` to inspect accumulated events. First to return `true` wins — collapse is triggered for that reality.

## ActionModel

```go
type ActionModel struct {
    Src         string         `json:"src"`
    Args        map[string]any `json:"args,omitempty"`
    Description *string        `json:"description,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
}
```

Synchronous. Error aborts the flow. Entry actions can call `EmitEvent(name, data)` to queue internal events.

## InvokeModel

```go
type InvokeModel struct {
    Src         string         `json:"src"`
    Args        map[string]any `json:"args,omitempty"`
    Description *string        `json:"description,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
}
```

Async fire-and-forget. Failures do not affect reality transitions.

## ConditionModel

```go
type ConditionModel struct {
    Src         string         `json:"src"`
    Args        map[string]any `json:"args,omitempty"`
    Description *string        `json:"description,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
}
```

Must return `true` for the transition to execute. Multiple conditions on a transition use AND logic.

## UniversalConstantsModel

```go
type UniversalConstantsModel struct {
    EntryInvokes        []*InvokeModel `json:"entryInvokes,omitempty"`
    ExitInvokes         []*InvokeModel `json:"exitInvokes,omitempty"`
    EntryActions        []*ActionModel `json:"entryActions,omitempty"`
    ExitActions         []*ActionModel `json:"exitActions,omitempty"`
    InvokesOnTransition []*InvokeModel `json:"invokesOnTransition,omitempty"`
    ActionsOnTransition []*ActionModel `json:"actionsOnTransition,omitempty"`
}
```

Can be defined at machine-level or universe-level. Execute BEFORE reality-specific operations.

## Validation Rules

### ID Pattern
```
^[A-Za-z](?:[A-Za-z0-9_-]*[A-Za-z0-9])?$
```
Start with letter, alphanumeric/underscore/hyphen, end with alphanumeric. Min length 1.

### Version Pattern
```
^[A-Za-z0-9][A-Za-z0-9._-]*$
```
Examples: `1.0.0`, `v1`, `1.0.0-alpha.1`

### Behavior Source Pattern
```
^\S+$
```
No whitespace. Examples: `builtin:observer:containsAllEvents`, `myAction`, `namespace:name`

### Argument Key Pattern
```
^[A-Za-z][A-Za-z0-9_-]*$
```
Start with letter, alphanumeric/underscore/hyphen.

### Argument Values
String, number, boolean, null, nested objects, arrays (recursive).

## Complete Example

7-universe admission process with superposition, cross-universe transitions, observers, and final states:

```json
{
  "id": "admission",
  "canonicalName": "admission-machine",
  "version": "0.1.0",
  "description": "Define all flows of an admission",
  "initials": ["U:admission-in-waiting-confirmation"],
  "universes": {
    "admission-in-waiting-confirmation": {
      "id": "admission-in-waiting-confirmation",
      "canonicalName": "admission-in-waiting-confirmation",
      "version": "0.1.0",
      "initial": "CREATED",
      "realities": {
        "CREATED": {
          "id": "CREATED",
          "type": "transition",
          "always": [{ "targets": ["WAITING_CONFIRMATION"] }]
        },
        "WAITING_CONFIRMATION": {
          "id": "WAITING_CONFIRMATION",
          "type": "transition",
          "on": {
            "confirm": [{ "targets": ["CONFIRMED"] }],
            "reject": [{ "targets": ["U:admission-rejected"] }]
          }
        },
        "CONFIRMED": {
          "id": "CONFIRMED",
          "type": "final",
          "always": [{
            "targets": ["U:admission-form-process", "U:admission-contract-process"]
          }]
        }
      }
    },
    "admission-form-process": {
      "id": "admission-form-process",
      "canonicalName": "admission-form-process",
      "version": "0.1.0",
      "initial": "FILLING_FORM",
      "realities": {
        "FILLING_FORM": {
          "id": "FILLING_FORM",
          "type": "transition",
          "on": {
            "fill-form": [{ "targets": ["FILLED"] }],
            "manual-complete": [{ "targets": ["FILLED"] }]
          }
        },
        "FILLED": {
          "id": "FILLED",
          "type": "final",
          "always": [{ "targets": ["U:admission-waiting-processes:WAITING_PROCESSES"] }]
        }
      }
    },
    "admission-contract-process": {
      "id": "admission-contract-process",
      "canonicalName": "admission-contract-process",
      "version": "0.1.0",
      "initial": "SIGNING_DATA_CONSENT",
      "realities": {
        "SIGNING_DATA_CONSENT": {
          "id": "SIGNING_DATA_CONSENT",
          "type": "transition",
          "on": {
            "sign": [{ "targets": ["DATA_CONSENT_SIGNED"] }],
            "manual-complete": [{ "targets": ["SIGNING_DATA_CONSENT"] }]
          }
        },
        "DATA_CONSENT_SIGNED": {
          "id": "DATA_CONSENT_SIGNED",
          "type": "transition",
          "always": [{ "targets": ["U:admission-waiting-processes:WAITING_PROCESSES"] }]
        },
        "SIGNING_ENROLLMENT_CONTRACT": {
          "id": "SIGNING_ENROLLMENT_CONTRACT",
          "type": "transition",
          "on": {
            "sign": [{ "targets": ["ENROLLMENT_CONTRACT_SIGNED"] }],
            "manual-complete": [{ "targets": ["ENROLLMENT_CONTRACT_SIGNED"] }]
          }
        },
        "ENROLLMENT_CONTRACT_SIGNED": {
          "id": "ENROLLMENT_CONTRACT_SIGNED",
          "type": "final",
          "always": [{ "targets": ["U:admission-completed"] }]
        }
      }
    },
    "admission-waiting-processes": {
      "id": "admission-waiting-processes",
      "canonicalName": "waiting-processes",
      "version": "0.1.0",
      "realities": {
        "WAITING_PROCESSES": {
          "id": "WAITING_PROCESSES",
          "type": "final",
          "observers": [
            {
              "src": "builtin:observer:containsAllEvents",
              "args": { "p1": "fill-form", "p2": "sign" }
            },
            {
              "src": "builtin:observer:containsAllEvents",
              "args": { "p1": "manual-complete" }
            }
          ],
          "always": [{
            "targets": ["U:admission-contract-process:SIGNING_ENROLLMENT_CONTRACT"]
          }]
        }
      }
    },
    "admission-rejected": {
      "id": "admission-rejected",
      "canonicalName": "admission-rejected",
      "version": "0.1.0",
      "initial": "REJECTED",
      "realities": {
        "REJECTED": {
          "id": "REJECTED",
          "type": "unsuccessfulFinal",
          "entryActions": [{
            "src": "disableAdmission",
            "description": "Set admission context to disabled"
          }]
        }
      }
    },
    "admission-cancelled": {
      "id": "admission-cancelled",
      "canonicalName": "admission-cancelled",
      "version": "0.1.0",
      "realities": {
        "CANCELLED": {
          "id": "CANCELLED",
          "type": "unsuccessfulFinal",
          "entryActions": [{
            "src": "disableAdmission",
            "description": "Set admission context to disabled"
          }]
        }
      }
    },
    "admission-completed": {
      "id": "admission-completed",
      "canonicalName": "admission-completed",
      "version": "0.1.0",
      "initial": "COMPLETED",
      "realities": {
        "COMPLETED": {
          "id": "COMPLETED",
          "type": "final",
          "on": {
            "cancel": [{ "targets": ["U:admission-cancelled"] }]
          }
        }
      }
    }
  }
}
```

**Flow**: `CREATED` → auto → `WAITING_CONFIRMATION` → on "confirm" → `CONFIRMED` → auto → starts `admission-form-process` + `admission-contract-process` in parallel → both feed into `admission-waiting-processes` (superposition) → observer checks if "fill-form" + "sign" both accumulated → collapses → continues to enrollment contract → `COMPLETED`.
