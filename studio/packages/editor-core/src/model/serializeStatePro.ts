import type {
  BehaviorRef,
  EditorState,
  EditorTransition,
  JsonObject,
  MetadataPackBinding,
  MetadataPackBindingMap,
  MetadataPackRegistry,
  MetadataScope,
  RealityType,
  SerializeIssue,
  SerializeResult,
  StateProMachine,
  StateProReality,
  StateProTransition,
  UniversalConstants,
} from "../types";
import { cleanBehaviorRef, tryParseJsonObject } from "../utils/behavior";
import {
  buildMachineEntityRef,
  buildMetadataPackRegistryIndex,
  buildRealityEntityRef,
  buildTransitionEntityRef,
  buildUniverseEntityRef,
  collectOwnedPointersFromPack,
  deepMergeJsonObjects,
  findBindingOwnershipCollisionsDetailed,
  getBindingsForEntity,
  listManualConflictingPointers,
  mergePackBindingsToMetadata,
  normalizeTransitionsOrder,
  pointerTokens,
  validateBindingWithPack,
} from "../utils";
import { stripStickyNoteKeysFromMetadata } from "./stickyNotesMetadata";
import { validateStateProMachine } from "./validateStatePro";

const cleanupUndefinedFields = <T extends Record<string, unknown>>(obj: T): T => {
  Object.keys(obj).forEach((key) => {
    if (obj[key] === undefined) {
      delete obj[key];
    }
  });
  return obj;
};

const cleanBehaviorRefs = (behaviors: BehaviorRef[] = []): BehaviorRef[] => {
  return behaviors
    .filter((behavior) => behavior && behavior.src)
    .map((behavior) => cleanBehaviorRef(behavior));
};

const normalizeTransitionConditions = (
  condition: BehaviorRef | undefined,
  conditions: BehaviorRef[] | undefined,
): BehaviorRef[] | undefined => {
  const mergedConditions = [
    ...(condition ? [condition] : []),
    ...(conditions || []),
  ];
  const cleanedConditions = cleanBehaviorRefs(mergedConditions);
  if (cleanedConditions.length === 0) {
    return undefined;
  }

  const seen = new Set<string>();
  const dedupedConditions = cleanedConditions.filter((candidate) => {
    const src = candidate?.src || "";
    if (!src || seen.has(src)) {
      return false;
    }
    seen.add(src);
    return true;
  });

  return dedupedConditions.length > 0 ? dedupedConditions : undefined;
};

const compactUC = (uc?: UniversalConstants): Partial<UniversalConstants> | undefined => {
  if (!uc) return undefined;

  const compact: Partial<UniversalConstants> = {};
  if (uc.entryActions.length > 0) compact.entryActions = cleanBehaviorRefs(uc.entryActions);
  if (uc.exitActions.length > 0) compact.exitActions = cleanBehaviorRefs(uc.exitActions);
  if (uc.entryInvokes.length > 0) compact.entryInvokes = cleanBehaviorRefs(uc.entryInvokes);
  if (uc.exitInvokes.length > 0) compact.exitInvokes = cleanBehaviorRefs(uc.exitInvokes);
  if (uc.actionsOnTransition.length > 0) {
    compact.actionsOnTransition = cleanBehaviorRefs(uc.actionsOnTransition);
  }
  if (uc.invokesOnTransition.length > 0) {
    compact.invokesOnTransition = cleanBehaviorRefs(uc.invokesOnTransition);
  }

  return Object.keys(compact).length > 0 ? compact : undefined;
};

const toRealityType = (realityType: RealityType):
  | "transition"
  | "final"
  | "unsuccessfulFinal" => {
  if (realityType === "success") {
    return "final";
  }

  if (realityType === "error") {
    return "unsuccessfulFinal";
  }

  return "transition";
};

const pointerToDotPath = (pointer: string): string => {
  const tokens = pointerTokens(pointer);
  if (tokens.length === 0) {
    return "";
  }
  return `.${tokens.join(".")}`;
};

