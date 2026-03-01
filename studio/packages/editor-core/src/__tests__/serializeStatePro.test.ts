import { describe, expect, it } from "vitest";

import { createInitialEditorState, serializeStatePro } from "../model";
import type { EditorNode, EditorTransition, StateProMachine } from "../types";

const createUniverseNode = (
  nodeId: string,
  universeId: string,
  canonicalName: string,
): Extract<EditorNode, { type: "universe" }> => ({
  id: nodeId,
  type: "universe",
  x: 1500,
  y: 1500,
  w: 300,
  h: 300,
  data: {
    id: universeId,
    name: universeId,
    canonicalName,
    version: "1.0.0",
    metadata: "{}",
    universalConstants: {
      entryActions: [],
      exitActions: [],
      entryInvokes: [],
      exitInvokes: [],
      actionsOnTransition: [],
      invokesOnTransition: [],
    },
  },
});

const createInitialRealityNode = (
  nodeId: string,
  realityId: string,
  universeNodeId: string,
): Extract<EditorNode, { type: "reality" }> => ({
  id: nodeId,
  type: "reality",
  x: 1600,
  y: 1600,
  data: {
    id: realityId,
    name: realityId,
    universeId: universeNodeId,
    isInitial: true,
    realityType: "normal",
  },
});

