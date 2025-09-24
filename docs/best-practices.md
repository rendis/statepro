# Best Practices

This guide covers recommended patterns and practices for building robust, maintainable quantum state machines with StatePro.

## Machine Design Principles

### 1. Single Responsibility Universes

Each universe should represent a single, cohesive workflow or process.

```json
// ✅ Good: Each universe has a clear, focused responsibility
{
  "universes": {
    "user-authentication": {
      "realities": {
        "AUTHENTICATING": {...},
        "AUTHENTICATED": {...},
        "AUTHENTICATION_FAILED": {...}
      }
    },
    "order-processing": {
      "realities": {
        "PENDING": {...},
        "PROCESSING": {...},
        "COMPLETED": {...}
      }
    }
  }
}
```

```json
// ❌ Avoid: Mixed concerns in a single universe
{
  "universes": {
    "user-and-orders": {
      "realities": {
        "USER_LOGGING_IN": {...},
        "ORDER_PROCESSING": {...},
        "EMAIL_SENDING": {...}
      }
    }
  }
}
```

### 2. Meaningful Naming Conventions

Use consistent, descriptive naming for all identifiers.

```json
{
  "id": "e-commerce-checkout",           // kebab-case for IDs
  "universes": {
    "payment-processing": {              // kebab-case for universe names
      "realities": {
        "WAITING_PAYMENT": {...},        // SCREAMING_SNAKE_CASE for realities
        "PROCESSING_PAYMENT": {...},
        "PAYMENT_COMPLETED": {...}
      }
    }
  }
}
```

**Events:**

```go
// Use domain.action pattern
event := statepro.NewEventBuilder("payment.process").Build()
event := statepro.NewEventBuilder("user.signup").Build()
event := statepro.NewEventBuilder("order.cancel").Build()
```

**Custom Executors:**

```go
// Use type:descriptiveName pattern
builtin.RegisterAction("action:sendWelcomeEmail", handler)
builtin.RegisterObserver("observer:businessHoursOnly", guard)
builtin.RegisterCondition("condition:hasPremiumAccess", checker)
```

### 3. Strategic Use of Metadata

Use metadata for debugging, tooling, and non-functional information.

```json
{
  "PROCESSING_PAYMENT": {
    "type": "transition",
    "on": {
      "payment.success": [{"targets": ["COMPLETED"]}],
      "payment.failure": [{"targets": ["FAILED"]}]
    },
    "metadata": {
      "displayName": "Processing Payment",
      "icon": "💳",
      "timeout": 30000,
      "retryPolicy": "exponential-backoff",
      "documentation": "Handles payment processing via external gateway"
    }
  }
}
```

## Event Design Patterns

### 1. Structured Event Data

Design event payloads consistently across your application.

```go
// ✅ Good: Consistent structure
type EventData struct {
    UserID       string                 `json:"userId"`
    CorrelationID string                `json:"correlationId"`
    Timestamp    time.Time              `json:"timestamp"`
    Payload      map[string]interface{} `json:"payload"`
}

event := statepro.NewEventBuilder("order.create").
    SetData(map[string]any{
        "userId":        "user-123",
        "correlationId": "req-456",
        "timestamp":     time.Now(),
        "payload": map[string]any{
            "productId": "prod-789",
            "quantity":  2,
            "price":     29.99,
        },
    }).
    SetCorrelationId("req-456").
    Build()
```

### 2. Event Versioning Strategy

Plan for event evolution from the beginning.

```go
// Version in event names
event := statepro.NewEventBuilder("order.create.v2").Build()

// Or version in event data
event := statepro.NewEventBuilder("order.create").
    SetData(map[string]any{
        "version": "2.0",
        "userId":  "123",
        // ... other fields
    }).
    Build()
```

### 3. Command vs Event Distinction

Use clear naming to distinguish between commands (requests) and events (facts).

