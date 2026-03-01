import "@testing-library/jest-dom";
import { createEvent, fireEvent, render, screen, within } from "@testing-library/react";
import { afterAll, beforeAll, describe, expect, it, vi } from "vitest";

import { StateProEditor } from "../StateProEditor";
import { TransitionBadge } from "../components/canvas";
import type { EditorTransition } from "../types";

class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}

beforeAll(() => {
  vi.stubGlobal("ResizeObserver", ResizeObserverMock);
});

afterAll(() => {
  vi.unstubAllGlobals();
});

describe("canvas text selection behavior", () => {
  it("prevents native text selection when starting a connection from a reality port", () => {
    render(<StateProEditor locale="es" />);

    const sourcePort = screen.getByTestId("reality-source-port-real-1");
    const mouseDownEvent = createEvent.mouseDown(sourcePort, {
      bubbles: true,
      cancelable: true,
      clientX: 1200,
      clientY: 1100,
    });

    fireEvent(sourcePort, mouseDownEvent);

    expect(mouseDownEvent.defaultPrevented).toBe(true);
  });

  it("prevents native text selection when starting a connection from transition output ports", () => {
    render(<StateProEditor locale="es" />);

    const badgeOutputPort = screen.getByTestId("transition-badge-port-right-tr-1");
    const badgeMouseDown = createEvent.mouseDown(badgeOutputPort, {
      bubbles: true,
      cancelable: true,
      clientX: 1180,
      clientY: 1120,
    });
    fireEvent(badgeOutputPort, badgeMouseDown);

    const routeOutputPort = screen.getByTestId("transition-port-right-hit-tr-1");
    const routeMouseDown = createEvent.mouseDown(routeOutputPort, {
      bubbles: true,
      cancelable: true,
      clientX: 1180,
      clientY: 1120,
    });
    fireEvent(routeOutputPort, routeMouseDown);

    expect(badgeMouseDown.defaultPrevented).toBe(true);
    expect(routeMouseDown.defaultPrevented).toBe(true);
  });

  it("uses alias cursor in connection ports (reality/transition/universe target)", () => {
    render(<StateProEditor locale="es" />);

    expect(screen.getByTestId("reality-source-port-real-1")).toHaveClass("cursor-alias");
    expect(screen.getByTestId("reality-target-port-real-1")).toHaveClass("cursor-alias");
    expect(screen.getByTestId("transition-badge-port-right-tr-1")).toHaveClass("cursor-alias");
    expect(screen.getByTestId("transition-port-right-hit-tr-1")).toHaveClass("cursor-alias");
    expect(screen.getByTestId("universe-target-port-univ-1")).toHaveClass("cursor-alias");
  });

  it("prevents text selection on double click for universe and reality nodes and keeps edit flow", () => {
    render(<StateProEditor locale="es" />);

    const universeNode = screen.getByTestId("universe-node-univ-1");
    const universeDblClick = new MouseEvent("dblclick", {
      bubbles: true,
      cancelable: true,
    });
    fireEvent(universeNode, universeDblClick);

    expect(universeDblClick.defaultPrevented).toBe(true);
    expect(screen.getByRole("button", { name: /guardar inspector/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /guardar inspector/i }));

    const realityWrapper = screen.getByTestId("reality-node-wrapper-real-1");
    const realityDblClick = new MouseEvent("dblclick", {
      bubbles: true,
      cancelable: true,
    });
    fireEvent(realityWrapper, realityDblClick);

    expect(realityDblClick.defaultPrevented).toBe(true);
    expect(screen.getByRole("button", { name: /guardar inspector/i })).toBeInTheDocument();
  });

  it("applies select-none wrappers to universe/reality labels and transition badge", () => {
    render(<StateProEditor locale="es" />);

    const universeNode = screen.getByTestId("universe-node-univ-1");
    const universeLabel = within(universeNode).getByText("main-universe");
    expect(universeLabel).toHaveClass("select-none");

    const realityWrapper = screen.getByTestId("reality-node-wrapper-real-1");
    const realityLabel = within(realityWrapper).getByText("idle");
    expect(realityLabel).toHaveClass("select-none");

    const transitionBadge = screen.getByTestId("transition-badge-tr-1");
    expect(transitionBadge).toHaveClass("select-none");
    expect(screen.getByTestId("transition-badge-port-left-tr-1")).toBeInTheDocument();
    expect(screen.getByTestId("transition-badge-port-right-tr-1")).toBeInTheDocument();
  });
});

describe("transition badge interaction behavior", () => {
  it("prevents text selection on double click while selecting and opening edit", () => {
    const onSelect = vi.fn();
    const onEdit = vi.fn();

    const transition: EditorTransition = {
      id: "tr-test",
      sourceRealityId: "real-1",
      triggerKind: "on",
      eventName: "START_PROCESS",
      type: "default",
      condition: undefined,
      conditions: [],
      actions: [],
      invokes: [],
      description: "",
      metadata: "",
      targets: ["processing"],
      order: 0,
    };

    render(
      <svg>
        <TransitionBadge
          x={100}
          y={100}
          transition={transition}
          selected={false}
          onSelect={onSelect}
          onEdit={onEdit}
        />
      </svg>,
    );

    const badge = screen.getByTestId("transition-badge-tr-test");
    const dblClickEvent = new MouseEvent("dblclick", {
      bubbles: true,
      cancelable: true,
    });

    fireEvent(badge, dblClickEvent);

    expect(dblClickEvent.defaultPrevented).toBe(true);
    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onEdit).toHaveBeenCalledTimes(1);
  });
});