describe("serializeStatePro", () => {
  it("usa universe.data.id como key en universes", () => {
    const state = createInitialEditorState();

    const universe = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    if (!universe) {
      throw new Error("Missing universe fixture");
    }

    universe.id = "u-node-xyz";
    universe.data.id = "human-universe";

    const reality = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality",
    );
    if (!reality) {
      throw new Error("Missing reality fixture");
    }

    reality.data.universeId = "u-node-xyz";

    const { machine } = serializeStatePro(state);

    expect(machine.universes["human-universe"]).toBeDefined();
    expect(machine.universes["human-universe"]?.realities.idle).toBeDefined();
  });

  it("bloquea export cuando hay universe.data.id duplicado", () => {
    const state = createInitialEditorState();
    state.nodes.push(
      createUniverseNode("u-2-node", "main-universe", "main-universe-v2"),
      createInitialRealityNode("real-3", "idle-v2", "u-2-node"),
    );

    const result = serializeStatePro(state);

    expect(result.canExport).toBe(false);
    expect(
      result.issues.some(
        (issue) =>
          issue.messageKey === "issue.duplicatedUniverseId" &&
          issue.messageParams?.universeId === "main-universe",
      ),
    ).toBe(true);
  });

  it("bloquea export cuando hay canonicalName de universo duplicado", () => {
    const state = createInitialEditorState();
    state.nodes.push(
      createUniverseNode("u-2-node", "payments", "main-universe"),
      createInitialRealityNode("real-3", "pending", "u-2-node"),
    );

    const result = serializeStatePro(state);

    expect(result.canExport).toBe(false);
    expect(
      result.issues.some(
        (issue) =>
          issue.messageKey === "issue.duplicatedUniverseCanonicalName" &&
          issue.messageParams?.canonicalName === "main-universe",
      ),
    ).toBe(true);
  });

  it("preserva notify interno y lo marca como error semántico", () => {
    const state = createInitialEditorState();

    state.transitions = [
      {
        id: "tr-local-notify",
        sourceRealityId: "real-1",
        triggerKind: "on",
        eventName: "NEXT",
        type: "notify",
        condition: undefined,
        conditions: [],
        actions: [],
        invokes: [],
        description: "",
        metadata: "",
        targets: ["processing"],
        order: 0,
      },
    ];

    const result = serializeStatePro(state);
    const transition = result.machine.universes["main-universe"]?.realities.idle.on?.NEXT[0];

    expect(transition).toBeDefined();
    expect(transition?.type).toBe("notify");
    expect(result.canExport).toBe(false);
    expect(result.issues.some((issue) => issue.code === "SEMANTIC_ERROR")).toBe(true);
  });

  it("serializa transición multiplexada en un solo objeto con múltiples targets", () => {
    const state = createInitialEditorState();

    const secondUniverse: Extract<EditorNode, { type: "universe" }> = {
      id: "u-2-node",
      type: "universe",
      x: 1500,
      y: 1500,
      w: 300,
      h: 300,
      data: {
        id: "payments",
        name: "payments",
        canonicalName: "payments",
        version: "1.0.0",
        metadata: "{}",
        universalConstants: {
          entryActions: [],
          exitActions: [],
          entryInvokes: [],
          exitInvokes: [],
          actionsOnTransition: [],
          invokesOnTransition: [],
        },
      },
    };

    const secondReality: Extract<EditorNode, { type: "reality" }> = {
      id: "real-pending",
      type: "reality",
      x: 1600,
      y: 1600,
      data: {
        id: "pending",
        name: "pending",
        universeId: "u-2-node",
        isInitial: true,
        realityType: "success",
      },
    };

    state.nodes.push(secondUniverse, secondReality);

    const multiplexTransition: EditorTransition = {
      id: "tr-multiplex",
      sourceRealityId: "real-1",
      triggerKind: "on",
      eventName: "GO",
      type: "default",
      condition: undefined,
      conditions: [],
      actions: [],
      invokes: [],
      description: "",
      metadata: "",
      targets: ["processing", "U:payments:pending"],
      order: 0,
    };

    state.transitions = [multiplexTransition];

    const result = serializeStatePro(state);
    const transition = result.machine.universes["main-universe"]?.realities.idle.on?.GO[0];

    expect(transition?.targets).toEqual(["processing", "U:payments:pending"]);
    expect(result.canExport).toBe(true);
  });

  it("migra condition legado a conditions al serializar", () => {
    const state = createInitialEditorState();

    state.transitions = [
      {
        id: "tr-guards",
        sourceRealityId: "real-1",
        triggerKind: "on",
        eventName: "GO",
        type: "default",
        condition: { src: "condition:primary" },
        conditions: [{ src: "condition:secondary-a" }, { src: "condition:secondary-b" }],
        actions: [],
        invokes: [],
        description: "",
        metadata: "",
        targets: ["processing"],
        order: 0,
      },
    ];

    const result = serializeStatePro(state);
    const transition = result.machine.universes["main-universe"]?.realities.idle.on?.GO[0];

    expect(transition).not.toHaveProperty("condition");
    expect(transition?.conditions?.map((condition) => condition.src)).toEqual([
      "condition:primary",
      "condition:secondary-a",
      "condition:secondary-b",
    ]);
    expect(result.canExport).toBe(true);
  });

  it("ignora visualOffset al serializar una transición", () => {
    const state = createInitialEditorState();

    state.transitions = [
      {
        id: "tr-visual",
        sourceRealityId: "real-1",
        triggerKind: "on",
        eventName: "MOVE_ME",
        type: "default",
        condition: undefined,
        conditions: [],
        actions: [],
        invokes: [],
        description: "",
        metadata: "",
        targets: ["processing"],
        order: 0,
        visualOffset: {
          x: 80,
          y: -40,
        },
      },
    ];

    const result = serializeStatePro(state);
    const transition = result.machine.universes["main-universe"]?.realities.idle.on?.MOVE_ME[0];

    expect(transition).toBeDefined();
    expect(transition).not.toHaveProperty("visualOffset");
  });

  it("no exporta notas visuales (_ui_*) en metadata", () => {
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

    universe.data.metadata = '{\"origin\":\"manual\",\"_ui_note\":{\"text\":\"stale\",\"colorIndex\":4}}';
    universe.data.note = { text: "Universe note", colorIndex: 2 };
    reality.data.metadata = '{\"hint\":\"manual reality\",\"_ui_note\":{\"text\":\"stale\"}}';
    reality.data.note = { text: "Reality note", colorIndex: 1 };

    state.transitions = [
      {
        ...state.transitions[0],
        metadata: '{\"debug\":true,\"_ui_note\":{\"text\":\"stale transition\",\"colorIndex\":4}}',
        note: { text: "Transition note", colorIndex: 3 },
      },
    ];

    state.machineConfig.metadata =
      '{\"author\":\"rendis\",\"_ui_notes\":[{\"x\":1,\"y\":2,\"text\":\"stale\",\"colorIndex\":4}]}';
    state.nodes.push({
      id: "note-1",
      type: "note",
      x: 320,
      y: 180,
      data: {
        text: "Global note",
        colorIndex: 4,
        isCollapsed: true,
      },
    });

    const result = serializeStatePro(state);
    const machineMetadata = result.machine.metadata as Record<string, unknown>;

    expect(machineMetadata.author).toBe("rendis");
    expect(machineMetadata._ui_notes).toBeUndefined();

    expect(
      (result.machine.universes["main-universe"]?.metadata as Record<string, unknown>)._ui_note,
    ).toBeUndefined();
    expect(
      (
        result.machine.universes["main-universe"]?.realities.idle.metadata as Record<string, unknown>
      )._ui_note,
    ).toBeUndefined();
    expect(
      (
        result.machine.universes["main-universe"]?.realities.idle.on?.START_PROCESS?.[0]
          ?.metadata as Record<string, unknown>
      )._ui_note,
    ).toBeUndefined();
  });

  it("limpia keys _ui_* heredadas cuando no hay notas en estado", () => {
    const state = createInitialEditorState();

    const universe = state.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    if (!universe) {
      throw new Error("Missing universe fixture");
    }

    universe.data.metadata = '{\"custom\":true,\"_ui_note\":{\"text\":\"legacy\",\"colorIndex\":2}}';
    state.machineConfig.metadata = '{\"name\":\"machine\",\"_ui_notes\":[{\"x\":1,\"y\":1,\"text\":\"legacy\"}]}';
    state.transitions = [
      {
        ...state.transitions[0],
        metadata: '{\"custom\":true,\"_ui_note\":{\"text\":\"legacy transition\"}}',
      },
    ];

    const result = serializeStatePro(state);
    const machineMetadata = result.machine.metadata as Record<string, unknown>;
    const universeMetadata = result.machine.universes["main-universe"]?.metadata as Record<
      string,
      unknown
    >;
    const transitionMetadata = result.machine.universes["main-universe"]?.realities.idle.on
      ?.START_PROCESS?.[0]?.metadata as Record<string, unknown>;

    expect(machineMetadata._ui_notes).toBeUndefined();
    expect(universeMetadata._ui_note).toBeUndefined();
    expect(transitionMetadata._ui_note).toBeUndefined();
  });

  it("sanitiza description en BehaviorRef y migra condition legado en snapshot importado sin mutar el origen", () => {
    const state = createInitialEditorState();

    const importedMachine: StateProMachine = {
      id: "imported-machine",
      canonicalName: "imported-machine",
      version: "1.0.0",
      initials: ["U:main"],
      universalConstants: {
        entryActions: [
          {
            src: "builtin:action:machineEntry",
            description: "Machine UC action",
          },
        ],
        invokesOnTransition: [
          {
            src: "builtin:invoke:machineTransition",
            description: "Machine UC invoke",
          },
        ],
      },
      universes: {
        main: {
          id: "main",
          canonicalName: "main",
          version: "1.0.0",
          initial: "idle",
          universalConstants: {
            exitActions: [
              {
                src: "builtin:action:universeExit",
                description: "Universe UC action",
              },
            ],
          },
          realities: {
            idle: {
              id: "idle",
              type: "transition",
              entryActions: [
                {
                  src: "builtin:action:idleEntry",
                  description: "Idle entry action",
                },
              ],
              exitInvokes: [
                {
                  src: "builtin:invoke:idleExit",
                  description: "Idle exit invoke",
                },
              ],
              always: [
                {
                  targets: ["done"],
                  description: "Always transition description",
                  condition: {
                    src: "condition:ready",
                    description: "Always condition",
                  },
                  actions: [
                    {
                      src: "builtin:action:alwaysAction",
                      description: "Always action",
                    },
                  ],
                  invokes: [
                    {
                      src: "builtin:invoke:alwaysInvoke",
                      description: "Always invoke",
                    },
                  ],
                },
              ],
              on: {
                NEXT: [
                  {
                    targets: ["done"],
                    condition: {
                      src: "condition:next",
                      description: "On condition",
                    },
                    actions: [
                      {
                        src: "builtin:action:onAction",
                        description: "On action",
                      },
                    ],
                    invokes: [
                      {
                        src: "builtin:invoke:onInvoke",
                        description: "On invoke",
                      },
                    ],
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
    };

    state.lastImportedMachine = importedMachine;
    state.isDirtyFromImport = false;
    state.machineConfig.id = "modified-id-that-must-not-be-exported";

    const result = serializeStatePro(state);
    const exportedMachine = result.machine;

    expect(exportedMachine.id).toBe("imported-machine");
    expect(result.canExport).toBe(true);

    expect(exportedMachine.universalConstants?.entryActions?.[0]).not.toHaveProperty(
      "description",
    );
    expect(exportedMachine.universalConstants?.invokesOnTransition?.[0]).not.toHaveProperty(
      "description",
    );
    expect(
      exportedMachine.universes.main.universalConstants?.exitActions?.[0],
    ).not.toHaveProperty("description");
    expect(exportedMachine.universes.main.realities.idle.entryActions?.[0]).not.toHaveProperty(
      "description",
    );
    expect(exportedMachine.universes.main.realities.idle.exitInvokes?.[0]).not.toHaveProperty(
      "description",
    );
    expect(
      exportedMachine.universes.main.realities.idle.always?.[0].conditions?.[0],
    ).not.toHaveProperty("description");
    expect(exportedMachine.universes.main.realities.idle.always?.[0].conditions?.[0]?.src).toBe(
      "condition:ready",
    );
    expect(
      exportedMachine.universes.main.realities.idle.always?.[0].actions?.[0],
    ).not.toHaveProperty("description");
    expect(
      exportedMachine.universes.main.realities.idle.always?.[0].invokes?.[0],
    ).not.toHaveProperty("description");
    expect(
      exportedMachine.universes.main.realities.idle.on?.NEXT[0].conditions?.[0],
    ).not.toHaveProperty("description");
    expect(exportedMachine.universes.main.realities.idle.on?.NEXT[0].conditions?.[0]?.src).toBe(
      "condition:next",
    );
    expect(
      exportedMachine.universes.main.realities.idle.on?.NEXT[0].actions?.[0],
    ).not.toHaveProperty("description");
    expect(
      exportedMachine.universes.main.realities.idle.on?.NEXT[0].invokes?.[0],
    ).not.toHaveProperty("description");

    expect(exportedMachine.universes.main.realities.idle.always?.[0].description).toBe(
      "Always transition description",
    );

    expect(importedMachine.universalConstants?.entryActions?.[0]?.description).toBe(
      "Machine UC action",
    );
    expect(importedMachine.universes.main.universalConstants?.exitActions?.[0]?.description).toBe(
      "Universe UC action",
    );
    expect(
      importedMachine.universes.main.realities.idle.always?.[0].condition?.description,
    ).toBe("Always condition");
    expect(
      importedMachine.universes.main.realities.idle.on?.NEXT[0].actions?.[0].description,
    ).toBe("On action");
  });
});
