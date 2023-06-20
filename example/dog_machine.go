package main

import (
	"fmt"
	"github.com/rendis/statepro"
	"github.com/rendis/statepro/piece"
	"os"
	"strings"
)

const (
	machineId      = "DogMachine"
	machineVersion = "1.0"
)

func runDogMachineExamples() {
	/*************** DOG MACHINE **************/
	/* Uncomment the example you want to run */
	/*******************************************/

	initDogMachine()

	//_, _ = getDogMachine(nil)

	//getDogMachineFromMap(nil)

	//getDogMachineInfo()

	//dogMachineSendEvent()

	//initDogMachineOnState()

	//selfEventBehavior()
}

// show how to register a machine and init all machines
func initDogMachine() {
	statepro.SetDefinitionPath("example/statepro.yml")
	definitions := &DogMachineDefinitions[Dog]{}
	_ = statepro.AddMachine[Dog](machineId, machineVersion, definitions)
	statepro.InitMachines()
	fmt.Println("DogMachine registered and initialized")
}

// show how to get a machine by id
func getDogMachine(dog *Dog) (string, piece.ProMachine[Dog]) {
	statepro.SetDefinitionPath("example/statepro.yml")
	definitions := &DogMachineDefinitions[Dog]{}
	dogMachineId := statepro.AddMachine[Dog](machineId, machineVersion, definitions)
	statepro.InitMachines()

	if dog == nil {
		dog = &Dog{}
	}
	mui := statepro.BuildMachineCompositeId(machineId, machineVersion)
	dogMachine, err := statepro.GetMachineByUniqueId[Dog](mui, dog)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nDogMachineId: %s\n\n", dogMachineId)

	return dogMachineId, dogMachine
}

// show how to add json machine definition as []byte
func getDogMachineFromMap(dog *Dog) (string, piece.ProMachine[Dog]) {
	// read json definition from file
	jsonDefinition, err := os.ReadFile("example/source/01_dog_machine.json")
	if err != nil {
		fmt.Printf("error reading json definition: %s\n", err.Error())
		panic(err)
	}

	var jsonMap = make(map[string]string)
	jsonMap["01_dog_machine"] = string(jsonDefinition)
	statepro.AddDefinitions(jsonMap)

	dogDefinitions := &DogMachineDefinitions[Dog]{}
	dogMachineId := statepro.AddMachine[Dog](machineId, machineVersion, dogDefinitions)

	statepro.InitMachines()

	if dog == nil {
		dog = &Dog{}
	}
	mui := statepro.BuildMachineCompositeId(machineId, machineVersion)
	dogMachine, err := statepro.GetMachineByUniqueId[Dog](mui, dog)
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

	currentState := dogMachine.GetState()
	fmt.Printf("DogMachine current state: %s\n", currentState)

	nextEvents := dogMachine.GetNextEvents()
	fmt.Printf("DogMachine next events: %v\n", nextEvents)

	isFinalState := dogMachine.IsFinalState()
	fmt.Printf("DogMachine is final state: %v\n", isFinalState)
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
	//_, dogMachine := getDogMachine(dog)
	_, dogMachine := getDogMachineFromMap(dog)

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
