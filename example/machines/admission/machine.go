package admission

import (
	"context"
	"github.com/rendis/statepro/v3/experimental"
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

func (a AdmissionQM) ExecuteObserver(ctx context.Context, args experimental.ObserverExecutorArgs) (bool, error) {
	return false, nil
}

func (a AdmissionQM) ExecuteAction(ctx context.Context, args experimental.ActionExecutorArgs) error {

	action := args.GetAction()

	switch action.Src {
	case "logEntryToStatus":
		return logEntryToStatus(ctx, args.GetRealityName(), args.GetUniverseName())
	case "logExitFromStatus":
		return logExitFromStatus(ctx, args.GetRealityName(), args.GetUniverseName())
	default:
		return nil
	}
}

func (a AdmissionQM) ExecuteInvoke(ctx context.Context, args experimental.InvokeExecutorArgs) {
	invoke := args.GetInvoke()
	event := args.GetEvent()
	mctx := args.GetContext()

	switch invoke.Src {
	case "notifyStatusChanged":
		notifyStatusChanged(ctx, mctx, event)
	default:
		return
	}
}
