import Ajv2020 from "ajv/dist/2020";

import type {
  SerializeIssue,
  StateProMachine,
  StateProReality,
  StateProTransition,
} from "../types";
import schema from "../schema/quantum-machine.schema.json";
import {
  isValidIdentifier,
  parseTargetReference,
} from "../utils";

const ajv = new Ajv2020({
  allErrors: true,
  strict: false,
  allowUnionTypes: true,
});

const validateSchema = ajv.compile(schema as object);

type AjvSchemaError = {
  instancePath?: string;
  schemaPath?: string;
  keyword: string;
  message?: string;
  params?: Record<string, unknown>;
};

const toFieldPath = (jsonPointer: string): string => {
  if (!jsonPointer) {
    return "machine";
  }

  return jsonPointer
    .split("/")
    .filter(Boolean)
    .map((segment) => segment.replace(/~1/g, "/").replace(/~0/g, "~"))
    .join(".");
};

const dedupeIssues = (issues: SerializeIssue[]): SerializeIssue[] => {
  const seen = new Set<string>();

  return issues.filter((issue) => {
    const messageIdentity = issue.messageKey
      ? `${issue.messageKey}:${JSON.stringify(issue.messageParams || {})}`
      : issue.message;
    const key = `${issue.code}|${issue.severity}|${issue.field}|${messageIdentity}`;

    if (seen.has(key)) {
      return false;
    }

    seen.add(key);
    return true;
  });
};

const isLowSignalSchemaError = (error: AjvSchemaError): boolean => {
  const lowSignalKeywords = new Set(["anyOf", "oneOf", "allOf", "if", "then", "else", "not"]);
  if (lowSignalKeywords.has(error.keyword)) {
    return true;
  }

  if (error.keyword === "required" && typeof error.schemaPath === "string") {
    if (
      error.schemaPath.includes("/anyOf/") ||
      error.schemaPath.includes("/oneOf/") ||
      error.schemaPath.includes("/allOf/")
    ) {
      return true;
    }
  }

  return false;
};

const toSchemaIssueMessage = (error: AjvSchemaError): string => {
  if (error.keyword === "required") {
    const missingProperty = error.params?.missingProperty;
    if (typeof missingProperty === "string" && missingProperty.trim().length > 0) {
      return `Missing required property '${missingProperty}'`;
    }
  }

  if (error.keyword === "additionalProperties") {
    const additionalProperty = error.params?.additionalProperty;
    if (typeof additionalProperty === "string" && additionalProperty.trim().length > 0) {
      return `Unexpected property '${additionalProperty}'`;
    }
  }

  if (error.keyword === "type") {
    const expectedType = error.params?.type;
    if (typeof expectedType === "string" && expectedType.trim().length > 0) {
      return `Expected type '${expectedType}'`;
    }
  }

  return error.message || "Schema validation error";
};

const hasEffectiveTransitionFlow = (reality: StateProReality): boolean => {
  if ((reality.always?.length || 0) > 0) {
    return true;
  }

  if (!reality.on || typeof reality.on !== "object") {
    return false;
  }

  return Object.values(reality.on).some((transitionGroup) => (transitionGroup?.length || 0) > 0);
};

const hasNonEmptyUniversalConstants = (
  constants: StateProMachine["universalConstants"] | undefined,
): boolean => {
  if (!constants) {
    return false;
  }

  return (
    (constants.entryActions?.length || 0) > 0 ||
    (constants.exitActions?.length || 0) > 0 ||
    (constants.entryInvokes?.length || 0) > 0 ||
    (constants.exitInvokes?.length || 0) > 0 ||
    (constants.actionsOnTransition?.length || 0) > 0 ||
    (constants.invokesOnTransition?.length || 0) > 0
  );
};

const validateTransitionTargetsSemantics = (
  machine: StateProMachine,
  universeId: string,
  realityId: string,
  transitionPath: string,
  transition: StateProTransition,
  issues: SerializeIssue[],
): void => {
  if (!transition || !Array.isArray(transition.targets) || transition.targets.length === 0) {
    issues.push({
      code: "SEMANTIC_ERROR",
      severity: "error",
      field: transitionPath,
      message: "Transition must define at least one target",
      messageKey: "issue.transitionNeedsTarget",
    });
    return;
  }

  const sourceUniverse = machine.universes[universeId];

  transition.targets.forEach((target, targetIndex) => {
    const targetPath = `${transitionPath}.targets[${targetIndex}]`;
    const parsed = parseTargetReference(target);
    if (!parsed) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: targetPath,
        message: `Invalid target reference '${target}'`,
        messageKey: "issue.invalidTargetReference",
        messageParams: { target },
      });
      return;
    }

    if (transition.type === "notify" && parsed.kind === "reality") {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: targetPath,
        message: "Notify transition cannot target internal realities",
        messageKey: "issue.notifyInternalTarget",
      });
    }

    if (parsed.kind === "reality") {
      const targetRealityId = parsed.realityId || "";
      if (!sourceUniverse?.realities?.[targetRealityId]) {
        issues.push({
          code: "SEMANTIC_ERROR",
          severity: "error",
          field: targetPath,
          message: `Unknown internal reality '${targetRealityId}' in universe '${universeId}'`,
          messageKey: "issue.unknownInternalReality",
          messageParams: {
            realityId: targetRealityId,
            universeId,
          },
        });
      }
      return;
    }

    const targetUniverseId = parsed.universeId || "";
    const targetUniverse = machine.universes[targetUniverseId];
    if (!targetUniverse) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: targetPath,
        message: `Unknown universe '${targetUniverseId}'`,
        messageKey: "issue.unknownUniverse",
        messageParams: { universeId: targetUniverseId },
      });
      return;
    }

    if (parsed.kind === "universeReality") {
      const targetRealityId = parsed.realityId || "";
      if (!targetUniverse.realities?.[targetRealityId]) {
        issues.push({
          code: "SEMANTIC_ERROR",
          severity: "error",
          field: targetPath,
          message: `Unknown reality '${targetRealityId}' in universe '${targetUniverseId}'`,
          messageKey: "issue.unknownRealityInUniverse",
          messageParams: {
            realityId: targetRealityId,
            universeId: targetUniverseId,
          },
        });
      }
    }
  });
};

