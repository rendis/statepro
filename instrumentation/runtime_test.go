package instrumentation

import "testing"

func TestRuntimeOptions_RechazaLimitesNegativos(t *testing.T) {
	for _, options := range []RuntimeOptions{{MaxConcurrentInvokes: -1}, {MaxAccumulatedEvents: -1}, {MaxTrackingEntries: -1}} {
		if err := options.Validate(); err == nil {
			t.Fatalf("opcion invalida aceptada: %+v", options)
		}
	}
	for _, options := range []RuntimeOptions{{}, {MaxConcurrentInvokes: 2, MaxAccumulatedEvents: 100, MaxTrackingEntries: 20, StrictObservers: true}} {
		if err := options.Validate(); err != nil {
			t.Fatal(err)
		}
	}
}
