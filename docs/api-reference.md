# API Reference

This document provides a comprehensive reference for StatePro's public API. All types and methods are defined under the `github.com/rendis/statepro/v3` module.

## Table of Contents

- [Core Functions](#core-functions)
- [Model Types](#model-types)
- [Instrumentation Interfaces](#instrumentation-interfaces)
- [Event System](#event-system)
- [Registry System](#registry-system)
- [Error Types](#error-types)

## Core Functions

### `NewQuantumMachine`

Creates a new quantum machine from a model.

```go
func NewQuantumMachine(model theoretical.QuantumMachineModel, opts ...QuantumMachineOption) (instrumentation.QuantumMachine, error)
```

**Parameters:**

- `model` - The quantum machine model containing universe definitions
- `opts` - Optional configuration options (currently none available)

**Returns:**

- `instrumentation.QuantumMachine` - The executable quantum machine
- `error` - Any initialization errors

**Example:**

```go
qm, err := statepro.NewQuantumMachine(model)
if err != nil {
    log.Fatal("Failed to create quantum machine:", err)
}
```

### Deserialization Functions

#### `DeserializeQuantumMachineFromBinary`

Deserializes a quantum machine model from JSON bytes.

```go
func DeserializeQuantumMachineFromBinary(data []byte) (theoretical.QuantumMachineModel, error)
```

**Parameters:**

- `data` - JSON bytes representing the machine model

**Returns:**

- `theoretical.QuantumMachineModel` - The deserialized model
- `error` - Any deserialization errors

#### `DeserializeQuantumMachineFromFile`

Deserializes a quantum machine model from a JSON file.

```go
func DeserializeQuantumMachineFromFile(path string) (theoretical.QuantumMachineModel, error)
```

**Parameters:**

- `path` - File path to the JSON definition

**Returns:**

- `theoretical.QuantumMachineModel` - The deserialized model
- `error` - Any file reading or deserialization errors

### Event Builder Functions

#### `NewEventBuilder`

Creates a new event builder for constructing events.

```go
func NewEventBuilder(name string) instrumentation.EventBuilder
```

**Parameters:**

- `name` - The event name/type

**Returns:**

- `instrumentation.EventBuilder` - A builder for constructing the event

**Example:**

```go
event := statepro.NewEventBuilder("confirm").
    SetData(map[string]any{"userId": "123"}).
    SetCorrelationId("req-456").
    Build()
```

## Model Types

### `QuantumMachineModel`

The root model representing a complete quantum state machine.

```go
type QuantumMachineModel struct {
    Id                string                        `json:"id"`
    CanonicalName     string                        `json:"canonicalName,omitempty"`
    Version           string                        `json:"version,omitempty"`
    Initials          []string                      `json:"initials"`
    Universes         map[string]UniverseModel      `json:"universes"`
    UniversalConstants *UniversalConstantsModel     `json:"universalConstants,omitempty"`
    Metadata          map[string]any                `json:"metadata,omitempty"`
}
```

### `UniverseModel`

Represents a single universe within a quantum machine.

```go
type UniverseModel struct {
    Id             string                    `json:"id"`
    CanonicalName  string                    `json:"canonicalName,omitempty"`
    Version        string                    `json:"version,omitempty"`
    Initial        string                    `json:"initial,omitempty"`
    Realities      map[string]RealityModel   `json:"realities"`
    Constants      *ConstantsModel           `json:"constants,omitempty"`
    Metadata       map[string]any            `json:"metadata,omitempty"`
}
```

### `RealityModel`

Defines a single state/reality within a universe.

```go
type RealityModel struct {
    Type           RealityType                      `json:"type"`
    On             map[string][]TransitionModel     `json:"on,omitempty"`
    Always         []TransitionModel                `json:"always,omitempty"`
    Entry          []string                         `json:"entry,omitempty"`
    Exit           []string                         `json:"exit,omitempty"`
    Metadata       map[string]any                   `json:"metadata,omitempty"`
}
```

### `TransitionModel`

Defines a transition between realities.

```go
type TransitionModel struct {
    Targets    []string       `json:"targets"`
    Observers  []string       `json:"observers,omitempty"`
    Actions    []string       `json:"actions,omitempty"`
    Invokes    []InvokeModel  `json:"invokes,omitempty"`
    Metadata   map[string]any `json:"metadata,omitempty"`
}
```

### Reality Types

```go
type RealityType string

const (
    RealityTypeTransition         RealityType = "transition"
    RealityTypeFinal             RealityType = "final"
    RealityTypeUnsuccessfulFinal RealityType = "unsuccessfulFinal"
)
```

## Instrumentation Interfaces

### `QuantumMachine`

The main interface for interacting with a quantum state machine.

```go
type QuantumMachine interface {
    Init(ctx context.Context, machineContext any) error
    SendEvent(ctx context.Context, event Event) (bool, error)
    GetSnapshot() Snapshot
}
```

#### Methods

##### `Init`

Initializes the quantum machine with optional context.

**Parameters:**

- `ctx` - Go context for cancellation/timeout
- `machineContext` - Optional context object available to all executors

**Returns:**

- `error` - Any initialization errors

##### `SendEvent`

Sends an event to the quantum machine for processing.

**Parameters:**

- `ctx` - Go context for cancellation/timeout
- `event` - The event to process

**Returns:**

- `bool` - True if the event was handled by any universe
- `error` - Any processing errors

##### `GetSnapshot`

Gets the current state snapshot of the machine.

**Returns:**

- `Snapshot` - Current machine state and history

### `Event`

Represents an event that can trigger state transitions.

```go
type Event interface {
    GetEventName() string
    GetData() map[string]any
    GetCorrelationId() string
    GetTimestamp() time.Time
}
```

### `EventBuilder`

Builder for constructing events.

```go
type EventBuilder interface {
    SetData(data map[string]any) EventBuilder
    SetCorrelationId(correlationId string) EventBuilder
    Build() Event
}
```

### `Snapshot`

Represents the current state of a quantum machine.

```go
type Snapshot interface {
    GetResume() Resume
    GetTrackingHistory() []TrackingEntry
}
```

### `Resume`

Contains the active state information.

```go
type Resume interface {
    ActiveUniverses map[string][]string  // universe -> realities
}
```

## Event System

### Event Structure

Events in StatePro carry the following information:

- **Name**: The event type/identifier
- **Data**: Key-value payload data
- **Correlation ID**: For tracking related events
- **Timestamp**: When the event was created

### Event Processing

Events are processed as follows:

1. Event is sent to all active universes
2. Each universe checks if current realities accept the event
3. Observers (guards) are evaluated for applicable transitions
4. Actions are executed during transitions
5. Invokes are triggered (async operations)
6. New realities are activated

### Event Example

```go
// Create an event with data
event := statepro.NewEventBuilder("user.signup").
    SetData(map[string]any{
        "email": "user@example.com",
        "plan": "premium",
    }).
    SetCorrelationId("signup-123").
    Build()

// Send the event
handled, err := qm.SendEvent(ctx, event)
```

## Registry System

The builtin registry allows registration of custom executors.

### Action Registration

```go
func RegisterAction(name string, executor ActionExecutor) error
```

**Example:**

```go
builtin.RegisterAction("action:sendEmail", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    email := args.GetEventData()["email"].(string)
    return sendEmail(email)
})
```

### Observer Registration

```go
func RegisterObserver(name string, executor ObserverExecutor) error
```

**Example:**

```go
builtin.RegisterObserver("observer:businessHours", func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    hour := time.Now().Hour()
    return hour >= 9 && hour <= 17, nil
})
```

### Condition Registration

```go
func RegisterCondition(name string, executor ConditionExecutor) error
```

### Invoke Registration

```go
func RegisterInvoke(name string, executor InvokeExecutor) error
```

## Error Types

### Common Errors

- **Deserialization Errors**: Invalid JSON structure
- **Validation Errors**: Missing required fields or invalid values
- **Runtime Errors**: Event processing failures
- **Executor Errors**: Custom action/observer failures

### Error Handling

```go
qm, err := statepro.NewQuantumMachine(model)
if err != nil {
    // Handle model validation errors
    log.Fatal("Model validation failed:", err)
}

handled, err := qm.SendEvent(ctx, event)
if err != nil {
    // Handle event processing errors
    log.Error("Event processing failed:", err)
}
```

## Best Practices

### Model Design

- Use descriptive IDs for universes and realities
- Include metadata for debugging and tooling
- Keep individual universes focused on single concerns

### Event Design

- Use consistent naming conventions (e.g., `domain.action`)
- Include relevant context in event data
- Use correlation IDs for tracking related events

### Error Handling

- Always check errors from `NewQuantumMachine`
- Handle both processing errors and unhandled events
- Use appropriate logging levels for debugging

### Performance

- Minimize work in observers (guards)
- Use async invokes for non-critical operations
- Consider event batching for high-throughput scenarios

## Integration Examples

### Web Handler Integration

```go
func handleUserAction(w http.ResponseWriter, r *http.Request) {
    event := statepro.NewEventBuilder("user.action").
        SetData(extractRequestData(r)).
        SetCorrelationId(r.Header.Get("X-Request-ID")).
        Build()

    handled, err := qm.SendEvent(r.Context(), event)
    if err != nil {
        http.Error(w, "Processing failed", 500)
        return
    }

    if !handled {
        http.Error(w, "Invalid action", 400)
        return
    }

    w.WriteHeader(200)
}
```

### Background Job Integration

```go
func processJob(job Job) error {
    event := statepro.NewEventBuilder("job.process").
        SetData(job.Data).
        SetCorrelationId(job.ID).
        Build()

    handled, err := qm.SendEvent(context.Background(), event)
    if err != nil {
        return fmt.Errorf("job processing failed: %w", err)
    }

    if !handled {
        return fmt.Errorf("job type not supported: %s", job.Type)
    }

    return nil
}
```

For more detailed examples, see the [examples](../example/) directory in the repository.
