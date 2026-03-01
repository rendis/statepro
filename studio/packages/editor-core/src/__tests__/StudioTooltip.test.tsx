import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { StudioTooltip } from "../components/shared";

describe("StudioTooltip", () => {
  it("usa side top por defecto y aparece en hover/focus", async () => {
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
    expect(tooltip.className).toContain("bottom-full left-1/2 -translate-x-1/2 mb-2");

    const bubble = screen.getByText("Default tooltip");
    expect(bubble.className).toContain("whitespace-nowrap");

    await user.unhover(trigger);
    expect(screen.queryByRole("tooltip")).not.toBeInTheDocument();

    await user.tab();
    expect(await screen.findByRole("tooltip")).toBeInTheDocument();
  });

  it("permite side right", async () => {
    const user = userEvent.setup();
    render(
      <StudioTooltip label="Right tooltip" side="right">
        <button type="button">trigger</button>
      </StudioTooltip>,
    );

    await user.hover(screen.getByRole("button", { name: "trigger" }));
    const tooltip = await screen.findByRole("tooltip");
    expect(tooltip.className).toContain("left-full top-1/2 -translate-y-1/2 ml-2");
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
});
