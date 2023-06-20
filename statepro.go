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
	"strings"
)

// GetMachineByCompositeId returns a ProMachine instance for the given compositeId, context and params.
func GetMachineByCompositeId[ContextType any](compositeId string, context *ContextType, params ...any) (piece.ProMachine[ContextType], error) {

	pmInfo, ok := proMachines[compositeId]
	if !ok {
		return nil, fmt.Errorf("machine '%s' does not exist", compositeId)
	}

	pm, ok := pmInfo.gMachine.(*piece.GMachine[ContextType])
	if !ok {
		return nil, fmt.Errorf("machine '%s' does not exist", compositeId)
	}

	fromSource, toSource := getContextSourceHandlers[ContextType](pmInfo.machineDefinitionRegistryName)

	if context == nil && fromSource == nil {
		return nil, errors.New("context is nil, please set a context or a 'ContextFromSource handler")
	}

	if context == nil {
		newContext, err := getContextFromSource[ContextType](fromSource, params)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error getting context from 'ContextFromSource' handler: %s", err.Error()))
		}
		context = newContext
	}

	return piece.NewProMachine[ContextType](pm, context, fromSource, toSource), nil
}

// BuildMachineCompositeId builds a machine composite id from the given machine id and version.
func BuildMachineCompositeId(machineId, version string) string {
	machineId = strings.TrimSpace(machineId)
	version = strings.TrimSpace(version)
	return buildMachineCompositeId(machineId, version)
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
		buildGMachines()
		log.Print("[INFO] Statepro properties loaded")
	})
}

func getContextSourceHandlers[ContextType any](machineDefinitionRegistryName string) (piece.ContextFromSourceFnDefinition[ContextType], piece.ContextToSourceFnDefinition[ContextType]) {
	var fromSource piece.ContextFromSourceFnDefinition[ContextType] = nil
	var toSource piece.ContextToSourceFnDefinition[ContextType] = nil

	if method := getFromSourceHandler(machineDefinitionRegistryName); method != nil {
		fromSource = method.(func(params ...any) (*ContextType, error))
	}

	if method := getToSourceHandler(machineDefinitionRegistryName); method != nil {
		toSource = method.(func(ContextType) error)
	}

	return fromSource, toSource
}

func getContextFromSource[ContextType any](fromSource piece.ContextFromSourceFnDefinition[ContextType], params []any) (*ContextType, error) {
	if fromSource == nil {
		return nil, fmt.Errorf("no ContextFromSource handler defined")
	}

	// avoid nil pointer exception
	if params == nil {
		params = []any{}
	}

	newContext, err := fromSource(params...)
	if newContext == nil {
		return nil, err
	}

	return newContext, err
}
