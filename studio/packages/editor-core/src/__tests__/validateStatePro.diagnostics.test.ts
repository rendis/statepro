import { describe, expect, it } from "vitest";
import { validateStateProMachine } from "../model/validateStatePro";
import type { SerializeIssue, StateProMachine } from "../types";

const fixture = (): StateProMachine => ({
  id: "machine", canonicalName: "machine", version: "1.0.0", initials: ["U:main"],
  universes: { main: {
    id: "main", canonicalName: "main", version: "1.0.0", initial: "idle",
    realities: {
      idle: { id: "idle", type: "transition", on: { GO: [{ targets: ["done"] }] } },
      done: { id: "done", type: "final" },
    },
  } },
});

type DiagnosticCase = {
  name: string;
  change: (machine: StateProMachine) => void;
  field: string;
  message: string;
  messageKey: string;
  messageParams?: SerializeIssue["messageParams"];
};
const cases: DiagnosticCase[] = [
  { name: "id de maquina", change: m => { m.id = "bad id"; }, field: "id", message: "Machine id has invalid format", messageKey: "issue.machineIdInvalid" },
  { name: "nombre canonico", change: m => { m.canonicalName = "bad name"; }, field: "canonicalName", message: "Machine canonicalName has invalid format", messageKey: "issue.machineCanonicalInvalid" },
  { name: "universos ausentes", change: m => { m.universes = {}; }, field: "universes", message: "Machine must define at least one universe", messageKey: "issue.machineNeedsUniverse" },
  { name: "universo nulo", change: m => { m.universes.main = null as never; }, field: "universes.main", message: "Universe cannot be null", messageKey: "issue.universeNull" },
  { name: "id de universo distinto a su clave", change: m => { m.universes.main.id = "other"; }, field: "universes.main.id", message: "Universe key 'main' must match universe.id 'other'", messageKey: "issue.universeKeyMismatch", messageParams: { universeKey: "main", universeId: "other" } },
  { name: "realidades vacias", change: m => { m.universes.main.realities = {}; }, field: "universes.main.realities", message: "Universe must define at least one reality", messageKey: "issue.universeNeedsReality" },
  { name: "realidad nula", change: m => { m.universes.main.realities.idle = null as never; }, field: "universes.main.realities.idle", message: "Reality cannot be null", messageKey: "issue.realityNull" },
  { name: "id de realidad distinto a su clave", change: m => { m.universes.main.realities.idle.id = "other"; }, field: "universes.main.realities.idle.id", message: "Reality key 'idle' must match reality.id 'other'", messageKey: "issue.realityKeyMismatch", messageParams: { realityKey: "idle", realityId: "other" } },
  { name: "transicion sin flujo", change: m => { m.universes.main.realities.idle.on = { GO: [] }; }, field: "universes.main.realities.idle", message: "Transition reality must define non-empty 'on' or non-empty 'always'", messageKey: "issue.transitionRealityNeedsOnOrAlways" },
  { name: "destinos vacios", change: m => { m.universes.main.realities.idle.on!.GO[0].targets = []; }, field: "universes.main.realities.idle.on.GO[0]", message: "Transition must define at least one target", messageKey: "issue.transitionNeedsTarget" },
  { name: "segundo destino malformado", change: m => { m.universes.main.realities.idle.on!.GO[0].targets = ["done", "U:"]; }, field: "universes.main.realities.idle.on.GO[0].targets[1]", message: "Invalid target reference 'U:'", messageKey: "issue.invalidTargetReference", messageParams: { target: "U:" } },
  { name: "notify interno", change: m => { m.universes.main.realities.idle.on!.GO[0].type = "notify"; }, field: "universes.main.realities.idle.on.GO[0].targets[0]", message: "Notify transition cannot target internal realities", messageKey: "issue.notifyInternalTarget" },
  { name: "realidad interna desconocida", change: m => { m.universes.main.realities.idle.on!.GO[0].targets = ["missing"]; }, field: "universes.main.realities.idle.on.GO[0].targets[0]", message: "Unknown internal reality 'missing' in universe 'main'", messageKey: "issue.unknownInternalReality", messageParams: { realityId: "missing", universeId: "main" } },
  { name: "universo externo desconocido", change: m => { m.universes.main.realities.idle.on!.GO[0].targets = ["U:missing"]; }, field: "universes.main.realities.idle.on.GO[0].targets[0]", message: "Unknown universe 'missing'", messageKey: "issue.unknownUniverse", messageParams: { universeId: "missing" } },
  { name: "realidad externa desconocida", change: m => { m.universes.main.realities.idle.on!.GO[0].targets = ["U:main:missing"]; }, field: "universes.main.realities.idle.on.GO[0].targets[0]", message: "Unknown reality 'missing' in universe 'main'", messageKey: "issue.unknownRealityInUniverse", messageParams: { realityId: "missing", universeId: "main" } },
  { name: "initial de universo inexistente", change: m => { m.universes.main.initial = "missing"; }, field: "universes.main.initial", message: "Initial reality 'missing' does not exist in universe 'main'", messageKey: "issue.initialRealityMissing", messageParams: { realityId: "missing", universeId: "main" } },
  { name: "segunda referencia inicial interna", change: m => { m.initials = ["U:main", "done"]; }, field: "initials[1]", message: "Initial reference must be U:<universe> or U:<universe>:<reality>", messageKey: "issue.initialReferenceFormat" },
  { name: "universo inicial desconocido", change: m => { m.initials = ["U:missing"]; }, field: "initials[0]", message: "Unknown universe 'missing' in initials", messageKey: "issue.initialUnknownUniverse", messageParams: { universeId: "missing" } },
  { name: "realidad inicial desconocida", change: m => { m.initials = ["U:main:missing"]; }, field: "initials[0]", message: "Unknown reality 'missing' for universe 'main' in initials", messageKey: "issue.initialUnknownReality", messageParams: { realityId: "missing", universeId: "main" } },
  { name: "condicion sin src", change: m => { m.universes.main.realities.idle.on!.GO[0].condition = {} as never; }, field: "universes.main.realities.idle.on.GO[0].condition.src", message: "Condition must have src", messageKey: "issue.conditionMissingSrc" },
  { name: "always con condicion incompleta", change: m => { m.universes.main.realities.idle.always = [{ targets: ["done"], condition: {} as never }]; }, field: "universes.main.realities.idle.always[0].condition.src", message: "Condition must have src", messageKey: "issue.conditionMissingSrc" },
  { name: "segundo always sin destino", change: m => { m.universes.main.realities.idle.always = [{ targets: ["done"] }, { targets: [] }]; }, field: "universes.main.realities.idle.always[1]", message: "Transition must define at least one target", messageKey: "issue.transitionNeedsTarget" },
];

