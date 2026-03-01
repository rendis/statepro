import "@testing-library/jest-dom";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterAll, beforeAll, describe, expect, it, vi } from "vitest";

import { StateProEditor } from "../StateProEditor";

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

const selectUniverseNode = (nodeId: string, options?: { shiftKey?: boolean }) => {
  const node = screen.getByTestId(`universe-node-${nodeId}`);
  fireEvent.mouseDown(node, {
    clientX: 520,
    clientY: 420,
    button: 0,
    shiftKey: options?.shiftKey ?? false,
  });
  fireEvent.mouseUp(node, {
    clientX: 520,
    clientY: 420,
    button: 0,
    shiftKey: options?.shiftKey ?? false,
  });
};

const selectRealityNode = (nodeId: string, options?: { shiftKey?: boolean }) => {
  const wrapper = screen.getByTestId(`reality-node-wrapper-${nodeId}`);
  const node = wrapper.firstElementChild as HTMLElement;
  fireEvent.mouseDown(node, {
    clientX: 1160,
    clientY: 1160,
    button: 0,
    shiftKey: options?.shiftKey ?? false,
  });
  fireEvent.mouseUp(node, {
    clientX: 1160,
    clientY: 1160,
    button: 0,
    shiftKey: options?.shiftKey ?? false,
  });
};

const selectNewestUniverse = () => {
  const nodes = screen.getAllByTestId(/universe-node-/);
  const newestNode = nodes[nodes.length - 1] as HTMLElement | undefined;
  if (!newestNode) {
    throw new Error("No universe node found");
  }
  fireEvent.mouseDown(newestNode, {
    clientX: 580,
    clientY: 460,
    button: 0,
  });
  fireEvent.mouseUp(newestNode, {
    clientX: 580,
    clientY: 460,
    button: 0,
  });
};

const openVisualModeMenu = async (user: ReturnType<typeof userEvent.setup>) => {
  await user.click(screen.getByTestId("toolbar-visual-mode-trigger"));
  await waitFor(() => {
    expect(screen.getByTestId("toolbar-visual-mode-menu")).toBeInTheDocument();
  });
};

const selectVisualMode = async (
  user: ReturnType<typeof userEvent.setup>,
  mode: "off" | "hide" | "dim",
) => {
  await openVisualModeMenu(user);
  await user.click(screen.getByTestId(`toolbar-visual-mode-${mode}`));
};

