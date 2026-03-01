import type {
  AnchoredNoteData,
  ApplyStudioLayoutResult,
  EditorNode,
  EditorState,
  GlobalNoteData,
  ParseStudioLayoutResult,
  StudioLayoutDocument,
  StudioLayoutIssue,
} from "../types";
import {
  buildRealityEntityRef,
  buildTransitionEntityRef,
  buildUniverseEntityRef,
  normalizeMetadataPackBindingMap,
  normalizeMetadataPackRegistry,
  normalizeTransitionsOrder,
} from "../utils";

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const pushIssue = (
  issues: StudioLayoutIssue[],
  severity: StudioLayoutIssue["severity"],
  field: string,
  message: string,
): void => {
  issues.push({
    severity,
    field,
    message,
  });
};

const toFiniteNumber = (
  value: unknown,
  issues: StudioLayoutIssue[],
  field: string,
): number | null => {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    pushIssue(issues, "error", field, "Must be a finite number.");
    return null;
  }
  return value;
};

const toStringValue = (
  value: unknown,
  issues: StudioLayoutIssue[],
  field: string,
): string | null => {
  if (typeof value !== "string" || !value.trim()) {
    pushIssue(issues, "error", field, "Must be a non-empty string.");
    return null;
  }
  return value;
};

const toAnchoredNote = (
  value: unknown,
  issues: StudioLayoutIssue[],
  field: string,
): AnchoredNoteData | null => {
  if (value == null) {
    return null;
  }
  if (!isRecord(value)) {
    pushIssue(issues, "error", field, "Note must be an object or null.");
    return null;
  }

  const text = typeof value.text === "string" ? value.text : null;
  const colorIndex =
    typeof value.colorIndex === "number" && Number.isFinite(value.colorIndex)
      ? Math.max(0, Math.trunc(value.colorIndex))
      : null;

  if (text === null || colorIndex === null) {
    pushIssue(
      issues,
      "error",
      field,
      "Note requires 'text' (string) and 'colorIndex' (finite number).",
    );
    return null;
  }

  return {
    text,
    colorIndex,
  };
};

const toGlobalNoteData = (
  value: unknown,
  issues: StudioLayoutIssue[],
  field: string,
): GlobalNoteData | null => {
  if (!isRecord(value)) {
    pushIssue(issues, "error", field, "Global note data must be an object.");
    return null;
  }

  const base = toAnchoredNote(value, issues, field);
  if (!base) {
    return null;
  }

  if ("isCollapsed" in value && typeof value.isCollapsed !== "boolean") {
    pushIssue(issues, "error", `${field}.isCollapsed`, "Must be a boolean when present.");
    return null;
  }

  return {
    ...base,
    isCollapsed: typeof value.isCollapsed === "boolean" ? value.isCollapsed : false,
  };
};

