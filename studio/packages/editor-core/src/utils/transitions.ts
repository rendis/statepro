import type { EditorNode, EditorTransition, NodeSizeMap, TransitionLeg } from "../types";
import {
  buildNodeReferenceIndex,
  parseTargetReference,
  resolveTargetReferenceToNodeId,
} from "./references";
import { BADGE_WIDTH } from "./transitionBadgeLayout";

interface Point {
  x: number;
  y: number;
}

interface CubicCurve {
  start: Point;
  controlA: Point;
  controlB: Point;
  end: Point;
}

export interface TransitionLegGeometry {
  d: string;
  midpoint: Point;
}

export interface TransitionRouteSegment {
  id: string;
  d: string;
  role: "inbound" | "bridge" | "outbound";
  hasArrow: boolean;
}

export interface TransitionRouteGeometry {
  anchor: Point;
  hubCenter: Point;
  leftPort: Point;
  rightPort: Point;
  segments: TransitionRouteSegment[];
}

const DEFAULT_OFFSET = { x: 0, y: 0 } as const;
const clamp = (value: number, min: number, max: number) =>
  Math.max(min, Math.min(max, value));
const PORT_RUNOUT_PX = 28;
const PORT_RUNOUT_MIN_PX = 16;
const PORT_RUNOUT_MAX_PX = 48;
const PORT_STUB_PX = 14;
const PORT_STUB_MIN_PX = 2;
const MIN_CURVE_SPAN_PX = 10;

const getCubicBezierPoint = (
  start: Point,
  controlA: Point,
  controlB: Point,
  end: Point,
  t: number,
): Point => {
  const inverseT = 1 - t;
  const a = inverseT ** 3;
  const b = 3 * inverseT ** 2 * t;
  const c = 3 * inverseT * t ** 2;
  const d = t ** 3;

  return {
    x: a * start.x + b * controlA.x + c * controlB.x + d * end.x,
    y: a * start.y + b * controlA.y + c * controlB.y + d * end.y,
  };
};

const getCurvePoint = (curve: CubicCurve, t: number) =>
  getCubicBezierPoint(curve.start, curve.controlA, curve.controlB, curve.end, t);

const curveToPath = (curve: CubicCurve) =>
  `M ${curve.start.x} ${curve.start.y} C ${curve.controlA.x} ${curve.controlA.y}, ${curve.controlB.x} ${curve.controlB.y}, ${curve.end.x} ${curve.end.y}`;

const resolvePortRunout = (
  start: Point,
  end: Point,
  preferredRunout: number = PORT_RUNOUT_PX,
): number => {
  const clampedPreferred = clamp(preferredRunout, PORT_RUNOUT_MIN_PX, PORT_RUNOUT_MAX_PX);
  const distance = Math.sqrt(
    (end.x - start.x) * (end.x - start.x) +
      (end.y - start.y) * (end.y - start.y),
  );
  const distanceLimited = clamp(
    distance * 0.45,
    PORT_RUNOUT_MIN_PX,
    PORT_RUNOUT_MAX_PX,
  );

  return Math.min(clampedPreferred, distanceLimited);
};

const buildPortLockedCubic = ({
  start,
  end,
  startNormalX,
  endNormalX,
  runout,
}: {
  start: Point;
  end: Point;
  startNormalX: -1 | 1;
  endNormalX: -1 | 1;
  runout: number;
}): CubicCurve => ({
  start,
  controlA: {
    x: start.x + startNormalX * runout,
    y: start.y,
  },
  controlB: {
    x: end.x + endNormalX * runout,
    y: end.y,
  },
  end,
});

