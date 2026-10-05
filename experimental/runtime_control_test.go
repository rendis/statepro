package experimental

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

func buildControlledQM(t *testing.T, options instrumentation.RuntimeOptions, realities map[string]*theoretical.RealityModel) (*ExQuantumMachine, *ExUniverse) {
	t.Helper()
	base, u := buildQM(t, "stateA", realities)
	machine, err := NewExQuantumMachineWithOptions(base.model, []*ExUniverse{u}, options)
	if err != nil {
		t.Fatal(err)
	}
	qm := machine.(*ExQuantumMachine)
	t.Cleanup(func() {
		_ = qm.Close()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := qm.WaitInvokes(ctx); err != nil {
			t.Errorf("invokes pendientes: %v", err)
		}
	})
	return qm, u
}

func awaitRuntimeSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatal("operacion bloqueada")
	}
}

func TestRuntime_CancelaEsperaDelLockSinEsperarAlCallback(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	registerTestAction(t, "test:runtime-lock", func(context.Context, instrumentation.ActionExecutorArgs) error { close(entered); <-release; return nil })
	qm, _ := buildControlledQM(t, instrumentation.RuntimeOptions{}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryAction("test:runtime-lock"))})
	done := make(chan error, 1)
	go func() { done <- qm.Init(context.Background(), nil) }()
	awaitRuntimeSignal(t, entered)
	ctx, cancel := context.WithCancel(context.Background())
	waiting := make(chan error, 1)
	go func() { _, err := qm.SendEvent(ctx, NewEventBuilder("GO").Build()); waiting <- err }()
	cancel()
	select {
	case err := <-waiting:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		close(release)
		t.Fatal("cancelacion espera el lock")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := qm.GetSnapshotContext(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRuntime_ReentradaDetectadaYContextoReutilizableTrasCallback(t *testing.T) {
	var qm *ExQuantumMachine
	var saved context.Context
	registerTestAction(t, "test:runtime-reentry", func(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
		saved = ctx
		if _, err := qm.SendEvent(ctx, NewEventBuilder("GO").Build()); !errors.Is(err, instrumentation.ErrReentrantCall) {
			return errors.New("reentrada no rechazada")
		}
		if _, err := qm.GetSnapshotContext(ctx); !errors.Is(err, instrumentation.ErrReentrantCall) {
			return errors.New("snapshot reentrante no rechazado")
		}
		if err := qm.LoadSnapshotContext(ctx, nil, nil); !errors.Is(err, instrumentation.ErrReentrantCall) {
			return errors.New("restauracion reentrante no rechazada")
		}
		if _, err := instrumentation.GetSnapshotContext(ctx, qm); !errors.Is(err, instrumentation.ErrReentrantCall) {
			return errors.New("helper de captura oculta reentrada")
		}
		if err := instrumentation.LoadSnapshotContext(ctx, qm, nil, nil); !errors.Is(err, instrumentation.ErrReentrantCall) {
			return errors.New("helper de restauracion oculta reentrada")
		}
		if err := qm.WaitInvokes(ctx); !errors.Is(err, instrumentation.ErrReentrantCall) {
			return errors.New("espera reentrante no rechazada")
		}
		if _, err := instrumentation.GetSnapshotWithError(args); err != nil {
			return err
		}
		return nil
	})
	qm, _ = buildControlledQM(t, instrumentation.RuntimeOptions{}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryAction("test:runtime-reentry"))})
	done := make(chan error, 1)
	go func() { done <- qm.Init(context.Background(), nil) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("reentrada bloqueada")
	}
	if _, err := qm.GetSnapshotContext(saved); err != nil {
		t.Fatal(err)
	}
}

func TestRuntime_CancelacionDuranteAccionNoInterrumpeOperacionAdmitida(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var secondSawCancel atomic.Bool
	registerTestAction(t, "test:runtime-cancel-action", func(context.Context, instrumentation.ActionExecutorArgs) error { cancel(); return nil })
	registerTestAction(t, "test:runtime-second-action", func(ctx context.Context, _ instrumentation.ActionExecutorArgs) error {
		secondSawCancel.Store(ctx.Err() != nil)
		return nil
	})
	qm, u := buildControlledQM(t, instrumentation.RuntimeOptions{}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryAction("test:runtime-cancel-action"), withEntryAction("test:runtime-second-action"))})
	if err := qm.Init(ctx, nil); err != nil {
		t.Fatalf("operacion admitida abortada por cancelacion: %v", err)
	}
	if !secondSawCancel.Load() || !u.initialized || !u.realityInitialized {
		t.Fatal("el callback siguiente debe ejecutarse y observar la cancelacion")
	}
}

