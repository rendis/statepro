import {
  createEmptyUniversalConstants,
  createInitialMetadataPackBindingMap,
  defaultRegistry,
} from "./defaults";
import type {
  EditorNode,
  EditorState,
  EditorTransition,
  JsonObject,
  StateProMachine,
  UniversalConstants,
  BehaviorRef,
} from "../types";

const toJsonText = (value: unknown): string => {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return "{}";
  }

  return JSON.stringify(value, null, 2);
};

const toBehaviorArray = (behaviors?: BehaviorRef[]): BehaviorRef[] => {
  return Array.isArray(behaviors)
    ? behaviors.filter((behavior) => behavior && behavior.src).map((behavior) => ({ ...behavior }))
    : [];
};

const mergeDeprecatedConditionToConditions = (
  transition: {
    condition?: BehaviorRef;
    conditions?: BehaviorRef[];
  },
): BehaviorRef[] => {
  const merged = [
    ...(transition.condition ? [{ ...transition.condition }] : []),
    ...toBehaviorArray(transition.conditions),
  ];

  const seen = new Set<string>();
  const unique: BehaviorRef[] = [];
  merged.forEach((condition) => {
    const src = condition?.src || "";
    if (!src || seen.has(src)) {
      return;
    }
    seen.add(src);
    unique.push(condition);
  });

  return unique;
};

const normalizeDeprecatedConditionModel = (machine: StateProMachine): StateProMachine => {
  const normalizedUniverses = Object.fromEntries(
    Object.entries(machine.universes || {}).map(([universeId, universe]) => {
      const normalizedRealities = Object.fromEntries(
        Object.entries(universe.realities || {}).map(([realityId, reality]) => {
          const normalizedAlways = (reality.always || []).map((transition) => {
            const normalizedConditions = mergeDeprecatedConditionToConditions(transition);
            const { condition: _deprecatedCondition, ...restTransition } = transition;
            return {
              ...restTransition,
              conditions: normalizedConditions.length > 0 ? normalizedConditions : undefined,
            };
          });

          const normalizedOn =
            reality.on == null
              ? reality.on
              : Object.fromEntries(
                  Object.entries(reality.on).map(([eventName, transitions]) => [
                    eventName,
                    (transitions || []).map((transition) => {
                      const normalizedConditions = mergeDeprecatedConditionToConditions(transition);
                      const { condition: _deprecatedCondition, ...restTransition } = transition;
                      return {
                        ...restTransition,
                        conditions:
                          normalizedConditions.length > 0 ? normalizedConditions : undefined,
                      };
                    }),
                  ]),
                );

          return [
            realityId,
            {
              ...reality,
              always: normalizedAlways,
              on: normalizedOn,
            },
          ];
        }),
      );

      return [
        universeId,
        {
          ...universe,
          realities: normalizedRealities,
        },
      ];
    }),
  ) as StateProMachine["universes"];

  return {
    ...machine,
    universes: normalizedUniverses,
  };
};

const toUniversalConstants = (
  constants?: Partial<UniversalConstants>,
): UniversalConstants => {
  const empty = createEmptyUniversalConstants();
  if (!constants) {
    return empty;
  }

  return {
    entryActions: toBehaviorArray(constants.entryActions),
    exitActions: toBehaviorArray(constants.exitActions),
    entryInvokes: toBehaviorArray(constants.entryInvokes),
    exitInvokes: toBehaviorArray(constants.exitInvokes),
    actionsOnTransition: toBehaviorArray(constants.actionsOnTransition),
    invokesOnTransition: toBehaviorArray(constants.invokesOnTransition),
  };
};

const toRealityType = (type: "transition" | "final" | "unsuccessfulFinal") => {
  if (type === "final") {
    return "success" as const;
  }

  if (type === "unsuccessfulFinal") {
    return "error" as const;
  }

  return "normal" as const;
};

