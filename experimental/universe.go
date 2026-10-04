package experimental

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/rendis/statepro/v3/builtin"
	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/internal/util"
	"github.com/rendis/statepro/v3/theoretical"
)

const (
	startEventName                             = "start"
	startOnEventName                           = "startOn"
	initializingUniverseErrMsgTemplate         = "error initializing universe '%s'"
	realitiesNotExistErrMsgTemplate            = "universe '%s' does not have reality '%s' defined"
	errorExecutingOnEntryProcessMsgTemplate    = "error executing on entry process for universe '%s' and reality '%s'"
	errorExecutingAlwaysTransitionsMsgTemplate = "error executing always transitions for reality '%s'"

	// maxEmitDepth is the maximum nesting depth for emitted events.
	// Prevents infinite loops when entry actions emit events that cause transitions
	// whose target reality's entry actions emit events again (A → B → A → ...).
	maxEmitDepth = 10
)

type UniverseInfoSnapshot struct {
	ID                         string            `json:"id"`
	CanonicalName              string            `json:"canonicalName"`
	Version                    string            `json:"version"`
	Initialized                bool              `json:"initialized"`
	CurrentReality             *string           `json:"currentReality,omitempty"`
	RealityInitialized         bool              `json:"realityInitialized"`
	InSuperposition            bool              `json:"inSuperposition"`
	RealityBeforeSuperposition *string           `json:"realityBeforeSuperposition,omitempty"`
	Accumulator                *eventAccumulator `json:"accumulator,omitempty"`
	Metadata                   map[string]any    `json:"metadata,omitempty"`
}

func NewExUniverse(model *theoretical.UniverseModel) *ExUniverse {
	u := &ExUniverse{
		model: model,
	}

	u.metadata = make(map[string]any)
	for k, v := range model.Metadata {
		u.metadata[k] = v
	}

	return u
}

type ExUniverse struct {
	// initialized true when the universe is initialized
	// the universe is initialized when the first operation is executed
	initialized bool

	// universeContext is the universe context
	universeContext any

	// model of the ExUniverse
	model *theoretical.UniverseModel

	// machineLawsExecutor is the machine laws executor
	constantsLawsExecutor instrumentation.ConstantsLawsExecutor

	// currentReality is the current reality of the ExUniverse
	currentReality *string

	// realityBeforeSuperposition is the reality before the ExUniverse entered in superposition
	realityBeforeSuperposition *string

	// isFinalReality true when the current reality type belongs to the final states
	// used to know when the ExUniverse has been exited and finalized
	isFinalReality bool

	// realityInitialized is true when the "entry operation" of the current reality are executed
	realityInitialized bool

	// inSuperposition is true if the ExUniverse is in superposition
	inSuperposition bool

	// externalTargets is the list of external targets
	// used for ExQuantumMachine to send events to external targets
	// externalTargets are cleared on each interaction with the ExUniverse
	externalTargets []string

	// eventAccumulator universe Event accumulator
	// used to accumulate events for each reality when the ExUniverse is in superposition (inSuperposition == true)
	eventAccumulator instrumentation.Accumulator

	// tracking
	tracking []string

	// metadata
	metadata   map[string]any
	metadataMu sync.Mutex

	// emitDepth tracks the current nesting depth of processEmittedEvents calls.
	// Used to prevent infinite loops when emitted events cause transitions whose
	// entry actions emit more events. Maximum depth is maxEmitDepth.
	emitDepth int

	// getSnapshotFn returns a snapshot without taking the machine mutex.
	// Used from actions that already run under quantumMachineMtx.
	options                instrumentation.RuntimeOptions
	invokes                *invokeManager
	owner                  *ExQuantumMachine
	invokeError            error // Written only by synchronous admission under the machine lock.
	getSnapshotFn          func() *instrumentation.MachineSnapshot
	getSnapshotWithErrorFn func() (*instrumentation.MachineSnapshot, error)
}

//------------------------------- External Operations -------------------------------//

// handleEvent handles an Event where depending on the state of the universe
func (u *ExUniverse) handleEvent(ctx context.Context, realityName *string, evt instrumentation.Event, universeContext any) ([]string, error) {
	var handleEventFn func() error
	u.setUniverseContext(universeContext)

	if realityName != nil {
		handleEventFn = func() error { return u.receiveEventToReality(ctx, *realityName, evt) }
	} else {
		handleEventFn = func() error { return u.receiveEvent(ctx, evt) }
	}

	externalTargets, err := u.universeDecorator(handleEventFn)
	if err != nil {
		err = errors.Join(fmt.Errorf("universe '%s'. error handling event '%s'", u.model.ID, evt.GetEventName()), err)
	}
	return externalTargets, err
}

// start starts the universe on the default reality (initial reality)
// start set initial reality as the current reality and execute:
// - always operations
// - initial operations
func (u *ExUniverse) start(ctx context.Context, universeContext any, event instrumentation.Event) ([]string, instrumentation.Event, error) {
	u.setUniverseContext(universeContext)
	if event == nil {
		event = NewEventBuilder(startEventName).
			SetEvtType(instrumentation.EventTypeStart).
			Build()
	}

	var initFn = func() error {
		if u.model.Initial == nil {
			u.initOnSuperposition()
			return nil
		}

		if err := u.initializeUniverseOn(ctx, *u.model.Initial, event); err != nil {
			return errors.Join(fmt.Errorf(initializingUniverseErrMsgTemplate, u.model.ID), err)
		}
		return nil
	}

	externalTargets, err := u.universeDecorator(initFn)
	return externalTargets, event, err
}

