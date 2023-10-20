package main

import (
	"fmt"
	"github.com/rendis/statepro/v3"
	"github.com/rendis/statepro/v3/example/machines/admission"
	"github.com/rendis/statepro/v3/example/universes/contract"
	"github.com/rendis/statepro/v3/theoretical"

	"os"
)

func main() {

	var path = "example/definitions/v0.1.json"

	arrByte, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	tmDef, err := theoretical.BuildQuantumMachineModelFromJSONDefinition(arrByte)
	if err != nil {
		panic(err)
	}

	qml1 := admission.NewAdmissionQMLink()

	if err := statepro.RegisterQuantumMachineLaws(qml1); err != nil {
		panic(err)
	}

	u1 := contract.NewContractUniverse()

	if err := statepro.RegisterUniverseLaws(u1); err != nil {
		panic(err)
	}

	qm, err := statepro.BuildQuantumMachine(tmDef)
	if err != nil {
		panic(err)
	}

	fmt.Println(qm)
}
