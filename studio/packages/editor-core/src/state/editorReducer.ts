import { createInitialEditorState } from "../model/defaults";
import { removeTransitionsReferencingDeletedNodes } from "../utils";
import type {
  BehaviorRegistryItem,
  EditorNode,
  EditorState,
  MetadataPackBindingMap,
  MetadataPackRegistry,
  NodeSize,
  SelectedElementRef,
  EditorTransition,
  UniversalConstants,
} from "../types";

export type EditorAction =
  | { type: "reset" }
  | { type: "hydrate-from-import"; payload: EditorState }
  | { type: "mark-dirty-from-import" }
  | { type: "set-selected-element"; payload: SelectedElementRef | null }
  | { type: "set-node-size"; payload: { nodeId: string; size: NodeSize } }
  | { type: "set-node-sizes"; payload: Record<string, NodeSize> }
  | {
      type: "update-node-sizes";
      payload: (prev: Record<string, NodeSize>) => Record<string, NodeSize>;
    }
  | { type: "set-registry"; payload: BehaviorRegistryItem[] }
  | { type: "set-metadata-pack-registry"; payload: MetadataPackRegistry }
  | { type: "set-metadata-pack-bindings"; payload: MetadataPackBindingMap }
  | { type: "set-machine-config"; payload: EditorState["machineConfig"] }
  | {
      type: "update-machine-config";
      payload: {
        field: keyof EditorState["machineConfig"];
        value: EditorState["machineConfig"][keyof EditorState["machineConfig"]];
      };
    }
  | {
      type: "update-machine-universal-constants";
      payload: {
        field: keyof UniversalConstants;
        value: UniversalConstants[keyof UniversalConstants];
      };
    }
  | { type: "set-nodes"; payload: EditorNode[] }
  | { type: "set-transitions"; payload: EditorTransition[] }
  | {
      type: "add-universe";
      payload: { universe: Extract<EditorNode, { type: "universe" }> };
    }
  | {
      type: "add-reality";
      payload: {
        reality: Extract<EditorNode, { type: "reality" }>;
        expandedUniverse?: { id: string; w: number; h: number };
      };
    }
  | { type: "add-transition"; payload: { transition: EditorTransition } }
  | {
      type: "update-node-data";
      payload: { nodeId: string; patch: Record<string, unknown> };
    }
  | {
      type: "update-transition";
      payload: { transitionId: string; patch: Partial<EditorTransition> };
    }
  | {
      type: "delete-element";
      payload: { kind: "node" | "transition"; id: string };
    }
  | {
      type: "clone-reality";
      payload: {
        reality: Extract<EditorNode, { type: "reality" }>;
        expandedUniverse?: { id: string; w: number; h: number };
      };
    }
  | {
      type: "clone-universe";
      payload: {
        universe: Extract<EditorNode, { type: "universe" }>;
        realities: Array<Extract<EditorNode, { type: "reality" }>>;
        transitions: EditorTransition[];
      };
    };

const clearSelectionIfInvalid = (state: EditorState): EditorState => {
  if (!state.selectedElement) return state;

  if (state.selectedElement.kind === "node") {
    const exists = state.nodes.some((node) => node.id === state.selectedElement?.id);
    return exists ? state : { ...state, selectedElement: null };
  }

  const transitionExists = state.transitions.some(
    (transition) => transition.id === state.selectedElement?.id,
  );
  return transitionExists ? state : { ...state, selectedElement: null };
};

const enforceSingleInitialReality = (nodes: EditorNode[], updatedNodeId: string): EditorNode[] => {
  const updatedNode = nodes.find((node) => node.id === updatedNodeId);
  if (!updatedNode || updatedNode.type !== "reality" || !updatedNode.data.isInitial) {
    return nodes;
  }

  return nodes.map((node) => {
    if (
      node.type === "reality" &&
      node.id !== updatedNodeId &&
      node.data.universeId === updatedNode.data.universeId &&
      node.data.isInitial
    ) {
      return {
        ...node,
        data: {
          ...node.data,
          isInitial: false,
        },
      };
    }
    return node;
  });
};

