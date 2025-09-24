package main

import (
	"context"
	"os"

	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/debugger/bot"
	"github.com/rendis/statepro/v3/instrumentation"
)

func main() {
	registerLaws()

	ctx := context.Background()
	qm := loadDefinition()

	var events = []instrumentation.Event{
		statepro.NewEventBuilder("confirm").Build(),
	}

	var count = 0
	var provider bot.EventProvider = func(_ *instrumentation.MachineSnapshot) (instrumentation.Event, error) {
		if count >= len(events) {
			return nil, nil
		}

		event := events[count]
		count++
		return event, nil
	}

	b, err := bot.NewBot(qm, provider, true)
	if err != nil {
		panic(err)
	}

	err = b.Run(ctx, nil)
	if err != nil {
		panic(err)
	}
}

func loadDefinition() instrumentation.QuantumMachine {
	var path = "example/bot/state_machine.json"

	arrByte, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	tmDef, err := statepro.DeserializeQuantumMachineFromBinary(arrByte)
	if err != nil {
		panic(err)
	}

	qm, err := statepro.NewQuantumMachine(tmDef)
	if err != nil {
		panic(err)
	}

	return qm
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
	args.AddToUniverseMetadata("templateId", templateId)
	return nil
}
