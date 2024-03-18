package main

import (
	"context"
	"fmt"
	"github.com/rendis/abslog/v3"
	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/instrumentation"
	"log"
	"os"
)

func main() {
	registerLaws()
	qm := loadDefinition()
	ctx := context.Background()
	machineCtx := &AdmissionQMContext{
		Active: true,
	}

	err := qm.Init(ctx, machineCtx)
	if err != nil {
		log.Fatal(err)
	}

	// confirm
	event := statepro.NewEventBuilder("confirm").Build()
	if _, err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	snapshot := qm.GetSnapshot()
	log.Printf("Snapshot generated: %+v", snapshot.Tracking)

	// fill-form
	event = statepro.NewEventBuilder("fill-form").Build()
	if _, err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	snapshot = qm.GetSnapshot()
	log.Printf("Snapshot generated: %+v", snapshot.Tracking)

	// sign
	event = statepro.NewEventBuilder("sign").Build()
	if _, err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	snapshot = qm.GetSnapshot()
	log.Printf("Snapshot generated: %+v", snapshot.Tracking)

	// sign
	event = statepro.NewEventBuilder("sign").Build()
	if _, err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	snapshot = qm.GetSnapshot()
	log.Printf("Snapshot generated: %+v", snapshot.Tracking)

	// cancel
	event = statepro.NewEventBuilder("cancel").Build()
	if _, err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	snapshot = qm.GetSnapshot()
	log.Printf("Snapshot generated: %+v", snapshot.Tracking)

	// reply cancel entry actions
	if err = qm.ReplayOnEntry(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("done")
}

func loadDefinition() instrumentation.QuantumMachine {
	var path = "example/v0.1.json"

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
	if err := builtin.RegisterAction("disableAdmission", DisableAdmission); err != nil {
		abslog.Fatalf("failed to register custom action '%s'", "disableAdmission", err)
	}
}

func DisableAdmission(ctx context.Context, args instrumentation.ActionExecutorArgs) error {
	adm, err := toContext(ctx, args.GetContext())
	if err != nil {
		return err
	}

	adm.Active = false
	return nil
}

func toContext(ctx context.Context, admissionCtx any) (*AdmissionQMContext, error) {
	admission, ok := admissionCtx.(*AdmissionQMContext)
	if !ok {
		abslog.ErrorCtx(ctx, "context is not an admission")
		return nil, fmt.Errorf("context is not an admission")
	}

	return admission, nil
}
