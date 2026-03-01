import { describe, expect, it } from "vitest";

import type { EditorNode, NodeSizeMap } from "../types";
import {
  computeCanvasSize,
  clampZoom,
  computeCenteredScroll,
  computeFitZoom,
  getContentBounds,
} from "../utils/viewport";

describe("viewport utils", () => {
  it("calcula bounds de contenido incluyendo universos y realidades", () => {
    const nodes: EditorNode[] = [
      {
        id: "univ-1",
        type: "universe",
        x: 100,
        y: 200,
        w: 300,
        h: 200,
        data: {
          id: "main-universe",
          name: "main-universe",
          canonicalName: "main-universe",
          version: "1.0.0",
        },
      },
      {
        id: "real-1",
        type: "reality",
        x: 500,
        y: 250,
        data: {
          id: "idle",
          name: "idle",
          universeId: "univ-1",
          isInitial: true,
          realityType: "normal",
        },
      },
    ];

    const nodeSizes: NodeSizeMap = {
      "real-1": { w: 180, h: 120 },
    };

    const bounds = getContentBounds(nodes, nodeSizes);

    expect(bounds).not.toBeNull();
    expect(bounds).toMatchObject({
      minX: 100,
      minY: 200,
      maxX: 680,
      maxY: 400,
      width: 580,
      height: 200,
      centerX: 390,
      centerY: 300,
    });
  });

  it("usa fallback de tamaño para realities sin medición", () => {
    const nodes: EditorNode[] = [
      {
        id: "real-1",
        type: "reality",
        x: 10,
        y: 20,
        data: {
          id: "idle",
          name: "idle",
          universeId: "univ-1",
          isInitial: false,
          realityType: "normal",
        },
      },
    ];

    const bounds = getContentBounds(nodes, {});

    expect(bounds).not.toBeNull();
    expect(bounds).toMatchObject({
      minX: 10,
      minY: 20,
      maxX: 202,
      maxY: 170,
    });
  });

  it("clampZoom respeta límites configurados", () => {
    expect(clampZoom(1.2, 0.2, 2)).toBe(1.2);
    expect(clampZoom(0.1, 0.2, 2)).toBe(0.2);
    expect(clampZoom(4, 0.2, 2)).toBe(2);
  });

  it("computeFitZoom calcula zoom de encuadre y lo acota", () => {
    const fit = computeFitZoom(
      {
        minX: 100,
        minY: 100,
        maxX: 700,
        maxY: 400,
        width: 600,
        height: 300,
        centerX: 400,
        centerY: 250,
      },
      1200,
      800,
      80,
      0.2,
      2,
    );

    expect(fit).toBeCloseTo(1.733, 3);
    expect(
      computeFitZoom(
        {
          minX: 0,
          minY: 0,
          maxX: 10,
          maxY: 10,
          width: 10,
          height: 10,
          centerX: 5,
          centerY: 5,
        },
        5000,
        5000,
        0,
        0.2,
        2,
      ),
    ).toBe(2);
  });

  it("computeCenteredScroll centra y limita scroll dentro del canvas", () => {
    expect(computeCenteredScroll(1600, 1600, 1, 1000, 800, 3200)).toEqual({
      left: 1100,
      top: 1200,
    });

    expect(computeCenteredScroll(50, 50, 1, 1000, 800, 3200)).toEqual({
      left: 0,
      top: 0,
    });

    expect(computeCenteredScroll(4000, 4000, 2, 1000, 800, 3200)).toEqual({
      left: 5400,
      top: 5600,
    });
  });

  it("computeCanvasSize extiende dimensiones según contenido y padding", () => {
    expect(computeCanvasSize(null, 3200, 600)).toEqual({
      width: 3200,
      height: 3200,
    });

    expect(
      computeCanvasSize(
        {
          minX: 100,
          minY: 100,
          maxX: 4100,
          maxY: 3500,
          width: 4000,
          height: 3400,
          centerX: 2100,
          centerY: 1800,
        },
        3200,
        600,
      ),
    ).toEqual({
      width: 4700,
      height: 4100,
    });
  });
});