const collectDuplicatedValues = (values: string[]): string[] => {
  const seen = new Set<string>();
  const duplicates = new Set<string>();

  values.filter(Boolean).forEach((value) => {
    if (seen.has(value)) {
      duplicates.add(value);
      return;
    }
    seen.add(value);
  });

  return Array.from(duplicates);
};

interface ResolveMetadataArgs {
  scope: MetadataScope;
  entityRef: string;
  rawMetadata: string | undefined;
  fieldPrefix: string;
  registry: MetadataPackRegistry;
  bindings: MetadataPackBindingMap;
  issues: SerializeIssue[];
}

const resolveMetadataForEntity = ({
  scope,
  entityRef,
  rawMetadata,
  fieldPrefix,
  registry,
  bindings,
  issues,
}: ResolveMetadataArgs): JsonObject => {
  const parsedManual = tryParseJsonObject(rawMetadata);
  if (!parsedManual.ok && parsedManual.error) {
    issues.push({
      code: "INVALID_JSON",
      severity: "error",
      field: `${fieldPrefix}.metadata`,
      message: parsedManual.error.message,
      messageKey: parsedManual.error.messageKey,
      messageParams: parsedManual.error.messageParams,
    });
  }

  const manualMetadata = structuredClone(parsedManual.value);
  const manualMetadataWithoutStickyNotes = stripStickyNoteKeysFromMetadata(manualMetadata);
  if (Object.prototype.hasOwnProperty.call(manualMetadataWithoutStickyNotes, "__studio")) {
    delete manualMetadataWithoutStickyNotes.__studio;
  }

  const registryIndex = buildMetadataPackRegistryIndex(registry);
  const entityBindings = getBindingsForEntity(bindings, scope, entityRef);

  const collisions = findBindingOwnershipCollisionsDetailed(entityBindings, registryIndex);
  collisions.forEach((collision) => {
    issues.push({
      code: "SEMANTIC_ERROR",
      severity: "error",
      field: `${fieldPrefix}.metadataPacks`,
      message:
        `Ownership collision (${collision.relation}): ` +
        `binding '${collision.leftBindingId}' (pack '${collision.leftPackId}', path '${collision.leftPointer}') ` +
        `vs binding '${collision.rightBindingId}' (pack '${collision.rightPackId}', path '${collision.rightPointer}').`,
      messageKey: "issue.packOwnershipCollision",
      messageParams: {
        relation: collision.relation,
        leftBindingId: collision.leftBindingId,
        leftPackId: collision.leftPackId,
        leftPointer: collision.leftPointer,
        rightBindingId: collision.rightBindingId,
        rightPackId: collision.rightPackId,
        rightPointer: collision.rightPointer,
      },
    });
  });

  const validBindings: MetadataPackBinding[] = [];
  entityBindings.forEach((binding) => {
    const pack = registryIndex.get(binding.packId);
    if (!pack) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `${fieldPrefix}.metadataPacks.${binding.id}.packId`,
        message: `Metadata pack '${binding.packId}' does not exist.`,
        messageKey: "issue.packDoesNotExist",
        messageParams: { packId: binding.packId },
      });
      return;
    }

    if (!pack.scopes.includes(scope)) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `${fieldPrefix}.metadataPacks.${binding.id}.scope`,
        message: `Pack '${pack.id}' cannot be applied to scope '${scope}'.`,
        messageKey: "issue.packScopeNotApplicable",
        messageParams: {
          packId: pack.id,
          scope,
        },
      });
      return;
    }

    const valueErrors = validateBindingWithPack(binding, pack);
    valueErrors.forEach((errorMessage, index) => {
      issues.push({
        code: "SCHEMA_ERROR",
        severity: "error",
        field: `${fieldPrefix}.metadataPacks.${binding.id}.values[${index}]`,
        message: errorMessage.message,
        messageKey: "issue.invalidValueDetailed",
        messageParams: {
          path: errorMessage.path,
          detail: errorMessage.detail,
        },
      });
    });

    const ownedPointers = collectOwnedPointersFromPack(pack);
    const manualConflicts = listManualConflictingPointers(
      manualMetadataWithoutStickyNotes,
      ownedPointers,
    );
    manualConflicts.forEach((pointer) => {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "warning",
        field: `${fieldPrefix}.metadata${pointerToDotPath(pointer)}`,
        message: `Manual metadata value was overridden by pack '${pack.id}' for owned path '${pointer}'.`,
        messageKey: "issue.manualMetadataOverridden",
        messageParams: {
          packId: pack.id,
          pointer,
        },
      });
    });

    validBindings.push(binding);
  });

  const packMetadata = mergePackBindingsToMetadata(validBindings, registryIndex);
  return deepMergeJsonObjects(manualMetadataWithoutStickyNotes, packMetadata);
};

