package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/theoretical"
)

func TestCLI_EnviarEventoYRestaurarHistorial(t *testing.T) {
	initial := "idle"
	qm, err := statepro.NewQuantumMachine(&theoretical.QuantumMachineModel{
		ID: "machine", Initials: []string{"U:main"}, Universes: map[string]*theoretical.UniverseModel{
			"main": {ID: "main", CanonicalName: "main", Initial: &initial, Realities: map[string]*theoretical.RealityModel{
				"idle": {ID: "idle", Type: theoretical.RealityTypeTransition, On: map[string][]*theoretical.TransitionModel{"GO": {{Targets: []string{"done"}}}}},
				"done": {ID: "done", Type: theoretical.RealityTypeFinal},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	snapshot := qm.GetSnapshot()
	container := &smContainer{qm: qm, events: []*debuggerEvent{{Name: "GO", Title: "Continue", uid: uuid.New()}}, history: []*containerHistory{{snapshot: snapshot, event: getSnapshotEvent(), pos: 0}}}
	prev := &model{}
	built, _ := buildSendEventModel(prev, container)
	send := built.(*model)
	if !strings.Contains(send.View(), "Send an event") {
		t.Fatal("vista ausente")
	}
	send.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !container.events[0].Sent || len(container.history) != 2 || qm.GetSnapshot().Resume.FinalizedUniverses["main"] != "done" {
		t.Fatal("evento no registrado")
	}
	// History is independent of the saved-snapshot picker.
	historyModel, _ := buildHistoryViewerModel(prev, container)
	history := historyModel.(*model)
	if history == prev || !strings.Contains(history.View(), "View history") {
		t.Fatal("historial no disponible")
	}
	history.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if len(container.history) != 1 || container.events[0].Sent || qm.GetSnapshot().Resume.ActiveUniverses["main"] != "idle" {
		t.Fatal("rollback no restaura estado y marcas")
	}
	container.snapshots = []*debuggerSnapshot{{Title: "Checkpoint", Snapshot: snapshot}}
	loadModel, _ := buildLoadSnapshotModel(prev, container)
	load := loadModel.(*model)
	if !strings.Contains(load.View(), "Choose a snapshot") {
		t.Fatal("vista de snapshots ausente")
	}
	load.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if len(container.history) != 1 || container.history[0].pos != 0 {
		t.Fatal("carga no reinicia historial")
	}
}

func TestCLI_ListasSinSeleccionNoProvocanPanic(t *testing.T) {
	for _, kind := range []string{"events", "snapshots", "history"} {
		t.Run(kind, func(t *testing.T) {
			l := list.New(nil, list.NewDefaultDelegate(), 80, 24)
			m := &model{helperModel: &l, container: &smContainer{}}
			switch kind {
			case "events":
				m.view, m.update = sendEventModelView, sendEventModelUpdate
			case "snapshots":
				m.view, m.update = loadSnapshotModelView, loadSnapshotModelUpdate
			case "history":
				m.view, m.update = historyViewerModelView, historyViewerModelUpdate
			}
			_ = m.View()
			for _, msg := range []tea.KeyMsg{{Type: tea.KeyEnter}, {Type: tea.KeyRunes, Runes: []rune{'v'}}, {Type: tea.KeyRunes, Runes: []rune{'l'}}, {Type: tea.KeyRunes, Runes: []rune{'r'}}} {
				m.Update(msg)
			}
		})
	}
}

func TestCLI_CopiaContextoConCamposPrivadosYJSONConErrores(t *testing.T) {
	type applicationContext struct {
		Name    string
		private int
	}
	original := &applicationContext{Name: "example", private: 7}
	copy := copyStructPointer(original).(*applicationContext)
	if copy == original || !reflect.DeepEqual(copy, original) {
		t.Fatal("copia de contexto incorrecta")
	}
	copy.Name = "changed"
	if original.Name != "example" {
		t.Fatal("copia comparte valor de contexto")
	}
	var missing *applicationContext
	if copyStructPointer(missing) != nil {
		t.Fatal("nil tipado no preservado")
	}
	encoded, err := formatToJson(map[string]any{"count": json.Number("9007199254740993")}, false)
	if err != nil || !strings.Contains(string(encoded), "9007199254740993") {
		t.Fatal("formato pierde precision")
	}
	if _, err := formatToJson(make(chan int), false); err == nil {
		t.Fatal("error JSON oculto")
	}
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := loadJSON(&map[string]any{}, path, false); err == nil {
		t.Fatal("JSON invalido aceptado")
	}
	if err := loadJSON(&map[string]any{}, path+"missing", true); err != nil {
		t.Fatal(err)
	}
	if err := loadJSON(&map[string]any{}, path+"missing", false); err == nil {
		t.Fatal("archivo requerido ausente aceptado")
	}
}
