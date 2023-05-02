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

//func GetMachine[T any](id string) (*piece.GMachine[T], error) {
//	if ptrPM, ok := proMachines[buildKey[T](id)]; ok {
//		if pm, ok := ptrPM.(*piece.GMachine[T]); ok {
//			return pm, nil
//		}
//	}
//	return nil, fmt.Errorf("machine '%s' does not exist", id)
//}

func GetMachine[T any](id string) (piece.ProMachine[T], error) {
	if ptrPM, ok := proMachines[buildKey[T](id)]; ok {
		if pm, ok := ptrPM.(*piece.GMachine[T]); ok {
			return pm, nil
		}
	}
	return nil, fmt.Errorf("machine '%s' does not exist", id)
}

func InitMachines() {
	loadXMachines()
	notifyXMachines()
}

/*
func BuildProMachine[T any](id string) {
	if xm, ok := xMachines[id]; ok {
		gm, err := xparse.ParseXMachine[T](xm)
		if err != nil {
			log.Fatalf("Error parsing machine '%s': %s", id, err)
		}
		proMachines[buildKey[T](id)] = gm
		return
	}
	log.Fatalf("GMachine definition id '%s' does not exist.", id)
}
*/

/*
var machineRegistry = make(map[string]any)

func RegisterMachine[T any](id string) {
	m := &piece.GMachine[T]{}
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

func GetMachine[T any](id string) *piece.GMachine[T] {
	if m, ok := machineRegistry[id]; ok {
		return m.(*piece.GMachine[T])
	}
	return nil
}
*/

/*
type MachineBuilder[T any] interface {
	WithState(stateName string) MachineBuilder[T]
	WithContext(context T) MachineBuilder[T]
	Start() StatePro[T]
}

type StatePro[T any] interface {
	PossibleEvents() []string
	ContainsEvent(eventName string) bool
	Send(eventName string) error
}
*/