func TestRuntime_CancelacionTrasAccionDeTransicionCompletaLaTransicion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var sideEffects atomic.Int32
	registerTestAction(t, "test:runtime-cancel-transition", func(context.Context, instrumentation.ActionExecutorArgs) error {
		sideEffects.Add(1)
		cancel()
		return nil
	})
	stateA := newTransitionReality("stateA", withOnTransition("GO", []string{"stateB"}, nil))
	stateA.On["GO"][0].Actions = []*theoretical.ActionModel{{Src: "test:runtime-cancel-transition"}}
	qm, u := buildControlledQM(t, instrumentation.RuntimeOptions{}, map[string]*theoretical.RealityModel{"stateA": stateA, "stateB": newTransitionReality("stateB")})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if handled, err := qm.SendEvent(ctx, NewEventBuilder("GO").Build()); !handled || err != nil {
		t.Fatalf("transicion con efectos ejecutados no debe abortarse: %v, %v", handled, err)
	}
	if *u.currentReality != "stateB" || sideEffects.Load() != 1 {
		t.Fatalf("estado inconsistente: %s, efectos %d", *u.currentReality, sideEffects.Load())
	}
}

func TestRuntime_ReentradaEntreCallbacksDeDosMaquinas(t *testing.T) {
	var first, second *ExQuantumMachine
	registerTestAction(t, "test:runtime-nested-first", func(ctx context.Context, _ instrumentation.ActionExecutorArgs) error { return second.Init(ctx, nil) })
	registerTestAction(t, "test:runtime-nested-second", func(ctx context.Context, _ instrumentation.ActionExecutorArgs) error {
		_, err := first.SendEvent(ctx, NewEventBuilder("GO").Build())
		if !errors.Is(err, instrumentation.ErrReentrantCall) {
			return errors.New("reentrada anidada no rechazada")
		}
		return nil
	})
	first, _ = buildControlledQM(t, instrumentation.RuntimeOptions{}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryAction("test:runtime-nested-first"))})
	second, _ = buildControlledQM(t, instrumentation.RuntimeOptions{}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryAction("test:runtime-nested-second"))})
	done := make(chan error, 1)
	go func() { done <- first.Init(context.Background(), nil) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("reentrada anidada bloqueada")
	}
}

func TestRuntime_InvokeAsincronoNoHeredaReentradaDelCallback(t *testing.T) {
	var first, second *ExQuantumMachine
	started := make(chan struct{})
	result := make(chan error, 1)
	registerTestAction(t, "test:runtime-async-parent", func(ctx context.Context, _ instrumentation.ActionExecutorArgs) error {
		if err := second.Init(ctx, nil); err != nil {
			return err
		}
		awaitRuntimeSignal(t, started)
		return nil
	})
	registerTestInvoke(t, "test:runtime-async-child", func(ctx context.Context, _ instrumentation.InvokeExecutorArgs) {
		err := first.checkCallbackContext(ctx)
		close(started)
		if err == nil {
			_, err = first.SendEvent(ctx, NewEventBuilder("GO").Build())
		}
		result <- err
	})
	first, _ = buildControlledQM(t, instrumentation.RuntimeOptions{}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryAction("test:runtime-async-parent"))})
	second, _ = buildControlledQM(t, instrumentation.RuntimeOptions{}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryInvoke("test:runtime-async-child"))})
	if err := first.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("invoke asincrono bloqueado")
	}
}

