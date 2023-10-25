package admission

import (
	"context"
	"github.com/rendis/statepro/v3/instrumentation"
)

func NewAdmissionQM() *AdmissionQM {
	return &AdmissionQM{}
}

type AdmissionQM struct {
}

func (a AdmissionQM) GetQuantumMachineId() string {
	return "admission-default-machine"
}

func (a AdmissionQM) GetQuantumMachineDescription() string {
	return "Admission quantum machine laws"
}

func (a AdmissionQM) ExecuteAction(ctx context.Context, args instrumentation.ActionExecutorArgs) error {

	action := args.GetAction()

	switch action.Src {
	case "logEntryToStatus":
		return logEntryToStatus(ctx, args.GetRealityName(), args.GetUniverseCanonicalName())
	case "logExitFromStatus":
		return logExitFromStatus(ctx, args.GetRealityName(), args.GetUniverseCanonicalName())
	default:
		return nil
	}
}

func (a AdmissionQM) ExecuteInvoke(ctx context.Context, args instrumentation.InvokeExecutorArgs) {
	invoke := args.GetInvoke()
	event := args.GetEvent()
	mctx := args.GetContext()

	switch invoke.Src {
	case "notifyStatusChanged":
		notifyStatusChanged(ctx, mctx, event)
	case "logTransition":
		logTransition(ctx, args)
	default:
		return
	}
}
