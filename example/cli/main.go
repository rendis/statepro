package main

import (
	"context"
	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/debugger/cli"
	"github.com/rendis/statepro/v3/instrumentation"
)

func main() {
	registerLaws()
	debugger := cli.NewStateMachineDebugger()
	debugger.SetStateMachinePath("example/cli/state_machine.json")
	debugger.SetEventsPath("example/cli/events.json")
	debugger.Run(nil)
}

func registerLaws() {
	_ = builtin.RegisterAction("disableAdmission", DisableAdmission)
}

func DisableAdmission(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
	return nil
}