const normalizeStudioLayoutDocument = (raw: unknown): ParseStudioLayoutResult => {
  const issues: StudioLayoutIssue[] = [];
  if (!isRecord(raw)) {
    return {
      document: null,
      issues: [
        {
          severity: "error",
          field: "root",
          message: "Layout JSON must be an object.",
        },
      ],
      canImport: false,
    };
  }

  const version = raw.version;
  if (version !== 1) {
    pushIssue(issues, "error", "version", "Layout version must be 1.");
  }

  const machineRefRaw = raw.machineRef;
  const machineId = isRecord(machineRefRaw)
    ? toStringValue(machineRefRaw.id, issues, "machineRef.id")
    : null;
  const machineCanonicalName = isRecord(machineRefRaw)
    ? toStringValue(machineRefRaw.canonicalName, issues, "machineRef.canonicalName")
    : null;
  const machineVersion = isRecord(machineRefRaw)
    ? toStringValue(machineRefRaw.version, issues, "machineRef.version")
    : null;
  if (!isRecord(machineRefRaw)) {
    pushIssue(issues, "error", "machineRef", "machineRef must be an object.");
  }

  const nodesRaw = raw.nodes;
  const universeEntriesRaw = isRecord(nodesRaw) ? nodesRaw.universes : undefined;
  const realityEntriesRaw = isRecord(nodesRaw) ? nodesRaw.realities : undefined;
  const globalNotesRaw = isRecord(nodesRaw) ? nodesRaw.globalNotes : undefined;

  if (!isRecord(nodesRaw)) {
    pushIssue(issues, "error", "nodes", "nodes must be an object.");
  }
  if (!Array.isArray(universeEntriesRaw)) {
    pushIssue(issues, "error", "nodes.universes", "nodes.universes must be an array.");
  }
  if (!Array.isArray(realityEntriesRaw)) {
    pushIssue(issues, "error", "nodes.realities", "nodes.realities must be an array.");
  }
  if (!Array.isArray(globalNotesRaw)) {
    pushIssue(issues, "error", "nodes.globalNotes", "nodes.globalNotes must be an array.");
  }

  const universes: StudioLayoutDocument["nodes"]["universes"] = [];
  (Array.isArray(universeEntriesRaw) ? universeEntriesRaw : []).forEach((entry, index) => {
    const field = `nodes.universes[${index}]`;
    if (!isRecord(entry)) {
      pushIssue(issues, "error", field, "Universe layout entry must be an object.");
      return;
    }

    const entityRef = toStringValue(entry.entityRef, issues, `${field}.entityRef`);
    const x = toFiniteNumber(entry.x, issues, `${field}.x`);
    const y = toFiniteNumber(entry.y, issues, `${field}.y`);
    const w = toFiniteNumber(entry.w, issues, `${field}.w`);
    const h = toFiniteNumber(entry.h, issues, `${field}.h`);
    const note = toAnchoredNote(entry.note, issues, `${field}.note`);
    if (!entityRef || x === null || y === null || w === null || h === null) {
      return;
    }

    universes.push({
      entityRef,
      x,
      y,
      w,
      h,
      note,
    });
  });

  const realities: StudioLayoutDocument["nodes"]["realities"] = [];
  (Array.isArray(realityEntriesRaw) ? realityEntriesRaw : []).forEach((entry, index) => {
    const field = `nodes.realities[${index}]`;
    if (!isRecord(entry)) {
      pushIssue(issues, "error", field, "Reality layout entry must be an object.");
      return;
    }

    const entityRef = toStringValue(entry.entityRef, issues, `${field}.entityRef`);
    const x = toFiniteNumber(entry.x, issues, `${field}.x`);
    const y = toFiniteNumber(entry.y, issues, `${field}.y`);
    const note = toAnchoredNote(entry.note, issues, `${field}.note`);
    if (!entityRef || x === null || y === null) {
      return;
    }

    realities.push({
      entityRef,
      x,
      y,
      note,
    });
  });

  const globalNotes: StudioLayoutDocument["nodes"]["globalNotes"] = [];
  (Array.isArray(globalNotesRaw) ? globalNotesRaw : []).forEach((entry, index) => {
    const field = `nodes.globalNotes[${index}]`;
    if (!isRecord(entry)) {
      pushIssue(issues, "error", field, "Global note entry must be an object.");
      return;
    }

    const x = toFiniteNumber(entry.x, issues, `${field}.x`);
    const y = toFiniteNumber(entry.y, issues, `${field}.y`);
    const data = toGlobalNoteData(entry.data, issues, `${field}.data`);
    if (x === null || y === null || !data) {
      return;
    }

    globalNotes.push({
      x,
      y,
      data,
    });
  });

  const transitionsRaw = raw.transitions;
  if (!Array.isArray(transitionsRaw)) {
    pushIssue(issues, "error", "transitions", "transitions must be an array.");
  }
  const transitions: StudioLayoutDocument["transitions"] = [];
  (Array.isArray(transitionsRaw) ? transitionsRaw : []).forEach((entry, index) => {
    const field = `transitions[${index}]`;
    if (!isRecord(entry)) {
      pushIssue(issues, "error", field, "Transition layout entry must be an object.");
      return;
    }

    const entityRef = toStringValue(entry.entityRef, issues, `${field}.entityRef`);
    const note = toAnchoredNote(entry.note, issues, `${field}.note`);
    let visualOffset: { x: number; y: number } | undefined;
    if (entry.visualOffset !== undefined) {
      if (!isRecord(entry.visualOffset)) {
        pushIssue(issues, "error", `${field}.visualOffset`, "visualOffset must be an object.");
      } else {
        const x = toFiniteNumber(entry.visualOffset.x, issues, `${field}.visualOffset.x`);
        const y = toFiniteNumber(entry.visualOffset.y, issues, `${field}.visualOffset.y`);
        if (x !== null && y !== null) {
          visualOffset = { x, y };
        }
      }
    }

    if (!entityRef) {
      return;
    }

    transitions.push({
      entityRef,
      visualOffset,
      note,
    });
  });

  const packsRaw = raw.packs;
  if (!isRecord(packsRaw)) {
    pushIssue(issues, "error", "packs", "packs must be an object.");
  }
  const rawPackRegistry = isRecord(packsRaw) ? packsRaw.packRegistry : undefined;
  const rawBindings = isRecord(packsRaw) ? packsRaw.bindings : undefined;
  if (!Array.isArray(rawPackRegistry)) {
    pushIssue(issues, "error", "packs.packRegistry", "packs.packRegistry must be an array.");
  }
  if (!isRecord(rawBindings)) {
    pushIssue(issues, "error", "packs.bindings", "packs.bindings must be an object.");
  }
  const packRegistry = normalizeMetadataPackRegistry(rawPackRegistry);
  const bindings = normalizeMetadataPackBindingMap(rawBindings);

  const hasErrors = issues.some((issue) => issue.severity === "error");
  if (
    hasErrors ||
    !machineId ||
    !machineCanonicalName ||
    !machineVersion
  ) {
    return {
      document: null,
      issues,
      canImport: false,
    };
  }

  return {
    document: {
      version: 1,
      machineRef: {
        id: machineId,
        canonicalName: machineCanonicalName,
        version: machineVersion,
      },
      nodes: {
        universes,
        realities,
        globalNotes,
      },
      transitions,
      packs: {
        packRegistry,
        bindings,
      },
    },
    issues,
    canImport: true,
  };
};