const validateSemantic = (machine: StateProMachine): SerializeIssue[] => {
  const issues: SerializeIssue[] = [];

  if (!machine?.universes || Object.keys(machine.universes).length === 0) {
    issues.push({
      code: "SEMANTIC_ERROR",
      severity: "error",
      field: "universes",
      message: "Machine must define at least one universe",
      messageKey: "issue.machineNeedsUniverse",
    });
    return issues;
  }

  Object.entries(machine.universes).forEach(([universeKey, universe]) => {
    if (!universe) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `universes.${universeKey}`,
        message: "Universe cannot be null",
        messageKey: "issue.universeNull",
      });
      return;
    }

    if (universe.id !== universeKey) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `universes.${universeKey}.id`,
        message: `Universe key '${universeKey}' must match universe.id '${universe.id}'`,
        messageKey: "issue.universeKeyMismatch",
        messageParams: {
          universeKey,
          universeId: universe.id,
        },
      });
    }

    if (hasNonEmptyUniversalConstants(universe.universalConstants)) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "warning",
        field: `universes.${universeKey}.universalConstants`,
        message: "Universe universalConstants are defined but current runtime ignores them",
        messageKey: "issue.universeConstantsRuntimeIgnored",
      });
    }

    if (!universe.realities || Object.keys(universe.realities).length === 0) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `universes.${universeKey}.realities`,
        message: "Universe must define at least one reality",
        messageKey: "issue.universeNeedsReality",
      });
      return;
    }

    Object.entries(universe.realities).forEach(([realityKey, reality]) => {
      if (!reality) {
        issues.push({
          code: "SEMANTIC_ERROR",
          severity: "error",
          field: `universes.${universeKey}.realities.${realityKey}`,
          message: "Reality cannot be null",
          messageKey: "issue.realityNull",
        });
        return;
      }

      if (reality.id !== realityKey) {
        issues.push({
          code: "SEMANTIC_ERROR",
          severity: "error",
          field: `universes.${universeKey}.realities.${realityKey}.id`,
          message: `Reality key '${realityKey}' must match reality.id '${reality.id}'`,
          messageKey: "issue.realityKeyMismatch",
          messageParams: {
            realityKey,
            realityId: reality.id,
          },
        });
      }

      if (reality.type === "transition" && !hasEffectiveTransitionFlow(reality)) {
        issues.push({
          code: "SEMANTIC_ERROR",
          severity: "error",
          field: `universes.${universeKey}.realities.${realityKey}`,
          message: "Transition reality must define non-empty 'on' or non-empty 'always'",
          messageKey: "issue.transitionRealityNeedsOnOrAlways",
        });
      }

      (reality.always || []).forEach((transition, transitionIndex) => {
        validateTransitionTargetsSemantics(
          machine,
          universeKey,
          realityKey,
          `universes.${universeKey}.realities.${realityKey}.always[${transitionIndex}]`,
          transition,
          issues,
        );
      });

      Object.entries(reality.on || {}).forEach(([eventName, transitions]) => {
        (transitions || []).forEach((transition, transitionIndex) => {
          validateTransitionTargetsSemantics(
            machine,
            universeKey,
            realityKey,
            `universes.${universeKey}.realities.${realityKey}.on.${eventName}[${transitionIndex}]`,
            transition,
            issues,
          );
        });
      });
    });

    if (universe.initial && !universe.realities[universe.initial]) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `universes.${universeKey}.initial`,
        message: `Initial reality '${universe.initial}' does not exist in universe '${universeKey}'`,
        messageKey: "issue.initialRealityMissing",
        messageParams: {
          realityId: universe.initial,
          universeId: universeKey,
        },
      });
    }
  });

  (machine.initials || []).forEach((initial, initialIndex) => {
    const parsed = parseTargetReference(initial);
    const field = `initials[${initialIndex}]`;

    if (!parsed || parsed.kind === "reality") {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field,
        message: "Initial reference must be U:<universe> or U:<universe>:<reality>",
        messageKey: "issue.initialReferenceFormat",
      });
      return;
    }

    const universeId = parsed.universeId || "";
    const universe = machine.universes[universeId];
    if (!universe) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field,
        message: `Unknown universe '${universeId}' in initials`,
        messageKey: "issue.initialUnknownUniverse",
        messageParams: {
          universeId,
        },
      });
      return;
    }

    if (parsed.kind === "universeReality") {
      const realityId = parsed.realityId || "";
      if (!universe.realities[realityId]) {
        issues.push({
          code: "SEMANTIC_ERROR",
          severity: "error",
          field,
          message: `Unknown reality '${realityId}' for universe '${universeId}' in initials`,
          messageKey: "issue.initialUnknownReality",
          messageParams: {
            realityId,
            universeId,
          },
        });
      }
    }
  });

  return issues;
};

