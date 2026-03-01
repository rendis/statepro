import { describe, expect, it } from "vitest";

import { getTransitionLegGeometry, getTransitionRouteGeometry } from "../utils";
import type { EditorNode, EditorTransition, NodeSizeMap, TransitionLeg } from "../types";

type ParsedPathCommand = {
  command: "M" | "L" | "C";
  values: number[];
};

const parsePathCommands = (d: string): ParsedPathCommand[] => {
  const chunks = d.match(/[MLC][^MLC]*/g) || [];
  return chunks.map((chunk) => {
    const command = chunk[0] as "M" | "L" | "C";
    const numeric = (chunk.slice(1).match(/-?\d+(?:\.\d+)?/g) || []).map(Number);
    return { command, values: numeric };
  });
};

const parsePathGeometry = (d: string) => {
  const commands = parsePathCommands(d);
  const move = commands.find((entry) => entry.command === "M");
  const cubic = commands.find((entry) => entry.command === "C");

  if (!move || move.values.length < 2 || !cubic || cubic.values.length < 6) {
    throw new Error(`Unexpected path command: ${d}`);
  }

  let current = { x: move.values[0], y: move.values[1] };
  let leadingLineEnd: { x: number; y: number } | null = null;
  let trailingLineStart: { x: number; y: number } | null = null;
  let trailingLineEnd: { x: number; y: number } | null = null;
  let seenCubic = false;

  commands.forEach((entry) => {
    if (entry.command === "M" && entry.values.length >= 2) {
      current = { x: entry.values[0], y: entry.values[1] };
      return;
    }

    if (entry.command === "L" && entry.values.length >= 2) {
      if (!seenCubic) {
        leadingLineEnd = { x: entry.values[0], y: entry.values[1] };
      } else {
        trailingLineStart = { ...current };
        trailingLineEnd = { x: entry.values[0], y: entry.values[1] };
      }
      current = { x: entry.values[0], y: entry.values[1] };
      return;
    }

    if (entry.command === "C" && entry.values.length >= 6) {
      seenCubic = true;
      current = { x: entry.values[4], y: entry.values[5] };
    }
  });

  return {
    startX: move.values[0],
    startY: move.values[1],
    control1X: cubic.values[0],
    control1Y: cubic.values[1],
    control2X: cubic.values[2],
    control2Y: cubic.values[3],
    cubicEndX: cubic.values[4],
    cubicEndY: cubic.values[5],
    endX: current.x,
    endY: current.y,
    leadingLineEndX: leadingLineEnd?.x ?? null,
    leadingLineEndY: leadingLineEnd?.y ?? null,
    trailingLineStartX: trailingLineStart?.x ?? null,
    trailingLineStartY: trailingLineStart?.y ?? null,
    trailingLineEndX: trailingLineEnd?.x ?? null,
    trailingLineEndY: trailingLineEnd?.y ?? null,
  };
};

const parsePathControls = (d: string) => {
  return parsePathGeometry(d);
};

const parseMoveTo = (d: string) => {
  const parsed = parsePathGeometry(d);
  return { x: parsed.startX, y: parsed.startY };
};

const parseEndPoint = (d: string) => {
  const parsed = parsePathGeometry(d);
  return { x: parsed.endX, y: parsed.endY };
};

