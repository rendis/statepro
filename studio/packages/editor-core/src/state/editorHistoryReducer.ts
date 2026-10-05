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
import deepEqual from "fast-deep-equal";

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

const snapshotsEqual = (
  left: EditorHistorySnapshot,
  right: EditorHistorySnapshot,
): boolean => {
  // Unchanged graph sections keep their identity across immutable reducer edits.
  // Only traverse sections which actually changed, rather than serializing the graph.
  // Optional fields may be absent from a supplied checkpoint. Comparing only
  // its own keys would miss a present imported model that must be cleared.
  const keys = new Set([...Object.keys(left), ...Object.keys(right)]);
  return (Array.from(keys) as Array<keyof EditorHistorySnapshot>).every(
    (key) => Object.is(left[key], right[key]) || deepEqual(left[key], right[key]),
  );
};

// Snapshots from createHistorySnapshot are already detached copies. The first
// commit takes ownership of them instead of cloning the whole graph again.
const detachedSnapshots = new WeakSet<EditorHistorySnapshot>();

const pushPast = (
  past: EditorHistorySnapshot[],
  snapshot: EditorHistorySnapshot,
): EditorHistorySnapshot[] => {
  const stored = detachedSnapshots.delete(snapshot) ? snapshot : clone(snapshot);
  const nextPast = [...past, stored];
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
  // Clone the checkpoint once, including all graph sections. The restored editor
  // still owns a detached copy; mutating it cannot change past/future entries.
  const restored = clone(snapshot);
  return {
    ...present,
    nodes: restored.nodes,
    transitions: restored.transitions,
    machineConfig: restored.machineConfig,
    registry: restored.registry,
    metadataPackRegistry: restored.metadataPackRegistry,
    metadataPackBindings: restored.metadataPackBindings,
    lastImportedMachine: restored.lastImportedMachine,
    isDirtyFromImport: restored.isDirtyFromImport,
    selectedElement: validateSelectedElement(present.selectedElement, restored.nodes, restored.transitions),
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

/**
 * Returns a snapshot detached from `state`. Committing it with `commit-snapshot`
 * transfers ownership to the history without another clone; do not mutate it afterwards.
 */
export const createHistorySnapshot = (state: EditorState): EditorHistorySnapshot => {
  const snapshot = clone(toSnapshot(state));
  detachedSnapshots.add(snapshot);
  return snapshot;
};

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

      if (nextPresent === basePresent) return state;

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
