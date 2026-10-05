package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/instrumentation"
)

const cliDefinition = `{"id":"machine","canonicalName":"machine","version":"1.0.0","initials":["U:main"],"universes":{"main":{"id":"main","canonicalName":"main","version":"1.0.0","initial":"idle","realities":{"idle":{"id":"idle","type":"final"}}}}}`

func writeCLIFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCLI_ConstruyeHistorialInicialYLeeArchivosOpcionales(t *testing.T) {
	dir := t.TempDir()
	application := &struct{ Name string }{}
	opt := &debuggerOptions{
		stateMachinePath: writeCLIFile(t, dir, "machine.json", cliDefinition),
		smContextPath:    writeCLIFile(t, dir, "context.json", `{"Name":"application"}`),
		smContext:        application,
		eventsPath:       writeCLIFile(t, dir, "events.json", `[{"name":"GO","title":"Continue"}]`),
		snapshotsPath:    writeCLIFile(t, dir, "snapshots.json", `[{"snapshot":{}}]`),
	}
	container, err := buildContainer(opt)
	if err != nil {
		t.Fatal(err)
	}
	if application.Name != "application" || container.smContext != application || len(container.history) != 1 || container.history[0].snapshot == nil || container.history[0].snapshot.Resume.FinalizedUniverses["main"] != "idle" {
		t.Fatal("estado inicial o contexto incorrectos")
	}
	if len(container.events) != 1 || container.events[0].uid == uuid.Nil || len(container.snapshots) != 2 || container.snapshots[0].Title != "Reset state machine (default)" || container.snapshots[1].Title != "Snapshot 2" {
		t.Fatal("archivos, identidades o titulos incorrectos")
	}
	opt.eventsPath, opt.snapshotsPath, opt.smContextPath = "", "", ""
	if _, err := buildContainer(opt); err != nil {
		t.Fatalf("archivos opcionales ausentes: %v", err)
	}
}

func TestCLI_ArchivosInvalidosDevuelvenErroresSinPanic(t *testing.T) {
	for _, kind := range []string{"context", "machine", "events", "snapshots", "null-event", "null-snapshot", "missing-machine"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			opt := &debuggerOptions{stateMachinePath: writeCLIFile(t, dir, "machine.json", cliDefinition), smContext: &struct{ Name string }{}}
			invalid := writeCLIFile(t, dir, "invalid.json", "{")
			want := ""
			switch kind {
			case "context":
				opt.smContextPath = invalid
				want = "error loading sm context"
			case "machine":
				opt.stateMachinePath = invalid
				want = "error loading definition"
			case "events":
				opt.eventsPath = invalid
				want = "error loading events"
			case "snapshots":
				opt.snapshotsPath = invalid
				want = "error loading snapshots"
			case "null-event":
				opt.eventsPath = writeCLIFile(t, dir, "events.json", "[null]")
				want = "event 1 must not be null"
			case "null-snapshot":
				opt.snapshotsPath = writeCLIFile(t, dir, "snapshots.json", "[null]")
				want = "snapshot 1 must not be null"
			case "missing-machine":
				opt.stateMachinePath += "missing"
				want = "error reading file"
			}
			if container, err := buildContainer(opt); container != nil || err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("error esperado %q: %v", want, err)
			}
		})
	}
}

func TestCLI_CapturaInicialFallidaSePropaga(t *testing.T) {
	if err := builtin.RegisterAction("test:cli-invalid-metadata", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.AddToUniverseMetadata("unserializable", make(chan int))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	definition := strings.Replace(cliDefinition, `"type":"final"`, `"type":"final","entryActions":[{"src":"test:cli-invalid-metadata"}]`, 1)
	opt := &debuggerOptions{stateMachinePath: writeCLIFile(t, t.TempDir(), "machine.json", definition)}
	if container, err := buildContainer(opt); container != nil || err == nil || !strings.Contains(err.Error(), "error capturing initial snapshot") {
		t.Fatalf("captura inicial fallida oculta: %v", err)
	}
}
