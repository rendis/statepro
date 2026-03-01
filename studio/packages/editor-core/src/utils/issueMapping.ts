import type { EditorNode, EditorTransition, SerializeIssue } from "../types";

export interface IssueIndex {
  machine: SerializeIssue[];
  universes: Map<string, SerializeIssue[]>;
  realities: Map<string, SerializeIssue[]>;
  transitions: Map<string, SerializeIssue[]>;
}

const pushIssue = (
  bucket: Map<string, SerializeIssue[]>,
  key: string,
  issue: SerializeIssue,
): void => {
  const current = bucket.get(key) || [];
  current.push(issue);
  bucket.set(key, current);
};

const buildUniverseNodeByDataId = (
  nodes: EditorNode[],
): Map<string, Extract<EditorNode, { type: "universe" }>> => {
  const map = new Map<string, Extract<EditorNode, { type: "universe" }>>();
  nodes.forEach((node) => {
    if (node.type === "universe") {
      map.set(node.data.id, node);
    }
  });
  return map;
};

const buildRealityNodeByUniverseAndDataId = (
  nodes: EditorNode[],
): Map<string, Extract<EditorNode, { type: "reality" }>> => {
  const map = new Map<string, Extract<EditorNode, { type: "reality" }>>();
  nodes.forEach((node) => {
    if (node.type !== "reality") {
      return;
    }
    const key = `${node.data.universeId}::${node.data.id}`;
    map.set(key, node);
  });
  return map;
};

const buildTransitionByGroupAndOrder = (
  transitions: EditorTransition[],
  nodes: EditorNode[],
): Map<string, Map<number, string>> => {
  const groups = new Map<string, Map<number, string>>();
  const sourceRealityByNodeId = new Map<
    string,
    Extract<EditorNode, { type: "reality" }>
  >(
    nodes
      .filter((node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality")
      .map((node) => [node.id, node]),
  );
  const universeByNodeId = new Map<string, Extract<EditorNode, { type: "universe" }>>(
    nodes
      .filter((node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe")
      .map((node) => [node.id, node]),
  );

  transitions.forEach((transition) => {
    const sourceReality = sourceRealityByNodeId.get(transition.sourceRealityId);
    if (!sourceReality) {
      return;
    }
    const universe = universeByNodeId.get(sourceReality.data.universeId);
    if (!universe) {
      return;
    }

    const groupKey =
      transition.triggerKind === "always"
        ? `${universe.data.id}::${sourceReality.data.id}::always`
        : `${universe.data.id}::${sourceReality.data.id}::on::${transition.eventName || ""}`;

    const currentGroup = groups.get(groupKey) || new Map<number, string>();
    currentGroup.set(transition.order, transition.id);
    groups.set(groupKey, currentGroup);
  });

  return groups;
};

const isMachineField = (field: string): boolean => {
  if (!field || field === "machine" || field.startsWith("machine.")) {
    return true;
  }

  return (
    field === "id" ||
    field === "canonicalName" ||
    field === "version" ||
    field.startsWith("metadata") ||
    field.startsWith("metadataPacks") ||
    field.startsWith("initials") ||
    field.startsWith("universalConstants")
  );
};

const readArrayIndex = (field: string, matcher: RegExp): number | null => {
  const match = field.match(matcher);
  if (!match) {
    return null;
  }
  const raw = match[1] ?? match[2];
  if (raw === undefined) {
    return null;
  }
  const parsed = Number.parseInt(raw, 10);
  return Number.isFinite(parsed) ? parsed : null;
};

export const buildIssueIndex = (
  issues: SerializeIssue[],
  nodes: EditorNode[],
  transitions: EditorTransition[],
): IssueIndex => {
  const index: IssueIndex = {
    machine: [],
    universes: new Map(),
    realities: new Map(),
    transitions: new Map(),
  };

  const universeByDataId = buildUniverseNodeByDataId(nodes);
  const realityByUniverseAndDataId = buildRealityNodeByUniverseAndDataId(nodes);
  const transitionByGroupAndOrder = buildTransitionByGroupAndOrder(transitions, nodes);

  issues.forEach((issue) => {
    const field = issue.field || "";
    let assigned = false;

    const transitionById = field.match(/^transition:([^.\[]+)/);
    if (transitionById?.[1]) {
      pushIssue(index.transitions, transitionById[1], issue);
      assigned = true;
    }

    const universeMeta = field.match(/^universe:([^.\[]+)/);
    if (universeMeta?.[1]) {
      const universeNode = universeByDataId.get(universeMeta[1]);
      if (universeNode) {
        pushIssue(index.universes, universeNode.id, issue);
        assigned = true;
      }
    }

    const realityMeta = field.match(/^reality:([^.\[]+)\.([^.\[]+)/);
    if (realityMeta?.[1] && realityMeta[2]) {
      const universeNode = universeByDataId.get(realityMeta[1]);
      const realityNode = universeNode
        ? realityByUniverseAndDataId.get(`${universeNode.id}::${realityMeta[2]}`)
        : null;
      if (realityNode) {
        pushIssue(index.realities, realityNode.id, issue);
        assigned = true;
      }
    }

    const universePath = field.match(/^universes\.([^.[]+)/);
    const universeDataId = universePath?.[1];
    const universeNode = universeDataId ? universeByDataId.get(universeDataId) : null;
    if (universeNode) {
      pushIssue(index.universes, universeNode.id, issue);
      assigned = true;

      const realityPath = field.match(/^universes\.([^.[]+)\.realities\.([^.[]+)/);
      const realityDataId = realityPath?.[2];
      const realityNode = realityDataId
        ? realityByUniverseAndDataId.get(`${universeNode.id}::${realityDataId}`)
        : null;
      if (realityNode) {
        pushIssue(index.realities, realityNode.id, issue);
        assigned = true;

        const alwaysIndex = readArrayIndex(
          field,
          /\.always(?:\[(\d+)\]|\.([0-9]+))(?:\.|$)/,
        );
        if (alwaysIndex !== null) {
          const groupKey = `${universeNode.data.id}::${realityNode.data.id}::always`;
          const transitionId = transitionByGroupAndOrder.get(groupKey)?.get(alwaysIndex);
          if (transitionId) {
            pushIssue(index.transitions, transitionId, issue);
            assigned = true;
          }
        }

        const onMatch = field.match(
          /\.on\.([^.[]+)(?:\[(\d+)\]|\.([0-9]+))(?:\.|$)/,
        );
        if (onMatch?.[1]) {
          const eventName = onMatch[1];
          const rawIndex = onMatch[2] ?? onMatch[3];
          const eventIndex = rawIndex ? Number.parseInt(rawIndex, 10) : null;
          if (eventIndex !== null && Number.isFinite(eventIndex)) {
            const groupKey = `${universeNode.data.id}::${realityNode.data.id}::on::${eventName}`;
            const transitionId = transitionByGroupAndOrder.get(groupKey)?.get(eventIndex);
            if (transitionId) {
              pushIssue(index.transitions, transitionId, issue);
              assigned = true;
            }
          }
        }
      }
    }

    if (!assigned && isMachineField(field)) {
      index.machine.push(issue);
      assigned = true;
    }

    if (!assigned) {
      index.machine.push(issue);
    }
  });

  return index;
};
