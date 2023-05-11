package main

import (
	"fmt"
	"github.com/rendis/statepro"
	"github.com/rendis/statepro/piece"
	"strings"
	"time"
)

func runDogMachineExamples() {
	/*************** DOG MACHINE **************/
	/* Uncomment the example you want to run */
	/*******************************************/

	invokeServices()
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

// invoke services
func invokeServices() {
	dog := &Dog{}
	_, dogMachine := getDogMachine(dog)

	evt := piece.BuildEvent("NOISE").Build()
	_ = dogMachine.SendEvent(evt)

	// invocations are launched on a separate goroutines, so we need to wait a bit before main goroutine exits
	time.Sleep(1 * time.Second)

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
