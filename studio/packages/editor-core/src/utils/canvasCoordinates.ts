export interface ViewportPoint {
  x: number;
  y: number;
}

export interface CanvasPoint {
  x: number;
  y: number;
}

export interface CanvasViewportRect {
  left: number;
  top: number;
}

export interface CanvasTransformState {
  scale: number;
  positionX: number;
  positionY: number;
}

export const viewportPointToCanvasPoint = (
  point: ViewportPoint,
  viewportRect: CanvasViewportRect | null | undefined,
  transform: CanvasTransformState | null | undefined,
): CanvasPoint | null => {
  if (!viewportRect || !transform) {
    return null;
  }

  if (!Number.isFinite(transform.scale) || transform.scale === 0) {
    return null;
  }

  return {
    x: (point.x - viewportRect.left - transform.positionX) / transform.scale,
    y: (point.y - viewportRect.top - transform.positionY) / transform.scale,
  };
};
