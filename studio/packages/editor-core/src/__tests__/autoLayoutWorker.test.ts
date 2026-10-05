import { afterEach, describe, expect, it, vi } from "vitest";
import type { EditorNode } from "../types";

const nodes: EditorNode[] = [{
  id: "u-1", type: "universe", x: 12, y: 24, w: 400, h: 300,
  data: { id: "main", name: "main", canonicalName: "main", version: "1.0.0" },
}];

type FakeWorker = EventTarget & { terminate: ReturnType<typeof vi.fn>; url: string };

type LayoutMode = "ok" | "pending" | "elk-error";

interface PrepareOptions {
  mode?: LayoutMode;
  failConstructor?: boolean;
  workerThrows?: boolean;
}

const layoutGraph = (graph: { children?: Array<Record<string, unknown>> }) =>
  ({ ...graph, children: graph.children?.map(child => ({ ...child, x: 0, y: 0 })) });

const prepare = async ({ mode = "ok", failConstructor = false, workerThrows = false }: PrepareOptions = {}) => {
  const workers: FakeWorker[] = [];
  const control = { mode };
  class FakeWorkerImpl extends EventTarget {
    terminate = vi.fn();
    constructor(public url: string) {
      super();
      if (workerThrows) throw new DOMException("cross-origin worker", "SecurityError");
      workers.push(this as unknown as FakeWorker);
    }
  }
  vi.stubGlobal("Worker", FakeWorkerImpl);
  const bundledLayout = vi.fn(async (graph: { children?: Array<Record<string, unknown>> }) => layoutGraph(graph));
  vi.doMock("elkjs/lib/elk.bundled.js", () => ({
    default: class { layout = bundledLayout; },
  }));
  const construct = vi.fn();
  vi.doMock("elkjs/lib/elk-api.js", () => ({
    default: class {
      constructor(options: { workerFactory: () => unknown; algorithms: string[] }) {
        construct(options);
        options.workerFactory();
        if (failConstructor) throw new Error("constructor failed");
      }
      async layout(graph: { children?: Array<Record<string, unknown>> }) {
        if (control.mode === "pending") return new Promise(() => {});
        if (control.mode === "elk-error") throw new Error("invalid graph");
        return layoutGraph(graph);
      }
    },
  }));
  const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
  const { computeAutoLayout } = await import("../utils/autoLayout");
  return { computeAutoLayout, workers, bundledLayout, construct, control, warn };
};

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  vi.doUnmock("elkjs/lib/elk-api.js");
  vi.doUnmock("elkjs/lib/elk.bundled.js");
  vi.resetModules();
});

describe("Layout en worker reutilizable", () => {
  it("reutiliza un solo worker para layouts secuenciales y concurrentes sin cargar el motor completo", async () => {
    const { computeAutoLayout, workers, bundledLayout, construct } = await prepare();
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
    expect(workers).toHaveLength(1);
    expect(workers[0].url).toBe("/elk-worker.js");
    expect(workers[0].terminate).not.toHaveBeenCalled();
    expect(construct).toHaveBeenCalledOnce();
    expect(construct).toHaveBeenCalledWith(expect.objectContaining({ algorithms: ["layered"] }));
    expect(bundledLayout).not.toHaveBeenCalled();
  });

  it.each(["error", "messageerror"])(
    "si el worker falla al cargar (%s) usa el motor del hilo principal y no reintenta esa URL",
    async event => {
      const { computeAutoLayout, workers, bundledLayout, control, warn } = await prepare({ mode: "pending" });
      const result = computeAutoLayout(nodes, [], {}, "/missing-worker.js");
      await vi.waitFor(() => expect(workers).toHaveLength(1));
      workers[0].dispatchEvent(new Event(event));
      expect((await result)[0]).toMatchObject({ x: 12, y: 24 });
      expect(workers[0].terminate).toHaveBeenCalledOnce();
      expect(bundledLayout).toHaveBeenCalledOnce();
      expect(warn).toHaveBeenCalled();

      control.mode = "ok";
      await computeAutoLayout(nodes, [], {}, "/missing-worker.js");
      expect(workers).toHaveLength(1);
      expect(bundledLayout).toHaveBeenCalledTimes(2);
    },
  );

  it.each([
    ["el constructor de Worker lanza (CSP u otro origen)", { workerThrows: true }],
    ["falla la construccion del motor", { failConstructor: true }],
  ])("usa el motor del hilo principal si %s", async (_label, options) => {
    const { computeAutoLayout, workers, bundledLayout } = await prepare(options);
    expect(await computeAutoLayout(nodes, [], {}, "/elk-worker.js")).toHaveLength(1);
    expect(bundledLayout).toHaveBeenCalledOnce();
    workers.forEach(worker => expect(worker.terminate).toHaveBeenCalledOnce());
  });

  it("usa el motor del hilo principal cuando Worker no existe", async () => {
    const { computeAutoLayout, bundledLayout } = await prepare();
    vi.stubGlobal("Worker", undefined);
    expect(await computeAutoLayout(nodes, [], {}, "/elk-worker.js")).toHaveLength(1);
    expect(bundledLayout).toHaveBeenCalledOnce();
  });

  it("si el worker cae tras funcionar, recupera la solicitud en curso y crea uno nuevo despues", async () => {
    const { computeAutoLayout, workers, bundledLayout, control } = await prepare();
    await computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    control.mode = "pending";
    const result = computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    workers[0].dispatchEvent(new Event("error"));
    expect(await result).toHaveLength(1);
    expect(bundledLayout).toHaveBeenCalledOnce();

    control.mode = "ok";
    await computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    expect(workers).toHaveLength(2);
    expect(bundledLayout).toHaveBeenCalledOnce();
  });

  it("interrumpe un worker bloqueado sin congelar el hilo principal y el siguiente layout usa uno nuevo", async () => {
    const { computeAutoLayout, workers, bundledLayout, control } = await prepare({ mode: "pending" });
    vi.useFakeTimers();
    const result = computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    const rejection = expect(result).rejects.toThrow("timed out");
    await vi.waitFor(() => expect(workers).toHaveLength(1));
    await vi.advanceTimersByTimeAsync(30_000);
    await rejection;
    expect(workers[0].terminate).toHaveBeenCalledOnce();
    expect(bundledLayout).not.toHaveBeenCalled();

    control.mode = "ok";
    expect(await computeAutoLayout(nodes, [], {}, "/elk-worker.js")).toHaveLength(1);
    expect(workers).toHaveLength(2);
    expect(workers[1].terminate).not.toHaveBeenCalled();
  });

  it("propaga errores de ELK sin cambiar de motor y conserva el worker", async () => {
    const { computeAutoLayout, workers, bundledLayout, control } = await prepare({ mode: "elk-error" });
    await expect(computeAutoLayout(nodes, [], {}, "/elk-worker.js")).rejects.toThrow("invalid graph");
    expect(bundledLayout).not.toHaveBeenCalled();
    expect(workers[0].terminate).not.toHaveBeenCalled();

    control.mode = "ok";
    await computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    expect(workers).toHaveLength(1);
  });

  it("libera el worker inactivo y crea otro en la siguiente solicitud", async () => {
    const { computeAutoLayout, workers } = await prepare();
    vi.useFakeTimers();
    await computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    await vi.advanceTimersByTimeAsync(59_000);
    expect(workers[0].terminate).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1_000);
    expect(workers[0].terminate).toHaveBeenCalledOnce();

    await computeAutoLayout(nodes, [], {}, "/elk-worker.js");
    expect(workers).toHaveLength(2);
  });
});
