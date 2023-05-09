package main

import (
	"fmt"
	"github.com/rendis/statepro"
	"github.com/rendis/statepro/piece"
	"strings"
)

func runDogMachineExamples() {
	// ------- 01-example-basic -----
	//initDogMachine()

	//_, _ = getDogMachine(nil)

	//getDogMachineInfo()

	//dogMachineSendEvent()

	//initDogMachineOnState()

	//selfEventBehavior()

	// ------- 02-read-example-basic -----
	accessContextAndEventValueBasic()
}

// show how to register a machine and init all machines
func initDogMachine() {
	definitions := &DogMachineDefinitions[Dog]{}
	_ = statepro.AddMachine[Dog](definitions)
	statepro.InitMachines()
	fmt.Println("DogMachine registered and initialized")
}

// show how to get a machine by id
func getDogMachine(dog *Dog) (string, piece.ProMachine[Dog]) {
	definitions := &DogMachineDefinitions[Dog]{}
	dogMachineId := statepro.AddMachine[Dog](definitions)
	statepro.InitMachines()

	if dog == nil {
		dog = &Dog{}
	}
	dogMachine, err := statepro.GetMachineById[Dog](dogMachineId, dog)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nDogMachineId: %s\n\n", dogMachineId)

	return dogMachineId, dogMachine
}

// show how to get machine info
func getDogMachineInfo() {
	dog := &Dog{}
	_, dogMachine := getDogMachine(dog)

	nextEvents := dogMachine.GetNextEvents()
	fmt.Printf("DogMachine next events: %v\n", nextEvents)

	isFinalState := dogMachine.IsFinalState()
	fmt.Printf("DogMachine is final state: %v\n", isFinalState)

	currentState := dogMachine.GetState()
	fmt.Printf("DogMachine current state: %s\n", currentState)
}

// show how to build and send an event
func dogMachineSendEvent() {
	dog := &Dog{}
	_, dogMachine := getDogMachine(dog)

	evt := piece.BuildEvent("NOISE").Build()
	_ = dogMachine.SendEvent(evt)

	printMachineInfo(dogMachine)
}

// show how to init a machine on a specific state
func initDogMachineOnState() {
	dog := &Dog{}
	_, dogMachine := getDogMachine(dog)

	initOnState := "Awake"
	_ = dogMachine.PlaceOn(initOnState)

	printMachineInfo(dogMachine)

	evt := piece.BuildEvent("TIRED").Build()
	_ = dogMachine.SendEvent(evt)

	printMachineInfo(dogMachine)
}

// show it-self event behavior
func selfEventBehavior() {
	dog := &Dog{}
	_, dogMachine := getDogMachine(dog)

	evt := piece.BuildEvent("DREAM").Build()
	_ = dogMachine.SendEvent(evt)

	printMachineInfo(dogMachine)
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
