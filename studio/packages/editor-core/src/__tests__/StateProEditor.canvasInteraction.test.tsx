import "@testing-library/jest-dom";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterAll, beforeAll, describe, expect, it, vi } from "vitest";

import { StateProEditor } from "../StateProEditor";
import type { StudioExternalValue } from "../types";

class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}

const originalClientWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "clientWidth");
const originalClientHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "clientHeight");

beforeAll(() => {
  vi.stubGlobal("ResizeObserver", ResizeObserverMock);
  Object.defineProperty(HTMLElement.prototype, "clientWidth", {
    configurable: true,
    get() {
      return 1280;
    },
  });
  Object.defineProperty(HTMLElement.prototype, "clientHeight", {
    configurable: true,
    get() {
      return 720;
    },
  });
});

afterAll(() => {
  if (originalClientWidth) {
    Object.defineProperty(HTMLElement.prototype, "clientWidth", originalClientWidth);
  }
  if (originalClientHeight) {
    Object.defineProperty(HTMLElement.prototype, "clientHeight", originalClientHeight);
  }
  vi.unstubAllGlobals();
});

const getTransformLayer = (): HTMLElement => {
  const layer = document.querySelector(".react-transform-component");
  if (!(layer instanceof HTMLElement)) {
    throw new Error("Transform layer not found");
  }
  return layer;
};

const getTransitionPath = (role: "inbound" | "bridge" | "outbound"): Element => {
  const path = document.querySelector(
    `path.pointer-events-none.transition-colors[data-segment-role="${role}"]`,
  );
  if (!path || path.tagName.toLowerCase() !== "path") {
    throw new Error(`Transition path not found for role: ${role}`);
  }
  return path as Element;
};

const getTransitionOutboundSegmentCount = (): number => {
  return document.querySelectorAll(
    'path.pointer-events-none.transition-colors[data-segment-role="outbound"]',
  ).length;
};

const getUniverseTargetPort = (): HTMLElement => {
  return screen.getByTestId("universe-target-port-univ-1");
};

const getRealityTargetPort = (realityNodeId: string): HTMLElement => {
  return screen.getByTestId(`reality-target-port-${realityNodeId}`);
};

const getGlobalNoteNode = (): HTMLElement => {
  const element = document.querySelector('[data-testid^="global-note-node-"]');
  if (!(element instanceof HTMLElement)) {
    throw new Error("Global note node not found");
  }
  return element;
};

type ParsedPathCommand = {
  command: "M" | "L" | "C";
  values: number[];
};

const parsePathCommands = (d: string): ParsedPathCommand[] => {
  const chunks = d.match(/[MLC][^MLC]*/g) || [];
  return chunks.map((chunk) => ({
    command: chunk[0] as "M" | "L" | "C",
    values: (chunk.slice(1).match(/-?\d+(?:\.\d+)?/g) || []).map(Number),
  }));
};

const parsePathCommand = (d: string) => {
  const commands = parsePathCommands(d);
  const move = commands.find((entry) => entry.command === "M");
  const cubic = commands.find((entry) => entry.command === "C");
  if (!move || move.values.length < 2 || !cubic || cubic.values.length < 6) {
    throw new Error(`Unexpected path command: ${d}`);
  }

  let current = { x: move.values[0], y: move.values[1] };
  commands.forEach((entry) => {
    if (entry.command === "M" && entry.values.length >= 2) {
      current = { x: entry.values[0], y: entry.values[1] };
      return;
    }
    if (entry.command === "L" && entry.values.length >= 2) {
      current = { x: entry.values[0], y: entry.values[1] };
      return;
    }
    if (entry.command === "C" && entry.values.length >= 6) {
      current = { x: entry.values[4], y: entry.values[5] };
    }
  });

  return {
    startX: move.values[0],
    startY: move.values[1],
    control1X: cubic.values[0],
    control1Y: cubic.values[1],
    control2X: cubic.values[2],
    control2Y: cubic.values[3],
    endX: current.x,
    endY: current.y,
  };
};

const parseMoveTo = (d: string) => {
  const parsed = parsePathCommand(d);
  return { x: parsed.startX, y: parsed.startY };
};

const parseEndPoint = (d: string) => {
  const parsed = parsePathCommand(d);
  return { x: parsed.endX, y: parsed.endY };
};

const parseTranslate = (value: string | null): { x: number; y: number } => {
  const match = value?.match(/translate\(([-\d.]+)\s+([-\d.]+)\)/);
  if (!match) {
    throw new Error(`Unexpected translate transform: ${value}`);
  }

  return { x: Number(match[1]), y: Number(match[2]) };
};

