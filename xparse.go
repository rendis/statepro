package statepro

import (
	"encoding/json"
	"fmt"
	"github.com/rendis/statepro/piece"
	"io/ioutil"
	"strings"
)

func GetXMachine(file string) (*XMachine, error) {
	xm := &XMachine{}

	byteArr, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(byteArr, xm)
	if err != nil {
		return nil, err
	}
	return xm, nil
}

func ParseXMachineToGMachine[ContextType any](x *XMachine) (*piece.GMachine[ContextType], error) {
	gMachine := &piece.GMachine[ContextType]{}
	err := validateXMachine(x)
	if err != nil {
		return nil, err
	}

	gMachine.Id = *x.Id
	gMachine.Context = nil

	if x.States != nil {
		gMachine.States = make(map[string]*piece.GState[ContextType], len(*x.States))
		for xname, xstate := range *x.States {
			gstate, err := parseXState[ContextType](xname, xstate, gMachine)
			if err != nil {
				return nil, err
			}
			gMachine.States[xname] = gstate
		}

		if _, ok := gMachine.States[*x.Initial]; !ok {
			return nil, fmt.Errorf("initial state '%s' does not exist", *x.Initial)
		}
		gMachine.EntryState = gMachine.States[*x.Initial]
		gMachine.CurrentState = gMachine.EntryState
	}

	return gMachine, nil
}

func parseXState[ContextType any](xStateName string, xs XState, pm *piece.GMachine[ContextType]) (*piece.GState[ContextType], error) {
	gs := piece.GState[ContextType]{}
	gs.Name = &xStateName

	// Always
	xAlways, err := xs.GetAlways()
	if err != nil {
		return nil, err
	}
	tAlways, err := parseXEvent[ContextType](xAlways, pm)
	if err != nil {
		return nil, err
	}
	gs.Always = tAlways

	// Entry
	xEntry, err := xs.GetEntryActions()
	if err != nil {
		return nil, err
	}

	gs.Entry, err = parseXActions[ContextType](xEntry, pm)
	if err != nil {
		return nil, err
	}

	// Exit
	xExit, err := xs.GetExitActions()
	if err != nil {
		return nil, err
	}

	gs.Exit, err = parseXActions[ContextType](xExit, pm)
	if err != nil {
		return nil, err
	}

	// On
	xon, err := xs.GetOn()
	if err != nil {
		return nil, err
	}
	if len(xon) > 0 {
		gs.On = make(map[string]*piece.GTransition[ContextType], len(xon))
		for evtName, xts := range xon {
			gts, err := parseXEvent[ContextType](xts, pm)
			if err != nil {
				return nil, err
			}
			xEvtName := evtName
			gs.On[xEvtName] = gts
		}
	}

	// Services
	xis, err := xs.GetInvoke()
	if err != nil {
		return nil, err
	}
	gss, err := parseXInvoke[ContextType](xis, pm)
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

func parseXInvoke[ContextType any](xis []XInvoke, pm *piece.GMachine[ContextType]) ([]*piece.GService[ContextType], error) {
	var gss = make([]*piece.GService[ContextType], len(xis))
	for i, xi := range xis {
		gs := &piece.GService[ContextType]{}
		srcName := strings.ToLower(strings.TrimSpace(*xi.Src))

		if xi.Id != nil {
			gs.Id = xi.Id
		} else {
			gs.Id = &srcName
		}
		gs.Src = &srcName

		srv, err := GetSrv[ContextType](srcName)
		if err != nil {
			return nil, err
		}

		gs.Inv, err = piece.CastToSrv[ContextType](srv)
		if err != nil {
			return nil, err
		}

		// OnDone
		if xi.OnDone != nil {
			gOnDone, err := parseXEvent[ContextType](*xi.OnDone, pm)
			if err != nil {
				return nil, err
			}
			gs.OnDone = gOnDone
		}

		// OnError
		if xi.OnError != nil {
			gOnError, err := parseXEvent[ContextType](*xi.OnError, pm)
			if err != nil {
				return nil, err
			}
			gs.OnError = gOnError
		}

		gss[i] = gs
	}
	return gss, nil
}

func parseXEvent[ContextType any](xts []XTransition, pm *piece.GMachine[ContextType]) (*piece.GTransition[ContextType], error) {
	gt := piece.GTransition[ContextType]{}
	gt.Guards = make([]*piece.GGuard[ContextType], len(xts))
	for i, xt := range xts {
		gg := piece.GGuard[ContextType]{}
		gg.Condition = xt.Condition
		gg.Target = xt.Target
		xActs, err := xt.GetActions()
		if err != nil {
			return nil, err
		}
		gg.Actions, err = parseXActions[ContextType](xActs, pm)
		if err != nil {
			return nil, err
		}

		if xt.Condition != nil && len(*xt.Condition) > 0 {
			condName := strings.ToLower(strings.TrimSpace(*xt.Condition))
			predicate, err := GetPredicate[ContextType](condName)
			if err != nil {
				return nil, err
			}

			gg.Predicate, err = piece.CastPredicate[ContextType](predicate)
			if err != nil {
				return nil, fmt.Errorf("failed to cast predicate '%s': %s", *xt.Condition, err)
			}
		}

		gt.Guards[i] = &gg
	}
	return &gt, nil
}

func parseXActions[ContextType any](xActs []string, pm *piece.GMachine[ContextType]) ([]*piece.GAction[ContextType], error) {
	var gActs = make([]*piece.GAction[ContextType], len(xActs))
	for i, xAct := range xActs {
		xActName := strings.ToLower(strings.TrimSpace(xAct))
		gAct := piece.GAction[ContextType]{}
		gAct.Name = xActName
		gActs[i] = &gAct

		// Get actions from Register Actions
		f, err := GetAction[ContextType](xActName)
		if err != nil {
			return nil, err
		}

		// Cast to piece.GAction
		gAct.Act, err = piece.CastToAct[ContextType](f)
		if err != nil {
			return nil, fmt.Errorf("failed to cast action '%s': %s", xAct, err)
		}

		gAct.Tool = pm
	}
	return gActs, nil
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
