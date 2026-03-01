import type { EditorNode, NodeSizeMap, RealityNode, UniverseNode } from "../types";

export const overlaps = (
  a: { x: number; y: number; w: number; h: number },
  b: { x: number; y: number; w: number; h: number },
  margin = 20,
): boolean => {
  return !(
    a.x + a.w + margin <= b.x ||
    a.x >= b.x + b.w + margin ||
    a.y + a.h + margin <= b.y ||
    a.y >= b.y + b.h + margin
  );
};

export const getRealitySize = (nodeId: string, nodeSizes: NodeSizeMap): { w: number; h: number } => {
  return nodeSizes[nodeId] || { w: 192, h: 150 };
};

export const getUniverseChildren = (nodes: EditorNode[], universeId: string): RealityNode[] => {
  return nodes.filter(
    (n): n is RealityNode => n.type === "reality" && n.data.universeId === universeId,
  );
};

export const getUniverseById = (nodes: EditorNode[], id: string): UniverseNode | undefined => {
  return nodes.find((n): n is UniverseNode => n.type === "universe" && n.id === id);
};

export const fitUniverseToChildren = (
  universe: UniverseNode,
  children: RealityNode[],
  nodeSizes: NodeSizeMap,
): { w: number; h: number } => {
  if (children.length === 0) {
    return { w: 240, h: 220 };
  }

  let maxRight = universe.x + 240;
  let maxBottom = universe.y + 220;

  children.forEach((child) => {
    const size = getRealitySize(child.id, nodeSizes);
    maxRight = Math.max(maxRight, child.x + size.w + 20);
    maxBottom = Math.max(maxBottom, child.y + size.h + 20);
  });

  return {
    w: maxRight - universe.x,
    h: maxBottom - universe.y,
  };
};
