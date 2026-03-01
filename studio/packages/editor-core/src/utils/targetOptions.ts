import type { EditorNode } from "../types";
import { parseTargetReference } from "./references";

export type TransitionTargetOptionKind =
  | "internalReality"
  | "externalUniverse"
  | "externalReality";

export interface TransitionTargetOption {
  value: string;
  label: string;
  kind: TransitionTargetOptionKind;
  universeId?: string;
  realityId?: string;
}

export interface TransitionTargetDisplayItem {
  value: string;
  label: string;
  kind: TransitionTargetOptionKind | "invalid";
  valid: boolean;
}

const getUniverseNodes = (
  nodes: EditorNode[],
): Extract<EditorNode, { type: "universe" }>[] => {
  return nodes.filter((node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe");
};

const getRealityNodes = (
  nodes: EditorNode[],
): Extract<EditorNode, { type: "reality" }>[] => {
  return nodes.filter((node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality");
};

export const buildTransitionTargetOptions = (
  sourceRealityNodeId: string,
  nodes: EditorNode[],
): TransitionTargetOption[] => {
  const sourceReality = nodes.find(
    (node): node is Extract<EditorNode, { type: "reality" }> =>
      node.type === "reality" && node.id === sourceRealityNodeId,
  );

  if (!sourceReality) {
    return [];
  }

  const universes = getUniverseNodes(nodes);
  const realities = getRealityNodes(nodes);

  const internalRealityOptions = realities
    .filter((reality) => reality.data.universeId === sourceReality.data.universeId)
    .map((reality) => ({
      value: reality.data.id,
      label: reality.data.id,
      kind: "internalReality" as const,
      universeId:
        universes.find((universe) => universe.id === reality.data.universeId)?.data.id || undefined,
      realityId: reality.data.id,
    }));

  const externalUniverseOptions = universes.map((universe) => ({
    value: `U:${universe.data.id}`,
    label: `U:${universe.data.id}`,
    kind: "externalUniverse" as const,
    universeId: universe.data.id,
  }));

  const externalRealityOptions = realities
    .map((reality) => {
      const universe = universes.find((candidate) => candidate.id === reality.data.universeId);
      if (!universe) {
        return null;
      }

      return {
        value: `U:${universe.data.id}:${reality.data.id}`,
        label: `U:${universe.data.id}:${reality.data.id}`,
        kind: "externalReality" as const,
        universeId: universe.data.id,
        realityId: reality.data.id,
      };
    })
    .filter((option): option is NonNullable<typeof option> => option !== null);

  return [...internalRealityOptions, ...externalUniverseOptions, ...externalRealityOptions];
};

export const buildTransitionTargetDisplayItems = (
  targets: string[],
  sourceRealityNodeId: string,
  nodes: EditorNode[],
): TransitionTargetDisplayItem[] => {
  const options = buildTransitionTargetOptions(sourceRealityNodeId, nodes);
  const optionByValue = new Map(options.map((option) => [option.value, option]));

  return targets.map((target) => {
    const option = optionByValue.get(target);
    if (option) {
      return {
        value: target,
        label: option.label,
        kind: option.kind,
        valid: true,
      };
    }

    const parsed = parseTargetReference(target);
    if (!parsed) {
      return {
        value: target,
        label: target,
        kind: "invalid",
        valid: false,
      };
    }

    return {
      value: target,
      label: target,
      kind: "invalid",
      valid: false,
    };
  });
};