```go
// Commands (imperatives)
statepro.NewEventBuilder("order.create").Build()     // Request to create
statepro.NewEventBuilder("payment.process").Build() // Request to process

// Events (past tense)
statepro.NewEventBuilder("order.created").Build()     // Order was created
statepro.NewEventBuilder("payment.processed").Build() // Payment was processed
```

## State Machine Architecture

### 1. Hierarchical State Organization

Break complex workflows into manageable universes.

```json
{
  "id": "user-onboarding",
  "initials": ["U:registration"],
  "universes": {
    "registration": {
      "initial": "COLLECTING_INFO",
      "realities": {
        "COLLECTING_INFO": {
          "on": {
            "info.submit": [{"targets": ["VALIDATING"]}]
          }
        },
        "VALIDATING": {
          "on": {
            "validation.success": [{"targets": ["U:email-verification"]}],
            "validation.failure": [{"targets": ["COLLECTING_INFO"]}]
          }
        }
      }
    },
    "email-verification": {
      "initial": "SENDING_EMAIL",
      "realities": {
        "SENDING_EMAIL": {
          "entry": ["action:sendVerificationEmail"],
          "on": {
            "email.sent": [{"targets": ["WAITING_VERIFICATION"]}]
          }
        },
        "WAITING_VERIFICATION": {
          "on": {
            "email.verified": [{"targets": ["U:profile-setup"]}],
            "email.timeout": [{"targets": ["SENDING_EMAIL"]}]
          }
        }
      }
    }
  }
}
```

### 2. Error Handling Patterns

Design explicit error states and recovery paths.

```json
{
  "PROCESSING": {
    "type": "transition",
    "on": {
      "success": [{"targets": ["COMPLETED"]}],
      "error": [{"targets": ["ERROR_HANDLING"]}],
      "timeout": [{"targets": ["TIMEOUT_HANDLING"]}]
    }
  },
  "ERROR_HANDLING": {
    "type": "transition",
    "entry": ["action:logError", "action:notifySupport"],
    "on": {
      "retry": [{
        "targets": ["PROCESSING"],
        "observers": ["observer:retryLimitNotExceeded"]
      }],
      "abandon": [{"targets": ["FAILED"]}]
    }
  },
  "FAILED": {
    "type": "unsuccessfulFinal",
    "entry": ["action:cleanupResources"]
  }
}
```

### 3. Timeout Handling

Implement timeout mechanisms for long-running operations.

```go
// Register timeout observer
builtin.RegisterObserver("observer:timeoutCheck", func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    data := args.GetEventData()
    startTime, ok := data["startTime"].(time.Time)
    if !ok {
        return false, fmt.Errorf("missing startTime in event data")
    }

    timeout := 30 * time.Second
    return time.Since(startTime) < timeout, nil
})

// Use in transitions
{
  "WAITING_RESPONSE": {
    "on": {
      "response": [{"targets": ["PROCESSING"]}],
      "timeout": [{
        "targets": ["TIMEOUT_HANDLING"],
        "observers": ["observer:timeoutCheck"]
      }]
    }
  }
}
```

## Performance Optimization

### 1. Efficient Observer Design

Keep guard conditions fast and predictable.

```go
// ✅ Good: Fast, simple conditions
builtin.RegisterObserver("observer:businessHours", func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    hour := time.Now().Hour()
    return hour >= 9 && hour <= 17, nil
})

// ❌ Avoid: Expensive operations in observers
builtin.RegisterObserver("observer:userExists", func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    userID := args.GetEventData()["userId"].(string)
    // Don't do expensive DB queries in observers
    user, err := database.FindUser(userID) // Slow!
    return user != nil, err
})
```

### 2. Minimize Event Data Size

Keep event payloads lean and use references when possible.

