package statepro

import (
	"fmt"
	"github.com/rendis/statepro/piece"
	"log"
)

func GetMachineById[ContextType any](machineId string, context *ContextType) (piece.ProMachine[ContextType], error) {
	if ptrPM, ok := proMachines[machineId]; ok {
		if pm, ok := ptrPM.(*piece.GMachine[ContextType]); ok {
			return piece.NewProMachine[ContextType](pm, context), nil
		}
	}
	return nil, fmt.Errorf("machine '%s' does not exist", machineId)
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

/*
func BuildProMachine[ContextType any](id string) {
	if xm, ok := xMachines[id]; ok {
		gm, err := xparse.ParseXMachine[ContextType](xm)
		if err != nil {
			log.Fatalf("Error parsing machine '%s': %s", id, err)
		}
		proMachines[buildKeyFromDefinition[ContextType](id)] = gm
		return
	}
	log.Fatalf("GMachine definition id '%s' does not exist.", id)
}
*/

/*
var machineRegistry = make(map[string]any)

func RegisterMachine[ContextType any](id string) {
	m := &piece.GMachine[ContextType]{}
	if _, ok := machineRegistry[id]; ok {
		log.Fatalf("GMachine %s already registered", id)
	}
	machineRegistry[id] = m
}

func LoadXMachines() {
	prop := loadProp()
	prefix := ""
	if prop.Scanner.FilePrefix != nil {
		prefix = *prop.Scanner.FilePrefix
	}
	defPaths := getDefinitionPaths(prefix, prop.Scanner.Paths)
	var m *piece.GMachine[any]
	for _, path := range defPaths {
		var err error

		m, err = xparse.LoadXFile[any](path)
		if err != nil {
			log.Fatalf("Error loading machine '%s': %s", path, err)
		}
		if _, ok := machineRegistry[m.Id]; ok {
			log.Fatalf("GMachine '%s' already exists", m.Id)
		}
		machineRegistry[m.Id] = m
		log.Printf("Loaded machine '%s'", m.Id)
	}
}

func GetMachineFromDefinition[ContextType any](id string) *piece.GMachine[ContextType] {
	if m, ok := machineRegistry[id]; ok {
		return m.(*piece.GMachine[ContextType])
	}
	return nil
}
*/

/*
type MachineBuilder[ContextType any] interface {
	WithState(stateName string) MachineBuilder[ContextType]
	WithContext(context ContextType) MachineBuilder[ContextType]
	Start() StatePro[ContextType]
}

type StatePro[ContextType any] interface {
	PossibleEvents() []string
	ContainsEvent(eventName string) bool
	Send(eventName string) error
}
*/
