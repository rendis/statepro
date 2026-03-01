import { describe, expect, it } from "vitest";

import { selectSkeletonTransitionIds } from "../utils/transitionLod";

describe("selectSkeletonTransitionIds", () => {
  it("siempre incluye transiciones marcadas como always-visible", () => {
    const selected = selectSkeletonTransitionIds({
      transitions: [{ id: "a" }, { id: "b" }, { id: "c" }],
      alwaysVisibleTransitionIds: new Set(["c"]),
      transitionBadgeAnchors: new Map([
        ["a", { x: 10, y: 0 }],
        ["b", { x: 20, y: 0 }],
        ["c", { x: 30, y: 0 }],
      ]),
      viewportCenter: { x: 0, y: 0 },
      limit: 1,
    });

    expect(selected.has("c")).toBe(true);
  });

  it("respeta el limite y ordena por distancia al viewport center", () => {
    const selected = selectSkeletonTransitionIds({
      transitions: [{ id: "t1" }, { id: "t2" }, { id: "t3" }, { id: "t4" }],
      alwaysVisibleTransitionIds: [],
      transitionBadgeAnchors: new Map([
        ["t1", { x: 200, y: 0 }],
        ["t2", { x: 50, y: 0 }],
        ["t3", { x: 120, y: 0 }],
        ["t4", { x: 10, y: 0 }],
      ]),
      viewportCenter: { x: 0, y: 0 },
      limit: 2,
    });

    expect(Array.from(selected)).toEqual(["t4", "t2"]);
  });

  it("cuando hay empate de distancia, conserva orden deterministico por indice", () => {
    const selected = selectSkeletonTransitionIds({
      transitions: [{ id: "x1" }, { id: "x2" }, { id: "x3" }],
      alwaysVisibleTransitionIds: [],
      transitionBadgeAnchors: new Map([
        ["x1", { x: 100, y: 0 }],
        ["x2", { x: -100, y: 0 }],
        ["x3", { x: 200, y: 0 }],
      ]),
      viewportCenter: { x: 0, y: 0 },
      limit: 2,
    });

    expect(Array.from(selected)).toEqual(["x1", "x2"]);
  });
});
