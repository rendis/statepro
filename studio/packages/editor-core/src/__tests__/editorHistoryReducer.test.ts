import { describe, expect, it } from "vitest";

import { createInitialEditorState } from "../model";
import {
  COALESCE_WINDOW_MS,
  HISTORY_LIMIT,
  createInitialEditorHistoryState,
  editorHistoryReducer,
} from "../state";
import type { EditorHistoryState } from "../state";

const applyMachineId = (
  state: EditorHistoryState,
  id: string,
  mode: "record" | "coalesce" | "silent" = "record",
  group = "machine-id",
  now = Date.now(),
): EditorHistoryState => {
  return editorHistoryReducer(state, {
    type: "apply-editor-action",
    mode,
    group,
    now,
    action: {
      type: "set-machine-config",
      payload: {
        ...state.present.machineConfig,
        id,
      },
    },
  });
};

describe("editorHistoryReducer", () => {
  it("permite cambios visuales sin marcar dirty-from-import cuando se indica explícitamente", () => {
    const imported = createInitialEditorState();
    imported.lastImportedMachine = {
      id: "m-1",
      canonicalName: "m-1",
      version: "1.0.0",
      initials: ["U:main"],
      universes: {
        main: {
          id: "main",
          canonicalName: "main",
          version: "1.0.0",
          realities: {
            idle: {
              id: "idle",
              type: "transition",
              always: [{ targets: ["idle"] }],
            },
          },
        },
      },
    };
    imported.isDirtyFromImport = false;

    const initial = createInitialEditorHistoryState(imported);
    const visualNext = editorHistoryReducer(initial, {
      type: "apply-editor-action",
      mode: "silent",
      markDirtyFromImport: false,
      action: {
        type: "set-nodes",
        payload: imported.nodes.map((node) =>
          node.type === "universe" ? { ...node, x: node.x + 10 } : node,
        ),
      },
    });

    expect(visualNext.present.isDirtyFromImport).toBe(false);

    const domainNext = editorHistoryReducer(visualNext, {
      type: "apply-editor-action",
      mode: "record",
      action: {
        type: "set-machine-config",
        payload: {
          ...visualNext.present.machineConfig,
          id: "changed-id",
        },
      },
    });

    expect(domainNext.present.isDirtyFromImport).toBe(true);
  });

  it("record apila en past y limpia future", () => {
    const initial = createInitialEditorHistoryState(createInitialEditorState());
    const first = applyMachineId(initial, "machine-a", "record");
    const undone = editorHistoryReducer(first, { type: "undo" });

    expect(undone.future.length).toBe(1);

    const branch = applyMachineId(undone, "machine-b", "record");
    expect(branch.past.length).toBe(1);
    expect(branch.future.length).toBe(0);
    expect(branch.present.machineConfig.id).toBe("machine-b");
  });

  it("undo/redo mueve snapshots correctamente", () => {
    const initial = createInitialEditorHistoryState(createInitialEditorState());
    const a = applyMachineId(initial, "machine-a", "record");
    const b = applyMachineId(a, "machine-b", "record");

    const undo1 = editorHistoryReducer(b, { type: "undo" });
    expect(undo1.present.machineConfig.id).toBe("machine-a");
    expect(undo1.past.length).toBe(1);
    expect(undo1.future.length).toBe(1);

    const undo2 = editorHistoryReducer(undo1, { type: "undo" });
    expect(undo2.present.machineConfig.id).toBe(initial.present.machineConfig.id);
    expect(undo2.past.length).toBe(0);
    expect(undo2.future.length).toBe(2);

    const redo1 = editorHistoryReducer(undo2, { type: "redo" });
    expect(redo1.present.machineConfig.id).toBe("machine-a");
    expect(redo1.past.length).toBe(1);
    expect(redo1.future.length).toBe(1);
  });

  it("coalesce agrupa por group dentro de la ventana temporal", () => {
    const initial = createInitialEditorHistoryState(createInitialEditorState());
    const t0 = 1_000;
    const a = applyMachineId(initial, "machine-a", "coalesce", "machine-id", t0);
    const b = applyMachineId(
      a,
      "machine-b",
      "coalesce",
      "machine-id",
      t0 + COALESCE_WINDOW_MS - 50,
    );
    const c = applyMachineId(
      b,
      "machine-c",
      "coalesce",
      "machine-id",
      t0 + 2 * COALESCE_WINDOW_MS + 50,
    );

    expect(a.past.length).toBe(1);
    expect(b.past.length).toBe(1);
    expect(c.past.length).toBe(2);
  });

  it("limita past con HISTORY_LIMIT", () => {
    let state = createInitialEditorHistoryState(createInitialEditorState());

    for (let index = 0; index < HISTORY_LIMIT + 20; index += 1) {
      state = applyMachineId(state, `machine-${index}`, "record");
    }

    expect(state.past.length).toBe(HISTORY_LIMIT);
  });

  it("silent no modifica historial", () => {
    const initial = createInitialEditorHistoryState(createInitialEditorState());
    const next = applyMachineId(initial, "silent-machine", "silent");

    expect(next.past.length).toBe(0);
    expect(next.future.length).toBe(0);
    expect(next.present.machineConfig.id).toBe("silent-machine");
  });

  it("reset-history limpia past/future y reemplaza present", () => {
    const initial = createInitialEditorHistoryState(createInitialEditorState());
    const edited = applyMachineId(initial, "machine-a", "record");
    const resetState = createInitialEditorState();
    resetState.machineConfig.id = "reset-machine";

    const reset = editorHistoryReducer(edited, {
      type: "reset-history",
      payload: resetState,
    });

    expect(reset.past.length).toBe(0);
    expect(reset.future.length).toBe(0);
    expect(reset.present.machineConfig.id).toBe("reset-machine");
  });
});
