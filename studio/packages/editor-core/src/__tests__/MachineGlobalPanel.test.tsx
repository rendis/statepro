import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { MachineGlobalPanel } from "../features/machine";
import { StudioI18nProvider } from "../i18n";
import type { EditorNode, MachineConfig } from "../types";

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
      description: "",
      tags: [],
      metadata: "{}",
      universalConstants: {
        entryActions: [],
        exitActions: [],
        entryInvokes: [],
        exitInvokes: [],
        actionsOnTransition: [],
        invokesOnTransition: [],
      },
    },
  },
];

const machineConfig: MachineConfig = {
  id: "machine-1",
  canonicalName: "machine-canonical",
  version: "1.0.0",
  description: "",
  initials: [],
  universalConstants: {
    entryActions: [],
    exitActions: [],
    entryInvokes: [],
    exitInvokes: [],
    actionsOnTransition: [],
    invokesOnTransition: [],
  },
  metadata: "{}",
};

const emptyMetadataPackBindings = {
  machine: [],
  universe: [],
  reality: [],
  transition: [],
};

describe("MachineGlobalPanel", () => {
  it("aplica color por fase en labels de universal constants", async () => {
    const user = userEvent.setup();

    render(
      <StudioI18nProvider locale="es">
      <MachineGlobalPanel
        config={machineConfig}
        nodes={nodes}
        registry={[]}
        metadataPackRegistry={[]}
        metadataPackBindings={emptyMetadataPackBindings}
        setConfig={vi.fn() as never}
        setMetadataPackBindings={vi.fn() as never}
        openBehaviorModal={vi.fn() as never}
      />
      </StudioI18nProvider>,
    );

    await user.click(screen.getByText(/maquina de estados/i));
    expect(screen.queryByText("Entry Phase")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Universal Constants" }));

    const entryLabel = screen.getByText("Entry Phase");
    const exitLabel = screen.getByText("Exit Phase");
    const transitionLabel = screen.getByText("On Transition");

    expect(entryLabel).toHaveClass("text-yellow-400");
    expect(exitLabel).toHaveClass("text-orange-400");
    expect(transitionLabel).toHaveClass("text-slate-300");

    expect(entryLabel).not.toHaveClass("text-slate-500");
    expect(exitLabel).not.toHaveClass("text-slate-500");
    expect(transitionLabel).not.toHaveClass("text-slate-500");
  });

  it("muestra metadata en tab avanzado", async () => {
    const user = userEvent.setup();

    render(
      <StudioI18nProvider locale="es">
      <MachineGlobalPanel
        config={machineConfig}
        nodes={nodes}
        registry={[]}
        metadataPackRegistry={[]}
        metadataPackBindings={emptyMetadataPackBindings}
        setConfig={vi.fn() as never}
        setMetadataPackBindings={vi.fn() as never}
        openBehaviorModal={vi.fn() as never}
      />
      </StudioI18nProvider>,
    );

    await user.click(screen.getByText(/maquina de estados/i));
    await user.click(screen.getByRole("button", { name: "Avanzado" }));

    expect(screen.getByText("metadata (JSON)")).toBeInTheDocument();
    expect(screen.queryByText("Entry Phase")).not.toBeInTheDocument();
  });

  it("cierra el dropdown de initials con Escape y con click fuera", async () => {
    const user = userEvent.setup();

    render(
      <StudioI18nProvider locale="es">
      <MachineGlobalPanel
        config={machineConfig}
        nodes={nodes}
        registry={[]}
        metadataPackRegistry={[]}
        metadataPackBindings={emptyMetadataPackBindings}
        setConfig={vi.fn() as never}
        setMetadataPackBindings={vi.fn() as never}
        openBehaviorModal={vi.fn() as never}
      />
      </StudioI18nProvider>,
    );

    await user.click(screen.getByText(/maquina de estados/i));
    await user.click(screen.getByText("Seleccionar referencias..."));
    expect(screen.getByText("U:main")).toBeInTheDocument();

    await user.keyboard("{Escape}");
    expect(screen.queryByText("U:main")).not.toBeInTheDocument();

    await user.click(screen.getByText("Seleccionar referencias..."));
    expect(screen.getByText("U:main")).toBeInTheDocument();

    await user.click(screen.getByText("version *"));
    expect(screen.queryByText("U:main")).not.toBeInTheDocument();
  });

  it("mantiene machine id en solo lectura y normaliza canonicalName", async () => {
    const user = userEvent.setup();

    const Harness = () => {
      const [config, setConfig] = useState<MachineConfig>(machineConfig);
      return (
        <MachineGlobalPanel
          config={config}
          nodes={nodes}
          registry={[]}
          metadataPackRegistry={[]}
          metadataPackBindings={emptyMetadataPackBindings}
          setConfig={setConfig}
          setMetadataPackBindings={vi.fn() as never}
          openBehaviorModal={vi.fn() as never}
        />
      );
    };

    render(
      <StudioI18nProvider locale="es">
        <Harness />
      </StudioI18nProvider>,
    );

    await user.click(screen.getByText(/maquina de estados/i));

    const idInput = screen.getByDisplayValue("machine-1");
    const canonicalInput = screen.getByDisplayValue("machine-canonical");

    expect(idInput).toHaveAttribute("readonly");
    expect(idInput).toHaveValue("machine-1");

    await user.clear(canonicalInput);
    await user.type(canonicalInput, "__99Canonical--");
    expect(canonicalInput).toHaveValue("canonical-");

    await user.tab();
    expect(canonicalInput).toHaveValue("canonical");
  });

  it("restaura el último valor válido cuando canonical queda inválido al editar", async () => {
    const user = userEvent.setup();

    const Harness = () => {
      const [config, setConfig] = useState<MachineConfig>(machineConfig);
      return (
        <MachineGlobalPanel
          config={config}
          nodes={nodes}
          registry={[]}
          metadataPackRegistry={[]}
          metadataPackBindings={emptyMetadataPackBindings}
          setConfig={setConfig}
          setMetadataPackBindings={vi.fn() as never}
          openBehaviorModal={vi.fn() as never}
        />
      );
    };

    render(
      <StudioI18nProvider locale="es">
        <Harness />
      </StudioI18nProvider>,
    );

    await user.click(screen.getByText(/maquina de estados/i));

    const canonicalInput = screen.getByDisplayValue("machine-canonical");

    await user.clear(canonicalInput);
    await user.type(canonicalInput, "__99---");
    expect(canonicalInput).toHaveValue("");
    await user.tab();
    expect(canonicalInput).toHaveValue("machine-canonical");
  });
});
