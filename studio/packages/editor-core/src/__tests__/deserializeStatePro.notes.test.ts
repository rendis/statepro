import { describe, expect, it } from "vitest";

import { deserializeStatePro } from "../model";
import type { EditorNode, StateProMachine } from "../types";

const createMachineWithNotes = (): StateProMachine => ({
  id: "machine-with-notes",
  canonicalName: "machine-with-notes",
  version: "1.0.0",
  metadata: {
    machineFlag: true,
    _ui_note: {
      text: "ignored-machine-note",
      colorIndex: 1,
    },
    _ui_notes: [
      {
        x: 120,
        y: 160,
        text: "Global note",
        colorIndex: 3,
        isCollapsed: true,
      },
      {
        x: 220,
        y: 260,
        text: 12,
        colorIndex: 999,
      },
      {
        x: "invalid",
        y: 300,
        text: "discarded",
        colorIndex: 2,
      },
    ],
  },
  universes: {
    main: {
      id: "main",
      canonicalName: "main",
      version: "1.0.0",
      metadata: {
        universeFlag: "keep",
        _ui_note: {
          text: "Universe note",
          colorIndex: 2,
        },
        _ui_notes: [{ x: 1, y: 2, text: "invalid-scope-note", colorIndex: 0 }],
      },
      initial: "idle",
      realities: {
        idle: {
          id: "idle",
          type: "transition",
          metadata: {
            realityFlag: "keep",
            _ui_note: {
              text: "Reality note",
              colorIndex: 1,
            },
          },
          on: {
            GO: [
              {
                targets: ["done"],
                metadata: {
                  transitionFlag: true,
                  _ui_note: {
                    text: "Transition note",
                    colorIndex: 4,
                  },
                  _ui_notes: [{ x: 1, y: 2, text: "invalid-scope-note", colorIndex: 0 }],
                },
              },
            ],
          },
        },
        done: {
          id: "done",
          type: "final",
        },
      },
    },
  },
});

describe("deserializeStatePro visual metadata isolation", () => {
  it("no migra _ui_notes a nodos de nota global", () => {
    const state = deserializeStatePro(createMachineWithNotes());
    const globalNotes = state.nodes.filter(
      (node): node is Extract<EditorNode, { type: "note" }> => node.type === "note",
    );

    expect(globalNotes).toHaveLength(0);
  });

  it("no migra _ui_note a notas ancladas de entidad", () => {
    const state = deserializeStatePro(createMachineWithNotes());

    const universe = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.type === "universe" && node.data.id === "main",
    );
    const reality = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.type === "reality" && node.data.id === "idle",
    );
    const transition = state.transitions[0];

    expect(universe?.data.note).toBeUndefined();
    expect(reality?.data.note).toBeUndefined();
    expect(transition?.note).toBeUndefined();
  });

  it("deja metadata intacta en import de modelo y limpia estado de packs", () => {
    const state = deserializeStatePro(createMachineWithNotes());

    const universe = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.type === "universe" && node.data.id === "main",
    );
    const reality = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.type === "reality" && node.data.id === "idle",
    );
    const transition = state.transitions[0];

    const machineMetadata = JSON.parse(state.machineConfig.metadata || "{}");
    const universeMetadata = JSON.parse(universe?.data.metadata || "{}");
    const realityMetadata = JSON.parse(reality?.data.metadata || "{}");
    const transitionMetadata = JSON.parse(transition?.metadata || "{}");

    expect(machineMetadata.machineFlag).toBe(true);
    expect(universeMetadata.universeFlag).toBe("keep");
    expect(realityMetadata.realityFlag).toBe("keep");
    expect(transitionMetadata.transitionFlag).toBe(true);

    expect(state.metadataPackRegistry).toEqual([]);
    expect(state.metadataPackBindings).toEqual({
      machine: [],
      universe: [],
      reality: [],
      transition: [],
    });
  });
});