// startOnReality starts the universe on the given reality
// startOnReality set the given reality as the current reality and execute:
// - always operations
// - initial operations
func (u *ExUniverse) startOnReality(ctx context.Context, realityName string, universeContext any, event instrumentation.Event) ([]string, instrumentation.Event, error) {
	u.setUniverseContext(universeContext)
	if event == nil {
		event = NewEventBuilder(startOnEventName).
			SetEvtType(instrumentation.EventTypeStartOn).
			Build()
	}

	var initFn = func() error {
		if err := u.initializeUniverseOn(ctx, realityName, event); err != nil {
			return errors.Join(fmt.Errorf(initializingUniverseErrMsgTemplate, u.model.ID), err)
		}
		return nil
	}

	externalTargets, err := u.universeDecorator(initFn)
	return externalTargets, event, err
}

// getSnapshot returns a snapshot of the universe
func (u *ExUniverse) getSnapshot() instrumentation.SerializedUniverseSnapshot {
	snapshot, err := u.getSnapshotWithError()
	if err != nil {
		slog.Error("snapshot capture failed", "universe", u.model.ID, "error", err)
	}
	return snapshot
}

func (u *ExUniverse) getSnapshotWithError() (instrumentation.SerializedUniverseSnapshot, error) {
	u.metadataMu.Lock()
	metadataCopy := cloneAnyMap(u.metadata)
	u.metadataMu.Unlock()

	var infoSnapshot = UniverseInfoSnapshot{
		ID:                         u.model.ID,
		CanonicalName:              u.model.CanonicalName,
		Version:                    u.model.Version,
		Initialized:                u.initialized,
		CurrentReality:             u.currentReality,
		RealityInitialized:         u.realityInitialized,
		InSuperposition:            u.inSuperposition,
		RealityBeforeSuperposition: u.realityBeforeSuperposition,
		Metadata:                   metadataCopy,
	}

	if u.eventAccumulator != nil {
		if accumulator, ok := u.eventAccumulator.(*eventAccumulator); ok {
			infoSnapshot.Accumulator = accumulator
		}
	}

	m, err := util.StructToMap(infoSnapshot)
	return m, err
}

// decodeSnapshot validates a detached snapshot without changing live state.
func (u *ExUniverse) decodeSnapshot(universeSnapshot instrumentation.SerializedUniverseSnapshot) (*UniverseInfoSnapshot, error) {
	if universeSnapshot == nil {
		return nil, fmt.Errorf("snapshot for universe '%s' must not be nil", u.model.ID)
	}
	snapshot, err := util.MapToStructWithNumbers[UniverseInfoSnapshot](universeSnapshot)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error loading snapshot for universe '%s'", u.model.ID), err)
	}
	util.NormalizeJSONNumbers(snapshot.Metadata)
	if snapshot.Accumulator != nil {
		for _, events := range snapshot.Accumulator.RealitiesEvents {
			for _, event := range events {
				if event != nil {
					util.NormalizeJSONNumbers(event.Data)
				}
			}
		}
	}
	invalid := func(reason string) (*UniverseInfoSnapshot, error) {
		return nil, fmt.Errorf("invalid snapshot for universe '%s': %s", u.model.ID, reason)
	}
	if snapshot.ID != "" && snapshot.ID != u.model.ID {
		return invalid("universe ID does not match")
	}
	if snapshot.InSuperposition && (!snapshot.Initialized || snapshot.CurrentReality != nil) {
		return invalid("superposition requires an initialized universe without a current reality")
	}
	if snapshot.Initialized && !snapshot.InSuperposition && snapshot.CurrentReality == nil {
		return invalid("initialized universe requires a current reality or superposition")
	}
	if !snapshot.Initialized && (snapshot.CurrentReality != nil || snapshot.RealityInitialized || snapshot.RealityBeforeSuperposition != nil) {
		return invalid("uninitialized universe cannot contain an active reality")
	}
	for _, reality := range []*string{snapshot.CurrentReality, snapshot.RealityBeforeSuperposition} {
		if reality == nil {
			continue
		}
		model, err := u.getRealityModel(*reality)
		if err != nil {
			return nil, err
		}
		if model == nil {
			return invalid("reality does not exist")
		}
	}
	if snapshot.Accumulator != nil {
		if limit := u.options.MaxAccumulatedEvents; limit > 0 && snapshot.Accumulator.CountAllEvents() > limit {
			return nil, &instrumentation.ResourceLimitError{Resource: "accumulated events", Limit: limit}
		}
		for reality, events := range snapshot.Accumulator.RealitiesEvents {
			model, err := u.getRealityModel(reality)
			if err != nil {
				return nil, err
			}
			if model == nil {
				return invalid("accumulator reality does not exist")
			}
			for _, event := range events {
				if event == nil {
					return invalid("accumulator contains a nil event")
				}
			}
		}
	}
	// Older/empty superposition snapshots may omit their empty accumulator.
	if snapshot.InSuperposition && snapshot.Accumulator == nil {
		snapshot.Accumulator = &eventAccumulator{RealitiesEvents: make(map[string][]*Event)}
	}
	return snapshot, nil
}

func (u *ExUniverse) applySnapshot(snapshot *UniverseInfoSnapshot) {
	u.initialized = snapshot.Initialized
	u.currentReality = snapshot.CurrentReality
	u.realityInitialized = snapshot.RealityInitialized
	u.inSuperposition = snapshot.InSuperposition
	u.realityBeforeSuperposition = snapshot.RealityBeforeSuperposition
	u.eventAccumulator = nil
	if snapshot.Accumulator != nil {
		u.eventAccumulator = snapshot.Accumulator
	}
	// Keep the map identity: in-flight invoke arguments share this map and mutex.
	u.metadataMu.Lock()
	if len(snapshot.Metadata) > 0 && u.metadata == nil {
		u.metadata = make(map[string]any)
	}
	clear(u.metadata)
	for k, v := range snapshot.Metadata {
		u.metadata[k] = v
	}
	u.metadataMu.Unlock()

	u.isFinalReality = false
	reality := u.currentReality
	if u.inSuperposition {
		reality = u.realityBeforeSuperposition
	}
	if reality != nil {
		realityModel, _ := u.getRealityModel(*reality) // already validated
		u.isFinalReality = theoretical.IsFinalState(realityModel.Type)
	}
}

