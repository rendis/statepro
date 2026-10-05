package cli

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/instrumentation"
	"os"
)

type debuggerEvent struct {
	Title  string         `json:"title"`
	Name   string         `json:"name"`
	Params map[string]any `json:"params"`
	Sent   bool           `json:"sent"`
	uid    uuid.UUID
}

type debuggerSnapshot struct {
	Title    string                           `json:"title"`
	Snapshot *instrumentation.MachineSnapshot `json:"snapshot"`
}

type containerHistory struct {
	event    *debuggerEvent
	snapshot *instrumentation.MachineSnapshot
	context  any
	pos      int
}

type smContainer struct {
	smContext any
	qm        instrumentation.QuantumMachine
	events    []*debuggerEvent
	snapshots []*debuggerSnapshot
	history   []*containerHistory
}

func buildContainer(opt *debuggerOptions) (*smContainer, error) {
	var container smContainer
	var err error

	if err = loadJSON(opt.smContext, opt.smContextPath, true); err != nil {
		return nil, fmt.Errorf("error loading sm context: %w", err)
	}

	container.smContext = opt.smContext
	container.qm, err = loadDefinition(opt.stateMachinePath, opt.smContext)
	if err != nil {
		return nil, fmt.Errorf("error loading definition: %w", err)
	}

	if container.qm != nil {
		snapshot, captureErr := instrumentation.GetSnapshotContext(context.Background(), container.qm)
		if captureErr != nil {
			return nil, fmt.Errorf("error capturing initial snapshot: %w", captureErr)
		}
		container.history = []*containerHistory{
			{
				snapshot: snapshot,
				context:  copyStructPointer(container.smContext),
				event:    getSnapshotEvent(),
				pos:      0,
			},
		}
	}

	if err = loadJSON(&container.events, opt.eventsPath, true); err != nil {
		return nil, fmt.Errorf("error loading events: %w", err)
	}
	for i, event := range container.events {
		if event == nil {
			return nil, fmt.Errorf("event %d must not be null", i+1)
		}
		event.uid = uuid.New()
	}

	if err = loadJSON(&container.snapshots, opt.snapshotsPath, true); err != nil {
		return nil, fmt.Errorf("error loading snapshots: %w", err)
	}
	for i, snapshot := range container.snapshots {
		if snapshot == nil {
			return nil, fmt.Errorf("snapshot %d must not be null", i+1)
		}
	}

	if err := setDefaultSnapshot(&container); err != nil {
		return nil, err
	}
	setSnapshotsTitle(&container)
	return &container, nil
}

func loadDefinition(path string, smContext any) (instrumentation.QuantumMachine, error) {
	arrByte, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	tmDef, err := statepro.DeserializeQuantumMachineFromBinary(arrByte)
	if err != nil {
		return nil, fmt.Errorf("error deserializing quantum machine: %w", err)
	}

	qm, err := statepro.NewQuantumMachine(tmDef)
	if err != nil {
		return nil, fmt.Errorf("error creating quantum machine: %w", err)
	}

	if err = qm.Init(context.Background(), smContext); err != nil {
		return nil, fmt.Errorf("error initializing quantum machine: %w", err)
	}

	return qm, nil
}

func setDefaultSnapshot(container *smContainer) error {
	if container.qm == nil {
		return nil
	}

	first, err := instrumentation.GetSnapshotContext(context.Background(), container.qm)
	if err != nil {
		return fmt.Errorf("error capturing default snapshot: %w", err)
	}
	debuggerSnap := &debuggerSnapshot{
		Title:    "Reset state machine (default)",
		Snapshot: first,
	}

	container.snapshots = append([]*debuggerSnapshot{debuggerSnap}, container.snapshots...)
	return nil
}

func setSnapshotsTitle(container *smContainer) {
	for i, snap := range container.snapshots {
		if snap.Title == "" {
			snap.Title = fmt.Sprintf("Snapshot %d", i+1)
		}
	}
}
