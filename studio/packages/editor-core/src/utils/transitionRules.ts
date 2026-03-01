import type { EditorNode, EditorTransition } from "../types";

const isUniverseRef = (ref: string): boolean => ref.startsWith("U:");

const parseUniverseRealityRef = (ref: string): { universeId: string; realityId?: string } | null => {
  if (!isUniverseRef(ref)) {
    return null;
  }

  const parts = ref.split(":");
  if (parts.length === 2) {
    return { universeId: parts[1] || "" };
  }
  if (parts.length === 3) {
    return { universeId: parts[1] || "", realityId: parts[2] || "" };
  }

  return null;
};

export const isInternalTargetRef = (
  sourceRealityId: string,
  targetRef: string,
  nodes: EditorNode[],
): boolean => {
  if (isUniverseRef(targetRef)) {
    return false;
  }

  const sourceReality = nodes.find(
    (node): node is Extract<EditorNode, { type: "reality" }> =>
      node.id === sourceRealityId && node.type === "reality",
  );
  if (!sourceReality) {
    return false;
  }

  const targetReality = nodes.find(
    (node): node is Extract<EditorNode, { type: "reality" }> =>
      node.type === "reality" &&
      node.data.universeId === sourceReality.data.universeId &&
      node.data.id === targetRef,
  );

  return Boolean(targetReality);
};

export const hasInternalTargets = (transition: EditorTransition, nodes: EditorNode[]): boolean => {
  return transition.targets.some((targetRef) =>
    isInternalTargetRef(transition.sourceRealityId, targetRef, nodes),
  );
};

export const isInvalidNotifyTransition = (
  transition: EditorTransition,
  nodes: EditorNode[],
): boolean => {
  if (transition.type !== "notify") {
    return false;
  }

  return hasInternalTargets(transition, nodes);
};

export const resolveTargetUniverseIdFromRef = (targetRef: string): string | null => {
  const parsed = parseUniverseRealityRef(targetRef);
  return parsed?.universeId || null;
};
