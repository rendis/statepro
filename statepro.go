package statepro

import (
	"fmt"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

func NewQuantumMachine(qmModel *theoretical.QuantumMachineModel) (instrumentation.QuantumMachine, error) {
	if qmModel == nil {
		return nil, fmt.Errorf("quantum machine model must not be nil")
	}
	var universes []*experimental.ExUniverse
	for id, model := range qmModel.Universes {
		if model == nil {
			return nil, fmt.Errorf("universe '%s' model must not be nil", id)
		}
		universes = append(universes, experimental.NewExUniverse(model))
	}
	return experimental.NewExQuantumMachine(qmModel, universes)
}

func NewEventBuilder(eventName string) instrumentation.EventBuilder {
	return experimental.NewEventBuilder(eventName)
}