func TestRuntime_LimiteCompartidoEntreUniversos(t *testing.T) {
	started := make(chan struct{})
	var rejected atomic.Bool
	registerTestInvoke(t, "test:runtime-universe-one", func(ctx context.Context, _ instrumentation.InvokeExecutorArgs) { close(started); <-ctx.Done() })
	registerTestInvoke(t, "test:runtime-universe-two", func(context.Context, instrumentation.InvokeExecutorArgs) { rejected.Store(true) })
	base, first, second := buildMultiUniverseQM(t,
		"stateA", map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryInvoke("test:runtime-universe-one"))},
		"stateA", map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryInvoke("test:runtime-universe-two"))},
	)
	configured, err := NewExQuantumMachineWithOptions(base.model, []*ExUniverse{first, second}, instrumentation.RuntimeOptions{MaxConcurrentInvokes: 1})
	if err != nil {
		t.Fatal(err)
	}
	qm := configured.(*ExQuantumMachine)
	defer qm.Close()
	var limit *instrumentation.ResourceLimitError
	if err := qm.Init(context.Background(), nil); !errors.As(err, &limit) {
		t.Fatalf("limite por maquina omitido: %v", err)
	}
	awaitRuntimeSignal(t, started)
	if rejected.Load() {
		t.Fatal("segundo universo excede limite")
	}
}

func TestRuntime_PanicLiberaSlotYPosicionCancelaInvokes(t *testing.T) {
	registerTestInvoke(t, "test:runtime-pool-panic", func(context.Context, instrumentation.InvokeExecutorArgs) { panic("intentional pool panic") })
	qm, _ := buildControlledQM(t, instrumentation.RuntimeOptions{MaxConcurrentInvokes: 1}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryInvoke("test:runtime-pool-panic"))})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := qm.WaitInvokes(ctx); err != nil {
		t.Fatal(err)
	}
	if err := qm.ReplayOnEntry(ctx); err != nil {
		t.Fatal(err)
	}
	if err := qm.WaitInvokes(ctx); err != nil {
		t.Fatal(err)
	}
	started, ended := make(chan struct{}), make(chan struct{})
	registerTestInvoke(t, "test:runtime-position-cancel", func(ctx context.Context, _ instrumentation.InvokeExecutorArgs) {
		close(started)
		<-ctx.Done()
		close(ended)
	})
	positioned, _ := buildControlledQM(t, instrumentation.RuntimeOptions{CancelInvokesOnExit: true}, map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryInvoke("test:runtime-position-cancel")), "stateB": newFinalReality("stateB"),
	})
	if err := positioned.Init(ctx, nil); err != nil {
		t.Fatal(err)
	}
	awaitRuntimeSignal(t, started)
	if err := positioned.PositionMachine(ctx, nil, "u1", "stateB", false); err != nil {
		t.Fatal(err)
	}
	awaitRuntimeSignal(t, ended)
}

func TestRuntime_InvokesRespetanSalidaOptInYCierre(t *testing.T) {
	for _, cancelOnExit := range []bool{false, true} {
		t.Run(map[bool]string{false: "contrato_legacy", true: "cancelacion_opt_in"}[cancelOnExit], func(t *testing.T) {
			started, ended := make(chan struct{}), make(chan struct{})
			name := "test:runtime-exit-legacy"
			if cancelOnExit {
				name = "test:runtime-exit-controlled"
			}
			registerTestInvoke(t, name, func(ctx context.Context, _ instrumentation.InvokeExecutorArgs) {
				close(started)
				<-ctx.Done()
				close(ended)
			})
			qm, _ := buildControlledQM(t, instrumentation.RuntimeOptions{CancelInvokesOnExit: cancelOnExit}, map[string]*theoretical.RealityModel{
				"stateA": newTransitionReality("stateA", withEntryInvoke(name), withOnTransition("GO", []string{"stateB"}, nil)), "stateB": newFinalReality("stateB"),
			})
			if err := qm.Init(context.Background(), nil); err != nil {
				t.Fatal(err)
			}
			awaitRuntimeSignal(t, started)
			if _, err := qm.SendEvent(context.Background(), NewEventBuilder("GO").Build()); err != nil {
				t.Fatal(err)
			}
			if cancelOnExit {
				awaitRuntimeSignal(t, ended)
			} else {
				select {
				case <-ended:
					t.Fatal("default cancela invoke al salir")
				default:
				}
			}
			_ = qm.Close()
			awaitRuntimeSignal(t, ended)
			if err := qm.ReplayOnEntry(context.Background()); !errors.Is(err, instrumentation.ErrMachineClosed) {
				t.Fatal(err)
			}
		})
	}
}

