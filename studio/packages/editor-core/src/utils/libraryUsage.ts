import type {
  BehaviorRef,
  EditorNode,
  EditorTransition,
  MachineConfig,
  UniversalConstants,
} from "../types";

const UNIVERSAL_CONSTANT_FIELDS: Array<keyof UniversalConstants> = [
  "entryActions",
  "exitActions",
  "entryInvokes",
  "exitInvokes",
  "actionsOnTransition",
  "invokesOnTransition",
];

const REALITY_BEHAVIOR_FIELDS = [
  "observers",
  "entryActions",
  "exitActions",
  "entryInvokes",
  "exitInvokes",
] as const;

const TRANSITION_ARRAY_FIELDS = ["conditions", "actions", "invokes"] as const;

type BehaviorUsageScope = "machine" | "universe" | "reality" | "transition";

export interface BehaviorUsageLocation {
  scope: BehaviorUsageScope;
  containerId: string;
  field: string;
  label: string;
  index?: number;
}

export interface UsageSummary {
  total: number;
  locations: BehaviorUsageLocation[];
}

export interface BehaviorStateSlices {
  machineConfig: MachineConfig;
  nodes: EditorNode[];
  transitions: EditorTransition[];
}

interface RemoveReferencesResult {
  nextNodes: EditorNode[];
  nextTransitions: EditorTransition[];
  nextMachineConfig: MachineConfig;
  removedCount: number;
}

interface AppendArrayUsageArgs {
  src: string;
  values: BehaviorRef[] | undefined;
  scope: BehaviorUsageScope;
  containerId: string;
  containerLabel: string;
  field: string;
  usages: BehaviorUsageLocation[];
}

const appendArrayUsage = ({
  src,
  values,
  scope,
  containerId,
  containerLabel,
  field,
  usages,
}: AppendArrayUsageArgs): void => {
  if (!Array.isArray(values) || values.length === 0) {
    return;
  }

  values.forEach((behavior, index) => {
    if (behavior?.src !== src) {
      return;
    }

    usages.push({
      scope,
      containerId,
      field,
      index,
      label: `${containerLabel} · ${field}[${index}]`,
    });
  });
};

const removeFromRequiredBehaviorArray = (
  values: BehaviorRef[] | undefined,
  src: string,
): { next: BehaviorRef[]; removed: number } => {
  const current = Array.isArray(values) ? values : [];
  if (current.length === 0) {
    return { next: current, removed: 0 };
  }

  const next = current.filter((behavior) => behavior?.src !== src);
  const removed = current.length - next.length;
  return {
    next: removed > 0 ? next : current,
    removed,
  };
};

const removeFromOptionalBehaviorArray = (
  values: BehaviorRef[] | undefined,
  src: string,
): { next: BehaviorRef[] | undefined; removed: number } => {
  if (!Array.isArray(values)) {
    return { next: values, removed: 0 };
  }

  if (values.length === 0) {
    return { next: values, removed: 0 };
  }

  const next = values.filter((behavior) => behavior?.src !== src);
  const removed = values.length - next.length;
  return {
    next: removed > 0 ? next : values,
    removed,
  };
};

