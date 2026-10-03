import { afterEach, describe, expect, it, vi } from "vitest";

import type { EditorNode } from "../types";

const nodes: EditorNode[] = [{
  id: "u-1", type: "universe", x: 12, y: 24, w: 400, h: 300,
  data: { id: "main", name: "main", canonicalName: "main", version: "1.0.0" },
}];

afterEach(() => {
  vi.doUnmock("elkjs/lib/elk.bundled.js");
  vi.resetModules();
});

describe("Carga diferida de ELK", () => {
  it("no carga ELK al importar ni al distribuir un grafo vacío", async () => {
    const load = vi.fn(() => ({ default: class {} }));
    vi.doMock("elkjs/lib/elk.bundled.js", load);
    const { computeAutoLayout } = await import("../utils/autoLayout");
    expect(load).not.toHaveBeenCalled();
    const empty: EditorNode[] = [];
    expect(await computeAutoLayout(empty, [], {})).toBe(empty);
    expect(load).not.toHaveBeenCalled();
  });

  it("comparte la instancia entre solicitudes concurrentes y reintenta una carga fallida", async () => {
    const construct = vi.fn();
    let fail = true;
    vi.doMock("elkjs/lib/elk.bundled.js", () => ({
      default: class {
        constructor() {
          construct();
          if (fail) throw new Error("ELK unavailable");
        }
        async layout(graph: { children?: Array<Record<string, unknown>> }) {
          return { ...graph, children: graph.children?.map(child => ({ ...child, x: 0, y: 0 })) };
        }
      },
    }));
    const { computeAutoLayout } = await import("../utils/autoLayout");
    await expect(computeAutoLayout(nodes, [], {})).rejects.toThrow("ELK unavailable");
    fail = false;
    const results = await Promise.all([
      computeAutoLayout(nodes, [], {}), computeAutoLayout(nodes, [], {}),
    ]);
    expect(construct).toHaveBeenCalledTimes(2);
    results.forEach(result => {
      expect(result[0]).toMatchObject({ x: 12, y: 24, w: 240, h: 220 });
    });
  });
});
