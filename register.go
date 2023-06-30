package statepro

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
)

type definitionMethodInfoRegistryType map[string]map[string]*definitionMethodInfo
type definitionSourceInfoRegistryType map[string]any

type definitionMethodInfo struct {
	fixedName    string
	originalName string
	pos          int
	val          *reflect.Value
}

type specialMethodInfo struct {
	originalName string
	required     bool
}

var (
	registryMutex sync.Mutex

	// MachineLinks registry
	machineLinksDefinitionRegistry = make(map[string]MachineLinks[any])

	// Methods registry
	predicatesRegistry     = make(definitionMethodInfoRegistryType)
	actionsRegistry        = make(definitionMethodInfoRegistryType)
	srvRegistry            = make(definitionMethodInfoRegistryType)
	sourceHandlersRegistry = make(definitionSourceInfoRegistryType)

	// Special methods registry
	alwaysActionsRegistry  = make([]string, 0)
	onEntryActionsRegistry = make([]string, 0)
	onExitActionsRegistry  = make([]string, 0)
)

var methodsToIgnore = map[string]bool{
	"ApplyId":           true,
	"ContextFromSource": true,
	"ContextToSource":   true,
}

func addMachineLinks[ContextType any](ml MachineLinks[ContextType]) error {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	// only allow adding machine definitions before statepro initialization
	if stateProInitialized {
		return fmt.Errorf("cannot add machine links after statepro initialized")
	}

	// validate machine registry definition
	if err := validateMachineRegistryDefinition[ContextType](ml); err != nil {
		return err
	}

	// load machine implementation defined in DefinitionType and register it
	registryDefTypStr, err := loadMachineImplementation[ContextType](ml)
	if err != nil {
		return err
	}

	// check if machine definition already registered
	if _, ok := machineLinksDefinitionRegistry[registryDefTypStr]; ok {
		return fmt.Errorf("machine definition '%s' already registered", registryDefTypStr)
	}

	// register machine definition
	machineLinksDefinitionRegistry[registryDefTypStr] = ml

	return nil
}

func validateMachineRegistryDefinition[ContextType any](ml MachineLinks[ContextType]) error {
	// non null
	if ml == nil {
		return fmt.Errorf("machine links cannot be null")
	}

	mdType := reflect.TypeOf(ml)

	// must be a pointer
	if mdType.Kind() != reflect.Ptr {
		return fmt.Errorf("machine links must be a pointer. Got '%s'", mdType.Kind().String())
	}

	// must be a pointer to a struct
	if mdType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf(
			"machine links must be a pointer to a struct. Got '%s'",
			mdType.Elem().Kind().String(),
		)
	}

	return nil
}

func loadMachineImplementation[ContextType any](ml MachineLinks[ContextType]) (string, error) {
	registryDefTyp := reflect.TypeOf(ml)
	registryDefVal := reflect.ValueOf(ml)
	registryDefTypStr := registryDefTyp.Elem().String()

	sysCtxParamTyp := reflect.TypeOf((*context.Context)(nil)).Elem()
	ctxParamTyp := reflect.TypeOf((*ContextType)(nil)).Elem()
	evtParamTyp := reflect.TypeOf((*Event)(nil)).Elem()
	actToolParamTyp := reflect.TypeOf((*ActionTool[ContextType])(nil)).Elem()

	for i := 0; i < registryDefTyp.NumMethod(); i++ {
		method := registryDefTyp.Method(i)
		methodName := method.Name

		// skip ignored methods
		if _, ok := methodsToIgnore[methodName]; ok {
			continue
		}

		lowerName := strings.ToLower(method.Name)
		methodInfo := &definitionMethodInfo{lowerName, methodName, i, &registryDefVal}

		var appendToRegistryErr error
		var methodType string
		switch {
		// type tPredicate[ContextType any] func(context.Context, *ContextType, Event) (bool, error)
		case isPredicate(method.Type, sysCtxParamTyp, ctxParamTyp, evtParamTyp):
			appendToRegistryErr = appendMachineMethodInfo(methodInfo, registryDefTypStr, predicatesRegistry)
			methodType = "predicate"

		// type tAction[ContextType any] func(context.Context, *ContextType, Event, ActionTool) error
		case isAction(method.Type, sysCtxParamTyp, ctxParamTyp, evtParamTyp, actToolParamTyp):
			appendToRegistryErr = appendMachineMethodInfo(methodInfo, registryDefTypStr, actionsRegistry)
			methodType = "action"

		// type tInvocation[ContextType any] func(context.Context, ContextType, Event)
		case isService(method.Type, sysCtxParamTyp, ctxParamTyp, evtParamTyp):
			appendToRegistryErr = appendMachineMethodInfo(methodInfo, registryDefTypStr, srvRegistry)
			methodType = "service"

		// default case is ignored
		default:
			log.Printf(
				"[WARNING] Skipping method '%s' of type '%s' in machine definition '%s' because it does not match any of the supported types (predicate, action, service or context source handler)",
				methodName, method.Type, registryDefTypStr,
			)
		}

		appendSourceHandler(ml, registryDefTypStr)

		if appendToRegistryErr != nil {
			return "", errors.Join(
				fmt.Errorf("error registering %s '%s' in machine definition '%s'", methodType, methodName, registryDefTypStr),
				appendToRegistryErr,
			)
		}
	}

	return registryDefTypStr, nil
}