// canHandleEvent returns true if a concrete (non-superposition, non-final) reality
// has an On handler for the event. Superposition universes are selected separately
// by getLazyActiveUniverses when no concrete handler exists, so a directed
// cross-universe transition does not also broadcast into the target.
func (u *ExUniverse) canHandleEvent(evt instrumentation.Event) bool {
	if !u.initialized || u.inSuperposition || u.isFinalReality {
		return false
	}
	if u.currentReality != nil && u.canRealityHandleEvent(*u.currentReality, evt) {
		return true
	}
	return false
}

func (u *ExUniverse) canAccumulateInSuperposition() bool {
	return u.initialized && u.inSuperposition
}

// replayOnEntry replays the on entry process for the current reality
func (u *ExUniverse) replayOnEntry(ctx context.Context, evt instrumentation.Event, universeContext any) error {
	u.setUniverseContext(universeContext)
	return u.executeOnEntry(ctx, evt)
}

// positionStatic positions the universe on the specified reality without executing any actions.
// This method only sets the universe state flags and does not trigger entry actions, always transitions,
// or any other execution flows. It's designed for testing, debugging, and scenarios where you want
// to place the machine in a specific state without side effects.
func (u *ExUniverse) positionStatic(realityID string, universeContext any) error {
	u.setUniverseContext(universeContext)

	// Validate reality exists
	realityModel, err := u.getRealityModel(realityID)
	if err != nil {
		return err
	}

	// Set universe as initialized
	u.initialized = true

	// Set current reality directly
	if u.options.CancelInvokesOnExit && u.invokes != nil {
		u.invokes.cancel(u.model.ID, "")
	}
	u.currentReality = &realityID
	u.addStateToTracking(u.currentReality)

	// Set reality flags
	u.realityInitialized = true
	u.isFinalReality = theoretical.IsFinalState(realityModel.Type)

	// Clear superposition state
	u.inSuperposition = false
	u.realityBeforeSuperposition = nil
	u.eventAccumulator = nil

	return nil
}

// isActive returns true if the universe is active
func (u *ExUniverse) isActive() bool {
	return u.initialized && !u.inSuperposition && !u.isFinalReality
}

//------------------------------- Internal Operations -------------------------------//

func (u *ExUniverse) setUniverseContext(universeContext any) {
	u.universeContext = universeContext
}

func (u *ExUniverse) addStateToTracking(state *string) {
	if state == nil {
		return
	}
	u.tracking = u.retainTracking(append(u.tracking, *state))
}

func (u *ExUniverse) retainTracking(entries []string) []string {
	limit := u.options.MaxTrackingEntries
	if limit > 0 && len(entries) > limit {
		return append([]string(nil), entries[len(entries)-limit:]...)
	}
	return entries
}

func (u *ExUniverse) popTrackingIfLast(state string) {
	if len(u.tracking) == 0 {
		return
	}
	if u.tracking[len(u.tracking)-1] == state {
		u.tracking = u.tracking[:len(u.tracking)-1]
	}
}

func (u *ExUniverse) universeDecorator(operation func() error) ([]string, error) {
	// clear externalTargets
	u.externalTargets = nil

	// execute operation
	if err := operation(); err != nil {
		return nil, err
	}

	// return externalTargets
	return u.externalTargets, nil
}

// receiveEventToReality receives an Event only if one of the following conditions is met:
//   - the Universe is in superposition.
//   - if Universe is not initialized:
//     -- if reality is the initial reality and has no observers -> initialize universe on reality.
//     -- otherwise, initialize universe on superposition.
//   - not in superposition but the current reality is the target reality and not a final reality
func (u *ExUniverse) receiveEventToReality(ctx context.Context, realityName string, event instrumentation.Event) error {
	reality, ok := u.model.Realities[realityName]
	if !ok {
		return fmt.Errorf(realitiesNotExistErrMsgTemplate, u.model.ID, realityName)
	}

	// if not initialized
	if !u.initialized {
		// if realityName is the initial reality and has no observers -> initialize universe on reality
		if u.model.Initial != nil && realityName == *u.model.Initial && len(reality.Observers) == 0 {
			return u.initializeUniverseOn(ctx, realityName, event)
		}
		u.initOnSuperposition()
	}

	// handling superposition
	if u.inSuperposition {
		isNewReality, err := u.accumulateEventForReality(ctx, realityName, event, true)
		if err != nil {
			return errors.Join(fmt.Errorf("error accumulating Event for reality '%s'", realityName), err)
		}

		if isNewReality {
			// establish new realityName
			if err = u.establishNewReality(ctx, realityName, event); err != nil {
				return errors.Join(fmt.Errorf("error establishing new reality '%s'", realityName), err)
			}
		}

		return nil
	}

	// handling if current reality is the target reality
	if *u.currentReality == realityName {
		return u.onEvent(ctx, event)
	}

	// return error
	var currentRealityName string
	if u.currentReality != nil {
		currentRealityName = *u.currentReality
	}
	return fmt.Errorf(
		"universe '%s' can't receive Event '%s' to reality '%s'. inSuperposition: '%t', currentReality: '%s', IsFinalReality: '%t'",
		u.model.ID, event.GetEventName(), realityName, u.inSuperposition, currentRealityName, u.isFinalReality,
	)
}

