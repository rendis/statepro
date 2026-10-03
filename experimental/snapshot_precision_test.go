package experimental

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

func TestSnapshot_PreservaNumerosEnCapturaJSONYRestauracion(t *testing.T) {
	qm, u := buildQM(t, "stateA", map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA")})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	u.metadata = map[string]any{
		"integer":  int64(9007199254740993),
		"unsigned": uint64(18446744073709551615),
		"decimal":  json.Number("0.12345678901234567890123456789"),
		"nested":   []any{map[string]any{"count": json.Number("9007199254740993e0")}},
		"small":    42,
	}
	snapshot, err := instrumentation.GetSnapshotWithError(qm)
	if err != nil {
		t.Fatal(err)
	}
	metadata := snapshot.Snapshots["u1"]["metadata"].(map[string]any)
	if metadata["small"] != float64(42) {
		t.Fatal("numeros representables deben conservar el tipo legacy")
	}
	for key, expected := range map[string]string{"integer": "9007199254740993", "unsigned": "18446744073709551615", "decimal": "0.12345678901234567890123456789"} {
		if metadata[key] != json.Number(expected) {
			t.Fatalf("numero alterado %s: %#v", key, metadata[key])
		}
	}
	encoded, err := snapshot.ToJson()
	if err != nil {
		t.Fatal(err)
	}
	var decoded instrumentation.MachineSnapshot
	if err := json.Unmarshal([]byte(encoded), &decoded); err != nil {
		t.Fatal(err)
	}
	if err := qm.LoadSnapshot(&decoded, nil); err != nil {
		t.Fatal(err)
	}
	restored, err := qm.GetSnapshotWithError()
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, err := restored.ToJson()
	if err != nil {
		t.Fatal(err)
	}
	if encoded != roundtrip {
		t.Fatalf("roundtrip modifica numeros:\n%s\n%s", encoded, roundtrip)
	}
	// Captured nested metadata must not alias live metadata after restoration.
	u.metadata["nested"].([]any)[0].(map[string]any)["count"] = "changed"
	if metadata["nested"].([]any)[0].(map[string]any)["count"] != json.Number("9007199254740993e0") {
		t.Fatal("snapshot comparte metadata anidada")
	}
}

func TestSnapshot_ErrorNoDevuelveCapturaParcial(t *testing.T) {
	realities := map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA")}
	qm, _, bad := buildMultiUniverseQM(t, "stateA", realities, "stateA", realities)
	bad.metadata["unsupported"] = make(chan int)
	snapshot, err := qm.GetSnapshotWithError()
	var serializationError *json.UnsupportedTypeError
	if snapshot != nil || !errors.As(err, &serializationError) || !strings.Contains(err.Error(), "u2") {
		t.Fatalf("captura parcial o error perdido: %#v, %v", snapshot, err)
	}
	if qm.GetSnapshot() != nil {
		t.Fatal("API legacy no debe devolver un snapshot incompleto")
	}
}

func TestSnapshot_PreservaNumerosEnEventosAcumulados(t *testing.T) {
	qm, u := buildQM(t, "stateA", map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA")})
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	u.realityBeforeSuperposition = u.currentReality
	u.currentReality = nil
	u.inSuperposition = true
	u.eventAccumulator = newEventAccumulator()
	u.eventAccumulator.Accumulate("stateA", &Event{Name: "COUNT", Data: map[string]any{
		"count": int64(9007199254740993), "ratio": json.Number("0.12345678901234567890123456789"), "small": 42,
	}})
	snapshot, err := qm.GetSnapshotWithError()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := snapshot.ToJson()
	if err != nil {
		t.Fatal(err)
	}
	var decoded instrumentation.MachineSnapshot
	if err := json.Unmarshal([]byte(encoded), &decoded); err != nil {
		t.Fatal(err)
	}
	if err := qm.LoadSnapshot(&decoded, nil); err != nil {
		t.Fatal(err)
	}
	events := u.eventAccumulator.GetStatistics().GetRealitiesEvents()["stateA"]
	if len(events) != 1 {
		t.Fatalf("eventos alterados: %v", events)
	}
	data := events[0].GetData()
	if data["count"] != json.Number("9007199254740993") || data["ratio"] != json.Number("0.12345678901234567890123456789") || data["small"] != float64(42) {
		t.Fatalf("datos de evento alterados: %#v", data)
	}
}

func TestSnapshot_AccionCapturaConErrorSinReentrarMutex(t *testing.T) {
	var captured error
	registerTestAction(t, "test:checked-snapshot", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		_, captured = instrumentation.GetSnapshotWithError(args)
		return nil
	})
	qm, u := buildQM(t, "stateA", map[string]*theoretical.RealityModel{"stateA": newTransitionReality("stateA", withEntryAction("test:checked-snapshot"))})
	u.metadata["unsupported"] = make(chan int)
	done := make(chan error, 1)
	go func() { done <- qm.Init(context.Background(), nil) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
		var serializationError *json.UnsupportedTypeError
		if !errors.As(captured, &serializationError) {
			t.Fatalf("error de captura perdido: %v", captured)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("captura desde args bloqueada por reentrada")
	}
}