```go
// ✅ Good: Reference large objects
event := statepro.NewEventBuilder("document.process").
    SetData(map[string]any{
        "documentId": "doc-123",
        "userId":     "user-456",
    }).
    Build()

// ❌ Avoid: Including large payloads
event := statepro.NewEventBuilder("document.process").
    SetData(map[string]any{
        "documentContent": largeFileContent, // Potentially MB of data
        "userId":          "user-456",
    }).
    Build()
```

### 3. Async Operations with Invokes

Use invokes for operations that don't need to block state transitions.

```go
builtin.RegisterInvoke("invoke:notifyUsers", func(ctx context.Context, args instrumentation.InvokeExecutorArgs) error {
    // This runs asynchronously and won't block the state transition
    go func() {
        users := getInterestedUsers()
        for _, user := range users {
            sendNotification(user, args.GetEventData())
        }
    }()
    return nil
})
```

## Testing Strategies

### 1. Unit Testing Custom Executors

Test your custom actions, observers, and conditions independently.

```go
func TestProcessOrderAction(t *testing.T) {
    // Mock dependencies
    orderService := &MockOrderService{}

    // Register with test-specific name
    builtin.RegisterAction("test:processOrder", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
        return processOrder(orderService, args.GetEventData())
    })

    // Test the action logic
    eventData := map[string]any{
        "orderId": "order-123",
        "userId":  "user-456",
    }

    err := processOrder(orderService, eventData)
    assert.NoError(t, err)
    assert.True(t, orderService.ProcessOrderCalled)
}
```

### 2. Integration Testing with Bot

Use the bot pattern for repeatable integration tests.

```go
func TestUserOnboardingFlow(t *testing.T) {
    // Load test state machine
    model, err := statepro.DeserializeQuantumMachineFromFile("test_onboarding.json")
    require.NoError(t, err)

    qm, err := statepro.NewQuantumMachine(model)
    require.NoError(t, err)

    err = qm.Init(context.Background(), nil)
    require.NoError(t, err)

    // Test complete flow
    events := []string{"user.register", "email.verify", "profile.complete"}

    for _, eventName := range events {
        event := statepro.NewEventBuilder(eventName).Build()
        handled, err := qm.SendEvent(context.Background(), event)
        assert.NoError(t, err)
        assert.True(t, handled, "Event %s should be handled", eventName)
    }

    // Verify final state
    snapshot := qm.GetSnapshot()
    resume := snapshot.GetResume()

    // Should end in onboarding-complete universe
    assert.Contains(t, resume.ActiveUniverses, "onboarding-complete")
}
```

### 3. Property-Based Testing

Test state machine properties with random event sequences.

```go
func TestStateMachineInvariants(t *testing.T) {
    qm := createTestMachine(t)

    // Generate random valid event sequences
    for i := 0; i < 100; i++ {
        events := generateRandomEventSequence()

        qm.Init(context.Background(), nil)

        for _, event := range events {
            handled, err := qm.SendEvent(context.Background(), event)

            // Invariant: No processing errors
            assert.NoError(t, err, "Event processing should never error")

            // Invariant: Active universes should never be empty after init
            snapshot := qm.GetSnapshot()
            resume := snapshot.GetResume()
            assert.NotEmpty(t, resume.ActiveUniverses, "Should always have active universes")
        }
    }
}
```

## Monitoring and Observability

### 1. Structured Logging

Implement consistent logging across all custom executors.

```go
type MachineLogger struct {
    logger *slog.Logger
}

func (ml *MachineLogger) LogTransition(universe, from, to, event string) {
    ml.logger.Info("state transition",
        "universe", universe,
        "from", from,
        "to", to,
        "event", event,
    )
}

// Use in actions
builtin.RegisterAction("action:processPayment", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    logger.Info("processing payment",
        "universe", args.GetCurrentUniverse(),
        "reality", args.GetCurrentReality(),
        "event", args.GetEventName(),
        "correlationId", args.GetCorrelationId(),
    )

    // Process payment...

    return nil
})
```

### 2. Metrics Collection

Track key metrics for monitoring machine health.

