import "@testing-library/jest-dom";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { PropertiesModal } from "../features/modals";
import type { EditorNode, EditorTransition } from "../types";

const nodes: EditorNode[] = [
  {
    id: "u-1",
    type: "universe",
    x: 0,
    y: 0,
    w: 400,
    h: 300,
    data: {
      id: "main",
      name: "main",
      canonicalName: "main",
      version: "1.0.0",
    },
  },
  {
    id: "r-1",
    type: "reality",
    x: 10,
    y: 10,
    data: {
      id: "idle",
      name: "idle",
      universeId: "u-1",
      isInitial: true,
      realityType: "normal",
    },
  },
  {
    id: "r-2",
    type: "reality",
    x: 20,
    y: 20,
    data: {
      id: "next",
      name: "next",
      universeId: "u-1",
      isInitial: false,
      realityType: "normal",
    },
  },
];

const transition: EditorTransition = {
  id: "tr-1",
  sourceRealityId: "r-1",
  triggerKind: "on",
  eventName: "GO",
  type: "default",
  condition: undefined,
  conditions: [],
  actions: [],
  invokes: [],
  description: "",
  metadata: "",
  targets: ["next"],
  order: 0,
};

describe("PropertiesModal transition behavior", () => {
  it("actualiza canonicalName de universo sin reescribir id", async () => {
    const updateNodeData = vi.fn();
    const universe = nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    if (!universe) {
      throw new Error("Universe fixture not found");
    }

    render(
      <PropertiesModal
        element={universe}
        nodes={nodes}
        transitions={[transition]}
        onClose={vi.fn()}
        updateNodeData={updateNodeData}
        updateTransitionData={vi.fn()}
        moveTransition={vi.fn()}
        openBehaviorModal={vi.fn() as never}
        registry={[]}
        metadataPackRegistry={[]}
        metadataPackBindings={{
          machine: [],
          universe: [],
          reality: [],
          transition: [],
        }}
        setMetadataPackBindings={vi.fn() as never}
      />,
    );

    const canonicalInput = screen
      .getAllByDisplayValue("main")
      .find((input) => !(input as HTMLInputElement).readOnly) as HTMLInputElement;
    expect(canonicalInput).toBeDefined();
    fireEvent.change(canonicalInput, { target: { value: "payments flow" } });
    fireEvent.blur(canonicalInput, { target: { value: "payments flow" } });

    expect(updateNodeData).toHaveBeenCalledWith("canonicalName", "payments-flow");
    expect(updateNodeData).not.toHaveBeenCalledWith("id", expect.anything());
    expect(updateNodeData).not.toHaveBeenCalledWith("name", expect.anything());
  });

  it("deshabilita notify cuando el target es interno al mismo universo", async () => {
    const user = userEvent.setup();
    const updateTransitionData = vi.fn();

    render(
      <PropertiesModal
        element={{ type: "transition", id: transition.id, data: transition }}
        nodes={nodes}
        transitions={[transition]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        updateTransitionData={updateTransitionData}
        moveTransition={vi.fn()}
        openBehaviorModal={vi.fn() as never}
        registry={[]}
        metadataPackRegistry={[]}
        metadataPackBindings={{
          machine: [],
          universe: [],
          reality: [],
          transition: [],
        }}
        setMetadataPackBindings={vi.fn() as never}
      />,
    );

    const comboBoxes = screen.getAllByRole("combobox");
    const typeSelect = comboBoxes[1] as HTMLSelectElement;

    expect(typeSelect).toBeDisabled();
    expect(screen.queryByRole("option", { name: /notify/i })).not.toBeInTheDocument();
    expect(screen.getByText(/notify mode safely disabled/i)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /save inspector/i }));
    expect(updateTransitionData).not.toHaveBeenCalledWith("type", "notify");
  });

  it("permite reordenar transición y muestra errores inline del elemento", async () => {
    const user = userEvent.setup();
    const moveTransition = vi.fn();

    const secondTransition: EditorTransition = {
      ...transition,
      id: "tr-2",
      order: 1,
    };

    render(
      <PropertiesModal
        element={{ type: "transition", id: transition.id, data: transition }}
        nodes={nodes}
        transitions={[transition, secondTransition]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        updateTransitionData={vi.fn()}
        moveTransition={moveTransition}
        openBehaviorModal={vi.fn() as never}
        registry={[]}
        metadataPackRegistry={[]}
        metadataPackBindings={{
          machine: [],
          universe: [],
          reality: [],
          transition: [],
        }}
        setMetadataPackBindings={vi.fn() as never}
        issues={[
          {
            code: "SEMANTIC_ERROR",
            severity: "error",
            field: "transition:tr-1.targets[0]",
            message: "Invalid target",
          },
        ]}
      />,
    );

    expect(screen.getByText(/validation errors/i)).toBeInTheDocument();
    expect(screen.getByText(/priority/i)).toBeInTheDocument();

    await user.click(screen.getByTitle(/move down priority/i));
    expect(moveTransition).toHaveBeenCalledWith("down");
  });

  it("usa selector de targets sin input libre y permite limpiar target inválido", async () => {
    const user = userEvent.setup();
    const updateTransitionData = vi.fn();

    const transitionWithInvalidTarget: EditorTransition = {
      ...transition,
      triggerKind: "always",
      eventName: undefined,
      targets: ["BAD::REF"],
    };

    render(
      <PropertiesModal
        element={{ type: "transition", id: transitionWithInvalidTarget.id, data: transitionWithInvalidTarget }}
        nodes={nodes}
        transitions={[transitionWithInvalidTarget]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        updateTransitionData={updateTransitionData}
        moveTransition={vi.fn()}
        openBehaviorModal={vi.fn() as never}
        registry={[]}
        metadataPackRegistry={[]}
        metadataPackBindings={{
          machine: [],
          universe: [],
          reality: [],
          transition: [],
        }}
        setMetadataPackBindings={vi.fn() as never}
      />,
    );

    expect(screen.queryByPlaceholderText(/REALITY_ID or U:universe/i)).not.toBeInTheDocument();
    expect(screen.getByTestId("transition-target-selector")).toBeInTheDocument();
    expect(screen.getByText(/invalid/i)).toBeInTheDocument();

    await user.click(screen.getByTitle(/delete target/i));
    expect(updateTransitionData).toHaveBeenCalledWith("targets", []);

    updateTransitionData.mockClear();
    await user.click(screen.getByTestId("transition-target-selector"));
    await user.click(screen.getByText("next"));

    expect(updateTransitionData).toHaveBeenCalledWith("targets", ["BAD::REF", "next"]);
  });

  it("en guards deduplica conditions por src", async () => {
    const user = userEvent.setup();
    const updateTransitionData = vi.fn();
    const openBehaviorModal = vi.fn((nextState) => {
      const resolvedState =
        typeof nextState === "function" ? nextState({} as never) : nextState;
      if (resolvedState?.isOpen && resolvedState?.onSave) {
        resolvedState.onSave({ src: "condition:secondary" });
      }
    });

    const guardedTransition: EditorTransition = {
      ...transition,
      condition: undefined,
      conditions: [{ src: "condition:secondary" }],
    };

    render(
      <PropertiesModal
        element={{ type: "transition", id: guardedTransition.id, data: guardedTransition }}
        nodes={nodes}
        transitions={[guardedTransition]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        updateTransitionData={updateTransitionData}
        moveTransition={vi.fn()}
        openBehaviorModal={openBehaviorModal as never}
        registry={[
          { src: "condition:secondary", type: "condition", simScript: "return true;" },
        ]}
        metadataPackRegistry={[]}
        metadataPackBindings={{
          machine: [],
          universe: [],
          reality: [],
          transition: [],
        }}
        setMetadataPackBindings={vi.fn() as never}
      />,
    );

    await user.click(screen.getByRole("button", { name: /guards/i }));
    await user.click(screen.getByRole("button", { name: /condition/i }));

    expect(updateTransitionData).toHaveBeenCalledWith("conditions", [
      { src: "condition:secondary" },
    ]);
  });

  it("persiste cambio de evento al hacer click fuera del selector", async () => {
    const user = userEvent.setup();
    const updateTransitionData = vi.fn();

    render(
      <PropertiesModal
        element={{ type: "transition", id: transition.id, data: transition }}
        nodes={nodes}
        transitions={[transition]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        updateTransitionData={updateTransitionData}
        moveTransition={vi.fn()}
        openBehaviorModal={vi.fn() as never}
        registry={[]}
        metadataPackRegistry={[]}
        metadataPackBindings={{
          machine: [],
          universe: [],
          reality: [],
          transition: [],
        }}
        setMetadataPackBindings={vi.fn() as never}
      />,
    );

    await user.click(screen.getByText("GO"));

    const eventInput = screen.getByPlaceholderText("Type or select");
    await user.clear(eventInput);
    await user.type(eventInput, "NEXT_EVENT");
    fireEvent.mouseDown(document.body);

    expect(updateTransitionData).toHaveBeenCalledWith("eventName", "NEXT_EVENT");
  });
});
