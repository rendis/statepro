import { describe, expect, it } from "vitest";

import { viewportPointToCanvasPoint } from "../utils";

describe("canvasCoordinates", () => {
  it("convierte viewport a canvas con escala 1 y offset", () => {
    const point = viewportPointToCanvasPoint(
      { x: 260, y: 180 },
      { left: 40, top: 20 },
      { scale: 1, positionX: 100, positionY: 50 },
    );

    expect(point).toEqual({ x: 120, y: 110 });
  });

  it("convierte viewport a canvas con escala distinta de 1", () => {
    const point = viewportPointToCanvasPoint(
      { x: 900, y: 700 },
      { left: 100, top: 80 },
      { scale: 2, positionX: -200, positionY: 40 },
    );

    expect(point).toEqual({ x: 500, y: 290 });
  });

  it("soporta puntos negativos y borde", () => {
    const point = viewportPointToCanvasPoint(
      { x: -50, y: 0 },
      { left: 10, top: 10 },
      { scale: 0.5, positionX: 20, positionY: -30 },
    );

    expect(point).toEqual({ x: -160, y: 40 });
  });

  it("es idempotente en mediciones repetidas con transform fijo", () => {
    const viewportRect = { left: 24, top: 12 };
    const transform = { scale: 1.25, positionX: 140, positionY: -32 };
    const measuredViewportPoint = { x: 640, y: 360 };

    const results = Array.from({ length: 5 }, () =>
      viewportPointToCanvasPoint(measuredViewportPoint, viewportRect, transform),
    );
    const baseline = results[0];
    expect(baseline).toBeTruthy();

    results.forEach((result) => {
      expect(result).toEqual(baseline);
    });
  });

  it("retorna null cuando no hay transform valido", () => {
    const point = viewportPointToCanvasPoint(
      { x: 100, y: 100 },
      { left: 0, top: 0 },
      { scale: 0, positionX: 0, positionY: 0 },
    );

    expect(point).toBeNull();
  });
});