// receiveEvent receives an Event only if one of the following conditions is met:
//   - the Universe is in superposition (the Event will be accumulated for each reality)
//   - if Universe is not initialized:
//     -- the Universe is initialized on the initial reality.
//   - not in superposition and the current reality not a final reality
func (u *ExUniverse) receiveEvent(ctx context.Context, event instrumentation.Event) error {
	// handling not initialized universe and initial reality
	if !u.initialized && u.model.Initial != nil {
		if err := u.receiveEventToReality(ctx, *u.model.Initial, event); err != nil {
			return errors.Join(fmt.Errorf(initializingUniverseErrMsgTemplate, u.model.ID), err)
		}
		return nil
	}

	// handling not initialized universe and not initial reality
	if !u.initialized && u.model.Initial == nil {
		u.initOnSuperposition()
	}

	// handling superposition
	if u.inSuperposition {
		isNewReality, realityName, err := u.accumulateEventForAllRealities(ctx, event)
		if err != nil {
			return errors.Join(fmt.Errorf("error accumulating Event for all realities"), err)
		}

		if !isNewReality {
			return nil
		}

		// establish new reality
		if err = u.establishNewReality(ctx, realityName, event); err != nil {
			return errors.Join(fmt.Errorf("error establishing new reality '%s'", realityName), err)
		}

		return nil
	}

	// handling not final current reality
	return u.onEvent(ctx, event)
}

func (u *ExUniverse) onEvent(ctx context.Context, event instrumentation.Event) error {
	realityModel, err := u.getRealityModel(*u.currentReality)
	if err != nil {
		return errors.Join(fmt.Errorf("error getting reality '%s'", *u.currentReality), err)
	}

	transitions, ok := realityModel.On[event.GetEventName()]
	if !ok {
		return fmt.Errorf("reality '%s' does not have transitions for Event '%s'", realityModel.ID, event.GetEventName())
	}

	onApprovedTransition, err := u.getApprovedTransition(ctx, transitions, event)
	if err != nil {
		return errors.Join(fmt.Errorf("error executing on transitions for reality '%s'", realityModel.ID), err)
	}

	if onApprovedTransition == nil {
		return nil
	}

	// execute cyclic transition while there are approved transitions
	if err = u.doCyclicTransition(ctx, onApprovedTransition, event); err != nil {
		return errors.Join(fmt.Errorf("error executing on transitions for reality '%s'", realityModel.ID), err)
	}

	return nil
}

func (u *ExUniverse) initializeUniverseOn(ctx context.Context, realityName string, event instrumentation.Event) error {
	u.initialized = true

	if err := u.establishNewReality(ctx, realityName, event); err != nil {
		u.initialized = false
		u.currentReality = nil
		u.realityInitialized = false
		u.isFinalReality = false
		u.inSuperposition = false
		return errors.Join(fmt.Errorf("error establishing initial reality '%s'", realityName), err)
	}

	return nil
}

func (u *ExUniverse) establishNewReality(ctx context.Context, reality string, event instrumentation.Event) error {
	previousTracking := u.tracking
	previousReality := u.currentReality
	previousFinal := u.isFinalReality
	previousRealityInitialized := u.realityInitialized
	previousSuperposition := u.inSuperposition
	previousBeforeSuperposition := u.realityBeforeSuperposition

	u.currentReality = &reality
	u.addStateToTracking(u.currentReality)
	if err := u.executeOnEntry(ctx, event); err != nil {
		u.currentReality = previousReality
		u.isFinalReality = previousFinal
		u.realityInitialized = previousRealityInitialized
		u.inSuperposition = previousSuperposition
		u.realityBeforeSuperposition = previousBeforeSuperposition
		u.tracking = previousTracking
		return errors.Join(fmt.Errorf(errorExecutingOnEntryProcessMsgTemplate, u.model.ID, reality), err)
	}
	u.realityInitialized = true

	// emitted events during entry may have changed currentReality (e.g. initSuperposition sets it to nil)
	if u.currentReality == nil {
		return nil
	}

	// clear Event accumulator
	u.eventAccumulator = nil

	// quit superposition
	u.inSuperposition = false
	u.realityBeforeSuperposition = nil

	// re-read reality from currentReality — emitted events during entry may have changed it
	realityModel, err := u.getRealityModel(*u.currentReality)
	if err != nil {
		return err
	}

	u.isFinalReality = theoretical.IsFinalState(realityModel.Type)

	// execute always
	if err = u.executeAlways(ctx, realityModel, event); err != nil {
		return errors.Join(fmt.Errorf("error executing always transitions for universe '%s'", u.model.ID), err)
	}

	return nil
}

func (u *ExUniverse) executeOnEntry(ctx context.Context, event instrumentation.Event) error {
	if err := u.executeOnEntryProcess(ctx, event); err != nil {
		return errors.Join(fmt.Errorf(errorExecutingOnEntryProcessMsgTemplate, u.model.ID, *u.currentReality), err)
	}
	return nil
}

func (u *ExUniverse) executeAlways(ctx context.Context, realityModel *theoretical.RealityModel, event instrumentation.Event) error {
	approvedTransition, err := u.getApprovedTransition(ctx, realityModel.Always, event)
	if err != nil {
		return errors.Join(fmt.Errorf(errorExecutingAlwaysTransitionsMsgTemplate, realityModel.ID), err)
	}

	// execute cyclic transition while there are approved transitions
	if err = u.doCyclicTransition(ctx, approvedTransition, event); err != nil {
		return errors.Join(fmt.Errorf(errorExecutingAlwaysTransitionsMsgTemplate, realityModel.ID), err)
	}

	return nil
}

