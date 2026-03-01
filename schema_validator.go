package statepro

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rendis/statepro/v3/theoretical"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const quantumMachineSchemaResource = "statepro://schema/quantum-machine.schema.json"

var (
	//go:embed schema/quantum-machine.schema.json
	quantumMachineSchemaBytes []byte

	quantumMachineSchemaOnce sync.Once
	quantumMachineSchema     *jsonschema.Schema
	quantumMachineSchemaErr  error
)

// QuantumMachineJSONSchema returns the JSON Schema used to validate state machine definitions.
func QuantumMachineJSONSchema() []byte {
	return append([]byte(nil), quantumMachineSchemaBytes...)
}

// ValidateQuantumMachineBySchema validates the model using the embedded JSON Schema.
func ValidateQuantumMachineBySchema(source *theoretical.QuantumMachineModel) error {
	if source == nil {
		return fmt.Errorf("source model cannot be nil")
	}

	m, err := SerializeQuantumMachineToMap(source)
	if err != nil {
		return fmt.Errorf("error serializing model: %w", err)
	}

	return ValidateQuantumMachineBySchemaFromMap(m)
}

// ValidateQuantumMachineBySchemaFromMap validates a map definition using the embedded JSON Schema.
func ValidateQuantumMachineBySchemaFromMap(source map[string]any) error {
	if source == nil {
		return fmt.Errorf("source map cannot be nil")
	}

	return validateQuantumMachineBySchema(source)
}

// ValidateQuantumMachineBySchemaFromBinary validates a JSON definition using the embedded JSON Schema.
func ValidateQuantumMachineBySchemaFromBinary(source []byte) error {
	if len(source) == 0 {
		return fmt.Errorf("source payload cannot be empty")
	}

	var payload any
	if err := json.Unmarshal(source, &payload); err != nil {
		return fmt.Errorf("invalid json payload: %w", err)
	}

	if looksLikeJSONSchemaDocument(payload) {
		return fmt.Errorf(
			"payload appears to be a JSON Schema document, not a QuantumMachineModel instance; " +
				"use a JSON Schema meta-schema (for example draft 2020-12) to validate schema documents",
		)
	}

	return validateQuantumMachineBySchema(payload)
}

func validateQuantumMachineBySchema(payload any) error {
	schema, err := getQuantumMachineSchema()
	if err != nil {
		return fmt.Errorf("error loading quantum machine schema: %w", err)
	}

	if err = schema.Validate(payload); err != nil {
		return fmt.Errorf("json schema validation failed: %w", err)
	}

	return nil
}

func getQuantumMachineSchema() (*jsonschema.Schema, error) {
	quantumMachineSchemaOnce.Do(func() {
		compiler := jsonschema.NewCompiler()
		var schemaDocument any

		if err := json.Unmarshal(quantumMachineSchemaBytes, &schemaDocument); err != nil {
			quantumMachineSchemaErr = err
			return
		}

		if err := compiler.AddResource(quantumMachineSchemaResource, schemaDocument); err != nil {
			quantumMachineSchemaErr = err
			return
		}

		quantumMachineSchema, quantumMachineSchemaErr = compiler.Compile(quantumMachineSchemaResource)
	})

	if quantumMachineSchemaErr != nil {
		return nil, quantumMachineSchemaErr
	}

	return quantumMachineSchema, nil
}

func looksLikeJSONSchemaDocument(payload any) bool {
	m, ok := payload.(map[string]any)
	if !ok {
		return false
	}

	_, hasSchema := m["$schema"]
	_, hasDefs := m["$defs"]
	_, hasProperties := m["properties"]
	_, hasID := m["id"]
	_, hasCanonicalName := m["canonicalName"]
	_, hasUniverses := m["universes"]
	_, hasInitials := m["initials"]

	if hasID || hasCanonicalName || hasUniverses || hasInitials {
		return false
	}

	return hasSchema && (hasDefs || hasProperties)
}
