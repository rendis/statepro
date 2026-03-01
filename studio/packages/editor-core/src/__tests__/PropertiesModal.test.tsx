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

const noopRenameHandlers = {
  commitUniverseIdRename: (_universeNodeId: string, nextUniverseIdDraft: string) => nextUniverseIdDraft,
  commitUniverseCanonicalRename: (
    _universeNodeId: string,
    nextCanonicalDraft: string,
    options: { syncId: boolean },
  ) => ({
    id: options.syncId ? nextCanonicalDraft : "main",
    canonicalName: nextCanonicalDraft,
  }),
  commitRealityIdRename: (_realityNodeId: string, nextRealityIdDraft: string) => nextRealityIdDraft,
};

describe("PropertiesModal transition behavior", () => {
  it("en universo emparejado bloquea id y edita canonical con syncId", () => {
    const universe = nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    if (!universe) {
      throw new Error("Universe fixture not found");
    }
    const commitUniverseCanonicalRename = vi.fn(() => ({
      id: "payments-flow",
      canonicalName: "payments-flow",
    }));

    render(
      <PropertiesModal
        element={universe}
        nodes={nodes}
        transitions={[transition]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        commitUniverseIdRename={noopRenameHandlers.commitUniverseIdRename}
        commitUniverseCanonicalRename={commitUniverseCanonicalRename}
        commitRealityIdRename={noopRenameHandlers.commitRealityIdRename}
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

    const idInput = screen
      .getAllByDisplayValue("main")
      .find((input) => (input as HTMLInputElement).readOnly) as HTMLInputElement;
    const canonicalInput = screen
      .getAllByDisplayValue("main")
      .find((input) => !(input as HTMLInputElement).readOnly) as HTMLInputElement;
    expect(idInput).toHaveAttribute("readonly");

    fireEvent.change(canonicalInput, { target: { value: "payments flow" } });
    fireEvent.blur(canonicalInput);

    expect(commitUniverseCanonicalRename).toHaveBeenCalledWith("u-1", "payments-flow", {
      syncId: true,
    });
    expect(idInput).toHaveValue("payments-flow");
    expect(canonicalInput).toHaveValue("payments-flow");
  });

  it("en universo desemparejado permite editar id y activarlo vuelve a canonical", async () => {
    const user = userEvent.setup();
    const commitUniverseIdRename = vi.fn(() => "billing");
    const commitUniverseCanonicalRename = vi.fn(() => ({
      id: "main-canonical",
      canonicalName: "main-canonical",
    }));
    const universe = nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    if (!universe) {
      throw new Error("Universe fixture not found");
    }
    const unpairedUniverse = {
      ...universe,
      data: {
        ...universe.data,
        canonicalName: "main-canonical",
      },
    };

    render(
      <PropertiesModal
        element={unpairedUniverse}
        nodes={nodes}
        transitions={[transition]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        commitUniverseIdRename={commitUniverseIdRename}
        commitUniverseCanonicalRename={commitUniverseCanonicalRename}
        commitRealityIdRename={noopRenameHandlers.commitRealityIdRename}
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

    const idInput = screen.getByDisplayValue("main") as HTMLInputElement;
    expect(idInput).not.toHaveAttribute("readonly");
    const linkToggle = screen.getByRole("checkbox", { name: /link id and canonicalname/i });
    expect(linkToggle).not.toBeChecked();

    fireEvent.change(idInput, { target: { value: "billing" } });
    fireEvent.blur(idInput);
    expect(commitUniverseIdRename).toHaveBeenCalledWith("u-1", "billing");
    expect(idInput).toHaveValue("billing");

    await user.click(linkToggle);
    expect(commitUniverseCanonicalRename).toHaveBeenCalledWith("u-1", "main-canonical", {
      syncId: true,
    });
    expect(idInput).toHaveAttribute("readonly");
    expect(idInput).toHaveValue("main-canonical");
  });

  it("en realidad permite editar id y confirmar en blur", () => {
    const commitRealityIdRename = vi.fn(() => "waiting-room");
    const reality = nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.type === "reality" && node.id === "r-1",
    );
    if (!reality) {
      throw new Error("Reality fixture not found");
    }

    render(
      <PropertiesModal
        element={reality}
        nodes={nodes}
        transitions={[transition]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        commitUniverseIdRename={noopRenameHandlers.commitUniverseIdRename}
        commitUniverseCanonicalRename={noopRenameHandlers.commitUniverseCanonicalRename}
        commitRealityIdRename={commitRealityIdRename}
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

    const realityIdInput = screen.getByDisplayValue("idle") as HTMLInputElement;
    fireEvent.change(realityIdInput, { target: { value: "waiting room" } });
    fireEvent.blur(realityIdInput);

    expect(commitRealityIdRename).toHaveBeenCalledWith("r-1", "waiting-room");
    expect(realityIdInput).toHaveValue("waiting-room");
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
        commitUniverseIdRename={noopRenameHandlers.commitUniverseIdRename}
        commitUniverseCanonicalRename={noopRenameHandlers.commitUniverseCanonicalRename}
        commitRealityIdRename={noopRenameHandlers.commitRealityIdRename}
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

  it("muestra labels simplificados en trigger y mode", () => {
    const transitionWithExternalTarget: EditorTransition = {
      ...transition,
      id: "tr-external",
      targets: ["U:billing"],
    };

    render(
      <PropertiesModal
        element={{ type: "transition", id: transitionWithExternalTarget.id, data: transitionWithExternalTarget }}
        nodes={nodes}
        transitions={[transitionWithExternalTarget]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        commitUniverseIdRename={noopRenameHandlers.commitUniverseIdRename}
        commitUniverseCanonicalRename={noopRenameHandlers.commitUniverseCanonicalRename}
        commitRealityIdRename={noopRenameHandlers.commitRealityIdRename}
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

    expect(screen.getByRole("option", { name: "On" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Always" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Default" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Notify" })).toBeInTheDocument();

    expect(screen.queryByRole("option", { name: "Specific event (on)" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "Automatic (always)" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "Default (Advance)" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "Notify (Notify/Keep)" })).not.toBeInTheDocument();
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
        commitUniverseIdRename={noopRenameHandlers.commitUniverseIdRename}
        commitUniverseCanonicalRename={noopRenameHandlers.commitUniverseCanonicalRename}
        commitRealityIdRename={noopRenameHandlers.commitRealityIdRename}
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

    await user.click(screen.getByRole("button", { name: /move down priority/i }));
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
        commitUniverseIdRename={noopRenameHandlers.commitUniverseIdRename}
        commitUniverseCanonicalRename={noopRenameHandlers.commitUniverseCanonicalRename}
        commitRealityIdRename={noopRenameHandlers.commitRealityIdRename}
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

    await user.click(screen.getByRole("button", { name: /delete target/i }));
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
        commitUniverseIdRename={noopRenameHandlers.commitUniverseIdRename}
        commitUniverseCanonicalRename={noopRenameHandlers.commitUniverseCanonicalRename}
        commitRealityIdRename={noopRenameHandlers.commitRealityIdRename}
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
        commitUniverseIdRename={noopRenameHandlers.commitUniverseIdRename}
        commitUniverseCanonicalRename={noopRenameHandlers.commitUniverseCanonicalRename}
        commitRealityIdRename={noopRenameHandlers.commitRealityIdRename}
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