// doCyclicTransition executes transition and always transitions of the current reality and goes to the next reality if necessary
// doCyclicTransition ends when:
//   - there are no approved transition
//   - there are approved transition but no targets
//   - There are approved transition but the target is another universe
//   - There are approved transition, but it is a notification transition
func (u *ExUniverse) doCyclicTransition(
	ctx context.Context, approvedTransition *theoretical.TransitionModel, event instrumentation.Event,
) error {
	visitedTargets := map[string]int{}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if approvedTransition == nil || len(approvedTransition.Targets) == 0 {
			return nil
		}

		args := instrumentation.QuantumMachineExecutorArgs{
			Context:               u.universeContext,
			RealityName:           *u.currentReality,
			UniverseID:            u.model.ID,
			UniverseCanonicalName: u.model.CanonicalName,
			Event:                 event,
		}

		if err := u.constantsLawsExecutor.ExecuteTransitionAction(ctx, &args); err != nil {
			return errors.Join(fmt.Errorf("error executing constants transition actions for reality '%s'", *u.currentReality), err)
		}

		if err := u.executeUniverseConstantActions(ctx, instrumentation.ActionTypeTransition, event, nil); err != nil {
			return errors.Join(fmt.Errorf("error executing universe constant transition actions for reality '%s'", *u.currentReality), err)
		}

		if err := u.executeActions(ctx, approvedTransition.Actions, event, instrumentation.ActionTypeTransition, nil); err != nil {
			return errors.Join(fmt.Errorf("error executing transition actions for reality '%s'", *u.currentReality), err)
		}

		if err := u.executeInvokeGroup(ctx, "transition", approvedTransition.Invokes, event, &args); err != nil {
			return err
		}

		if approvedTransition.IsNotification() {
			u.externalTargets = approvedTransition.Targets
			return nil
		}

		if err := u.executeOnExitProcess(ctx, event); err != nil {
			return errors.Join(fmt.Errorf("error executing on exit process for universe '%s' and reality '%s'", u.model.ID, *u.currentReality), err)
		}

		if len(approvedTransition.Targets) > 1 {
			return u.initSuperposition(approvedTransition.Targets)
		}

		refTyp, _, err := processReference(approvedTransition.Targets[0])
		if err != nil {
			return errors.Join(fmt.Errorf("error processing reference '%s'", approvedTransition.Targets[0]), err)
		}
		if refTyp != RefTypeReality {
			return u.initSuperposition(approvedTransition.Targets)
		}

		next := approvedTransition.Targets[0]
		visitedTargets[next]++
		if visitedTargets[next] > 1 {
			return fmt.Errorf(
				"cyclic transition detected in universe '%s': reality '%s' visited more than once in the same cascade",
				u.model.ID, next,
			)
		}

		previousTracking := u.tracking
		previousReality := u.currentReality
		previousFinal := u.isFinalReality

		u.currentReality = &approvedTransition.Targets[0]
		u.addStateToTracking(u.currentReality)

		if err := u.executeOnEntry(ctx, event); err != nil {
			u.currentReality = previousReality
			u.isFinalReality = previousFinal
			u.realityInitialized = previousReality != nil
			u.tracking = previousTracking
			return errors.Join(fmt.Errorf(errorExecutingOnEntryProcessMsgTemplate, u.model.ID, next), err)
		}

		if u.currentReality == nil || u.inSuperposition {
			return nil
		}

		realityModel, err := u.getRealityModel(*u.currentReality)
		if err != nil {
			return err
		}
		u.isFinalReality = theoretical.IsFinalState(realityModel.Type)

		if approvedTransition, err = u.getApprovedTransition(ctx, realityModel.Always, event); err != nil {
			return errors.Join(fmt.Errorf(errorExecutingAlwaysTransitionsMsgTemplate, realityModel.ID), err)
		}
	}
}

func (u *ExUniverse) getApprovedTransition(
	ctx context.Context, transitionModels []*theoretical.TransitionModel, event instrumentation.Event,
) (*theoretical.TransitionModel, error) {
	if len(transitionModels) == 0 {
		return nil, nil
	}

	for _, transition := range transitionModels {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var conditions []*theoretical.ConditionModel

		if transition.Condition != nil {
			conditions = append(conditions, transition.Condition)
		}

		if transition.Conditions != nil {
			conditions = append(conditions, transition.Conditions...)
		}

		doTransition, err := u.executeConditions(ctx, event, conditions)
		if err != nil {
			condSrc := "<unknown>"
			if transition.Condition != nil {
				condSrc = transition.Condition.Src
			} else if len(transition.Conditions) > 0 {
				condSrc = transition.Conditions[0].Src
			}
			return nil, errors.Join(fmt.Errorf("error executing transition condition '%s'", condSrc), err)
		}

		if doTransition {
			return transition, nil
		}
	}

	return nil, nil
}

func (u *ExUniverse) executeConditions(
	ctx context.Context, event instrumentation.Event, conditionsModel []*theoretical.ConditionModel,
) (bool, error) {
	// if conditionsModel is empty then return true, the transition is always executed
	if len(conditionsModel) == 0 {
		return true, nil
	}

	args := &conditionExecutorArgs{
		context:               u.universeContext,
		realityName:           *u.currentReality,
		universeCanonicalName: u.model.CanonicalName,
		universeID:            u.model.ID,
		universeMetadata:      u.metadata,
		metadataMu:            &u.metadataMu,
		event:                 event,
	}

	for _, conditionModel := range conditionsModel {
		args.condition = *conditionModel
		doTransition, err := u.runConditionExecutor(ctx, args)
		if err != nil {
			return false, errors.Join(fmt.Errorf("error executing condition '%s'", conditionModel.Src), err)
		}

		if !doTransition {
			return false, nil
		}
	}

	return true, nil
}

func (u *ExUniverse) initSuperposition(targets []string) error {
	// set superposition
	u.realityBeforeSuperposition = u.currentReality
	u.currentReality = nil
	u.inSuperposition = true
	u.externalTargets = targets
	u.eventAccumulator = newEventAccumulator()
	return nil
}

// initOnSuperposition initializes the universe directly into superposition.
// This is the correct path for universes without an Initial reality.
// No reality is established and no always transitions execute until an
// observer approves a reality via establishNewReality.
func (u *ExUniverse) initOnSuperposition() {
	u.initialized = true
	u.realityBeforeSuperposition = nil
	u.currentReality = nil
	u.inSuperposition = true
	u.externalTargets = nil
	u.eventAccumulator = newEventAccumulator()
}

