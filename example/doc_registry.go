package main

import (
	"fmt"
	"github.com/rendis/statepro/piece"
)

type DogMachineDefinitions[T Dog] struct{}

func (DogMachineDefinitions[T]) GetMachineTemplateId() string {
	return "DogMachine"
}

// action -> (dog Dog, evt piece.Event, actTool piece.ActionTool[Dog]) error
// predicate -> (dog Dog, evt Event) (bool, error)
// invocation -> (dog Dog, evt Event) ServiceResponse

func (DogMachineDefinitions[T]) NotifySleeping(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is sleeping")
	return nil
}

// 01 - Basic example
//func (DogMachineDefinitions[T]) IncreaseEnergy(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
//	fmt.Println("- (A) Dog is increasing energy")
//	return nil
//}

// 02 - Read example
func (DogMachineDefinitions[T]) IncreaseEnergy(dog Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is increasing energy")
	fmt.Println("      · Dog name: ", dog.Name)
	fmt.Println("      · Dog energy level: ", dog.EnergyLevel)
	return nil
}

func (DogMachineDefinitions[T]) NotifyMovement(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
	fmt.Println("- (A) Dog is moving")
	return nil
}

// 01 - Basic example
//func (DogMachineDefinitions[T]) StartEating(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
//	fmt.Println("- (A) Dog is eating")
//	return nil
//}

// 02 - Read example
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
