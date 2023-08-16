package statepro

import (
	"encoding/json"
	"fmt"
	"strings"
)

func buildGMachine[ContextType any](registryType string, xMachine *XMachine) (*gMachine[ContextType], error) {
	gMachine := &gMachine[ContextType]{
		id: *xMachine.Id,
		xm: xMachine,
	}

	gMachine.states = make(map[string]*gState[ContextType], len(xMachine.States))
	gMachine.finalStates = make([]string, 0)
	var inEventsByState = make(map[string]map[string]struct{}, len(xMachine.States))

	// parse states and add to machine
	// add final states
	// get in events by state
	for xName, xstate := range xMachine.States {
		// parse state and add to machine
		gState, err := parseXState[ContextType](registryType, xName, xstate)
		if err != nil {
			return nil, err
		}
		gMachine.states[xName] = gState

		// add final state
		if gState.StateType == StateTypeFinal {
			gMachine.finalStates = append(gMachine.finalStates, xName)
		}

		// get in events
		for evtName, transitions := range gState.On {
			for _, guard := range transitions.Guards {
				targetStateName := *guard.Target
				if _, ok := inEventsByState[targetStateName]; !ok {
					inEventsByState[targetStateName] = make(map[string]struct{})
				}
				inEventsByState[targetStateName][evtName] = struct{}{}
			}
		}
	}

	// add in events to machine
	gMachine.inEventsByState = make(map[string]map[string]struct{}, len(inEventsByState))
	for stateName, inEvents := range inEventsByState {
		gMachine.inEventsByState[stateName] = make(map[string]struct{}, len(inEvents))
		for evtName := range inEvents {
			gMachine.inEventsByState[stateName][evtName] = struct{}{}
		}
	}

	gMachine.entryState = gMachine.states[*xMachine.Initial]
	gMachine.currentState = gMachine.entryState

	return gMachine, nil
}

func parseXMachine(definition []byte) (*XMachine, error) {
	xm := &XMachine{}
	err := json.Unmarshal(definition, xm)
	return xm, err
}

func parseXState[ContextType any](registryType, xStateName string, xs *XState) (*gState[ContextType], error) {
	gs := gState[ContextType]{}
	gs.Name = &xStateName

	var err error

	// Always
	if gs.Always, err = parseXEvent[ContextType](xStateName, registryType, xs.Always); err != nil {
		return nil, err
	}

	// Entry
	if gs.Entry, err = parseXActions[ContextType](registryType, xs.Entry); err != nil {
		return nil, err
	}

	// Exit
	if gs.Exit, err = parseXActions[ContextType](registryType, xs.Exit); err != nil {
		return nil, err
	}

	// On
	gs.On = make(map[string]*gTransition[ContextType], len(xs.On))
	for evtName, xts := range xs.On {
		gts, err := parseXEvent[ContextType](xStateName, registryType, xts)
		if err != nil {
			return nil, err
		}
		xEvtName := evtName
		gs.On[xEvtName] = gts
	}

	// Services
	if gs.Services, err = parseXInvoke[ContextType](xStateName, registryType, xs.Invoke); err != nil {
		return nil, err
	}

	// Type of state
	if xs.Type == nil {
		gs.StateType = StateTypeNormal
	} else {
		switch *xs.Type {
		case "initial":
			gs.StateType = StateTypeInitial
		case "final":
			gs.StateType = StateTypeFinal
		case "history":
			gs.StateType = StateTypeHistory
		case "shared":
			gs.StateType = StateTypeShared
		default:
			return nil, fmt.Errorf("unknown state type: %s", *xs.Type)
		}
	}

	return &gs, nil
}

func parseXEvent[ContextType any](stateName, registryType string, xts []*XTransition) (*gTransition[ContextType], error) {
	gt := gTransition[ContextType]{}

	if len(xts) == 0 {
		return &gt, nil
	}

	gt.Guards = make([]*gGuard[ContextType], len(xts))
	for i, xt := range xts {
		gg := gGuard[ContextType]{}
		gg.Condition = xt.Condition

		if xt.Target == nil {
			gg.Target = &stateName
		} else {
			gg.Target = xt.Target
		}

		var err error
		if gg.Actions, err = parseXActions[ContextType](registryType, xt.Actions); err != nil {
			return nil, err
		}

		if xt.Condition != nil && len(*xt.Condition) > 0 {
			originalCondName := strings.TrimSpace(*xt.Condition)
			condName := strings.ToLower(originalCondName)
			predicate, err := getPredicate(registryType, condName, originalCondName)
			if err != nil {
				return nil, err
			}

			gg.Predicate, err = castPredicate[ContextType](predicate)
			if err != nil {
				return nil, fmt.Errorf("failed to cast predicate '%s': %s", *xt.Condition, err)
			}
		}

		gt.Guards[i] = &gg
	}
	return &gt, nil
}

func parseXActions[ContextType any](registryType string, xActs []string) ([]*gAction[ContextType], error) {
	var gActs = make([]*gAction[ContextType], len(xActs))
	for i, xAct := range xActs {
		originalXActName := strings.TrimSpace(xAct)
		xActName := strings.ToLower(originalXActName)
		gAct := gAction[ContextType]{}
		gAct.Name = xActName
		gActs[i] = &gAct

		// Get actions from Register Actions
		f, err := getAction(registryType, xActName, originalXActName)
		if err != nil {
			return nil, err
		}

		// Cast to piece.gAction
		gAct.Act, err = castToAct[ContextType](f)
		if err != nil {
			return nil, fmt.Errorf("failed to cast action '%s': %s", xAct, err)
		}
	}
	return gActs, nil
}

func parseXInvoke[ContextType any](stateName, registryType string, xis []*XInvoke) ([]*gService[ContextType], error) {
	var gss = make([]*gService[ContextType], len(xis))

	for i, xi := range xis {
		gs := &gService[ContextType]{}
		originalSrcName := strings.TrimSpace(*xi.Src)
		srcName := strings.ToLower(originalSrcName)

		if xi.Id != nil {
			gs.Id = xi.Id
		} else {
			gs.Id = &srcName
		}
		gs.Src = &srcName

		srv, err := getSrv(registryType, srcName, originalSrcName)
		if err != nil {
			return nil, err
		}

		if gs.Inv, err = castToSrv[ContextType](srv); err != nil {
			return nil, err
		}

		// OnDone
		if gs.OnDone, err = parseXEvent[ContextType](stateName, registryType, xi.OnDone); err != nil {
			return nil, err
		}

		// OnError
		if xi.OnError != nil {
			gOnError, err := parseXEvent[ContextType](stateName, registryType, xi.OnError)
			if err != nil {
				return nil, err
			}
			gs.OnError = gOnError
		}

		gss[i] = gs
	}
	return gss, nil
}
