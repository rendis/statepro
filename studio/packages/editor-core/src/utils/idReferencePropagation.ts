import type {
  EditorNode,
  EditorTransition,
  MachineConfig,
  MetadataPackBinding,
  MetadataPackBindingMap,
} from "../types";
import {
  buildRealityEntityRef,
  buildUniverseEntityRef,
} from "./metadataPacks";
import { parseTargetReference } from "./references";

export interface RenameReferencesResult {
  nodes: EditorNode[];
  transitions: EditorTransition[];
  machineConfig: MachineConfig;
  metadataPackBindings: MetadataPackBindingMap;
}

export interface RenameUniverseIdParams {
  nodes: EditorNode[];
  transitions: EditorTransition[];
  machineConfig: MachineConfig;
  metadataPackBindings: MetadataPackBindingMap;
  universeNodeId: string;
  nextUniverseId: string;
}

export interface RenameRealityIdParams {
  nodes: EditorNode[];
  transitions: EditorTransition[];
  machineConfig: MachineConfig;
  metadataPackBindings: MetadataPackBindingMap;
  realityNodeId: string;
  nextRealityId: string;
}

const remapUniverseTargetRef = (
  targetRef: string,
  oldUniverseId: string,
  nextUniverseId: string,
): string => {
  const parsed = parseTargetReference(targetRef);
  if (!parsed) {
    return targetRef;
  }

  if (parsed.kind === "universe" && parsed.universeId === oldUniverseId) {
    return `U:${nextUniverseId}`;
  }

  if (parsed.kind === "universeReality" && parsed.universeId === oldUniverseId) {
    return `U:${nextUniverseId}:${parsed.realityId || ""}`;
  }

  return targetRef;
};

const remapRealityTargetRef = (
  targetRef: string,
  oldRealityId: string,
  nextRealityId: string,
  realityUniverseId: string,
  transitionSourceUniverseId: string | null,
): string => {
  const parsed = parseTargetReference(targetRef);
  if (!parsed) {
    return targetRef;
  }

  if (
    parsed.kind === "reality" &&
    parsed.realityId === oldRealityId &&
    transitionSourceUniverseId === realityUniverseId
  ) {
    return nextRealityId;
  }

  if (
    parsed.kind === "universeReality" &&
    parsed.universeId === realityUniverseId &&
    parsed.realityId === oldRealityId
  ) {
    return `U:${realityUniverseId}:${nextRealityId}`;
  }

  return targetRef;
};

const remapInitialUniverseRef = (
  initialRef: string,
  oldUniverseId: string,
  nextUniverseId: string,
): string => {
  const parsed = parseTargetReference(initialRef);
  if (!parsed) {
    return initialRef;
  }

  if (parsed.kind === "universe" && parsed.universeId === oldUniverseId) {
    return `U:${nextUniverseId}`;
  }

  if (parsed.kind === "universeReality" && parsed.universeId === oldUniverseId) {
    return `U:${nextUniverseId}:${parsed.realityId || ""}`;
  }

  return initialRef;
};

const remapInitialRealityRef = (
  initialRef: string,
  universeId: string,
  oldRealityId: string,
  nextRealityId: string,
): string => {
  const parsed = parseTargetReference(initialRef);
  if (!parsed) {
    return initialRef;
  }

  if (
    parsed.kind === "universeReality" &&
    parsed.universeId === universeId &&
    parsed.realityId === oldRealityId
  ) {
    return `U:${universeId}:${nextRealityId}`;
  }

  return initialRef;
};

const mapScopeBindings = (
  bindings: MetadataPackBinding[],
  mapEntityRef: (entityRef: string) => string,
): MetadataPackBinding[] => {
  return bindings.map((binding) => {
    const nextEntityRef = mapEntityRef(binding.entityRef);
    if (nextEntityRef === binding.entityRef) {
      return binding;
    }

    return {
      ...binding,
      entityRef: nextEntityRef,
    };
  });
};

