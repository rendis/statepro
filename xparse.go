package statepro

import (
	"encoding/json"
	"fmt"
	piece2 "github.com/rendis/statepro/piece"
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

func ParseXMachineToGMachine[T any](x *XMachine) (*piece2.GMachine[T], error) {
	gmachine := &piece2.GMachine[T]{}
	err := validateXMachine(x)
	if err != nil {
		return nil, err
	}

	gmachine.Id = *x.Id
	gmachine.Context = nil

	if x.States != nil {
		gmachine.States = make(map[string]*piece2.GState[T], len(*x.States))
		for xname, xstate := range *x.States {
			gstate, err := parseXState[T](xname, xstate, gmachine)
			if err != nil {
				return nil, err
			}
			gmachine.States[xname] = gstate
		}

		if _, ok := gmachine.States[*x.Initial]; !ok {
			return nil, fmt.Errorf("initial state '%s' does not exist", *x.Initial)
		}
		gmachine.EntryState = gmachine.States[*x.Initial]
		gmachine.CurrentState = gmachine.EntryState
	}

	return gmachine, nil
}

func parseXState[T any](xStateName string, xs XState, pm *piece2.GMachine[T]) (*piece2.GState[T], error) {
	gs := piece2.GState[T]{}
	gs.Name = &xStateName

	// Always
	xalways, err := xs.GetAlways()
	if err != nil {
		return nil, err
	}
	talways, err := parseXEvent[T](xalways, pm)
	if err != nil {
		return nil, err
	}
	gs.Always = talways

	// Entry
	xentry, err := xs.GetEntryActions()
	if err != nil {
		return nil, err
	}

	gs.Entry, err = parseXActions[T](xentry, pm)
	if err != nil {
		return nil, err
	}

	// Exit
	xexit, err := xs.GetExitActions()
	if err != nil {
		return nil, err
	}

	gs.Exit, err = parseXActions[T](xexit, pm)
	if err != nil {
		return nil, err
	}

	// On
	xon, err := xs.GetOn()
	if err != nil {
		return nil, err
	}
	if len(xon) > 0 {
		gs.On = make(map[string]*piece2.GTransition[T], len(xon))
		for evtName, xts := range xon {
			gts, err := parseXEvent[T](xts, pm)
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
	gss, err := parseXInvoke[T](xis, pm)
	if err != nil {
		return nil, err
	}
	gs.Services = gss

	// Type of state
	if xs.Type == nil {
		gs.StateType = piece2.StateTypeNormal
	}
	if xs.Type != nil {
		switch *xs.Type {
		case "initial":
			gs.StateType = piece2.StateTypeInitial
		case "final":
			gs.StateType = piece2.StateTypeFinal
		case "history":
			gs.StateType = piece2.StateTypeHistory
		case "shared":
			gs.StateType = piece2.StateTypeShared
		default:
			return nil, fmt.Errorf("unknown state type: %s", *xs.Type)
		}
	}

	return &gs, nil
}

func parseXInvoke[T any](xis []XInvoke, pm *piece2.GMachine[T]) ([]*piece2.GService[T], error) {
	var gss = make([]*piece2.GService[T], len(xis))
	for i, xi := range xis {
		gs := &piece2.GService[T]{}
		srcName := strings.ToLower(strings.TrimSpace(*xi.Src))

		if xi.Id != nil {
			gs.Id = xi.Id
		} else {
			gs.Id = &srcName
		}
		gs.Src = &srcName

		srv, err := GetSrv[T](srcName)
		if err != nil {
			return nil, err
		}

		gs.Inv, err = piece2.CastToSrv[T](srv)
		if err != nil {
			return nil, err
		}

		// OnDone
		if xi.OnDone != nil {
			gOnDone, err := parseXEvent[T](*xi.OnDone, pm)
			if err != nil {
				return nil, err
			}
			gs.OnDone = gOnDone
		}

		// OnError
		if xi.OnError != nil {
			gOnError, err := parseXEvent[T](*xi.OnError, pm)
			if err != nil {
				return nil, err
			}
			gs.OnError = gOnError
		}

		gss[i] = gs
	}
	return gss, nil
}

func parseXEvent[T any](xts []XTransition, pm *piece2.GMachine[T]) (*piece2.GTransition[T], error) {
	gt := piece2.GTransition[T]{}
	gt.Guards = make([]*piece2.GGuard[T], len(xts))
	for i, xt := range xts {
		gg := piece2.GGuard[T]{}
		gg.Condition = xt.Condition
		gg.Target = xt.Target
		xacts, err := xt.GetActions()
		if err != nil {
			return nil, err
		}
		gg.Actions, err = parseXActions[T](xacts, pm)
		if err != nil {
			return nil, err
		}

		if xt.Condition != nil && len(*xt.Condition) > 0 {
			condName := strings.ToLower(strings.TrimSpace(*xt.Condition))
			predicate, err := GetPredicate[T](condName)
			if err != nil {
				return nil, err
			}

			gg.Predicate, err = piece2.CastPredicate[T](predicate)
			if err != nil {
				return nil, fmt.Errorf("failed to cast predicate '%s': %s", *xt.Condition, err)
			}
		}

		gt.Guards[i] = &gg
	}
	return &gt, nil
}

func parseXActions[T any](xacts []string, pm *piece2.GMachine[T]) ([]*piece2.GAction[T], error) {
	var gacts = make([]*piece2.GAction[T], len(xacts))
	for i, xact := range xacts {
		xactName := strings.ToLower(strings.TrimSpace(xact))
		gact := piece2.GAction[T]{}
		gact.Name = xactName
		gacts[i] = &gact

		// Get actions from Register Actions
		f, err := GetAction[T](xactName)
		if err != nil {
			return nil, err
		}

		// Cast to piece.GAction
		gact.Act, err = piece2.CastToAct[T](f)
		if err != nil {
			return nil, fmt.Errorf("failed to cast action '%s': %s", xact, err)
		}

		gact.Tool = pm
	}
	return gacts, nil
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
