import { describe, expect, it } from "vitest";

import { deserializeStatePro } from "../model";
import type { StateProMachine } from "../types";

describe("deserializeStatePro condition migration", () => {
  it("migrates legacy condition into conditions for editor state and imported snapshot", () => {
    const legacyMachine: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      universes: {
        main: {
          id: "main",
          canonicalName: "main",
          version: "1.0.0",
          initial: "idle",
          realities: {
            idle: {
              id: "idle",
              type: "transition",
              on: {
                GO: [
                  {
                    targets: ["done"],
                    condition: { src: "condition:primary" },
                    conditions: [
                      { src: "condition:secondary" },
                      { src: "condition:primary" },
                    ],
                  },
                ],
              },
            },
            done: {
              id: "done",
              type: "final",
            },
          },
        },
      },
    };

    const state = deserializeStatePro(legacyMachine);
    const editorTransition = state.transitions.find((transition) => transition.eventName === "GO");
    const importedTransition = state.lastImportedMachine?.universes.main.realities.idle.on?.GO[0];

    expect(editorTransition?.condition).toBeUndefined();
    expect(editorTransition?.conditions.map((condition) => condition.src)).toEqual([
      "condition:primary",
      "condition:secondary",
    ]);

    expect(importedTransition).not.toHaveProperty("condition");
    expect(importedTransition?.conditions?.map((condition) => condition.src)).toEqual([
      "condition:primary",
      "condition:secondary",
    ]);

    expect(legacyMachine.universes.main.realities.idle.on?.GO[0].condition?.src).toBe(
      "condition:primary",
    );
  });
});
