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

	sendEventInsideAction()

	//sendDataInEventAndAccessItInAction_Ex1()

	//sendDataInEventAndAccessItInAction_Ex2()

	//actionErrorHandling()
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

// send event inside action
func sendEventInsideAction() {
	dog := &Dog{
		Name:        "GouGou",
		EnergyLevel: 50,
	}
	_, dogMachine := getDogMachine(dog)

	initOnState := "Awake"
	_ = dogMachine.PlaceOn(initOnState)

	printMachineInfo(dogMachine)

	evt := piece.BuildEvent("TIRED").Build()
	_ = dogMachine.SendEvent(evt)

	printMachineInfo(dogMachine)
}

// send data in event and access it in action - Example 1
func sendDataInEventAndAccessItInAction_Ex1() {
	dog := &Dog{}
	_, dogMachine := getDogMachine(dog)

	printMachineInfo(dogMachine)

	evtData := EvtData{
		textToNextState: "I'm tired, I want to sleep",
	}
	evt := piece.BuildEvent("SMELLS").
		WithData(evtData).
		Build()
	resp := dogMachine.SendEvent(evt)

	printMachineInfo(dogMachine)

	lastEvent := resp.GetLastEvent()

	evt = piece.BuildEvent("FOUND_SOMETHING_INTERESTING").
		WithData(lastEvent.GetData()).
		Build()

	resp = dogMachine.SendEvent(evt)

	printMachineInfo(dogMachine)

	lastEvent = resp.GetLastEvent()
	evtData, _ = lastEvent.GetData().(EvtData)
	fmt.Printf("\n- (Last evt) Last message: %s\n", evtData.textToNextState)
}

// send data in event and access it in action - Example 1
func sendDataInEventAndAccessItInAction_Ex2() {
	dog := &Dog{}
	_, dogMachine := getDogMachine(dog)

	initOnState := "Awake"
	_ = dogMachine.PlaceOn(initOnState)
	printMachineInfo(dogMachine)

	evtData := EvtData{
		textToNextState: "I'm awake... just being cute",
		tooTired:        true,
	}
	evt := piece.BuildEvent("TIRED").
		WithData(evtData).
		Build()
	resp := dogMachine.SendEvent(evt)
	printMachineInfo(dogMachine)

	lastEvent := resp.GetLastEvent()
	evtData, _ = lastEvent.GetData().(EvtData)
	fmt.Printf("\n- (Last evt) Last message: %s\n", evtData.textToNextState)
}

// action error handling
func actionErrorHandling() {
	dog := &Dog{}
	_, dogMachine := getDogMachine(dog)
	_ = dogMachine.PlaceOn("Playing")

	printMachineInfo(dogMachine)

	evt := piece.BuildEvent("TIRED").
		Build()

	resp := dogMachine.SendEvent(evt)

	if resp.Error() != nil {
		fmt.Printf("\n- (Error) %s\n", resp.Error().Error())
	}

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
