import "@testing-library/jest-dom";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
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

const countUniverseNodes = (): number => screen.queryAllByTestId(/universe-node-/).length;
const countRealityNodes = (): number => screen.queryAllByTestId(/reality-node-wrapper-/).length;
const countTransitionBadges = (): number => screen.queryAllByTestId(/transition-badge-/).length;

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

const alwaysTransitionsValue: StudioExternalValue = {
  definition: {
    id: "always-order-machine",
    canonicalName: "always-order-machine",
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
            always: [
              {
                targets: ["done"],
              },
              {
                targets: ["done"],
              },
            ],
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

describe("StateProEditor history", () => {
  it("clona el mismo universo con ids derivados únicos", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const sourceUniverseNode = screen.getByTestId("universe-node-univ-1");

    fireEvent.mouseDown(sourceUniverseNode, { clientX: 520, clientY: 420, button: 0 });
    fireEvent.mouseUp(sourceUniverseNode, { clientX: 520, clientY: 420, button: 0 });
    await user.click(screen.getByTitle(/clonar universo/i));

    await waitFor(() => {
      expect(countUniverseNodes()).toBe(2);
      expect(screen.getByText("main-universe-copy")).toBeInTheDocument();
    });

    fireEvent.mouseDown(sourceUniverseNode, { clientX: 520, clientY: 420, button: 0 });
    fireEvent.mouseUp(sourceUniverseNode, { clientX: 520, clientY: 420, button: 0 });
    await user.click(screen.getByTitle(/clonar universo/i));

    await waitFor(() => {
      expect(countUniverseNodes()).toBe(3);
      expect(screen.getByText("main-universe-copy-2")).toBeInTheDocument();
    });
  });

  it("crear universo -> undo lo elimina -> redo lo restaura", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const undoButton = screen.getByRole("button", { name: /deshacer/i });
    const redoButton = screen.getByRole("button", { name: /rehacer/i });

    expect(countUniverseNodes()).toBe(1);
    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));

    await waitFor(() => {
      expect(countUniverseNodes()).toBe(2);
      expect(undoButton).toBeEnabled();
    });

    await user.click(undoButton);
    await waitFor(() => {
      expect(countUniverseNodes()).toBe(1);
      expect(redoButton).toBeEnabled();
    });

    await user.click(redoButton);
    await waitFor(() => {
      expect(countUniverseNodes()).toBe(2);
    });
  });

  it("drag de universo se guarda como 1 paso de historial", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const canvas = screen.getByTestId("editor-canvas");
    const universeNode = screen.getByTestId("universe-node-univ-1");
    const undoButton = screen.getByRole("button", { name: /deshacer/i });
    const redoButton = screen.getByRole("button", { name: /rehacer/i });

    const initialLeft = universeNode.style.left;
    const initialTop = universeNode.style.top;

    fireEvent.mouseDown(universeNode, { clientX: 520, clientY: 420, button: 0 });
    fireEvent.mouseMove(canvas, { clientX: 700, clientY: 520, buttons: 1 });
    fireEvent.mouseUp(canvas, { clientX: 700, clientY: 520, button: 0 });

    await waitFor(() => {
      expect(universeNode.style.left).not.toBe(initialLeft);
      expect(universeNode.style.top).not.toBe(initialTop);
      expect(undoButton).toBeEnabled();
    });

    const movedLeft = universeNode.style.left;
    const movedTop = universeNode.style.top;

    await user.click(undoButton);
    await waitFor(() => {
      expect(universeNode.style.left).toBe(initialLeft);
      expect(universeNode.style.top).toBe(initialTop);
      expect(redoButton).toBeEnabled();
    });

    await user.click(redoButton);
    await waitFor(() => {
      expect(universeNode.style.left).toBe(movedLeft);
      expect(universeNode.style.top).toBe(movedTop);
    });
  });

  it("edición rápida de texto queda coalescida en un solo undo", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await user.click(screen.getByText(/maquina de estados/i));
    const machineCanonicalInput = screen.getByDisplayValue("admission-process");
    await user.clear(machineCanonicalInput);
    await user.type(machineCanonicalInput, "history-coalesced-canonical");
    expect(machineCanonicalInput).toHaveValue("history-coalesced-canonical");

    await user.click(screen.getByRole("button", { name: /deshacer/i }));
    await waitFor(() => {
      expect(screen.getByDisplayValue("admission-process")).toBeInTheDocument();
    });
  });

  it("después de undo + nueva edición, redo queda deshabilitado", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const redoButton = screen.getByRole("button", { name: /rehacer/i });
    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    await user.click(screen.getByRole("button", { name: /deshacer/i }));
    expect(redoButton).toBeEnabled();

    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    await waitFor(() => {
      expect(redoButton).toBeDisabled();
    });
  });

  it("importar modelo JSON reinicia historial y no permite volver al estado previo", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    expect(screen.getByRole("button", { name: /deshacer/i })).toBeEnabled();

    await user.click(screen.getByRole("button", { name: /importar json \/ exportar json/i }));
    fireEvent.change(screen.getByPlaceholderText(/pega aqui el quantummachine json/i), {
      target: { value: sampleImportMachine },
    });
    await user.click(screen.getByRole("button", { name: /importar modelo/i }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /deshacer/i })).toBeDisabled();
      expect(screen.queryByPlaceholderText(/pega aqui el quantummachine json/i)).not.toBeInTheDocument();
    });
  });

  it("shortcuts de teclado funcionan y se ignoran cuando el foco está en inputs", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    await user.click(screen.getByRole("button", { name: /universo en blanco/i }));
    expect(countUniverseNodes()).toBe(2);

    fireEvent.keyDown(window, { key: "z", ctrlKey: true });
    await waitFor(() => {
      expect(countUniverseNodes()).toBe(1);
    });

    fireEvent.keyDown(window, { key: "y", ctrlKey: true });
    await waitFor(() => {
      expect(countUniverseNodes()).toBe(2);
    });

    fireEvent.keyDown(window, { key: "z", ctrlKey: true });
    await waitFor(() => {
      expect(countUniverseNodes()).toBe(1);
    });

    fireEvent.keyDown(window, { key: "z", ctrlKey: true, shiftKey: true });
    await waitFor(() => {
      expect(countUniverseNodes()).toBe(2);
    });

    await user.click(screen.getByText(/maquina de estados/i));
    const machineIdInput = screen.getByDisplayValue("admission-process-machine");
    machineIdInput.focus();
    fireEvent.keyDown(machineIdInput, { key: "z", ctrlKey: true });

    expect(countUniverseNodes()).toBe(2);
  });

  it("persiste cambio de trigger a always desde el modal de transición", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const transitionBadge = screen.getByTestId("transition-badge-tr-1");
    fireEvent.doubleClick(transitionBadge, { clientX: 1120, clientY: 1120, button: 0 });

    const triggerSelect = (await screen.findAllByRole("combobox"))[0] as HTMLSelectElement;
    expect(triggerSelect.value).toBe("on");
    expect(screen.getByText(/evento a escuchar|event to listen/i)).toBeInTheDocument();

    await user.selectOptions(triggerSelect, "always");

    await waitFor(() => {
      const nextTriggerSelect = screen.getAllByRole("combobox")[0] as HTMLSelectElement;
      expect(nextTriggerSelect.value).toBe("always");
    });

    expect(screen.queryByText(/evento a escuchar|event to listen/i)).not.toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/ALWAYS/)).toBeInTheDocument();
    });
  });

  it("muestra order/total en badges de transiciones always", async () => {
    render(<StateProEditor locale="es" defaultValue={alwaysTransitionsValue} />);

    await waitFor(() => {
      expect(screen.getByText("ALWAYS 1 / 2")).toBeInTheDocument();
      expect(screen.getByText("ALWAYS 2 / 2")).toBeInTheDocument();
    });
  });

  it("Delete elimina universo, realidad y transición seleccionados", async () => {
    render(<StateProEditor locale="es" />);

    const realityWrapper = screen.getByTestId("reality-node-wrapper-real-2");
    const realityNode = realityWrapper.firstElementChild as HTMLElement;
    fireEvent.mouseDown(realityNode, { clientX: 1320, clientY: 1260, button: 0 });
    fireEvent.mouseUp(realityNode, { clientX: 1320, clientY: 1260, button: 0 });
    fireEvent.keyDown(window, { key: "Delete" });
    await waitFor(() => {
      expect(screen.queryByTestId("reality-node-wrapper-real-2")).not.toBeInTheDocument();
      expect(countRealityNodes()).toBe(1);
      expect(countTransitionBadges()).toBe(0);
    });

    const universeNode = screen.getByTestId("universe-node-univ-1");
    fireEvent.mouseDown(universeNode, { clientX: 520, clientY: 420, button: 0 });
    fireEvent.mouseUp(universeNode, { clientX: 520, clientY: 420, button: 0 });
    fireEvent.keyDown(window, { key: "Delete" });
    await waitFor(() => {
      expect(countUniverseNodes()).toBe(0);
      expect(countRealityNodes()).toBe(0);
    });
  });

  it("Delete/Backspace ignora elementos seleccionados cuando el foco está en inputs", async () => {
    const user = userEvent.setup();
    render(<StateProEditor locale="es" />);

    const realityNode = await screen.findByTestId("reality-node-wrapper-real-1");
    fireEvent.mouseDown(realityNode, { clientX: 1100, clientY: 1120, button: 0 });
    fireEvent.mouseUp(realityNode, { clientX: 1100, clientY: 1120, button: 0 });

    await user.click(screen.getByText(/maquina de estados/i));
    const machineIdInput = screen.getByDisplayValue("admission-process-machine");
    machineIdInput.focus();

    fireEvent.keyDown(machineIdInput, { key: "Delete" });
    fireEvent.keyDown(machineIdInput, { key: "Backspace" });

    await waitFor(() => {
      expect(screen.getByTestId("reality-node-wrapper-real-1")).toBeInTheDocument();
    });
  });
});