// isPredicate -> (context.Context, *ContextType, Event) (bool, error)
func isPredicate(methodTyp, sysCtxParam, tParam, evtParam reflect.Type) bool {
	in := methodTyp.NumIn() == 4 && // 4 because of the receiver
		methodTyp.In(1) == sysCtxParam && // context.Context
		methodTyp.In(2) == reflect.PtrTo(tParam) && // *ContextType
		methodTyp.In(3) == evtParam // Event

	out := methodTyp.NumOut() == 2 && methodTyp.Out(0) == reflect.TypeOf(true) && methodTyp.Out(1) == reflect.TypeOf((*error)(nil)).Elem()
	return in && out
}

// isAction -> (context.Context, *ContextType, Event, ActionTool) error
func isAction(methodTyp, sysCtxParam, tParam, evtParam, actToolParam reflect.Type) bool {
	in := methodTyp.NumIn() == 5 && // 5 because of the receiver
		methodTyp.In(1) == sysCtxParam && // context.Context
		methodTyp.In(2) == reflect.PtrTo(tParam) && // *ContextType
		methodTyp.In(3) == evtParam && // Event
		methodTyp.In(4) == actToolParam // ActionTool

	out := methodTyp.NumOut() == 1 && methodTyp.Out(0) == reflect.TypeOf((*error)(nil)).Elem()

	return in && out
}

// isService -> (context.Context, ContextType, Event)
func isService(methodTyp, sysCtxParam, tParam, evtParam reflect.Type) bool {
	in := methodTyp.NumIn() == 4 && // 4 because of the receiver
		methodTyp.In(1) == sysCtxParam && // context.Context
		methodTyp.In(2) == tParam && // ContextType
		methodTyp.In(3) == evtParam // Event

	out := methodTyp.NumOut() == 0
	return in && out
}

func appendMachineMethodInfo(methodInfo *definitionMethodInfo, registryType string, reg definitionMethodInfoRegistryType) error {
	if _, ok := reg[registryType]; !ok {
		reg[registryType] = make(map[string]*definitionMethodInfo)
	}

	if _, ok := reg[registryType][methodInfo.fixedName]; ok {
		return fmt.Errorf("'%s' already registered for type '%s'", methodInfo.fixedName, registryType)
	}

	reg[registryType][methodInfo.fixedName] = methodInfo
	return nil
}

func appendSourceHandler[ContextType any](ml MachineLinks[ContextType], registryType string) {
	if toSourceHandler, ok := ml.(ProMachineToSourceHandler[ContextType]); ok {
		sourceHandlersRegistry[registryType] = toSourceHandler
		return
	}
	sourceHandlersRegistry[registryType] = nil
}

func registrySpecialMethods(actions []ActionOption) {
	for _, opt := range actions {
		switch opt.ExecutionType {
		case ExecutionTypeAlways:
			alwaysActionsRegistry = append(alwaysActionsRegistry, opt.ActionName)
		case ExecutionTypeOnEntry:
			onEntryActionsRegistry = append(onEntryActionsRegistry, opt.ActionName)
		case ExecutionTypeOnExit:
			onExitActionsRegistry = append(onExitActionsRegistry, opt.ActionName)
		default:
			log.Printf("[WARNING] Skipping unhadled execution type '%s' for action '%s'", opt.ExecutionType, opt.ActionName)
		}
	}
}