export const renameUniverseId = ({
  nodes,
  transitions,
  machineConfig,
  metadataPackBindings,
  universeNodeId,
  nextUniverseId,
}: RenameUniverseIdParams): RenameReferencesResult => {
  const universeNode = nodes.find(
    (node): node is Extract<EditorNode, { type: "universe" }> =>
      node.type === "universe" && node.id === universeNodeId,
  );

  if (!universeNode || !nextUniverseId) {
    return {
      nodes,
      transitions,
      machineConfig,
      metadataPackBindings,
    };
  }

  const oldUniverseId = universeNode.data.id;
  if (!oldUniverseId || oldUniverseId === nextUniverseId) {
    const nextNodes = nodes.map((node) => {
      if (node.type !== "universe" || node.id !== universeNodeId) {
        return node;
      }

      return {
        ...node,
        data: {
          ...node.data,
          id: nextUniverseId,
          name: nextUniverseId,
        },
      };
    });

    return {
      nodes: nextNodes,
      transitions,
      machineConfig,
      metadataPackBindings,
    };
  }

  const nextNodes = nodes.map((node) => {
    if (node.type !== "universe" || node.id !== universeNodeId) {
      return node;
    }

    return {
      ...node,
      data: {
        ...node.data,
        id: nextUniverseId,
        name: nextUniverseId,
      },
    };
  });

  const nextTransitions = transitions.map((transition) => ({
    ...transition,
    targets: transition.targets.map((target) =>
      remapUniverseTargetRef(target, oldUniverseId, nextUniverseId),
    ),
  }));

  const nextMachineConfig: MachineConfig = {
    ...machineConfig,
    initials: machineConfig.initials.map((initial) =>
      remapInitialUniverseRef(initial, oldUniverseId, nextUniverseId),
    ),
  };

  const oldUniverseEntityRef = buildUniverseEntityRef(oldUniverseId);
  const nextUniverseEntityRef = buildUniverseEntityRef(nextUniverseId);
  const oldRealityPrefix = `${oldUniverseEntityRef}:R:`;
  const nextRealityPrefix = `${nextUniverseEntityRef}:R:`;

  const nextBindings: MetadataPackBindingMap = {
    machine: metadataPackBindings.machine,
    universe: mapScopeBindings(metadataPackBindings.universe, (entityRef) =>
      entityRef === oldUniverseEntityRef ? nextUniverseEntityRef : entityRef,
    ),
    reality: mapScopeBindings(metadataPackBindings.reality, (entityRef) =>
      entityRef.startsWith(oldRealityPrefix)
        ? `${nextRealityPrefix}${entityRef.slice(oldRealityPrefix.length)}`
        : entityRef,
    ),
    transition: mapScopeBindings(metadataPackBindings.transition, (entityRef) =>
      entityRef.startsWith(oldRealityPrefix)
        ? `${nextRealityPrefix}${entityRef.slice(oldRealityPrefix.length)}`
        : entityRef,
    ),
  };

  return {
    nodes: nextNodes,
    transitions: nextTransitions,
    machineConfig: nextMachineConfig,
    metadataPackBindings: nextBindings,
  };
};

export const renameRealityId = ({
  nodes,
  transitions,
  machineConfig,
  metadataPackBindings,
  realityNodeId,
  nextRealityId,
}: RenameRealityIdParams): RenameReferencesResult => {
  const realityNode = nodes.find(
    (node): node is Extract<EditorNode, { type: "reality" }> =>
      node.type === "reality" && node.id === realityNodeId,
  );

  if (!realityNode || !nextRealityId) {
    return {
      nodes,
      transitions,
      machineConfig,
      metadataPackBindings,
    };
  }

  const oldRealityId = realityNode.data.id;
  const parentUniverseNode = nodes.find(
    (node): node is Extract<EditorNode, { type: "universe" }> =>
      node.type === "universe" && node.id === realityNode.data.universeId,
  );

  if (!parentUniverseNode) {
    return {
      nodes,
      transitions,
      machineConfig,
      metadataPackBindings,
    };
  }

  const universeId = parentUniverseNode.data.id;

  const nextNodes = nodes.map((node) => {
    if (node.type !== "reality" || node.id !== realityNodeId) {
      return node;
    }

    return {
      ...node,
      data: {
        ...node.data,
        id: nextRealityId,
        name: nextRealityId,
      },
    };
  });

  const universeByNodeId = new Map(
    nodes
      .filter(
        (node): node is Extract<EditorNode, { type: "universe" }> =>
          node.type === "universe",
      )
      .map((node) => [node.id, node]),
  );
  const sourceUniverseDataIdByRealityNodeId = new Map<string, string>();
  nodes.forEach((node) => {
    if (node.type !== "reality") {
      return;
    }
    const parentUniverse = universeByNodeId.get(node.data.universeId);
    if (!parentUniverse) {
      return;
    }

    sourceUniverseDataIdByRealityNodeId.set(node.id, parentUniverse.data.id);
  });

  const nextTransitions = transitions.map((transition) => {
    const sourceUniverseId =
      sourceUniverseDataIdByRealityNodeId.get(transition.sourceRealityId) || null;

    return {
      ...transition,
      targets: transition.targets.map((target) =>
        remapRealityTargetRef(
          target,
          oldRealityId,
          nextRealityId,
          universeId,
          sourceUniverseId,
        ),
      ),
    };
  });

  const nextMachineConfig: MachineConfig = {
    ...machineConfig,
    initials: machineConfig.initials.map((initial) =>
      remapInitialRealityRef(initial, universeId, oldRealityId, nextRealityId),
    ),
  };

  const oldRealityEntityRef = buildRealityEntityRef(universeId, oldRealityId);
  const nextRealityEntityRef = buildRealityEntityRef(universeId, nextRealityId);
  const oldTransitionPrefix = `${oldRealityEntityRef}:T:`;
  const nextTransitionPrefix = `${nextRealityEntityRef}:T:`;

  const nextBindings: MetadataPackBindingMap = {
    machine: metadataPackBindings.machine,
    universe: metadataPackBindings.universe,
    reality: mapScopeBindings(metadataPackBindings.reality, (entityRef) =>
      entityRef === oldRealityEntityRef ? nextRealityEntityRef : entityRef,
    ),
    transition: mapScopeBindings(metadataPackBindings.transition, (entityRef) =>
      entityRef.startsWith(oldTransitionPrefix)
        ? `${nextTransitionPrefix}${entityRef.slice(oldTransitionPrefix.length)}`
        : entityRef,
    ),
  };

  return {
    nodes: nextNodes,
    transitions: nextTransitions,
    machineConfig: nextMachineConfig,
    metadataPackBindings: nextBindings,
  };
};