func (u *ExUniverse) executeOnEntryProcess(ctx context.Context, event instrumentation.Event) error {
	// if current reality is nil -> return
	if u.currentReality == nil {
		return nil
	}

	// if reality is already initialized and the Event is not a replay on entry -> return
	if u.realityInitialized && !event.GetFlags().ReplayOnEntry {
		return nil
	}

	realityModel, err := u.getRealityModel(*u.currentReality)
	if err != nil {
		return err
	}

	// collector for emitted events from entry actions
	var emittedEvents []instrumentation.EmittedEvent

	// execute on constants entry actions
	args := &instrumentation.QuantumMachineExecutorArgs{
		Context:               u.universeContext,
		RealityName:           realityModel.ID,
		UniverseCanonicalName: u.model.CanonicalName,
		UniverseID:            u.model.ID,
		Event:                 event,
		EmittedEvents:         &emittedEvents,
	}
	if err = u.constantsLawsExecutor.ExecuteEntryAction(ctx, args); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on entry machine actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	if err = u.executeUniverseConstantActions(ctx, instrumentation.ActionTypeEntry, event, &emittedEvents); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on entry universe constant actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	// execute on entry reality actions (pass emittedEvents collector)
	if err = u.executeActions(ctx, realityModel.EntryActions, event, instrumentation.ActionTypeEntry, &emittedEvents); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on entry actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	// execute on entry constants invokes, invokes are executed asynchronously
	if err := u.executeInvokeGroup(ctx, "entry", realityModel.EntryInvokes, event, args); err != nil {
		return err
	}

	u.realityInitialized = true

	// process emitted events after all entry actions complete and reality is initialized
	if len(emittedEvents) > 0 {
		if err = u.processEmittedEvents(ctx, emittedEvents); err != nil {
			return errors.Join(
				fmt.Errorf("error processing emitted events for reality '%s'", realityModel.ID),
				err,
			)
		}
	}

	return nil
}

// processEmittedEvents processes events emitted by entry actions via EmitEvent.
// Events are processed in FIFO order. For each event, it looks up the current reality's
// On handlers. If a handler exists and its conditions pass, the transition is executed
// via doCyclicTransition. The first event that triggers a transition wins; remaining
// events are discarded.
//
// Nesting is tracked via emitDepth to prevent infinite loops (max depth: maxEmitDepth).
// If the depth limit is exceeded, the error propagates up and the universe is left in an
// inconsistent state (partially mutated). The caller should treat the machine instance as
// invalid. This is an unrecoverable condition that indicates a bug in the state definition.
func (u *ExUniverse) processEmittedEvents(
	ctx context.Context, emittedEvents []instrumentation.EmittedEvent,
) error {
	u.emitDepth++
	defer func() { u.emitDepth-- }()

	if u.emitDepth > maxEmitDepth {
		realityName := "<nil>"
		if u.currentReality != nil {
			realityName = *u.currentReality
		}
		return fmt.Errorf(
			"emitted event depth exceeded maximum (%d) in universe '%s', reality '%s' — possible infinite loop",
			maxEmitDepth, u.model.ID, realityName,
		)
	}

	if u.currentReality == nil {
		return nil
	}
	realityModel, err := u.getRealityModel(*u.currentReality)
	if err != nil {
		return err
	}

	for _, emitted := range emittedEvents {
		if err := ctx.Err(); err != nil {
			return err
		}
		transitions, ok := realityModel.On[emitted.Name]
		if !ok {
			continue
		}

		// build an event from the emitted data
		evt := NewEventBuilder(emitted.Name).
			SetEvtType(instrumentation.EventTypeEmitted).
			SetData(emitted.Data).
			Build()

		approvedTransition, err := u.getApprovedTransition(ctx, transitions, evt)
		if err != nil {
			return errors.Join(fmt.Errorf("error evaluating emitted event '%s' conditions", emitted.Name), err)
		}

		if approvedTransition == nil {
			continue
		}

		// first approved transition wins — execute and stop processing remaining events
		if err = u.doCyclicTransition(ctx, approvedTransition, evt); err != nil {
			return errors.Join(fmt.Errorf("error executing transition for emitted event '%s'", emitted.Name), err)
		}
		break
	}

	return nil
}

func (u *ExUniverse) executeOnExitProcess(ctx context.Context, event instrumentation.Event) error {
	if u.currentReality == nil {
		return nil
	}

	realityModel, err := u.getRealityModel(*u.currentReality)
	if err != nil {
		return err
	}

	// execute on exit constants actions
	args := &instrumentation.QuantumMachineExecutorArgs{
		Context:               u.universeContext,
		RealityName:           realityModel.ID,
		UniverseCanonicalName: u.model.CanonicalName,
		UniverseID:            u.model.ID,
		Event:                 event,
	}

	if err = u.constantsLawsExecutor.ExecuteExitAction(ctx, args); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on exit machine actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	if err = u.executeUniverseConstantActions(ctx, instrumentation.ActionTypeExit, event, nil); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on exit universe constant actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	if err = u.executeActions(ctx, realityModel.ExitActions, event, instrumentation.ActionTypeExit, nil); err != nil {
		return errors.Join(
			fmt.Errorf("error executing on exit actions for reality '%s'", realityModel.ID),
			err,
		)
	}

	if u.options.CancelInvokesOnExit && u.invokes != nil {
		u.invokes.cancel(u.model.ID, realityModel.ID)
	}
	if err := u.executeInvokeGroup(ctx, "exit", realityModel.ExitInvokes, event, args); err != nil {
		return err
	}

	u.realityInitialized = false
	return nil
}

