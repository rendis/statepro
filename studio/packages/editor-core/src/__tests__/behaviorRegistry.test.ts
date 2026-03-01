import { describe, expect, it } from "vitest";

import {
  buildBuiltinBehaviorRegistry,
  buildEditorStateFromExternalValue,
  composeBehaviorRegistry,
  mergeBehaviorRegistryWithExternal,
  normalizeBehaviorRegistry,
} from "../model";
import type { StudioExternalValue } from "../types";

const MINIMAL_EXTERNAL_VALUE: StudioExternalValue = {
  definition: {
    id: "machine-id",
    canonicalName: "machine",
    version: "1.0.0",
    universes: {},
  },
};

describe("behaviorRegistry composition", () => {
  it("always includes runtime built-ins and excludes removed studio seeds", () => {
    const registry = normalizeBehaviorRegistry([], { locale: "en" });
    const sources = new Set(registry.map((entry) => entry.src));

    expect(sources.has("builtin:action:logBasicInfo")).toBe(true);
    expect(sources.has("builtin:action:logArgs")).toBe(true);
    expect(sources.has("builtin:observer:containsAllEvents")).toBe(true);

    expect(sources.has("builtin:invoke:emitEvent")).toBe(false);
    expect(sources.has("condition:isAdult")).toBe(false);
  });

  it("keeps official built-in description while accepting external simScript override", () => {
    const builtin = buildBuiltinBehaviorRegistry("en").find(
      (entry) => entry.src === "builtin:action:logArgs",
    );
    expect(builtin).toBeDefined();

    const composed = composeBehaviorRegistry({
      locale: "en",
      currentRegistry: [],
      externalRegistry: [
        {
          src: "builtin:action:logArgs",
          type: "action",
          description: "External override that must be ignored",
          simScript: "// external override script",
        },
      ],
    });

    const resolved = composed.find((entry) => entry.src === "builtin:action:logArgs");
    expect(resolved).toBeDefined();
    expect(resolved?.description).toBe(builtin?.description);
    expect(resolved?.simScript).toBe("// external override script");
  });

  it("drops stale external entries when external catalog changes", () => {
    const previous = composeBehaviorRegistry({
      locale: "en",
      currentRegistry: [],
      externalRegistry: [
        {
          src: "custom:action:external",
          type: "action",
          description: "External action",
          simScript: "return true;",
        },
      ],
      preferExternalForExternalSources: true,
    });

    const next = mergeBehaviorRegistryWithExternal(previous, {
      locale: "en",
      externalRegistry: [],
      previousExternalSources: ["custom:action:external"],
    });

    expect(next.some((entry) => entry.src === "custom:action:external")).toBe(false);
    expect(next.some((entry) => entry.src === "builtin:action:logBasicInfo")).toBe(true);
  });

  it("merges built-ins with external library behaviors when loading external value", () => {
    const state = buildEditorStateFromExternalValue(MINIMAL_EXTERNAL_VALUE, {
      locale: "en",
      libraryBehaviors: [
        {
          src: "custom:observer:businessHours",
          type: "observer",
          description: "Business hours check",
          simScript: "return true;",
        },
      ],
    });

    const sources = new Set(state.registry.map((entry) => entry.src));
    expect(sources.has("builtin:action:logBasicInfo")).toBe(true);
    expect(sources.has("builtin:observer:alwaysTrue")).toBe(true);
    expect(sources.has("custom:observer:businessHours")).toBe(true);
  });
});
