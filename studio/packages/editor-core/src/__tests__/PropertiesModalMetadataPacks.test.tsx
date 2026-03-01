import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { PropertiesModal } from "../features/modals";
import type { EditorNode, EditorTransition } from "../types";

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
    id: "r-1",
    type: "reality",
    x: 10,
    y: 10,
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
    x: 20,
    y: 20,
    data: {
      id: "next",
      name: "next",
      universeId: "u-1",
      isInitial: false,
      realityType: "normal",
    },
  },
];

const transition: EditorTransition = {
  id: "tr-1",
  sourceRealityId: "r-1",
  triggerKind: "on",
  eventName: "GO",
  type: "default",
  condition: undefined,
  conditions: [],
  actions: [],
  invokes: [],
  description: "",
  metadata: "",
  targets: ["next"],
  order: 0,
};

describe("PropertiesModal metadata packs tab", () => {
  it("allows attaching a compatible pack to transition", async () => {
    const user = userEvent.setup();
    const setMetadataPackBindings = vi.fn();

    render(
      <PropertiesModal
        element={{ type: "transition", id: transition.id, data: transition }}
        nodes={nodes}
        transitions={[transition]}
        onClose={vi.fn()}
        updateNodeData={vi.fn()}
        updateTransitionData={vi.fn()}
        moveTransition={vi.fn()}
        openBehaviorModal={vi.fn() as never}
        registry={[]}
        metadataPackRegistry={[
          {
            id: "pack-transition",
            label: "Transition Pack",
            scopes: ["transition"],
            schema: {
              type: "object",
              properties: {
                mode: { type: "string" },
              },
            },
          },
        ]}
        metadataPackBindings={{
          machine: [],
          universe: [],
          reality: [],
          transition: [],
        }}
        setMetadataPackBindings={setMetadataPackBindings as never}
      />,
    );

    await user.click(screen.getByRole("button", { name: /metadata packs/i }));
    await user.selectOptions(screen.getByRole("combobox"), "pack-transition");
    await user.click(screen.getByRole("button", { name: /attach/i }));

    expect(setMetadataPackBindings).toHaveBeenCalledTimes(1);
  });
});
