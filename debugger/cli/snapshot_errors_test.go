package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rendis/statepro/v3/instrumentation"
)

type failingSnapshots struct {
	instrumentation.QuantumMachine
	captureErr error
	restoreErr error
}

func (m *failingSnapshots) SendEvent(context.Context, instrumentation.Event) (bool, error) {
	return true, nil
}
func (m *failingSnapshots) GetSnapshotContext(context.Context) (*instrumentation.MachineSnapshot, error) {
	return nil, m.captureErr
}
func (m *failingSnapshots) LoadSnapshotContext(context.Context, *instrumentation.MachineSnapshot, any) error {
	return m.restoreErr
}

func TestCLI_ErrorDeCapturaNoCreaCheckpointYEventoSigueMarcado(t *testing.T) {
	wantErr := errors.New("metadata cannot be serialized")
	checkpoint := &containerHistory{snapshot: &instrumentation.MachineSnapshot{}, event: getSnapshotEvent()}
	container := &smContainer{qm: &failingSnapshots{captureErr: wantErr}, events: []*debuggerEvent{{Name: "GO"}}, history: []*containerHistory{checkpoint}}
	built, _ := buildSendEventModel(&model{}, container)
	m := built.(*model)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !errors.Is(m.err, wantErr) || !strings.Contains(m.View(), wantErr.Error()) {
		t.Fatal("error de captura oculto")
	}
	if !container.events[0].Sent || len(container.history) != 1 || container.history[0] != checkpoint {
		t.Fatal("evento o historial incorrectos tras captura fallida")
	}
	if err := setDefaultSnapshot(container); !errors.Is(err, wantErr) || len(container.snapshots) != 0 {
		t.Fatal("checkpoint inicial invalido creado")
	}
}

func TestCLI_RestauracionFallidaConservaHistorialYMarcas(t *testing.T) {
	wantErr := errors.New("invalid snapshot")
	for _, kind := range []string{"snapshot", "history"} {
		t.Run(kind, func(t *testing.T) {
			checkpoint := &containerHistory{snapshot: &instrumentation.MachineSnapshot{}, event: getSnapshotEvent()}
			container := &smContainer{qm: &failingSnapshots{restoreErr: wantErr}, events: []*debuggerEvent{{Name: "GO", Sent: true}}, history: []*containerHistory{checkpoint}, snapshots: []*debuggerSnapshot{{Title: "checkpoint", Snapshot: checkpoint.snapshot}}}
			var built tea.Model
			var key rune
			if kind == "snapshot" {
				built, _ = buildLoadSnapshotModel(&model{}, container)
				key = 'l'
			} else {
				built, _ = buildHistoryViewerModel(&model{}, container)
				key = 'r'
			}
			m := built.(*model)
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
			if !errors.Is(m.err, wantErr) || !strings.Contains(m.View(), wantErr.Error()) {
				t.Fatal("error de restauracion oculto")
			}
			if len(container.history) != 1 || container.history[0] != checkpoint || !container.events[0].Sent {
				t.Fatal("historial modificado tras restauracion rechazada")
			}
		})
	}
}

func TestCLI_CapturaFallidaTrasRestaurarNoInventaHistorial(t *testing.T) {
	wantErr := errors.New("capture failed")
	checkpoint := &containerHistory{snapshot: &instrumentation.MachineSnapshot{}, event: getSnapshotEvent()}
	container := &smContainer{qm: &failingSnapshots{captureErr: wantErr}, history: []*containerHistory{checkpoint}, snapshots: []*debuggerSnapshot{{Title: "checkpoint", Snapshot: checkpoint.snapshot}}}
	built, _ := buildLoadSnapshotModel(&model{}, container)
	m := built.(*model)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if !errors.Is(m.err, wantErr) || container.history[0] != checkpoint {
		t.Fatal("captura fallida oculta o checkpoint sustituido")
	}
}
