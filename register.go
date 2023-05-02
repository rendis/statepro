package statepro

import (
	"fmt"
	"github.com/rendis/statepro/piece"
	"log"
	"reflect"
	"strings"
	"sync"
)

type logic struct {
	name string
	pos  int
	val  *reflect.Value
}

type logicRegistry map[reflect.Type]map[string]logic

var proMutex sync.Mutex
var xchanWait sync.WaitGroup
var xchans = make(map[machineKey]chan *XMachine)
var xMachines = make(map[string]*XMachine)
var proMachines = make(map[machineKey]any)
var predicatesRegistry = make(logicRegistry)
var actionsRegistry = make(logicRegistry)
var srvRegistry = make(logicRegistry)

type MachineDefinitions[T any] interface {
	GetMachineId() string
}

func AddMachine[T any](d MachineDefinitions[T]) {
	key := buildKey[T](d.GetMachineId())
	xid := d.GetMachineId()
	if _, ok := xchans[key]; ok {
		log.Fatalf("Machine definition id '%s' for type '%s' already exists.", xid, key.typ)
	}
	xchan := make(chan *XMachine, 1)
	xchans[key] = xchan
	xchanWait.Add(1)
	loadImplementation[T](d)
	go asyncBuilder[T](key, xchan)
}

func loadXMachines() {
	p := LoadProp()
	defPaths := GetDefinitionPaths(p.GetPrefix(), p.Scanner.Paths)

	for _, path := range defPaths {
		m, err := GetXMachine(path)

		if err != nil {
			log.Fatalf("Error loading machine '%s': %s", path, err)
		}

		if _, ok := xMachines[*m.Id]; ok {
			log.Fatalf("Machine definition id '%s' already exists in the path '%s'.", *m.Id, path)
		}

		xMachines[*m.Id] = m
	}
}

func notifyXMachines() {
	for k, xchan := range xchans {
		xm, ok := xMachines[k.id]
		if !ok {
			log.Fatalf("Definition for machine '%s' does not exist.", k.id)
		}
		xchan <- xm
		close(xchan)
	}
	xchanWait.Wait()
}

func asyncBuilder[T any](key machineKey, xchan <-chan *XMachine) {
	defer xchanWait.Done()
	xm := <-xchan
	gm, err := ParseXMachineToGMachine[T](xm)
	if err != nil {
		log.Fatalf("Error parsing machine '%s': %s", *xm.Id, err)
	}
	proMutex.Lock()
	proMachines[key] = gm
	proMutex.Unlock()
}

func buildKey[T any](id string) machineKey {
	return machineKey{id, reflect.TypeOf(new(T)).Elem()}
}

func appendLogic(l logic, typ reflect.Type, reg logicRegistry) error {
	if _, ok := reg[typ]; !ok {
		reg[typ] = make(map[string]logic)
	}

	if _, ok := reg[typ][l.name]; ok {
		return fmt.Errorf("'%s' already registered for type '%s'", l.name, typ)
	}

	reg[typ][l.name] = l
	return nil
}

func appendPredicate(l logic, typ reflect.Type) {
	err := appendLogic(l, typ, predicatesRegistry)
	if err != nil {
		log.Fatalf("Error appending predicate: %s", err)
	}
}

func appendAction(l logic, typ reflect.Type) {
	err := appendLogic(l, typ, actionsRegistry)
	if err != nil {
		log.Fatalf("Error appending action: %s", err)
	}
}

func appendSrv(l logic, typ reflect.Type) {
	err := appendLogic(l, typ, srvRegistry)
	if err != nil {
		log.Fatalf("Error appending service: %s", err)
	}
}

func loadImplementation[T any](d MachineDefinitions[T]) {
	id := d.GetMachineId()
	log.Printf("Loading implementation for machine '%s'", id)
	typ := reflect.TypeOf(d)
	val := reflect.ValueOf(d)

	tParam := reflect.TypeOf((*T)(nil)).Elem()
	evtParam := reflect.TypeOf((*piece.Event)(nil)).Elem()
	actToolParam := reflect.TypeOf((*piece.ActionTool[T])(nil)).Elem()

	for i := 0; i < typ.NumMethod(); i++ {
		if m := typ.Method(i); m.Name != "GetMachineId" {
			name := strings.ToLower(m.Name)
			l := logic{name, i, &val}
			switch {
			// type TPredicate[T any] func(T, Event) (bool, error)
			case isPredicate(m.Type, tParam, evtParam):
				appendPredicate(l, tParam)
			// type TAction[T any] func(T, Event, ActionTool[T]) error
			case isAction(m.Type, tParam, evtParam, actToolParam):
				appendAction(l, tParam)
			// type TInvocation[T any] func(T, Event) ServiceResponse
			case isService(m.Type, tParam, evtParam):
				appendSrv(l, tParam)
			default:
				log.Printf("Skipping method '%s' of type '%s' in machine '%s'\n", m.Name, m.Type, id)
			}
		}
	}
}

func isPredicate(typ, tParam, evtParam reflect.Type) bool {
	in := typ.NumIn() == 3 && typ.In(1) == tParam && typ.In(2) == evtParam
	out := typ.NumOut() == 2 && typ.Out(0) == reflect.TypeOf(true) && typ.Out(1) == reflect.TypeOf((*error)(nil)).Elem()
	return in && out
}

func isAction(typ, tParam, evtParam, actToolParam reflect.Type) bool {
	in := typ.NumIn() == 4 && typ.In(1) == tParam && typ.In(2) == evtParam && typ.In(3) == actToolParam
	out := typ.NumOut() == 1 && typ.Out(0) == reflect.TypeOf((*error)(nil)).Elem()
	return in && out
}

func isService(typ, tParam, evtParam reflect.Type) bool {
	in := typ.NumIn() == 3 && typ.In(1) == tParam && typ.In(2) == evtParam
	out := typ.NumOut() == 1 && typ.Out(0) == reflect.TypeOf((*piece.ServiceResponse)(nil)).Elem()
	return in && out
}

func getBehavior[T any](name, behavior string, reg logicRegistry) (any, error) {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	r, ok := reg[typ]
	if !ok {
		return nil, fmt.Errorf("no %s registered for name '%s' and type '%s'", behavior, name, typ)
	}

	b, ok := r[name]
	if !ok {
		return nil, fmt.Errorf("no %s registered for name '%s' and type '%s'", behavior, name, typ)
	}

	return b.val.Method(b.pos).Interface(), nil
}

func GetAction[T any](name string) (any, error) {
	return getBehavior[T](name, "action", actionsRegistry)
}

func GetPredicate[T any](name string) (any, error) {
	return getBehavior[T](name, "predicate", predicatesRegistry)
}

func GetSrv[T any](name string) (any, error) {
	return getBehavior[T](name, "service", srvRegistry)
}