const validateTransitionShape = (transition: StateProTransition, fieldPrefix: string): SerializeIssue[] => {
  const issues: SerializeIssue[] = [];

  if (transition.condition && !transition.condition.src) {
    issues.push({
      code: "SEMANTIC_ERROR",
      severity: "error",
      field: `${fieldPrefix}.condition.src`,
      message: "Condition must have src",
      messageKey: "issue.conditionMissingSrc",
    });
  }

  if (Array.isArray(transition.conditions) && transition.conditions.length > 0) {
    const seen = new Set<string>();
    const duplicates = new Set<string>();

    transition.conditions.forEach((condition) => {
      const src = condition?.src || "";
      if (!src) {
        return;
      }
      if (seen.has(src)) {
        duplicates.add(src);
        return;
      }
      seen.add(src);
    });

    duplicates.forEach((src) => {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `${fieldPrefix}.conditions`,
        message: `Duplicated condition '${src}' in conditions`,
        messageKey: "issue.duplicatedTransitionCondition",
        messageParams: { src },
      });
    });
  }

  return issues;
};

const validateIdentifiers = (machine: StateProMachine): SerializeIssue[] => {
  const issues: SerializeIssue[] = [];

  if (!isValidIdentifier(machine.id)) {
    issues.push({
      code: "SEMANTIC_ERROR",
      severity: "error",
      field: "id",
      message: "Machine id has invalid format",
      messageKey: "issue.machineIdInvalid",
    });
  }

  if (!isValidIdentifier(machine.canonicalName)) {
    issues.push({
      code: "SEMANTIC_ERROR",
      severity: "error",
      field: "canonicalName",
      message: "Machine canonicalName has invalid format",
      messageKey: "issue.machineCanonicalInvalid",
    });
  }

  Object.entries(machine.universes || {}).forEach(([universeId, universe]) => {
    if (!isValidIdentifier(universeId)) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `universes.${universeId}`,
        message: "Universe key has invalid format",
        messageKey: "issue.universeKeyInvalid",
      });
    }

    Object.entries(universe.realities || {}).forEach(([realityId, reality]) => {
      if (!isValidIdentifier(realityId)) {
        issues.push({
          code: "SEMANTIC_ERROR",
          severity: "error",
          field: `universes.${universeId}.realities.${realityId}`,
          message: "Reality key has invalid format",
          messageKey: "issue.realityKeyInvalid",
        });
      }

      (reality.always || []).forEach((transition, transitionIndex) => {
        issues.push(
          ...validateTransitionShape(
            transition,
            `universes.${universeId}.realities.${realityId}.always[${transitionIndex}]`,
          ),
        );
      });

      Object.entries(reality.on || {}).forEach(([eventName, transitions]) => {
        transitions.forEach((transition, transitionIndex) => {
          issues.push(
            ...validateTransitionShape(
              transition,
              `universes.${universeId}.realities.${realityId}.on.${eventName}[${transitionIndex}]`,
            ),
          );
        });
      });
    });
  });

  return issues;
};

export const validateStateProMachine = (
  machine: StateProMachine,
  seedIssues: SerializeIssue[] = [],
): { issues: SerializeIssue[]; canExport: boolean } => {
  const rawIssues: SerializeIssue[] = [...seedIssues];

  const schemaOk = validateSchema(machine);
  if (!schemaOk && Array.isArray(validateSchema.errors)) {
    validateSchema.errors.forEach((rawError) => {
      const error = rawError as AjvSchemaError;
      if (isLowSignalSchemaError(error)) {
        return;
      }

      rawIssues.push({
        code: "SCHEMA_ERROR",
        severity: "error",
        field: toFieldPath(error.instancePath || ""),
        message: toSchemaIssueMessage(error),
      });
    });
  }

  rawIssues.push(...validateIdentifiers(machine));
  rawIssues.push(...validateSemantic(machine));

  const issues = dedupeIssues(rawIssues);

  const canExport = issues.every((issue) => issue.severity !== "error");

  return {
    issues,
    canExport,
  };
};
