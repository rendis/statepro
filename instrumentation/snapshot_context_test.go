package instrumentation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
)

type legacySnapshots struct {
	instrumentation.QuantumMachine
	get  func() *instrumentation.MachineSnapshot
	load func(*instrumentation.MachineSnapshot, any) error
}

func (m legacySnapshots) GetSnapshot() *instrumentation.MachineSnapshot { return m.get() }
func (m legacySnapshots) LoadSnapshot(s *instrumentation.MachineSnapshot, c any) error {
	return m.load(s, c)
}

type checkedSnapshots struct {
	legacySnapshots
	checked func() (*instrumentation.MachineSnapshot, error)
}

func (m checkedSnapshots) GetSnapshotWithError() (*instrumentation.MachineSnapshot, error) {
	return m.checked()
}

type contextualSnapshots struct {
	checkedSnapshots
	capture func(context.Context) (*instrumentation.MachineSnapshot, error)
	restore func(context.Context, *instrumentation.MachineSnapshot, any) error
}

func (m contextualSnapshots) GetSnapshotContext(ctx context.Context) (*instrumentation.MachineSnapshot, error) {
	return m.capture(ctx)
}
func (m contextualSnapshots) LoadSnapshotContext(ctx context.Context, s *instrumentation.MachineSnapshot, c any) error {
	return m.restore(ctx, s, c)
}

func TestSnapshotContext_PriorizaCapacidadesYPropagaErrores(t *testing.T) {
	snapshot := &instrumentation.MachineSnapshot{}
	wantErr := errors.New("capture failed")
	legacy := legacySnapshots{get: func() *instrumentation.MachineSnapshot { return snapshot }}
	checked := checkedSnapshots{legacySnapshots: legacy, checked: func() (*instrumentation.MachineSnapshot, error) { return nil, wantErr }}
	ctx := context.WithValue(context.Background(), struct{}{}, "caller")
	contextual := contextualSnapshots{checkedSnapshots: checked, capture: func(got context.Context) (*instrumentation.MachineSnapshot, error) {
		if got != ctx {
			t.Fatal("contexto sustituido")
		}
		return snapshot, nil
	}}
	if got, err := instrumentation.GetSnapshotContext(ctx, contextual); err != nil || got != snapshot {
		t.Fatalf("captura contextual: %v", err)
	}
	if got, err := instrumentation.GetSnapshotContext(ctx, checked); got != nil || !errors.Is(err, wantErr) {
		t.Fatalf("error oculto: %v", err)
	}
	if got, err := instrumentation.GetSnapshotContext(ctx, legacy); err != nil || got != snapshot {
		t.Fatalf("compatibilidad: %v", err)
	}
	legacy.get = func() *instrumentation.MachineSnapshot { return nil }
	if _, err := instrumentation.GetSnapshotContext(ctx, legacy); err == nil {
		t.Fatal("captura nil aceptada")
	}
}

func TestSnapshotContext_CancelacionAntesYDespuesDeProveedorAntiguo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	called := false
	machine := legacySnapshots{
		get:  func() *instrumentation.MachineSnapshot { called = true; return &instrumentation.MachineSnapshot{} },
		load: func(*instrumentation.MachineSnapshot, any) error { called = true; return nil },
	}
	cancel()
	if _, err := instrumentation.GetSnapshotContext(ctx, machine); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := instrumentation.LoadSnapshotContext(ctx, machine, nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if called {
		t.Fatal("proveedor ejecutado tras cancelacion")
	}
	if _, err := instrumentation.GetSnapshotContext(nil, machine); err == nil {
		t.Fatal("contexto nil aceptado")
	}
	if err := instrumentation.LoadSnapshotContext(nil, machine, nil, nil); err == nil {
		t.Fatal("contexto nil aceptado")
	}
	ctx, cancel = context.WithCancel(context.Background())
	machine.get = func() *instrumentation.MachineSnapshot { cancel(); return &instrumentation.MachineSnapshot{} }
	if got, err := instrumentation.GetSnapshotContext(ctx, machine); got != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelacion durante captura ignorada: %v", err)
	}
}

func TestSnapshotContext_RestauraConContextoYFallbackCompatible(t *testing.T) {
	ctx := context.Background()
	snapshot := &instrumentation.MachineSnapshot{}
	wantErr := errors.New("restore failed")
	legacy := legacySnapshots{load: func(got *instrumentation.MachineSnapshot, machineContext any) error {
		if got != snapshot || machineContext != "application" {
			t.Fatal("argumentos de restauracion alterados")
		}
		return wantErr
	}}
	contextual := contextualSnapshots{checkedSnapshots: checkedSnapshots{legacySnapshots: legacy}, restore: func(got context.Context, s *instrumentation.MachineSnapshot, c any) error {
		if got != ctx || s != snapshot || c != "application" {
			t.Fatal("argumentos contextuales alterados")
		}
		return nil
	}}
	if err := instrumentation.LoadSnapshotContext(ctx, contextual, snapshot, "application"); err != nil {
		t.Fatal(err)
	}
	if err := instrumentation.LoadSnapshotContext(ctx, legacy, snapshot, "application"); !errors.Is(err, wantErr) {
		t.Fatal(err)
	}
}
