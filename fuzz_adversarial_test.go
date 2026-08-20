package statepro

import (
	"testing"
)

// FuzzValidateDefinitionBinary feeds arbitrary bytes through schema+semantic validation.
// Hostile / truncated / huge payloads must never panic.
func FuzzValidateDefinitionBinary(f *testing.F) {
	seeds := []string{
		``,
		`{}`,
		`null`,
		`[]`,
		`{"id":"m"}`,
		minimalValidMachineJSON(),
		`{"id":"machine","canonicalName":"machine","version":"1.0.0","initials":["U:main"],"universes":{}}`,
		`{"id":1,"universes":[]}`,
		`{not json`,
		string([]byte{0x00, 0xff, 0xfe}),
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, payload []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("ValidateQuantumMachineDefinitionFromBinary panicked: %v", r)
			}
		}()
		_ = ValidateQuantumMachineDefinitionFromBinary(payload)
	})
}

// FuzzDeserializeQuantumMachine ensures serde never panics on hostile maps/binaries.
func FuzzDeserializeQuantumMachine(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(minimalValidMachineJSON()))
	f.Add([]byte(`{"universes":"nope"}`))
	f.Add([]byte(`{"id":"x","universes":{"u":{"id":"u"}}}`))

	f.Fuzz(func(t *testing.T, payload []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("deserialize panicked: %v", r)
			}
		}()
		model, err := DeserializeQuantumMachineFromBinary(payload)
		if err != nil || model == nil {
			return
		}
		_, _ = SerializeQuantumMachineToBinary(model)
		_, _ = SerializeQuantumMachineToMap(model)
	})
}
