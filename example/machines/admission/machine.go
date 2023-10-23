package admission

import (
	"context"
	"fmt"
	"github.com/rendis/statepro/v3/experimental"
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

func (a AdmissionQM) ExecuteObserver(ctx context.Context, args experimental.ObserverExecutorArgs) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (a AdmissionQM) ExecuteAction(ctx context.Context, args experimental.ActionExecutorArgs) error {

	action := args.GetAction()

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

func (a AdmissionQM) ExecuteInvoke(ctx context.Context, args experimental.InvokeExecutorArgs) {
	invoke := args.GetInvoke()
	event := args.GetEvent()
	mctx := args.GetContext()

	switch invoke.Src {
	case "notifyStatusChanged":
		notifyStatusChanged(ctx, mctx, event)
	default:
		log.Printf("ERROR: invoke not found. Invoke name: '%s'\n", invoke.Src)
	}
}
