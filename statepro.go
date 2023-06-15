// StatePro is a Golang library designed to efficiently and adaptively handle Finite State Machines in microservices.
//
// Inspired by XState but focused on backend development, its JSON representation is compatible with XState's
// visual creator (stately.ai), facilitating its design and visualization.
package statepro

import (
	"errors"
	"fmt"
	"github.com/rendis/statepro/piece"
	"log"
)

// GetMachine returns a ProMachine instance for the given machineId and context.
func GetMachine[ContextType any](machineId string, context *ContextType) (piece.ProMachine[ContextType], error) {

	pmInfo, ok := proMachines[machineId]
	if !ok {
		return nil, fmt.Errorf("machine '%s' does not exist", machineId)
	}

	pm, ok := pmInfo.gMachine.(*piece.GMachine[ContextType])
	if !ok {
		return nil, fmt.Errorf("machine '%s' does not exist", machineId)
	}

	fromSource, toSource := getContextSourceHandlers[ContextType](pmInfo.machineDefinitionRegistryName)

	if context == nil && fromSource == nil {
		return nil, errors.New("context is nil, please set a context or a 'ContextFromSource handler")
	}

	if context == nil {
		newContext, err := getContextFromSource[ContextType](fromSource)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error getting context from 'ContextFromSource' handler: %s", err.Error()))
		}
		context = newContext
	}

	return piece.NewProMachine[ContextType](pm, context, fromSource, toSource), nil
}

// InitMachines loads the statepro properties and initializes the machines.
func InitMachines() {
	loadPropOnce.Do(func() {
		defer func() {
			cleanStatepro()
		}()
		log.Print("[INFO] Loading statepro properties")
		isPropLoaded = true
		loadXMachines()
		notifyXMachines()
		log.Print("[INFO] Statepro properties loaded")
	})
}

func getContextSourceHandlers[ContextType any](machineDefinitionRegistryName string) (piece.ContextFromSourceFnDefinition[ContextType], piece.ContextToSourceFnDefinition[ContextType]) {
	var fromSource piece.ContextFromSourceFnDefinition[ContextType] = nil
	var toSource piece.ContextToSourceFnDefinition[ContextType] = nil

	if method := getFromSourceHandler(machineDefinitionRegistryName); method != nil {
		fromSource = method.(func(params ...any) (ContextType, error))
	}

	if method := getToSourceHandler(machineDefinitionRegistryName); method != nil {
		toSource = method.(func(ContextType) error)
	}

	return fromSource, toSource
}

func getContextFromSource[ContextType any](fromSource piece.ContextFromSourceFnDefinition[ContextType]) (*ContextType, error) {
	if fromSource == nil {
		return nil, fmt.Errorf("no ContextFromSource handler defined")
	}
	newContext, err := fromSource()
	return &newContext, err
}
