import { afterEach, describe, expect, it } from "vitest";

import {
  createInitialEditorState,
  deserializeStatePro,
  serializeStatePro,
} from "../model";
import {
  createInitialEditorHistoryState,
  editorHistoryReducer,
} from "../state/editorHistoryReducer";
import { editorReducer } from "../state/editorReducer";
import type {
  MetadataPackBinding,
  MetadataPackDefinition,
  StateProMachine,
} from "../types";
import {
  deepMergeJsonObjects,
  getValueAtPointer,
  mergePackBindingsToMetadata,
  setValueAtPointer,
  validateBindingWithPack,
} from "../utils/metadataPacks";

afterEach(() => {
  delete (Object.prototype as Record<string, unknown>).reviewMarker;
});

describe("regresiones de la revisión", () => {
  it.each(["/__proto__/reviewMarker", "/constructor/prototype/reviewMarker"])(
    "escribe %s como datos propios sin atravesar prototipos",
    (pointer) => {
      const result = setValueAtPointer({}, pointer, true);
      expect(({} as Record<string, unknown>).reviewMarker).toBeUndefined();
      expect(getValueAtPointer(result, pointer)).toBe(true);
      expect(getValueAtPointer({}, pointer)).toBeUndefined();
    },
  );

  it("fusiona claves JSON especiales sin modificar el prototipo del resultado", () => {
    const input = JSON.parse('{"__proto__":{"reviewMarker":true}}');
    const result = deepMergeJsonObjects({}, input);
    expect(Object.getPrototypeOf(result)).toBe(Object.prototype);
    expect(Object.hasOwn(result, "__proto__")).toBe(true);
    expect(({} as Record<string, unknown>).reviewMarker).toBeUndefined();
  });

  it("mantiene las constantes de packs en datos propios al fusionar metadata", () => {
    const pack: MetadataPackDefinition = {
      id: "review-pack",
      label: "Review",
      scopes: ["machine"],
      schema: { type: "object" },
      ui: {
        "/__proto__/reviewMarker": { constant: true, constantValue: true },
      },
    };
    const binding: MetadataPackBinding = {
      id: "binding",
      packId: pack.id,
      scope: "machine",
      entityRef: "machine",
      values: {},
    };
    const metadata = mergePackBindingsToMetadata(
      [binding],
      new Map([[pack.id, pack]]),
    );
    expect(({} as Record<string, unknown>).reviewMarker).toBeUndefined();
    expect(getValueAtPointer(metadata, "/__proto__/reviewMarker")).toBe(true);
    const state = createInitialEditorState();
    state.metadataPackRegistry = [pack];
    state.metadataPackBindings.machine = [binding];
    const exported = serializeStatePro(state);
    expect(exported.canExport, JSON.stringify(exported.issues)).toBe(true);
    expect(
      getValueAtPointer(exported.machine.metadata, "/__proto__/reviewMarker"),
    ).toBe(true);
    expect(({} as Record<string, unknown>).reviewMarker).toBeUndefined();
  });

  it("revalida esquemas editados y packs de diferentes instancias con el mismo id", () => {
    const pack: MetadataPackDefinition = {
      id: "review-validation",
      label: "Review",
      scopes: ["machine"],
      schema: {
        $id: "https://example.test/review-schema",
        type: "object",
        properties: { value: { type: "string" } },
      },
    };
    const binding: MetadataPackBinding = {
      id: "binding",
      packId: pack.id,
      scope: "machine",
      entityRef: "machine",
      values: { value: "text" },
    };
    expect(validateBindingWithPack(binding, pack)).toEqual([]);
    pack.schema.properties = { value: { type: "number" } };
    expect(validateBindingWithPack(binding, pack)).not.toEqual([]);
    expect(
      validateBindingWithPack({ ...binding, values: { value: 42 } }, pack),
    ).toEqual([]);
    expect(
      validateBindingWithPack(binding, {
        ...pack,
        schema: { type: "object", required: ["missing"] },
      }),
    ).not.toEqual([]);
  });

  it("no crea estados ni historial nuevos para tamaños de nodos sin cambios", () => {
    const state = createInitialEditorState();
    expect(
      editorReducer(state, {
        type: "update-node-sizes",
        payload: (previous) => previous,
      }),
    ).toBe(state);
    expect(
      editorReducer(state, {
        type: "set-node-sizes",
        payload: state.nodeSizes,
      }),
    ).toBe(state);
    const history = createInitialEditorHistoryState(state);
    expect(
      editorHistoryReducer(history, {
        type: "apply-editor-action",
        action: { type: "update-node-sizes", payload: (previous) => previous },
        mode: "silent",
        markDirtyFromImport: false,
      }),
    ).toBe(history);
  });

  it.each([false, true])(
    "preserva orden, argumentos y multiplicidad de condiciones (editado=%s)",
    (edited) => {
      const conditions = [
        { src: "condition:limit", args: { limit: 7 } },
        { src: "condition:limit", args: { limit: 9 } },
      ];
      const machine: StateProMachine = {
        id: "machine",
        canonicalName: "machine",
        version: "1.0.0",
        initials: ["U:main"],
        universes: {
          main: {
            id: "main",
            canonicalName: "main",
            version: "1.0.0",
            initial: "idle",
            realities: {
              idle: {
                id: "idle",
                type: "transition",
                on: {
                  GO: [
                    { targets: ["done"], condition: conditions[0], conditions },
                  ],
                },
              },
              done: { id: "done", type: "final" },
            },
          },
        },
      };
      const state = deserializeStatePro(machine);
      state.isDirtyFromImport = edited;
      const expected = [conditions[0], ...conditions];
      expect(state.transitions[0].conditions).toEqual(expected);
      const result = serializeStatePro(state);
      expect(result.canExport, JSON.stringify(result.issues)).toBe(true);
      expect(
        result.machine.universes.main.realities.idle.on?.GO[0].conditions,
      ).toEqual(expected);
      expect(machine.universes.main.realities.idle.on?.GO[0].condition).toEqual(
        conditions[0],
      );
    },
  );
});
