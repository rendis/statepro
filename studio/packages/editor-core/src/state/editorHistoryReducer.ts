import { createInitialEditorState } from "../model/defaults";
import type {
  BehaviorRegistryItem,
  EditorNode,
  EditorState,
  EditorTransition,
  MachineConfig,
  MetadataPackBindingMap,
  MetadataPackRegistry,
  SelectedElementRef,
  StateProMachine,
} from "../types";
import { editorReducer, type EditorAction } from "./editorReducer";

export const HISTORY_LIMIT = 100;
export const COALESCE_WINDOW_MS = 600;

export type HistoryApplyMode = "record" | "coalesce" | "silent";

export interface EditorHistorySnapshot {
  nodes: EditorNode[];
  transitions: EditorTransition[];
  machineConfig: MachineConfig;
  registry: BehaviorRegistryItem[];
  metadataPackRegistry: MetadataPackRegistry;
  metadataPackBindings: MetadataPackBindingMap;
  lastImportedMachine?: StateProMachine;
  isDirtyFromImport: boolean;
}

export interface EditorHistoryState {
  past: EditorHistorySnapshot[];
  present: EditorState;
  future: EditorHistorySnapshot[];
  lastCoalesceGroup: string | null;
  lastCoalesceAt: number | null;
}

export type EditorHistoryAction =
  | {
      type: "apply-editor-action";
      action: EditorAction;
      mode?: HistoryApplyMode;
      group?: string;
      markDirtyFromImport?: boolean;
      now?: number;
    }
  | {
      type: "commit-snapshot";
      payload: EditorHistorySnapshot;
    }
  | { type: "undo" }
  | { type: "redo" }
  | { type: "reset-history"; payload: EditorState };

const clone = <T>(value: T): T => structuredClone(value);

const toSnapshot = (state: EditorState): EditorHistorySnapshot => ({
  nodes: state.nodes,
  transitions: state.transitions,
  machineConfig: state.machineConfig,
  registry: state.registry,
  metadataPackRegistry: state.metadataPackRegistry,
  metadataPackBindings: state.metadataPackBindings,
  lastImportedMachine: state.lastImportedMachine,
  isDirtyFromImport: state.isDirtyFromImport,
});

const serializeSnapshot = (snapshot: EditorHistorySnapshot): string => JSON.stringify(snapshot);

const snapshotsEqual = (
  left: EditorHistorySnapshot,
  right: EditorHistorySnapshot,
): boolean => {
  return serializeSnapshot(left) === serializeSnapshot(right);
};

const pushPast = (
  past: EditorHistorySnapshot[],
  snapshot: EditorHistorySnapshot,
): EditorHistorySnapshot[] => {
  const nextPast = [...past, clone(snapshot)];
  if (nextPast.length <= HISTORY_LIMIT) {
    return nextPast;
  }
  return nextPast.slice(nextPast.length - HISTORY_LIMIT);
};

const validateSelectedElement = (
  selectedElement: SelectedElementRef | null,
  nodes: EditorNode[],
  transitions: EditorTransition[],
): SelectedElementRef | null => {
  if (!selectedElement) {
    return null;
  }

  if (selectedElement.kind === "node") {
    const hasNode = nodes.some((node) => node.id === selectedElement.id);
    return hasNode ? selectedElement : null;
  }

  const hasTransition = transitions.some((transition) => transition.id === selectedElement.id);
  return hasTransition ? selectedElement : null;
};

const applySnapshotToPresent = (
  present: EditorState,
  snapshot: EditorHistorySnapshot,
): EditorState => {
  const nextNodes = clone(snapshot.nodes);
  const nextTransitions = clone(snapshot.transitions);
  return {
    ...present,
    nodes: nextNodes,
    transitions: nextTransitions,
    machineConfig: clone(snapshot.machineConfig),
    registry: clone(snapshot.registry),
    metadataPackRegistry: clone(snapshot.metadataPackRegistry),
    metadataPackBindings: clone(snapshot.metadataPackBindings),
    lastImportedMachine: snapshot.lastImportedMachine
      ? clone(snapshot.lastImportedMachine)
      : undefined,
    isDirtyFromImport: snapshot.isDirtyFromImport,
    selectedElement: validateSelectedElement(present.selectedElement, nextNodes, nextTransitions),
  };
};

