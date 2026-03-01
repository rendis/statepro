import { describe, expect, it } from "vitest";

import type { EditorNode } from "../types";
import { searchCanvasNodes, type CanvasSearchFilters } from "../utils";

const buildNodes = (): EditorNode[] => [
  {
    id: "univ-1",
    type: "universe",
    x: 100,
    y: 100,
    w: 500,
    h: 320,
    data: {
      id: "main-universe",
      name: "main-universe",
      canonicalName: "admission-main",
      version: "1.0.0",
      tags: ["payments", "vip-core"],
    },
  },
  {
    id: "univ-2",
    type: "universe",
    x: 800,
    y: 100,
    w: 500,
    h: 320,
    data: {
      id: "alpha",
      name: "alpha",
      canonicalName: "alpha-canonical",
      version: "1.0.0",
      tags: [],
    },
  },
  {
    id: "univ-3",
    type: "universe",
    x: 1450,
    y: 100,
    w: 500,
    h: 320,
    data: {
      id: "alphabet",
      name: "alphabet",
      canonicalName: "alphabet-canonical",
      version: "1.0.0",
      tags: [],
    },
  },
  {
    id: "real-1",
    type: "reality",
    x: 150,
    y: 190,
    data: {
      id: "idle",
      name: "idle",
      universeId: "univ-1",
      isInitial: true,
      realityType: "normal",
    },
  },
  {
    id: "real-2",
    type: "reality",
    x: 390,
    y: 190,
    data: {
      id: "proc",
      name: "processing",
      universeId: "univ-1",
      isInitial: false,
      realityType: "success",
    },
  },
];

describe("searchCanvasNodes", () => {
  const filters = (overrides: Partial<CanvasSearchFilters>): CanvasSearchFilters => ({
    universe: true,
    tag: true,
    reality: true,
    ...overrides,
  });

  it("retorna vacío cuando la query está vacía", () => {
    expect(searchCanvasNodes(buildNodes(), "  ")).toEqual([]);
  });

  it("encuentra universos por id", () => {
    const results = searchCanvasNodes(buildNodes(), "main-universe");
    expect(results).toHaveLength(1);
    expect(results[0]).toMatchObject({
      nodeId: "univ-1",
      nodeType: "universe",
      matchedField: "id",
    });
  });

  it("encuentra universos por canonicalName", () => {
    const results = searchCanvasNodes(buildNodes(), "admission");
    expect(results).toHaveLength(1);
    expect(results[0]).toMatchObject({
      nodeId: "univ-1",
      nodeType: "universe",
      matchedField: "canonicalName",
    });
  });

  it("encuentra universos por contención de tags", () => {
    const results = searchCanvasNodes(buildNodes(), "vip");
    expect(results).toHaveLength(1);
    expect(results[0]).toMatchObject({
      nodeId: "univ-1",
      nodeType: "universe",
      matchedField: "tag",
      matchedTag: "vip-core",
    });
  });

  it("encuentra realidades por id y name", () => {
    const byId = searchCanvasNodes(buildNodes(), "proc");
    expect(byId).toHaveLength(1);
    expect(byId[0]).toMatchObject({
      nodeId: "real-2",
      nodeType: "reality",
      matchedField: "id",
    });

    const byName = searchCanvasNodes(buildNodes(), "process");
    expect(byName).toHaveLength(1);
    expect(byName[0]).toMatchObject({
      nodeId: "real-2",
      nodeType: "reality",
      matchedField: "name",
    });
  });

  it("ordena por ranking exact > prefix > contains", () => {
    const nodes: EditorNode[] = [
      ...buildNodes(),
      {
        id: "univ-4",
        type: "universe",
        x: 2000,
        y: 100,
        w: 500,
        h: 320,
        data: {
          id: "my-alpha",
          name: "my-alpha",
          canonicalName: "my-alpha",
          version: "1.0.0",
          tags: [],
        },
      },
    ];

    const results = searchCanvasNodes(nodes, "alpha");
    expect(results.map((entry) => entry.nodeId).slice(0, 3)).toEqual(["univ-2", "univ-3", "univ-4"]);
  });

  it("deduplica por nodo y conserva el mejor match", () => {
    const results = searchCanvasNodes(buildNodes(), "main");
    expect(results).toHaveLength(1);
    expect(results[0]).toMatchObject({
      nodeId: "univ-1",
      matchedField: "id",
    });
  });

  it("aplica scope universe sin incluir tags ni realidades", () => {
    const onlyUniverse = searchCanvasNodes(buildNodes(), "admission", {
      filters: filters({ tag: false, reality: false }),
    });
    expect(onlyUniverse).toHaveLength(1);
    expect(onlyUniverse[0]).toMatchObject({
      nodeId: "univ-1",
      nodeType: "universe",
      matchedField: "canonicalName",
    });

    const noTagMatch = searchCanvasNodes(buildNodes(), "vip", {
      filters: filters({ tag: false, reality: false }),
    });
    expect(noTagMatch).toEqual([]);
  });

  it("aplica scope tag de forma independiente", () => {
    const results = searchCanvasNodes(buildNodes(), "vip", {
      filters: filters({ universe: false, reality: false, tag: true }),
    });

    expect(results).toHaveLength(1);
    expect(results[0]).toMatchObject({
      nodeId: "univ-1",
      matchedField: "tag",
    });
  });

  it("aplica scope reality sin incluir universos", () => {
    const results = searchCanvasNodes(buildNodes(), "process", {
      filters: filters({ universe: false, tag: false, reality: true }),
    });

    expect(results).toHaveLength(1);
    expect(results[0]).toMatchObject({
      nodeId: "real-2",
      nodeType: "reality",
      matchedField: "name",
    });
  });

  it("retorna vacío cuando todos los filtros están apagados", () => {
    const results = searchCanvasNodes(buildNodes(), "main", {
      filters: {
        universe: false,
        tag: false,
        reality: false,
      },
    });
    expect(results).toEqual([]);
  });
});
