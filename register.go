package statepro

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rendis/statepro/piece"
	"log"
	"reflect"
	"strings"
	"sync"
)

type definitionMethodInfo struct {
	fixedName    string
	originalName string
	pos          int
	val          *reflect.Value
}

var loadPropOnce sync.Once // used to load prop only once
var isPropLoaded = false   // used to load prop only once

type definitionMethodInfoRegistry map[reflect.Type]map[string]definitionMethodInfo

var xMachines = make(map[string]*XMachine)                  // templateId -> *XMachine
var xChannelsByTemplate = make(map[string][]chan *XMachine) // templateId -> []chan *XMachine, used to send XMachine to all channels that are waiting for it
var xChannelsByTemplateWait sync.WaitGroup                  // used to wait for all channels on xChannelsByTemplate to be closed

var proMachinesMutex sync.Mutex
var proMachines = make(map[string]any) // machineId -> *piece.GMachine[ContextType]

var predicatesRegistry = make(definitionMethodInfoRegistry)
var actionsRegistry = make(definitionMethodInfoRegistry)
var srvRegistry = make(definitionMethodInfoRegistry)

type MachineRegistryDefinitions[ContextType any] interface {
	GetMachineTemplateId() string
}

func AddMachine[ContextType any](machineRegistry MachineRegistryDefinitions[ContextType]) string {

	// only allow adding machines before statepro properties have been loaded
	proMachinesMutex.Lock()
	if isPropLoaded {
		log.Fatalf("Cannot add machine '%s' after statepro properties have been loaded", machineRegistry.GetMachineTemplateId())
	}
	proMachinesMutex.Unlock()

	// validate machine registry definition
	if err := validateMachineRegistryDefinition[ContextType](machineRegistry); err != nil {
		log.Fatalf("Error validating machine registry definition: %s", err)
	}

	templateId := machineRegistry.GetMachineTemplateId()
	var machineId = uuid.New().String() + ":" + templateId

	// load machine implementation defined in DefinitionType and register it
	loadMachineImplementation[ContextType](machineRegistry)

	if _, ok := xChannelsByTemplate[templateId]; !ok {
		xChannelsByTemplate[templateId] = make([]chan *XMachine, 0)
	}
	ch := make(chan *XMachine, 1)
	xChannelsByTemplate[templateId] = append(xChannelsByTemplate[templateId], ch)

	xChannelsByTemplateWait.Add(1)
	go asyncGMachineBuilder[ContextType](machineId, ch)

	return machineId
}

func validateMachineRegistryDefinition[ContextType any](machineRegistry MachineRegistryDefinitions[ContextType]) error {
	// non null
	if machineRegistry == nil {
		return fmt.Errorf("machine registry definition cannot be null")
	}

	// templateId not empty
	templateId := machineRegistry.GetMachineTemplateId()
	if strings.TrimSpace(templateId) == "" {
		return fmt.Errorf("machine registry definition cannot have an empty template id")
	}

	// must be a pointer
	if reflect.TypeOf(machineRegistry).Kind() != reflect.Ptr {
		return fmt.Errorf("machine registry definition with id '%s' must be a pointer", templateId)
	}

	// must be a pointer to a struct
	if reflect.TypeOf(machineRegistry).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("machine registry definition with id '%s' must be a pointer to a struct", templateId)
	}

	return nil
}

