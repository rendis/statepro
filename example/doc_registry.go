package main

import (
	"fmt"
	"github.com/rendis/statepro/piece"
)

type DogMachineDefinitions[T Dog] struct{}

func (DogMachineDefinitions[T]) GetMachineTemplateId() string {
	return "DogMachine"
}

// Actions definitions
// action -> (dog Dog, evt Event, actTool ActionTool[Dog]) error

func (DogMachineDefinitions[T]) NotifySleeping(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is sleeping")
	return nil
}

// 03 - Write example
func (DogMachineDefinitions[T]) IncreaseEnergy(dog Dog, _ piece.Event, actTool piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is increasing energy")
	fmt.Println("      · Dog name: ", dog.Name)
	fmt.Println("      · Dog energy level before eat: ", dog.EnergyLevel)
	dog.Eat()
	actTool.Assign(dog)
	return nil
}

func (DogMachineDefinitions[T]) NotifyMovement(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is moving")
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
	return nil
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

func (DogMachineDefinitions[T]) StartSniffing(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is sniffing")
	return nil
}

func (DogMachineDefinitions[T]) StopSniffing(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog stopped sniffing")
	return nil
}

func (DogMachineDefinitions[T]) Investigate(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is investigating")
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

func (DogMachineDefinitions[T]) GoToEat(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is going to eat")
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

func (DogMachineDefinitions[T]) MoreHungryThanTired(_ Dog, _ piece.Event) (bool, error) {
	response := true
	fmt.Printf("- (G) Check if dog is more hungry than tired. Dog is more hungry than tired: %v\n", response)
	return response, nil
}
