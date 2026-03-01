import { describe, expect, it } from "vitest";

import {
  collectBehaviorUsages,
  removeBehaviorReferences,
  type BehaviorStateSlices,
} from "../utils/libraryUsage";

const TARGET_SRC = "custom:action:target";

const createState = (): BehaviorStateSlices => ({
  machineConfig: {
    id: "machine-1",
    canonicalName: "machine-1",
    version: "1.0.0",
    description: "",
    initials: [],
    metadata: "{}",
    universalConstants: {
      entryActions: [{ src: TARGET_SRC }, { src: "keep:action:machine" }],
      exitActions: [],
      entryInvokes: [],
      exitInvokes: [{ src: TARGET_SRC }],
      actionsOnTransition: [{ src: TARGET_SRC }, { src: TARGET_SRC }],
      invokesOnTransition: [],
    },
  },
  nodes: [
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
        universalConstants: {
          entryActions: [{ src: TARGET_SRC }],
          exitActions: [],
          entryInvokes: [],
          exitInvokes: [],
          actionsOnTransition: [],
          invokesOnTransition: [],
        },
      },
    },
    {
      id: "r-1",
      type: "reality",
      x: 20,
      y: 20,
      data: {
        id: "idle",
        name: "idle",
        universeId: "u-1",
        isInitial: true,
        realityType: "normal",
        observers: [{ src: TARGET_SRC }],
        entryActions: [{ src: "keep:action:reality" }, { src: TARGET_SRC }],
        exitActions: [],
        entryInvokes: [],
        exitInvokes: [],
      },
    },
  ],
  transitions: [
    {
      id: "tr-1",
      sourceRealityId: "r-1",
      triggerKind: "on",
      eventName: "GO",
      type: "default",
      condition: { src: TARGET_SRC },
      conditions: [{ src: "keep:condition" }, { src: TARGET_SRC }],
      actions: [{ src: TARGET_SRC }, { src: TARGET_SRC }],
      invokes: [{ src: "keep:invoke" }],
      description: "",
      metadata: "",
      targets: ["idle"],
      order: 0,
    },
  ],
});

describe("libraryUsage", () => {
  it("detecta usos en machine, universe, reality y transition", () => {
    const state = createState();

    const summary = collectBehaviorUsages(TARGET_SRC, state);

    expect(summary.total).toBe(11);
    expect(
      summary.locations.some(
        (location) =>
          location.label === "Machine:machine-1 · universalConstants.entryActions[0]",
      ),
    ).toBe(true);
    expect(
      summary.locations.some(
        (location) => location.label === "Universe:main · universalConstants.entryActions[0]",
      ),
    ).toBe(true);
    expect(
      summary.locations.some((location) => location.label === "Reality:idle · observers[0]"),
    ).toBe(true);
    expect(
      summary.locations.some((location) => location.label === "Transition:tr-1 · condition"),
    ).toBe(true);
    expect(
      summary.locations.some((location) => location.label === "Transition:tr-1 · actions[1]"),
    ).toBe(true);
  });

  it("elimina referencias objetivo y conserva las restantes", () => {
    const state = createState();

    const result = removeBehaviorReferences(TARGET_SRC, state);

    expect(result.removedCount).toBe(11);

    expect(result.nextMachineConfig.universalConstants.entryActions).toEqual([
      { src: "keep:action:machine" },
    ]);
    expect(result.nextMachineConfig.universalConstants.actionsOnTransition).toEqual([]);

    const universe = result.nextNodes.find((node) => node.id === "u-1");
    expect(universe?.type === "universe" ? universe.data.universalConstants?.entryActions : []).toEqual([]);

    const reality = result.nextNodes.find((node) => node.id === "r-1");
    expect(reality?.type === "reality" ? reality.data.observers : []).toEqual([]);
    expect(reality?.type === "reality" ? reality.data.entryActions : []).toEqual([
      { src: "keep:action:reality" },
    ]);

    expect(result.nextTransitions[0]?.condition).toBeUndefined();
    expect(result.nextTransitions[0]?.conditions).toEqual([{ src: "keep:condition" }]);
    expect(result.nextTransitions[0]?.actions).toEqual([]);
    expect(result.nextTransitions[0]?.invokes).toEqual([{ src: "keep:invoke" }]);

    // Purity check.
    expect(
      state.machineConfig.universalConstants.entryActions.some(
        (behavior) => behavior.src === TARGET_SRC,
      ),
    ).toBe(true);
    expect(state.transitions[0]?.condition?.src).toBe(TARGET_SRC);
  });
});