```go
var (
    eventCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "statepro_events_total",
            Help: "Total number of events processed",
        },
        []string{"machine_id", "event_name", "handled"},
    )

    transitionDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "statepro_transition_duration_seconds",
            Help: "Duration of state transitions",
        },
        []string{"machine_id", "universe", "from_state", "to_state"},
    )
)

// Instrument event processing
func (w *InstrumentedMachine) SendEvent(ctx context.Context, event Event) (bool, error) {
    start := time.Now()
    handled, err := w.machine.SendEvent(ctx, event)

    eventCounter.WithLabelValues(w.machineID, event.GetEventName(), strconv.FormatBool(handled)).Inc()

    if err == nil {
        duration := time.Since(start).Seconds()
        transitionDuration.WithLabelValues(w.machineID, "universe", "from", "to").Observe(duration)
    }

    return handled, err
}
```

## Security Considerations

### 1. Input Validation

Validate all event data before processing.

```go
builtin.RegisterAction("action:updateProfile", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
    data := args.GetEventData()

    // Validate required fields
    userID, ok := data["userId"].(string)
    if !ok || userID == "" {
        return fmt.Errorf("invalid or missing userId")
    }

    // Validate format
    if !isValidUserID(userID) {
        return fmt.Errorf("malformed userId: %s", userID)
    }

    // Continue with processing...
    return nil
})
```

### 2. Authorization Checks

Implement authorization in observers when needed.

```go
builtin.RegisterObserver("observer:userCanModifyOrder", func(ctx context.Context, args instrumentation.ObserverExecutorArgs) (bool, error) {
    data := args.GetEventData()

    userID := data["userId"].(string)
    orderID := data["orderId"].(string)

    // Check if user owns the order
    order, err := orderService.GetOrder(orderID)
    if err != nil {
        return false, err
    }

    return order.UserID == userID, nil
})
```

### 3. Sensitive Data Handling

Avoid storing sensitive data in event payloads or state machine context.

```go
// ❌ Avoid: Sensitive data in events
event := statepro.NewEventBuilder("payment.process").
    SetData(map[string]any{
        "creditCardNumber": "4111-1111-1111-1111", // Don't do this!
        "cvv": "123",
    }).
    Build()

// ✅ Good: Use tokens/references
event := statepro.NewEventBuilder("payment.process").
    SetData(map[string]any{
        "paymentTokenId": "token-abc123",
        "amount": 99.99,
    }).
    Build()
```

## Documentation Standards

### 1. Machine Documentation

Document your state machines at multiple levels.

```json
{
  "id": "order-fulfillment",
  "metadata": {
    "description": "Handles order processing from confirmation to delivery",
    "version": "2.1.0",
    "author": "fulfillment-team",
    "documentation": "https://wiki.company.com/order-fulfillment"
  },
  "universes": {
    "payment-processing": {
      "metadata": {
        "description": "Processes payments via external gateway",
        "timeout": "30 seconds",
        "retryPolicy": "3 attempts with exponential backoff"
      }
    }
  }
}
```

### 2. Code Comments

Document complex business logic in your executors.

```go
// RegisterOrderProcessingActions sets up all actions needed for order processing.
// This includes payment validation, inventory checks, and shipping coordination.
func RegisterOrderProcessingActions() {
    // Validates payment information and processes charge
    // Integrates with PaymentGateway v3 API
    // Timeout: 30 seconds, Retries: 3 attempts
    builtin.RegisterAction("action:processPayment", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
        data := args.GetEventData()

        // Extract payment details
        amount, ok := data["amount"].(float64)
        if !ok {
            return fmt.Errorf("invalid amount in payment data")
        }

        // Business rule: Amounts over $10,000 require manual approval
        if amount > 10000 {
            return triggerManualApproval(data)
        }

        return processPaymentNormally(data)
    })
}
```

By following these best practices, you'll create maintainable, robust, and scalable quantum state machines that are easy to debug, test, and extend.
