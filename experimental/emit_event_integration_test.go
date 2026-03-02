package experimental

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/rendis/statepro/v3/instrumentation"
	"github.com/rendis/statepro/v3/theoretical"
)

// ---------- helpers ----------

// loadTestModel loads and deserializes a JSON state machine definition from testdata/.
func loadTestModel(t *testing.T, filename string) *theoretical.QuantumMachineModel {
	t.Helper()
	b, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}
	var model theoretical.QuantumMachineModel
	if err := json.Unmarshal(b, &model); err != nil {
		t.Fatalf("failed to deserialize %s: %v", filename, err)
	}
	return &model
}

// buildFromModel creates and initializes a quantum machine from a model, allowing the caller
// to override Initials to select which universes to start.
func buildFromModel(t *testing.T, model *theoretical.QuantumMachineModel, initials []string) (*ExQuantumMachine, map[string]*ExUniverse) {
	t.Helper()
	if initials != nil {
		model.Initials = initials
	}
	universes := make([]*ExUniverse, 0, len(model.Universes))
	universeMap := make(map[string]*ExUniverse, len(model.Universes))
	for _, um := range model.Universes {
		u := NewExUniverse(um)
		universes = append(universes, u)
		universeMap[um.ID] = u
	}
	qm, err := NewExQuantumMachine(model, universes)
	if err != nil {
		t.Fatalf("NewExQuantumMachine failed: %v", err)
	}
	return qm.(*ExQuantumMachine), universeMap
}

// initMachine initializes the machine with a nil context.
func initMachine(t *testing.T, qm *ExQuantumMachine) {
	t.Helper()
	if err := qm.Init(context.Background(), nil); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
}

// sendTestEvent sends a named event and returns the snapshot.
func sendTestEvent(t *testing.T, qm *ExQuantumMachine, eventName string) *instrumentation.MachineSnapshot {
	t.Helper()
	event := NewEventBuilder(eventName).Build()
	_, err := qm.SendEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("SendEvent(%s) failed: %v", eventName, err)
	}
	return qm.GetSnapshot()
}

// sendTestEventWithData sends a named event with data.
func sendTestEventWithData(t *testing.T, qm *ExQuantumMachine, eventName string, data map[string]any) *instrumentation.MachineSnapshot {
	t.Helper()
	event := NewEventBuilder(eventName).SetData(data).Build()
	_, err := qm.SendEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("SendEvent(%s) failed: %v", eventName, err)
	}
	return qm.GetSnapshot()
}

// assertUniverseReality checks the current reality of a specific universe.
func assertUniverseReality(t *testing.T, universes map[string]*ExUniverse, universeID, expectedReality string) {
	t.Helper()
	u, ok := universes[universeID]
	if !ok {
		t.Fatalf("universe %q not found", universeID)
	}
	if u.currentReality == nil {
		if expectedReality == "" {
			return // expected nil
		}
		t.Fatalf("universe %q: expected reality %q, got nil (superposition or uninitialized)", universeID, expectedReality)
	}
	if *u.currentReality != expectedReality {
		t.Fatalf("universe %q: expected reality %q, got %q", universeID, expectedReality, *u.currentReality)
	}
}

// assertUniverseFinalized checks that a universe is finalized.
func assertUniverseFinalized(t *testing.T, universes map[string]*ExUniverse, universeID string) {
	t.Helper()
	u, ok := universes[universeID]
	if !ok {
		t.Fatalf("universe %q not found", universeID)
	}
	if !u.isFinalReality {
		t.Fatalf("universe %q: expected isFinalReality=true, got false", universeID)
	}
}

// ---------- Tests ----------