func loadXMachines() {
	p := loadProp()
	defPaths := getDefinitionPaths(p.getPrefix(), p.Scanner.Paths)

	for _, path := range defPaths {
		m, err := getXMachine(path)

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
	for templateId, xChannels := range xChannelsByTemplate {
		xm, ok := xMachines[templateId]
		if !ok {
			log.Fatalf("Definition for machine '%s' does not exist.", templateId)
		}

		for _, ch := range xChannels {
			ch <- xm
			close(ch)
		}
	}
	xChannelsByTemplateWait.Wait()
}

func asyncGMachineBuilder[ContextType any](machineId string, xChan <-chan *XMachine) {
	defer xChannelsByTemplateWait.Done()
	xm := <-xChan
	gm, err := parseXMachineToGMachine[ContextType](xm)
	if err != nil {
		log.Fatalf("Error parsing machine '%s': %s", *xm.Id, err)
	}
	proMachinesMutex.Lock()
	proMachines[machineId] = gm
	proMachinesMutex.Unlock()
}

func appendMachineMethodInfo(methodInfo definitionMethodInfo, typ reflect.Type, reg definitionMethodInfoRegistry) error {
	if _, ok := reg[typ]; !ok {
		reg[typ] = make(map[string]definitionMethodInfo)
	}

	if _, ok := reg[typ][methodInfo.fixedName]; ok {
		return fmt.Errorf("'%s' already registered for type '%s'", methodInfo.fixedName, typ)
	}

	reg[typ][methodInfo.fixedName] = methodInfo
	return nil
}

func appendPredicate(methodInfo definitionMethodInfo, typ reflect.Type) {
	err := appendMachineMethodInfo(methodInfo, typ, predicatesRegistry)
	if err != nil {
		log.Fatalf("Error appending predicate: %s", err)
	}
}

func appendAction(methodInfo definitionMethodInfo, typ reflect.Type) {
	err := appendMachineMethodInfo(methodInfo, typ, actionsRegistry)
	if err != nil {
		log.Fatalf("Error appending action: %s", err)
	}
}

func appendSrv(methodInfo definitionMethodInfo, typ reflect.Type) {
	err := appendMachineMethodInfo(methodInfo, typ, srvRegistry)
	if err != nil {
		log.Fatalf("Error appending service: %s", err)
	}
}

func loadMachineImplementation[ContextType any](machineRegistry MachineRegistryDefinitions[ContextType]) {
	machineId := machineRegistry.GetMachineTemplateId()
	log.Printf("Loading implementation for machine id '%s'", machineId)

	defTyp := reflect.TypeOf(machineRegistry)
	defVal := reflect.ValueOf(machineRegistry)

	ctxParamTyp := reflect.TypeOf((*ContextType)(nil)).Elem()
	evtParamTyp := reflect.TypeOf((*piece.Event)(nil)).Elem()
	actToolParamTyp := reflect.TypeOf((*piece.ActionTool[ContextType])(nil)).Elem()

	for i := 0; i < defTyp.NumMethod(); i++ {
		method := defTyp.Method(i)
		if method.Name == "GetMachineTemplateId" {
			continue
		}

		lowerName := strings.ToLower(method.Name)
		methodInfo := definitionMethodInfo{lowerName, method.Name, i, &defVal}
		switch {

		// type TPredicate[ContextType any] func(ContextType, Event) (bool, error)
		case isPredicate(method.Type, ctxParamTyp, evtParamTyp):
			appendPredicate(methodInfo, ctxParamTyp)

		// type TAction[ContextType any] func(ContextType, Event, ActionTool[ContextType]) error
		case isAction(method.Type, ctxParamTyp, evtParamTyp, actToolParamTyp):
			appendAction(methodInfo, ctxParamTyp)

		// type TInvocation[ContextType any] func(ContextType, Event) ServiceResponse
		case isService(method.Type, ctxParamTyp, evtParamTyp):
			appendSrv(methodInfo, ctxParamTyp)

		// default case is ignored
		default:
			log.Printf("Skipping method '%s' of type '%s' in machine '%s'\n", method.Name, method.Type, machineId)
		}
	}

	log.Printf("Loaded implementation for machine id '%s'", machineId)
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

func getBehavior[ContextType any](fixedName, originalName, behavior string, reg definitionMethodInfoRegistry) (any, error) {
	typ := reflect.TypeOf((*ContextType)(nil)).Elem()
	r, ok := reg[typ]
	if !ok {
		return nil, fmt.Errorf("no %s registered for '%s' and type '%s'", behavior, originalName, typ)
	}

	b, ok := r[fixedName]
	if !ok {
		return nil, fmt.Errorf("no %s registered for '%s' and type '%s'", behavior, originalName, typ)
	}

	return b.val.Method(b.pos).Interface(), nil
}

func getAction[ContextType any](fixedName, originalName string) (any, error) {
	return getBehavior[ContextType](fixedName, originalName, "action", actionsRegistry)
}

func getPredicate[ContextType any](fixedName, originalName string) (any, error) {
	return getBehavior[ContextType](fixedName, originalName, "predicate", predicatesRegistry)
}

func getSrv[ContextType any](fixedName, originalName string) (any, error) {
	return getBehavior[ContextType](fixedName, originalName, "service", srvRegistry)
}
