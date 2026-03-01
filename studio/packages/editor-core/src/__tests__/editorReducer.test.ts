import { describe, expect, it } from "vitest";

import { createInitialEditorState } from "../model";
import { editorReducer } from "../state";

describe("editorReducer", () => {
  it("mantiene un solo initial por universo", () => {
    const state = createInitialEditorState();

    const next = editorReducer(state, {
      type: "update-node-data",
      payload: {
        nodeId: "real-2",
        patch: { isInitial: true },
      },
    });

    const idle = next.nodes.find((node) => node.id === "real-1" && node.type === "reality");
    const processing = next.nodes.find((node) => node.id === "real-2" && node.type === "reality");

    expect(idle?.type === "reality" ? idle.data.isInitial : false).toBe(false);
    expect(processing?.type === "reality" ? processing.data.isInitial : false).toBe(true);
  });

  it("borra universo con cascade de realities y transitions", () => {
    const state = createInitialEditorState();

    const next = editorReducer(state, {
      type: "delete-element",
      payload: { kind: "node", id: "univ-1" },
    });

    expect(next.nodes.length).toBe(0);
    expect(next.transitions.length).toBe(0);
  });

  it("borra transición cuando la realidad eliminada era target", () => {
    const state = createInitialEditorState();

    const next = editorReducer(state, {
      type: "delete-element",
      payload: { kind: "node", id: "real-2" },
    });

    expect(next.nodes.some((node) => node.id === "real-2")).toBe(false);
    expect(next.transitions.length).toBe(0);
  });

  it("borra transición cuando apunta a universo eliminado", () => {
    const state = createInitialEditorState();
    const templateUniverse = state.nodes.find(
      (node): node is Extract<(typeof state.nodes)[number], { type: "universe" }> =>
        node.type === "universe",
    );
    const templateTransition = state.transitions.find((transition) => transition.id === "tr-1");
    expect(templateUniverse).toBeDefined();
    expect(templateTransition).toBeDefined();
    if (!templateUniverse || !templateTransition) {
      return;
    }

    state.nodes.push({
      ...templateUniverse,
      id: "univ-2",
      x: 1700,
      y: 1000,
      data: {
        ...templateUniverse.data,
        id: "aux-universe",
        name: "aux-universe",
        canonicalName: "aux-universe",
      },
    });
    state.transitions.push({
      ...templateTransition,
      id: "tr-cross-universe",
      sourceRealityId: "real-1",
      eventName: "GO_AUX",
      targets: ["U:aux-universe"],
      order: 1,
    });

    const next = editorReducer(state, {
      type: "delete-element",
      payload: { kind: "node", id: "univ-2" },
    });

    expect(next.transitions.some((transition) => transition.id === "tr-cross-universe")).toBe(false);
    expect(next.transitions.some((transition) => transition.id === "tr-1")).toBe(true);
  });

  it("borra transición cuando apunta a realidad externa eliminada", () => {
    const state = createInitialEditorState();
    const templateUniverse = state.nodes.find(
      (node): node is Extract<(typeof state.nodes)[number], { type: "universe" }> =>
        node.type === "universe",
    );
    const templateReality = state.nodes.find(
      (node): node is Extract<(typeof state.nodes)[number], { type: "reality" }> =>
        node.type === "reality",
    );
    const templateTransition = state.transitions.find((transition) => transition.id === "tr-1");
    expect(templateUniverse).toBeDefined();
    expect(templateReality).toBeDefined();
    expect(templateTransition).toBeDefined();
    if (!templateUniverse || !templateReality || !templateTransition) {
      return;
    }

    state.nodes.push({
      ...templateUniverse,
      id: "univ-2",
      x: 1700,
      y: 1000,
      data: {
        ...templateUniverse.data,
        id: "aux-universe",
        name: "aux-universe",
        canonicalName: "aux-universe",
      },
    });
    state.nodes.push({
      ...templateReality,
      id: "real-3",
      x: 1760,
      y: 1120,
      data: {
        ...templateReality.data,
        id: "review",
        name: "review",
        universeId: "univ-2",
      },
    });
    state.transitions.push({
      ...templateTransition,
      id: "tr-cross-reality",
      sourceRealityId: "real-1",
      eventName: "GO_REVIEW",
      targets: ["U:aux-universe:review"],
      order: 1,
    });

    const next = editorReducer(state, {
      type: "delete-element",
      payload: { kind: "node", id: "real-3" },
    });

    expect(next.transitions.some((transition) => transition.id === "tr-cross-reality")).toBe(false);
    expect(next.transitions.some((transition) => transition.id === "tr-1")).toBe(true);
  });

  it("marca dirty después de una edición sobre snapshot importado", () => {
    const state = createInitialEditorState();
    state.lastImportedMachine = {
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
              type: "final",
            },
          },
        },
      },
    };
    state.isDirtyFromImport = false;

    const next = editorReducer(state, {
      type: "mark-dirty-from-import",
    });

    expect(next.isDirtyFromImport).toBe(true);
  });

  it("compone updates funcionales consecutivos en transiciones sin revertir trigger", () => {
    const state = createInitialEditorState();

    const next = editorReducer(
      editorReducer(state, {
        type: "update-transitions",
        payload: (previous) =>
          previous.map((transition) =>
            transition.id === "tr-1"
              ? {
                  ...transition,
                  triggerKind: "always",
                  eventName: undefined,
                }
              : transition,
          ),
      }),
      {
        type: "update-transitions",
        payload: (previous) =>
          previous.map((transition) =>
            transition.id === "tr-1"
              ? {
                  ...transition,
                  description: "updated-after-trigger",
                }
              : transition,
          ),
      },
    );

    const transition = next.transitions.find((entry) => entry.id === "tr-1");
    expect(transition?.triggerKind).toBe("always");
    expect(transition?.eventName).toBeUndefined();
    expect(transition?.description).toBe("updated-after-trigger");
  });
});
