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
	debugger.SetStateMachinePath("./state_machine.json")
	//debugger.SetStateMachinePath("example/cli/state_machine.json")
	debugger.SetEventsPath("./events.json")
	//debugger.SetEventsPath("example/cli/events.json")
	debugger.SetSnapshotsPath("./snapshots.json")
	//debugger.SetSnapshotsPath("example/cli/snapshots.json")
	debugger.Run(nil)
}

func registerLaws() {
	_ = builtin.RegisterAction("action:disableAdmission", disableAdmission)
	_ = builtin.RegisterAction("action:createForm", createForm)
	_ = builtin.RegisterAction("action:sendContract", sendContract)
	_ = builtin.RegisterCondition("condition:ifItIsTheTemplate", ifItIsTheTemplate)
	_ = builtin.RegisterAction("action:createPayment", createPayment)
}

func disableAdmission(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
	return nil
}

func ifItIsTheTemplate(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
	// get templateId from args
	templateId, ok := getTemplateIdFromArg(args.GetCondition().Args)
	if !ok {
		return false, nil
	}

	// get templateId from event
	eventTemplateIdAny, ok := args.GetEvent().GetData()["templateId"]
	if !ok {
		return false, nil
	}

	fromTemplateId, ok := eventTemplateIdAny.(string)
	if !ok {
		return false, nil
	}

	isTheTemplate := templateId == fromTemplateId
	if !isTheTemplate {
		return false, nil
	}

	return true, nil
}

func getTemplateIdFromArg(args map[string]any) (string, bool) {
	templateIdAny, ok := args["templateId"]
	if !ok {
		return "", false
	}

	templateId, ok := templateIdAny.(string)
	return templateId, ok
}

func createForm(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
	return nil
}

func sendContract(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
	return nil
}

func createPayment(_ context.Context, args instrumentation.ActionExecutorArgs) error {
	const templateId = "payment_generated_id"
	args.UpdateUniverseMetadata("templateId", templateId)
	return nil
}