const buildTransitionOutput = (
  transition: EditorTransition,
  resolvedMetadata: JsonObject,
): StateProTransition => {
  const output: StateProTransition = cleanupUndefinedFields({
    targets: transition.targets.filter(Boolean),
    type: transition.type,
    conditions: normalizeTransitionConditions(transition.condition, transition.conditions),
    actions: transition.actions.length > 0 ? cleanBehaviorRefs(transition.actions) : undefined,
    invokes: transition.invokes.length > 0 ? cleanBehaviorRefs(transition.invokes) : undefined,
    description: transition.description || undefined,
    metadata: Object.keys(resolvedMetadata).length > 0 ? resolvedMetadata : undefined,
  });

  return output;
};

const sanitizeBehaviorRefList = (behaviors: BehaviorRef[] | undefined): BehaviorRef[] | undefined => {
  if (!behaviors) {
    return undefined;
  }

  return cleanBehaviorRefs(behaviors);
};

const sanitizeUniversalConstants = (
  constants: Partial<UniversalConstants> | undefined,
): Partial<UniversalConstants> | undefined => {
  if (!constants) {
    return undefined;
  }

  return {
    ...constants,
    entryActions: sanitizeBehaviorRefList(constants.entryActions),
    exitActions: sanitizeBehaviorRefList(constants.exitActions),
    entryInvokes: sanitizeBehaviorRefList(constants.entryInvokes),
    exitInvokes: sanitizeBehaviorRefList(constants.exitInvokes),
    actionsOnTransition: sanitizeBehaviorRefList(constants.actionsOnTransition),
    invokesOnTransition: sanitizeBehaviorRefList(constants.invokesOnTransition),
  };
};

const sanitizeTransitionBehaviorRefs = (transition: StateProTransition): StateProTransition => {
  return {
    ...transition,
    condition: undefined,
    conditions: normalizeTransitionConditions(transition.condition, transition.conditions),
    actions: sanitizeBehaviorRefList(transition.actions),
    invokes: sanitizeBehaviorRefList(transition.invokes),
  };
};

const sanitizeMachineBehaviorRefs = (machine: StateProMachine): StateProMachine => {
  const sanitizedUniverses = Object.fromEntries(
    Object.entries(machine.universes || {}).map(([universeId, universe]) => {
      const sanitizedRealities = Object.fromEntries(
        Object.entries(universe.realities || {}).map(([realityId, reality]) => {
          const sanitizedOn =
            reality.on == null
              ? reality.on
              : Object.fromEntries(
                  Object.entries(reality.on).map(([eventName, transitions]) => [
                    eventName,
                    (Array.isArray(transitions) ? transitions : []).map(
                      sanitizeTransitionBehaviorRefs,
                    ),
                  ]),
                );

          return [
            realityId,
            {
              ...reality,
              observers: sanitizeBehaviorRefList(reality.observers),
              entryActions: sanitizeBehaviorRefList(reality.entryActions),
              exitActions: sanitizeBehaviorRefList(reality.exitActions),
              entryInvokes: sanitizeBehaviorRefList(reality.entryInvokes),
              exitInvokes: sanitizeBehaviorRefList(reality.exitInvokes),
              always: reality.always?.map(sanitizeTransitionBehaviorRefs),
              on: sanitizedOn,
            },
          ];
        }),
      );

      return [
        universeId,
        {
          ...universe,
          universalConstants: sanitizeUniversalConstants(universe.universalConstants),
          realities: sanitizedRealities,
        },
      ];
    }),
  ) as StateProMachine["universes"];

  return {
    ...machine,
    universalConstants: sanitizeUniversalConstants(machine.universalConstants),
    universes: sanitizedUniverses,
  };
};

