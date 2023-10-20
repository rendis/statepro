package admission

import (
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/theoretical"
)

func NewAdmissionQMLink() experimental.QuantumMachineLaws {
	return &AdmissionQMLink{}
}

type AdmissionQMLink struct {
	universes map[string]experimental.UniverseLaws
}

func (q *AdmissionQMLink) GetQuantumMachineId() string {
	return "1"
}

func (q *AdmissionQMLink) GetQuantumMachineVersion() string {
	return "0.1"
}

func (q *AdmissionQMLink) GetQuantumMachineDescription() string {
	return "AdmissionQM"
}

func (q *AdmissionQMLink) ExecuteObserver(
	quantumMachineContext any,
	accumulatorStatistics experimental.AccumulatorStatistics,
	event experimental.Event,
	observer theoretical.ObserverModel,
) (bool, error) {
	return false, nil
}

func (q *AdmissionQMLink) ExecuteAction(
	quantumMachineContext any,
	event experimental.Event,
	action theoretical.ActionModel,
) error {
	return nil
}

func (q *AdmissionQMLink) ExecuteInvoke(
	quantumMachineContext any,
	event experimental.Event,
	invoke theoretical.InvokeModel,
) error {
	return nil
}
