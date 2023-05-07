package statepro

import (
	"fmt"
	"github.com/rendis/statepro/piece"
	"reflect"
)

type machineKey struct {
	id  string
	typ reflect.Type
}

// var xMachines = make(map[string]*xparse.XMachine)
// var proMachines = make(map[machineKey]any)

//func GetMachine[ContextType any](id string) (*piece.GMachine[ContextType], error) {
//	if ptrPM, ok := proMachines[buildKey[ContextType](id)]; ok {
//		if pm, ok := ptrPM.(*piece.GMachine[ContextType]); ok {
//			return pm, nil
//		}
//	}
//	return nil, fmt.Errorf("machine '%s' does not exist", id)
//}

func GetMachine[ContextType any](d MachineDefinition[ContextType]) (piece.ProMachine[ContextType], error) {
	if ptrPM, ok := proMachines[buildKey(d)]; ok {
		if pm, ok := ptrPM.(*piece.GMachine[ContextType]); ok {
			return pm, nil
		}
	}
	return nil, fmt.Errorf("machine '%s' does not exist for type '%s'", d.GetMachineId(), reflect.TypeOf(d))
}

func InitMachines() {
	loadXMachines()
	notifyXMachines()
}

/*
func BuildProMachine[ContextType any](id string) {
	if xm, ok := xMachines[id]; ok {
		gm, err := xparse.ParseXMachine[ContextType](xm)
		if err != nil {
			log.Fatalf("Error parsing machine '%s': %s", id, err)
		}
		proMachines[buildKey[ContextType](id)] = gm
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

func GetMachine[ContextType any](id string) *piece.GMachine[ContextType] {
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