// 1. Happy path: Init → EmitEvent auto-advances through the full flow.
// CREATING_FORM → emit "create-form" → FILLING_FORM → SendEvent("fill-form") → FILLED →
// always → contract-process:CREATING_CONTRACT → emit "create-contract" → SIGNING_CONTRACT →
// SendEvent("sign") → SIGNED → always → completed:DONE
func TestEmitEvent_JSONLoad_HappyPath(t *testing.T) {
	registerTestAction(t, "action:createFormTest", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		templateId := args.GetAction().Args["templateId"].(string)
		args.EmitEvent("create-form", map[string]any{"templateId": templateId})
		return nil
	})
	registerTestAction(t, "action:createContractTest", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		templateId := args.GetAction().Args["templateId"].(string)
		args.EmitEvent("create-contract", map[string]any{"templateId": templateId})
		return nil
	})
	registerTestCondition(t, "condition:ifTemplateTest", func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		expected := args.GetCondition().Args["templateId"].(string)
		actual, ok := args.GetEvent().GetData()["templateId"].(string)
		return ok && actual == expected, nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	qm, universes := buildFromModel(t, model, []string{"U:form-process"})
	initMachine(t, qm)

	// After Init: CREATING_FORM entry action emitted "create-form" → condition passed → FILLING_FORM
	assertUniverseReality(t, universes, "form-process", "FILLING_FORM")

	// SendEvent fill-form → FILLED → always → contract-process:CREATING_CONTRACT → emit → SIGNING_CONTRACT
	sendTestEvent(t, qm, "fill-form")
	assertUniverseFinalized(t, universes, "form-process")
	assertUniverseReality(t, universes, "contract-process", "SIGNING_CONTRACT")

	// SendEvent sign → SIGNED → always → completed:DONE
	sendTestEvent(t, qm, "sign")
	assertUniverseFinalized(t, universes, "contract-process")
	assertUniverseReality(t, universes, "completed", "DONE")
	assertUniverseFinalized(t, universes, "completed")
}

// 2. Condition blocks: entry action emits event but condition rejects it → stays in CREATING_FORM.
func TestEmitEvent_JSONLoad_ConditionBlocks(t *testing.T) {
	registerTestAction(t, "action:createFormTest-cond", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		// Emit with WRONG templateId
		args.EmitEvent("create-form", map[string]any{"templateId": "wrong-tpl"})
		return nil
	})
	registerTestCondition(t, "condition:ifTemplateTest-cond", func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		expected := args.GetCondition().Args["templateId"].(string)
		actual, ok := args.GetEvent().GetData()["templateId"].(string)
		return ok && actual == expected, nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	// Override action/condition names in the loaded model
	formUniverse := model.Universes["form-process"]
	formUniverse.Realities["CREATING_FORM"].EntryActions[0].Src = "action:createFormTest-cond"
	formUniverse.Realities["CREATING_FORM"].On["create-form"][0].Condition.Src = "condition:ifTemplateTest-cond"

	qm, universes := buildFromModel(t, model, []string{"U:form-process"})
	initMachine(t, qm)

	// Condition fails → stays in CREATING_FORM
	assertUniverseReality(t, universes, "form-process", "CREATING_FORM")
}

// 3. No emit: action does NOT call EmitEvent → machine stays waiting (backward compat).
func TestEmitEvent_JSONLoad_NoEmit(t *testing.T) {
	registerTestAction(t, "action:createFormTest-noemit", func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		// Intentionally does NOT call EmitEvent
		return nil
	})
	registerTestCondition(t, "condition:ifTemplateTest-noemit", func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		return true, nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	formUniverse := model.Universes["form-process"]
	formUniverse.Realities["CREATING_FORM"].EntryActions[0].Src = "action:createFormTest-noemit"
	formUniverse.Realities["CREATING_FORM"].On["create-form"][0].Condition.Src = "condition:ifTemplateTest-noemit"

	qm, universes := buildFromModel(t, model, []string{"U:form-process"})
	initMachine(t, qm)

	// No emit → stays in CREATING_FORM
	assertUniverseReality(t, universes, "form-process", "CREATING_FORM")
}

