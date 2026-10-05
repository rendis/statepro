import { afterEach, describe, expect, it, vi } from "vitest";
import type { EditorNode } from "../types";

const nodes: EditorNode[] = [{
  id: "u-1", type: "universe", x: 12, y: 24, w: 400, h: 300,
  data: { id: "main", name: "main", canonicalName: "main", version: "1.0.0" },
}];

const prepare = async (pending = false, failConstructor = false) => {
  const workers: Array<EventTarget & { terminate: ReturnType<typeof vi.fn>; url: string }> = [];
  class FakeWorker extends EventTarget {
    terminate = vi.fn();
    constructor(public url: string) { super(); workers.push(this); }
  }
  vi.stubGlobal("Worker", FakeWorker);
  const bundled = vi.fn(() => { throw new Error("bundled ELK should not load"); });
  vi.doMock("elkjs/lib/elk.bundled.js", bundled);
  const construct = vi.fn();
  vi.doMock("elkjs/lib/elk-api.js", () => ({
    default: class {
      constructor(options: { workerFactory: () => unknown; algorithms: string[] }) {
        construct(options);
        options.workerFactory();
        if (failConstructor) throw new Error("constructor failed");
      }
      async layout(graph: { children?: Array<Record<string, unknown>> }) {
        if (pending) return new Promise(() => {});
        return { ...graph, children: graph.children?.map(child => ({ ...child, x: 0, y: 0 })) };
      }
    },
  }));
  const { computeAutoLayout } = await import("../utils/autoLayout");
  return { computeAutoLayout, workers, bundled, construct };
};

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.doUnmock("elkjs/lib/elk-api.js");
  vi.doUnmock("elkjs/lib/elk.bundled.js");
  vi.resetModules();
});

describe("Layout aislado en worker", () => {
  it("usa el worker solicitado sin cargar el motor completo en el hilo principal", async () => {
    const { computeAutoLayout, workers, bundled, construct } = await prepare();
    const empty: EditorNode[] = [];
    expect(await computeAutoLayout(empty, [], {}, "/elk-worker.js")).toBe(empty);
    expect(workers).toHaveLength(0);
    await computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    const [a, b] = await Promise.all([
      computeAutoLayout(nodes, [], {}, "/elk-worker.js"),
      computeAutoLayout(nodes, [], {}, "/elk-worker.js"),
    ]);
    expect(a[0]).toMatchObject({ x: 12, y: 24, w: 240, h: 220 });
    expect(b).toEqual(a);
    expect(workers).toHaveLength(3);
    workers.forEach(worker => {
      expect(worker.url).toBe("/elk-worker.js");
      expect(worker.terminate).toHaveBeenCalledOnce();
    });
    expect(construct).toHaveBeenCalledWith(expect.objectContaining({ algorithms: ["layered"] }));
    expect(bundled).not.toHaveBeenCalled();
  });

  it.each(["error", "messageerror"])("rechaza %s y libera el worker en vez de quedar esperando", async event => {
    const { computeAutoLayout, workers } = await prepare(true);
    const result = computeAutoLayout(nodes, [], {}, "/missing-worker.js");
    const rejection = expect(result).rejects.toThrow("layout worker failed");
    await vi.waitFor(() => expect(workers).toHaveLength(1));
    workers[0].dispatchEvent(new Event(event));
    await rejection;
    expect(workers[0].terminate).toHaveBeenCalledOnce();
  });

  it("interrumpe un worker bloqueado y permite una solicitud posterior", async () => {
    const { computeAutoLayout, workers } = await prepare(true);
    vi.useFakeTimers();
    const result = computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    const rejection = expect(result).rejects.toThrow("timed out");
    await vi.waitFor(() => expect(workers).toHaveLength(1));
    await vi.advanceTimersByTimeAsync(30_000);
    await rejection;
    expect(workers[0].terminate).toHaveBeenCalledOnce();
    vi.useRealTimers();
    vi.resetModules();
    const retry = await prepare();
    expect(await retry.computeAutoLayout(nodes, [], {}, "/elk-worker.js")).toHaveLength(1);
    expect(retry.workers[0].terminate).toHaveBeenCalledOnce();
  });

  it("libera el worker si falla la construccion del motor", async () => {
    const { computeAutoLayout, workers } = await prepare(false, true);
    await expect(computeAutoLayout(nodes, [], {}, "/elk-worker.js")).rejects.toThrow("constructor failed");
    expect(workers[0].terminate).toHaveBeenCalledOnce();
  });
});
