package statepro

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rendis/statepro/v3/theoretical"
)

func minimalValidMachineJSON() string {
	return `{
		"id":"machine",
		"canonicalName":"machine",
		"version":"1.0.0",
		"initials":["U:main"],
		"universes":{
			"main":{
				"id":"main",
				"canonicalName":"main",
				"version":"1.0.0",
				"initial":"A",
				"realities":{
					"A":{"id":"A","type":"transition","always":[{"targets":["END"]}]},
					"END":{"id":"END","type":"final"}
				}
			}
		}
	}`
}

func TestAdv_ValidateDefinition_NilAndEmpty(t *testing.T) {
	if err := ValidateQuantumMachineDefinition(nil); err == nil {
		t.Fatal("expected error for nil model")
	}
	if err := ValidateQuantumMachineDefinitionFromMap(nil); err == nil {
		t.Fatal("expected error for nil map")
	}
	if err := ValidateQuantumMachineDefinitionFromBinary(nil); err == nil {
		t.Fatal("expected error for nil binary")
	}
	if err := ValidateQuantumMachineDefinitionFromBinary([]byte{}); err == nil {
		t.Fatal("expected error for empty binary")
	}
}

func TestAdv_ValidateDefinition_FromModelAndMap_Happy(t *testing.T) {
	var model theoretical.QuantumMachineModel
	if err := json.Unmarshal([]byte(minimalValidMachineJSON()), &model); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := ValidateQuantumMachineDefinition(&model); err != nil {
		t.Fatalf("ValidateQuantumMachineDefinition: %v", err)
	}

	var asMap map[string]any
	if err := json.Unmarshal([]byte(minimalValidMachineJSON()), &asMap); err != nil {
		t.Fatalf("map unmarshal: %v", err)
	}
	if err := ValidateQuantumMachineDefinitionFromMap(asMap); err != nil {
		t.Fatalf("ValidateQuantumMachineDefinitionFromMap: %v", err)
	}
}