// 4. External event after stuck: action doesn't emit, then external SendEvent advances.
func TestEmitEvent_JSONLoad_ExternalEventAfterStuck(t *testing.T) {
	registerTestAction(t, "action:createFormTest-ext", func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return nil // no emit
	})
	registerTestCondition(t, "condition:ifTemplateTest-ext", func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		expected := args.GetCondition().Args["templateId"].(string)
		actual, ok := args.GetEvent().GetData()["templateId"].(string)
		return ok && actual == expected, nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	formUniverse := model.Universes["form-process"]
	formUniverse.Realities["CREATING_FORM"].EntryActions[0].Src = "action:createFormTest-ext"
	formUniverse.Realities["CREATING_FORM"].On["create-form"][0].Condition.Src = "condition:ifTemplateTest-ext"

	qm, universes := buildFromModel(t, model, []string{"U:form-process"})
	initMachine(t, qm)

	// Stuck in CREATING_FORM
	assertUniverseReality(t, universes, "form-process", "CREATING_FORM")

	// External SendEvent with correct data → advances
	sendTestEventWithData(t, qm, "create-form", map[string]any{"templateId": "tpl-1"})
	assertUniverseReality(t, universes, "form-process", "FILLING_FORM")
}

// 5. Multiple different events emitted (FIFO): first emit matches → transitions, second is discarded.
func TestEmitEvent_JSONLoad_MultipleDifferentEmits_FirstWins(t *testing.T) {
	registerTestAction(t, "action:firstEmitAction-fw", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("event-alpha", nil)
		return nil
	})
	registerTestAction(t, "action:secondEmitAction-fw", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("event-beta", nil)
		return nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	multiU := model.Universes["multi-emit"]
	multiU.Realities["PROCESSING"].EntryActions[0].Src = "action:firstEmitAction-fw"
	multiU.Realities["PROCESSING"].EntryActions[1].Src = "action:secondEmitAction-fw"

	qm, universes := buildFromModel(t, model, []string{"U:multi-emit"})
	initMachine(t, qm)

	// First emit "event-alpha" wins → ALPHA_DONE, "event-beta" discarded
	assertUniverseReality(t, universes, "multi-emit", "ALPHA_DONE")
}

// 6. Multiple same events: both actions emit "event-alpha" → only first processes.
func TestEmitEvent_JSONLoad_MultipleSameEmits(t *testing.T) {
	callCount := 0
	registerTestAction(t, "action:firstEmitAction-same", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		callCount++
		args.EmitEvent("event-alpha", nil)
		return nil
	})
	registerTestAction(t, "action:secondEmitAction-same", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		callCount++
		args.EmitEvent("event-alpha", nil)
		return nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	multiU := model.Universes["multi-emit"]
	multiU.Realities["PROCESSING"].EntryActions[0].Src = "action:firstEmitAction-same"
	multiU.Realities["PROCESSING"].EntryActions[1].Src = "action:secondEmitAction-same"

	qm, universes := buildFromModel(t, model, []string{"U:multi-emit"})
	initMachine(t, qm)

	// Both emitted "event-alpha", first one transitions, second is discarded
	assertUniverseReality(t, universes, "multi-emit", "ALPHA_DONE")
	if callCount != 2 {
		t.Fatalf("expected both entry actions to run, got callCount=%d", callCount)
	}
}

// 7. Fallback to second event: first emit has no handler, second has handler → second wins.
func TestEmitEvent_JSONLoad_FallbackToSecondEmit(t *testing.T) {
	registerTestAction(t, "action:firstEmitAction-fb", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("nonexistent-event", nil) // no On handler for this
		return nil
	})
	registerTestAction(t, "action:secondEmitAction-fb", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("event-beta", nil) // this one has a handler
		return nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	multiU := model.Universes["multi-emit"]
	multiU.Realities["PROCESSING"].EntryActions[0].Src = "action:firstEmitAction-fb"
	multiU.Realities["PROCESSING"].EntryActions[1].Src = "action:secondEmitAction-fb"

	qm, universes := buildFromModel(t, model, []string{"U:multi-emit"})
	initMachine(t, qm)

	// First emit "nonexistent-event" → no handler → skip. Second emit "event-beta" → BETA_DONE
	assertUniverseReality(t, universes, "multi-emit", "BETA_DONE")
}

