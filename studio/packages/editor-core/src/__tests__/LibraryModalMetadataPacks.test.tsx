import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { LibraryModal } from "../features/modals";

describe("LibraryModal metadata packs", () => {
  it("shows metadata packs tab and saves a new pack", async () => {
    const user = userEvent.setup();
    const setMetadataPackRegistry = vi.fn();

    render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={vi.fn()}
        resolveUsage={vi.fn(() => ({ total: 0, locations: [] }))}
        onDeleteBehavior={vi.fn()}
        metadataPackRegistry={[]}
        setMetadataPackRegistry={setMetadataPackRegistry}
      />,
    );

    await user.click(screen.getByRole("button", { name: /metadata packs/i }));
    await user.click(screen.getByRole("button", { name: /new metadata pack/i }));

    const labelInput = screen.getByLabelText(/label/i);

    await user.clear(labelInput);
    await user.type(labelInput, "Pack UX");

    expect(screen.getByLabelText(/pack id \(auto\)/i)).toHaveValue("pack-ux");

    await user.click(screen.getByRole("button", { name: /save pack/i }));

    expect(setMetadataPackRegistry).toHaveBeenCalledTimes(1);
    const nextRegistry = setMetadataPackRegistry.mock.calls[0]?.[0];
    expect(Array.isArray(nextRegistry)).toBe(true);
    expect(nextRegistry[0]?.id).toBe("pack-ux");
  });

  it("blocks save and shows inline errors for hierarchical field collisions", async () => {
    const user = userEvent.setup();

    render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={vi.fn()}
        resolveUsage={vi.fn(() => ({ total: 0, locations: [] }))}
        onDeleteBehavior={vi.fn()}
        metadataPackRegistry={[]}
        setMetadataPackRegistry={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: /metadata packs/i }));
    await user.click(screen.getByRole("button", { name: /new metadata pack/i }));

    await user.type(screen.getByLabelText(/label/i), "Pack Collision");

    await user.clear(screen.getByPlaceholderText("profile.email"));
    await user.type(screen.getByPlaceholderText("profile.email"), "name");

    await user.click(screen.getByRole("button", { name: /\+ add field/i }));
    const pathInputs = screen.getAllByPlaceholderText("profile.email");
    await user.clear(pathInputs[1]!);
    await user.type(pathInputs[1]!, "name.id");

    expect(screen.getAllByText(/collides with/i).length).toBeGreaterThan(0);
    expect(screen.queryByText(/colisiones de ownership activas/i)).toBeNull();
    expect(screen.getByRole("button", { name: /save pack/i })).toBeDisabled();
  });

  it("persists constant values for non-select fields", async () => {
    const user = userEvent.setup();
    const setMetadataPackRegistry = vi.fn();

    render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={vi.fn()}
        resolveUsage={vi.fn(() => ({ total: 0, locations: [] }))}
        onDeleteBehavior={vi.fn()}
        metadataPackRegistry={[]}
        setMetadataPackRegistry={setMetadataPackRegistry}
      />,
    );

    await user.click(screen.getByRole("button", { name: /metadata packs/i }));
    await user.click(screen.getByRole("button", { name: /new metadata pack/i }));

    await user.type(screen.getByLabelText(/label/i), "Pack Constant");

    expect(screen.queryByRole("checkbox", { name: /constant/i })).toBeNull();
    await user.selectOptions(screen.getAllByRole("combobox")[0]!, "text");
    await user.click(screen.getByRole("checkbox", { name: /constant/i }));
    await user.type(screen.getByPlaceholderText("Constant value"), "fixed-value");

    await user.click(screen.getByRole("button", { name: /save pack/i }));

    expect(setMetadataPackRegistry).toHaveBeenCalledTimes(1);
    const nextRegistry = setMetadataPackRegistry.mock.calls[0]?.[0];
    const uiPointer = nextRegistry[0]?.ui?.["/category"];
    expect(uiPointer?.constant).toBe(true);
    expect(uiPointer?.constantValue).toBe("fixed-value");
  });

  it("blocks save when another pack already uses the same label", async () => {
    const user = userEvent.setup();

    render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={vi.fn()}
        resolveUsage={vi.fn(() => ({ total: 0, locations: [] }))}
        onDeleteBehavior={vi.fn()}
        metadataPackRegistry={[
          {
            id: "base-profile",
            label: "Base Profile",
            scopes: ["machine"],
            schema: { type: "object", properties: {} },
          },
        ]}
        setMetadataPackRegistry={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: /metadata packs/i }));
    await user.click(screen.getByRole("button", { name: /new metadata pack/i }));
    await user.type(screen.getByLabelText(/label/i), "base   profile");

    expect(
      screen.getByText(/another pack with label 'Base Profile' already exists/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /save pack/i })).toBeDisabled();
  });

  it("shows scope badges in metadata pack list cards", async () => {
    const user = userEvent.setup();

    render(
      <LibraryModal
        isOpen
        onClose={vi.fn()}
        registry={[]}
        setRegistry={vi.fn()}
        resolveUsage={vi.fn(() => ({ total: 0, locations: [] }))}
        onDeleteBehavior={vi.fn()}
        metadataPackRegistry={[
          {
            id: "pack-scoped",
            label: "Pack Scoped",
            scopes: ["machine", "reality"],
            schema: { type: "object", properties: {} },
          },
        ]}
        setMetadataPackRegistry={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: /metadata packs/i }));

    expect(screen.getByText("pack-scoped")).toBeInTheDocument();
    expect(screen.getByText(/^machine$/i)).toBeInTheDocument();
    expect(screen.getByText(/^reality$/i)).toBeInTheDocument();
  });
});
