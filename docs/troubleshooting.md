# Troubleshooting Guide

This guide helps you diagnose and resolve common issues when working with StatePro quantum state machines.

## Common Issues

### 1. Events Not Being Handled

**Symptoms:**
- `SendEvent` returns `handled = false`
- State machine doesn't transition as expected
- Events appear to be ignored

**Possible Causes & Solutions:**

#### Event Name Mismatch
```go
// ❌ Wrong - case sensitive
event := statepro.NewEventBuilder("Confirm").Build()

// ✅ Correct - matches JSON definition
event := statepro.NewEventBuilder("confirm").Build()
```

#### Current Reality Doesn't Accept Event
Check your JSON definition to ensure the current reality has the event handler:

```json
{
  "WAITING_CONFIRMATION": {
    "type": "transition",
    "on": {
      "confirm": [{"targets": ["CONFIRMED"]}],
      "cancel": [{"targets": ["CANCELLED"]}]
    }
  }
}
```

#### Observer/Guard Conditions Failing
```go
// Check if observers are rejecting the transition
builtin.RegisterObserver("observer:businessHours", func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    // Log the evaluation for debugging
    result := isBusinessHours()
    log.Printf("Business hours check: %v", result)
    return result, nil
})
```

#### No Active Universes
Verify that universes are properly initialized:

```go
snapshot := qm.GetSnapshot()
resume := snapshot.GetResume()
log.Printf("Active universes: %v", resume.ActiveUniverses)
```

### 2. Machine Initialization Failures

**Symptoms:**
- `NewQuantumMachine` returns an error
- `Init` method fails
- Invalid model errors

**Common Solutions:**

#### JSON Validation Errors
```bash
# Use a JSON validator to check syntax
cat state_machine.json | jq .
```

#### Missing Required Fields
Ensure your JSON has all required fields:

```json
{
  "id": "required-field",
  "initials": ["U:universe-name"],
  "universes": {
    "universe-name": {
      "id": "universe-name",
      "realities": {
        "STATE_NAME": {
          "type": "transition"
        }
      }
    }
  }
}
```

#### Invalid Reality Types
```json
{
  "SOME_STATE": {
    "type": "transition",     // ✅ Valid: "transition", "final", "unsuccessfulFinal"
    "on": {
      "event": [{"targets": ["NEXT_STATE"]}]
    }
  }
}
```

### 3. Runtime Errors During Event Processing

**Symptoms:**
- `SendEvent` returns an error
- Actions or observers throw exceptions
- Unexpected application crashes

**Debugging Steps:**

#### Enable Detailed Logging
```go
// Add logging to your custom executors
builtin.RegisterAction("action:processData", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    log.Printf("Executing processData action with event: %+v", args.GetEventData())

    err := doProcessing()
    if err != nil {
        log.Printf("Processing failed: %v", err)
        return err
    }

    return nil
})
```

#### Check Action/Observer Implementations
```go
// Handle potential nil pointer errors
builtin.RegisterAction("action:updateUser", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    data := args.GetEventData()
    userID, ok := data["userId"].(string)
    if !ok || userID == "" {
        return fmt.Errorf("missing or invalid userId in event data")
    }

    // Continue with processing...
    return nil
})
```

### 4. Superposition Issues

**Symptoms:**
- Multiple realities active when expecting one
- Unexpected parallel execution
- Events processed multiple times

**Understanding Superposition:**

Superposition occurs when:
- Multiple universes are active simultaneously
- A universe is in multiple realities at once
- Transitions spawn new universes

**Solutions:**

#### Check Initial Configuration
```json
{
  "initials": [
    "U:universe1",    // Starts universe1
    "U:universe2"     // Also starts universe2 - both will be active!
  ]
}
```

#### Monitor Active States
```go
func debugActiveStates(qm instrumentation.QuantumMachine) {
    snapshot := qm.GetSnapshot()
    resume := snapshot.GetResume()

    for universe, realities := range resume.ActiveUniverses {
        log.Printf("Universe %s has realities: %v", universe, realities)
    }
}
```

### 5. Memory and Performance Issues

**Symptoms:**
- High memory usage
- Slow event processing
- Growing snapshot sizes

**Optimization Tips:**

#### Minimize Event Data
```go
// ❌ Avoid large payloads
event := statepro.NewEventBuilder("process").
    SetData(map[string]any{
        "largeFile": fileContents, // Could be MBs
    }).
    Build()

// ✅ Use references instead
event := statepro.NewEventBuilder("process").
    SetData(map[string]any{
        "fileId": "abc123",
    }).
    Build()
```