describe("Diagnosticos observables del validador", () => {
  it.each(cases)("preserva campo, traduccion y detalle: $name", ({ change, field, message, messageKey, messageParams }) => {
    const machine = fixture();
    change(machine);
    const result = validateStateProMachine(machine);
    expect(result.canExport).toBe(false);
    expect(result.issues.filter(issue => issue.messageKey === messageKey)).toEqual([{
      code: "SEMANTIC_ERROR", severity: "error", field, message, messageKey,
      ...(messageParams ? { messageParams } : {}),
    }]);
  });

  it("informa identificadores invalidos de claves y permite varias ramas efectivas", () => {
    const machine = fixture();
    const universe = machine.universes.main;
    universe.id = "bad universe";
    universe.realities["bad reality"] = { id: "bad reality", type: "final" };
    machine.universes = { "bad universe": universe };
    machine.initials = [];
    expect(validateStateProMachine(machine).issues).toEqual(expect.arrayContaining([
      { code: "SEMANTIC_ERROR", severity: "error", field: "universes.bad universe", message: "Universe key has invalid format", messageKey: "issue.universeKeyInvalid" },
      { code: "SEMANTIC_ERROR", severity: "error", field: "universes.bad universe.realities.bad reality", message: "Reality key has invalid format", messageKey: "issue.realityKeyInvalid" },
    ]));
    const valid = fixture();
    valid.universes.main.realities.idle.on = { EMPTY: [], GO: [{ targets: ["U:main:done"] }] };
    valid.initials = ["U:main:idle"];
    expect(validateStateProMachine(valid)).toEqual({ canExport: true, issues: [] });
  });

  it("conserva advertencias y errores distintos y deduplica mensajes localizables", () => {
    const base: SerializeIssue = { code: "SEMANTIC_ERROR", severity: "warning", field: "id", message: "warning" };
    const localized: SerializeIssue = { ...base, messageKey: "issue.unknownUniverse", messageParams: { universeId: "one" } };
    const seeds: SerializeIssue[] = [base, { ...base }, localized, { ...localized, message: "translated warning" }, { ...localized, messageParams: { universeId: "two" } }, { ...base, field: "version" }, { ...base, message: "another warning" }];
    const before = structuredClone(seeds);
    expect(validateStateProMachine(fixture(), seeds)).toEqual({ canExport: true, issues: [base, localized, seeds[4], seeds[5], seeds[6]] });
    expect(seeds).toEqual(before);
    const error = { ...base, severity: "error" as const };
    expect(validateStateProMachine(fixture(), [base, error]).issues).toEqual([base, error]);
    expect(validateStateProMachine(fixture(), [base, error]).canExport).toBe(false);
  });

  it("traduce rutas JSON Pointer con slash, tilde e indices", () => {
    const machine = fixture();
    machine.universes["a/b~c"] = { id: "a/b~c", canonicalName: "bad", version: "1.0.0", realities: 4 as never };
    machine.universes.main.realities.idle.on!.GO[0].targets = [9 as never];
    expect(validateStateProMachine(machine).issues).toEqual(expect.arrayContaining([
      { code: "SCHEMA_ERROR", severity: "error", field: "universes.a/b~c.realities", message: "Expected type 'object'" },
      { code: "SCHEMA_ERROR", severity: "error", field: "universes.main.realities.idle.on.GO.0.targets.0", message: "Expected type 'string'" },
    ]));
  });

  it("identifica propiedades requeridas, tipos y restricciones sin ruido de anyOf", () => {
    const machine = fixture();
    delete (machine as Partial<StateProMachine>).canonicalName;
    machine.id = "bad id";
    machine.universes.main.realities.idle.on!.GO[0].targets = [];
    const schemas = validateStateProMachine(machine).issues.filter(issue => issue.code === "SCHEMA_ERROR");
    expect(schemas).toEqual(expect.arrayContaining([
      { code: "SCHEMA_ERROR", severity: "error", field: "machine", message: "Missing required property 'canonicalName'" },
      expect.objectContaining({ field: "id", message: expect.stringContaining("pattern") }),
      expect.objectContaining({ field: "universes.main.realities.idle.on.GO.0.targets", message: expect.stringContaining("items") }),
    ]));
    expect(schemas.some(issue => /anyOf|oneOf|allOf/.test(issue.message))).toBe(false);
  });
});