export const collectBehaviorUsages = (
  src: string,
  state: BehaviorStateSlices,
): UsageSummary => {
  const usages: BehaviorUsageLocation[] = [];

  UNIVERSAL_CONSTANT_FIELDS.forEach((field) => {
    appendArrayUsage({
      src,
      values: state.machineConfig.universalConstants?.[field],
      scope: "machine",
      containerId: state.machineConfig.id || "machine",
      containerLabel: `Machine:${state.machineConfig.id || "machine"}`,
      field: `universalConstants.${field}`,
      usages,
    });
  });

  state.nodes.forEach((node) => {
    if (node.type === "universe") {
      const universeId = node.data.id || node.id;
      UNIVERSAL_CONSTANT_FIELDS.forEach((field) => {
        appendArrayUsage({
          src,
          values: node.data.universalConstants?.[field],
          scope: "universe",
          containerId: universeId,
          containerLabel: `Universe:${universeId}`,
          field: `universalConstants.${field}`,
          usages,
        });
      });
      return;
    }

    if (node.type === "note") {
      return;
    }

    const realityId = node.data.id || node.id;
    REALITY_BEHAVIOR_FIELDS.forEach((field) => {
      appendArrayUsage({
        src,
        values: node.data[field],
        scope: "reality",
        containerId: realityId,
        containerLabel: `Reality:${realityId}`,
        field,
        usages,
      });
    });
  });

  state.transitions.forEach((transition) => {
    if (transition.condition?.src === src) {
      usages.push({
        scope: "transition",
        containerId: transition.id,
        field: "condition",
        label: `Transition:${transition.id} · condition`,
      });
    }

    TRANSITION_ARRAY_FIELDS.forEach((field) => {
      appendArrayUsage({
        src,
        values: transition[field],
        scope: "transition",
        containerId: transition.id,
        containerLabel: `Transition:${transition.id}`,
        field,
        usages,
      });
    });
  });

  return {
    total: usages.length,
    locations: usages,
  };
};

export const removeBehaviorReferences = (
  src: string,
  state: BehaviorStateSlices,
): RemoveReferencesResult => {
  let removedCount = 0;
  let machineChanged = false;

  const nextMachineUC = { ...state.machineConfig.universalConstants };
  UNIVERSAL_CONSTANT_FIELDS.forEach((field) => {
    const result = removeFromRequiredBehaviorArray(nextMachineUC[field], src);
    if (result.removed === 0) {
      return;
    }

    machineChanged = true;
    removedCount += result.removed;
    nextMachineUC[field] = result.next;
  });

  const nextMachineConfig = machineChanged
    ? { ...state.machineConfig, universalConstants: nextMachineUC }
    : state.machineConfig;

  let nodesChanged = false;
  const nextNodes = state.nodes.map((node) => {
    if (node.type === "universe") {
      if (!node.data.universalConstants) {
        return node;
      }

      let universeChanged = false;
      const nextUC = { ...node.data.universalConstants };

      UNIVERSAL_CONSTANT_FIELDS.forEach((field) => {
        const result = removeFromRequiredBehaviorArray(nextUC[field], src);
        if (result.removed === 0) {
          return;
        }

        universeChanged = true;
        removedCount += result.removed;
        nextUC[field] = result.next;
      });

      if (!universeChanged) {
        return node;
      }

      nodesChanged = true;
      return {
        ...node,
        data: {
          ...node.data,
          universalConstants: nextUC,
        },
      };
    }

    if (node.type === "note") {
      return node;
    }

    let realityChanged = false;
    const nextRealityData = { ...node.data };

    REALITY_BEHAVIOR_FIELDS.forEach((field) => {
      const result = removeFromOptionalBehaviorArray(nextRealityData[field], src);
      if (result.removed === 0) {
        return;
      }

      realityChanged = true;
      removedCount += result.removed;
      nextRealityData[field] = result.next;
    });

    if (!realityChanged) {
      return node;
    }

    nodesChanged = true;
    return {
      ...node,
      data: nextRealityData,
    };
  });

  let transitionsChanged = false;
  const nextTransitions = state.transitions.map((transition) => {
    let hasChanges = false;
    let nextTransition = transition;

    if (transition.condition?.src === src) {
      removedCount += 1;
      hasChanges = true;
      nextTransition = {
        ...nextTransition,
        condition: undefined,
      };
    }

    TRANSITION_ARRAY_FIELDS.forEach((field) => {
      const result = removeFromRequiredBehaviorArray(nextTransition[field], src);
      if (result.removed === 0) {
        return;
      }

      removedCount += result.removed;
      hasChanges = true;
      nextTransition = {
        ...nextTransition,
        [field]: result.next,
      };
    });

    if (!hasChanges) {
      return transition;
    }

    transitionsChanged = true;
    return nextTransition;
  });

  return {
    nextNodes: nodesChanged ? nextNodes : state.nodes,
    nextTransitions: transitionsChanged ? nextTransitions : state.transitions,
    nextMachineConfig,
    removedCount,
  };
};
