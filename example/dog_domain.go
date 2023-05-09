package main

import (
	"fmt"
	"time"
)

type DogState int

const (
	Sleeping DogState = iota
	Awake
	Eating
	Playing
	Sniffing
	Investigating
)

func (ds DogState) String() string {
	return [...]string{"Sleeping", "Awake", "Eating", "Playing", "Sniffing", "Investigating"}[ds]
}

type Dog struct {
	Name         string
	EnergyLevel  int
	IsHungry     bool
	IsTired      bool
	CurrentState DogState
}

func (d *Dog) Eat() {
	fmt.Printf("      · (M) %s is eating\n", d.Name)
	d.IsHungry = false
	d.EnergyLevel = 100
	d.CurrentState = Eating
}

func (d *Dog) Sleep() {
	fmt.Printf("      · (M) %s is sleeping\n", d.Name)
	d.IsTired = false
	d.EnergyLevel += 10
	d.CurrentState = Sleeping
	time.Sleep(1 * time.Second)
}

func (d *Dog) Play() {
	fmt.Printf("      · (M) %s is playing\n", d.Name)
	d.EnergyLevel -= 5
	if d.EnergyLevel <= 5 {
		d.IsTired = true
	}
	d.CurrentState = Playing
}

// ------------------------------------

type EvtData struct {
	date            time.Time
	textToNextState string
}