func (u *ExUniverse) executeActions(
	ctx context.Context,
	actionModels []*theoretical.ActionModel,
	event instrumentation.Event,
	actionType instrumentation.ActionType,
	emittedEvents *[]instrumentation.EmittedEvent,
) error {
	if len(actionModels) == 0 {
		return nil
	}

	// execute actions
	for _, action := range actionModels {
		if err := ctx.Err(); err != nil {
			return err
		}
		args := &actionExecutorArgs{
			context:                u.universeContext,
			realityName:            *u.currentReality,
			universeCanonicalName:  u.model.CanonicalName,
			universeID:             u.model.ID,
			universeMetadata:       u.metadata,
			metadataMu:             &u.metadataMu,
			event:                  event,
			action:                 *action,
			actionType:             actionType,
			getSnapshotFn:          u.snapshotProvider(),
			getSnapshotWithErrorFn: u.snapshotWithErrorProvider(),
			emittedEvents:          emittedEvents,
		}
		if err := u.runActionExecutor(ctx, action.Src, args); err != nil {
			return errors.Join(fmt.Errorf("error executing action '%s'", action.Src), err)
		}
	}

	return nil
}

func (u *ExUniverse) executeInvokes(
	ctx context.Context, invokeModels []*theoretical.InvokeModel, event instrumentation.Event,
) {
	if len(invokeModels) == 0 {
		return
	}

	// execute invokes
	for _, invoke := range invokeModels {
		args := &invokeExecutorArgs{
			context:               u.universeContext,
			realityName:           *u.currentReality,
			universeCanonicalName: u.model.CanonicalName,
			universeID:            u.model.ID,
			universeMetadata:      u.metadata,
			metadataMu:            &u.metadataMu,
			event:                 event,
			invoke:                *invoke,
		}
		u.runInvokeExecutor(ctx, args)
	}
}

func (u *ExUniverse) accumulateEventForReality(
	ctx context.Context, realityName string, event instrumentation.Event, directEvent bool,
) (bool, error) {
	realityModel, err := u.getRealityModel(realityName)
	if err != nil {
		return false, err
	}

	// if reality has no observers -> return false
	if len(realityModel.Observers) == 0 {
		if directEvent {
			return true, nil
		}
		return false, nil
	}

	if err := u.validateObservers(realityModel); err != nil {
		return false, err
	}
	if err := u.checkAccumulatorBudget(1); err != nil {
		return false, err
	}
	// accumulate Event
	u.eventAccumulator.Accumulate(realityName, event)

	// execute observers
	isNewReality, err := u.executeObservers(ctx, realityModel, event)
	if err != nil {
		return false, errors.Join(fmt.Errorf("error executing observers for reality '%s'", realityModel.ID), err)
	}

	return isNewReality, nil
}

func (u *ExUniverse) accumulateEventForAllRealities(ctx context.Context, event instrumentation.Event) (bool, string, error) {
	// Reserve the worst-case fan-out before invoking observers or adding any copies.
	copies := 0
	for _, model := range u.model.Realities {
		if err := u.validateObservers(model); err != nil {
			return false, "", err
		}
		if len(model.Observers) > 0 {
			copies++
		}
	}
	if err := u.checkAccumulatorBudget(copies); err != nil {
		return false, "", err
	}
	for _, reality := range sortedMapKeys(u.model.Realities) {
		isNewReality, err := u.accumulateEventForReality(ctx, reality, event, false)
		if err != nil {
			return false, "", errors.Join(fmt.Errorf("error accumulating Event for reality '%s'", reality), err)
		}

		if isNewReality {
			return true, reality, nil
		}
	}

	return false, "", nil
}

func (u *ExUniverse) executeObservers(
	ctx context.Context, realityModel *theoretical.RealityModel, event instrumentation.Event,
) (bool, error) {
	var firstErr error

	for _, observer := range realityModel.Observers {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		args := &observerExecutorArgs{
			context:               u.universeContext,
			realityName:           realityModel.ID,
			universeCanonicalName: u.model.CanonicalName,
			universeID:            u.model.ID,
			universeMetadata:      u.metadata,
			metadataMu:            &u.metadataMu,
			accumulatorStatistics: u.eventAccumulator.GetStatistics(),
			event:                 event,
			observer:              *observer,
		}
		isApproved, err := u.runObserverExecutor(ctx, observer.Src, args)
		if err != nil {
			joined := errors.Join(fmt.Errorf("error executing observer '%s'", observer.Src), err)
			if firstErr == nil {
				firstErr = joined
			}
			continue
		}

		if isApproved {
			return true, nil
		}
	}

	return false, firstErr
}

func (u *ExUniverse) canRealityHandleEvent(realityName string, evt instrumentation.Event) bool {
	realityModel, err := u.getRealityModel(realityName)
	if err != nil {
		return false
	}

	if realityModel == nil {
		return false
	}

	_, ok := realityModel.On[evt.GetEventName()]
	return ok
}

//------------- executors -------------

func (u *ExUniverse) runObserverExecutor(ctx context.Context, src string, args *observerExecutorArgs) (bool, error) {
	if u.options.StrictObservers && (src == "" || builtin.GetObserver(src) == nil) {
		return false, fmt.Errorf("%w: %q", instrumentation.ErrUnknownObserver, src)
	}
	if src == "" {
		return true, nil
	}

	if err := ctx.Err(); err != nil {
		return false, err
	}
	if fn := builtin.GetObserver(src); fn != nil {
		callbackCtx, done := u.owner.callbackContext(ctx)
		defer done()
		approved, err := fn(callbackCtx, args)
		if err != nil {
			return false, err
		}
		if err := ctx.Err(); err != nil {
			return false, err
		}
		return approved, nil
	}

	slog.WarnContext(ctx, "observer not found (default: return true)", "src", src)
	return true, nil
}