export const editorReducer = (state: EditorState, action: EditorAction): EditorState => {
  switch (action.type) {
    case "reset": {
      return createInitialEditorState();
    }

    case "hydrate-from-import": {
      return clearSelectionIfInvalid(action.payload);
    }

    case "mark-dirty-from-import": {
      if (!state.lastImportedMachine || state.isDirtyFromImport) {
        return state;
      }

      return {
        ...state,
        isDirtyFromImport: true,
      };
    }

    case "set-selected-element": {
      return {
        ...state,
        selectedElement: action.payload,
      };
    }

    case "set-node-size": {
      return {
        ...state,
        nodeSizes: {
          ...state.nodeSizes,
          [action.payload.nodeId]: action.payload.size,
        },
      };
    }

    case "set-node-sizes": {
      return {
        ...state,
        nodeSizes: action.payload,
      };
    }

    case "update-node-sizes": {
      return {
        ...state,
        nodeSizes: action.payload(state.nodeSizes),
      };
    }

    case "set-registry": {
      return {
        ...state,
        registry: action.payload,
      };
    }

    case "set-metadata-pack-registry": {
      return {
        ...state,
        metadataPackRegistry: action.payload,
      };
    }

    case "set-metadata-pack-bindings": {
      return {
        ...state,
        metadataPackBindings: action.payload,
      };
    }

    case "set-machine-config": {
      return {
        ...state,
        machineConfig: action.payload,
      };
    }

    case "update-machine-config": {
      return {
        ...state,
        machineConfig: {
          ...state.machineConfig,
          [action.payload.field]: action.payload.value,
        },
      };
    }

    case "update-machine-universal-constants": {
      return {
        ...state,
        machineConfig: {
          ...state.machineConfig,
          universalConstants: {
            ...state.machineConfig.universalConstants,
            [action.payload.field]: action.payload.value,
          },
        },
      };
    }

    case "set-nodes": {
      return clearSelectionIfInvalid({
        ...state,
        nodes: action.payload,
      });
    }

    case "set-transitions": {
      return clearSelectionIfInvalid({
        ...state,
        transitions: action.payload,
      });
    }

    case "add-universe": {
      return {
        ...state,
        nodes: [...state.nodes, action.payload.universe],
        selectedElement: {
          kind: "node",
          id: action.payload.universe.id,
        },
      };
    }

    case "add-reality": {
      const nextNodes = state.nodes.map((node) => {
        if (action.payload.expandedUniverse && node.id === action.payload.expandedUniverse.id) {
          return {
            ...node,
            w: action.payload.expandedUniverse.w,
            h: action.payload.expandedUniverse.h,
          };
        }
        return node;
      });

      nextNodes.push(action.payload.reality);

      return {
        ...state,
        nodes: enforceSingleInitialReality(nextNodes, action.payload.reality.id),
        selectedElement: {
          kind: "node",
          id: action.payload.reality.id,
        },
      };
    }

    case "add-transition": {
      return {
        ...state,
        transitions: [...state.transitions, action.payload.transition],
        selectedElement: {
          kind: "transition",
          id: action.payload.transition.id,
        },
      };
    }

    case "update-node-data": {
      let nextNodes = state.nodes.map((node) => {
        if (node.id !== action.payload.nodeId) {
          return node;
        }

        return {
          ...node,
          data: {
            ...node.data,
            ...action.payload.patch,
          },
        } as EditorNode;
      });

      nextNodes = enforceSingleInitialReality(nextNodes, action.payload.nodeId);

      return {
        ...state,
        nodes: nextNodes,
      };
    }

    case "update-transition": {
      return {
        ...state,
        transitions: state.transitions.map((transition) => {
          if (transition.id !== action.payload.transitionId) {
            return transition;
          }
          return {
            ...transition,
            ...action.payload.patch,
          };
        }),
      };
    }

    case "delete-element": {
      if (action.payload.kind === "transition") {
        return clearSelectionIfInvalid({
          ...state,
          transitions: state.transitions.filter((transition) => transition.id !== action.payload.id),
        });
      }

      const node = state.nodes.find((candidate) => candidate.id === action.payload.id);
      if (!node) {
        return state;
      }

      let nodeIdsToDelete = [node.id];
      if (node.type === "universe") {
        const childRealityIds = state.nodes
          .filter(
            (candidate) =>
              candidate.type === "reality" && candidate.data.universeId === node.id,
          )
          .map((candidate) => candidate.id);
        nodeIdsToDelete = [...nodeIdsToDelete, ...childRealityIds];
      }

      const nextNodes = state.nodes.filter((candidate) => !nodeIdsToDelete.includes(candidate.id));
      const nextTransitions = removeTransitionsReferencingDeletedNodes(
        state.transitions,
        state.nodes,
        nodeIdsToDelete,
      );

      return clearSelectionIfInvalid({
        ...state,
        nodes: nextNodes,
        transitions: nextTransitions,
      });
    }

    case "clone-reality": {
      const nextNodes = [...state.nodes, action.payload.reality].map((node) => {
        if (action.payload.expandedUniverse && node.id === action.payload.expandedUniverse.id) {
          return {
            ...node,
            w: action.payload.expandedUniverse.w,
            h: action.payload.expandedUniverse.h,
          };
        }

        return node;
      });

      return {
        ...state,
        nodes: nextNodes,
        selectedElement: {
          kind: "node",
          id: action.payload.reality.id,
        },
      };
    }

    case "clone-universe": {
      return {
        ...state,
        nodes: [...state.nodes, action.payload.universe, ...action.payload.realities],
        transitions: [...state.transitions, ...action.payload.transitions],
        selectedElement: {
          kind: "node",
          id: action.payload.universe.id,
        },
      };
    }

    default: {
      return state;
    }
  }
};