#### Limit Tracking History
Currently, all state changes are tracked. For high-throughput scenarios, consider:
- Periodic snapshot clearing (if available in future versions)
- External state persistence
- Event sourcing patterns

## Debugging Tools

### 1. Interactive CLI Debugger

The Bubble Tea debugger is your primary tool for understanding machine behavior:

```bash
# Copy an example to start debugging
cp -r example/cli debug-session
cd debug-session

# Modify state_machine.json and events.json for your scenario
vim state_machine.json

# Launch the debugger
go run main.go
```

**Debugger Features:**
- Real-time state visualization
- Event replay and testing
- Snapshot inspection
- Interactive event sending

### 2. Programmatic Bot

For automated testing and reproducible scenarios:

```bash
# Copy the bot example
cp -r example/bot test-automation
cd test-automation

# Customize the event sequence
vim events.json

# Run the automation
go run main.go
```

### 3. Custom Logging

Add comprehensive logging to understand execution flow:

```go
builtin.RegisterAction("action:debug", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    log.Printf("=== DEBUG ACTION ===")
    log.Printf("Current Universe: %s", args.GetCurrentUniverse())
    log.Printf("Current Reality: %s", args.GetCurrentReality())
    log.Printf("Event Name: %s", args.GetEventName())
    log.Printf("Event Data: %+v", args.GetEventData())
    log.Printf("Machine Context: %+v", args.GetContext())
    log.Printf("===================")
    return nil
})
```

## Common Patterns and Solutions

### Pattern: Conditional Transitions

**Problem:** Need different transition behavior based on data

**Solution:** Use observers with context checking

```go
builtin.RegisterObserver("observer:checkUserRole", func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    data := args.GetEventData()
    role, ok := data["userRole"].(string)
    if !ok {
        return false, fmt.Errorf("userRole missing from event")
    }

    return role == "admin", nil
})
```

### Pattern: Error Handling

**Problem:** Need to handle failures gracefully

**Solution:** Use unsuccessful final states

```json
{
  "PROCESSING": {
    "type": "transition",
    "on": {
      "success": [{"targets": ["COMPLETED"]}],
      "error": [{"targets": ["FAILED"]}]
    }
  },
  "COMPLETED": {
    "type": "final"
  },
  "FAILED": {
    "type": "unsuccessfulFinal",
    "metadata": {
      "errorHandling": "retry"
    }
  }
}
```

### Pattern: Async Operations

**Problem:** Need to trigger background work

**Solution:** Use invokes for non-blocking operations

```json
{
  "PROCESSING": {
    "type": "transition",
    "always": [{
      "targets": ["WAITING_CALLBACK"],
      "invokes": [{"executor": "invoke:backgroundJob"}]
    }]
  }
}
```

```go
builtin.RegisterInvoke("invoke:backgroundJob", func(ctx context.Context, args instrumentation.InvokeExecutorArgs) error {
    go func() {
        // Do background work
        time.Sleep(5 * time.Second)

        // Send completion event
        event := statepro.NewEventBuilder("jobComplete").Build()
        qm.SendEvent(context.Background(), event)
    }()

    return nil
})
```

## Performance Tips

### 1. Event Design
- Keep event data minimal
- Use consistent naming conventions
- Include only necessary context

### 2. Observer Optimization
- Make guard conditions fast
- Avoid heavy computations
- Cache expensive lookups

### 3. Action Efficiency
- Keep actions focused and quick
- Use invokes for heavy operations
- Minimize external dependencies

### 4. Model Structure
- Keep individual universes simple
- Avoid deeply nested transitions
- Use metadata for debugging info only

## When to Ask for Help

If you've tried the above solutions and still have issues:

1. **Create a Minimal Reproduction**
   - Simplify your state machine to the essential problem
   - Create a small Go program that demonstrates the issue
   - Include the JSON definition and relevant Go code

2. **Gather Debug Information**
   - Current StatePro version (`go.mod`)
   - Go version (`go version`)
   - Complete error messages and stack traces
   - Relevant logs with timestamps

3. **Document Expected vs Actual Behavior**
   - What should happen
   - What actually happens
   - Steps to reproduce

4. **Check Existing Resources**
   - [Examples directory](../example/) for similar patterns
   - [API Reference](api-reference.md) for method documentation
   - [Concepts guide](concepts.md) for understanding fundamentals

Remember: StatePro's quantum nature means that understanding superposition and multi-universe execution is key to successful debugging. When in doubt, use the interactive debugger to visualize what's actually happening in your state machine.