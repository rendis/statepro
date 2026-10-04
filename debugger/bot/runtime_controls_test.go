package bot_test

import (
	"context"
	"errors"
	"testing"
	"time"

	statepro "github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/debugger/bot"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

func TestBot_CancelaRestauracionMientrasMaquinaRealEjecutaCallback(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	if err := builtin.RegisterAction("test:bot-runtime-block", func(context.Context, instrumentation.ActionExecutorArgs) error {
		close(entered)
		<-release
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	defer builtin.RegisterAction("test:bot-runtime-block", nil)
	initial := "idle"
	qm, err := statepro.NewQuantumMachine(&theoretical.QuantumMachineModel{
		ID: "machine", Initials: []string{"U:main"},
		Universes: map[string]*theoretical.UniverseModel{"main": {
			ID: "main", Initial: &initial, Realities: map[string]*theoretical.RealityModel{"idle": {
				ID: "idle", Type: "transition", EntryActions: []*theoretical.ActionModel{{Src: "test:bot-runtime-block"}},
			}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := bot.NewBot(qm, func(*instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		t.Error("proveedor ejecutado tras cancelacion")
		return nil, nil
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- qm.Init(context.Background(), nil) }()
	defer func() {
		close(release)
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("callback no iniciado")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := b.Run(ctx, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestBot_CancelacionAntesYDuranteProveedor(t *testing.T) {
	for _, preCanceled := range []bool{false, true} {
		qm := &hostileQM{MockQuantumMachine: MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}}}
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		provider := func(*instrumentation.MachineSnapshot) (instrumentation.Event, error) {
			calls++
			cancel()
			return &MockEvent{name: "handled"}, nil
		}
		b, err := bot.NewBot(qm, provider, false)
		if err != nil {
			t.Fatal(err)
		}
		if preCanceled {
			cancel()
		}
		if err := b.Run(ctx, nil); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if qm.sendCalls != 0 || (preCanceled && calls != 0) {
			t.Fatal("cancelacion ejecuta trabajo adicional")
		}
	}
}

func TestBot_LimiteRetieneUltimosEventosYRechazaNegativos(t *testing.T) {
	qm := &MockQuantumMachine{snapshot: &instrumentation.MachineSnapshot{}}
	count := 0
	provider := func(*instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		count++
		if count > 5 {
			return nil, nil
		}
		return &MockEvent{name: "handled"}, nil
	}
	b, err := bot.NewBot(qm, provider, false, bot.WithHistoryLimit(2))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Run(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if len(b.GetHistory()) != 2 {
		t.Fatalf("historial sin limite: %d", len(b.GetHistory()))
	}
	if _, err := bot.NewBot(qm, provider, false, bot.WithHistoryLimit(-1)); err == nil {
		t.Fatal("limite negativo aceptado")
	}
}

type failingSnapshotMachine struct {
	MockQuantumMachine
	err error
}

func (m *failingSnapshotMachine) GetSnapshotWithError() (*instrumentation.MachineSnapshot, error) {
	return nil, m.err
}
func TestBot_NoOcultaErroresDeCaptura(t *testing.T) {
	failure := errors.New("snapshot failure")
	qm := &failingSnapshotMachine{err: failure}
	if _, err := bot.NewBot(qm, func(*instrumentation.MachineSnapshot) (instrumentation.Event, error) { return nil, nil }, false); !errors.Is(err, failure) {
		t.Fatal(err)
	}
}
