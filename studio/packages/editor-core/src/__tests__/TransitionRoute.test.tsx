import "@testing-library/jest-dom";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { TransitionRoute } from "../components/canvas";
import type { EditorNode, EditorTransition, NodeSizeMap, TransitionLeg } from "../types";

const nodes: EditorNode[] = [
  {
    id: "real-src",
    type: "reality",
    x: 100,
    y: 120,
    data: {
      id: "source",
      name: "source",
      universeId: "univ-1",
      isInitial: true,
      realityType: "normal",
    },
  },
  {
    id: "real-a",
    type: "reality",
    x: 390,
    y: 140,
    data: {
      id: "target-a",
      name: "target-a",
      universeId: "univ-1",
      isInitial: false,
      realityType: "normal",
    },
  },
  {
    id: "real-b",
    type: "reality",
    x: 420,
    y: 280,
    data: {
      id: "target-b",
      name: "target-b",
      universeId: "univ-1",
      isInitial: false,
      realityType: "success",
    },
  },
];

const nodeSizes: NodeSizeMap = {
  "real-src": { w: 192, h: 150 },
  "real-a": { w: 192, h: 150 },
  "real-b": { w: 192, h: 150 },
};

const transition: EditorTransition = {
  id: "tr-1",
  sourceRealityId: "real-src",
  triggerKind: "on",
  eventName: "GO",
  type: "default",
  condition: undefined,
  conditions: [],
  actions: [],
  invokes: [],
  description: "",
  metadata: "",
  targets: ["target-a", "target-b"],
  order: 0,
};

const legs: TransitionLeg[] = [
  {
    id: "tr-1::0",
    transitionId: "tr-1",
    source: "real-src",
    target: "real-a",
    targetRef: "target-a",
  },
  {
    id: "tr-1::1",
    transitionId: "tr-1",
    source: "real-src",
    target: "real-b",
    targetRef: "target-b",
  },
];

describe("TransitionRoute", () => {
  it("selects transition from left/right ports and outbound branch", () => {
    const onSelect = vi.fn();
    const { container } = render(
      <svg>
        <TransitionRoute
          transition={transition}
          legs={legs}
          nodes={nodes}
          nodeSizes={nodeSizes}
          selected={false}
          onSelect={onSelect}
        />
      </svg>,
    );

    fireEvent.click(screen.getByTestId("transition-port-left-hit-tr-1"));
    fireEvent.click(screen.getByTestId("transition-port-right-hit-tr-1"));

    const outboundHit = container.querySelector(
      'path[data-segment-role="outbound"][data-testid^="transition-segment-hit-"]',
    );
    if (!outboundHit || outboundHit.tagName.toLowerCase() !== "path") {
      throw new Error("Outbound segment hit area not found");
    }
    fireEvent.click(outboundHit);

    expect(onSelect).toHaveBeenCalledTimes(3);
  });

  it("dispatches output-port mousedown callback from right hit area", () => {
    const onOutputPortMouseDown = vi.fn();
    render(
      <svg>
        <TransitionRoute
          transition={transition}
          legs={legs}
          nodes={nodes}
          nodeSizes={nodeSizes}
          selected={false}
          onSelect={vi.fn()}
          onOutputPortMouseDown={onOutputPortMouseDown}
        />
      </svg>,
    );

    const rightPort = screen.getByTestId("transition-port-right-hit-tr-1");
    expect(rightPort).toHaveClass("cursor-alias");
    fireEvent.mouseDown(rightPort);
    expect(onOutputPortMouseDown).toHaveBeenCalledTimes(1);
  });

  it("emits hover state when hovering a port", () => {
    const onHover = vi.fn();
    render(
      <svg>
        <TransitionRoute
          transition={transition}
          legs={legs}
          nodes={nodes}
          nodeSizes={nodeSizes}
          selected={false}
          onSelect={vi.fn()}
          onHover={onHover}
        />
      </svg>,
    );

    const leftPortHit = screen.getByTestId("transition-port-left-hit-tr-1");
    fireEvent.mouseEnter(leftPortHit);
    fireEvent.mouseLeave(leftPortHit);

    expect(onHover).toHaveBeenNthCalledWith(1, true);
    expect(onHover).toHaveBeenNthCalledWith(2, false);
  });
});
