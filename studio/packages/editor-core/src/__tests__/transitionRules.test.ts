import { describe, expect, it } from "vitest";

import { buildInvalidNotifyTransitionMap, isInternalTargetRef, isInvalidNotifyTransition } from "../utils";
import type { EditorNode, EditorTransition } from "../types";

const baseNodes: EditorNode[] = [
  {
    id: "u-1",
    type: "universe",
    x: 0,
    y: 0,
    w: 300,
    h: 200,
    data: {
      id: "u-1",
      name: "u-1",
      canonicalName: "u-1",
      version: "1.0.0",
    },
  },
  {
    id: "u-2",
    type: "universe",
    x: 400,
    y: 0,
    w: 300,
    h: 200,
    data: {
      id: "u-2",
      name: "u-2",
      canonicalName: "u-2",
      version: "1.0.0",
    },
  },
  {
    id: "r-1",
    type: "reality",
    x: 10,
    y: 10,
    data: {
      id: "r-1",
      name: "r-1",
      universeId: "u-1",
      isInitial: true,
      realityType: "normal",
    },
  },
  {
    id: "r-2",
    type: "reality",
    x: 20,
    y: 20,
    data: {
      id: "r-2",
      name: "r-2",
      universeId: "u-1",
      isInitial: false,
      realityType: "normal",
    },
  },
  {
    id: "r-3",
    type: "reality",
    x: 420,
    y: 20,
    data: {
      id: "r-3",
      name: "r-3",
      universeId: "u-2",
      isInitial: false,
      realityType: "normal",
    },
  },
];

const transition = (type: "default" | "notify", targets: string[]): EditorTransition => ({
  id: `tr-${type}-${targets.join("-")}`,
  sourceRealityId: "r-1",
  triggerKind: "on",
  eventName: "GO",
  type,
  condition: undefined,
  conditions: [],
  actions: [],
  invokes: [],
  description: "",
  metadata: "",
  targets,
  order: 0,
});

describe("transitionRules", () => {
  it("detecta target interno por ref de realidad", () => {
    expect(isInternalTargetRef("r-1", "r-2", baseNodes)).toBe(true);
  });

  it("retorna false para target externo U:universe:reality", () => {
    expect(isInternalTargetRef("r-1", "U:u-2:r-3", baseNodes)).toBe(false);
  });

  it("marca notify inválido cuando contiene target interno", () => {
    expect(isInvalidNotifyTransition(transition("notify", ["r-2"]), baseNodes)).toBe(true);
  });

  it("no marca notify inválido si todos los targets son externos", () => {
    expect(
      isInvalidNotifyTransition(transition("notify", ["U:u-2", "U:u-2:r-3"]), baseNodes),
    ).toBe(false);
  });

  it("default no se marca inválido", () => {
    expect(isInvalidNotifyTransition(transition("default", ["r-2"]), baseNodes)).toBe(false);
  });

  it("precalcula invalid notify por transición en lote", () => {
    const transitions: EditorTransition[] = [
      transition("notify", ["r-2"]),
      transition("notify", ["U:u-2:r-3"]),
      transition("default", ["r-2"]),
    ];
    transitions[0].id = "notify-invalid";
    transitions[1].id = "notify-valid";
    transitions[2].id = "default-valid";

    const result = buildInvalidNotifyTransitionMap(transitions, baseNodes);

    expect(result.get("notify-invalid")).toBe(true);
    expect(result.get("notify-valid")).toBe(false);
    expect(result.get("default-valid")).toBe(false);
  });
});
