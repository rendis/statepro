import "@testing-library/jest-dom";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
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
});
