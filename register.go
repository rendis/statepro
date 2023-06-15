package statepro

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rendis/statepro/piece"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
)

// MachineRegistryDefinitions is the interface that must be implemented by a machine registry definition.
// It is used to link a machine definition with its implementation using the 'id' property.
type MachineRegistryDefinitions[ContextType any] interface {
	GetMachineTemplateId() string
}

type definitionMethodInfoRegistry map[string]map[string]*definitionMethodInfo

type definitionMethodInfo struct {
	fixedName    string
	originalName string
	pos          int
	val          *reflect.Value
}

type proMachineInfo struct {
	gMachine                      any // *piece.GMachine[ContextType]
	machineDefinitionRegistryName string
}

var (
	loadPropOnce sync.Once // used to load prop only once
	isPropLoaded = false   // used to load prop only once
)

var (
	definitionsArr   = make(map[string][]byte) // used to store machine definitions
	definitionsMutex sync.Mutex                // used to lock definitionsArr
)

var (
	xMachines               = make(map[string]*XMachine)        // templateId -> *XMachine
	xChannelsByTemplate     = make(map[string][]chan *XMachine) // templateId -> []chan *XMachine, used to send XMachine to all channels that are waiting for it
	xChannelsByTemplateWait sync.WaitGroup                      // used to wait for all channels on xChannelsByTemplate to be closed
)

var (
	proMachinesMutex sync.Mutex
	proMachines      = make(map[string]*proMachineInfo) // machineId -> *piece.GMachine[ContextType]
)

var (
	predicatesRegistry            = make(definitionMethodInfoRegistry)
	actionsRegistry               = make(definitionMethodInfoRegistry)
	srvRegistry                   = make(definitionMethodInfoRegistry)
	contextSourceHandlersRegistry = make(definitionMethodInfoRegistry)
)

const (
	getMachineTemplateIdMethodName   = "GetMachineTemplateId"
	contextFromSourceMethodName      = "ContextFromSource"
	contextFromSourceMethodNameFixed = "contextfromsource"
	contextToSourceMethodName        = "ContextToSource"
	contextToSourceMethodNameFixed   = "contexttosource"
)

// AddMachine registers a machine definition and returns a unique machine id.
func AddMachine[ContextType any](machineRegistry MachineRegistryDefinitions[ContextType]) string {

	// only allow adding machines before statepro properties have been loaded
	proMachinesMutex.Lock()
	if isPropLoaded {
		log.Fatalf("[FATAL] Cannot add machine '%s' after statepro properties have been loaded", machineRegistry.GetMachineTemplateId())
	}
	proMachinesMutex.Unlock()

	// validate machine registry definition
	if err := validateMachineRegistryDefinition[ContextType](machineRegistry); err != nil {
		log.Fatalf("[FATAL] Error validating machine registry definition: %s", err)
	}

	templateId := machineRegistry.GetMachineTemplateId()
	var machineId = uuid.New().String() + ":" + templateId

	// load machine implementation defined in DefinitionType and register it
	registryDefType := loadMachineImplementation[ContextType](machineRegistry)

	if _, ok := xChannelsByTemplate[templateId]; !ok {
		xChannelsByTemplate[templateId] = make([]chan *XMachine, 0)
	}
	ch := make(chan *XMachine, 1)
	xChannelsByTemplate[templateId] = append(xChannelsByTemplate[templateId], ch)

	xChannelsByTemplateWait.Add(1)
	go asyncGMachineBuilder[ContextType](machineId, registryDefType, ch)

	return machineId
}

// AddDefinitions registers a machine json definitions
func AddDefinitions(definitions map[string]string) {
	definitionsMutex.Lock()
	defer definitionsMutex.Unlock()

	if definitions == nil || len(definitions) == 0 {
		log.Printf("[WARN] No machine definitions found to add")
		return
	}

	for k, v := range definitions {

		if _, ok := definitionsArr[k]; ok {
			log.Fatalf("[FATAL] Machine definition id '%s' already exists.", k)
		}
		definitionsArr[k] = []byte(v)
	}
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
	definitions, err := getMachineDefinitions()
	if err != nil {
		log.Fatalf("[FATAL] Error loading machine definitions: %s", err)
	}

	if definitions == nil || len(definitions) == 0 {
		log.Fatalf("[FATAL] No machine definitions found")
	}

	for path, definition := range definitions {
		m, err := getXMachine(definition)

		if err != nil {
			log.Fatalf("[FATAL] Error loading machine '%s': %s", path, err)
		}

		if _, ok := xMachines[*m.Id]; ok {
			log.Fatalf("[FATAL] Machine definition id '%s' already exists in the path '%s'.", *m.Id, path)
		}

		xMachines[*m.Id] = m
	}
}

