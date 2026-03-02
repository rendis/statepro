import { describe, expect, it } from "vitest";

import { resolveConnectionRenderMode } from "../utils/connectionRenderPolicy";

describe("connectionRenderPolicy", () => {
  it("usa skeleton en navegación con grafo grande", () => {
    expect(
      resolveConnectionRenderMode({
        isNavigating: true,
        transitionCount: 120,
        navigatingFullTransitionThreshold: 80,
      }),
    ).toBe("skeleton");
  });

  it("usa full en navegación con grafo pequeño", () => {
    expect(
      resolveConnectionRenderMode({
        isNavigating: true,
        transitionCount: 40,
        navigatingFullTransitionThreshold: 80,
      }),
    ).toBe("full");
  });

  it("usa full fuera de navegación (incluye edición)", () => {
    expect(
      resolveConnectionRenderMode({
        isNavigating: false,
        transitionCount: 240,
        navigatingFullTransitionThreshold: 80,
      }),
    ).toBe("full");
  });
});
