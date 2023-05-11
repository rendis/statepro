package statepro

import (
	"errors"
	"fmt"
	"github.com/rendis/statepro/piece"
	"log"
)

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

func InitMachines() {
	loadPropOnce.Do(func() {
		log.Print("Loading statepro properties")
		isPropLoaded = true
		loadXMachines()
		notifyXMachines()
		log.Print("Statepro properties loaded")
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
