package admission

import (
	"context"
	"fmt"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/theoretical"
	"log"
)

func NewAdmissionQM() *AdmissionQM {
	return &AdmissionQM{}
}

type AdmissionQM struct {
}

func (a AdmissionQM) GetQuantumMachineId() string {
	return "admission_default_machine"
}

func (a AdmissionQM) GetQuantumMachineDescription() string {
	return "Admission quantum machine laws"
}

func (a AdmissionQM) ExecuteObserver(
	ctx context.Context,
	quantumMachineContext any,
	accumulatorStatistics experimental.AccumulatorStatistics,
	event experimental.Event,
	observer theoretical.ObserverModel,
) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (a AdmissionQM) ExecuteAction(
	ctx context.Context, quantumMachineContext any, event experimental.Event, action theoretical.ActionModel,
) error {
	switch action.Src {
	case "logEntryToStatus":
		return logEntryToStatus(ctx)
	case "logExitFromStatus":
		return logExitFromStatus(ctx)
	default:
		log.Printf("ERROR: action not found. Action name: '%s'\n", action.Src)
		return fmt.Errorf("action not found. Action name: '%s'", action.Src)
	}
}

func (a AdmissionQM) ExecuteInvoke(
	ctx context.Context, quantumMachineContext any, event experimental.Event, invoke theoretical.InvokeModel,
) {
	switch invoke.Src {
	case "notifyStatusChanged":
		notifyStatusChanged(ctx, quantumMachineContext, event)
	default:
		log.Printf("ERROR: invoke not found. Invoke name: '%s'\n", invoke.Src)
	}
}
