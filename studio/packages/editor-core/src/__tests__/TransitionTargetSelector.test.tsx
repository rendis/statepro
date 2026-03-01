import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it } from "vitest";

import { TransitionTargetSelector } from "../components/shared";
import { StudioI18nProvider } from "../i18n";
import type { EditorNode } from "../types";

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
    id: "u-2",
    type: "universe",
    x: 450,
    y: 0,
    w: 400,
    h: 300,
    data: {
      id: "payments",
      name: "payments",
      canonicalName: "payments",
      version: "1.0.0",
    },
  },
  {
    id: "r-1",
    type: "reality",
    x: 20,
    y: 20,
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
    x: 100,
    y: 20,
    data: {
      id: "next",
      name: "next",
      universeId: "u-1",
      isInitial: false,
      realityType: "normal",
    },
  },
  {
    id: "r-3",
    type: "reality",
    x: 520,
    y: 20,
    data: {
      id: "pending",
      name: "pending",
      universeId: "u-2",
      isInitial: true,
      realityType: "normal",
    },
  },
];

const renderHarness = (initialTargets: string[]) => {
  const Harness = () => {
    const [targets, setTargets] = useState(initialTargets);
    return (
      <TransitionTargetSelector
        sourceRealityNodeId="r-1"
        targets={targets}
        nodes={nodes}
        onChange={setTargets}
      />
    );
  };

  return render(
    <StudioI18nProvider locale="en">
      <Harness />
    </StudioI18nProvider>,
  );
};

describe("TransitionTargetSelector", () => {
  it("selects targets from dropdown and removes chips without free typing", async () => {
    const user = userEvent.setup();
    renderHarness([]);

    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(screen.getByText("Select targets...")).toBeInTheDocument();

    await user.click(screen.getByTestId("transition-target-selector"));
    await user.click(screen.getByText("next"));

    expect(screen.getByText("next")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /delete target/i }));
    expect(screen.queryByText("next")).not.toBeInTheDocument();
  });

  it("shows invalid imported target chip and allows removing it", async () => {
    const user = userEvent.setup();
    renderHarness(["BAD::REF"]);

    expect(screen.getByText("BAD::REF")).toBeInTheDocument();
    expect(screen.getByText(/invalid/i)).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /delete target/i }));
    expect(screen.queryByText("BAD::REF")).not.toBeInTheDocument();
    expect(screen.getByText("Select targets...")).toBeInTheDocument();
  });
});
