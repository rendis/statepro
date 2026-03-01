package statepro

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rendis/statepro/v3/theoretical"
)

// ValidateQuantumMachineDefinition validates both schema and semantic integrity.
// It is stricter than ValidateQuantumMachineBySchema and ensures references resolve.
func ValidateQuantumMachineDefinition(source *theoretical.QuantumMachineModel) error {
	if source == nil {
		return fmt.Errorf("source model cannot be nil")
	}

	if err := ValidateQuantumMachineBySchema(source); err != nil {
		return err
	}

	if err := validateQuantumMachineSemantics(source); err != nil {
		return fmt.Errorf("semantic validation failed: %w", err)
	}

	return nil
}

// ValidateQuantumMachineDefinitionFromMap validates both schema and semantic integrity from map payload.
func ValidateQuantumMachineDefinitionFromMap(source map[string]any) error {
	if source == nil {
		return fmt.Errorf("source map cannot be nil")
	}

	if err := ValidateQuantumMachineBySchemaFromMap(source); err != nil {
		return err
	}

	model, err := DeserializeQuantumMachineFromMap(source)
	if err != nil {
		return fmt.Errorf("error deserializing model for semantic validation: %w", err)
	}

	if err = validateQuantumMachineSemantics(model); err != nil {
		return fmt.Errorf("semantic validation failed: %w", err)
	}

	return nil
}

// ValidateQuantumMachineDefinitionFromBinary validates both schema and semantic integrity from JSON payload.
func ValidateQuantumMachineDefinitionFromBinary(source []byte) error {
	if len(source) == 0 {
		return fmt.Errorf("source payload cannot be empty")
	}

	if err := ValidateQuantumMachineBySchemaFromBinary(source); err != nil {
		return err
	}

	model, err := DeserializeQuantumMachineFromBinary(source)
	if err != nil {
		return fmt.Errorf("error deserializing model for semantic validation: %w", err)
	}

	if err = validateQuantumMachineSemantics(model); err != nil {
		return fmt.Errorf("semantic validation failed: %w", err)
	}

	return nil
}

type semanticValidationErrors struct {
	errors []string
}

func (v *semanticValidationErrors) add(format string, args ...any) {
	v.errors = append(v.errors, fmt.Sprintf(format, args...))
}

func (v *semanticValidationErrors) hasErrors() bool {
	return len(v.errors) > 0
}

func (v *semanticValidationErrors) Error() string {
	if len(v.errors) == 0 {
		return ""
	}
	return strings.Join(v.errors, "; ")
}

func validateQuantumMachineSemantics(model *theoretical.QuantumMachineModel) error {
	if model == nil {
		return fmt.Errorf("model cannot be nil")
	}

	errCollector := &semanticValidationErrors{}

	if len(model.Universes) == 0 {
		errCollector.add("machine must define at least one universe")
	}

	for universeKey, universe := range model.Universes {
		if universe == nil {
			errCollector.add("universe '%s' cannot be nil", universeKey)
			continue
		}

		if universe.ID != universeKey {
			errCollector.add("universe key '%s' must match universe.id '%s'", universeKey, universe.ID)
		}

		if len(universe.Realities) == 0 {
			errCollector.add("universe '%s' must define at least one reality", universeKey)
			continue
		}

		for realityKey, reality := range universe.Realities {
			if reality == nil {
				errCollector.add("reality '%s' in universe '%s' cannot be nil", realityKey, universeKey)
				continue
			}

			if reality.ID != realityKey {
				errCollector.add("reality key '%s' in universe '%s' must match reality.id '%s'", realityKey, universeKey, reality.ID)
			}

			if reality.Type == theoretical.RealityTypeTransition && !hasEffectiveTransitionFlow(reality) {
				errCollector.add("transition reality '%s' in universe '%s' must define non-empty 'on' or non-empty 'always'", realityKey, universeKey)
			}
		}

		if universe.Initial != nil && *universe.Initial != "" {
			if _, ok := universe.Realities[*universe.Initial]; !ok {
				errCollector.add("universe '%s' initial '%s' does not reference an existing reality", universeKey, *universe.Initial)
			}
		}
	}

	for idx, initial := range model.Initials {
		kind, universeID, realityID, ok := parseStateReference(initial)
		if !ok {
			errCollector.add("initials[%d] has invalid reference '%s'", idx, initial)
			continue
		}

		if kind == referenceReality {
			errCollector.add("initials[%d] must be external reference (U:<universe> or U:<universe>:<reality>), got '%s'", idx, initial)
			continue
		}

		u, exists := model.Universes[universeID]
		if !exists || u == nil {
			errCollector.add("initials[%d] references unknown universe '%s'", idx, universeID)
			continue
		}

		if kind == referenceUniverseReality {
			if _, exists = u.Realities[realityID]; !exists {
				errCollector.add("initials[%d] references unknown reality '%s' in universe '%s'", idx, realityID, universeID)
			}
		}
	}

	for universeID, universe := range model.Universes {
		if universe == nil {
			continue
		}

		for realityID, reality := range universe.Realities {
			if reality == nil {
				continue
			}

			for tIdx, transition := range reality.Always {
				validateTransitionSemantics(errCollector, model, universeID, realityID, "always", tIdx, transition)
			}

			for eventName, transitions := range reality.On {
				for tIdx, transition := range transitions {
					validateTransitionSemantics(errCollector, model, universeID, realityID, "on."+eventName, tIdx, transition)
				}
			}
		}
	}

	if errCollector.hasErrors() {
		return errCollector
	}
	return nil
}