const largeTransitionsValue: StudioExternalValue = {
  definition: {
    id: "large-transitions-machine",
    canonicalName: "large-transitions-machine",
    version: "1.0.0",
    initials: ["U:main"],
    universes: {
      main: {
        id: "main",
        canonicalName: "main",
        version: "1.0.0",
        initial: "idle",
        realities: {
          idle: {
            id: "idle",
            type: "transition",
            always: Array.from({ length: 96 }, () => ({
              targets: ["done"],
            })),
          },
          done: {
            id: "done",
            type: "final",
          },
        },
      },
    },
  },
};

describe("StateProEditor canvas interaction", () => {
  it("renderiza nodos en primer paint con performance aggressive", async () => {
    render(
      <StateProEditor
        locale="es"
        features={{
          performance: {
            mode: "aggressive",
            staticPressureThreshold: 1,
            onEmaMs: 1,
            offEmaMs: 1,
            onMissRatio: 0,
            offMissRatio: 0,
          },
        }}
      />,
    );

    await waitFor(() => {
      expect(screen.getByTestId("universe-node-univ-1")).toBeInTheDocument();
      expect(screen.getByTestId("reality-node-wrapper-real-1")).toBeInTheDocument();
    });
  });

  it("permite panear al arrastrar en fondo vacío", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const transformLayer = getTransformLayer();

    await waitFor(() => {
      expect(transformLayer.style.transform).toContain("scale(");
    });

    const initialTransform = transformLayer.style.transform;

    fireEvent.mouseDown(canvas, { clientX: 640, clientY: 360, button: 0 });
    fireEvent.mouseMove(window, { clientX: 760, clientY: 430, buttons: 1 });
    fireEvent.mouseUp(window, { clientX: 760, clientY: 430, button: 0 });

    await waitFor(() => {
      expect(transformLayer.style.transform).not.toBe(initialTransform);
    });
  });

  it("no deja canvas en blanco durante paneo con performance aggressive", async () => {
    render(
      <StateProEditor
        locale="es"
        features={{
          performance: {
            mode: "aggressive",
            staticPressureThreshold: 1,
            onEmaMs: 1,
            offEmaMs: 1,
            onMissRatio: 0,
            offMissRatio: 0,
          },
        }}
      />,
    );

    const canvas = screen.getByTestId("editor-canvas");
    const transformLayer = getTransformLayer();
    const initialTransform = transformLayer.style.transform;

    expect(screen.getByTestId("universe-node-univ-1")).toBeInTheDocument();

    fireEvent.mouseDown(canvas, { clientX: 640, clientY: 360, button: 0 });
    fireEvent.mouseMove(window, { clientX: 760, clientY: 440, buttons: 1 });
    fireEvent.mouseUp(window, { clientX: 760, clientY: 440, button: 0 });

    await waitFor(() => {
      expect(transformLayer.style.transform).not.toBe(initialTransform);
      expect(screen.getByTestId("universe-node-univ-1")).toBeInTheDocument();
      expect(screen.getByTestId("reality-node-wrapper-real-1")).toBeInTheDocument();
    });
  });

  it("en pan/zoom con grafo grande mantiene rutas skeleton visibles y oculta badges temporalmente", async () => {
    render(<StateProEditor locale="es" defaultValue={largeTransitionsValue} />);

    const canvas = screen.getByTestId("editor-canvas");
    await waitFor(() => {
      expect(screen.queryAllByTestId(/transition-badge-/).length).toBeGreaterThan(0);
    });

    fireEvent.mouseDown(canvas, { clientX: 640, clientY: 360, button: 0 });
    fireEvent.mouseMove(window, { clientX: 760, clientY: 440, buttons: 1 });

    await waitFor(() => {
      const skeletonSegments = document.querySelectorAll(".studio-transition-skeleton");
      expect(skeletonSegments.length).toBeGreaterThan(0);
      expect(screen.queryAllByTestId(/transition-badge-/).length).toBe(0);
    });

    fireEvent.mouseUp(window, { clientX: 760, clientY: 440, button: 0 });

    await waitFor(
      () => {
        expect(screen.queryAllByTestId(/transition-badge-/).length).toBeGreaterThan(0);
      },
      { timeout: 1200 },
    );
  });

  it("no activa skeleton al solo seleccionar un nodo sin arrastrar", async () => {
    render(<StateProEditor locale="es" defaultValue={largeTransitionsValue} />);

    await waitFor(() => {
      expect(screen.queryAllByTestId(/transition-badge-/).length).toBeGreaterThan(0);
      expect(screen.queryAllByTestId(/reality-node-wrapper-/).length).toBeGreaterThan(0);
    });
    const realityNode = screen.queryAllByTestId(/reality-node-wrapper-/)[0];
    if (!realityNode) {
      throw new Error("Reality node wrapper not found");
    }

    fireEvent.mouseDown(realityNode, { clientX: 1100, clientY: 1120, button: 0 });
    fireEvent.mouseUp(realityNode, { clientX: 1100, clientY: 1120, button: 0 });

    await waitFor(() => {
      expect(document.querySelectorAll(".studio-transition-skeleton").length).toBe(0);
      expect(screen.queryAllByTestId(/transition-badge-/).length).toBeGreaterThan(0);
    });
  });

  it("no inicia paneo al arrastrar sobre un nodo interactivo", async () => {
    render(<StateProEditor locale="es" />);

    const universeNode = screen.getByTestId("universe-node-univ-1");
    const transformLayer = getTransformLayer();

    await waitFor(() => {
      expect(transformLayer.style.transform).toContain("scale(");
    });

    const initialTransform = transformLayer.style.transform;

    fireEvent.mouseDown(universeNode, { clientX: 520, clientY: 420, button: 0 });
    fireEvent.mouseMove(window, { clientX: 620, clientY: 500, buttons: 1 });
    fireEvent.mouseUp(window, { clientX: 620, clientY: 500, button: 0 });

    expect(transformLayer.style.transform).toBe(initialTransform);
  });

  it("hace deselect con click en fondo, pero no tras drag de paneo", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const universeNode = screen.getByTestId("universe-node-univ-1");

    fireEvent.mouseDown(universeNode, { clientX: 520, clientY: 420, button: 0 });
    fireEvent.mouseUp(universeNode, { clientX: 520, clientY: 420, button: 0 });
    await waitFor(() => {
      expect(screen.getByTitle(/eliminar universo/i)).toBeInTheDocument();
    });

    fireEvent.click(canvas, { clientX: 200, clientY: 200 });
    expect(screen.queryByTitle(/eliminar universo/i)).not.toBeInTheDocument();

    fireEvent.mouseDown(universeNode, { clientX: 520, clientY: 420, button: 0 });
    fireEvent.mouseUp(universeNode, { clientX: 520, clientY: 420, button: 0 });
    await waitFor(() => {
      expect(screen.getByTitle(/eliminar universo/i)).toBeInTheDocument();
    });

    fireEvent.mouseDown(canvas, { clientX: 640, clientY: 360, button: 0 });
    fireEvent.mouseMove(window, { clientX: 740, clientY: 440, buttons: 1 });
    fireEvent.mouseUp(window, { clientX: 740, clientY: 440, button: 0 });
    fireEvent.click(canvas, { clientX: 740, clientY: 440 });

    await waitFor(() => {
      expect(screen.getByTitle(/eliminar universo/i)).toBeInTheDocument();
    });

    fireEvent.click(canvas, { clientX: 220, clientY: 220 });
    expect(screen.queryByTitle(/eliminar universo/i)).not.toBeInTheDocument();
  });

  it("muestra cursor grabbing mientras se arrastra el canvas", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");

    expect(canvas).not.toHaveClass("cursor-grabbing");

    fireEvent.mouseDown(canvas, { clientX: 640, clientY: 360, button: 0 });
    fireEvent.mouseMove(window, { clientX: 760, clientY: 430, buttons: 1 });

    await waitFor(() => {
      expect(canvas).toHaveClass("cursor-grabbing");
    });

    fireEvent.mouseUp(window, { clientX: 760, clientY: 430, button: 0 });

    await waitFor(() => {
      expect(canvas).not.toHaveClass("cursor-grabbing");
    });
  });

  it("arrastra transición moviendo curva y manteniendo endpoints anclados", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const badge = screen.getByTestId("transition-badge-tr-1");
    const initialInbound = parsePathCommand(getTransitionPath("inbound").getAttribute("d") || "");
    const initialOutbound = parsePathCommand(getTransitionPath("outbound").getAttribute("d") || "");

    fireEvent.mouseDown(badge, { clientX: 1120, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 1220, clientY: 1180, buttons: 1 });
    fireEvent.mouseUp(canvas, { clientX: 1220, clientY: 1180, button: 0 });

    await waitFor(() => {
      const movedInbound = parsePathCommand(getTransitionPath("inbound").getAttribute("d") || "");
      const movedOutbound = parsePathCommand(getTransitionPath("outbound").getAttribute("d") || "");

      expect(movedInbound.startX).toBeCloseTo(initialInbound.startX, 5);
      expect(movedInbound.startY).toBeCloseTo(initialInbound.startY, 5);
      expect(movedInbound.endX).not.toBe(initialInbound.endX);
      expect(movedOutbound.endX).toBeCloseTo(initialOutbound.endX, 5);
      expect(movedOutbound.endY).toBeCloseTo(initialOutbound.endY, 5);
      expect(movedOutbound.startX).not.toBe(initialOutbound.startX);
    });
  });

  it("mantiene tangentes locales fijas al forzar la transicion hacia la izquierda", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const badge = screen.getByTestId("transition-badge-tr-1");

    fireEvent.mouseDown(badge, { clientX: 1120, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 760, clientY: 960, buttons: 1 });
    fireEvent.mouseUp(canvas, { clientX: 760, clientY: 960, button: 0 });

    await waitFor(() => {
      const inbound = parsePathCommand(getTransitionPath("inbound").getAttribute("d") || "");
      const outbound = parsePathCommand(getTransitionPath("outbound").getAttribute("d") || "");

      expect(inbound.control1X).toBeGreaterThan(inbound.startX);
      expect(inbound.control2X).toBeLessThan(inbound.endX);
      expect(outbound.control1X).toBeGreaterThan(outbound.startX);
      expect(outbound.control2X).toBeLessThan(outbound.endX);
    });
  });

  it("muestra puertos izquierdo y derecho en el badge de transicion single-target", () => {
    render(<StateProEditor locale="es" />);

    expect(screen.getByTestId("transition-badge-port-left-tr-1")).toBeInTheDocument();
    expect(screen.getByTestId("transition-badge-port-right-tr-1")).toBeInTheDocument();
  });

  it("mantiene endpoints estables tras re-renders del badge (sin crecimiento acumulativo)", async () => {
    render(<StateProEditor locale="es" />);

    const badge = screen.getByTestId("transition-badge-tr-1");

    fireEvent.click(badge);

    await waitFor(() => {
      expect(getTransformLayer().style.transform).toContain("scale(");
    });
    const settledTransform = getTransformLayer().style.transform;
    await waitFor(() => {
      expect(getTransformLayer().style.transform).toBe(settledTransform);
    });

    let baselineInboundEnd = { x: 0, y: 0 };
    let baselineOutboundStart = { x: 0, y: 0 };

    await waitFor(() => {
      const inbound = getTransitionPath("inbound").getAttribute("d") || "";
      const outbound = getTransitionPath("outbound").getAttribute("d") || "";
      baselineInboundEnd = parseEndPoint(inbound);
      baselineOutboundStart = parseMoveTo(outbound);
      expect(Number.isFinite(baselineInboundEnd.x)).toBe(true);
      expect(Number.isFinite(baselineInboundEnd.y)).toBe(true);
      expect(Number.isFinite(baselineOutboundStart.x)).toBe(true);
      expect(Number.isFinite(baselineOutboundStart.y)).toBe(true);
    });

    for (let iteration = 0; iteration < 5; iteration += 1) {
      fireEvent.mouseEnter(badge);
      fireEvent.mouseLeave(badge);
      fireEvent.click(badge);
    }

    await waitFor(() => {
      const inbound = getTransitionPath("inbound").getAttribute("d") || "";
      const outbound = getTransitionPath("outbound").getAttribute("d") || "";
      const finalInboundEnd = parseEndPoint(inbound);
      const finalOutboundStart = parseMoveTo(outbound);

      expect(finalInboundEnd.x).toBeCloseTo(baselineInboundEnd.x, 5);
      expect(finalInboundEnd.y).toBeCloseTo(baselineInboundEnd.y, 5);
      expect(finalOutboundStart.x).toBeCloseTo(baselineOutboundStart.x, 5);
      expect(finalOutboundStart.y).toBeCloseTo(baselineOutboundStart.y, 5);
    });
  });

  it("al arrastrar transición no activa paneo y cambia ancla del badge", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const badge = screen.getByTestId("transition-badge-tr-1");
    const transformLayer = getTransformLayer();

    await waitFor(() => {
      expect(transformLayer.style.transform).toContain("scale(");
    });

    const initialTransform = transformLayer.style.transform;
    const initialBadge = screen.getByTestId("transition-badge-tr-1");
    const initialTranslate = parseTranslate(initialBadge.getAttribute("transform"));

    fireEvent.mouseDown(badge, { clientX: 1120, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 1250, clientY: 1200, buttons: 1 });
    fireEvent.mouseUp(canvas, { clientX: 1250, clientY: 1200, button: 0 });

    const movedBadge = screen.getByTestId("transition-badge-tr-1");
    const movedTranslate = parseTranslate(movedBadge.getAttribute("transform"));

    expect(transformLayer.style.transform).toBe(initialTransform);
    expect(movedTranslate.x).not.toBe(initialTranslate.x);
    expect(movedTranslate.y).not.toBe(initialTranslate.y);
  });

  it("al iniciar conexión desde puerto de transición no arrastra el badge", () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const badge = screen.getByTestId("transition-badge-tr-1");
    const outputPort = screen.getByTestId("transition-badge-port-right-tr-1");
    const initialTranslate = parseTranslate(badge.getAttribute("transform"));

    fireEvent.mouseDown(outputPort, { clientX: 1180, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 1290, clientY: 1200, buttons: 1 });
    fireEvent.mouseUp(canvas, { clientX: 1290, clientY: 1200, button: 0 });

    const finalTranslate = parseTranslate(
      screen.getByTestId("transition-badge-tr-1").getAttribute("transform"),
    );

    expect(finalTranslate.x).toBeCloseTo(initialTranslate.x, 5);
    expect(finalTranslate.y).toBeCloseTo(initialTranslate.y, 5);
  });

  it("al iniciar conexión desde puerto de realidad no arrastra el nodo", () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const sourcePort = screen.getByTestId("reality-source-port-real-1");
    const realityNode = screen.getByTestId("reality-node-wrapper-real-1").firstElementChild;
    if (!(realityNode instanceof HTMLElement)) {
      throw new Error("Reality node element not found");
    }

    const initialLeft = realityNode.style.left;
    const initialTop = realityNode.style.top;

    fireEvent.mouseDown(sourcePort, { clientX: 1210, clientY: 1130, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 1320, clientY: 1250, buttons: 1 });
    fireEvent.mouseUp(canvas, { clientX: 1320, clientY: 1250, button: 0 });

    const finalNode = screen.getByTestId("reality-node-wrapper-real-1").firstElementChild;
    if (!(finalNode instanceof HTMLElement)) {
      throw new Error("Reality node element not found after drag");
    }
    expect(finalNode.style.left).toBe(initialLeft);
    expect(finalNode.style.top).toBe(initialTop);
  });

  it("aplica cursor alias en canvas durante gesto de conexión", () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const sourcePort = screen.getByTestId("reality-source-port-real-1");

    expect(canvas).not.toHaveClass("cursor-alias");
    fireEvent.mouseDown(sourcePort, { clientX: 1210, clientY: 1130, button: 0 });
    expect(canvas).toHaveClass("cursor-alias");
    fireEvent.mouseUp(canvas, { clientX: 1210, clientY: 1130, button: 0 });
    expect(canvas).not.toHaveClass("cursor-alias");
  });

  it("agrega target arrastrando desde salida de badge hacia puerto target", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const outputPort = screen.getByTestId("transition-badge-port-right-tr-1");
    const universeTargetPort = getUniverseTargetPort();

    expect(getTransitionOutboundSegmentCount()).toBe(1);

    fireEvent.mouseDown(outputPort, { clientX: 1180, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 1060, clientY: 1240, buttons: 1 });
    fireEvent.mouseUp(universeTargetPort, { clientX: 1060, clientY: 1240, button: 0 });

    await waitFor(() => {
      expect(getTransitionOutboundSegmentCount()).toBe(2);
    });
  });

  it("agrega target arrastrando desde salida de ruta hacia puerto target", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const outputPort = screen.getByTestId("transition-port-right-hit-tr-1");
    const universeTargetPort = getUniverseTargetPort();

    expect(getTransitionOutboundSegmentCount()).toBe(1);

    fireEvent.mouseDown(outputPort, { clientX: 1180, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 1060, clientY: 1240, buttons: 1 });
    fireEvent.mouseUp(universeTargetPort, { clientX: 1060, clientY: 1240, button: 0 });

    await waitFor(() => {
      expect(getTransitionOutboundSegmentCount()).toBe(2);
    });
  });

  it("no duplica target al repetir drop sobre destino existente", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const outputPort = screen.getByTestId("transition-badge-port-right-tr-1");
    const existingTargetPort = getRealityTargetPort("real-2");

    expect(getTransitionOutboundSegmentCount()).toBe(1);

    fireEvent.mouseDown(outputPort, { clientX: 1180, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 1300, clientY: 1260, buttons: 1 });
    fireEvent.mouseUp(existingTargetPort, { clientX: 1300, clientY: 1260, button: 0 });

    fireEvent.mouseDown(outputPort, { clientX: 1180, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 1300, clientY: 1260, buttons: 1 });
    fireEvent.mouseUp(existingTargetPort, { clientX: 1300, clientY: 1260, button: 0 });

    await waitFor(() => {
      expect(getTransitionOutboundSegmentCount()).toBe(1);
    });
  });

  it("no cambia targets al soltar conexión fuera de puertos target", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const outputPort = screen.getByTestId("transition-badge-port-right-tr-1");

    expect(getTransitionOutboundSegmentCount()).toBe(1);

    fireEvent.mouseDown(outputPort, { clientX: 1180, clientY: 1120, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 900, clientY: 900, buttons: 1 });
    fireEvent.mouseUp(canvas, { clientX: 900, clientY: 900, button: 0 });

    await waitFor(() => {
      expect(getTransitionOutboundSegmentCount()).toBe(1);
    });
  });

  it("crea nota global desde toolbar", () => {
    render(<StateProEditor locale="es" />);

    fireEvent.click(screen.getByRole("button", { name: /anadir nota global/i }));

    expect(document.querySelectorAll('[data-testid^="global-note-node-"]')).toHaveLength(1);
  });

  it("mueve nota global dentro del canvas", () => {
    render(<StateProEditor locale="es" />);

    fireEvent.click(screen.getByRole("button", { name: /anadir nota global/i }));

    const canvas = screen.getByTestId("editor-canvas");
    const initialNode = getGlobalNoteNode();
    const initialLeft = Number.parseFloat(initialNode.style.left || "0");
    const initialTop = Number.parseFloat(initialNode.style.top || "0");

    fireEvent.mouseDown(initialNode, { clientX: 500, clientY: 320, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 650, clientY: 460, buttons: 1 });
    fireEvent.mouseUp(canvas, { clientX: 650, clientY: 460, button: 0 });

    const movedNode = getGlobalNoteNode();
    const movedLeft = Number.parseFloat(movedNode.style.left || "0");
    const movedTop = Number.parseFloat(movedNode.style.top || "0");

    expect(movedLeft).not.toBe(initialLeft);
    expect(movedTop).not.toBe(initialTop);
  });

  it("colapsa y expande nota global", async () => {
    render(<StateProEditor locale="es" />);

    fireEvent.click(screen.getByRole("button", { name: /anadir nota global/i }));
    expect(screen.getByPlaceholderText("Nota global flotante...")).toBeInTheDocument();

    fireEvent.click(screen.getByTitle("Colapsar"));

    await waitFor(() => {
      expect(screen.queryByPlaceholderText("Nota global flotante...")).not.toBeInTheDocument();
    });
    expect(screen.getByTitle("Expandir")).toBeInTheDocument();

    fireEvent.click(screen.getByTitle("Expandir"));

    await waitFor(() => {
      expect(screen.getByPlaceholderText("Nota global flotante...")).toBeInTheDocument();
    });
  });

  it("muestra indicador de nota anclada cuando el elemento pierde foco", async () => {
    render(<StateProEditor locale="es" />);

    const realityWrapper = screen.getByTestId("reality-node-wrapper-real-1");
    const realityNode = realityWrapper.firstElementChild as HTMLElement;
    fireEvent.mouseDown(realityNode, { clientX: 1120, clientY: 1130, button: 0 });
    fireEvent.mouseUp(realityNode, { clientX: 1120, clientY: 1130, button: 0 });

    fireEvent.mouseDown(await within(realityWrapper).findByTitle(/anadir nota/i), { button: 0 });
    expect(screen.getByPlaceholderText("Escribe una nota aqui...")).toBeInTheDocument();

    fireEvent.click(screen.getByTestId("editor-canvas"), { clientX: 220, clientY: 220 });

    await waitFor(() => {
      expect(screen.getByTitle(/ver nota/i)).toBeInTheDocument();
    });
    expect(screen.queryByPlaceholderText("Escribe una nota aqui...")).not.toBeInTheDocument();
  });
});
