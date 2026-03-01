import type { EditorNode } from "../types";

export type TargetReferenceKind = "reality" | "universe" | "universeReality";

export interface ParsedTargetReference {
  kind: TargetReferenceKind;
  universeId?: string;
  realityId?: string;
}

const identifierPattern = /^[A-Za-z](?:[A-Za-z0-9_-]*[A-Za-z0-9])?$/;

export const isValidIdentifier = (value: string): boolean => {
  return identifierPattern.test(value);
};

export const parseTargetReference = (ref: string): ParsedTargetReference | null => {
  if (!ref) {
    return null;
  }

  if (ref.startsWith("U:")) {
    const parts = ref.split(":");
    if (parts.length === 2) {
      const universeId = parts[1] || "";
      if (!isValidIdentifier(universeId)) {
        return null;
      }
      return {
        kind: "universe",
        universeId,
      };
    }

    if (parts.length === 3) {
      const universeId = parts[1] || "";
      const realityId = parts[2] || "";
      if (!isValidIdentifier(universeId) || !isValidIdentifier(realityId)) {
        return null;
      }
      return {
        kind: "universeReality",
        universeId,
        realityId,
      };
    }

    return null;
  }

  if (!isValidIdentifier(ref)) {
    return null;
  }

  return {
    kind: "reality",
    realityId: ref,
  };
};

export interface NodeReferenceIndex {
  universeById: Map<string, Extract<EditorNode, { type: "universe" }>>;
  universeNodeByDataId: Map<string, Extract<EditorNode, { type: "universe" }>>;
  realityById: Map<string, Extract<EditorNode, { type: "reality" }>>;
  realityByUniverseAndDataId: Map<string, Extract<EditorNode, { type: "reality" }>>;
}

export const buildNodeReferenceIndex = (nodes: EditorNode[]): NodeReferenceIndex => {
  const universeById = new Map<string, Extract<EditorNode, { type: "universe" }>>();
  const universeNodeByDataId = new Map<string, Extract<EditorNode, { type: "universe" }>>();
  const realityById = new Map<string, Extract<EditorNode, { type: "reality" }>>();
  const realityByUniverseAndDataId = new Map<string, Extract<EditorNode, { type: "reality" }>>();

  nodes.forEach((node) => {
    if (node.type === "universe") {
      universeById.set(node.id, node);
      universeNodeByDataId.set(node.data.id, node);
      return;
    }

    if (node.type !== "reality") {
      return;
    }

    realityById.set(node.id, node);
    const key = `${node.data.universeId}::${node.data.id}`;
    realityByUniverseAndDataId.set(key, node);
  });

  return {
    universeById,
    universeNodeByDataId,
    realityById,
    realityByUniverseAndDataId,
  };
};

export const buildTargetReferenceFromNodes = (
  sourceRealityNodeId: string,
  targetNodeId: string,
  nodes: EditorNode[],
): string | null => {
  const sourceNode = nodes.find(
    (node): node is Extract<EditorNode, { type: "reality" }> =>
      node.id === sourceRealityNodeId && node.type === "reality",
  );
  const targetNode = nodes.find((node) => node.id === targetNodeId);

  if (!sourceNode || !targetNode) {
    return null;
  }

  if (targetNode.type === "universe") {
    return `U:${targetNode.data.id}`;
  }

  if (targetNode.type !== "reality") {
    return null;
  }

  if (targetNode.data.universeId === sourceNode.data.universeId) {
    return targetNode.data.id;
  }

  const universeNode = nodes.find(
    (node): node is Extract<EditorNode, { type: "universe" }> =>
      node.type === "universe" && node.id === targetNode.data.universeId,
  );

  if (!universeNode) {
    return null;
  }

  return `U:${universeNode.data.id}:${targetNode.data.id}`;
};

export const resolveTargetReferenceToNodeId = (
  sourceRealityNodeId: string,
  targetRef: string,
  nodes: EditorNode[],
  index: NodeReferenceIndex,
): string | null => {
  const parsed = parseTargetReference(targetRef);
  if (!parsed) {
    return null;
  }

  if (parsed.kind === "universe") {
    return index.universeNodeByDataId.get(parsed.universeId || "")?.id || null;
  }

  if (parsed.kind === "universeReality") {
    const universeNode = index.universeNodeByDataId.get(parsed.universeId || "");
    if (!universeNode) {
      return null;
    }
    const key = `${universeNode.id}::${parsed.realityId || ""}`;
    return index.realityByUniverseAndDataId.get(key)?.id || null;
  }

  const sourceReality =
    index.realityById.get(sourceRealityNodeId) ||
    nodes.find(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.type === "reality" && node.id === sourceRealityNodeId,
    );
  if (!sourceReality) {
    return null;
  }

  const key = `${sourceReality.data.universeId}::${parsed.realityId || ""}`;
  return index.realityByUniverseAndDataId.get(key)?.id || null;
};
