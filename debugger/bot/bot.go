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
func NewBot(qm instrumentation.QuantumMachine, eventProvider EventProvider) (SMBot, error) {
	if qm == nil {
		return nil, fmt.Errorf("quantum machine cannot be nil")
	}

	if eventProvider == nil {
		return nil, fmt.Errorf("event provider cannot be nil")
	}

	return &bot{
		qm:              qm,
		initialSnapshot: qm.GetSnapshot(),
		eventProvider:   eventProvider,
	}, nil
}

type bot struct {
	qm              instrumentation.QuantumMachine
	initialSnapshot *instrumentation.MachineSnapshot
	history         []*EventHistory
	eventProvider   EventProvider
}

func (b *bot) Run(ctx context.Context, machineContext any) error {
	b.history = nil
	if err := b.qm.LoadSnapshot(b.initialSnapshot, machineContext); err != nil {
		return fmt.Errorf("error loading initial snapshot: %w", err)
	}

	if err := b.qm.Init(ctx, machineContext); err != nil {
		return fmt.Errorf("error initializing quantum machine: %w", err)
	}

	for {
		event, err := b.eventProvider(b.qm.GetSnapshot())
		if err != nil {
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
			return fmt.Errorf("event '%s' was not handled", event.GetEventName())
		}

		b.history = append(b.history, &EventHistory{
			Event:    event,
			Snapshot: b.qm.GetSnapshot(),
		})
	}

	return nil
}

func (b *bot) GetHistory() []*EventHistory {
	return b.history
}

func (b *bot) GetQuantumMachine() instrumentation.QuantumMachine {
	return b.qm
}
