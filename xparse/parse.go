package xparse

import (
	"encoding/json"
	"fmt"
	"github.com/rendis/statepro/piece"
	"io/ioutil"
)

func LoadXFile[CTX any](file string) (*piece.GMachine[CTX], error) {
	xm := &XMachine{}

	byteArr, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byteArr, xm)
	if err != nil {
		return nil, err
	}

	return parseXMachine[CTX](xm)
}

func parseXMachine[CTX any](x *XMachine) (*piece.GMachine[CTX], error) {
	g := piece.GMachine[CTX]{}
	err := validateXMachine(x)
	if err != nil {
		return nil, err
	}

	g.Id = *x.Id
	g.Context = nil

	g.States = make(map[string]*piece.GState[CTX], len(*x.States))
	for xname, xstate := range *x.States {
		gstate, err := parseXState[CTX](xname, xstate)
		if err != nil {
			return nil, err
		}
		g.States[xname] = gstate
	}

	if _, ok := g.States[*x.Initial]; !ok {
		return nil, fmt.Errorf("initial state '%s' does not exist", *x.Initial)
	}
	g.EntryState = g.States[*x.Initial]
	g.CurrentState = g.EntryState

	return &g, nil
}

func parseXState[CTX any](xStateName string, xs XState) (*piece.GState[CTX], error) {
	gs := piece.GState[CTX]{}
	gs.Name = xStateName

	// Always
	xalways, err := xs.GetAlways()
	if err != nil {
		return nil, err
	}
	talways, err := parseXEvent[CTX](xalways)
	if err != nil {
		return nil, err
	}
	gs.Always = talways

	// Entry
	xentry, err := xs.GetEntryActions()
	if err != nil {
		return nil, err
	}
	gs.Entry = parseXActions[CTX](xentry)

	// Exit
	xexit, err := xs.GetExitActions()
	if err != nil {
		return nil, err
	}
	gs.Exit = parseXActions[CTX](xexit)

	// On
	xon, err := xs.GetOn()
	if err != nil {
		return nil, err
	}
	if len(xon) > 0 {
		gs.On = make(map[string]*piece.GTransition[CTX], len(xon))
		for evtName, xts := range xon {
			gts, err := parseXEvent[CTX](xts)
			if err != nil {
				return nil, err
			}
			xEvtName := string(evtName)
			gs.On[xEvtName] = gts
		}
	}

	// Services
	xis, err := xs.GetInvoke()
	if err != nil {
		return nil, err
	}
	gss, err := parseXInvoke[CTX](xis)
	if err != nil {
		return nil, err
	}
	gs.Services = gss

	// Type of state
	if xs.Type == nil {
		gs.StateType = piece.StateTypeNormal
	}
	if xs.Type != nil {
		switch *xs.Type {
		case "initial":
			gs.StateType = piece.StateTypeInitial
		case "final":
			gs.StateType = piece.StateTypeFinal
		case "history":
			gs.StateType = piece.StateTypeHistory
		case "shared":
			gs.StateType = piece.StateTypeShared
		default:
			return nil, fmt.Errorf("unknown state type: %s", *xs.Type)
		}
	}

	return &gs, nil
}

func parseXInvoke[CTX any](xis []XInvoke) ([]*piece.Service[CTX], error) {
	var gss = make([]*piece.Service[CTX], len(xis))
	for i, xi := range xis {
		gs := &piece.Service[CTX]{}
		if xi.Id != nil {
			gs.Id = *xi.Id
		} else {
			gs.Id = *xi.Src
		}
		gs.Src = *xi.Src

		// OnDone
		if xi.OnDone != nil {
			gOnDone, err := parseXEvent[CTX](*xi.OnDone)
			if err != nil {
				return nil, err
			}
			gs.OnDone = gOnDone
		}

		// OnError
		if xi.OnError != nil {
			gOnError, err := parseXEvent[CTX](*xi.OnError)
			if err != nil {
				return nil, err
			}
			gs.OnError = gOnError
		}

		// TODO: Add Invocation

		gss[i] = gs
	}
	return gss, nil
}

func parseXEvent[CTX any](xts []XTransition) (*piece.GTransition[CTX], error) {
	gt := piece.GTransition[CTX]{}
	gt.Guards = make([]*piece.GGuard[CTX], len(xts))
	for i, xt := range xts {
		gg := piece.GGuard[CTX]{}
		gg.Condition = xt.Condition
		gg.Target = xt.Target
		xacts, err := xt.GetActions()
		if err != nil {
			return nil, err
		}
		gg.Actions = parseXActions[CTX](xacts)
		gt.Guards[i] = &gg
	}
	return &gt, nil
}

func parseXActions[CTX any](xacts []XActionName) []*piece.GAction[CTX] {
	var gacts = make([]*piece.GAction[CTX], len(xacts))
	for i, xact := range xacts {
		gact := piece.GAction[CTX]{}
		gact.Name = string(xact)
		gacts[i] = &gact
	}
	return gacts
}

func validateXMachine(x *XMachine) error {
	if x.Id == nil {
		return fmt.Errorf("machine id is required")
	}

	if x.Initial == nil {
		return fmt.Errorf("initial state is required")
	}

	if x.States == nil {
		return fmt.Errorf("machine must have at least one state")
	}

	if _, ok := (*x.States)[*x.Initial]; !ok {
		return fmt.Errorf("initial state must be defined")
	}

	return nil
}