// external
func getContextToSourceHandlers[ContextType any](machineDefinitionRegistryName string) ProMachineToSourceHandler[ContextType] {
	if toSourceUnTyped := getToSourceHandlers(machineDefinitionRegistryName); toSourceUnTyped != nil {
		return toSourceUnTyped.(ProMachineToSourceHandler[ContextType])
	}
	return nil
}

func getToSourceHandlers(registryType string) any {
	def, _ := sourceHandlersRegistry[registryType]
	return def
}

func getMachineDefinitionsById[ContextType any](id, version string) (string, MachineLinks[ContextType]) {
	for key, md := range machineLinksDefinitionRegistry {
		if md.ApplyId(id, version) {
			return key, md
		}
	}
	return "", nil
}

func validateAllToSourceImplementationsRequired() error {
	var allRequiredErr error
	for key, toSource := range sourceHandlersRegistry {
		if toSource == nil {
			var cErr = fmt.Errorf("missing required 'ProMachineToSourceHandler' implementation for '%s'", key)
			if allRequiredErr == nil {
				allRequiredErr = cErr
			} else {
				allRequiredErr = errors.Join(allRequiredErr, cErr)
			}
		}
	}
	return allRequiredErr
}

func validateAllSpecialActionsRequired(actions []ActionOption) error {
	if len(actions) == 0 {
		return nil
	}

	allRequired := make([]ActionOption, 0)
	for _, opt := range actions {
		if opt.Required {
			allRequired = append(allRequired, opt)
		}
	}

	// prepare validation
	var allRequiredErr error
	var validatorWg sync.WaitGroup
	var errCollectorWg sync.WaitGroup
	var errCh = make(chan error, 10)
	var maxWorkerCh = make(chan struct{}, 10)

	// define validation function
	var validateRequiredFn = func(registryName string, methods map[string]*definitionMethodInfo, errCh chan<- error) {
		for _, opt := range allRequired {
			fixedRequiredName := strings.ToLower(opt.ActionName)
			if _, ok := methods[fixedRequiredName]; !ok {
				errCh <- fmt.Errorf("missing required action method '%s' for type '%s'", opt.ActionName, registryName)
			}
		}
	}

	// launch error collector
	errCollectorWg.Add(1)
	go func() {
		defer errCollectorWg.Done()
		for err := range errCh {
			allRequiredErr = errors.Join(allRequiredErr, err)
		}
	}()

	// launch validator workers
	for regName, actionsInfo := range actionsRegistry {
		validatorWg.Add(1)
		go func(regName string, actionsInfo map[string]*definitionMethodInfo) {
			defer validatorWg.Done()
			maxWorkerCh <- struct{}{}
			validateRequiredFn(regName, actionsInfo, errCh)
			<-maxWorkerCh
		}(regName, actionsInfo)
	}

	// wait for all validations to finish
	validatorWg.Wait()
	close(maxWorkerCh)

	// wait for error collector to finish
	close(errCh)
	errCollectorWg.Wait()

	return allRequiredErr
}

func getAction(registryType, fixedName, originalName string) (any, error) {
	b, err := getBehavior(fixedName, originalName, registryType, actionsRegistry)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error getting action '%s' for type '%s'", originalName, registryType), err)
	}
	return b, nil
}

func getPredicate(registryType, fixedName, originalName string) (any, error) {
	b, err := getBehavior(fixedName, originalName, registryType, predicatesRegistry)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error getting predicate '%s' for type '%s'", originalName, registryType), err)
	}
	return b, nil
}

func getSrv(registryType, fixedName, originalName string) (any, error) {
	b, err := getBehavior(fixedName, originalName, registryType, srvRegistry)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error getting service '%s' for type '%s'", originalName, registryType), err)
	}
	return b, nil
}

func getBehavior(fixedName, originalName, registryType string, reg definitionMethodInfoRegistryType) (any, error) {
	if r, ok := reg[registryType]; ok {
		if b, ok := r[fixedName]; ok {
			return b.val.Method(b.pos).Interface(), nil
		}
	}
	capitalized := strings.ToUpper(originalName[0:1]) + originalName[1:]
	return nil, fmt.Errorf("no '%s' registered for type '%s'", capitalized, registryType)
}
