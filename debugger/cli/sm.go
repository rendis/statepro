package cli

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rendis/statepro"
	"github.com/rendis/statepro/instrumentation"
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
		container.history = []*containerHistory{
			{
				snapshot: container.qm.GetSnapshot(),
				context:  copyStructPointer(container.smContext),
				event:    getSnapshotEvent(),
				pos:      0,
			},
		}
	}

	if err = loadJSON(&container.events, opt.eventsPath, true); err != nil {
		return nil, fmt.Errorf("error loading events: %w", err)
	}
	for _, event := range container.events {
		event.uid = uuid.New()
	}

	if err = loadJSON(&container.snapshots, opt.snapshotsPath, true); err != nil {
		return nil, fmt.Errorf("error loading snapshots: %w", err)
	}

	setDefaultSnapshot(&container)
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

func setDefaultSnapshot(container *smContainer) {
	if container.qm == nil {
		return
	}

	first := container.qm.GetSnapshot()
	debuggerSnap := &debuggerSnapshot{
		Title:    "Reset state machine (default)",
		Snapshot: first,
	}

	container.snapshots = append([]*debuggerSnapshot{debuggerSnap}, container.snapshots...)
}

func setSnapshotsTitle(container *smContainer) {
	for i, snap := range container.snapshots {
		if snap.Title == "" {
			snap.Title = fmt.Sprintf("Snapshot %d", i+1)
		}
	}
}
