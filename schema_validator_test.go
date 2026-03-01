package statepro

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuantumMachineJSONSchema_IsValidJSON(t *testing.T) {
	schema := QuantumMachineJSONSchema()
	if len(schema) == 0 {
		t.Fatal("expected non-empty schema")
	}

	var payload map[string]any
	if err := json.Unmarshal(schema, &payload); err != nil {
		t.Fatalf("expected schema to be valid JSON: %v", err)
	}

	if payload["$schema"] == nil {
		t.Fatal("expected $schema field in schema document")
	}
}

func TestQuantumMachineJSONSchema_CompilesAgainstMetaSchema(t *testing.T) {
	_, err := getQuantumMachineSchema()
	if err != nil {
		t.Fatalf("expected embedded schema to compile against JSON Schema meta-schema: %v", err)
	}
}

func TestValidateQuantumMachineBySchemaFromBinary_ValidExamples(t *testing.T) {
	examples := []string{
		"example/sm/state_machine.json",
		"example/cli/state_machine.json",
		"example/bot/state_machine.json",
		"schema/examples/neutral-machine.json",
	}

	for _, examplePath := range examples {
		t.Run(examplePath, func(t *testing.T) {
			b, err := os.ReadFile(examplePath)
			if err != nil {
				t.Fatalf("error reading fixture %s: %v", examplePath, err)
			}

			if err = ValidateQuantumMachineBySchemaFromBinary(b); err != nil {
				t.Fatalf("expected valid schema for %s, got: %v", examplePath, err)
			}
		})
	}
}

func TestValidateQuantumMachineBySchemaFromBinary_AllStateMachineFixtures(t *testing.T) {
	var files []string

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path == ".git" || path == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "state_machine.json") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("error walking repository for fixtures: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("expected at least one state_machine.json fixture")
	}

	for _, file := range files {
		file := file
		t.Run(file, func(t *testing.T) {
			b, readErr := os.ReadFile(file)
			if readErr != nil {
				t.Fatalf("error reading fixture %s: %v", file, readErr)
			}
			if validateErr := ValidateQuantumMachineBySchemaFromBinary(b); validateErr != nil {
				t.Fatalf("fixture %s is invalid against schema: %v", file, validateErr)
			}
		})
	}
}

func TestValidateQuantumMachineBySchemaFromBinary_SchemaDocumentPayload(t *testing.T) {
	schemaPayload := QuantumMachineJSONSchema()
	err := ValidateQuantumMachineBySchemaFromBinary(schemaPayload)
	if err == nil {
		t.Fatal("expected error when validating a JSON Schema document as machine payload")
	}

	if !strings.Contains(err.Error(), "appears to be a JSON Schema document") {
		t.Fatalf("expected schema-document hint in error, got: %v", err)
	}
}

func TestValidateQuantumMachineBySchema_ValidTypedModel(t *testing.T) {
	b, err := os.ReadFile("example/sm/state_machine.json")
	if err != nil {
		t.Fatalf("error reading fixture: %v", err)
	}

	model, err := DeserializeQuantumMachineFromBinary(b)
	if err != nil {
		t.Fatalf("error deserializing fixture: %v", err)
	}

	if err = ValidateQuantumMachineBySchema(model); err != nil {
		t.Fatalf("expected typed model to be valid, got: %v", err)
	}
}

func TestValidateQuantumMachineBySchemaFromBinary_InvalidCases(t *testing.T) {
	cases := []struct {
		name    string
		payload string
	}{
		{
			name: "missing machine id",
			payload: `{
				"canonicalName":"machine",
				"version":"1.0.0",
				"initials":["U:main"],
				"universes":{
					"main":{
						"id":"main",
						"canonicalName":"main",
						"version":"1.0.0",
						"realities":{
							"START":{"id":"START","type":"transition","always":[{"targets":["END"]}]},
							"END":{"id":"END","type":"final"}
						}
					}
				}
			}`,
		},
		{
			name: "invalid target reference",
			payload: `{
				"id":"machine",
				"canonicalName":"machine",
				"version":"1.0.0",
				"initials":["U:main"],
				"universes":{
					"main":{
						"id":"main",
						"canonicalName":"main",
						"version":"1.0.0",
						"realities":{
							"START":{
								"id":"START",
								"type":"transition",
								"on":{"go":[{"targets":["U::bad"]}]}
							},
							"END":{"id":"END","type":"final"}
						}
					}
				}
			}`,
		},
		{
			name: "notify transition cannot target internal reality",
			payload: `{
				"id":"machine",
				"canonicalName":"machine",
				"version":"1.0.0",
				"initials":["U:main"],
				"universes":{
					"main":{
						"id":"main",
						"canonicalName":"main",
						"version":"1.0.0",
						"realities":{
							"A":{
								"id":"A",
								"type":"transition",
								"on":{"go":[{"type":"notify","targets":["B"]}]}
							},
							"B":{"id":"B","type":"final"}
						}
					}
				}
			}`,
		},
		{
			name: "transition reality requires on or always",
			payload: `{
				"id":"machine",
				"canonicalName":"machine",
				"version":"1.0.0",
				"initials":["U:main"],
				"universes":{
					"main":{
						"id":"main",
						"canonicalName":"main",
						"version":"1.0.0",
						"realities":{
							"A":{"id":"A","type":"transition"},
							"B":{"id":"B","type":"final"}
						}
					}
				}
			}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateQuantumMachineBySchemaFromBinary([]byte(tc.payload))
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			if !strings.Contains(err.Error(), "json schema validation failed") {
				t.Fatalf("expected schema validation error, got: %v", err)
			}
		})
	}
}

func TestValidateQuantumMachineBySchemaFromMap_NilMap(t *testing.T) {
	err := ValidateQuantumMachineBySchemaFromMap(nil)
	if err == nil {
		t.Fatal("expected error for nil map")
	}
}
