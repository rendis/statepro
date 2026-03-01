import "@testing-library/jest-dom";
import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
const originalScrollIntoView = Object.getOwnPropertyDescriptor(Element.prototype, "scrollIntoView");
const scrollIntoViewMock = vi.fn();

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
  Object.defineProperty(Element.prototype, "scrollIntoView", {
    configurable: true,
    value: scrollIntoViewMock,
  });
});

afterAll(() => {
  if (originalClientWidth) {
    Object.defineProperty(HTMLElement.prototype, "clientWidth", originalClientWidth);
  }
  if (originalClientHeight) {
    Object.defineProperty(HTMLElement.prototype, "clientHeight", originalClientHeight);
  }
  if (originalScrollIntoView) {
    Object.defineProperty(Element.prototype, "scrollIntoView", originalScrollIntoView);
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

const valueWithTags: StudioExternalValue = {
  definition: {
    id: "machine-with-tags",
    canonicalName: "machine-with-tags",
    version: "1.0.0",
    initials: ["U:payments"],
    universes: {
      payments: {
        id: "payments",
        canonicalName: "payments-flow",
        version: "1.0.0",
        tags: ["vip-priority"],
        realities: {
          idle: {
            id: "idle",
            type: "final",
          },
        },
      },
    },
  },
};

const openSearchInput = async (user: ReturnType<typeof userEvent.setup>) => {
  await user.click(screen.getByTestId("toolbar-search-toggle"));
  return screen.findByRole("textbox", { name: /buscar nodos/i });
};

describe("StateProEditor viewport toolbar", () => {
  it("muestra controles de viewport y actualiza porcentaje de zoom", async () => {
    const user = userEvent.setup();

    render(<StateProEditor locale="es" />);

    expect(screen.getByRole("button", { name: /acercar/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /alejar/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /deshacer/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /rehacer/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /autolayout/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ajustar contenido/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /centrar contenido/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /deshacer/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /rehacer/i })).toBeDisabled();

    const zoomLabel = () => screen.getByText(/^\d+%$/);
    const initialZoom = zoomLabel().textContent;
    expect(initialZoom).toBeTruthy();

    await user.click(screen.getByRole("button", { name: /acercar/i }));
    await waitFor(() => {
      expect(zoomLabel().textContent).not.toBe(initialZoom);
    });
    const zoomAfterIn = zoomLabel().textContent;
    expect(zoomAfterIn).not.toBe(initialZoom);

    await user.click(screen.getByRole("button", { name: /alejar/i }));
    await waitFor(() => {
      expect(zoomLabel().textContent).not.toBe(zoomAfterIn);
    });
    const zoomAfterOut = zoomLabel().textContent;
    expect(zoomAfterOut).not.toBe(zoomAfterIn);
  });

  it("hace zoom con Ctrl/Cmd + wheel y no con wheel sola", async () => {
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const zoomLabel = () => screen.getByText(/^\d+%$/).textContent;

    await waitFor(() => {
      expect(zoomLabel()).toMatch(/^\d+%$/);
    });

    const initialZoom = zoomLabel();

    fireEvent.wheel(canvas, { deltaY: -120 });
    expect(zoomLabel()).toBe(initialZoom);

    fireEvent.keyDown(window, { key: "Control" });
    fireEvent.wheel(canvas, { deltaY: -120 });
    fireEvent.keyUp(window, { key: "Control" });

    await waitFor(() => {
      expect(zoomLabel()).not.toBe(initialZoom);
    });

    const zoomAfterCtrl = zoomLabel();

    fireEvent.keyDown(window, { key: "Meta" });
    fireEvent.wheel(canvas, { deltaY: 120 });
    fireEvent.keyUp(window, { key: "Meta" });

    await waitFor(() => {
      expect(zoomLabel()).not.toBe(zoomAfterCtrl);
    });
  });

  it("limita el zoom out al 20%", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const zoomOutButton = screen.getByRole("button", { name: /alejar/i });
    const zoomLabel = () => screen.getByText(/^\d+%$/).textContent ?? "";
    const parseZoom = (label: string) => Number.parseInt(label.replace("%", ""), 10);

    for (let i = 0; i < 20; i += 1) {
      await user.click(zoomOutButton);
      await act(async () => {
        await new Promise((resolve) => setTimeout(resolve, 220));
      });
    }

    await waitFor(() => {
      expect(parseZoom(zoomLabel())).toBe(20);
    });

    const zoomAtLimit = zoomLabel();
    await user.click(zoomOutButton);

    await waitFor(() => {
      expect(zoomLabel()).toBe(zoomAtLimit);
    });
  });

  it("aplica autolayout al presionar el boton", async () => {
    const user = userEvent.setup();

    render(<StateProEditor locale="es" />);

    const universeNode = screen.getByTestId("universe-node-univ-1");
    const initialWidth = universeNode.style.width;
    const autoLayoutButton = screen.getByRole("button", { name: /autolayout/i });

    await user.click(autoLayoutButton);

    await waitFor(() => {
      expect(universeNode.style.width).not.toBe(initialWidth);
    });
  });

  it("muestra buscador y lista coincidencias con iconografia por entidad", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    expect(screen.getByTestId("toolbar-search-toggle")).toBeInTheDocument();
    const searchInput = await openSearchInput(user);
    expect(searchInput).toBeInTheDocument();

    await user.type(searchInput, "main");

    const universeResult = await screen.findByTestId("toolbar-search-result-univ-1");
    expect(universeResult).toBeInTheDocument();
    expect(within(universeResult).getByTestId("toolbar-search-icon-universe")).toBeInTheDocument();

    expect(screen.getByTestId("toolbar-search-filter-universe")).toHaveAttribute(
      "title",
      "Buscar universos",
    );
    expect(screen.getByTestId("toolbar-search-filter-tag")).toHaveAttribute("title", "Buscar tags");
    expect(screen.getByTestId("toolbar-search-filter-reality")).toHaveAttribute(
      "title",
      "Buscar realidades",
    );
  });

  it("permite activar/desactivar filtros por iconos y refleja resultados", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" defaultValue={valueWithTags} />);

    const searchInput = await openSearchInput(user);
    const tagFilterButton = screen.getByTestId("toolbar-search-filter-tag");

    await user.type(searchInput, "vip");
    expect(await screen.findByTestId("toolbar-search-result-univ-1")).toBeInTheDocument();

    await user.click(tagFilterButton);

    await waitFor(() => {
      expect(screen.queryByTestId("toolbar-search-result-univ-1")).not.toBeInTheDocument();
      expect(screen.getByText(/sin coincidencias/i)).toBeInTheDocument();
    });
  });

  it("navega coincidencias con flechas, centra canvas y selecciona con Enter", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const transformLayer = getTransformLayer();
    await waitFor(() => {
      expect(transformLayer.style.transform).toContain("scale(");
    });

    const searchInput = await openSearchInput(user);
    await user.type(searchInput, "processing");
    await screen.findByTestId("toolbar-search-result-real-2");

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 250));
    });

    const initialTransform = transformLayer.style.transform;
    fireEvent.keyDown(searchInput, { key: "ArrowDown" });

    await waitFor(() => {
      expect(transformLayer.style.transform).not.toBe(initialTransform);
    });

    fireEvent.keyDown(searchInput, { key: "Enter" });

    await waitFor(() => {
      expect(screen.getByTitle(/eliminar realidad/i)).toBeInTheDocument();
    });
  });

  it("hace scroll automático del listado al navegar coincidencias", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const searchInput = await openSearchInput(user);
    await user.type(searchInput, "i");

    await screen.findByTestId("toolbar-search-result-univ-1");
    await screen.findByTestId("toolbar-search-result-real-1");
    scrollIntoViewMock.mockClear();

    fireEvent.keyDown(searchInput, { key: "ArrowDown" });

    await waitFor(() => {
      expect(scrollIntoViewMock).toHaveBeenCalled();
    });
  });

  it("cierra lista al perder foco del buscador y mantiene query", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const searchInput = await openSearchInput(user);
    await user.type(searchInput, "main");
    expect(await screen.findByTestId("toolbar-search-result-univ-1")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /autolayout/i }));

    await waitFor(() => {
      expect(screen.queryByTestId("toolbar-search-result-univ-1")).not.toBeInTheDocument();
    });

    const reopenedInput = await openSearchInput(user);
    expect(reopenedInput).toHaveValue("main");
  });

  it("busca por tags de universo y permite seleccionar por click", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" defaultValue={valueWithTags} />);

    const searchInput = await openSearchInput(user);
    await user.type(searchInput, "vip");

    const universeResult = await screen.findByTestId("toolbar-search-result-univ-1");
    expect(within(universeResult).getByTestId("toolbar-search-icon-universe")).toBeInTheDocument();

    await user.click(universeResult);

    await waitFor(() => {
      expect(screen.getByTitle(/eliminar universo/i)).toBeInTheDocument();
    });
  });

  it("aplica pulse temporal en nodo de canvas al navegar resultados", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const searchInput = await openSearchInput(user);
    await user.type(searchInput, "processing");
    await screen.findByTestId("toolbar-search-result-real-2");

    fireEvent.keyDown(searchInput, { key: "ArrowDown" });

    const realityNode = screen.getByTestId("reality-node-wrapper-real-2").firstElementChild as HTMLElement;
    await waitFor(() => {
      expect(realityNode.className).toMatch(/studio-search-pulse-(a|b)/);
    });

    await waitFor(
      () => {
        expect(realityNode.className).not.toMatch(/studio-search-pulse-(a|b)/);
      },
      { timeout: 2000 },
    );
  });
});
