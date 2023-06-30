// StatePro is a Golang library designed to efficiently and adaptively handle Finite State Machines in microservices.
//
// Inspired by XState but focused on backend development, its JSON representation is compatible with XState's
// visual creator (stately.ai), facilitating its design and visualization.
package statepro

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

var (
	stateProInitializedMtx sync.Mutex
	stateProInitialized    = false // used to check if statePro is initialized
	toSourceExecutionMode  = ExecutionTypeNone
)

// AddMachineLinks registers machine methods linked to the machine definition
func AddMachineLinks[ContextType any](ml MachineLinks[ContextType]) error {
	return addMachineLinks[ContextType](ml)
}

// InitStatePro initializes statepro. It must be called before using any other statepro function.
func InitStatePro(opt *StateProOptions) error {
	stateProInitializedMtx.Lock()
	defer func() {
		stateProInitialized = true
		stateProInitializedMtx.Unlock()
	}()

	// validate state pro
	if err := validateStatePro(opt); err != nil {
		return errors.Join(fmt.Errorf("statepro initialization failed"), err)
	}

	// process state pro options
	processOptions(opt)

	log.Println("StatePro initialized")
	return nil
}

// GetProMachine returns a ProMachine instance
func GetProMachine[ContextType any](_ context.Context, gmo *GetProMachineOptions[ContextType]) (ProMachine[ContextType], error) {
	var err error

	// check if statePro is initialized
	if !stateProInitialized {
		return nil, StateProNotInitializedErr
	}

	// validate options
	if err = gmo.validate(); err != nil {
		return nil, err
	}

	// get xMachine
	xMachine, err := parseXMachine(gmo.Definition)
	if err != nil {
		return nil, err
	}

	// get machine definition
	definitionRegistryName, md := getMachineDefinitionsById[ContextType](*xMachine.Id, xMachine.Version)
	if md == nil {
		return nil, NoMachineDefinitionAvailableForIdErr
	}

	// build machine
	var gm *gMachine[ContextType]
	if gm, err = buildGMachine[ContextType](definitionRegistryName, xMachine); err != nil {
		return nil, err
	}

	// fill machine fields
	gm.context = gmo.Context
	gm.contextToSourceFn = getContextToSourceHandlers[ContextType](definitionRegistryName)
	gm.toSourceExecutionMode = toSourceExecutionMode
	gm.alwaysNames = alwaysActionsRegistry
	gm.onEntryNames = onEntryActionsRegistry
	gm.onExitNames = onExitActionsRegistry

	return gm, nil
}

func validateStatePro(opt *StateProOptions) error {
	if opt == nil {
		return nil
	}

	// check special methods, no duplicates allowed
	var duplicatedErrs error
	specialCheck := make(map[string]bool)
	for _, sm := range opt.Actions {
		if _, ok := specialCheck[sm.ActionName]; ok {
			duplicatedErrs = errors.Join(duplicatedErrs, fmt.Errorf("duplicate action name '%s'", sm.ActionName))
		}
		specialCheck[sm.ActionName] = true
	}
	if duplicatedErrs != nil {
		return duplicatedErrs
	}

	// check special methods
	if err := validateAllSpecialActionsRequired(opt.Actions); err != nil {
		return errors.Join(fmt.Errorf("some required action are missing"), err)
	}

	// if ToSourceExecutionMode != None, check if all machine definitions have a 'toSource' implementation
	if opt.ToSourceExecutionMode != ExecutionTypeNone {
		err := validateAllToSourceImplementationsRequired()
		if err != nil {
			return errors.Join(
				fmt.Errorf("'ToSourceExecutionMode' is not 'None' but some machine definitions do not have a 'toSource' implementation"),
				err,
			)
		}
	}

	return nil
}

func processOptions(opt *StateProOptions) {
	if opt == nil {
		return
	}

	// set executeToSourceOnEachStateChangeOpt
	toSourceExecutionMode = opt.ToSourceExecutionMode

	// set actions
	registrySpecialMethods(opt.Actions)
}