const buildPortLockedStubbedPath = ({
  start,
  end,
  startNormalX,
  endNormalX,
  stubPx = PORT_STUB_PX,
  runoutPx = PORT_RUNOUT_PX,
}: {
  start: Point;
  end: Point;
  startNormalX: -1 | 1;
  endNormalX: -1 | 1;
  stubPx?: number;
  runoutPx?: number;
}): string => {
  const distance = Math.sqrt(
    (end.x - start.x) * (end.x - start.x) +
      (end.y - start.y) * (end.y - start.y),
  );
  const maxStubPerSide = Math.max(0, (distance - MIN_CURVE_SPAN_PX) / 2);
  const resolvedStub = clamp(Math.min(stubPx, maxStubPerSide), 0, stubPx);

  if (resolvedStub < PORT_STUB_MIN_PX) {
    const fallbackCurve = buildPortLockedCubic({
      start,
      end,
      startNormalX,
      endNormalX,
      runout: resolvePortRunout(start, end, runoutPx),
    });
    return curveToPath(fallbackCurve);
  }

  const stubStart: Point = {
    x: start.x + startNormalX * resolvedStub,
    y: start.y,
  };
  const stubEnd: Point = {
    x: end.x + endNormalX * resolvedStub,
    y: end.y,
  };
  const curveRunout = resolvePortRunout(stubStart, stubEnd, runoutPx);
  const curve = buildPortLockedCubic({
    start: stubStart,
    end: stubEnd,
    startNormalX,
    endNormalX,
    runout: curveRunout,
  });

  return `M ${start.x} ${start.y} L ${stubStart.x} ${stubStart.y} C ${curve.controlA.x} ${curve.controlA.y}, ${curve.controlB.x} ${curve.controlB.y}, ${curve.end.x} ${curve.end.y} L ${end.x} ${end.y}`;
};

const buildSingleSegmentCurve = (
  start: Point,
  end: Point,
  offset: Point,
): CubicCurve => {
  const deltaX = end.x - start.x;
  const deltaY = end.y - start.y;
  const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);
  const handleLength = Math.max(distance / 2.5, 50);

  return {
    start,
    controlA: {
      x: start.x + handleLength + offset.x,
      y: start.y + offset.y,
    },
    controlB: {
      x: end.x - handleLength + offset.x,
      y: end.y + offset.y,
    },
    end,
  };
};

export const getTransitionLegGeometry = (
  leg: TransitionLeg,
  nodes: EditorNode[],
  nodeSizes: NodeSizeMap,
  transition?: Pick<EditorTransition, "visualOffset">,
): TransitionLegGeometry | null => {
  const endpoints = getLegEndpoints(leg, nodes, nodeSizes);
  if (!endpoints) {
    return null;
  }

  const offset = transition?.visualOffset || DEFAULT_OFFSET;
  const segment = buildSingleSegmentPath(endpoints.start, endpoints.end, offset);

  return {
    d: segment.d,
    midpoint: segment.midpoint,
  };
};

const getLegEndpoints = (
  leg: TransitionLeg,
  nodes: EditorNode[],
  nodeSizes: NodeSizeMap,
): { start: Point; end: Point } | null => {
  const sourceSize = nodeSizes[leg.source] || { w: 192, h: 80 };
  const targetNode = nodes.find((node) => node.id === leg.target);
  const sourceNode = nodes.find((node) => node.id === leg.source);
  if (!sourceNode || !targetNode) {
    return null;
  }

  const targetIsUniverse = targetNode.type === "universe";
  const targetH = targetIsUniverse ? targetNode.h : nodeSizes[leg.target]?.h || 80;

  return {
    start: {
      x: sourceNode.x + sourceSize.w + 6,
      y: sourceNode.y + sourceSize.h / 2,
    },
    end: {
      x: targetNode.x - 10,
      y: targetNode.y + targetH / 2,
    },
  };
};

const buildSingleSegmentPath = (
  start: Point,
  end: Point,
  offset: Point,
): { d: string; midpoint: Point } => {
  const curve = buildSingleSegmentCurve(start, end, offset);

  return {
    d: curveToPath(curve),
    midpoint: getCurvePoint(curve, 0.5),
  };
};

