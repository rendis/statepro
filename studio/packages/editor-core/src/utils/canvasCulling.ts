import type { EditorNode, EditorTransition, NodeSizeMap, TransitionLeg } from "../types";

const DEFAULT_REALITY_WIDTH = 192;
const DEFAULT_REALITY_HEIGHT = 150;
const DEFAULT_NOTE_WIDTH = 224;
const DEFAULT_NOTE_HEIGHT = 160;
const COLLAPSED_NOTE_HEIGHT = 44;
const MIN_BOUNDS_SIZE = 4;

export interface CanvasViewportBounds {
  left: number;
  top: number;
  right: number;
  bottom: number;
}

type TransformStateLike = {
  scale: number;
  positionX: number;
  positionY: number;
} | null;

type ViewportRectLike = {
  width: number;
  height: number;
} | null;

export const computeCanvasViewportBounds = (
  viewport: ViewportRectLike,
  transform: TransformStateLike,
  padding = 240,
): CanvasViewportBounds | null => {
  if (
    !viewport ||
    !transform ||
    !Number.isFinite(transform.scale) ||
    transform.scale <= 0 ||
    !Number.isFinite(viewport.width) ||
    !Number.isFinite(viewport.height) ||
    viewport.width <= 0 ||
    viewport.height <= 0
  ) {
    return null;
  }

  const left = (-transform.positionX) / transform.scale - padding;
  const top = (-transform.positionY) / transform.scale - padding;
  const right = (viewport.width - transform.positionX) / transform.scale + padding;
  const bottom = (viewport.height - transform.positionY) / transform.scale + padding;

  if (![left, top, right, bottom].every(Number.isFinite)) {
    return null;
  }

  if (right - left < MIN_BOUNDS_SIZE || bottom - top < MIN_BOUNDS_SIZE) {
    return null;
  }

  return { left, top, right, bottom };
};

const getNodeRect = (
  node: EditorNode,
  nodeSizes: NodeSizeMap,
): { left: number; top: number; right: number; bottom: number } => {
  if (node.type === "universe") {
    return {
      left: node.x,
      top: node.y,
      right: node.x + node.w,
      bottom: node.y + node.h,
    };
  }

  if (node.type === "note") {
    const noteWidth = DEFAULT_NOTE_WIDTH;
    const noteHeight = node.data.isCollapsed ? COLLAPSED_NOTE_HEIGHT : DEFAULT_NOTE_HEIGHT;
    return {
      left: node.x,
      top: node.y,
      right: node.x + noteWidth,
      bottom: node.y + noteHeight,
    };
  }

  const size = nodeSizes[node.id] || { w: DEFAULT_REALITY_WIDTH, h: DEFAULT_REALITY_HEIGHT };
  return {
    left: node.x,
    top: node.y,
    right: node.x + size.w,
    bottom: node.y + size.h,
  };
};

const intersects = (
  bounds: CanvasViewportBounds,
  rect: { left: number; top: number; right: number; bottom: number },
): boolean => {
  return !(
    rect.right < bounds.left ||
    rect.left > bounds.right ||
    rect.bottom < bounds.top ||
    rect.top > bounds.bottom
  );
};

export interface CanvasCullOptions {
  nodes: EditorNode[];
  transitions: EditorTransition[];
  nodeSizes: NodeSizeMap;
  transitionLegsByTransitionId: Map<string, TransitionLeg[]>;
  viewportBounds: CanvasViewportBounds | null;
  alwaysVisibleNodeIds?: Iterable<string>;
  alwaysVisibleTransitionIds?: Iterable<string>;
}

export interface CanvasCullResult {
  visibleNodeIds: Set<string>;
  visibleTransitionIds: Set<string>;
}

export const buildCanvasCullResult = ({
  nodes,
  transitions,
  nodeSizes,
  transitionLegsByTransitionId,
  viewportBounds,
  alwaysVisibleNodeIds = [],
  alwaysVisibleTransitionIds = [],
}: CanvasCullOptions): CanvasCullResult => {
  const allVisibleResult: CanvasCullResult = {
    visibleNodeIds: new Set(nodes.map((node) => node.id)),
    visibleTransitionIds: new Set(transitions.map((transition) => transition.id)),
  };

  if (!viewportBounds) {
    return allVisibleResult;
  }

  const visibleNodeIds = new Set<string>();
  const visibleTransitionIds = new Set<string>();

  nodes.forEach((node) => {
    if (intersects(viewportBounds, getNodeRect(node, nodeSizes))) {
      visibleNodeIds.add(node.id);
    }
  });

  for (const forcedNodeId of alwaysVisibleNodeIds) {
    visibleNodeIds.add(forcedNodeId);
  }

  transitions.forEach((transition) => {
    const sourceVisible = visibleNodeIds.has(transition.sourceRealityId);
    const legs = transitionLegsByTransitionId.get(transition.id) || [];
    const hasVisibleTarget = legs.some((leg) => visibleNodeIds.has(leg.target));
    if (sourceVisible || hasVisibleTarget) {
      visibleTransitionIds.add(transition.id);
    }
  });

  for (const forcedTransitionId of alwaysVisibleTransitionIds) {
    visibleTransitionIds.add(forcedTransitionId);
    const transition = transitions.find((candidate) => candidate.id === forcedTransitionId);
    if (!transition) {
      continue;
    }

    visibleNodeIds.add(transition.sourceRealityId);
    const legs = transitionLegsByTransitionId.get(transition.id) || [];
    legs.forEach((leg) => visibleNodeIds.add(leg.target));
  }

  // Fail-open: never let aggressive culling hide the entire scene.
  if (nodes.length > 0 && visibleNodeIds.size === 0) {
    return allVisibleResult;
  }

  return {
    visibleNodeIds,
    visibleTransitionIds,
  };
};
