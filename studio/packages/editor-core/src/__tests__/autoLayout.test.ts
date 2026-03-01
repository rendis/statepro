import { describe, expect, it } from "vitest";

import type { EditorNode, EditorTransition, NodeSizeMap } from "../types";
import { computeAutoLayout } from "../utils/autoLayout";

const createUniverse = (
  id: string,
  dataId: string,
  x: number,
  y: number,
): Extract<EditorNode, { type: "universe" }> => ({
  id,
  type: "universe",
  x,
  y,
  w: 400,
  h: 300,
  data: {
    id: dataId,
    name: dataId,
    canonicalName: dataId,
    version: "1.0.0",
  },
});

const createReality = (
  id: string,
  dataId: string,
  universeId: string,
  x: number,
  y: number,
): Extract<EditorNode, { type: "reality" }> => ({
  id,
  type: "reality",
  x,
  y,
  data: {
    id: dataId,
    name: dataId,
    universeId,
    isInitial: false,
    realityType: "normal",
  },
});

const createTransition = (
  id: string,
  sourceRealityId: string,
  targets: string[],
): EditorTransition => ({
  id,
  sourceRealityId,
  triggerKind: "on",
  eventName: "GO",
  type: "default",
  condition: undefined,
  conditions: [],
  actions: [],
  invokes: [],
  description: "",
  metadata: "",
  targets,
  order: 0,
});