func getMachineDefinitions() (map[string][]byte, error) {
	defFromProp, err := getDefinitionsFromProp()
	if err != nil {
		return nil, fmt.Errorf("error getting machine definitions from properties: %s", err)
	}

	definitions := make(map[string][]byte)
	for k, v := range defFromProp {
		definitions[k] = v
	}

	for k, v := range definitionsArr {
		definitions[k] = v
	}

	return definitions, nil
}

func getDefinitionsFromProp() (map[string][]byte, error) {
	var definitions = make(map[string][]byte)

	if p := loadProp(); p != nil {
		defPaths := getDefinitionPaths(p.getPrefix(), p.StateProProp.Paths)

		for _, path := range defPaths {
			byteArr, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			definitions[path] = byteArr
		}
	}

	return definitions, nil
}

func cleanStatepro() {
	definitionsMutex.Lock()
	defer definitionsMutex.Unlock()

	// clear definitions
	for k := range definitionsArr {
		delete(definitionsArr, k)
	}
}

func notifyXMachines() {
	for templateId, xChannels := range xChannelsByTemplate {
		xm, ok := xMachines[templateId]
		if !ok {
			log.Fatalf("[FATAL] Definition for machine '%s' does not exist.", templateId)
		}

		for _, ch := range xChannels {
			ch <- xm
			close(ch)
		}
	}
	xChannelsByTemplateWait.Wait()
}

func asyncGMachineBuilder[ContextType any](machineId, registryType string, xChan <-chan *XMachine) {
	defer xChannelsByTemplateWait.Done()
	xm := <-xChan
	gm, err := parseXMachineToGMachine[ContextType](registryType, xm)
	if err != nil {
		log.Fatalf("[FATAL] Error parsing machine '%s': %s", *xm.Id, err)
	}

	gmInfo := &proMachineInfo{
		gMachine:                      gm,
		machineDefinitionRegistryName: registryType,
	}

	proMachinesMutex.Lock()
	proMachines[machineId] = gmInfo
	proMachinesMutex.Unlock()
}

func appendMachineMethodInfo(methodInfo *definitionMethodInfo, registryType string, reg definitionMethodInfoRegistry) error {
	if _, ok := reg[registryType]; !ok {
		reg[registryType] = make(map[string]*definitionMethodInfo)
	}

	if _, ok := reg[registryType][methodInfo.fixedName]; ok {
		return fmt.Errorf("'%s' already registered for type '%s'", methodInfo.fixedName, registryType)
	}

	reg[registryType][methodInfo.fixedName] = methodInfo
	return nil
}

func appendPredicate(methodInfo *definitionMethodInfo, registryType string) {
	err := appendMachineMethodInfo(methodInfo, registryType, predicatesRegistry)
	if err != nil {
		log.Fatalf("[FATAL] Error appending predicate: %s", err)
	}
}

func appendAction(methodInfo *definitionMethodInfo, registryType string) {
	err := appendMachineMethodInfo(methodInfo, registryType, actionsRegistry)
	if err != nil {
		log.Fatalf("[FATAL] Error appending action: %s", err)
	}
}

func appendSrv(methodInfo *definitionMethodInfo, registryType string) {
	err := appendMachineMethodInfo(methodInfo, registryType, srvRegistry)
	if err != nil {
		log.Fatalf("[FATAL] Error appending service: %s", err)
	}
}

func appendContextSourceHandler(methodInfo *definitionMethodInfo, registryType string) {
	err := appendMachineMethodInfo(methodInfo, registryType, contextSourceHandlersRegistry)
	if err != nil {
		log.Fatalf("[FATAL] Error appending context source handler: %s", err)
	}
}

func loadMachineImplementation[ContextType any](machineRegistry MachineRegistryDefinitions[ContextType]) string {
	machineId := machineRegistry.GetMachineTemplateId()
	log.Printf("[INFO] Loading implementation for machine id '%s'", machineId)

	registryDefTyp := reflect.TypeOf(machineRegistry)
	registryDefVal := reflect.ValueOf(machineRegistry)
	registryDefTypStr := registryDefTyp.Elem().String()

	ctxParamTyp := reflect.TypeOf((*ContextType)(nil)).Elem()
	evtParamTyp := reflect.TypeOf((*piece.Event)(nil)).Elem()
	actToolParamTyp := reflect.TypeOf((*piece.ActionTool[ContextType])(nil)).Elem()

	for i := 0; i < registryDefTyp.NumMethod(); i++ {
		method := registryDefTyp.Method(i)
		methodName := method.Name
		if methodName == getMachineTemplateIdMethodName {
			continue
		}

		lowerName := strings.ToLower(method.Name)
		methodInfo := &definitionMethodInfo{lowerName, methodName, i, &registryDefVal}
		switch {

		// type TPredicate[ContextType any] func(ContextType, Event) (bool, error)
		case isPredicate(method.Type, ctxParamTyp, evtParamTyp):
			appendPredicate(methodInfo, registryDefTypStr)

		// type TAction[ContextType any] func(ContextType, Event, ActionTool[ContextType]) error
		case isAction(method.Type, ctxParamTyp, evtParamTyp, actToolParamTyp):
			appendAction(methodInfo, registryDefTypStr)

		// type TInvocation[ContextType any] func(ContextType, Event)
		case isService(method.Type, ctxParamTyp, evtParamTyp):
			appendSrv(methodInfo, registryDefTypStr)

		case isFromSource(methodName, method.Type, ctxParamTyp):
			appendContextSourceHandler(methodInfo, registryDefTypStr)

		case isToSource(methodName, method.Type, ctxParamTyp):
			appendContextSourceHandler(methodInfo, registryDefTypStr)

		// default case is ignored
		default:
			log.Printf("[INFO] Skipping method '%s' of type '%s' in machine '%s'\n", method.Name, method.Type, machineId)
		}
	}

	log.Printf("[INFO] Loaded implementation for machine id '%s'", machineId)
	return registryDefTypStr
}

