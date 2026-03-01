import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { LibraryModal } from "../features/modals";

describe("LibraryModal", () => {
  it("muestra firma y hint alineados para action al crear un behavior", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={vi.fn()}
        resolveUsage={() => ({ total: 0, locations: [] })}
        onDeleteBehavior={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: /new behavior/i }));

    expect(screen.getByText("function executeAction(args, context) {")).toBeInTheDocument();
    expect(screen.getByText("Injected: args, context. Return value is ignored.")).toBeInTheDocument();

    const scriptArea = container.querySelector("textarea");
    expect(scriptArea).toBeInTheDocument();
    if (!scriptArea) {
      throw new Error("Script textarea not found");
    }
    expect((scriptArea as HTMLTextAreaElement).value).toContain("Running action...");
  });

  it("sin edición manual, al cambiar a invoke actualiza firma y plantilla", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={vi.fn()}
        resolveUsage={() => ({ total: 0, locations: [] })}
        onDeleteBehavior={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: /new behavior/i }));
    await user.click(screen.getByRole("button", { name: /^invoke$/i }));

    expect(
      screen.getByText("function executeInvoke(args, context, resolve, reject) {"),
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        "Injected: args, context, resolve, reject. Call resolve() or reject(error).",
      ),
    ).toBeInTheDocument();

    const scriptArea = container.querySelector("textarea");
    expect(scriptArea).toBeInTheDocument();
    if (!scriptArea) {
      throw new Error("Script textarea not found");
    }
    expect((scriptArea as HTMLTextAreaElement).value).toContain("resolve");
  });

  it("si el script fue editado manualmente, no se sobrescribe al cambiar tipo", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={vi.fn()}
        resolveUsage={() => ({ total: 0, locations: [] })}
        onDeleteBehavior={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: /new behavior/i }));

    const scriptArea = container.querySelector("textarea");
    expect(scriptArea).toBeInTheDocument();
    if (!scriptArea) {
      throw new Error("Script textarea not found");
    }

    await user.clear(scriptArea);
    await user.type(scriptArea, "// custom script");
    await user.click(screen.getByRole("button", { name: /^invoke$/i }));

    expect(scriptArea).toHaveValue("// custom script");
    expect(
      screen.getByText("function executeInvoke(args, context, resolve, reject) {"),
    ).toBeInTheDocument();
  });

  it("al editar behavior existente, cambiar tipo no sobrescribe simScript", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[
          {
            src: "condition:isAdult",
            type: "condition",
            description: "Checks age",
            simScript: "return args.age >= 18;",
          },
        ]}
        setRegistry={vi.fn()}
        resolveUsage={() => ({ total: 0, locations: [] })}
        onDeleteBehavior={vi.fn()}
      />,
    );

    await user.click(screen.getByText("condition:isAdult"));

    const scriptArea = container.querySelector("textarea");
    expect(scriptArea).toBeInTheDocument();
    if (!scriptArea) {
      throw new Error("Script textarea not found");
    }
    expect(scriptArea).toHaveValue("return args.age >= 18;");

    await user.click(screen.getByRole("button", { name: /^action$/i }));

    expect(scriptArea).toHaveValue("return args.age >= 18;");
    expect(screen.getByText("function executeAction(args, context) {")).toBeInTheDocument();
  });

  it("al guardar no persiste simScriptCustomized en registry", async () => {
    const user = userEvent.setup();
    const setRegistry = vi.fn();

    render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={setRegistry}
        resolveUsage={() => ({ total: 0, locations: [] })}
        onDeleteBehavior={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: /new behavior/i }));
    await user.type(
      screen.getByPlaceholderText("e.g.: myapp:action:validate"),
      "custom:action:test",
    );
    await user.click(screen.getByRole("button", { name: /save to library/i }));

    expect(setRegistry).toHaveBeenCalledTimes(1);
    const nextRegistry = setRegistry.mock.calls[0]?.[0] as Array<Record<string, unknown>>;
    expect(nextRegistry).toHaveLength(1);
    expect(nextRegistry[0]).toMatchObject({
      src: "custom:action:test",
      type: "action",
    });
    expect(nextRegistry[0]).not.toHaveProperty("simScriptCustomized");
  });

  it("abre confirmación inline, muestra usos y ejecuta eliminar al confirmar", async () => {
    const user = userEvent.setup();
    const setRegistry = vi.fn();
    const onDeleteBehavior = vi.fn();
    const resolveUsage = vi.fn(() => ({
      total: 2,
      locations: [
        {
          scope: "machine" as const,
          containerId: "machine-1",
          field: "universalConstants.entryActions",
          label: "Machine:machine-1 · universalConstants.entryActions[0]",
          index: 0,
        },
        {
          scope: "reality" as const,
          containerId: "idle",
          field: "entryActions",
          label: "Reality:idle · entryActions[0]",
          index: 0,
        },
      ],
    }));

    render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[
          {
            src: "custom:action:target",
            type: "action",
            description: "Target",
            simScript: "return true;",
          },
        ]}
        setRegistry={setRegistry}
        resolveUsage={resolveUsage}
        onDeleteBehavior={onDeleteBehavior}
      />,
    );

    const deleteButton = screen.getByRole("button", {
      name: /delete custom:action:target/i,
    });

    await user.click(deleteButton);

    expect(screen.getByText(/confirm deletion/i)).toBeInTheDocument();
    expect(screen.getByText(/used in 2 reference/i)).toBeInTheDocument();
    expect(
      screen.getByText("Machine:machine-1 · universalConstants.entryActions[0]"),
    ).toBeInTheDocument();
    expect(screen.getByText("Reality:idle · entryActions[0]")).toBeInTheDocument();
    expect(onDeleteBehavior).not.toHaveBeenCalled();
    expect(resolveUsage).toHaveBeenCalledWith("custom:action:target");

    await user.click(screen.getByRole("button", { name: /cancel/i }));
    expect(screen.queryByText(/confirm deletion/i)).not.toBeInTheDocument();

    await user.click(deleteButton);
    await user.click(screen.getByRole("button", { name: /delete and clean 2 reference/i }));

    expect(onDeleteBehavior).toHaveBeenCalledTimes(1);
    expect(onDeleteBehavior).toHaveBeenCalledWith("custom:action:target");
    expect(setRegistry).not.toHaveBeenCalled();
  });

  it("protege built-ins: sin delete y en modo solo lectura", async () => {
    const user = userEvent.setup();
    const setRegistry = vi.fn();

    const { container } = render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[
          {
            src: "builtin:action:logArgs",
            type: "action",
            description: "Official built-in description",
            simScript: "console.log(args);",
          },
        ]}
        behaviorSourceIndex={{
          "builtin:action:logArgs": "builtin",
        }}
        setRegistry={setRegistry}
        resolveUsage={() => ({ total: 0, locations: [] })}
        onDeleteBehavior={vi.fn()}
      />,
    );

    expect(
      screen.queryByRole("button", { name: /delete builtin:action:logargs/i }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByText("builtin:action:logArgs"));

    const sourceInput = screen.getByDisplayValue("builtin:action:logArgs");
    expect(sourceInput).toBeDisabled();

    const descriptionInput = screen.getByDisplayValue("Official built-in description");
    expect(descriptionInput).toBeDisabled();

    const scriptArea = container.querySelector("textarea");
    expect(scriptArea).toBeInTheDocument();
    if (!scriptArea) {
      throw new Error("Script textarea not found");
    }
    expect(scriptArea).toBeDisabled();
    expect(
      screen.getByText(
        /built-in behavior: source, type, description and simulation script are managed/i,
      ),
    ).toBeInTheDocument();

    expect(
      screen.getByRole("button", { name: /save to library/i }),
    ).toBeDisabled();
    expect(setRegistry).not.toHaveBeenCalled();
  });

  it("muestra badge de origen para built-in, external y user", async () => {
    const user = userEvent.setup();
    render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[
          {
            src: "builtin:action:logArgs",
            type: "action",
            description: "Built-in",
            simScript: "return true;",
          },
          {
            src: "external:action:notify",
            type: "action",
            description: "External behavior from app args",
            simScript: "return true;",
          },
          {
            src: "user:action:custom",
            type: "action",
            description: "User",
            simScript: "return true;",
          },
        ]}
        behaviorSourceIndex={{
          "builtin:action:logArgs": "builtin",
          "external:action:notify": "external",
          "user:action:custom": "user",
        }}
        setRegistry={vi.fn()}
        resolveUsage={() => ({ total: 0, locations: [] })}
        onDeleteBehavior={vi.fn()}
      />,
    );

    expect(screen.getByText("Built-in")).toBeInTheDocument();
    expect(screen.getByText("External")).toBeInTheDocument();
    expect(screen.getByText("User")).toBeInTheDocument();

    await user.hover(screen.getByText("External"));
    const tooltip = await screen.findByRole("tooltip");
    expect(tooltip).toHaveTextContent("external:action:notify");
    expect(tooltip).toHaveTextContent("External behavior from app args");
  });
});