describe("computeAutoLayout", () => {
  it("ordena realidades izquierda a derecha dentro del universo", async () => {
    const nodes: EditorNode[] = [
      createUniverse("u-node-1", "u-1", 1000, 1000),
      createReality("r-node-1", "r-1", "u-node-1", 1200, 1260),
      createReality("r-node-2", "r-2", "u-node-1", 1080, 1120),
      createReality("r-node-3", "r-3", "u-node-1", 1320, 1400),
    ];
    const transitions: EditorTransition[] = [
      createTransition("t-1", "r-node-1", ["r-2"]),
      createTransition("t-2", "r-node-2", ["r-3"]),
    ];
    const nodeSizes: NodeSizeMap = {
      "r-node-1": { w: 180, h: 120 },
      "r-node-2": { w: 180, h: 120 },
      "r-node-3": { w: 180, h: 120 },
    };

    const result = await computeAutoLayout(nodes, transitions, nodeSizes);
    const r1 = result.find(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.id === "r-node-1" && node.type === "reality",
    );
    const r2 = result.find(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.id === "r-node-2" && node.type === "reality",
    );
    const r3 = result.find(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.id === "r-node-3" && node.type === "reality",
    );
    const universe = result.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.id === "u-node-1" && node.type === "universe",
    );

    expect(r1).toBeDefined();
    expect(r2).toBeDefined();
    expect(r3).toBeDefined();
    expect(universe).toBeDefined();

    if (!r1 || !r2 || !r3 || !universe) {
      return;
    }

    expect(r1.x).toBeLessThan(r2.x);
    expect(r2.x).toBeLessThan(r3.x);
    expect(r1.x).toBeGreaterThanOrEqual(universe.x);
    expect(r3.x + (nodeSizes["r-node-3"]?.w || 0)).toBeLessThanOrEqual(universe.x + universe.w);
    expect(r2.x - (r1.x + (nodeSizes["r-node-1"]?.w || 0))).toBeGreaterThanOrEqual(280);
  });

  it("centra verticalmente un nodo de inicio unico dentro del universo", async () => {
    const nodes: EditorNode[] = [
      createUniverse("u-start-single", "u-start-single", 1000, 1000),
      createReality("r-start", "r-start", "u-start-single", 1100, 1040),
      createReality("r-a", "r-a", "u-start-single", 1380, 980),
      createReality("r-b", "r-b", "u-start-single", 1400, 1180),
      createReality("r-c", "r-c", "u-start-single", 1410, 1360),
      createReality("r-d", "r-d", "u-start-single", 1680, 1200),
    ];
    const transitions: EditorTransition[] = [
      createTransition("t-start-a", "r-start", ["r-a"]),
      createTransition("t-start-b", "r-start", ["r-b"]),
      createTransition("t-start-c", "r-start", ["r-c"]),
      createTransition("t-a-d", "r-a", ["r-d"]),
      createTransition("t-b-d", "r-b", ["r-d"]),
      createTransition("t-c-d", "r-c", ["r-d"]),
    ];
    const nodeSizes: NodeSizeMap = {
      "r-start": { w: 180, h: 120 },
      "r-a": { w: 180, h: 120 },
      "r-b": { w: 180, h: 120 },
      "r-c": { w: 180, h: 120 },
      "r-d": { w: 180, h: 120 },
    };

    const result = await computeAutoLayout(nodes, transitions, nodeSizes);

    const universe = result.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.id === "u-start-single" && node.type === "universe",
    );
    const startReality = result.find(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.id === "r-start" && node.type === "reality",
    );
    const laidOutRealities = result.filter(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.type === "reality" && node.data.universeId === "u-start-single",
    );

    expect(universe).toBeDefined();
    expect(startReality).toBeDefined();
    if (!universe || !startReality) {
      return;
    }

    const minRealityX = Math.min(...laidOutRealities.map((reality) => reality.x));
    expect(startReality.x).toBeCloseTo(minRealityX, 5);

    const startCenterY = startReality.y + (nodeSizes["r-start"]?.h || 0) / 2;
    const universeCenterY = universe.y + universe.h / 2;
    expect(Math.abs(startCenterY - universeCenterY)).toBeLessThanOrEqual(40);
  });

  it("alinea y equidista multiples nodos de inicio cerca del centro vertical", async () => {
    const nodes: EditorNode[] = [
      createUniverse("u-start-multi", "u-start-multi", 1000, 1000),
      createReality("r-start-a", "r-start-a", "u-start-multi", 1060, 920),
      createReality("r-start-b", "r-start-b", "u-start-multi", 1120, 1180),
      createReality("r-start-c", "r-start-c", "u-start-multi", 1180, 1450),
      createReality("r-target", "r-target", "u-start-multi", 1560, 1200),
      createReality("r-end", "r-end", "u-start-multi", 1860, 1200),
    ];
    const transitions: EditorTransition[] = [
      createTransition("t-a-target", "r-start-a", ["r-target"]),
      createTransition("t-b-target", "r-start-b", ["r-target"]),
      createTransition("t-c-target", "r-start-c", ["r-target"]),
      createTransition("t-target-end", "r-target", ["r-end"]),
    ];
    const nodeSizes: NodeSizeMap = {
      "r-start-a": { w: 180, h: 120 },
      "r-start-b": { w: 180, h: 120 },
      "r-start-c": { w: 180, h: 120 },
      "r-target": { w: 180, h: 120 },
      "r-end": { w: 180, h: 120 },
    };

    const result = await computeAutoLayout(nodes, transitions, nodeSizes);
    const universe = result.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.id === "u-start-multi" && node.type === "universe",
    );
    const starts = result
      .filter(
        (node): node is Extract<EditorNode, { type: "reality" }> =>
          node.type === "reality" &&
          node.data.universeId === "u-start-multi" &&
          node.id.startsWith("r-start-"),
      )
      .sort((left, right) => left.y - right.y);

    expect(universe).toBeDefined();
    expect(starts).toHaveLength(3);
    if (!universe || starts.length !== 3) {
      return;
    }

    expect(starts[0]?.x).toBeCloseTo(starts[1]?.x || 0, 5);
    expect(starts[1]?.x).toBeCloseTo(starts[2]?.x || 0, 5);

    const expectedStep = (nodeSizes["r-start-a"]?.h || 0) + 110;
    const firstStep = (starts[1]?.y || 0) - (starts[0]?.y || 0);
    const secondStep = (starts[2]?.y || 0) - (starts[1]?.y || 0);
    expect(firstStep).toBeCloseTo(expectedStep, 5);
    expect(secondStep).toBeCloseTo(expectedStep, 5);

    const startsTop = starts[0]?.y || 0;
    const startsBottom = (starts[2]?.y || 0) + (nodeSizes["r-start-c"]?.h || 0);
    const startsCenter = (startsTop + startsBottom) / 2;
    const universeCenterY = universe.y + universe.h / 2;
    expect(Math.abs(startsCenter - universeCenterY)).toBeLessThanOrEqual(40);
  });

  it("ordena universos izquierda a derecha segun transiciones cruzadas", async () => {
    const nodes: EditorNode[] = [
      createUniverse("u-node-1", "u-1", 900, 900),
      createUniverse("u-node-2", "u-2", 1600, 900),
      createReality("r-node-1", "r-1", "u-node-1", 950, 980),
      createReality("r-node-2", "r-2", "u-node-2", 1650, 980),
    ];
    const transitions: EditorTransition[] = [
      createTransition("t-cross", "r-node-1", ["U:u-2:r-2"]),
    ];

    const result = await computeAutoLayout(nodes, transitions, {});
    const u1 = result.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.id === "u-node-1" && node.type === "universe",
    );
    const u2 = result.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.id === "u-node-2" && node.type === "universe",
    );

    expect(u1).toBeDefined();
    expect(u2).toBeDefined();

    if (!u1 || !u2) {
      return;
    }

    expect(u1.x).toBeLessThan(u2.x);
  });

  it("mantiene tamano minimo para universo vacio", async () => {
    const nodes: EditorNode[] = [createUniverse("u-empty", "u-empty", 1000, 1000)];
    const result = await computeAutoLayout(nodes, [], {});
    const universe = result.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.id === "u-empty" && node.type === "universe",
    );

    expect(universe).toBeDefined();
    if (!universe) {
      return;
    }

    expect(universe.w).toBe(240);
    expect(universe.h).toBe(220);
  });

  it("ignora targets invalidos sin lanzar errores", async () => {
    const nodes: EditorNode[] = [
      createUniverse("u-node-1", "u-1", 1000, 1000),
      createReality("r-node-1", "r-1", "u-node-1", 1100, 1100),
    ];
    const transitions: EditorTransition[] = [
      createTransition("t-invalid", "r-node-1", ["INVALID::TARGET"]),
    ];

    await expect(computeAutoLayout(nodes, transitions, {})).resolves.toHaveLength(nodes.length);
  });

  it("mantiene layout valido cuando no hay nodos de inicio (ciclo puro)", async () => {
    const nodes: EditorNode[] = [
      createUniverse("u-cycle", "u-cycle", 1000, 1000),
      createReality("r-cycle-a", "r-cycle-a", "u-cycle", 1110, 1080),
      createReality("r-cycle-b", "r-cycle-b", "u-cycle", 1360, 1160),
      createReality("r-cycle-c", "r-cycle-c", "u-cycle", 1600, 1120),
    ];
    const transitions: EditorTransition[] = [
      createTransition("t-cycle-a-b", "r-cycle-a", ["r-cycle-b"]),
      createTransition("t-cycle-b-c", "r-cycle-b", ["r-cycle-c"]),
      createTransition("t-cycle-c-a", "r-cycle-c", ["r-cycle-a"]),
    ];

    const result = await computeAutoLayout(nodes, transitions, {});
    const cycleRealities = result.filter(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.type === "reality" && node.data.universeId === "u-cycle",
    );
    const universe = result.find(
      (node): node is Extract<EditorNode, { type: "universe" }> =>
        node.id === "u-cycle" && node.type === "universe",
    );

    expect(result).toHaveLength(nodes.length);
    expect(universe).toBeDefined();
    expect(cycleRealities).toHaveLength(3);
    cycleRealities.forEach((reality) => {
      expect(Number.isFinite(reality.x)).toBe(true);
      expect(Number.isFinite(reality.y)).toBe(true);
    });
  });

  it("mantiene orden estable para universos desconectados", async () => {
    const nodes: EditorNode[] = [
      createUniverse("u-a", "u-a", 900, 1000),
      createUniverse("u-b", "u-b", 1500, 1000),
      createUniverse("u-c", "u-c", 1200, 1000),
      createReality("r-a", "r-a", "u-a", 940, 1080),
      createReality("r-b", "r-b", "u-b", 1540, 1080),
      createReality("r-c", "r-c", "u-c", 1240, 1080),
    ];

    const initialOrder = nodes
      .filter((node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe")
      .sort((left, right) => left.x - right.x)
      .map((node) => node.id);

    const result = await computeAutoLayout(nodes, [], {});
    const nextOrder = result
      .filter((node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe")
      .sort((left, right) => left.x - right.x)
      .map((node) => node.id);

    expect(nextOrder).toEqual(initialOrder);
  });
});