func (u *ExUniverse) runActionExecutor(ctx context.Context, src string, args *actionExecutorArgs) error {
	if src == "" {
		return nil
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if fn := builtin.GetAction(src); fn != nil {
		callbackCtx, done := u.owner.callbackContext(ctx)
		defer done()
		err := fn(callbackCtx, args)
		if err != nil {
			return err
		}
		return ctx.Err()
	}

	slog.WarnContext(ctx, "action not found", "src", src)
	return nil
}

func (u *ExUniverse) runInvokeExecutor(ctx context.Context, args *invokeExecutorArgs) {
	if u.invokeError != nil || args.invoke.Src == "" {
		return
	}
	if fn := builtin.GetInvoke(args.invoke.Src); fn != nil {
		if u.invokes == nil {
			u.invokes = newInvokeManager(0)
		}
		u.invokeError = u.invokes.start(ctx, u.model.ID, args.realityName, args.invoke.Src, func(ctx context.Context) { fn(ctx, args) })
		if u.invokeError != nil {
			slog.ErrorContext(ctx, "invoke admission failed", "src", args.invoke.Src, "error", u.invokeError)
		}
		return
	}
	slog.WarnContext(ctx, "invoke not found", "src", args.invoke.Src)
}

func (u *ExUniverse) executeInvokeGroup(ctx context.Context, phase string, invokes []*theoretical.InvokeModel, event instrumentation.Event, args *instrumentation.QuantumMachineExecutorArgs) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	u.invokeError = nil
	switch phase {
	case "entry":
		u.constantsLawsExecutor.ExecuteEntryInvokes(ctx, args)
	case "exit":
		u.constantsLawsExecutor.ExecuteExitInvokes(ctx, args)
	case "transition":
		u.constantsLawsExecutor.ExecuteTransitionInvokes(ctx, args)
	}
	u.executeUniverseConstantInvokes(ctx, phase, event)
	u.executeInvokes(ctx, invokes, event)
	return u.invokeError
}

func (u *ExUniverse) validateObservers(model *theoretical.RealityModel) error {
	if !u.options.StrictObservers {
		return nil
	}
	for _, observer := range model.Observers {
		if observer == nil || observer.Src == "" || builtin.GetObserver(observer.Src) == nil {
			return fmt.Errorf("%w in universe %q, reality %q", instrumentation.ErrUnknownObserver, u.model.ID, model.ID)
		}
	}
	return nil
}

func (u *ExUniverse) checkAccumulatorBudget(copies int) error {
	limit := u.options.MaxAccumulatedEvents
	if limit == 0 || u.eventAccumulator == nil {
		return nil
	}
	if copies > limit-u.eventAccumulator.GetStatistics().CountAllEvents() {
		return &instrumentation.ResourceLimitError{Resource: "accumulated events", Limit: limit}
	}
	return nil
}

func (u *ExUniverse) runConditionExecutor(ctx context.Context, args *conditionExecutorArgs) (bool, error) {
	if args.condition.Src == "" {
		return true, nil
	}

	if err := ctx.Err(); err != nil {
		return false, err
	}
	if fn := builtin.GetCondition(args.condition.Src); fn != nil {
		callbackCtx, done := u.owner.callbackContext(ctx)
		defer done()
		approved, err := fn(callbackCtx, args)
		if err != nil {
			return false, err
		}
		if err := ctx.Err(); err != nil {
			return false, err
		}
		return approved, nil
	}

	slog.WarnContext(ctx, "condition not found (default: return false)", "src", args.condition.Src)
	return false, nil
}

func (u *ExUniverse) getRealityModel(realityName string) (*theoretical.RealityModel, error) {
	realityModel, ok := u.model.GetReality(realityName)
	if !ok {
		return nil, fmt.Errorf("reality '%s' does not exist in universe '%s'", realityName, u.model.ID)
	}
	return realityModel, nil
}

func (u *ExUniverse) snapshotProvider() func() *instrumentation.MachineSnapshot {
	if u.getSnapshotFn != nil {
		return u.getSnapshotFn
	}
	if u.constantsLawsExecutor != nil {
		return u.constantsLawsExecutor.GetSnapshot
	}
	return func() *instrumentation.MachineSnapshot { return nil }
}

func (u *ExUniverse) snapshotWithErrorProvider() func() (*instrumentation.MachineSnapshot, error) {
	if u.getSnapshotWithErrorFn != nil {
		return u.getSnapshotWithErrorFn
	}
	return func() (*instrumentation.MachineSnapshot, error) {
		return instrumentation.GetSnapshotWithError(u.constantsLawsExecutor)
	}
}

func (u *ExUniverse) universeConstants() *theoretical.UniversalConstantsModel {
	if u.model == nil {
		return nil
	}
	return u.model.UniversalConstants
}

func (u *ExUniverse) executeUniverseConstantActions(
	ctx context.Context,
	actionType instrumentation.ActionType,
	event instrumentation.Event,
	emittedEvents *[]instrumentation.EmittedEvent,
) error {
	uc := u.universeConstants()
	if uc == nil {
		return nil
	}

	var actions []*theoretical.ActionModel
	switch actionType {
	case instrumentation.ActionTypeEntry:
		actions = uc.EntryActions
	case instrumentation.ActionTypeExit:
		actions = uc.ExitActions
	case instrumentation.ActionTypeTransition:
		actions = uc.ActionsOnTransition
	}

	if err := u.executeActions(ctx, actions, event, actionType, emittedEvents); err != nil {
		return errors.Join(fmt.Errorf("error executing universe constant %s actions", actionType), err)
	}
	return nil
}

func (u *ExUniverse) executeUniverseConstantInvokes(ctx context.Context, kind string, event instrumentation.Event) {
	uc := u.universeConstants()
	if uc == nil {
		return
	}

	var invokes []*theoretical.InvokeModel
	switch kind {
	case "entry":
		invokes = uc.EntryInvokes
	case "exit":
		invokes = uc.ExitInvokes
	case "transition":
		invokes = uc.InvokesOnTransition
	}

	u.executeInvokes(ctx, invokes, event)
}