export const parseStudioLayout = (layoutJson: string): ParseStudioLayoutResult => {
  try {
    const parsed = JSON.parse(layoutJson) as unknown;
    return normalizeStudioLayoutDocument(parsed);
  } catch {
    return {
      document: null,
      issues: [
        {
          severity: "error",
          field: "root",
          message: "Invalid JSON.",
        },
      ],
      canImport: false,
    };
  }
};

const createGlobalNoteNodeId = (usedNodeIds: Set<string>, index: number): string => {
  let candidate = `global-note-${index + 1}`;
  let suffix = 1;
  while (usedNodeIds.has(candidate)) {
    suffix += 1;
    candidate = `global-note-${index + 1}-${suffix}`;
  }
  usedNodeIds.add(candidate);
  return candidate;
};

export const applyStudioLayoutDocument = (
  state: EditorState,
  document: StudioLayoutDocument,
): ApplyStudioLayoutResult => {
  const issues: StudioLayoutIssue[] = [];
  if (
    document.machineRef.id !== state.machineConfig.id ||
    document.machineRef.canonicalName !== state.machineConfig.canonicalName ||
    document.machineRef.version !== state.machineConfig.version
  ) {
    pushIssue(
      issues,
      "warning",
      "machineRef",
      "Layout machineRef differs from current model. Applying by logical entity refs.",
    );
  }

  const nextNodesNoGlobalNotes = state.nodes
    .filter((node) => node.type !== "note")
    .map((node) => structuredClone(node));
  const nextTransitions = normalizeTransitionsOrder(
    state.transitions.map((transition) => structuredClone(transition)),
  );

  const universeByNodeId = new Map(
    nextNodesNoGlobalNotes
      .filter(
        (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
      )
      .map((node) => [node.id, node]),
  );
  const realityByNodeId = new Map(
    nextNodesNoGlobalNotes
      .filter(
        (node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality",
      )
      .map((node) => [node.id, node]),
  );

  const universeByEntityRef = new Map<string, Extract<EditorNode, { type: "universe" }>>();
  universeByNodeId.forEach((node) => {
    universeByEntityRef.set(buildUniverseEntityRef(node.data.id), node);
  });

  const realityByEntityRef = new Map<string, Extract<EditorNode, { type: "reality" }>>();
  realityByNodeId.forEach((node) => {
    const parentUniverse = universeByNodeId.get(node.data.universeId);
    if (!parentUniverse) {
      return;
    }
    realityByEntityRef.set(
      buildRealityEntityRef(parentUniverse.data.id, node.data.id),
      node,
    );
  });

  const transitionByEntityRef = new Map<string, (typeof nextTransitions)[number]>();
  nextTransitions.forEach((transition) => {
    const sourceReality = realityByNodeId.get(transition.sourceRealityId);
    if (!sourceReality) {
      return;
    }
    const parentUniverse = universeByNodeId.get(sourceReality.data.universeId);
    if (!parentUniverse) {
      return;
    }
    transitionByEntityRef.set(
      buildTransitionEntityRef(
        parentUniverse.data.id,
        sourceReality.data.id,
        transition.triggerKind,
        transition.eventName,
        transition.order,
      ),
      transition,
    );
  });

  document.nodes.universes.forEach((snapshot, index) => {
    const node = universeByEntityRef.get(snapshot.entityRef);
    if (!node) {
      pushIssue(
        issues,
        "warning",
        `nodes.universes[${index}].entityRef`,
        `Unknown universe entityRef '${snapshot.entityRef}'.`,
      );
      return;
    }

    node.x = snapshot.x;
    node.y = snapshot.y;
    node.w = snapshot.w;
    node.h = snapshot.h;
    node.data.note = snapshot.note ? structuredClone(snapshot.note) : null;
  });

  document.nodes.realities.forEach((snapshot, index) => {
    const node = realityByEntityRef.get(snapshot.entityRef);
    if (!node) {
      pushIssue(
        issues,
        "warning",
        `nodes.realities[${index}].entityRef`,
        `Unknown reality entityRef '${snapshot.entityRef}'.`,
      );
      return;
    }

    node.x = snapshot.x;
    node.y = snapshot.y;
    node.data.note = snapshot.note ? structuredClone(snapshot.note) : null;
  });

  document.transitions.forEach((snapshot, index) => {
    const transition = transitionByEntityRef.get(snapshot.entityRef);
    if (!transition) {
      pushIssue(
        issues,
        "warning",
        `transitions[${index}].entityRef`,
        `Unknown transition entityRef '${snapshot.entityRef}'.`,
      );
      return;
    }

    transition.visualOffset = snapshot.visualOffset
      ? {
          x: snapshot.visualOffset.x,
          y: snapshot.visualOffset.y,
        }
      : undefined;
    transition.note = snapshot.note ? structuredClone(snapshot.note) : null;
  });

  const usedNodeIds = new Set(nextNodesNoGlobalNotes.map((node) => node.id));
  const nextGlobalNotes: Array<Extract<EditorNode, { type: "note" }>> = document.nodes.globalNotes.map(
    (snapshot, index) => ({
      id: createGlobalNoteNodeId(usedNodeIds, index),
      type: "note",
      x: snapshot.x,
      y: snapshot.y,
      data: structuredClone(snapshot.data),
    }),
  );

  const packRegistryChanged =
    JSON.stringify(state.metadataPackRegistry) !== JSON.stringify(document.packs.packRegistry);
  const packBindingsChanged =
    JSON.stringify(state.metadataPackBindings) !== JSON.stringify(document.packs.bindings);

  return {
    state: {
      ...state,
      nodes: [...nextNodesNoGlobalNotes, ...nextGlobalNotes],
      transitions: nextTransitions,
      selectedElement: null,
      metadataPackRegistry: structuredClone(document.packs.packRegistry),
      metadataPackBindings: structuredClone(document.packs.bindings),
      isDirtyFromImport:
        state.isDirtyFromImport || packRegistryChanged || packBindingsChanged,
    },
    issues,
  };
};

export const parseAndApplyStudioLayout = (
  layoutJson: string,
  state: EditorState,
): {
  state: EditorState;
  issues: StudioLayoutIssue[];
  canImport: boolean;
} => {
  const parsed = parseStudioLayout(layoutJson);
  if (!parsed.canImport || !parsed.document) {
    return {
      state,
      issues: parsed.issues,
      canImport: false,
    };
  }

  const applied = applyStudioLayoutDocument(state, parsed.document);
  return {
    state: applied.state,
    issues: [...parsed.issues, ...applied.issues],
    canImport: true,
  };
};
