package main

import (
	"context"
	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/instrumentation"
	"log"

	"os"
)

func main() {
	setLaws()
	qm := loadDefinition()
	ctx := context.Background()
	machineCtx := &AdmissionQMContext{}

	err := qm.Init(ctx, machineCtx)
	if err != nil {
		log.Fatal(err)
	}

	ss := qm.GetSnapshot()

	// confirm
	event := statepro.NewEventBuilder("confirmed").Build()

	if _, err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	ss = qm.GetSnapshot()

	if _, err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	ss = qm.GetSnapshot()

	//sing
	event = statepro.NewEventBuilder("signed").Build()

	if _, err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	ss = qm.GetSnapshot()

	qm2 := loadDefinition()
	if err = qm2.LoadSnapshot(ss, machineCtx); err != nil {
		log.Fatal(err)
	}

	//fill
	event = statepro.NewEventBuilder("filled-form").Build()

	if _, err = qm2.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	ss = qm2.GetSnapshot()

	//paid
	event = statepro.NewEventBuilder("paid").Build()

	if _, err = qm2.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	ss = qm2.GetSnapshot()

	tracking := ss.GetTracking()
	log.Printf("tracking: %v", tracking)

	log.Println("*************** DONE ***************")

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

func setLaws() {
	_ = builtin.RegisterAction("logEntryToStatus", logEntryToStatusAction)

	_ = builtin.RegisterInvoke("notifyStatusChanged", notifyStatusChangedInvk)
}
