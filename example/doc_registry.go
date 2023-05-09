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

// action -> (dog Dog, evt piece.Event, actTool piece.ActionTool[Dog]) error
// predicate -> (dog Dog, evt Event) (bool, error)
// invocation -> (dog Dog, evt Event) ServiceResponse

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

// 01 - Basic example
//func (DogMachineDefinitions[T]) IncreaseEnergy(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
//	fmt.Println("- (A) Dog is increasing energy")
//	return nil
//}

// 02 - Read example
//func (DogMachineDefinitions[T]) IncreaseEnergy(dog Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
//	fmt.Println("- (A) Dog is increasing energy")
//	fmt.Println("      · Dog name: ", dog.Name)
//	fmt.Println("      · Dog energy level: ", dog.EnergyLevel)
//	return nil
//}

// 03 - Write example
//func (DogMachineDefinitions[T]) IncreaseEnergy(dog Dog, _ piece.Event, actTool piece.ActionTool[Dog]) error {
//	fmt.Println("- (A) Dog is increasing energy")
//	fmt.Println("      · Dog name: ", dog.Name)
//	fmt.Println("      · Dog energy level before eat: ", dog.EnergyLevel)
//	//dog.Eat()
//	actTool.Assign(dog)
//	return nil
//}

// 04 - Events example
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

// 01 - Basic example
//func (DogMachineDefinitions[T]) DecreaseEnergy(_ Dog, _ piece.Event, _ piece.ActionTool[Dog]) error {
//	fmt.Println("- (A) Dog is decreasing energy")
//	return nil
//}

// 04 - Events example
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

// 04 - Events example
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

// 04 - Events example
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

// 04 - Events example
func (DogMachineDefinitions[T]) MoreHungryThanTired(_ Dog, evt piece.Event) (bool, error) {
	response := true
	if evt.HasData() {
		evtData, _ := evt.GetData().(EvtData)
		response = !evtData.tooTired
	}
	fmt.Printf("- (G) Check if dog is more hungry than tired. Dog is more hungry than tired: %v\n", response)
	return response, nil
}