const validateBindingEntityRefs = (
  bindings: MetadataPackBindingMap,
  knownRefs: Record<MetadataScope, Set<string>>,
  issues: SerializeIssue[],
): void => {
  const scopes: MetadataScope[] = ["machine", "universe", "reality", "transition"];
  scopes.forEach((scope) => {
    bindings[scope].forEach((binding) => {
      if (knownRefs[scope].has(binding.entityRef)) {
        return;
      }
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `metadataPacks.${scope}.${binding.id}.entityRef`,
        message: `Unknown entity reference '${binding.entityRef}' for scope '${scope}'.`,
        messageKey: "issue.unknownEntityReference",
        messageParams: {
          entityRef: binding.entityRef,
          scope,
        },
      });
    });
  });
};

export const serializeStatePro = (state: EditorState): SerializeResult => {
  const issues: SerializeIssue[] = [];
  const {
    nodes,
    transitions,
    machineConfig,
    metadataPackRegistry,
    metadataPackBindings,
    lastImportedMachine,
    isDirtyFromImport,
  } = state;

  if (lastImportedMachine && !isDirtyFromImport) {
    const snapshotMachine = sanitizeMachineBehaviorRefs(lastImportedMachine);
    const validation = validateStateProMachine(snapshotMachine);
    return {
      machine: snapshotMachine,
      issues: validation.issues,
      canExport: validation.canExport,
    };
  }

  const orderedTransitions = normalizeTransitionsOrder(transitions);
  const universeNodes = nodes.filter((node) => node.type === "universe");
  const realityNodes = nodes.filter((node) => node.type === "reality");
  const universeByNodeId = new Map(universeNodes.map((universe) => [universe.id, universe]));
  const realitiesByNodeId = new Map(realityNodes.map((reality) => [reality.id, reality]));

  const knownRefs: Record<MetadataScope, Set<string>> = {
    machine: new Set([buildMachineEntityRef()]),
    universe: new Set(universeNodes.map((universe) => buildUniverseEntityRef(universe.data.id))),
    reality: new Set(
      realityNodes
        .map((reality) => {
          const parentUniverse = universeByNodeId.get(reality.data.universeId);
          if (!parentUniverse) {
            return null;
          }
          return buildRealityEntityRef(parentUniverse.data.id, reality.data.id);
        })
        .filter((value): value is string => Boolean(value)),
    ),
    transition: new Set(
      orderedTransitions
        .map((transition) => {
          const sourceReality = realitiesByNodeId.get(transition.sourceRealityId);
          if (!sourceReality) {
            return null;
          }
          const parentUniverse = universeByNodeId.get(sourceReality.data.universeId);
          if (!parentUniverse) {
            return null;
          }
          return buildTransitionEntityRef(
            parentUniverse.data.id,
            sourceReality.data.id,
            transition.triggerKind,
            transition.eventName,
            transition.order,
          );
        })
        .filter((value): value is string => Boolean(value)),
    ),
  };

  validateBindingEntityRefs(metadataPackBindings, knownRefs, issues);

  const machineMetadata = resolveMetadataForEntity({
    scope: "machine",
    entityRef: buildMachineEntityRef(),
    rawMetadata: machineConfig.metadata,
    fieldPrefix: "machine",
    registry: metadataPackRegistry,
    bindings: metadataPackBindings,
    issues,
  });

  const machine: StateProMachine = {
    id: machineConfig.id,
    canonicalName: machineConfig.canonicalName,
    version: machineConfig.version,
    description: machineConfig.description || undefined,
    initials: machineConfig.initials.length > 0 ? [...machineConfig.initials] : undefined,
    universalConstants: compactUC(machineConfig.universalConstants),
    metadata: Object.keys(machineMetadata).length > 0 ? machineMetadata : undefined,
    universes: {},
  };

  const duplicatedUniverseIds = collectDuplicatedValues(
    universeNodes.map((universe) => universe.data.id),
  );
  duplicatedUniverseIds.forEach((universeId) => {
    issues.push({
      code: "SEMANTIC_ERROR",
      severity: "error",
      field: "machine.universes",
      message: `Duplicated universe id '${universeId}'.`,
      messageKey: "issue.duplicatedUniverseId",
      messageParams: { universeId },
    });
  });

  const duplicatedUniverseCanonicalNames = collectDuplicatedValues(
    universeNodes.map((universe) => universe.data.canonicalName || universe.data.id),
  );
  duplicatedUniverseCanonicalNames.forEach((canonicalName) => {
    issues.push({
      code: "SEMANTIC_ERROR",
      severity: "error",
      field: "machine.universes",
      message: `Duplicated universe canonicalName '${canonicalName}'.`,
      messageKey: "issue.duplicatedUniverseCanonicalName",
      messageParams: { canonicalName },
    });
  });

  universeNodes.forEach((universe) => {
    const universeId = universe.data.id;
    const universeMetadata = resolveMetadataForEntity({
      scope: "universe",
      entityRef: buildUniverseEntityRef(universe.data.id),
      rawMetadata: universe.data.metadata,
      fieldPrefix: `universe:${universeId}`,
      registry: metadataPackRegistry,
      bindings: metadataPackBindings,
      issues,
    });

    machine.universes[universeId] = cleanupUndefinedFields({
      id: universe.data.id,
      canonicalName: universe.data.canonicalName || universe.data.id,
      version: universe.data.version || "1.0.0",
      description: universe.data.description || undefined,
      tags: universe.data.tags && universe.data.tags.length > 0 ? [...universe.data.tags] : undefined,
      metadata: Object.keys(universeMetadata).length > 0 ? universeMetadata : undefined,
      universalConstants: compactUC(universe.data.universalConstants),
      realities: {},
    });
  });

  const transitionMetadataById = new Map<string, JsonObject>();
  orderedTransitions.forEach((transition) => {
    const sourceReality = realitiesByNodeId.get(transition.sourceRealityId);
    if (!sourceReality) {
      return;
    }
    const sourceUniverse = universeByNodeId.get(sourceReality.data.universeId);
    if (!sourceUniverse) {
      return;
    }

    const transitionMetadata = resolveMetadataForEntity({
      scope: "transition",
      entityRef: buildTransitionEntityRef(
        sourceUniverse.data.id,
        sourceReality.data.id,
        transition.triggerKind,
        transition.eventName,
        transition.order,
      ),
      rawMetadata: transition.metadata,
      fieldPrefix: `transition:${transition.id}`,
      registry: metadataPackRegistry,
      bindings: metadataPackBindings,
      issues,
    });

    transitionMetadataById.set(transition.id, transitionMetadata);
  });

  realityNodes.forEach((reality) => {
    const parentUniverseNode = universeByNodeId.get(reality.data.universeId);
    if (!parentUniverseNode) {
      return;
    }

    const parentUniverseId = parentUniverseNode.data.id;
    const parentUniverse = machine.universes[parentUniverseId];
    if (!parentUniverse) {
      return;
    }

    const realityId = reality.data.id;
    if (reality.data.isInitial) {
      parentUniverse.initial = realityId;
    }

    const realityMetadata = resolveMetadataForEntity({
      scope: "reality",
      entityRef: buildRealityEntityRef(parentUniverseId, realityId),
      rawMetadata: reality.data.metadata,
      fieldPrefix: `reality:${parentUniverseId}.${realityId}`,
      registry: metadataPackRegistry,
      bindings: metadataPackBindings,
      issues,
    });

    const realityOutput: StateProReality = cleanupUndefinedFields({
      id: realityId,
      type: toRealityType(reality.data.realityType),
      description: reality.data.description || undefined,
      metadata: Object.keys(realityMetadata).length > 0 ? realityMetadata : undefined,
      observers:
        reality.data.observers && reality.data.observers.length > 0
          ? cleanBehaviorRefs(reality.data.observers)
          : undefined,
      entryActions:
        reality.data.entryActions && reality.data.entryActions.length > 0
          ? cleanBehaviorRefs(reality.data.entryActions)
          : undefined,
      exitActions:
        reality.data.exitActions && reality.data.exitActions.length > 0
          ? cleanBehaviorRefs(reality.data.exitActions)
          : undefined,
      entryInvokes:
        reality.data.entryInvokes && reality.data.entryInvokes.length > 0
          ? cleanBehaviorRefs(reality.data.entryInvokes)
          : undefined,
      exitInvokes:
        reality.data.exitInvokes && reality.data.exitInvokes.length > 0
          ? cleanBehaviorRefs(reality.data.exitInvokes)
          : undefined,
    });

    const outgoingTransitions = orderedTransitions.filter(
      (transition) => transition.sourceRealityId === reality.id,
    );

    const alwaysTransitions = outgoingTransitions.filter(
      (transition) => transition.triggerKind === "always",
    );

    if (alwaysTransitions.length > 0) {
      realityOutput.always = alwaysTransitions.map((transition) =>
        buildTransitionOutput(transition, transitionMetadataById.get(transition.id) || {}),
      );
    }

    const eventTransitions = outgoingTransitions.filter((transition) => transition.triggerKind === "on");
    if (eventTransitions.length > 0) {
      const onMap: Record<string, StateProTransition[]> = {};
      eventTransitions.forEach((transition) => {
        const eventName = transition.eventName || "";
        if (!eventName) {
          issues.push({
            code: "SEMANTIC_ERROR",
            severity: "error",
            field: `transition:${transition.id}.eventName`,
            message: "On transition requires eventName",
            messageKey: "issue.transitionEventRequired",
          });
          return;
        }

        if (!onMap[eventName]) {
          onMap[eventName] = [];
        }

        onMap[eventName]?.push(
          buildTransitionOutput(transition, transitionMetadataById.get(transition.id) || {}),
        );
      });

      if (Object.keys(onMap).length > 0) {
        realityOutput.on = onMap;
      }
    }

    parentUniverse.realities[realityId] = realityOutput;
  });

  orderedTransitions.forEach((transition) => {
    const sourceReality = realitiesByNodeId.get(transition.sourceRealityId);
    if (!sourceReality) {
      issues.push({
        code: "SEMANTIC_ERROR",
        severity: "error",
        field: `transition:${transition.id}.sourceRealityId`,
        message: `Transition source reality '${transition.sourceRealityId}' does not exist`,
        messageKey: "issue.transitionSourceMissing",
        messageParams: {
          sourceRealityId: transition.sourceRealityId,
        },
      });
    }
  });

  const machineOutput = cleanupUndefinedFields(
    machine as unknown as Record<string, unknown>,
  ) as unknown as StateProMachine;

  const validation = validateStateProMachine(machineOutput, issues);

  return {
    machine: machineOutput,
    issues: validation.issues,
    canExport: validation.canExport,
  };
};
