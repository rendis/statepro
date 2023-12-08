package main

import (
	"context"
	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/instrumentation"
	"log"

	"os"
)

func main() {
	qm := loadDefinition()
	ctx := context.Background()
	machineCtx := &AdmissionQMContext{}

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