func TestRuntime_LimiteGlobalCubreInvokesDeConstantesYRealidad(t *testing.T) {
	var launched, entryActions atomic.Int32
	registerTestInvoke(t, "test:runtime-limited-constant", func(context.Context, instrumentation.InvokeExecutorArgs) { launched.Add(1) })
	registerTestInvoke(t, "test:runtime-limited-reality", func(context.Context, instrumentation.InvokeExecutorArgs) { launched.Add(1) })
	registerTestAction(t, "test:runtime-limited-entry", func(context.Context, instrumentation.ActionExecutorArgs) error { entryActions.Add(1); return nil })
	qm, u := buildControlledQM(t, instrumentation.RuntimeOptions{MaxConcurrentInvokes: 1}, map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withEntryAction("test:runtime-limited-entry"), withEntryInvoke("test:runtime-limited-reality")),
	})
	qm.model.UniversalConstants = &theoretical.UniversalConstantsModel{EntryInvokes: []*theoretical.InvokeModel{{Src: "test:runtime-limited-constant"}}}
	var limit *instrumentation.ResourceLimitError
	if err := qm.Init(context.Background(), nil); !errors.As(err, &limit) || limit.Resource != "concurrent invokes" {
		t.Fatalf("limite omitido: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := qm.WaitInvokes(ctx); err != nil {
		t.Fatal(err)
	}
	if launched.Load() != 0 || entryActions.Load() != 0 || u.realityInitialized {
		t.Fatalf("rechazo tras efectos: invokes %d, acciones %d", launched.Load(), entryActions.Load())
	}
}

func TestRuntime_LimiteDeInvokesRechazaTransicionAntesDeEfectos(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	var transitionActions atomic.Int32
	registerTestInvoke(t, "test:runtime-step-invoke", func(ctx context.Context, _ instrumentation.InvokeExecutorArgs) {
		select {
		case <-block:
		case <-ctx.Done():
		}
	})
	registerTestAction(t, "test:runtime-step-action", func(context.Context, instrumentation.ActionExecutorArgs) error { transitionActions.Add(1); return nil })
	stateA := newTransitionReality("stateA", withOnTransition("GO", []string{"stateB"}, nil), withOnTransition("FIT", []string{"stateC"}, nil))
	stateA.On["GO"][0].Actions = []*theoretical.ActionModel{{Src: "test:runtime-step-action"}}
	stateA.On["GO"][0].Invokes = []*theoretical.InvokeModel{{Src: "test:runtime-step-invoke"}}
	stateA.ExitInvokes = []*theoretical.InvokeModel{{Src: "test:runtime-step-invoke"}}
	stateB := newTransitionReality("stateB", withEntryInvoke("test:runtime-step-invoke"))
	stateC := newTransitionReality("stateC", withEntryInvoke("test:runtime-step-invoke"))
	qm, u := buildControlledQM(t, instrumentation.RuntimeOptions{MaxConcurrentInvokes: 2}, map[string]*theoretical.RealityModel{"stateA": stateA, "stateB": stateB, "stateC": stateC})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	// Transition, exit and target entry need three slots; only two exist.
	var limit *instrumentation.ResourceLimitError
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("GO").Build()); !errors.As(err, &limit) {
		t.Fatalf("limite omitido: %v", err)
	}
	if transitionActions.Load() != 0 || qm.invokes.running() != 0 || *u.currentReality != "stateA" || !u.realityInitialized {
		t.Fatalf("rechazo parcial: acciones %d, invokes %d, realidad %s", transitionActions.Load(), qm.invokes.running(), *u.currentReality)
	}
	// Exit and target entry fit exactly.
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("FIT").Build()); err != nil {
		t.Fatal(err)
	}
	if *u.currentReality != "stateC" || qm.invokes.running() != 2 {
		t.Fatalf("transicion dentro del limite: %s, invokes %d", *u.currentReality, qm.invokes.running())
	}
}