export const getTransitionRouteGeometry = (
  transition: EditorTransition,
  legs: TransitionLeg[],
  nodes: EditorNode[],
  nodeSizes: NodeSizeMap,
): TransitionRouteGeometry | null => {
  const validLegEndpoints = legs
    .map((leg) => ({ leg, endpoints: getLegEndpoints(leg, nodes, nodeSizes) }))
    .filter(
      (entry): entry is { leg: TransitionLeg; endpoints: { start: Point; end: Point } } =>
        Boolean(entry.endpoints),
    );

  if (validLegEndpoints.length === 0) {
    return null;
  }

  const offset = transition.visualOffset || DEFAULT_OFFSET;
  const baseStart = validLegEndpoints[0]?.endpoints.start;
  if (!baseStart) {
    return null;
  }

  const endPoints = validLegEndpoints.map((entry) => entry.endpoints.end);
  const centroid = {
    x: endPoints.reduce((sum, point) => sum + point.x, 0) / endPoints.length,
    y: endPoints.reduce((sum, point) => sum + point.y, 0) / endPoints.length,
  };

  const badgeHalfWidth = BADGE_WIDTH / 2;
  const sourcePadding = 24;
  const targetPadding = 24;
  const minEndX = Math.min(...endPoints.map((point) => point.x));
  const desiredHubX = baseStart.x + clamp((centroid.x - baseStart.x) * 0.45, 90, 240);
  const minHubX = baseStart.x + badgeHalfWidth + sourcePadding;
  const maxHubX = minEndX - badgeHalfWidth - targetPadding;
  const resolvedHubX =
    maxHubX >= minHubX
      ? clamp(desiredHubX, minHubX, maxHubX)
      : baseStart.x + (minEndX - baseStart.x) / 2;
  const hubCenter: Point = {
    x: resolvedHubX + offset.x,
    y: baseStart.y + (centroid.y - baseStart.y) * 0.35 + offset.y,
  };
  const portHalfGap = badgeHalfWidth;
  const leftPort: Point = {
    x: hubCenter.x - portHalfGap,
    y: hubCenter.y,
  };
  const rightPort: Point = {
    x: hubCenter.x + portHalfGap,
    y: hubCenter.y,
  };

  const inboundSegment: TransitionRouteSegment = {
    id: `${transition.id}::inbound`,
    d: buildPortLockedStubbedPath({
      start: baseStart,
      end: leftPort,
      startNormalX: 1,
      endNormalX: -1,
      stubPx: PORT_STUB_PX,
      runoutPx: PORT_RUNOUT_PX,
    }),
    role: "inbound",
    hasArrow: false,
  };

  const bridgeSegment: TransitionRouteSegment = {
    id: `${transition.id}::bridge`,
    d: buildPortLockedStubbedPath({
      start: leftPort,
      end: rightPort,
      startNormalX: 1,
      endNormalX: -1,
      stubPx: PORT_STUB_PX,
      runoutPx: PORT_RUNOUT_PX,
    }),
    role: "bridge",
    hasArrow: false,
  };

  const branchSegments = validLegEndpoints.map(({ leg, endpoints }, index) => {
    return {
      id: `${transition.id}::outbound::${index}::${leg.id}`,
      d: buildPortLockedStubbedPath({
        start: rightPort,
        end: endpoints.end,
        startNormalX: 1,
        endNormalX: -1,
        stubPx: PORT_STUB_PX,
        runoutPx: PORT_RUNOUT_PX,
      }),
      role: "outbound" as const,
      hasArrow: true,
    };
  });

  return {
    anchor: hubCenter,
    hubCenter,
    leftPort,
    rightPort,
    segments: [inboundSegment, bridgeSegment, ...branchSegments],
  };
};

export const getTransitionGroupKey = (transition: EditorTransition): string => {
  if (transition.triggerKind === "always") {
    return `${transition.sourceRealityId}::always`;
  }

  return `${transition.sourceRealityId}::on::${transition.eventName || ""}`;
};

export const buildTransitionLegs = (
  transitions: EditorTransition[],
  nodes: EditorNode[],
): TransitionLeg[] => {
  const index = buildNodeReferenceIndex(nodes);
  const legs: TransitionLeg[] = [];

  transitions.forEach((transition) => {
    transition.targets.forEach((targetRef, targetIndex) => {
      const targetNodeId = resolveTargetReferenceToNodeId(
        transition.sourceRealityId,
        targetRef,
        nodes,
        index,
      );

      if (!targetNodeId) {
        return;
      }

      legs.push({
        id: `${transition.id}::${targetIndex}`,
        transitionId: transition.id,
        source: transition.sourceRealityId,
        target: targetNodeId,
        targetRef,
      });
    });
  });

  return legs;
};