func isPredicate(methodTyp, tParam, evtParam reflect.Type) bool {
	// predicate -> (ContextType, Event) (bool, error)
	in := methodTyp.NumIn() == 3 && methodTyp.In(1) == tParam && methodTyp.In(2) == evtParam
	out := methodTyp.NumOut() == 2 && methodTyp.Out(0) == reflect.TypeOf(true) && methodTyp.Out(1) == reflect.TypeOf((*error)(nil)).Elem()
	return in && out
}

func isAction(methodTyp, tParam, evtParam, actToolParam reflect.Type) bool {
	// action -> (ContextType, Event, ActionTool[ContextType]) error
	in := methodTyp.NumIn() == 4 && methodTyp.In(1) == tParam && methodTyp.In(2) == evtParam && methodTyp.In(3) == actToolParam
	out := methodTyp.NumOut() == 1 && methodTyp.Out(0) == reflect.TypeOf((*error)(nil)).Elem()
	return in && out
}

func isService(methodTyp, tParam, evtParam reflect.Type) bool {
	// service -> (ContextType, Event)
	in := methodTyp.NumIn() == 3 && methodTyp.In(1) == tParam && methodTyp.In(2) == evtParam
	out := methodTyp.NumOut() == 0
	return in && out
}

func isFromSource(methodName string, methodTyp, tParam reflect.Type) bool {
	//ContextFromSource -> (params ... any) (ContextType, error)
	if methodName != contextFromSourceMethodName {
		return false
	}
	in := methodTyp.NumIn() > 1 && methodTyp.In(1) == reflect.TypeOf((*[]any)(nil)).Elem()
	out := methodTyp.NumOut() == 2 && methodTyp.Out(0) == tParam && methodTyp.Out(1) == reflect.TypeOf((*error)(nil)).Elem()
	return in && out
}

func isToSource(methodName string, methodTyp, tParam reflect.Type) bool {
	//ContextToSource -> (ContextType) error
	if methodName != contextToSourceMethodName {
		return false
	}
	in := methodTyp.NumIn() == 2 && methodTyp.In(1) == tParam
	out := methodTyp.NumOut() == 1 && methodTyp.Out(0) == reflect.TypeOf((*error)(nil)).Elem()
	return in && out
}

func getBehavior(fixedName, originalName, registryType, behavior string, reg definitionMethodInfoRegistry) (any, error) {
	if r, ok := reg[registryType]; ok {
		if b, ok := r[fixedName]; ok {
			return b.val.Method(b.pos).Interface(), nil
		}
	}
	capitalized := strings.ToUpper(originalName[0:1]) + originalName[1:]
	return nil, fmt.Errorf("no '%s' registered for '%s' in registry definition '%s'", behavior, capitalized, registryType)
}

func getAction(registryType, fixedName, originalName string) (any, error) {
	return getBehavior(fixedName, originalName, registryType, "action", actionsRegistry)
}

func getPredicate(registryType, fixedName, originalName string) (any, error) {
	return getBehavior(fixedName, originalName, registryType, "predicate", predicatesRegistry)
}

func getSrv(registryType, fixedName, originalName string) (any, error) {
	return getBehavior(fixedName, originalName, registryType, "service", srvRegistry)
}

func getFromSourceHandler(registryType string) any {
	handler, _ := getBehavior(
		contextFromSourceMethodNameFixed,
		contextFromSourceMethodName,
		registryType,
		"context from source",
		contextSourceHandlersRegistry,
	)
	return handler
}

func getToSourceHandler(registryType string) any {
	handler, _ := getBehavior(
		contextToSourceMethodNameFixed,
		contextToSourceMethodName,
		registryType,
		"context to source",
		contextSourceHandlersRegistry,
	)
	return handler
}