func hasEffectiveTransitionFlow(reality *theoretical.RealityModel) bool {
	if reality == nil {
		return false
	}

	if len(reality.Always) > 0 {
		return true
	}

	if len(reality.On) == 0 {
		return false
	}

	for _, transitions := range reality.On {
		if len(transitions) > 0 {
			return true
		}
	}

	return false
}

func validateTransitionSemantics(
	errCollector *semanticValidationErrors,
	model *theoretical.QuantumMachineModel,
	universeID string,
	realityID string,
	transitionPath string,
	transitionIndex int,
	transition *theoretical.TransitionModel,
) {
	if transition == nil {
		errCollector.add("universe '%s' reality '%s' transition '%s[%d]' cannot be null", universeID, realityID, transitionPath, transitionIndex)
		return
	}

	if len(transition.Targets) == 0 {
		errCollector.add("universe '%s' reality '%s' transition '%s[%d]' must define at least one target", universeID, realityID, transitionPath, transitionIndex)
		return
	}

	isNotify := transition.Type != nil && *transition.Type == theoretical.TransitionTypeNotify

	for targetIndex, target := range transition.Targets {
		kind, targetUniverseID, targetRealityID, ok := parseStateReference(target)
		if !ok {
			errCollector.add(
				"universe '%s' reality '%s' transition '%s[%d]' target[%d] has invalid reference '%s'",
				universeID, realityID, transitionPath, transitionIndex, targetIndex, target,
			)
			continue
		}

		if isNotify && kind == referenceReality {
			errCollector.add(
				"universe '%s' reality '%s' transition '%s[%d]' with type 'notify' cannot target internal reality '%s'",
				universeID, realityID, transitionPath, transitionIndex, target,
			)
		}

		switch kind {
		case referenceReality:
			u := model.Universes[universeID]
			if u == nil {
				errCollector.add("unknown source universe '%s' while validating target '%s'", universeID, target)
				continue
			}
			if _, exists := u.Realities[targetRealityID]; !exists {
				errCollector.add(
					"universe '%s' reality '%s' transition '%s[%d]' target[%d] references unknown internal reality '%s'",
					universeID, realityID, transitionPath, transitionIndex, targetIndex, targetRealityID,
				)
			}
		case referenceUniverse:
			if tu := model.Universes[targetUniverseID]; tu == nil {
				errCollector.add(
					"universe '%s' reality '%s' transition '%s[%d]' target[%d] references unknown universe '%s'",
					universeID, realityID, transitionPath, transitionIndex, targetIndex, targetUniverseID,
				)
			}
		case referenceUniverseReality:
			tu := model.Universes[targetUniverseID]
			if tu == nil {
				errCollector.add(
					"universe '%s' reality '%s' transition '%s[%d]' target[%d] references unknown universe '%s'",
					universeID, realityID, transitionPath, transitionIndex, targetIndex, targetUniverseID,
				)
				continue
			}
			if _, exists := tu.Realities[targetRealityID]; !exists {
				errCollector.add(
					"universe '%s' reality '%s' transition '%s[%d]' target[%d] references unknown reality '%s' in universe '%s'",
					universeID, realityID, transitionPath, transitionIndex, targetIndex, targetRealityID, targetUniverseID,
				)
			}
		}
	}
}

type stateReferenceType int

const (
	referenceUniverse stateReferenceType = iota
	referenceUniverseReality
	referenceReality
)

func parseStateReference(ref string) (stateReferenceType, string, string, bool) {
	if ref == "" {
		return 0, "", "", false
	}

	if strings.HasPrefix(ref, "U:") {
		parts := strings.Split(ref, ":")
		if len(parts) == 2 {
			if !isValidIdentifier(parts[1]) {
				return 0, "", "", false
			}
			return referenceUniverse, parts[1], "", true
		}
		if len(parts) == 3 {
			if !isValidIdentifier(parts[1]) || !isValidIdentifier(parts[2]) {
				return 0, "", "", false
			}
			return referenceUniverseReality, parts[1], parts[2], true
		}
		return 0, "", "", false
	}

	if !isValidIdentifier(ref) {
		return 0, "", "", false
	}
	return referenceReality, "", ref, true
}

func isValidIdentifier(v string) bool {
	if v == "" {
		return false
	}

	if v[0] < 'A' || (v[0] > 'Z' && v[0] < 'a') || v[0] > 'z' {
		return false
	}

	for i := 1; i < len(v); i++ {
		ch := v[i]
		isLetter := (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
		isDigit := ch >= '0' && ch <= '9'
		if !isLetter && !isDigit && ch != '_' && ch != '-' {
			return false
		}
	}

	last := v[len(v)-1]
	isLetter := (last >= 'A' && last <= 'Z') || (last >= 'a' && last <= 'z')
	isDigit := last >= '0' && last <= '9'
	return isLetter || isDigit
}

// Unmarshal helper kept for future callers that validate schema-document payloads directly.
func unmarshalAny(source []byte) (any, error) {
	var payload any
	if err := json.Unmarshal(source, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