describe("StateProEditor visual filter", () => {
  it("agrega y quita nodos de selección múltiple con Shift+click", () => {
    render(<StateProEditor locale="es" />);

    selectUniverseNode("univ-1");
    expect(screen.getByRole("button", { name: /eliminar universo/i })).toBeInTheDocument();

    selectRealityNode("real-1", { shiftKey: true });
    expect(screen.queryByRole("button", { name: /eliminar universo/i })).not.toBeInTheDocument();

    selectRealityNode("real-1", { shiftKey: true });
    expect(screen.getByRole("button", { name: /eliminar universo/i })).toBeInTheDocument();
  });

  it("oculta menús contextuales por componente cuando hay multi-selección", () => {
    render(<StateProEditor locale="es" />);

    selectRealityNode("real-1");
    expect(screen.getByRole("button", { name: /eliminar realidad/i })).toBeInTheDocument();

    selectUniverseNode("univ-1", { shiftKey: true });
    expect(screen.queryByRole("button", { name: /eliminar realidad/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /eliminar universo/i })).not.toBeInTheDocument();
  });

  it("control de visualización siempre visible y habilitado con o sin selección", () => {
    render(<StateProEditor locale="es" />);

    const visualTrigger = screen.getByTestId("toolbar-visual-mode-trigger");
    expect(visualTrigger).toBeInTheDocument();
    expect(visualTrigger).toBeEnabled();

    selectUniverseNode("univ-1");
    expect(visualTrigger).toBeEnabled();

    fireEvent.click(screen.getByTestId("editor-canvas"), {
      clientX: 200,
      clientY: 200,
    });

    expect(visualTrigger).toBeEnabled();
  });

  it("permite activar filtro sin selección y aplicarlo al interactuar con nodos", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    await waitFor(() => {
      expect(screen.getAllByTestId(/universe-node-/)).toHaveLength(2);
    });

    fireEvent.click(screen.getByTestId("editor-canvas"), {
      clientX: 120,
      clientY: 120,
    });

    await selectVisualMode(user, "hide");

    expect(screen.getByTestId("universe-node-univ-1")).toBeInTheDocument();

    selectNewestUniverse();
    await waitFor(() => {
      expect(screen.queryByTestId("universe-node-univ-1")).not.toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId("editor-canvas"), {
      clientX: 140,
      clientY: 140,
    });

    expect(screen.getByTestId("universe-node-univ-1")).toBeInTheDocument();
    expect(screen.getByTestId("transition-badge-tr-1")).toBeInTheDocument();

    selectNewestUniverse();
    await waitFor(() => {
      expect(screen.queryByTestId("universe-node-univ-1")).not.toBeInTheDocument();
    });
  });

  it("modo hide oculta nodos y transiciones no conectados", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    await waitFor(() => {
      expect(screen.getAllByTestId(/universe-node-/)).toHaveLength(2);
    });

    selectNewestUniverse();
    await selectVisualMode(user, "hide");

    await waitFor(() => {
      expect(screen.queryByTestId("universe-node-univ-1")).not.toBeInTheDocument();
      expect(screen.queryByTestId("transition-badge-tr-1")).not.toBeInTheDocument();
    });
  });

  it("vuelve a mostrar todo al deseleccionar canvas con modo activo", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    await waitFor(() => {
      expect(screen.getAllByTestId(/universe-node-/)).toHaveLength(2);
    });

    selectNewestUniverse();
    await selectVisualMode(user, "hide");

    await waitFor(() => {
      expect(screen.queryByTestId("universe-node-univ-1")).not.toBeInTheDocument();
    });

    fireEvent.click(screen.getByTestId("editor-canvas"), {
      clientX: 180,
      clientY: 180,
    });

    expect(screen.getByTestId("universe-node-univ-1")).toBeInTheDocument();
    expect(screen.getByTestId("transition-badge-tr-1")).toBeInTheDocument();
  });

  it("modo dim mantiene render y opaca elementos no conectados", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    await waitFor(() => {
      expect(screen.getAllByTestId(/universe-node-/)).toHaveLength(2);
    });

    selectNewestUniverse();
    await selectVisualMode(user, "dim");

    const unrelatedUniverse = screen.getByTestId("universe-node-univ-1");
    expect(unrelatedUniverse).toHaveClass("studio-visual-muted");
    expect(unrelatedUniverse).toHaveClass("pointer-events-none");

    const unrelatedTransition = screen.getByTestId("transition-badge-tr-1");
    expect(unrelatedTransition.closest(".studio-visual-muted")).toBeTruthy();
  });

  it("cambia hide/dim/off desde el selector de visualización", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    await waitFor(() => {
      expect(screen.getAllByTestId(/universe-node-/)).toHaveLength(2);
    });

    selectNewestUniverse();

    await selectVisualMode(user, "hide");
    await waitFor(() => {
      expect(screen.queryByTestId("universe-node-univ-1")).not.toBeInTheDocument();
    });

    await selectVisualMode(user, "dim");
    await waitFor(() => {
      expect(screen.getByTestId("universe-node-univ-1")).toBeInTheDocument();
    });
    expect(screen.getByTestId("universe-node-univ-1")).toHaveClass("studio-visual-muted");

    await selectVisualMode(user, "off");
    await waitFor(() => {
      expect(screen.getByTestId("universe-node-univ-1")).toBeInTheDocument();
    });
    expect(screen.getByTestId("universe-node-univ-1")).not.toHaveClass("studio-visual-muted");
  });
});
