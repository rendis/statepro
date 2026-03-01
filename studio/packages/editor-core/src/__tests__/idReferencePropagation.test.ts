import { describe, expect, it } from "vitest";

import { renameRealityId, renameUniverseId } from "../utils";
import type { EditorNode, EditorTransition, MachineConfig, MetadataPackBindingMap } from "../types";

const baseNodes: EditorNode[] = [
  {
    id: "u-1",
    type: "universe",
    x: 0,
    y: 0,
    w: 400,
    h: 300,
    data: {
      id: "main",
      name: "main",
      canonicalName: "main",
      version: "1.0.0",
    },
  },
  {
    id: "u-2",
    type: "universe",
    x: 500,
    y: 0,
    w: 400,
    h: 300,
    data: {
      id: "aux",
      name: "aux",
      canonicalName: "aux",
      version: "1.0.0",
    },
  },
  {
    id: "r-1",
    type: "reality",
    x: 10,
    y: 10,
    data: {
      id: "idle",
      name: "idle",
      universeId: "u-1",
      isInitial: true,
      realityType: "normal",
    },
  },
  {
    id: "r-2",
    type: "reality",
    x: 520,
    y: 10,
    data: {
      id: "pending",
      name: "pending",
      universeId: "u-2",
      isInitial: true,
      realityType: "normal",
    },
  },
];

const baseTransitions: EditorTransition[] = [
  {
    id: "tr-1",
    sourceRealityId: "r-1",
    triggerKind: "on",
    eventName: "GO",
    type: "default",
    condition: undefined,
    conditions: [],
    actions: [],
    invokes: [],
    metadata: "",
    targets: ["idle", "U:main", "U:main:idle", "U:aux:pending"],
    order: 0,
  },
  {
    id: "tr-2",
    sourceRealityId: "r-2",
    triggerKind: "on",
    eventName: "BACK",
    type: "default",
    condition: undefined,
    conditions: [],
    actions: [],
    invokes: [],
    metadata: "",
    targets: ["idle", "U:main:idle"],
    order: 0,
  },
];

const baseMachineConfig: MachineConfig = {
  id: "machine-id",
  canonicalName: "machine",
  version: "1.0.0",
  initials: ["U:main", "U:main:idle", "U:aux:pending", "U:aux:idle"],
  universalConstants: {
    entryActions: [],
    exitActions: [],
    entryInvokes: [],
    exitInvokes: [],
    actionsOnTransition: [],
    invokesOnTransition: [],
  },
  metadata: "{}",
};

const baseBindings: MetadataPackBindingMap = {
  machine: [],
  universe: [
    {
      id: "b-universe-main",
      packId: "pack-u",
      scope: "universe",
      entityRef: "U:main",
      values: {},
    },
    {
      id: "b-universe-aux",
      packId: "pack-u",
      scope: "universe",
      entityRef: "U:aux",
      values: {},
    },
  ],
  reality: [
    {
      id: "b-reality-main-idle",
      packId: "pack-r",
      scope: "reality",
      entityRef: "U:main:R:idle",
      values: {},
    },
    {
      id: "b-reality-aux-pending",
      packId: "pack-r",
      scope: "reality",
      entityRef: "U:aux:R:pending",
      values: {},
    },
  ],
  transition: [
    {
      id: "b-transition-main-idle",
      packId: "pack-t",
      scope: "transition",
      entityRef: "U:main:R:idle:T:on:GO:0",
      values: {},
    },
    {
      id: "b-transition-aux-pending",
      packId: "pack-t",
      scope: "transition",
      entityRef: "U:aux:R:pending:T:on:BACK:0",
      values: {},
    },
  ],
};

describe("idReferencePropagation", () => {
  it("renameUniverseId propaga referencias de universo, realidad y transition entityRef", () => {
    const result = renameUniverseId({
      nodes: structuredClone(baseNodes),
      transitions: structuredClone(baseTransitions),
      machineConfig: structuredClone(baseMachineConfig),
      metadataPackBindings: structuredClone(baseBindings),
      universeNodeId: "u-1",
      nextUniverseId: "primary",
    });

    const renamedUniverse = result.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe" && node.id === "u-1",
    );
    expect(renamedUniverse?.data.id).toBe("primary");
    expect(renamedUniverse?.data.name).toBe("primary");

    expect(result.transitions[0]?.targets).toEqual(["idle", "U:primary", "U:primary:idle", "U:aux:pending"]);
    expect(result.transitions[1]?.targets).toEqual(["idle", "U:primary:idle"]);

    expect(result.machineConfig.initials).toEqual(["U:primary", "U:primary:idle", "U:aux:pending", "U:aux:idle"]);

    expect(result.metadataPackBindings.universe.map((binding) => binding.entityRef)).toEqual([
      "U:primary",
      "U:aux",
    ]);
    expect(result.metadataPackBindings.reality.map((binding) => binding.entityRef)).toEqual([
      "U:primary:R:idle",
      "U:aux:R:pending",
    ]);
    expect(result.metadataPackBindings.transition.map((binding) => binding.entityRef)).toEqual([
      "U:primary:R:idle:T:on:GO:0",
      "U:aux:R:pending:T:on:BACK:0",
    ]);
  });

  it("renameRealityId remapea refs internas y externas segun universo fuente", () => {
    const result = renameRealityId({
      nodes: structuredClone(baseNodes),
      transitions: structuredClone(baseTransitions),
      machineConfig: structuredClone(baseMachineConfig),
      metadataPackBindings: structuredClone(baseBindings),
      realityNodeId: "r-1",
      nextRealityId: "ready",
    });

    const renamedReality = result.nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality" && node.id === "r-1",
    );
    expect(renamedReality?.data.id).toBe("ready");
    expect(renamedReality?.data.name).toBe("ready");

    // transition source in "main": internal + external refs to main.idle are remapped
    expect(result.transitions[0]?.targets).toEqual(["ready", "U:main", "U:main:ready", "U:aux:pending"]);
    // transition source in "aux": internal "idle" stays unchanged, external ref to main.idle is remapped
    expect(result.transitions[1]?.targets).toEqual(["idle", "U:main:ready"]);

    expect(result.machineConfig.initials).toEqual([
      "U:main",
      "U:main:ready",
      "U:aux:pending",
      "U:aux:idle",
    ]);

    expect(result.metadataPackBindings.reality.map((binding) => binding.entityRef)).toEqual([
      "U:main:R:ready",
      "U:aux:R:pending",
    ]);
    expect(result.metadataPackBindings.transition.map((binding) => binding.entityRef)).toEqual([
      "U:main:R:ready:T:on:GO:0",
      "U:aux:R:pending:T:on:BACK:0",
    ]);
  });
});
