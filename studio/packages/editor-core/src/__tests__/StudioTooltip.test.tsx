import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { StudioTooltip } from "../components/shared";

describe("StudioTooltip", () => {
  it("aparece en hover/focus y se oculta al salir", async () => {
    const user = userEvent.setup();
    render(
      <StudioTooltip label="Default tooltip">
        <button type="button">trigger</button>
      </StudioTooltip>,
    );

    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();

    const trigger = screen.getByRole("button", { name: "trigger" });
    await user.hover(trigger);

    const tooltip = await screen.findByRole("tooltip");
    expect(tooltip).toHaveAttribute("data-side", "top");
    const bubble = screen.getByText("Default tooltip");
    expect(bubble.className).toContain("whitespace-nowrap");

    await user.unhover(trigger);
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();

    await user.tab();
    expect(await screen.findByRole("tooltip")).toBeInTheDocument();
  });

  it("respeta side right", async () => {
    const user = userEvent.setup();
    render(
      <StudioTooltip label="Right tooltip" side="right">
        <button type="button">trigger</button>
      </StudioTooltip>,
    );

    await user.hover(screen.getByRole("button", { name: "trigger" }));
    const tooltip = await screen.findByRole("tooltip");
    expect(tooltip).toHaveAttribute("data-side", "right");
  });

  it("permite width wrap para textos largos", async () => {
    const user = userEvent.setup();
    render(
      <StudioTooltip label="Wrap tooltip" width="wrap">
        <button type="button">trigger</button>
      </StudioTooltip>,
    );

    await user.hover(screen.getByRole("button", { name: "trigger" }));
    const bubble = await screen.findByText("Wrap tooltip");
    expect(bubble.className).toContain("max-w-96");
    expect(bubble.className).toContain("whitespace-normal");
  });

  it("usa portal cuando portal=true", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <StudioTooltip label="Portal tooltip" portal>
        <button type="button">trigger</button>
      </StudioTooltip>,
    );

    await user.hover(screen.getByRole("button", { name: "trigger" }));
    const tooltip = await screen.findByRole("tooltip");

    expect(tooltip).toBeInTheDocument();
    expect(container.contains(tooltip)).toBe(false);
    expect(tooltip.className).toContain("pointer-events-none");
  });
});
