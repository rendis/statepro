package statepro

import (
	"github.com/rendis/statepro/experimental"
	"github.com/rendis/statepro/instrumentation"
	"github.com/rendis/statepro/theoretical"
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
