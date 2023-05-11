package main

import (
	"errors"
	"fmt"
	"github.com/rendis/statepro/piece"
)

type DogMachineDefinitions[T Dog] struct{}

func (DogMachineDefinitions[T]) GetMachineTemplateId() string {
	return "DogMachine"
}

// Actions definitions
// action -> (dog Dog, evt Event, actTool ActionTool[Dog]) error

func (DogMachineDefinitions[T]) NotifySleeping(_ Dog, evt piece.Event, actTool piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is sleeping")

	evtData, _ := evt.GetData().(EvtData)
	msgFromPrevState := evtData.textToNextState
	fmt.Printf("      · Message from state (%s): %s\n", evt.GetFrom(), msgFromPrevState)

	evtData.textToNextState = "I'm going to sleep"
	evt = evt.ToBuilder().WithData(evtData).Build()
	actTool.Propagate(evt)
	return nil
}

func (DogMachineDefinitions[T]) IncreaseEnergy(dog Dog, evt piece.Event, actTool piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is increasing energy")
	fmt.Println("      · Dog name: ", dog.Name)
	fmt.Println("      · Dog energy level before eat: ", dog.EnergyLevel)
	actTool.Assign(dog)

	evtData, _ := evt.GetData().(EvtData)
	msgFromPrevState := evtData.textToNextState
	fmt.Printf("      · Message from state (%s): %s\n", evt.GetFrom(), msgFromPrevState)

	evtData.textToNextState = "While I'm sleeping I'm going to increase my energy"
	evt = evt.ToBuilder().WithData(evtData).Build()
	actTool.Propagate(evt)

	return nil
}

func (DogMachineDefinitions[T]) NotifyMovement(dog Dog, _ piece.Event, actTool piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is moving")
	dog.EnergyLevel -= 10

	fmt.Println("      · Decreasing energy level by 10")
	actTool.Assign(dog)
	return nil
}

func (DogMachineDefinitions[T]) StartEating(dog Dog, evt piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is eating")
	fmt.Println("      · Dog name: ", dog.Name)
	fmt.Println("      · Dog energy level: ", dog.EnergyLevel)
	fmt.Println("      · Event name: ", evt.GetName())
	fmt.Println("      · Event from: ", evt.GetFrom())
	fmt.Println("      · Event type: ", evt.GetEvtType())
	return nil
}

func (DogMachineDefinitions[T]) StopEating(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog stopped eating")
	return nil
}

func (DogMachineDefinitions[T]) StartPlaying(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is playing")
	return nil
}

func (DogMachineDefinitions[T]) StopPlaying(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog stopped playing")
	return nil
}

func (DogMachineDefinitions[T]) DecreaseEnergy(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is decreasing energy")
	return errors.New("the dog is too tired to go to bed (he will stay in the kitchen)")
}

func (DogMachineDefinitions[T]) StartInvestigating(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is investigating")
	return nil
}

func (DogMachineDefinitions[T]) StopInvestigating(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog stopped investigating")
	return nil
}

func (DogMachineDefinitions[T]) LoseInterest(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog lost interest")
	return nil
}

func (DogMachineDefinitions[T]) StartSniffing(_ Dog, evt piece.Event, actTool piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is sniffing")

	evtData, _ := evt.GetData().(EvtData)
	msgFromPrevState := evtData.textToNextState
	fmt.Printf("      · Message from state (%s): %s\n", evt.GetFrom(), msgFromPrevState)

	evtData.textToNextState = "I'm a dog and I'm sniffing"
	evt = evt.ToBuilder().WithData(evtData).Build()
	actTool.Propagate(evt)

	return nil
}

func (DogMachineDefinitions[T]) StopSniffing(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog stopped sniffing")
	return nil
}

func (DogMachineDefinitions[T]) Investigate(_ Dog, evt piece.Event, actTool piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is investigating")

	evtData, _ := evt.GetData().(EvtData)
	msgFromPrevState := evtData.textToNextState
	fmt.Printf("      · Message from state (%s): %s\n", evt.GetFrom(), msgFromPrevState)

	evtData.textToNextState = "I'm a dog and I'm investigating"
	evt = evt.ToBuilder().WithData(evtData).Build()
	actTool.Propagate(evt)

	return nil
}

func (DogMachineDefinitions[T]) Alert(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) A loud noise was heard. Dog is alert")
	return nil
}

func (DogMachineDefinitions[T]) HaveADream(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is dreaming with a bone")
	return nil
}

func (DogMachineDefinitions[T]) GoToEat(_ Dog, _ piece.Event, actTool piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog more hungry than tired. Dog is going to eat")
	goToEatEvt := piece.BuildEvent("HUNGRY").Build()
	actTool.Send(goToEatEvt)
	return nil
}

// Invocations definitions
// invocation -> (dog Dog, evt Event)

func (DogMachineDefinitions[T]) SendAppNotification(_ Dog, _ piece.Event) {
	fmt.Printf("- (I) Send app notification\n")
}

func (DogMachineDefinitions[T]) TurnOnLights(_ Dog, _ piece.Event) {
	fmt.Printf("- (I) Turn on lights\n")
}

// Guards (predicates) definitions
// predicate -> (dog Dog, evt Event) (bool, error)

func (DogMachineDefinitions[T]) IsLoudNoise(_ Dog, _ piece.Event) (bool, error) {
	response := true
	fmt.Printf("- (G) Check if noise is loud. Noise is loud: %v\n", response)
	return response, nil
}

func (DogMachineDefinitions[T]) IsEnergyLow(_ Dog, _ piece.Event) (bool, error) {
	response := true
	fmt.Printf("- (G) Check if energy is low. Energy is low: %v\n", response)
	return response, nil
}

func (DogMachineDefinitions[T]) MoreHungryThanTired(_ Dog, evt piece.Event) (bool, error) {
	response := true
	if evt.HasData() {
		evtData, _ := evt.GetData().(EvtData)
		response = !evtData.tooTired
	}
	fmt.Printf("- (G) Check if dog is more hungry than tired. Dog is more hungry than tired: %v\n", response)
	return response, nil
}

// Context from/to source
// ContextFromSource -> (params ... any) (Dog, error)
// ContextToSource -> (dog Dog) error

func (DogMachineDefinitions[T]) ContextFromSource(params ...any) (Dog, error) {
	fmt.Println("- (C) Context from source")
	dog := Dog{
		Name:        "Bobby",
		EnergyLevel: 50,
	}

	fmt.Printf("      · Dog got from source. Dog name: %s, energy level: %d\n", dog.Name, dog.EnergyLevel)

	return dog, nil
}

func (DogMachineDefinitions[T]) ContextToSource(dog Dog) error {
	fmt.Printf("- (C) Save context to source. Dog name: %s, energy level: %d\n", dog.Name, dog.EnergyLevel)
	return nil
}
