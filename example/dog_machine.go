package main

import (
	"fmt"
	"github.com/rendis/statepro"
	"github.com/rendis/statepro/piece"
	"strings"
	"time"
)

func runDogMachineExamples() {
	// ------- 01-example-basic -----
	//initDogMachine()

	//_, _ = getDogMachine(nil)

	//getDogMachineInfo()

	//dogMachineSendEvent()

	//initDogMachineOnState()

	//selfEventBehavior()

	// ------- 02-read-basic-example -----
	//accessContextAndEventValueBasic()

	// ------- 03-write-basic-example -----
	//updateContextValue()

	// ------- 04-events-example -----
	//sendEventInsideAction()

	//sendDataInEventAndAccessItInAction_Ex1()

	//sendDataInEventAndAccessItInAction_Ex2()

	//actionErrorHandling()

	// ------- 05-invocations-services -----
	invokeServices()
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

// update context value
func updateContextValue() {
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

	evt = piece.BuildEvent("FULL").Build()
	_ = dogMachine.SendEvent(evt)

	context := dogMachine.GetContext()
	fmt.Printf("\n- (Current ctx) Dog energy level after eat: %d\n", context.EnergyLevel)

	printMachineInfo(dogMachine)
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
