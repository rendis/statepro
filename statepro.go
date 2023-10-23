package statepro

import (
	"fmt"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

func NewQuantumMachine(qmModel *theoretical.QuantumMachineModel) (instrumentation.QuantumMachine, error) {
	qmLaws := GetQuantumMachineLaws(qmModel.ID)
	if qmLaws == nil {
		return nil, fmt.Errorf("quantum machine laws not found. QuantumMachineId: '%s'", qmModel.ID)
	}

	var universes []*experimental.ExUniverse
	for _, model := range qmModel.Universes {
		laws := GetUniverseLaws(model.ID)
		if laws == nil {
			return nil, fmt.Errorf("laws not found for universe id '%s'", model.ID)
		}

		universe := experimental.NewExUniverse(model.ID, model, laws)
		universes = append(universes, universe)
	}

	return experimental.NewExQuantumMachine(qmModel, qmLaws, universes)
}

func NewEvent(eventName string, data map[string]any, evtType instrumentation.EventType) instrumentation.Event {
	return experimental.NewEvent(eventName, data, evtType)
}

func NewEventBuilder(eventName string) instrumentation.EventBuilder {
	return experimental.NewEventBuilder(eventName)
}