const shouldAutoMarkDirtyFromImport = (
  action: EditorAction,
  present: EditorState,
  markDirtyFromImport: boolean,
): boolean => {
  if (!markDirtyFromImport) {
    return false;
  }

  if (!present.lastImportedMachine || present.isDirtyFromImport) {
    return false;
  }

  switch (action.type) {
    case "reset":
    case "hydrate-from-import":
    case "mark-dirty-from-import":
    case "set-selected-element":
    case "set-node-size":
    case "set-node-sizes":
    case "update-node-sizes":
      return false;
    default:
      return true;
  }
};

export const createHistorySnapshot = (state: EditorState): EditorHistorySnapshot =>
  clone(toSnapshot(state));

export const createInitialEditorHistoryState = (
  initialState: EditorState = createInitialEditorState(),
): EditorHistoryState => ({
  past: [],
  present: initialState,
  future: [],
  lastCoalesceGroup: null,
  lastCoalesceAt: null,
});

export const editorHistoryReducer = (
  state: EditorHistoryState,
  action: EditorHistoryAction,
): EditorHistoryState => {
  switch (action.type) {
    case "apply-editor-action": {
      const mode = action.mode || "record";
      const basePresent = state.present;
      const presentWithDirty = shouldAutoMarkDirtyFromImport(
        action.action,
        basePresent,
        action.markDirtyFromImport ?? true,
      )
        ? editorReducer(basePresent, { type: "mark-dirty-from-import" })
        : basePresent;
      const nextPresent = editorReducer(presentWithDirty, action.action);

      if (mode === "silent") {
        return {
          ...state,
          present: nextPresent,
          lastCoalesceGroup: null,
          lastCoalesceAt: null,
        };
      }

      const baseSnapshot = toSnapshot(basePresent);
      const nextSnapshot = toSnapshot(nextPresent);

      if (snapshotsEqual(baseSnapshot, nextSnapshot)) {
        return {
          ...state,
          present: nextPresent,
        };
      }

      if (mode === "coalesce") {
        const now = action.now ?? Date.now();
        const canCoalesce =
          Boolean(action.group) &&
          state.lastCoalesceGroup === action.group &&
          state.lastCoalesceAt !== null &&
          now - state.lastCoalesceAt <= COALESCE_WINDOW_MS &&
          state.past.length > 0;

        return {
          ...state,
          present: nextPresent,
          future: [],
          past: canCoalesce ? state.past : pushPast(state.past, baseSnapshot),
          lastCoalesceGroup: action.group || null,
          lastCoalesceAt: now,
        };
      }

      return {
        ...state,
        present: nextPresent,
        past: pushPast(state.past, baseSnapshot),
        future: [],
        lastCoalesceGroup: null,
        lastCoalesceAt: null,
      };
    }

    case "commit-snapshot": {
      const currentSnapshot = toSnapshot(state.present);
      if (snapshotsEqual(action.payload, currentSnapshot)) {
        return {
          ...state,
          lastCoalesceGroup: null,
          lastCoalesceAt: null,
        };
      }

      return {
        ...state,
        past: pushPast(state.past, action.payload),
        future: [],
        lastCoalesceGroup: null,
        lastCoalesceAt: null,
      };
    }

    case "undo": {
      if (state.past.length === 0) {
        return state;
      }

      const previousSnapshot = state.past[state.past.length - 1];
      const currentSnapshot = toSnapshot(state.present);
      return {
        ...state,
        present: applySnapshotToPresent(state.present, previousSnapshot),
        past: state.past.slice(0, -1),
        future: [clone(currentSnapshot), ...state.future],
        lastCoalesceGroup: null,
        lastCoalesceAt: null,
      };
    }

    case "redo": {
      if (state.future.length === 0) {
        return state;
      }

      const nextSnapshot = state.future[0];
      const currentSnapshot = toSnapshot(state.present);
      return {
        ...state,
        present: applySnapshotToPresent(state.present, nextSnapshot),
        past: pushPast(state.past, currentSnapshot),
        future: state.future.slice(1),
        lastCoalesceGroup: null,
        lastCoalesceAt: null,
      };
    }

    case "reset-history": {
      return createInitialEditorHistoryState(action.payload);
    }

    default: {
      return state;
    }
  }
};