// 8. First emit condition fails, fallback to second emit.
func TestEmitEvent_JSONLoad_FirstConditionFails_FallbackToSecond(t *testing.T) {
	registerTestAction(t, "action:conditionalEmitAction-cf", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("guarded-event", nil)  // has condition that always rejects
		args.EmitEvent("fallback-event", nil) // no condition, should succeed
		return nil
	})
	registerTestCondition(t, "condition:alwaysRejectTest-cf", func(_ context.Context, _ instrumentation.ConditionExecutorArgs) (bool, error) {
		return false, nil // always reject
	})

	model := loadTestModel(t, "emit_event_machine.json")
	condU := model.Universes["conditional-emit"]
	condU.Realities["CHECKING"].EntryActions[0].Src = "action:conditionalEmitAction-cf"
	condU.Realities["CHECKING"].On["guarded-event"][0].Condition.Src = "condition:alwaysRejectTest-cf"

	qm, universes := buildFromModel(t, model, []string{"U:conditional-emit"})
	initMachine(t, qm)

	// "guarded-event" rejected by condition → "fallback-event" succeeds → FALLBACK_PATH
	assertUniverseReality(t, universes, "conditional-emit", "FALLBACK_PATH")
}

// 9. Emit event with no matching On handler → stays in current reality.
func TestEmitEvent_JSONLoad_NoHandler(t *testing.T) {
	registerTestAction(t, "action:emitNonExistent-nh", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("ghost-event", nil) // no handler exists
		return nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	noHandlerU := model.Universes["no-handler-emit"]
	noHandlerU.Realities["WAITING"].EntryActions[0].Src = "action:emitNonExistent-nh"

	qm, universes := buildFromModel(t, model, []string{"U:no-handler-emit"})
	initMachine(t, qm)

	// "ghost-event" has no handler → stays in WAITING
	assertUniverseReality(t, universes, "no-handler-emit", "WAITING")

	// But a real external event can still advance
	sendTestEvent(t, qm, "real-event")
	assertUniverseReality(t, universes, "no-handler-emit", "DONE")
}

// 10. Chained emits: STEP_A emits → STEP_B → STEP_B emits → STEP_C (all in one Init).
func TestEmitEvent_JSONLoad_ChainedEmits(t *testing.T) {
	registerTestAction(t, "action:emitAdvanceA-ch", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance-a", nil)
		return nil
	})
	registerTestAction(t, "action:emitAdvanceB-ch", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance-b", nil)
		return nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	chainedU := model.Universes["chained-emit"]
	chainedU.Realities["STEP_A"].EntryActions[0].Src = "action:emitAdvanceA-ch"
	chainedU.Realities["STEP_B"].EntryActions[0].Src = "action:emitAdvanceB-ch"

	qm, universes := buildFromModel(t, model, []string{"U:chained-emit"})
	initMachine(t, qm)

	// STEP_A emit → STEP_B → STEP_B emit → STEP_C (final)
	assertUniverseReality(t, universes, "chained-emit", "STEP_C")
	assertUniverseFinalized(t, universes, "chained-emit")
}

// 11. Cross-universe emit: entry action emits → transitions to external universe.
func TestEmitEvent_JSONLoad_CrossUniverseEmit(t *testing.T) {
	registerTestAction(t, "action:emitCrossUniverse-cu", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("go-external", nil)
		return nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	crossU := model.Universes["cross-universe-emit"]
	crossU.Realities["DISPATCHING"].EntryActions[0].Src = "action:emitCrossUniverse-cu"

	qm, universes := buildFromModel(t, model, []string{"U:cross-universe-emit"})
	initMachine(t, qm)

	// Emit "go-external" → target is U:completed → superposition → completed:DONE
	assertUniverseReality(t, universes, "completed", "DONE")
	assertUniverseFinalized(t, universes, "completed")
}

// 12. Snapshot verification: ensure snapshot reflects post-EmitEvent state correctly.
func TestEmitEvent_JSONLoad_SnapshotAfterEmit(t *testing.T) {
	registerTestAction(t, "action:createFormTest-snap", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		templateId := args.GetAction().Args["templateId"].(string)
		args.EmitEvent("create-form", map[string]any{"templateId": templateId})
		return nil
	})
	registerTestCondition(t, "condition:ifTemplateTest-snap", func(_ context.Context, args instrumentation.ConditionExecutorArgs) (bool, error) {
		expected := args.GetCondition().Args["templateId"].(string)
		actual, ok := args.GetEvent().GetData()["templateId"].(string)
		return ok && actual == expected, nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	formUniverse := model.Universes["form-process"]
	formUniverse.Realities["CREATING_FORM"].EntryActions[0].Src = "action:createFormTest-snap"
	formUniverse.Realities["CREATING_FORM"].On["create-form"][0].Condition.Src = "condition:ifTemplateTest-snap"

	qm, _ := buildFromModel(t, model, []string{"U:form-process"})
	initMachine(t, qm)

	snapshot := qm.GetSnapshot()
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}

	// Verify the snapshot shows form-process in FILLING_FORM (post-emit state).
	// ActiveUniverses is keyed by canonicalName.
	reality, ok := snapshot.Resume.ActiveUniverses["form-process"]
	if !ok {
		t.Fatal("snapshot: form-process not found in ActiveUniverses")
	}
	if reality != "FILLING_FORM" {
		t.Fatalf("snapshot: expected form-process in FILLING_FORM, got %q", reality)
	}
}

// 13. Multiple emits where none have handlers → stays, then external event works.
func TestEmitEvent_JSONLoad_AllEmitsMiss(t *testing.T) {
	registerTestAction(t, "action:firstEmitAction-miss", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("totally-unknown-1", nil)
		return nil
	})
	registerTestAction(t, "action:secondEmitAction-miss", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("totally-unknown-2", nil)
		return nil
	})

	model := loadTestModel(t, "emit_event_machine.json")
	multiU := model.Universes["multi-emit"]
	multiU.Realities["PROCESSING"].EntryActions[0].Src = "action:firstEmitAction-miss"
	multiU.Realities["PROCESSING"].EntryActions[1].Src = "action:secondEmitAction-miss"

	qm, universes := buildFromModel(t, model, []string{"U:multi-emit"})
	initMachine(t, qm)

	// Both emitted events have no handlers → stays in PROCESSING
	assertUniverseReality(t, universes, "multi-emit", "PROCESSING")

	// External event still works
	sendTestEvent(t, qm, "event-alpha")
	assertUniverseReality(t, universes, "multi-emit", "ALPHA_DONE")
}

// 14. Emit with nil data → works fine (data is optional).
func TestEmitEvent_JSONLoad_EmitNilData(t *testing.T) {
	registerTestAction(t, "action:emitAdvanceA-nil", func(_ context.Context, args instrumentation.ActionExecutorArgs) error {
		args.EmitEvent("advance-a", nil) // explicitly nil data
		return nil
	})
	registerTestAction(t, "action:emitAdvanceB-nil", func(_ context.Context, _ instrumentation.ActionExecutorArgs) error {
		return nil // no emit for B, stops at STEP_B
	})

	model := loadTestModel(t, "emit_event_machine.json")
	chainedU := model.Universes["chained-emit"]
	chainedU.Realities["STEP_A"].EntryActions[0].Src = "action:emitAdvanceA-nil"
	chainedU.Realities["STEP_B"].EntryActions[0].Src = "action:emitAdvanceB-nil"

	qm, universes := buildFromModel(t, model, []string{"U:chained-emit"})
	initMachine(t, qm)

	// Nil data works — STEP_A emit → STEP_B, STEP_B no emit → stays
	assertUniverseReality(t, universes, "chained-emit", "STEP_B")
}
