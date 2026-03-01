import "@testing-library/jest-dom";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { TransitionBadge } from "../components/canvas";
import type { EditorTransition } from "../types";
import { TRIGGER_MAX_CHARS } from "../utils";

const baseTransition: EditorTransition = {
  id: "tr-svg",
  sourceRealityId: "real-1",
  triggerKind: "on",
  eventName: "VERY_LONG_EVENT_NAME_FOR_TRANSITION_BADGE",
  type: "default",
  condition: undefined,
  conditions: [],
  actions: [],
  invokes: [],
  description: "",
  metadata: "",
  targets: ["target-a"],
  order: 0,
};

describe("TransitionBadge SVG", () => {
  it("renders deterministic left/right ports", () => {
    render(
      <svg>
        <TransitionBadge
          x={240}
          y={160}
          transition={baseTransition}
          selected={false}
          onSelect={vi.fn()}
        />
      </svg>,
    );

    expect(screen.getByTestId("transition-badge-port-left-tr-svg")).toBeInTheDocument();
    const rightPort = screen.getByTestId("transition-badge-port-right-tr-svg");
    expect(rightPort).toBeInTheDocument();
    expect(rightPort).toHaveClass("cursor-alias");
  });

  it("dispatches click, double-click and mousedown interactions", () => {
    const onSelect = vi.fn();
    const onEdit = vi.fn();
    const onMouseDown = vi.fn();
    const onOutputPortMouseDown = vi.fn();

    render(
      <svg>
        <TransitionBadge
          x={240}
          y={160}
          transition={baseTransition}
          selected={false}
          onSelect={onSelect}
          onEdit={onEdit}
          onMouseDown={onMouseDown}
          onOutputPortMouseDown={onOutputPortMouseDown}
        />
      </svg>,
    );

    const badge = screen.getByTestId("transition-badge-tr-svg");
    fireEvent.mouseDown(badge);
    fireEvent.mouseDown(screen.getByTestId("transition-badge-port-right-tr-svg"));
    fireEvent.click(badge);
    fireEvent.doubleClick(badge);

    expect(onMouseDown).toHaveBeenCalledTimes(1);
    expect(onOutputPortMouseDown).toHaveBeenCalledTimes(1);
    expect(onSelect).toHaveBeenCalledTimes(2);
    expect(onEdit).toHaveBeenCalledTimes(1);
  });

  it("truncates long trigger labels with fixed max chars", () => {
    render(
      <svg>
        <TransitionBadge
          x={240}
          y={160}
          transition={baseTransition}
          selected={false}
          onSelect={vi.fn()}
        />
      </svg>,
    );

    const expectedLabel = `${baseTransition.eventName?.slice(0, TRIGGER_MAX_CHARS - 3)}...`;
    expect(screen.getByText(expectedLabel)).toBeInTheDocument();
    expect(screen.queryByText(baseTransition.eventName || "")).not.toBeInTheDocument();
  });
});
