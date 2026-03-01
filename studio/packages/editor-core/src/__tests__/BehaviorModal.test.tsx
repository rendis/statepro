import "@testing-library/jest-dom";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { BehaviorModal } from "../features/modals";

describe("BehaviorModal", () => {
  it("muestra error si intenta guardar sin source", async () => {
    const user = userEvent.setup();

    render(
      <BehaviorModal
        isOpen
        type="action"
        initialData={null}
        onSave={vi.fn()}
        onClose={vi.fn()}
        registry={[
          {
            src: "builtin:action:test",
            type: "action",
            description: "Test action",
            simScript: "return true;",
          },
        ]}
      />,
    );

    await user.click(screen.getByRole("button", { name: /assign/i }));

    expect(screen.getByText(/must select a source/i)).toBeInTheDocument();
  });

  it("guarda src, args y metadata sin description", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();

    render(
      <BehaviorModal
        isOpen
        type="action"
        initialData={null}
        onSave={onSave}
        onClose={vi.fn()}
        registry={[
          {
            src: "builtin:action:test",
            type: "action",
            description: "Test action",
            simScript: "return true;",
          },
        ]}
      />,
    );

    expect(screen.queryByText(/^description$/i)).not.toBeInTheDocument();

    await user.selectOptions(screen.getByRole("combobox"), "builtin:action:test");

    const [argsInput, metadataInput] = screen.getAllByRole("textbox");
    fireEvent.change(argsInput, { target: { value: '{"foo":"bar"}' } });
    fireEvent.change(metadataInput, { target: { value: '{"scope":"test"}' } });

    await user.click(screen.getByRole("button", { name: /assign/i }));

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(onSave).toHaveBeenCalledWith({
      src: "builtin:action:test",
      args: { foo: "bar" },
      metadata: { scope: "test" },
    });

    const payload = onSave.mock.calls[0]?.[0] as Record<string, unknown>;
    expect(payload).not.toHaveProperty("description");
  });
});