describe("transition geometry", () => {
  it("applies same visualOffset to every leg of a multi-target transition", () => {
    const nodes: EditorNode[] = [
      {
        id: "real-src",
        type: "reality",
        x: 100,
        y: 120,
        data: {
          id: "source",
          name: "source",
          universeId: "univ-1",
          isInitial: true,
          realityType: "normal",
        },
      },
      {
        id: "real-a",
        type: "reality",
        x: 360,
        y: 140,
        data: {
          id: "target-a",
          name: "target-a",
          universeId: "univ-1",
          isInitial: false,
          realityType: "normal",
        },
      },
      {
        id: "real-b",
        type: "reality",
        x: 390,
        y: 280,
        data: {
          id: "target-b",
          name: "target-b",
          universeId: "univ-1",
          isInitial: false,
          realityType: "success",
        },
      },
    ];

    const nodeSizes: NodeSizeMap = {
      "real-src": { w: 192, h: 150 },
      "real-a": { w: 192, h: 150 },
      "real-b": { w: 192, h: 150 },
    };

    const legs: TransitionLeg[] = [
      {
        id: "tr-1::0",
        transitionId: "tr-1",
        source: "real-src",
        target: "real-a",
        targetRef: "target-a",
      },
      {
        id: "tr-1::1",
        transitionId: "tr-1",
        source: "real-src",
        target: "real-b",
        targetRef: "target-b",
      },
    ];

    const baseTransition: EditorTransition = {
      id: "tr-1",
      sourceRealityId: "real-src",
      triggerKind: "on",
      eventName: "GO",
      type: "default",
      condition: undefined,
      conditions: [],
      actions: [],
      invokes: [],
      description: "",
      metadata: "",
      targets: ["target-a", "target-b"],
      order: 0,
    };

    const withOffset: EditorTransition = {
      ...baseTransition,
      visualOffset: { x: 44, y: -28 },
    };

    const baseGeometries = legs.map((leg) => getTransitionLegGeometry(leg, nodes, nodeSizes, baseTransition));
    const movedGeometries = legs.map((leg) => getTransitionLegGeometry(leg, nodes, nodeSizes, withOffset));

    baseGeometries.forEach((geometry, index) => {
      const movedGeometry = movedGeometries[index];
      if (!geometry || !movedGeometry) {
        throw new Error("Expected transition geometry to exist");
      }

      const baseControls = parsePathControls(geometry.d);
      const movedControls = parsePathControls(movedGeometry.d);

      expect(movedControls.control1X - baseControls.control1X).toBeCloseTo(44, 5);
      expect(movedControls.control1Y - baseControls.control1Y).toBeCloseTo(-28, 5);
      expect(movedControls.control2X - baseControls.control2X).toBeCloseTo(44, 5);
      expect(movedControls.control2Y - baseControls.control2Y).toBeCloseTo(-28, 5);
    });
  });

  it("builds multi-target route with left/right hub ports and outbound branches", () => {
    const nodes: EditorNode[] = [
      {
        id: "real-src",
        type: "reality",
        x: 100,
        y: 120,
        data: {
          id: "source",
          name: "source",
          universeId: "univ-1",
          isInitial: true,
          realityType: "normal",
        },
      },
      {
        id: "real-a",
        type: "reality",
        x: 390,
        y: 140,
        data: {
          id: "target-a",
          name: "target-a",
          universeId: "univ-1",
          isInitial: false,
          realityType: "normal",
        },
      },
      {
        id: "real-b",
        type: "reality",
        x: 420,
        y: 280,
        data: {
          id: "target-b",
          name: "target-b",
          universeId: "univ-1",
          isInitial: false,
          realityType: "success",
        },
      },
    ];

    const nodeSizes: NodeSizeMap = {
      "real-src": { w: 192, h: 150 },
      "real-a": { w: 192, h: 150 },
      "real-b": { w: 192, h: 150 },
    };

    const legs: TransitionLeg[] = [
      {
        id: "tr-2::0",
        transitionId: "tr-2",
        source: "real-src",
        target: "real-a",
        targetRef: "target-a",
      },
      {
        id: "tr-2::1",
        transitionId: "tr-2",
        source: "real-src",
        target: "real-b",
        targetRef: "target-b",
      },
    ];

    const baseTransition: EditorTransition = {
      id: "tr-2",
      sourceRealityId: "real-src",
      triggerKind: "on",
      eventName: "GO",
      type: "default",
      condition: undefined,
      conditions: [],
      actions: [],
      invokes: [],
      description: "",
      metadata: "",
      targets: ["target-a", "target-b"],
      order: 0,
    };

    const movedTransition: EditorTransition = {
      ...baseTransition,
      visualOffset: { x: 30, y: -20 },
    };

    const baseRoute = getTransitionRouteGeometry(baseTransition, legs, nodes, nodeSizes);
    const movedRoute = getTransitionRouteGeometry(movedTransition, legs, nodes, nodeSizes);

    expect(baseRoute).toBeTruthy();
    expect(movedRoute).toBeTruthy();

    if (!baseRoute || !movedRoute) {
      return;
    }

    expect(baseRoute.anchor.x).toBeCloseTo(baseRoute.hubCenter.x, 5);
    expect(baseRoute.anchor.y).toBeCloseTo(baseRoute.hubCenter.y, 5);
    expect(baseRoute.leftPort.x).toBeLessThan(baseRoute.hubCenter.x);
    expect(baseRoute.rightPort.x).toBeGreaterThan(baseRoute.hubCenter.x);

    expect(baseRoute.segments).toHaveLength(4);
    expect(baseRoute.segments.filter((segment) => segment.role === "inbound")).toHaveLength(1);
    expect(baseRoute.segments.filter((segment) => segment.role === "bridge")).toHaveLength(1);
    expect(baseRoute.segments.filter((segment) => segment.role === "outbound")).toHaveLength(2);

    const inbound = baseRoute.segments.find((segment) => segment.role === "inbound");
    const bridge = baseRoute.segments.find((segment) => segment.role === "bridge");
    const outbound = baseRoute.segments.filter((segment) => segment.role === "outbound");

    expect(inbound?.hasArrow).toBe(false);
    expect(bridge?.hasArrow).toBe(false);
    outbound.forEach((segment) => expect(segment.hasArrow).toBe(true));

    const inboundEnd = parseEndPoint(inbound?.d || "");
    expect(inboundEnd.x).toBeCloseTo(baseRoute.leftPort.x, 5);
    expect(inboundEnd.y).toBeCloseTo(baseRoute.leftPort.y, 5);

    const inboundShape = parsePathGeometry(inbound?.d || "");
    expect(inboundShape.leadingLineEndX).not.toBeNull();
    expect(inboundShape.trailingLineStartX).not.toBeNull();
    if (
      inboundShape.leadingLineEndX === null ||
      inboundShape.trailingLineStartX === null
    ) {
      throw new Error("Expected stubbed path segments");
    }
    expect(Math.abs(inboundShape.leadingLineEndX - inboundShape.startX)).toBeCloseTo(14, 1);
    expect(Math.abs(inboundShape.endX - inboundShape.trailingLineStartX)).toBeCloseTo(14, 1);

    outbound.forEach((segment) => {
      const branchStart = parseMoveTo(segment.d);
      expect(branchStart.x).toBeCloseTo(baseRoute.rightPort.x, 5);
      expect(branchStart.y).toBeCloseTo(baseRoute.rightPort.y, 5);

      const branchShape = parsePathGeometry(segment.d);
      expect(branchShape.leadingLineEndX).not.toBeNull();
      expect(branchShape.trailingLineStartX).not.toBeNull();
      if (
        branchShape.leadingLineEndX === null ||
        branchShape.trailingLineStartX === null
      ) {
        throw new Error("Expected stubbed outbound path segments");
      }
      expect(Math.abs(branchShape.leadingLineEndX - branchShape.startX)).toBeCloseTo(14, 1);
      expect(Math.abs(branchShape.endX - branchShape.trailingLineStartX)).toBeCloseTo(14, 1);
    });

    expect(movedRoute.hubCenter.x - baseRoute.hubCenter.x).toBeCloseTo(30, 5);
    expect(movedRoute.hubCenter.y - baseRoute.hubCenter.y).toBeCloseTo(-20, 5);
    expect(movedRoute.leftPort.x - baseRoute.leftPort.x).toBeCloseTo(30, 5);
    expect(movedRoute.rightPort.y - baseRoute.rightPort.y).toBeCloseTo(-20, 5);
  });

  it("is deterministic across repeated geometry calculations", () => {
    const nodes: EditorNode[] = [
      {
        id: "real-src",
        type: "reality",
        x: 100,
        y: 120,
        data: {
          id: "source",
          name: "source",
          universeId: "univ-1",
          isInitial: true,
          realityType: "normal",
        },
      },
      {
        id: "real-a",
        type: "reality",
        x: 390,
        y: 140,
        data: {
          id: "target-a",
          name: "target-a",
          universeId: "univ-1",
          isInitial: false,
          realityType: "normal",
        },
      },
      {
        id: "real-b",
        type: "reality",
        x: 420,
        y: 280,
        data: {
          id: "target-b",
          name: "target-b",
          universeId: "univ-1",
          isInitial: false,
          realityType: "success",
        },
      },
    ];

    const nodeSizes: NodeSizeMap = {
      "real-src": { w: 192, h: 150 },
      "real-a": { w: 192, h: 150 },
      "real-b": { w: 192, h: 150 },
    };

    const legs: TransitionLeg[] = [
      {
        id: "tr-3::0",
        transitionId: "tr-3",
        source: "real-src",
        target: "real-a",
        targetRef: "target-a",
      },
      {
        id: "tr-3::1",
        transitionId: "tr-3",
        source: "real-src",
        target: "real-b",
        targetRef: "target-b",
      },
    ];

    const transition: EditorTransition = {
      id: "tr-3",
      sourceRealityId: "real-src",
      triggerKind: "on",
      eventName: "GO",
      type: "default",
      condition: undefined,
      conditions: [],
      actions: [],
      invokes: [],
      description: "",
      metadata: "",
      targets: ["target-a", "target-b"],
      order: 0,
    };

    const firstRoute = getTransitionRouteGeometry(transition, legs, nodes, nodeSizes);
    expect(firstRoute).toBeTruthy();
    if (!firstRoute) {
      return;
    }

    for (let iteration = 0; iteration < 5; iteration += 1) {
      const nextRoute = getTransitionRouteGeometry(transition, legs, nodes, nodeSizes);
      expect(nextRoute).toEqual(firstRoute);
    }

    const inbound = firstRoute.segments.find((segment) => segment.role === "inbound");
    const outbound = firstRoute.segments.filter((segment) => segment.role === "outbound");
    const inboundEnd = parseEndPoint(inbound?.d || "");

    expect(inboundEnd.x).toBeCloseTo(firstRoute.leftPort.x, 5);
    expect(inboundEnd.y).toBeCloseTo(firstRoute.leftPort.y, 5);
    outbound.forEach((segment) => {
      const outboundStart = parseMoveTo(segment.d);
      expect(outboundStart.x).toBeCloseTo(firstRoute.rightPort.x, 5);
      expect(outboundStart.y).toBeCloseTo(firstRoute.rightPort.y, 5);
    });
  });

  it("keeps fixed local tangents even when source/target are forced to opposite sides", () => {
    const nodes: EditorNode[] = [
      {
        id: "real-src",
        type: "reality",
        x: 520,
        y: 120,
        data: {
          id: "source",
          name: "source",
          universeId: "univ-1",
          isInitial: true,
          realityType: "normal",
        },
      },
      {
        id: "real-target",
        type: "reality",
        x: 260,
        y: 180,
        data: {
          id: "target",
          name: "target",
          universeId: "univ-1",
          isInitial: false,
          realityType: "normal",
        },
      },
    ];

    const nodeSizes: NodeSizeMap = {
      "real-src": { w: 192, h: 150 },
      "real-target": { w: 192, h: 150 },
    };

    const legs: TransitionLeg[] = [
      {
        id: "tr-4::0",
        transitionId: "tr-4",
        source: "real-src",
        target: "real-target",
        targetRef: "target",
      },
    ];

    const transition: EditorTransition = {
      id: "tr-4",
      sourceRealityId: "real-src",
      triggerKind: "on",
      eventName: "GO",
      type: "default",
      condition: undefined,
      conditions: [],
      actions: [],
      invokes: [],
      description: "",
      metadata: "",
      targets: ["target"],
      order: 0,
    };

    const route = getTransitionRouteGeometry(transition, legs, nodes, nodeSizes);
    expect(route).toBeTruthy();
    if (!route) {
      return;
    }

    const inbound = route.segments.find((segment) => segment.role === "inbound");
    const outbound = route.segments.find((segment) => segment.role === "outbound");
    expect(inbound).toBeTruthy();
    expect(outbound).toBeTruthy();
    if (!inbound || !outbound) {
      return;
    }

    const inboundControls = parsePathControls(inbound.d);
    const outboundControls = parsePathControls(outbound.d);

    // Always leave source towards East and approach the left badge port from West.
    expect(inboundControls.control1X).toBeGreaterThan(inboundControls.startX);
    expect(inboundControls.control2X).toBeLessThan(inboundControls.endX);

    // Always leave right badge port towards East and approach target from West.
    expect(outboundControls.control1X).toBeGreaterThan(outboundControls.startX);
    expect(outboundControls.control2X).toBeLessThan(outboundControls.endX);
  });
});
