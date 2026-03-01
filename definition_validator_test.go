package statepro

import (
	"os"
	"strings"
	"testing"
)

func TestValidateQuantumMachineDefinitionFromBinary_AllStateMachineFixtures(t *testing.T) {
	fixtures := []string{
		"example/sm/state_machine.json",
		"example/cli/state_machine.json",
		"example/bot/state_machine.json",
		"schema/examples/neutral-machine.json",
	}

	for _, file := range fixtures {
		file := file
		t.Run(file, func(t *testing.T) {
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("error reading fixture %s: %v", file, err)
			}
			if err = ValidateQuantumMachineDefinitionFromBinary(b); err != nil {
				t.Fatalf("expected fixture %s to be semantically valid, got: %v", file, err)
			}
		})
	}
}

func TestValidateQuantumMachineDefinitionFromBinary_InvalidSemanticCases(t *testing.T) {
	cases := []struct {
		name        string
		payload     string
		mustContain string
	}{
		{
			name: "unknown internal target reality",
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
						"initial":"A",
						"realities":{
							"A":{
								"id":"A",
								"type":"transition",
								"on":{"go":[{"targets":["B"]}]}
							},
							"END":{"id":"END","type":"final"}
						}
					}
				}
			}`,
			mustContain: "unknown internal reality 'B'",
		},
		{
			name: "unknown universe in initials",
			payload: `{
				"id":"machine",
				"canonicalName":"machine",
				"version":"1.0.0",
				"initials":["U:missing"],
				"universes":{
					"main":{
						"id":"main",
						"canonicalName":"main",
						"version":"1.0.0",
						"initial":"A",
						"realities":{
							"A":{
								"id":"A",
								"type":"transition",
								"always":[{"targets":["END"]}]
							},
							"END":{"id":"END","type":"final"}
						}
					}
				}
			}`,
			mustContain: "references unknown universe 'missing'",
		},
		{
			name: "universe map key mismatch",
			payload: `{
				"id":"machine",
				"canonicalName":"machine",
				"version":"1.0.0",
				"initials":["U:main"],
				"universes":{
					"main":{
						"id":"main-v2",
						"canonicalName":"main",
						"version":"1.0.0",
						"initial":"A",
						"realities":{
							"A":{
								"id":"A",
								"type":"transition",
								"always":[{"targets":["END"]}]
							},
							"END":{"id":"END","type":"final"}
						}
					}
				}
			}`,
			mustContain: "universe key 'main' must match universe.id 'main-v2'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateQuantumMachineDefinitionFromBinary([]byte(tc.payload))
			if err == nil {
				t.Fatal("expected semantic validation error, got nil")
			}

			if !strings.Contains(err.Error(), tc.mustContain) {
				t.Fatalf("expected error to contain %q, got: %v", tc.mustContain, err)
			}
		})
	}
}
