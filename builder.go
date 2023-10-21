package statepro

import (
	"fmt"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/theoretical"
)

func BuildExperimentalQuantumMachine(qmModel *theoretical.QuantumMachineModel) (*experimental.ExQuantumMachine, error) {
	qmId := buildUniqueId(qmModel.ID, qmModel.Version)
	qmLaws := GetQuantumMachineLaws(qmId)
	if qmLaws == nil {
		return nil, fmt.Errorf("quantum machine laws not found. QuantumMachineId: '%s'", qmId)
	}

	var universes []*experimental.ExUniverse
	for _, model := range qmModel.Universes {
		universeId := buildUniqueId(model.ID, model.Version)
		laws := GetUniverseLaws(universeId)
		if laws == nil {
			return nil, fmt.Errorf("laws not found for universe id '%s'", universeId)
		}

		universe := experimental.NewExUniverse(universeId, model, laws)
		universes = append(universes, universe)
	}

	return experimental.NewExQuantumMachine(qmModel, qmLaws, universes)
}

func buildUniqueId(parts ...string) string {
	var id string
	for _, part := range parts {
		id += part + "@"
	}
	return id[:len(id)-1]
}
