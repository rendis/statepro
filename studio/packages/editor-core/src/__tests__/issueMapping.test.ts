import { describe, expect, it } from "vitest";

import { buildIssueIndex } from "../utils";
import type { EditorNode, EditorTransition, SerializeIssue } from "../types";

const nodes: EditorNode[] = [
  {
    id: "u-node-1",
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
    id: "r-node-1",
    type: "reality",
    x: 10,
    y: 10,
    data: {
      id: "idle",
      name: "idle",
      universeId: "u-node-1",
      isInitial: true,
      realityType: "normal",
    },
  },
];

const transitions: EditorTransition[] = [
  {
    id: "tr-1",
    sourceRealityId: "r-node-1",
    triggerKind: "always",
    type: "default",
    condition: undefined,
    conditions: [],
    actions: [],
    invokes: [],
    description: "",
    metadata: "",
    targets: ["idle"],
    order: 0,
  },
];

describe("buildIssueIndex", () => {
  it("mapea errores de universes/realities/always[i] a nodos y transición", () => {
    const issues: SerializeIssue[] = [
      {
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: "universes.main.realities.idle.always[0].targets[0]",
        message: "Invalid target",
      },
    ];

    const index = buildIssueIndex(issues, nodes, transitions);

    expect(index.universes.get("u-node-1")?.length).toBe(1);
    expect(index.realities.get("r-node-1")?.length).toBe(1);
    expect(index.transitions.get("tr-1")?.length).toBe(1);
  });

  it("mapea errores directos transition:<id> a transición", () => {
    const issues: SerializeIssue[] = [
      {
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: "transition:tr-1.eventName",
        message: "Missing eventName",
      },
    ];

    const index = buildIssueIndex(issues, nodes, transitions);
    expect(index.transitions.get("tr-1")?.length).toBe(1);
  });
});
