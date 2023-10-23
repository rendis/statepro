package main

import (
	"context"
	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/example/domain"
	"github.com/rendis/statepro/v3/example/machines/admission"
	admissionUniverse "github.com/rendis/statepro/v3/example/universes/admission"
	"github.com/rendis/statepro/v3/example/universes/form"
	"github.com/rendis/statepro/v3/example/universes/payment"
	"github.com/rendis/statepro/v3/example/universes/sign"
	"github.com/rendis/statepro/v3/experimental"
	"github.com/rendis/statepro/v3/theoretical"
	"log"

	"os"
)

func main() {
	setLaws()
	qm := loadDefinition()
	ctx := context.Background()
	machineCtx := &domain.AdmissionQMContext{}

	err := qm.Init(ctx, machineCtx)
	if err != nil {
		log.Fatal(err)
	}

	// confirm
	event, _ := experimental.NewEventBuilder().
		SetEventName("confirm").
		Build()

	if err = qm.SendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	//sing
	event, _ = experimental.NewEventBuilder().
		SetEventName("sign").
		Build()

	if err = qm.LazySendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	ss := qm.GetSnapshot()

	qm2 := loadDefinition()
	if err = qm2.LoadSnapshot(ss); err != nil {
		log.Fatal(err)
	}

	//fill
	event, _ = experimental.NewEventBuilder().
		SetEventName("fill").
		Build()

	if err = qm2.LazySendEvent(ctx, event); err != nil {
		log.Fatal(err)
	}

	log.Println("done")

}

func loadDefinition() statepro.QuantumMachine {
	var path = "example/definitions/v0.1.json"

	arrByte, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	tmDef, err := theoretical.BuildQuantumMachineModelFromJSONDefinition(arrByte)
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
	qm := admission.NewAdmissionQM()
	au := admissionUniverse.NewAdmissionUniverse()
	fu := form.NewFormUniverse()
	su := sign.NewSignUniverse()
	pu := payment.NewPaymentUniverse()

	if err := statepro.RegisterQuantumMachineLaws(qm); err != nil {
		panic(err)
	}

	if err := statepro.RegisterUniverseLaws(au); err != nil {
		panic(err)
	}

	if err := statepro.RegisterUniverseLaws(fu); err != nil {
		panic(err)
	}

	if err := statepro.RegisterUniverseLaws(su); err != nil {
		panic(err)
	}

	if err := statepro.RegisterUniverseLaws(pu); err != nil {
		panic(err)
	}
}
