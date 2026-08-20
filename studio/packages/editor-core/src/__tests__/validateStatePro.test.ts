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

  it("no advierte que el runtime ignore universe.universalConstants", () => {
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
    expect(result.issues).toEqual([]);
    expect(
      result.issues.some(
        (issue) =>
          issue.field === "universes.main-universe.universalConstants" ||
          issue.messageKey === "issue.universeConstantsRuntimeIgnored",
      ),
    ).toBe(false);
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

  it("rechaza máquinas hostiles: refs rotas, notify interno y universo vacío", () => {
    const brokenTargets: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["U:ghost"],
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
                GO: [{ targets: ["U:ghost"], type: "notify" }],
              },
            },
            done: { id: "done", type: "final" },
          },
        },
      },
    };

    const broken = validateStateProMachine(brokenTargets);
    expect(broken.canExport).toBe(false);
    expect(
      broken.issues.some(
        (issue) =>
          issue.messageKey === "issue.initialUnknownUniverse" ||
          issue.messageKey === "issue.unknownUniverse",
      ),
    ).toBe(true);

    const emptyUniverses: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: [],
      universes: {},
    };
    const empty = validateStateProMachine(emptyUniverses);
    expect(empty.canExport).toBe(false);
    expect(
      empty.issues.some((issue) => issue.messageKey === "issue.machineNeedsUniverse"),
    ).toBe(true);

    const notifyInternal: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["U:main"],
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
                GO: [{ targets: ["done"], type: "notify" }],
              },
            },
            done: { id: "done", type: "final" },
          },
        },
      },
    };
    const notify = validateStateProMachine(notifyInternal);
    expect(notify.canExport).toBe(false);
    expect(
      notify.issues.some((issue) => issue.messageKey === "issue.notifyInternalTarget"),
    ).toBe(true);
  });

  it("detecta target vacío, realidad interna desconocida y mismatch de keys", () => {
    const emptyTarget: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["U:main"],
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
              on: { GO: [{ targets: [] }] },
            },
            done: { id: "done", type: "final" },
          },
        },
      },
    };
    const empty = validateStateProMachine(emptyTarget);
    expect(empty.canExport).toBe(false);
    expect(
      empty.issues.some((issue) => issue.messageKey === "issue.transitionNeedsTarget"),
    ).toBe(true);

    const unknownInternal: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["U:main"],
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
              on: { GO: [{ targets: ["missing"] }] },
            },
            done: { id: "done", type: "final" },
          },
        },
      },
    };
    const unknown = validateStateProMachine(unknownInternal);
    expect(unknown.canExport).toBe(false);
    expect(
      unknown.issues.some((issue) => issue.messageKey === "issue.unknownInternalReality"),
    ).toBe(true);

    const keyMismatch: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["U:main"],
      universes: {
        main: {
          id: "other",
          canonicalName: "main",
          version: "1.0.0",
          initial: "idle",
          realities: {
            idle: {
              id: "idle",
              type: "transition",
              always: [{ targets: ["done"] }],
            },
            done: { id: "done", type: "final" },
          },
        },
      },
    };
    const mismatch = validateStateProMachine(keyMismatch);
    expect(mismatch.canExport).toBe(false);
    expect(
      mismatch.issues.some((issue) => issue.messageKey === "issue.universeKeyMismatch"),
    ).toBe(true);

    const badInitialFormat: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["not-a-ref"],
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
              always: [{ targets: ["done"] }],
            },
            done: { id: "done", type: "final" },
          },
        },
      },
    };
    const badInitial = validateStateProMachine(badInitialFormat);
    expect(badInitial.canExport).toBe(false);
    expect(
      badInitial.issues.some((issue) => issue.messageKey === "issue.initialReferenceFormat"),
    ).toBe(true);

    const unknownExternalReality: StateProMachine = {
      id: "machine",
      canonicalName: "machine",
      version: "1.0.0",
      initials: ["U:main"],
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
              on: { GO: [{ targets: ["U:main:ghost"] }] },
            },
            done: { id: "done", type: "final" },
          },
        },
      },
    };
    const ext = validateStateProMachine(unknownExternalReality);
    expect(ext.canExport).toBe(false);
    expect(
      ext.issues.some((issue) => issue.messageKey === "issue.unknownRealityInUniverse"),
    ).toBe(true);
  });
});
