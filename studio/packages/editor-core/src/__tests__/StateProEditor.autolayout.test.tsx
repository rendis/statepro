import "@testing-library/jest-dom";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { StrictMode } from "react";
import userEvent from "@testing-library/user-event";
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import type { EditorNode, EditorTransition, NodeSizeMap } from "../types";

type Deferred<T> = {
  promise: Promise<T>;
  resolve: (value: T) => void;
  reject: (error?: unknown) => void;
};

const createDeferred = <T,>(): Deferred<T> => {
  let resolve: (value: T) => void = () => undefined;
  let reject: (error?: unknown) => void = () => undefined;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
};

let pendingLayout: Deferred<EditorNode[]> | null = null;
let lastNodesInput: EditorNode[] = [];

const sampleImportMachine = JSON.stringify(
  {
    id: "imported-machine",
    canonicalName: "imported-machine",
    version: "1.0.0",
    initials: ["U:main"],
    universes: {
      main: {
        id: "main",
        canonicalName: "main",
        version: "1.0.0",
        realities: {
          idle: {
            id: "idle",
            type: "final",
          },
        },
      },
    },
  },
  null,
  2,
);

const computeAutoLayoutMock = vi.fn(
  async (nodes: EditorNode[], _transitions: EditorTransition[], _nodeSizes: NodeSizeMap) => {
    lastNodesInput = nodes;
    pendingLayout = createDeferred<EditorNode[]>();
    return pendingLayout.promise;
  },
);

vi.mock("../utils/autoLayout", () => ({
  computeAutoLayout: (
    nodes: EditorNode[],
    transitions: EditorTransition[],
    nodeSizes: NodeSizeMap,
  ) => computeAutoLayoutMock(nodes, transitions, nodeSizes),
}));

import { StateProEditor } from "../StateProEditor";

class ResizeObserverMock {
  private callback: ResizeObserverCallback;

  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
  }

  observe(target: Element) {
    this.callback(
      [
        {
          target,
          contentRect: {
            x: 0,
            y: 0,
            width: 192,
            height: 80,
            top: 0,
            left: 0,
            right: 192,
            bottom: 80,
            toJSON: () => ({}),
          } as DOMRectReadOnly,
        } as ResizeObserverEntry,
      ],
      this as unknown as ResizeObserver,
    );
  }

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

beforeEach(() => {
  computeAutoLayoutMock.mockClear();
  pendingLayout = null;
  lastNodesInput = [];
});

const resolvePendingAutoLayout = () => {
  expect(pendingLayout).not.toBeNull();
  pendingLayout?.resolve(lastNodesInput);
};

describe("StateProEditor autolayout button", () => {
  it("en StrictMode no deja el boton pegado en Laying out", async () => {
    render(
      <StrictMode>
        <StateProEditor locale="es" />
      </StrictMode>,
    );

    const button = screen.getByRole("button", { name: /autolayout/i });
    await waitFor(() => {
      expect(computeAutoLayoutMock).toHaveBeenCalled();
      expect(button).toBeDisabled();
      expect(button).toHaveTextContent(/ordenando|laying out/i);
    });

    resolvePendingAutoLayout();

    await waitFor(() => {
      expect(button).not.toBeDisabled();
      expect(button).not.toHaveTextContent(/ordenando|laying out/i);
    });
  });

  it("no invalida autolayout inicial cuando cambian nodeSizes durante mount", async () => {
    render(<StateProEditor locale="es" />);

    const button = screen.getByRole("button", { name: /autolayout/i });
    await waitFor(() => {
      expect(computeAutoLayoutMock).toHaveBeenCalledTimes(1);
      expect(button).toBeDisabled();
    });

    expect(computeAutoLayoutMock).toHaveBeenCalledTimes(1);

    resolvePendingAutoLayout();

    await waitFor(() => {
      expect(button).not.toBeDisabled();
    });
  });

  it("se deshabilita mientras el autolayout esta corriendo", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const button = screen.getByRole("button", { name: /autolayout/i });
    await waitFor(() => {
      expect(computeAutoLayoutMock).toHaveBeenCalledTimes(1);
      expect(button).toBeDisabled();
    });

    resolvePendingAutoLayout();

    await waitFor(() => {
      expect(button).not.toBeDisabled();
    });

    await user.click(button);
    expect(computeAutoLayoutMock).toHaveBeenCalledTimes(2);
    expect(button).toBeDisabled();

    resolvePendingAutoLayout();

    await waitFor(() => {
      expect(button).not.toBeDisabled();
    });
  });

  it("resetea el offset visual de transiciones al aplicar autolayout", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await waitFor(() => {
      expect(computeAutoLayoutMock).toHaveBeenCalledTimes(1);
    });
    resolvePendingAutoLayout();
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /autolayout/i })).not.toBeDisabled();
    });

    const badge = await screen.findByTestId("transition-badge-tr-1");
    const canvas = screen.getByTestId("editor-canvas");
    const initialTransform = badge.getAttribute("transform");

    expect(initialTransform).toBeTruthy();

    fireEvent.mouseDown(badge, { clientX: 400, clientY: 300 });
    fireEvent.mouseMove(canvas, { clientX: 520, clientY: 360 });
    fireEvent.mouseUp(canvas, { clientX: 520, clientY: 360 });

    await waitFor(() => {
      expect(screen.getByTestId("transition-badge-tr-1").getAttribute("transform")).not.toBe(
        initialTransform,
      );
    });

    const button = screen.getByRole("button", { name: /autolayout/i });
    await user.click(button);

    expect(computeAutoLayoutMock).toHaveBeenCalledTimes(2);
    resolvePendingAutoLayout();

    await waitFor(() => {
      expect(screen.getByTestId("transition-badge-tr-1").getAttribute("transform")).toBe(
        initialTransform,
      );
    });
  });

  it("aplica autolayout por defecto al importar un modelo", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await waitFor(() => {
      expect(computeAutoLayoutMock).toHaveBeenCalledTimes(1);
    });
    resolvePendingAutoLayout();
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /autolayout/i })).not.toBeDisabled();
    });

    await user.click(screen.getByRole("button", { name: /importar json \/ exportar json/i }));
    fireEvent.change(screen.getByPlaceholderText(/pega aqui el quantummachine json/i), {
      target: { value: sampleImportMachine },
    });

    await user.click(screen.getByRole("button", { name: /importar modelo/i }));
    expect(computeAutoLayoutMock).toHaveBeenCalledTimes(2);
    expect(lastNodesInput).toHaveLength(2);

    resolvePendingAutoLayout();

    await waitFor(() => {
      expect(screen.queryByPlaceholderText(/pega aqui el quantummachine json/i)).not.toBeInTheDocument();
    });
  });
});
