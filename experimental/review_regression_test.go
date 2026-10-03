package experimental

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

func TestRevision_SnapshotInvalidoNoModificaNingunUniverso(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA"),
		"stateB": newTransitionReality("stateB"),
	}
	qm, _, _ := buildMultiUniverseQM(t, "stateA", realities, "stateA", realities)
	if err := qm.Init(context.Background(), "original"); err != nil {
		t.Fatal(err)
	}
	before := qm.GetSnapshot()
	bad := qm.GetSnapshot()
	bad.Snapshots["u1"]["currentReality"] = "stateB"
	bad.Snapshots["u2"]["currentReality"] = "missing"
	bad.Tracking["u1"] = []string{"stateB"}
	if err := qm.LoadSnapshot(bad, "replacement"); err == nil {
		t.Fatal("se esperaba rechazar el snapshot")
	}
	if !reflect.DeepEqual(before, qm.GetSnapshot()) || qm.machineContext != "original" {
		t.Fatal("rechazar un snapshot debe conservar estado, tracking y contexto")
	}
}

func TestRevision_SnapshotRechazaEstadosInconsistentes(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(instrumentation.SerializedUniverseSnapshot)
	}{
		{"sin_reality", func(s instrumentation.SerializedUniverseSnapshot) { delete(s, "currentReality") }},
		{"superposicion_con_reality", func(s instrumentation.SerializedUniverseSnapshot) { s["inSuperposition"] = true }},
		{"reality_anterior_desconocida", func(s instrumentation.SerializedUniverseSnapshot) { s["realityBeforeSuperposition"] = "missing" }},
		{"identidad_incorrecta", func(s instrumentation.SerializedUniverseSnapshot) { s["id"] = "another" }},
		{"no_inicializado_con_reality", func(s instrumentation.SerializedUniverseSnapshot) { s["initialized"] = false }},
		{"acumulador_con_reality_desconocida", func(s instrumentation.SerializedUniverseSnapshot) {
			s["accumulator"] = map[string]any{"realitiesEvents": map[string]any{"missing": []any{}}}
		}},
		{"acumulador_con_evento_nulo", func(s instrumentation.SerializedUniverseSnapshot) {
			s["accumulator"] = map[string]any{"realitiesEvents": map[string]any{"stateA": []any{nil}}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			qm, _ := buildQM(t, "stateA", map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA")})
			if err := qm.Init(context.Background(), nil); err != nil {
				t.Fatal(err)
			}
			before := qm.GetSnapshot()
			invalid := qm.GetSnapshot()
			test.change(invalid.Snapshots["u1"])
			if err := qm.LoadSnapshot(invalid, nil); err == nil {
				t.Fatal("se esperaba rechazar estado inconsistente")
			}
			if !reflect.DeepEqual(before, qm.GetSnapshot()) {
				t.Fatal("el rechazo debe ser atomico")
			}
		})
	}
}

func TestRevision_SnapshotRestauraMetadataYTracking(t *testing.T) {
	qm, u := buildQM(t, "stateA", map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA")})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	u.metadata["original"] = "value"
	snapshot := qm.GetSnapshot()
	snapshot.Tracking = nil
	u.metadata["later"] = true
	u.tracking = []string{"later"}
	if err := qm.LoadSnapshot(snapshot, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(u.metadata, map[string]any{"original": "value"}) {
		t.Fatalf("metadata no restaurada: %v", u.metadata)
	}
	if len(u.tracking) != 0 {
		t.Fatalf("tracking anterior retenido: %v", u.tracking)
	}
}

func TestRevision_SnapshotSuperposicionInicializaAcumuladorYFinal(t *testing.T) {
	qm, u := buildQM(t, "DONE", map[string]*theoretical.RealityModel{"DONE": newFinalReality("DONE"), "stateA": newTransitionReality("stateA")})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	snapshot := qm.GetSnapshot()
	s := snapshot.Snapshots["u1"]
	delete(s, "currentReality")
	delete(s, "accumulator")
	s["inSuperposition"] = true
	s["realityInitialized"] = false
	if err := qm.LoadSnapshot(snapshot, nil); err != nil {
		t.Fatal(err)
	}
	if u.isFinalReality || u.eventAccumulator == nil {
		t.Fatal("debe restaurar superposicion no final con acumulador vacio")
	}
	if _, err := qm.SendEvent(context.Background(), NewEventBuilder("tick").Build()); err != nil {
		t.Fatal(err)
	}
}

func TestRevision_LoadSnapshotConcurrenteConMetadataDeInvoke(t *testing.T) {
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	restored := make(chan struct{})
	registerTestInvoke(t, "test:review-metadata", func(_ context.Context, args instrumentation.InvokeExecutorArgs) {
		close(started)
		<-release
		defer close(done)
		for i := 0; i < 500; i++ {
			args.AddToUniverseMetadata("invoke", i)
			args.GetUniverseMetadata()
		}
		<-restored
		args.AddToUniverseMetadata("after", true)
	})
	qm, u := buildQM(t, "stateA", map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryInvoke("test:review-metadata"))})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	<-started
	snapshot := qm.GetSnapshot()
	close(release)
	for i := 0; i < 500; i++ {
		if err := qm.LoadSnapshot(snapshot, nil); err != nil {
			t.Fatal(err)
		}
	}
	close(restored)
	<-done
	if u.metadata["after"] != true {
		t.Fatal("los args del invoke deben conservar el mapa vivo tras restaurar")
	}
}

type reviewCustomEvent struct{ instrumentation.Event }

type reviewAdmissionContext struct {
	context.Context
	checked chan struct{}
	once    sync.Once
}

func (c *reviewAdmissionContext) Err() error {
	err := c.Context.Err()
	c.once.Do(func() { close(c.checked) })
	return err
}

func TestRevision_AcumuladorAceptaImplementacionPublicaDeEvent(t *testing.T) {
	event := NewEventBuilder("tick").SetData(map[string]any{"value": 1}).Build()
	acc := newEventAccumulator()
	acc.Accumulate("stateA", reviewCustomEvent{event})
	got := acc.GetStatistics().GetRealityEvents("stateA")["tick"]
	if got == nil || !reflect.DeepEqual(got.GetData(), event.GetData()) || got.GetEvtType() != event.GetEvtType() {
		t.Fatal("se deben conservar los campos del evento")
	}
}

func TestRevision_SendEventRechazaNilYContextoCancelado(t *testing.T) {
	qm, _ := buildQM(t, "stateA", map[string]*theoretical.RealityModel{
		"stateA": newTransitionReality("stateA", withOnTransition("go", []string{"stateB"}, nil)),
		"stateB": newTransitionReality("stateB"),
	})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	before := qm.GetSnapshot()
	var typedNil *Event
	for _, event := range []instrumentation.Event{nil, typedNil} {
		if handled, err := qm.SendEvent(context.Background(), event); err == nil || handled {
			t.Fatal("evento nil debe devolver error sin procesarse")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if handled, err := qm.SendEvent(ctx, NewEventBuilder("go").Build()); handled || !errors.Is(err, context.Canceled) {
		t.Fatalf("resultado de cancelacion: %v, %v", handled, err)
	}
	if !reflect.DeepEqual(before, qm.GetSnapshot()) {
		t.Fatal("no debe mutar antes de admitir el evento")
	}
	// Cancellation while waiting for the machine lock must also prevent admission.
	qm.quantumMachineMtx.Lock()
	ctx, cancel = context.WithCancel(context.Background())
	queued := &reviewAdmissionContext{Context: ctx, checked: make(chan struct{})}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handled, err := qm.SendEvent(queued, NewEventBuilder("go").Build())
		if handled || !errors.Is(err, context.Canceled) {
			t.Errorf("evento cancelado en cola: %v, %v", handled, err)
		}
	}()
	<-queued.checked
	cancel()
	qm.quantumMachineMtx.Unlock()
	wg.Wait()
}
