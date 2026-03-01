import { describe, expect, it } from "vitest";

import { validateStateProMachine } from "../model/validateStatePro";
import type { StateProMachine } from "../types";

describe("validateStateProMachine", () => {
  it("omite errores de schema de bajo valor y conserva el error semántico útil", () => {
    const machine: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      universes: {
        "main-universe": {
          id: "main-universe",
          canonicalName: "main-universe",
          version: "1.0.0",
          realities: {
            idle: {
              id: "idle",
              type: "transition",
            },
          },
        },
      },
    };

    const result = validateStateProMachine(machine);

    const schemaIssuesAtReality = result.issues.filter(
      (issue) =>
        issue.code === "SCHEMA_ERROR" &&
        issue.field === "universes.main-universe.realities.idle",
    );

    expect(schemaIssuesAtReality).toHaveLength(0);
    expect(
      result.issues.some(
        (issue) => issue.messageKey === "issue.transitionRealityNeedsOnOrAlways",
      ),
    ).toBe(true);
  });

  it("emite warning no bloqueante cuando universe.universalConstants está configurado", () => {
    const machine: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["U:main-universe"],
      universes: {
        "main-universe": {
          id: "main-universe",
          canonicalName: "main-universe",
          version: "1.0.0",
          initial: "idle",
          universalConstants: {
            entryActions: [{ src: "action:test" }],
          },
          realities: {
            idle: {
              id: "idle",
              type: "transition",
              always: [{ targets: ["done"] }],
            },
            done: {
              id: "done",
              type: "final",
            },
          },
        },
      },
    };

    const result = validateStateProMachine(machine);

    expect(result.canExport).toBe(true);
    expect(
      result.issues.some(
        (issue) =>
          issue.severity === "warning" &&
          issue.field === "universes.main-universe.universalConstants" &&
          issue.messageKey === "issue.universeConstantsRuntimeIgnored",
      ),
    ).toBe(true);
  });

  it("bloquea export cuando una transición repite la misma condition en conditions", () => {
    const machine: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["U:main-universe"],
      universes: {
        "main-universe": {
          id: "main-universe",
          canonicalName: "main-universe",
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
                    conditions: [
                      { src: "condition:isValid" },
                      { src: "condition:isValid" },
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

    const result = validateStateProMachine(machine);

    expect(result.canExport).toBe(false);
    expect(
      result.issues.some(
        (issue) =>
          issue.messageKey === "issue.duplicatedTransitionCondition" &&
          issue.field === "universes.main-universe.realities.idle.on.GO[0].conditions",
      ),
    ).toBe(true);
  });
});
