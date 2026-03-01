import { describe, expect, it } from "vitest";

import {
  applyStudioLayoutDocument,
  createInitialEditorState,
  parseAndApplyStudioLayout,
  parseStudioLayout,
  serializeStudioLayout,
} from "../model";
import type { EditorNode, StudioLayoutDocument } from "../types";

describe("studio layout JSON", () => {
  it("serializa estado visual (posiciones, notas, offsets y packs)", () => {
    const state = createInitialEditorState();
    const universe = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    const reality = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality",
    );
    if (!universe || !reality) {
      throw new Error("Missing fixtures");
    }

    universe.x = 1400;
    universe.y = 1500;
    universe.w = 740;
    universe.h = 490;
    universe.data.note = { text: "Universe visual note", colorIndex: 2 };
    reality.x = 1460;
    reality.y = 1590;
    reality.data.note = { text: "Reality visual note", colorIndex: 1 };
    state.transitions[0] = {
      ...state.transitions[0],
      visualOffset: { x: 24, y: -12 },
      note: { text: "Transition visual note", colorIndex: 3 },
    };
    state.nodes.push({
      id: "global-note-test",
      type: "note",
      x: 320,
      y: 180,
      data: {
        text: "Global note",
        colorIndex: 4,
        isCollapsed: true,
      },
    });
    state.metadataPackRegistry = [
      {
        id: "pack-layout",
        label: "Pack Layout",
        scopes: ["machine"],
        schema: {
          type: "object",
          properties: {
            profile: {
              type: "object",
            },
          },
        },
      },
    ];
    state.metadataPackBindings = {
      machine: [
        {
          id: "binding-layout",
          packId: "pack-layout",
          scope: "machine",
          entityRef: "machine",
          values: {
            profile: {
              email: "layout@example.com",
            },
          },
        },
      ],
      universe: [],
      reality: [],
      transition: [],
    };

    const layout = serializeStudioLayout(state);

    expect(layout.version).toBe(1);
    expect(layout.machineRef.id).toBe(state.machineConfig.id);
    expect(layout.nodes.universes[0]?.x).toBe(1400);
    expect(layout.nodes.universes[0]?.note?.text).toBe("Universe visual note");
    expect(layout.nodes.realities[0]?.note?.text).toBe("Reality visual note");
    expect(layout.transitions[0]?.visualOffset).toEqual({ x: 24, y: -12 });
    expect(layout.transitions[0]?.note?.text).toBe("Transition visual note");
    expect(layout.nodes.globalNotes).toHaveLength(1);
    expect(layout.packs.packRegistry).toHaveLength(1);
    expect(layout.packs.bindings.machine).toHaveLength(1);
  });

  it("parsea layout JSON y reporta errores cuando la estructura es inválida", () => {
    const invalid = parseStudioLayout('{"version":2,"machineRef":{}}');
    expect(invalid.canImport).toBe(false);
    expect(invalid.issues.some((issue) => issue.severity === "error")).toBe(true);
  });

  it("aplica layout por entityRef, reemplaza notas globales y metadata packs", () => {
    const source = createInitialEditorState();
    const sourceUniverse = source.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    const sourceReality = source.nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality",
    );
    if (!sourceUniverse || !sourceReality) {
      throw new Error("Missing fixtures");
    }

    sourceUniverse.x = 1800;
    sourceUniverse.y = 900;
    sourceUniverse.data.note = { text: "u-note", colorIndex: 1 };
    sourceReality.x = 1840;
    sourceReality.y = 990;
    sourceReality.data.note = { text: "r-note", colorIndex: 3 };
    source.transitions[0] = {
      ...source.transitions[0],
      visualOffset: { x: 60, y: -40 },
      note: { text: "t-note", colorIndex: 2 },
    };
    source.nodes.push({
      id: "global-note-source",
      type: "note",
      x: 200,
      y: 260,
      data: {
        text: "layout global",
        colorIndex: 2,
        isCollapsed: false,
      },
    });
    source.metadataPackRegistry = [
      {
        id: "pack-layout",
        label: "Pack Layout",
        scopes: ["machine"],
        schema: { type: "object", properties: { profile: { type: "object" } } },
      },
    ];
    source.metadataPackBindings = {
      machine: [
        {
          id: "binding-layout",
          packId: "pack-layout",
          scope: "machine",
          entityRef: "machine",
          values: { profile: { stage: "layout" } },
        },
      ],
      universe: [],
      reality: [],
      transition: [],
    };

    const layout = serializeStudioLayout(source);
    const target = createInitialEditorState();
    const applied = applyStudioLayoutDocument(target, layout);

    const appliedUniverse = applied.state.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    const appliedReality = applied.state.nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality",
    );
    const appliedGlobalNotes = applied.state.nodes.filter(
      (node): node is Extract<EditorNode, { type: "note" }> => node.type === "note",
    );

    expect(applied.issues).toHaveLength(0);
    expect(appliedUniverse?.x).toBe(1800);
    expect(appliedUniverse?.data.note?.text).toBe("u-note");
    expect(appliedReality?.x).toBe(1840);
    expect(appliedReality?.data.note?.text).toBe("r-note");
    expect(applied.state.transitions[0]?.visualOffset).toEqual({ x: 60, y: -40 });
    expect(applied.state.transitions[0]?.note?.text).toBe("t-note");
    expect(appliedGlobalNotes).toHaveLength(1);
    expect(applied.state.metadataPackRegistry).toHaveLength(1);
    expect(applied.state.metadataPackBindings.machine).toHaveLength(1);
  });

  it("parseAndApply ignora refs inexistentes y retorna warning no bloqueante", () => {
    const state = createInitialEditorState();
    const layout = serializeStudioLayout(state);
    const editedLayout: StudioLayoutDocument = {
      ...layout,
      nodes: {
        ...layout.nodes,
        universes: [
          ...layout.nodes.universes,
          {
            entityRef: "U:not-found",
            x: 0,
            y: 0,
            w: 300,
            h: 200,
            note: null,
          },
        ],
      },
      transitions: [
        ...layout.transitions,
        {
          entityRef: "U:not-found:R:not-found:T:on:EV:0",
          note: null,
        },
      ],
    };

    const result = parseAndApplyStudioLayout(JSON.stringify(editedLayout), state);
    expect(result.canImport).toBe(true);
    expect(result.issues.some((issue) => issue.severity === "warning")).toBe(true);
  });
});