func TestRuntime_RestauracionValidaCancelaInvokesPeroInvalidaNo(t *testing.T) {
	started, ended := make(chan struct{}), make(chan struct{})
	registerTestInvoke(t, "test:runtime-restore", func(ctx context.Context, _ instrumentation.InvokeExecutorArgs) {
		close(started)
		<-ctx.Done()
		close(ended)
	})
	qm, _ := buildControlledQM(t, instrumentation.RuntimeOptions{CancelInvokesOnExit: true}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryInvoke("test:runtime-restore"))})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	awaitRuntimeSignal(t, started)
	snapshot, err := qm.GetSnapshotWithError()
	if err != nil {
		t.Fatal(err)
	}
	if err := qm.LoadSnapshot(&instrumentation.MachineSnapshot{Snapshots: map[string]instrumentation.SerializedUniverseSnapshot{"u1": nil}}, nil); err == nil {
		t.Fatal("snapshot invalido aceptado")
	}
	select {
	case <-ended:
		t.Fatal("rechazar snapshot cancela invokes")
	default:
	}
	if err := qm.LoadSnapshot(snapshot, nil); err != nil {
		t.Fatal(err)
	}
	awaitRuntimeSignal(t, ended)
}

func TestRuntime_ObserverEstrictoYLimiteDeAcumulacionSinMutacionParcial(t *testing.T) {
	registerTestObserver(t, "test:runtime-decline", func(context.Context, instrumentation.ObserverExecutorArgs) (bool, error) { return false, nil })
	for _, strict := range []bool{false, true} {
		qm, u := buildControlledQM(t, instrumentation.RuntimeOptions{StrictObservers: strict}, map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withObserver("test:runtime-missing", nil))})
		u.initOnSuperposition()
		_, err := qm.SendEvent(context.Background(), NewEventBuilder("GO").Build())
		if strict && (!errors.Is(err, instrumentation.ErrUnknownObserver) || !u.inSuperposition || u.eventAccumulator.GetStatistics().CountAllEvents() != 0) {
			t.Fatalf("modo estricto no rechazo sin acumular: %v", err)
		}
		if !strict && (err != nil || u.inSuperposition) {
			t.Fatalf("default alterado: %v", err)
		}
	}
	qm, u := buildControlledQM(t, instrumentation.RuntimeOptions{MaxAccumulatedEvents: 3}, map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withObserver("test:runtime-decline", nil)), "stateB": newTransitionReality("stateB", withObserver("test:runtime-decline", nil)),
	})
	u.initOnSuperposition()
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("GO").Build()); err != nil {
		t.Fatal(err)
	}
	var limit *instrumentation.ResourceLimitError
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("GO").Build()); !errors.As(err, &limit) {
		t.Fatal(err)
	}
	if u.eventAccumulator.GetStatistics().CountAllEvents() != 2 {
		t.Fatal("fanout rechazado muta acumulador")
	}
	snapshot, err := qm.GetSnapshotWithError()
	if err != nil {
		t.Fatal(err)
	}
	qm.options.MaxAccumulatedEvents = 1
	u.options.MaxAccumulatedEvents = 1
	if err := qm.LoadSnapshot(snapshot, nil); !errors.As(err, &limit) {
		t.Fatalf("restauracion evade limite: %v", err)
	}
}

func TestRuntime_TrackingLimitadoPreservaRollback(t *testing.T) {
	failure := errors.New("entry failed")
	registerTestAction(t, "test:runtime-tracking-fail", func(context.Context, instrumentation.ActionExecutorArgs) error { return failure })
	qm, u := buildControlledQM(t, instrumentation.RuntimeOptions{MaxTrackingEntries: 2}, map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withOnTransition("GO", []string{"stateB"}, nil)), "stateB": newTransitionReality("stateB", withEntryAction("test:runtime-tracking-fail")),
	})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	u.tracking = []string{"older", "stateA"}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("GO").Build()); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(u.tracking, []string{"older", "stateA"}) {
		t.Fatalf("rollback pierde historial: %v", u.tracking)
	}
	for i := 0; i < 5; i++ {
		if err := qm.PositionMachine(context.Background(), nil, "u1", "stateA", false); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(u.tracking, []string{"stateA", "stateA"}) {
		t.Fatalf("historial sin limite: %v", u.tracking)
	}
}
