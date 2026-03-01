import type { EditorNode, NodeSizeMap } from "../types";

const DEFAULT_REALITY_WIDTH = 192;
const DEFAULT_REALITY_HEIGHT = 150;
const DEFAULT_NOTE_WIDTH = 224;
const DEFAULT_NOTE_HEIGHT = 160;

export interface ContentBounds {
  minX: number;
  minY: number;
  maxX: number;
  maxY: number;
  width: number;
  height: number;
  centerX: number;
  centerY: number;
}

export interface CanvasSize {
  width: number;
  height: number;
}

const clamp = (value: number, min: number, max: number): number => {
  return Math.min(max, Math.max(min, value));
};

export const getContentBounds = (
  nodes: EditorNode[],
  nodeSizes: NodeSizeMap,
): ContentBounds | null => {
  if (nodes.length === 0) {
    return null;
  }

  let minX = Number.POSITIVE_INFINITY;
  let minY = Number.POSITIVE_INFINITY;
  let maxX = Number.NEGATIVE_INFINITY;
  let maxY = Number.NEGATIVE_INFINITY;

  nodes.forEach((node) => {
    const width =
      node.type === "universe"
        ? node.w
        : node.type === "note"
          ? node.data.isCollapsed
            ? 224
            : DEFAULT_NOTE_WIDTH
          : nodeSizes[node.id]?.w || DEFAULT_REALITY_WIDTH;
    const height =
      node.type === "universe"
        ? node.h
        : node.type === "note"
          ? node.data.isCollapsed
            ? 44
            : DEFAULT_NOTE_HEIGHT
          : nodeSizes[node.id]?.h || DEFAULT_REALITY_HEIGHT;

    minX = Math.min(minX, node.x);
    minY = Math.min(minY, node.y);
    maxX = Math.max(maxX, node.x + width);
    maxY = Math.max(maxY, node.y + height);
  });

  if (!Number.isFinite(minX) || !Number.isFinite(minY) || !Number.isFinite(maxX) || !Number.isFinite(maxY)) {
    return null;
  }

  const width = Math.max(maxX - minX, 1);
  const height = Math.max(maxY - minY, 1);

  return {
    minX,
    minY,
    maxX,
    maxY,
    width,
    height,
    centerX: minX + width / 2,
    centerY: minY + height / 2,
  };
};

export const clampZoom = (zoom: number, minZoom = 0.2, maxZoom = 2): number => {
  return clamp(zoom, minZoom, maxZoom);
};

export const computeFitZoom = (
  bounds: ContentBounds,
  viewportWidth: number,
  viewportHeight: number,
  padding = 80,
  minZoom = 0.2,
  maxZoom = 2,
): number => {
  if (viewportWidth <= 0 || viewportHeight <= 0) {
    return clampZoom(1, minZoom, maxZoom);
  }

  const availableWidth = Math.max(viewportWidth - padding * 2, 1);
  const availableHeight = Math.max(viewportHeight - padding * 2, 1);
  const zoomByWidth = availableWidth / Math.max(bounds.width, 1);
  const zoomByHeight = availableHeight / Math.max(bounds.height, 1);
  const fitZoom = Math.min(zoomByWidth, zoomByHeight);

  if (!Number.isFinite(fitZoom) || fitZoom <= 0) {
    return clampZoom(1, minZoom, maxZoom);
  }

  return clampZoom(fitZoom, minZoom, maxZoom);
};

export const computeCenteredScroll = (
  centerX: number,
  centerY: number,
  zoom: number,
  viewportWidth: number,
  viewportHeight: number,
  canvasSize: number,
): { left: number; top: number } => {
  const nextLeft = centerX * zoom - viewportWidth / 2;
  const nextTop = centerY * zoom - viewportHeight / 2;

  const maxLeft = Math.max(canvasSize * zoom - viewportWidth, 0);
  const maxTop = Math.max(canvasSize * zoom - viewportHeight, 0);

  return {
    left: clamp(nextLeft, 0, maxLeft),
    top: clamp(nextTop, 0, maxTop),
  };
};

export const computeCanvasSize = (
  bounds: ContentBounds | null,
  minSize = 3200,
  edgePadding = 600,
): CanvasSize => {
  if (!bounds) {
    return {
      width: minSize,
      height: minSize,
    };
  }

  return {
    width: Math.max(minSize, Math.ceil(bounds.maxX + edgePadding)),
    height: Math.max(minSize, Math.ceil(bounds.maxY + edgePadding)),
  };
};