func TestAdv_ValidateDefinition_HostileSemantics(t *testing.T) {
	// Payloads that must be rejected. Some die at JSON Schema, others at semantics —
	// both layers are part of the security boundary.
	mustFail := []struct {
		name    string
		payload string
	}{
		{
			name: "no universes",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],"universes":{}
			}`,
		},
		{
			name: "nil universe entry",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],
				"universes":{"main":null}
			}`,
		},
		{
			name: "universe with no realities",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],
				"universes":{"main":{"id":"main","canonicalName":"main","version":"1.0.0","realities":{}}}
			}`,
		},
		{
			name: "transition reality without on/always",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"A",
					"realities":{"A":{"id":"A","type":"transition"},"END":{"id":"END","type":"final"}}
				}}
			}`,
		},
		{
			name: "initials as bare reality ref",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["A"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"A",
					"realities":{"A":{"id":"A","type":"transition","always":[{"targets":["END"]}]},"END":{"id":"END","type":"final"}}
				}}
			}`,
		},
		{
			name: "invalid initials format",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"A",
					"realities":{"A":{"id":"A","type":"transition","always":[{"targets":["END"]}]},"END":{"id":"END","type":"final"}}
				}}
			}`,
		},
		{
			name: "notify with internal target rejected by schema",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"A",
					"realities":{
						"A":{"id":"A","type":"transition","on":{"go":[{"type":"notify","targets":["END"]}]}},
						"END":{"id":"END","type":"final"}
					}
				}}
			}`,
		},
		{
			name:    "totally garbage payload",
			payload: `{"id":1,"universes":[]}`,
		},
	}

	for _, tc := range mustFail {
		t.Run(tc.name+"/binary", func(t *testing.T) {
			if err := ValidateQuantumMachineDefinitionFromBinary([]byte(tc.payload)); err == nil {
				t.Fatal("expected validation to reject hostile payload")
			}
		})
		t.Run(tc.name+"/map", func(t *testing.T) {
			var asMap map[string]any
			if err := json.Unmarshal([]byte(tc.payload), &asMap); err != nil {
				return // invalid JSON is already covered elsewhere
			}
			if err := ValidateQuantumMachineDefinitionFromMap(asMap); err == nil {
				t.Fatal("expected ValidateQuantumMachineDefinitionFromMap to reject hostile payload")
			}
		})
	}

	semanticCases := []struct {
		name        string
		payload     string
		mustContain string
	}{
		{
			name: "initial reality missing",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"missing",
					"realities":{"A":{"id":"A","type":"transition","always":[{"targets":["END"]}]},"END":{"id":"END","type":"final"}}
				}}
			}`,
			mustContain: "does not reference an existing reality",
		},
		{
			name: "unknown external target universe",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"A",
					"realities":{
						"A":{"id":"A","type":"transition","on":{"go":[{"targets":["U:ghost"]}]}},
						"END":{"id":"END","type":"final"}
					}
				}}
			}`,
			mustContain: "unknown universe",
		},
		{
			name: "reality key mismatch",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"A",
					"realities":{
						"A":{"id":"other","type":"transition","always":[{"targets":["END"]}]},
						"END":{"id":"END","type":"final"}
					}
				}}
			}`,
			mustContain: "must match reality.id",
		},
		{
			name: "unknown reality in U:universe:reality initial",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main:missing"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"A",
					"realities":{"A":{"id":"A","type":"transition","always":[{"targets":["END"]}]},"END":{"id":"END","type":"final"}}
				}}
			}`,
			mustContain: "unknown reality 'missing'",
		},
		{
			name: "transition with empty targets",
			payload: `{
				"id":"machine","canonicalName":"machine","version":"1.0.0",
				"initials":["U:main"],
				"universes":{"main":{
					"id":"main","canonicalName":"main","version":"1.0.0","initial":"A",
					"realities":{
						"A":{"id":"A","type":"transition","on":{"go":[{"targets":[]}]}},
						"END":{"id":"END","type":"final"}
					}
				}}
			}`,
			mustContain: "target",
		},
	}

	for _, tc := range semanticCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateQuantumMachineDefinitionFromBinary([]byte(tc.payload))
			if err == nil {
				t.Fatal("expected semantic validation error")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.mustContain)) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.mustContain)
			}
		})
	}
}

func TestAdv_SerializeDeserialize_RoundTripHostile(t *testing.T) {
	if m, err := SerializeQuantumMachineToMap(nil); err != nil || m != nil {
		t.Fatalf("nil serialize map: m=%v err=%v", m, err)
	}
	if b, err := SerializeQuantumMachineToBinary(nil); err != nil || b != nil {
		t.Fatalf("nil serialize binary: b=%v err=%v", b, err)
	}

	var model theoretical.QuantumMachineModel
	if err := json.Unmarshal([]byte(minimalValidMachineJSON()), &model); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	asMap, err := SerializeQuantumMachineToMap(&model)
	if err != nil {
		t.Fatalf("SerializeQuantumMachineToMap: %v", err)
	}
	back, err := DeserializeQuantumMachineFromMap(asMap)
	if err != nil {
		t.Fatalf("DeserializeQuantumMachineFromMap: %v", err)
	}
	if back.ID != model.ID {
		t.Fatalf("round-trip id %q != %q", back.ID, model.ID)
	}

	bin, err := SerializeQuantumMachineToBinary(&model)
	if err != nil {
		t.Fatalf("SerializeQuantumMachineToBinary: %v", err)
	}
	back2, err := DeserializeQuantumMachineFromBinary(bin)
	if err != nil {
		t.Fatalf("DeserializeQuantumMachineFromBinary: %v", err)
	}
	if back2.CanonicalName != model.CanonicalName {
		t.Fatalf("round-trip canonical %q", back2.CanonicalName)
	}
}

func TestAdv_Deserialize_GarbageJSON(t *testing.T) {
	if _, err := DeserializeQuantumMachineFromBinary([]byte(`{not json`)); err == nil {
		t.Fatal("expected error for garbage JSON")
	}
	if _, err := DeserializeQuantumMachineFromMap(map[string]any{"universes": "not-a-map"}); err == nil {
		t.Fatal("expected error for hostile map shape")
	}
}