export const deserializeStatePro = (machine: StateProMachine): EditorState => {
  const normalizedMachine = normalizeDeprecatedConditionModel(structuredClone(machine));
  const machineMetadataObject =
    normalizedMachine.metadata &&
    typeof normalizedMachine.metadata === "object" &&
    !Array.isArray(normalizedMachine.metadata)
      ? (normalizedMachine.metadata as JsonObject)
      : {};

  const nodes: EditorNode[] = [];
  const transitions: EditorTransition[] = [];
  const universeNodeIdByUniverseId = new Map<string, string>();
  const realityNodeIdByUniverseAndRealityId = new Map<string, string>();

  const universeEntries = Object.entries(normalizedMachine.universes || {});

  universeEntries.forEach(([universeId, universe], universeIndex) => {
    const universeNodeId = `univ-${universeIndex + 1}`;
    const universeX = 1050 + (universeIndex % 2) * 650;
    const universeY = 1050 + Math.floor(universeIndex / 2) * 520;

    const realityEntries = Object.entries(universe.realities || {});
    const rows = Math.max(1, Math.ceil(realityEntries.length / 2));
    const universeHeight = Math.max(320, 100 + rows * 190);
    const universeWidth = 560;

    const universeMetadata =
      universe.metadata && typeof universe.metadata === "object" && !Array.isArray(universe.metadata)
        ? (universe.metadata as JsonObject)
        : undefined;

    nodes.push({
      id: universeNodeId,
      type: "universe",
      x: universeX,
      y: universeY,
      w: universeWidth,
      h: universeHeight,
      data: {
        id: universe.id,
        name: universe.id,
        canonicalName: universe.canonicalName || universe.id,
        version: universe.version || "1.0.0",
        description: universe.description || "",
        tags: universe.tags ? [...universe.tags] : [],
        metadata: toJsonText(universeMetadata),
        universalConstants: toUniversalConstants(universe.universalConstants),
      },
    });

    universeNodeIdByUniverseId.set(universeId, universeNodeId);

    realityEntries.forEach(([realityId, reality], realityIndex) => {
      const col = realityIndex % 2;
      const row = Math.floor(realityIndex / 2);
      const realityNodeId = `real-${universeIndex + 1}-${realityIndex + 1}`;

      const realityMetadata =
        reality.metadata && typeof reality.metadata === "object" && !Array.isArray(reality.metadata)
          ? (reality.metadata as JsonObject)
          : undefined;

      nodes.push({
        id: realityNodeId,
        type: "reality",
        x: universeX + 30 + col * 240,
        y: universeY + 80 + row * 190,
        data: {
          id: reality.id,
          name: reality.id,
          universeId: universeNodeId,
          isInitial: universe.initial === reality.id,
          realityType: toRealityType(reality.type),
          description: reality.description || "",
          metadata: toJsonText(realityMetadata),
          observers: toBehaviorArray(reality.observers),
          entryActions: toBehaviorArray(reality.entryActions),
          exitActions: toBehaviorArray(reality.exitActions),
          entryInvokes: toBehaviorArray(reality.entryInvokes),
          exitInvokes: toBehaviorArray(reality.exitInvokes),
        },
      });

      realityNodeIdByUniverseAndRealityId.set(`${universeId}::${realityId}`, realityNodeId);
    });
  });

  let transitionCounter = 1;

  universeEntries.forEach(([universeId, universe]) => {
    Object.entries(universe.realities || {}).forEach(([realityId, reality]) => {
      const sourceRealityNodeId = realityNodeIdByUniverseAndRealityId.get(`${universeId}::${realityId}`);
      if (!sourceRealityNodeId) {
        return;
      }

      (reality.always || []).forEach((transition, transitionIndex) => {
        const transitionMetadata =
          transition.metadata &&
          typeof transition.metadata === "object" &&
          !Array.isArray(transition.metadata)
            ? (transition.metadata as JsonObject)
            : undefined;

        transitions.push({
          id: `tr-${transitionCounter++}`,
          sourceRealityId: sourceRealityNodeId,
          triggerKind: "always",
          eventName: undefined,
          type: transition.type || "default",
          condition: undefined,
          conditions: mergeDeprecatedConditionToConditions(transition),
          actions: toBehaviorArray(transition.actions),
          invokes: toBehaviorArray(transition.invokes),
          description: transition.description || "",
          metadata: toJsonText(transitionMetadata),
          targets: [...(transition.targets || [])],
          order: transitionIndex,
        });
      });

      Object.entries(reality.on || {}).forEach(([eventName, eventTransitions]) => {
        (eventTransitions || []).forEach((transition, transitionIndex) => {
          const transitionMetadata =
            transition.metadata &&
            typeof transition.metadata === "object" &&
            !Array.isArray(transition.metadata)
              ? (transition.metadata as JsonObject)
              : undefined;

          transitions.push({
            id: `tr-${transitionCounter++}`,
            sourceRealityId: sourceRealityNodeId,
            triggerKind: "on",
            eventName,
            type: transition.type || "default",
            condition: undefined,
            conditions: mergeDeprecatedConditionToConditions(transition),
            actions: toBehaviorArray(transition.actions),
            invokes: toBehaviorArray(transition.invokes),
            description: transition.description || "",
            metadata: toJsonText(transitionMetadata),
            targets: [...(transition.targets || [])],
            order: transitionIndex,
          });
        });
      });
    });
  });

  return {
    nodes,
    transitions,
    nodeSizes: {},
    selectedElement: null,
    machineConfig: {
      id: normalizedMachine.id,
      canonicalName: normalizedMachine.canonicalName,
      version: normalizedMachine.version,
      description: normalizedMachine.description || "",
      initials: normalizedMachine.initials ? [...normalizedMachine.initials] : [],
      universalConstants: toUniversalConstants(normalizedMachine.universalConstants),
      metadata: toJsonText(machineMetadataObject),
    },
    registry: structuredClone(defaultRegistry),
    metadataPackRegistry: [],
    metadataPackBindings: createInitialMetadataPackBindingMap(),
    lastImportedMachine: structuredClone(normalizedMachine),
    isDirtyFromImport: false,
  };
};
