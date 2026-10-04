import { describe, expect, it } from "vitest";

import { validateStateProMachine } from "../model/validateStatePro";
import { isValidIdentifier, parseTargetReference } from "../utils/references";
import type { StateProMachine } from "../types";

const validMachine = (): StateProMachine => ({
  id: "machine", canonicalName: "machine", version: "1.0.0", initials: ["U:main"],
  universes: { main: {
    id: "main", canonicalName: "main", version: "1.0.0", initial: "idle",
    realities: {
      idle: { id: "idle", type: "transition", on: { GO: [{ targets: ["done"] }] } },
      done: { id: "done", type: "final" },
    },
  } },
});

describe("Validacion de documentos externos", () => {
  it.each([null, false, 42, "machine", [], { universes: { main: null } },
    { universes: { main: { realities: { idle: null } } } },
    { universes: { main: { realities: { idle: { always: {}, on: { GO: {} } } } } } },
    { universes: { main: { realities: { idle: { always: [null] } } } }, initials: {} },
  ])("devuelve errores estructurados para JSON malformado %#", (document) => {
    const result = validateStateProMachine(document as unknown as StateProMachine);
    expect(result.canExport).toBe(false);
    expect(result.issues.some(issue => issue.severity === "error")).toBe(true);
  });

  it.each(["constructor", "toString"])("no resuelve %s desde el prototipo como universo o realidad", (name) => {
    const machine = validMachine();
    machine.initials = [`U:${name}`];
    machine.universes.main.realities.idle.on!.GO[0].targets = [name, `U:${name}`];
    const result = validateStateProMachine(machine);
    expect(result.canExport).toBe(false);
    expect(result.issues).toEqual(expect.arrayContaining([
      expect.objectContaining({ messageKey: "issue.initialUnknownUniverse", messageParams: { universeId: name } }),
      expect.objectContaining({ messageKey: "issue.unknownInternalReality", messageParams: { realityId: name, universeId: "main" } }),
      expect.objectContaining({ messageKey: "issue.unknownUniverse", messageParams: { universeId: name } }),
    ]));
  });

  it("acepta nombres propios que coinciden con propiedades del prototipo", () => {
    const machine = validMachine();
    const universe = machine.universes.main;
    universe.id = "constructor";
    universe.canonicalName = "constructor";
    Object.defineProperty(universe.realities, "constructor", {
      value: { id: "constructor", type: "final" }, enumerable: true,
    });
    universe.realities.idle.on!.GO[0].targets = ["constructor"];
    machine.universes = { constructor: universe };
    machine.initials = ["U:constructor"];
    expect(validateStateProMachine(machine)).toEqual({ issues: [], canExport: true });
  });

  it.each([undefined, null, true, 1, { toString: null, valueOf: null }])("rechaza referencias e identificadores que no son strings %#", (value) => {
    expect(isValidIdentifier(value as string)).toBe(false);
    expect(parseTargetReference(value as string)).toBeNull();
    const machine = validMachine();
    machine.universes.main.realities.idle.on!.GO[0].targets = [value as string];
    expect(validateStateProMachine(machine).canExport).toBe(false);
  });
});
