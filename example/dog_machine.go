package main

import (
	"fmt"
	"github.com/rendis/statepro"
	"github.com/rendis/statepro/piece"
	"strings"
)

func runDogMachineExamples() {
	/*************** DOG MACHINE **************/
	/* Uncomment the example you want to run */
	/*******************************************/
	statepro.SetDefinitionPath("example/statepro.yml")

	accessContextAndEventValueBasic()
}

// show how to get a machine by id
func getDogMachine(dog *Dog) (string, piece.ProMachine[Dog]) {
	definitions := &DogMachineDefinitions[Dog]{}
	dogMachineId := statepro.AddMachine[Dog](definitions)
	statepro.InitMachines()

	if dog == nil {
		dog = &Dog{}
	}
	dogMachine, err := statepro.GetMachine[Dog](dogMachineId, dog)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nDogMachineId: %s\n\n", dogMachineId)

	return dogMachineId, dogMachine
}

// access to context and event value, basic
func accessContextAndEventValueBasic() {
	dog := &Dog{
		Name:        "GouGou",
		EnergyLevel: 50,
	}
	_, dogMachine := getDogMachine(dog)

	initOnState := "Awake"
	_ = dogMachine.PlaceOn(initOnState)

	printMachineInfo(dogMachine)

	evt := piece.BuildEvent("HUNGRY").Build()
	_ = dogMachine.SendEvent(evt)

	printMachineInfo(dogMachine)
}

func printMachineInfo(machine piece.ProMachine[Dog]) {
	fmt.Println("=========================================")

	currentState := machine.GetState()
	fmt.Printf("* DogMachine current state: %s\n", currentState)

	isFinalState := machine.IsFinalState()
	fmt.Printf("* DogMachine is final state: %v\n", isFinalState)

	nextEvents := machine.GetNextEvents()
	// join next events by comma
	nextEventsStr := strings.Join(nextEvents, ", ")
	fmt.Printf("* DogMachine next events: %v\n", nextEventsStr)
	fmt.Println("=========================================")
}
