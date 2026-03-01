import { describe, expect, it } from "vitest";

import type { EditorNode, EditorTransition, NodeSizeMap, TransitionLeg } from "../types";
import { buildCanvasCullResult, computeCanvasViewportBounds } from "../utils/canvasCulling";

const nodes: EditorNode[] = [
  {
    id: "real-a",
    type: "reality",
    x: 40,
    y: 60,
    data: {
      id: "a",
      name: "A",
      universeId: "u-1",
      isInitial: true,
      realityType: "normal",
    },
  },
  {
    id: "real-b",
    type: "reality",
    x: 1200,
    y: 1200,
    data: {
      id: "b",
      name: "B",
      universeId: "u-2",
      isInitial: false,
      realityType: "normal",
    },
  },
  {
    id: "real-c",
    type: "reality",
    x: 1540,
    y: 1460,
    data: {
      id: "c",
      name: "C",
      universeId: "u-2",
      isInitial: false,
      realityType: "success",
    },
  },
];

const transitions: EditorTransition[] = [
  {
    id: "tr-visible",
    sourceRealityId: "real-a",
    triggerKind: "on",
    eventName: "GO",
    type: "default",
    condition: undefined,
    conditions: [],
    actions: [],
    invokes: [],
    description: "",
    metadata: "",
    targets: ["U:u-2:b"],
    order: 0,
  },
  {
    id: "tr-offscreen",
    sourceRealityId: "real-b",
    triggerKind: "on",
    eventName: "GO",
    type: "default",
    condition: undefined,
    conditions: [],
    actions: [],
    invokes: [],
    description: "",
    metadata: "",
    targets: ["U:u-2:c"],
    order: 0,
  },
];

const nodeSizes: NodeSizeMap = {
  "real-a": { w: 192, h: 150 },
  "real-b": { w: 192, h: 150 },
  "real-c": { w: 192, h: 150 },
};

const transitionLegsByTransitionId = new Map<string, TransitionLeg[]>([
  [
    "tr-visible",
    [
      {
        id: "tr-visible::0",
        transitionId: "tr-visible",
        source: "real-a",
        target: "real-b",
        targetRef: "U:u-2:b",
      },
    ],
  ],
  [
    "tr-offscreen",
    [
      {
        id: "tr-offscreen::0",
        transitionId: "tr-offscreen",
        source: "real-b",
        target: "real-c",
        targetRef: "U:u-2:c",
      },
    ],
  ],
]);

describe("canvas culling", () => {
  it("omite elementos fuera de viewport en perf mode", () => {
    const viewportBounds = {
      left: -100,
      top: -100,
      right: 500,
      bottom: 500,
    };

    const result = buildCanvasCullResult({
      nodes,
      transitions,
      nodeSizes,
      transitionLegsByTransitionId,
      viewportBounds,
    });

    expect(result.visibleNodeIds.has("real-a")).toBe(true);
    expect(result.visibleNodeIds.has("real-b")).toBe(false);
    expect(result.visibleNodeIds.has("real-c")).toBe(false);
    expect(result.visibleTransitionIds.has("tr-visible")).toBe(true);
    expect(result.visibleTransitionIds.has("tr-offscreen")).toBe(false);
  });

  it("conserva seleccionados/conectados aunque estén fuera de viewport", () => {
    const viewportBounds = {
      left: -100,
      top: -100,
      right: 500,
      bottom: 500,
    };

    const result = buildCanvasCullResult({
      nodes,
      transitions,
      nodeSizes,
      transitionLegsByTransitionId,
      viewportBounds,
      alwaysVisibleNodeIds: ["real-c"],
      alwaysVisibleTransitionIds: ["tr-offscreen"],
    });

    expect(result.visibleNodeIds.has("real-c")).toBe(true);
    expect(result.visibleTransitionIds.has("tr-offscreen")).toBe(true);
    expect(result.visibleNodeIds.has("real-b")).toBe(true);
  });

  it("calcula bounds de viewport a partir de transform", () => {
    const bounds = computeCanvasViewportBounds(
      { width: 1280, height: 720 },
      { scale: 1.5, positionX: -300, positionY: -150 },
      200,
    );

    expect(bounds).not.toBeNull();
    expect(bounds?.left).toBeCloseTo(0, 5);
    expect(bounds?.top).toBeCloseTo(-100, 5);
    expect(bounds?.right).toBeGreaterThan(bounds?.left || 0);
    expect(bounds?.bottom).toBeGreaterThan(bounds?.top || 0);
  });

  it("fail-open: si el culling deja 0 nodos visibles mantiene el render completo", () => {
    const result = buildCanvasCullResult({
      nodes,
      transitions,
      nodeSizes,
      transitionLegsByTransitionId,
      viewportBounds: {
        left: 5000,
        top: 5000,
        right: 5100,
        bottom: 5100,
      },
    });

    expect(result.visibleNodeIds.size).toBe(nodes.length);
    expect(result.visibleTransitionIds.size).toBe(transitions.length);
  });

  it("retorna null para viewport inválido", () => {
    const bounds = computeCanvasViewportBounds(
      { width: 0, height: 720 },
      { scale: 1, positionX: 0, positionY: 0 },
      200,
    );

    expect(bounds).toBeNull();
  });
});
