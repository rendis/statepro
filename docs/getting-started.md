# Getting Started with StatePro

This guide will help you build your first quantum state machine with StatePro in just a few minutes.

## Prerequisites

- Go 1.22 or newer
- Basic understanding of state machines
- Familiarity with JSON

## Installation

```bash
go get github.com/rendis/statepro/v3
```

## Your First Quantum Machine

Let's build a simple traffic light state machine to understand the core concepts.

### Step 1: Define the State Machine

Create a file called `traffic_light.json`:

```json
{
  "id": "traffic-light",
  "canonicalName": "simple-traffic-light",
  "version": "1.0.0",
  "initials": ["U:main-traffic-flow"],
  "universes": {
    "main-traffic-flow": {
      "id": "main-traffic-flow",
      "initial": "RED",
      "realities": {
        "RED": {
          "type": "transition",
          "on": {
            "timer": [{"targets": ["GREEN"]}]
          },
          "metadata": {
            "duration": 30,
            "color": "#FF0000"
          }
        },
        "GREEN": {
          "type": "transition",
          "on": {
            "timer": [{"targets": ["YELLOW"]}]
          },
          "metadata": {
            "duration": 45,
            "color": "#00FF00"
          }
        },
        "YELLOW": {
          "type": "transition",
          "on": {
            "timer": [{"targets": ["RED"]}]
          },
          "metadata": {
            "duration": 5,
            "color": "#FFFF00"
          }
        }
      }
    }
  }
}
```

### Step 2: Create the Go Application

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/rendis/statepro/v3"
    "github.com/rendis/statepro/v3/builtin"
    "github.com/rendis/statepro/v3/instrumentation"
)

func main() {
    // Load the state machine definition
    raw, err := os.ReadFile("traffic_light.json")
    if err != nil {
        log.Fatal("Error reading state machine:", err)
    }

    // Deserialize the model
    model, err := statepro.DeserializeQuantumMachineFromBinary(raw)
    if err != nil {
        log.Fatal("Error deserializing model:", err)
    }

    // Create the quantum machine
    qm, err := statepro.NewQuantumMachine(model)
    if err != nil {
        log.Fatal("Error creating quantum machine:", err)
    }

    // Initialize the machine
    ctx := context.Background()
    if err := qm.Init(ctx, nil); err != nil {
        log.Fatal("Error initializing machine:", err)
    }

    // Display initial state
    displayCurrentState(qm)

    // Simulate traffic light cycles
    for i := 0; i < 6; i++ {
        time.Sleep(2 * time.Second) // Simulate time passing

        // Send timer event
        timerEvent := statepro.NewEventBuilder("timer").Build()
        handled, err := qm.SendEvent(ctx, timerEvent)
        if err != nil {
            log.Fatal("Error sending event:", err)
        }

        if !handled {
            log.Println("Timer event was not handled")
        }

        displayCurrentState(qm)
    }
}

func displayCurrentState(qm instrumentation.QuantumMachine) {
    snapshot := qm.GetSnapshot()
    resume := snapshot.GetResume()

    fmt.Printf("🚦 Current State: %v\n", resume.ActiveUniverses)

    for universe, realities := range resume.ActiveUniverses {
        for _, reality := range realities {
            fmt.Printf("   Universe: %s, Reality: %s\n", universe, reality)
        }
    }
    fmt.Println("---")
}
```

### Step 3: Run the Application

```bash
go mod init traffic-light-demo
go mod tidy
go run main.go
```

You should see output like:

```plaintext
🚦 Current State: map[main-traffic-flow:[RED]]
   Universe: main-traffic-flow, Reality: RED
---
🚦 Current State: map[main-traffic-flow:[GREEN]]
   Universe: main-traffic-flow, Reality: GREEN
---
🚦 Current State: map[main-traffic-flow:[YELLOW]]
   Universe: main-traffic-flow, Reality: YELLOW
---
```

## Understanding the Example

### Key Concepts Demonstrated

1. **Quantum Machine**: The root container (`traffic-light`)
2. **Universe**: A single state machine (`main-traffic-flow`)
3. **Realities**: Individual states (`RED`, `GREEN`, `YELLOW`)
4. **Events**: Triggers for transitions (`timer`)
5. **Metadata**: Additional state information (duration, color)

### JSON Structure Breakdown

- `initials`: Defines which universes start active
- `universes`: Map of all available state machines
- `realities`: States within each universe
- `type: "transition"`: Allows transitions to other states
- `on`: Event-to-target mappings

## Next Steps

Now that you have a basic understanding, explore these advanced features:

### 1. Add Custom Actions

```go
// Register a custom action that runs when entering a state
_ = builtin.RegisterAction("action:logStateChange", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    fmt.Printf("🔄 State changed to: %s\n", args.GetCurrentReality())
    return nil
})
```

### 2. Add Observers (Guards)

```go
// Register an observer that controls when transitions can happen
_ = builtin.RegisterObserver("observer:timeOfDay", func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    // Only allow transitions during daytime
    hour := time.Now().Hour()
    return hour >= 6 && hour <= 18, nil
})
```

### 3. Explore Multi-Universe Scenarios

Try creating state machines with multiple parallel universes that can interact and spawn new universes based on events.

### 4. Use the Interactive Debugger

```bash
# Copy one of the example state machines
cp -r example/cli ./my-debug-session
cd my-debug-session

# Launch the interactive debugger
go run main.go
```

## Common Patterns

### Simple State Machine

- Single universe with linear transitions
- Good for workflows, approval processes

### Parallel Processing

- Multiple universes active simultaneously
- Good for concurrent workflows

### Conditional Branching

- Use observers to guard transitions
- Good for business rule enforcement

### Event Accumulation

- Collect multiple events before transitioning
- Good for batch processing scenarios

## Troubleshooting

### Events Not Handled

- Check that event names match exactly
- Verify the current reality accepts the event
- Ensure observers (if any) return true

### Unexpected States

- Use the debugger CLI to inspect active universes
- Check the snapshots to trace execution history
- Verify your JSON syntax and structure

## Resources

- 📖 [Complete Concepts Guide](concepts.md)
- 📝 [Detailed Modeling Reference](modeling.md)
- 🔧 [Debugging Tools](debugging.md)
- 🛠️ [API Reference](instrumentation.md)

Ready to build more complex state machines? Check out the [Modeling Guide](modeling.md) for advanced patterns and techniques.
