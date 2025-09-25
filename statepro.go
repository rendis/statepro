package statepro

import (
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

func NewQuantumMachine(qmModel *theoretical.QuantumMachineModel) (instrumentation.QuantumMachine, error) {
	var universes []*experimental.ExUniverse
	for _, model := range qmModel.Universes {
		universes = append(universes, experimental.NewExUniverse(model))
	}
	return experimental.NewExQuantumMachine(qmModel, universes)
}

func NewEventBuilder(eventName string) instrumentation.EventBuilder {
	return experimental.NewEventBuilder(eventName)
}
