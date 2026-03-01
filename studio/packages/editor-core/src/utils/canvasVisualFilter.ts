import type { EditorNode, EditorTransition } from "../types";
import {
  buildNodeReferenceIndex,
  resolveTargetReferenceToNodeId,
} from "./references";

export interface CanvasVisualFocusResult {
  focusedNodeIds: Set<string>;
  focusedTransitionIds: Set<string>;
}

const isUniverseNode = (
  node: EditorNode | undefined,
): node is Extract<EditorNode, { type: "universe" }> => Boolean(node && node.type === "universe");

const isRealityNode = (
  node: EditorNode | undefined,
): node is Extract<EditorNode, { type: "reality" }> => Boolean(node && node.type === "reality");

export const resolveVisualFocus = (
  nodes: EditorNode[],
  transitions: EditorTransition[],
  selectedNodeIds: string[],
): CanvasVisualFocusResult => {
  const focusedNodeIds = new Set<string>();
  const focusedTransitionIds = new Set<string>();

  if (selectedNodeIds.length === 0) {
    return {
      focusedNodeIds,
      focusedTransitionIds,
    };
  }

  const nodeById = new Map(nodes.map((node) => [node.id, node]));
  const realitiesByUniverseId = new Map<string, string[]>();
  const selectedUniverseIds = new Set<string>();
  const selectedRealityIds = new Set<string>();
  const index = buildNodeReferenceIndex(nodes);

  nodes.forEach((node) => {
    if (node.type !== "reality") {
      return;
    }

    const current = realitiesByUniverseId.get(node.data.universeId) || [];
    current.push(node.id);
    realitiesByUniverseId.set(node.data.universeId, current);
  });

  const normalizedSelection = Array.from(new Set(selectedNodeIds)).filter((id) => {
    const node = nodeById.get(id);
    return node?.type === "universe" || node?.type === "reality";
  });

  normalizedSelection.forEach((nodeId) => {
    const node = nodeById.get(nodeId);
    if (isUniverseNode(node)) {
      selectedUniverseIds.add(node.id);
      focusedNodeIds.add(node.id);

      const children = realitiesByUniverseId.get(node.id) || [];
      children.forEach((childId) => focusedNodeIds.add(childId));
      return;
    }

    if (isRealityNode(node)) {
      selectedRealityIds.add(node.id);
      focusedNodeIds.add(node.id);
      focusedNodeIds.add(node.data.universeId);
    }
  });

  transitions.forEach((transition) => {
    const sourceReality = nodeById.get(transition.sourceRealityId);
    const sourceUniverseId = isRealityNode(sourceReality)
      ? sourceReality.data.universeId
      : null;

    let isDirectlyConnected = false;
    if (isRealityNode(sourceReality)) {
      if (selectedRealityIds.has(sourceReality.id)) {
        isDirectlyConnected = true;
      }
      if (sourceUniverseId && selectedUniverseIds.has(sourceUniverseId)) {
        isDirectlyConnected = true;
      }
    }

    if (!isDirectlyConnected) {
      for (const targetRef of transition.targets) {
        const targetNodeId = resolveTargetReferenceToNodeId(
          transition.sourceRealityId,
          targetRef,
          nodes,
          index,
        );
        const targetNode = targetNodeId ? nodeById.get(targetNodeId) : undefined;

        if (isUniverseNode(targetNode) && selectedUniverseIds.has(targetNode.id)) {
          isDirectlyConnected = true;
          break;
        }

        if (isRealityNode(targetNode)) {
          if (selectedRealityIds.has(targetNode.id)) {
            isDirectlyConnected = true;
            break;
          }

          if (selectedUniverseIds.has(targetNode.data.universeId)) {
            isDirectlyConnected = true;
            break;
          }
        }
      }
    }

    if (!isDirectlyConnected) {
      return;
    }

    focusedTransitionIds.add(transition.id);

    if (isRealityNode(sourceReality)) {
      focusedNodeIds.add(sourceReality.id);
      focusedNodeIds.add(sourceReality.data.universeId);
    }

    transition.targets.forEach((targetRef) => {
      const targetNodeId = resolveTargetReferenceToNodeId(
        transition.sourceRealityId,
        targetRef,
        nodes,
        index,
      );
      const targetNode = targetNodeId ? nodeById.get(targetNodeId) : undefined;

      if (isUniverseNode(targetNode)) {
        focusedNodeIds.add(targetNode.id);
        return;
      }

      if (isRealityNode(targetNode)) {
        focusedNodeIds.add(targetNode.id);
        focusedNodeIds.add(targetNode.data.universeId);
      }
    });
  });

  return {
    focusedNodeIds,
    focusedTransitionIds,
  };
};
