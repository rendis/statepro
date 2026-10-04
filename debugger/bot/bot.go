package bot

import (
	"context"
	"fmt"

	"github.com/rendis/statepro/v3/instrumentation"
)

// EventProvider is a function type that provides the next event to be processed based on the current snapshot.
// It returns the next event to be processed and an error if any occurs.
// If the returned event is nil, it indicates that no more events are to be processed.
type EventProvider func(currentSnapshot *instrumentation.MachineSnapshot) (instrumentation.Event, error)

// SMBot is a state machine executor.
type SMBot interface {
	// Run starts sending events to the state machine as provided by the EventProvider.
	// In each execution, the history is cleared and the state machine is set to its initial state.
	Run(ctx context.Context, machineContext any) error

	// GetHistory retrieves all the events sent and the snapshots produced by them.
	GetHistory() []*EventHistory

	// GetQuantumMachine retrieves the quantum machine used by the bot.
	GetQuantumMachine() instrumentation.QuantumMachine
}

type EventHistory struct {
	Event    instrumentation.Event
	Snapshot *instrumentation.MachineSnapshot
}

// NewBot creates a new state machine bot.
// Parameters:
// - qm: the quantum machine to be used by the bot.
// - eventProvider: the function that provides the next event to be processed. If nil, the default sequential event provider is used.
// - initQuantumMachine: if true, the quantum machine will be initialized before processing events.
// - opts: optional configurations for the bot.
func NewBot(
	qm instrumentation.QuantumMachine,
	eventProvider EventProvider,
	initQuantumMachine bool,
	opts ...BotOption,
) (SMBot, error) {
	if qm == nil {
		return nil, fmt.Errorf("quantum machine cannot be nil")
	}

	if eventProvider == nil {
		return nil, fmt.Errorf("event provider cannot be nil")
	}

	initial, err := captureSnapshot(context.Background(), qm)
	if err != nil {
		return nil, fmt.Errorf("error capturing initial snapshot: %w", err)
	}
	b := &bot{
		qm:                 qm,
		initialSnapshot:    initial,
		eventProvider:      eventProvider,
		initQuantumMachine: initQuantumMachine,
	}

	for _, opt := range opts {
		opt(b)
	}
	if b.historyLimit < 0 {
		return nil, fmt.Errorf("history limit must not be negative")
	}
	return b, nil
}

type bot struct {
	qm                    instrumentation.QuantumMachine
	initialSnapshot       *instrumentation.MachineSnapshot
	eventProvider         EventProvider
	initQuantumMachine    bool
	history               []*EventHistory
	ignoreUnhandledEvents bool
	historyLimit          int
}

type BotOption func(*bot)

func WithIgnoreUnhandledEvents(ignore bool) BotOption {
	return func(b *bot) {
		b.ignoreUnhandledEvents = ignore
	}
}

// WithHistoryLimit retains the latest entries; zero preserves unlimited history.
func WithHistoryLimit(limit int) BotOption {
	return func(b *bot) { b.historyLimit = limit }
}

func captureSnapshot(ctx context.Context, qm instrumentation.QuantumMachine) (*instrumentation.MachineSnapshot, error) {
	var snapshot *instrumentation.MachineSnapshot
	var err error
	if contextual, ok := qm.(instrumentation.ContextSnapshotProvider); ok {
		snapshot, err = contextual.GetSnapshotContext(ctx)
	} else if checked, ok := qm.(instrumentation.SnapshotProvider); ok {
		snapshot, err = checked.GetSnapshotWithError()
	} else {
		snapshot = qm.GetSnapshot()
	}
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, fmt.Errorf("snapshot capture returned nil")
	}
	return snapshot, nil
}

func (b *bot) Run(ctx context.Context, machineContext any) error {
	if ctx == nil {
		return fmt.Errorf("context must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	b.history = nil
	var restoreErr error
	if contextual, ok := b.qm.(instrumentation.ContextSnapshotProvider); ok {
		restoreErr = contextual.LoadSnapshotContext(ctx, b.initialSnapshot, machineContext)
	} else {
		restoreErr = b.qm.LoadSnapshot(b.initialSnapshot, machineContext)
	}
	if restoreErr != nil {
		return fmt.Errorf("error loading initial snapshot: %w", restoreErr)
	}

	if b.initQuantumMachine {
		if err := b.qm.Init(ctx, machineContext); err != nil {
			return fmt.Errorf("error initializing quantum machine: %w", err)
		}
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		snapshot, err := captureSnapshot(ctx, b.qm)
		if err != nil {
			return fmt.Errorf("error capturing snapshot: %w", err)
		}
		event, err := b.eventProvider(snapshot)
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if event == nil {
			break
		}

		handled, err := b.qm.SendEvent(ctx, event)
		if err != nil {
			return err
		}

		if !handled {
			if b.ignoreUnhandledEvents {
				continue
			}
			return fmt.Errorf("event '%s' was not handled", event.GetEventName())
		}

		snapshot, err = captureSnapshot(ctx, b.qm)
		if err != nil {
			return fmt.Errorf("error capturing event snapshot: %w", err)
		}
		b.history = append(b.history, &EventHistory{
			Event:    event,
			Snapshot: snapshot,
		})
		if b.historyLimit > 0 && len(b.history) > b.historyLimit {
			copy(b.history, b.history[len(b.history)-b.historyLimit:])
			clear(b.history[b.historyLimit:])
			b.history = b.history[:b.historyLimit]
		}
	}

	return nil
}

func (b *bot) GetHistory() []*EventHistory {
	return b.history
}

func (b *bot) GetQuantumMachine() instrumentation.QuantumMachine {
	return b.qm
}