export const removeTransitionsReferencingDeletedNodes = (
  transitions: EditorTransition[],
  nodes: EditorNode[],
  nodeIdsToDelete: string[],
): EditorTransition[] => {
  if (nodeIdsToDelete.length === 0) {
    return transitions;
  }

  const deletedNodeIdSet = new Set(nodeIdsToDelete);
  const universeByNodeId = new Map<string, Extract<EditorNode, { type: "universe" }>>();
  const realityByNodeId = new Map<string, Extract<EditorNode, { type: "reality" }>>();

  nodes.forEach((node) => {
    if (node.type === "universe") {
      universeByNodeId.set(node.id, node);
      return;
    }

    if (node.type === "reality") {
      realityByNodeId.set(node.id, node);
    }
  });

  const deletedUniverseDataIds = new Set<string>();
  const deletedRealityByUniverseNodeAndDataId = new Set<string>();
  const deletedRealityByUniverseDataAndDataId = new Set<string>();

  nodes.forEach((node) => {
    if (!deletedNodeIdSet.has(node.id)) {
      return;
    }

    if (node.type === "universe") {
      deletedUniverseDataIds.add(node.data.id);
      return;
    }

    if (node.type !== "reality") {
      return;
    }

    deletedRealityByUniverseNodeAndDataId.add(
      `${node.data.universeId}::${node.data.id}`,
    );

    const universe = universeByNodeId.get(node.data.universeId);
    if (universe) {
      deletedRealityByUniverseDataAndDataId.add(
        `${universe.data.id}::${node.data.id}`,
      );
    }
  });

  return transitions.filter((transition) => {
    if (deletedNodeIdSet.has(transition.sourceRealityId)) {
      return false;
    }

    const sourceReality = realityByNodeId.get(transition.sourceRealityId);

    return !transition.targets.some((targetRef) => {
      const parsed = parseTargetReference(targetRef);
      if (!parsed) {
        return false;
      }

      if (parsed.kind === "universe") {
        return deletedUniverseDataIds.has(parsed.universeId || "");
      }

      if (parsed.kind === "universeReality") {
        const universeDataId = parsed.universeId || "";
        const realityDataId = parsed.realityId || "";

        if (deletedUniverseDataIds.has(universeDataId)) {
          return true;
        }

        return deletedRealityByUniverseDataAndDataId.has(
          `${universeDataId}::${realityDataId}`,
        );
      }

      if (!sourceReality) {
        return false;
      }

      return deletedRealityByUniverseNodeAndDataId.has(
        `${sourceReality.data.universeId}::${parsed.realityId || ""}`,
      );
    });
  });
};

export const sortTransitionsByExecutionOrder = (
  transitions: EditorTransition[],
): EditorTransition[] => {
  return [...transitions].sort((a, b) => {
    const keyA = getTransitionGroupKey(a);
    const keyB = getTransitionGroupKey(b);

    if (keyA === keyB) {
      return a.order - b.order;
    }

    return keyA.localeCompare(keyB);
  });
};

export const normalizeTransitionsOrder = (
  transitions: EditorTransition[],
): EditorTransition[] => {
  const grouped = new Map<string, EditorTransition[]>();

  transitions.forEach((transition) => {
    const key = getTransitionGroupKey(transition);
    const current = grouped.get(key) || [];
    current.push(transition);
    grouped.set(key, current);
  });

  const normalized: EditorTransition[] = [];

  grouped.forEach((groupTransitions) => {
    const sorted = [...groupTransitions].sort((a, b) => a.order - b.order);
    sorted.forEach((transition, index) => {
      normalized.push({
        ...transition,
        order: index,
      });
    });
  });

  return sortTransitionsByExecutionOrder(normalized);
};

export const moveTransitionInsideGroup = (
  transitions: EditorTransition[],
  transitionId: string,
  direction: "up" | "down",
): EditorTransition[] => {
  const target = transitions.find((transition) => transition.id === transitionId);
  if (!target) {
    return transitions;
  }

  const key = getTransitionGroupKey(target);
  const group = transitions
    .filter((transition) => getTransitionGroupKey(transition) === key)
    .sort((a, b) => a.order - b.order);

  const index = group.findIndex((transition) => transition.id === transitionId);
  if (index === -1) {
    return transitions;
  }

  const nextIndex = direction === "up" ? index - 1 : index + 1;
  if (nextIndex < 0 || nextIndex >= group.length) {
    return transitions;
  }

  const swapped = [...group];
  const tmp = swapped[index];
  swapped[index] = swapped[nextIndex] as EditorTransition;
  swapped[nextIndex] = tmp as EditorTransition;

  const remappedOrders = new Map<string, number>();
  swapped.forEach((transition, idx) => {
    remappedOrders.set(transition.id, idx);
  });

  return transitions.map((transition) => {
    if (getTransitionGroupKey(transition) !== key) {
      return transition;
    }

    return {
      ...transition,
      order: remappedOrders.get(transition.id) ?? transition.order,
    };
  });
};
