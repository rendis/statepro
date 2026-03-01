import {
  BookOpen,
  Code2,
  Copy,
  Info,
  Maximize,
  Play,
  PlusCircle,
  Settings2,
  StickyNote,
  Trash2,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useReducer, useRef, useState } from "react";
import {
  TransformComponent,
  TransformWrapper,
  type ReactZoomPanPinchContentRef,
} from "react-zoom-pan-pinch";

import {
  REALITY_TYPE_LABEL_KEYS,
  REALITY_TYPES,
  STUDIO_ICON_REGISTRY,
  STUDIO_ICONS,
} from "./constants";
import {
  AnchoredNote,
  CanvasToolbar,
  GlobalNoteNode,
  RealityNode,
  TransitionBadge,
  TransitionRoute,
} from "./components/canvas";
import {
  applyStudioLayoutDocument,
  buildEditorStateFromExternalValue,
  createInitialEditorState,
  deserializeStatePro,
  serializeStatePro,
  serializeStudioLayout,
} from "./model";
import {
  createHistorySnapshot,
  createInitialEditorHistoryState,
  editorHistoryReducer,
} from "./state";
import type {
  EditorAction,
  EditorHistorySnapshot,
  HistoryApplyMode,
} from "./state";
import type {
  BehaviorModalState,
  ConnectingStart,
  DragStartInfo,
  EditorNode,
  EditorState,
  MachineConfig,
  BehaviorRegistryItem,
  NodeSizeMap,
  ResizingStart,
  EditorTransition,
  TransitionLeg,
  StateProMachine,
  AnchoredNoteData,
  GlobalNoteData,
  StudioChangePayload,
  StudioExternalValue,
  StudioFeatureFlags,
  StudioLayoutDocument,
  StudioLayoutIssue,
  StudioUniverseTemplate,
} from "./types";
import {
  BADGE_WIDTH,
  DEFAULT_CANVAS_SEARCH_FILTERS,
  buildIssueIndex,
  buildTargetReferenceFromNodes,
  buildTransitionLegs,
  cleanIdentifier,
  collectBehaviorUsages,
  ensureUniqueIdentifier,
  getTransitionRouteGeometry,
  getTransitionGroupKey,
  isInvalidNotifyTransition,
  moveTransitionInsideGroup,
  searchCanvasNodes,
  viewportPointToCanvasPoint,
  removeTransitionsReferencingDeletedNodes,
  replacePackIdInBindings,
  removePackBindingsByPackId,
  removeBehaviorReferences,
  renameRealityId,
  renameUniverseId,
} from "./utils";
import type { CanvasSearchFilters } from "./utils";
import { computeAutoLayout } from "./utils/autoLayout";
import {
  computeCanvasSize,
  computeFitZoom,
  getContentBounds,
} from "./utils/viewport";
import {
  BehaviorModal,
  JsonIOModal,
  LibraryModal,
  PropertiesModal,
} from "./features/modals";
import { MachineGlobalPanel } from "./features/machine";
import {
  resolveInitialStudioLocale,
  resolveSerializeIssueMessage,
  StudioI18nProvider,
  useI18n,
  type StudioLocale,
} from "./i18n";

type SelectedEditorElement = EditorNode | { type: "transition"; id: string; data: EditorTransition };
type InspectableEditorNode = Extract<EditorNode, { type: "universe" | "reality" }>;
type TransitionDragStart = {
  transitionId: string;
  mouseX: number;
  mouseY: number;
  startOffsetX: number;
  startOffsetY: number;
};

const DEFAULT_REALITY_WIDTH = 192;
const DEFAULT_REALITY_HEIGHT = 150;
const DEFAULT_GLOBAL_NOTE_WIDTH = 224;
const COLLISION_MARGIN = 20;
const MIN_ZOOM = 0.2;
const MAX_ZOOM = 2;
const ZOOM_STEP = 0.1;
const FIT_PADDING = 80;
const MIN_CANVAS_SIZE = 3200;
const CANVAS_EDGE_PADDING = 600;
const VIEWPORT_ANIMATION_MS = 180;
const VIEWPORT_ANIMATION_TYPE = "easeOutCubic" as const;
const DEFAULT_CHANGE_DEBOUNCE_MS = 250;
const SEARCH_PULSE_DURATION_MS = 900;

type HistoryTrackingOptions = {
  mode?: HistoryApplyMode;
  group?: string;
  markDirtyFromImport?: boolean;
};

const isInspectableEditorNode = (
  element: SelectedEditorElement | null,
): element is InspectableEditorNode => {
  return Boolean(element && element.type !== "transition" && element.type !== "note");
};

export interface StateProEditorProps {
  value?: StudioExternalValue;
  defaultValue?: StudioExternalValue;
  onChange?: (payload: StudioChangePayload) => void;
  changeDebounceMs?: number;
  universeTemplates?: StudioUniverseTemplate[];
  libraryBehaviors?: BehaviorRegistryItem[];
  features?: StudioFeatureFlags;
  locale?: StudioLocale;
  defaultLocale?: StudioLocale;
  onLocaleChange?: (locale: StudioLocale) => void;
  persistLocale?: boolean;
  showLocaleSwitcher?: boolean;
}

interface StateProEditorInnerProps {
  initialLocale: StudioLocale;
  showLocaleSwitcher: boolean;
  value?: StudioExternalValue;
  defaultValue?: StudioExternalValue;
  onChange?: (payload: StudioChangePayload) => void;
  changeDebounceMs: number;
  universeTemplates: StudioUniverseTemplate[];
  libraryBehaviors?: BehaviorRegistryItem[];
  features: {
    json: {
      import: boolean;
      export: boolean;
    };
    library: {
      behaviors: {
        manage: boolean;
      };
      metadataPacks: {
        create: boolean;
      };
    };
  };
}

export function StateProEditor({
  value,
  defaultValue,
  onChange,
  changeDebounceMs = DEFAULT_CHANGE_DEBOUNCE_MS,
  universeTemplates = [],
  libraryBehaviors,
  features,
  locale,
  defaultLocale = "en",
  onLocaleChange,
  persistLocale = true,
  showLocaleSwitcher = true,
}: StateProEditorProps = {}) {
  const initialLocale = resolveInitialStudioLocale({
    locale,
    defaultLocale,
    persistLocale,
  });

  const resolvedFeatures = {
    json: {
      import: features?.json?.import ?? true,
      export: features?.json?.export ?? true,
    },
    library: {
      behaviors: {
        manage: features?.library?.behaviors?.manage ?? true,
      },
      metadataPacks: {
        create: features?.library?.metadataPacks?.create ?? true,
      },
    },
  };

  return (
    <StudioI18nProvider
      locale={locale}
      defaultLocale={defaultLocale}
      onLocaleChange={onLocaleChange}
      persistLocale={persistLocale}
    >
      <StateProEditorInner
        initialLocale={initialLocale}
        showLocaleSwitcher={showLocaleSwitcher}
        value={value}
        defaultValue={defaultValue}
        onChange={onChange}
        changeDebounceMs={changeDebounceMs}
        universeTemplates={universeTemplates}
        libraryBehaviors={libraryBehaviors}
        features={resolvedFeatures}
      />
    </StudioI18nProvider>
  );
}

function StateProEditorInner({
  initialLocale,
  showLocaleSwitcher,
  value,
  defaultValue,
  onChange,
  changeDebounceMs,
  universeTemplates,
  libraryBehaviors,
  features,
}: StateProEditorInnerProps) {
  const { locale, setLocale, t } = useI18n();
  const containerRef = useRef<HTMLDivElement | null>(null);
  const canvasRef = useRef<HTMLDivElement | null>(null);
  const isControlled = value !== undefined;
  const changeSourceRef = useRef<"user" | "external-sync">("user");

  const buildInitialState = useCallback((): EditorState => {
    if (value) {
      return buildEditorStateFromExternalValue(value, { libraryBehaviors });
    }

    if (defaultValue) {
      return buildEditorStateFromExternalValue(defaultValue, { libraryBehaviors });
    }

    const initialState = createInitialEditorState(initialLocale);
    if (libraryBehaviors) {
      return {
        ...initialState,
        registry: structuredClone(libraryBehaviors),
      };
    }

    return initialState;
  }, [defaultValue, initialLocale, libraryBehaviors, value]);

  const [historyState, dispatchHistory] = useReducer(
    editorHistoryReducer,
    undefined,
    () => createInitialEditorHistoryState(buildInitialState()),
  );
  const editorState = historyState.present;
  const {
    nodes,
    transitions,
    nodeSizes,
    registry,
    machineConfig,
    metadataPackRegistry,
    metadataPackBindings,
    lastImportedMachine,
  } = editorState;
  const latestEditorStateRef = useRef(editorState);
  latestEditorStateRef.current = editorState;
  const autoLayoutRunIdRef = useRef(0);
  const isMountedRef = useRef(true);
  const canUndo = historyState.past.length > 0;
  const canRedo = historyState.future.length > 0;
  const lastControlledValueSignatureRef = useRef<string | null>(null);
  const lastLibraryBehaviorsSignatureRef = useRef<string | null>(null);

  const [selectedElement, setSelectedElement] = useState<SelectedEditorElement | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchFilters, setSearchFilters] = useState<CanvasSearchFilters>({
    ...DEFAULT_CANVAS_SEARCH_FILTERS,
  });
  const [searchActiveIndex, setSearchActiveIndex] = useState(-1);
  const [searchPulseNodeId, setSearchPulseNodeId] = useState<string | null>(null);
  const [searchPulseTick, setSearchPulseTick] = useState(0);

  const applyEditorAction = useCallback(
    (
      action: EditorAction,
      mode: HistoryApplyMode = "record",
      group?: string,
      markDirtyFromImport = true,
    ) => {
      changeSourceRef.current = "user";
      dispatchHistory({
        type: "apply-editor-action",
        action,
        mode,
        group,
        markDirtyFromImport,
        now: Date.now(),
      });
    },
    [],
  );

  const applyRecord = useCallback(
    (action: EditorAction, markDirtyFromImport = true) =>
      applyEditorAction(action, "record", undefined, markDirtyFromImport),
    [applyEditorAction],
  );

  const applyCoalesced = useCallback(
    (action: EditorAction, group: string, markDirtyFromImport = true) =>
      applyEditorAction(action, "coalesce", group, markDirtyFromImport),
    [applyEditorAction],
  );

  const applySilent = useCallback(
    (action: EditorAction, markDirtyFromImport = true) =>
      applyEditorAction(action, "silent", undefined, markDirtyFromImport),
    [applyEditorAction],
  );

  const undo = useCallback(() => {
    dispatchHistory({ type: "undo" });
  }, []);

  const redo = useCallback(() => {
    dispatchHistory({ type: "redo" });
  }, []);

  const resetHistoryWith = useCallback((state: EditorState) => {
    dispatchHistory({ type: "reset-history", payload: state });
  }, []);
  const resetHistoryWithRef = useRef(resetHistoryWith);
  resetHistoryWithRef.current = resetHistoryWith;

  const markDirtyFromImport = () => {};

  const setNodes = (
    value: EditorNode[] | ((prev: EditorNode[]) => EditorNode[]),
    options: HistoryTrackingOptions = {},
  ) => {
    const mode = options.mode || "record";
    const group = options.group || "nodes";
    const markDirty = options.markDirtyFromImport ?? true;
    const action: EditorAction =
      typeof value === "function"
        ? { type: "update-nodes", payload: value }
        : { type: "set-nodes", payload: value };

    if (mode === "coalesce") {
      applyCoalesced(action, group, markDirty);
      return;
    }
    if (mode === "silent") {
      applySilent(action, markDirty);
      return;
    }
    applyRecord(action, markDirty);
  };

  const setTransitions = (
    value:
      | EditorTransition[]
      | ((prev: EditorTransition[]) => EditorTransition[]),
    options: HistoryTrackingOptions = {},
  ) => {
    const mode = options.mode || "record";
    const group = options.group || "transitions";
    const markDirty = options.markDirtyFromImport ?? true;
    const action: EditorAction =
      typeof value === "function"
        ? { type: "update-transitions", payload: value }
        : { type: "set-transitions", payload: value };

    if (mode === "coalesce") {
      applyCoalesced(action, group, markDirty);
      return;
    }
    if (mode === "silent") {
      applySilent(action, markDirty);
      return;
    }
    applyRecord(action, markDirty);
  };

  const setNodeSizes = (value: NodeSizeMap | ((prev: NodeSizeMap) => NodeSizeMap)) => {
    if (typeof value === "function") {
      applySilent({ type: "update-node-sizes", payload: value }, false);
      return;
    }
    applySilent({ type: "set-node-sizes", payload: value }, false);
  };

  const setRegistry = (
    value:
      | EditorState["registry"]
      | ((prev: EditorState["registry"]) => EditorState["registry"]),
    options: HistoryTrackingOptions = {},
  ) => {
    const mode = options.mode || "record";
    const group = options.group || "registry";
    const markDirty = options.markDirtyFromImport ?? true;
    const action: EditorAction =
      typeof value === "function"
        ? { type: "update-registry", payload: value }
        : { type: "set-registry", payload: value };

    if (mode === "coalesce") {
      applyCoalesced(action, group, markDirty);
      return;
    }
    if (mode === "silent") {
      applySilent(action, markDirty);
      return;
    }
    applyRecord(action, markDirty);
  };

  const setMetadataPackRegistry = (
    value:
      | EditorState["metadataPackRegistry"]
      | ((prev: EditorState["metadataPackRegistry"]) => EditorState["metadataPackRegistry"]),
    options: HistoryTrackingOptions = {},
  ) => {
    const mode = options.mode || "record";
    const group = options.group || "metadata-pack-registry";
    const markDirty = options.markDirtyFromImport ?? true;
    const action: EditorAction =
      typeof value === "function"
        ? { type: "update-metadata-pack-registry", payload: value }
        : { type: "set-metadata-pack-registry", payload: value };

    if (mode === "coalesce") {
      applyCoalesced(action, group, markDirty);
      return;
    }
    if (mode === "silent") {
      applySilent(action, markDirty);
      return;
    }
    applyRecord(action, markDirty);
  };

  const setMetadataPackBindings = (
    value:
      | EditorState["metadataPackBindings"]
      | ((prev: EditorState["metadataPackBindings"]) => EditorState["metadataPackBindings"]),
    options: HistoryTrackingOptions = {},
  ) => {
    const mode = options.mode || "record";
    const group = options.group || "metadata-pack-bindings";
    const markDirty = options.markDirtyFromImport ?? true;
    const action: EditorAction =
      typeof value === "function"
        ? { type: "update-metadata-pack-bindings", payload: value }
        : { type: "set-metadata-pack-bindings", payload: value };

    if (mode === "coalesce") {
      applyCoalesced(action, group, markDirty);
      return;
    }
    if (mode === "silent") {
      applySilent(action, markDirty);
      return;
    }
    applyRecord(action, markDirty);
  };

  const setMachineConfig = (
    value: MachineConfig | ((prev: MachineConfig) => MachineConfig),
    options: HistoryTrackingOptions = {},
  ) => {
    const mode = options.mode || "coalesce";
    const group = options.group || "machine-config";
    const markDirty = options.markDirtyFromImport ?? true;
    const action: EditorAction =
      typeof value === "function"
        ? { type: "update-machine-config-fn", payload: value }
        : { type: "set-machine-config", payload: value };

    if (mode === "coalesce") {
      applyCoalesced(action, group, markDirty);
      return;
    }
    if (mode === "silent") {
      applySilent(action, markDirty);
      return;
    }
    applyRecord(action, markDirty);
  };

  const [showJsonModal, setShowJsonModal] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isLibraryOpen, setIsLibraryOpen] = useState(false);
  const [hoveredTransitionId, setHoveredTransitionId] = useState<string | null>(null);

  useEffect(() => {
    if ((features.json.import || features.json.export) || !showJsonModal) {
      return;
    }
    setShowJsonModal(false);
  }, [features.json.export, features.json.import, showJsonModal]);

  const [behaviorModal, setBehaviorModal] = useState<BehaviorModalState>({
    isOpen: false,
    type: "action",
    initialData: null,
    onSave: null,
  });

  const [draggingNode, setDraggingNode] = useState<string | null>(null);
  const [resizingStart, setResizingStart] = useState<ResizingStart | null>(null);
  const [dragStartInfo, setDragStartInfo] = useState<DragStartInfo | null>(null);
  const [draggingTransition, setDraggingTransition] = useState<TransitionDragStart | null>(null);
  const [connectingStart, setConnectingStart] = useState<ConnectingStart | null>(null);
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [isCanvasPanning, setIsCanvasPanning] = useState(false);
  const [isAutoLayouting, setIsAutoLayouting] = useState(false);
  const transformRef = useRef<ReactZoomPanPinchContentRef | null>(null);
  const didInitialAutoLayoutRef = useRef(false);
  const didInitialAutoFitRef = useRef(false);
  const lastAutoFitImportRef = useRef<StateProMachine | null>(null);
  const gestureBaseSnapshotRef = useRef<EditorHistorySnapshot | null>(null);
  const panStartedRef = useRef(false);
  const panMovedRef = useRef(false);

  useEffect(() => {
    if (!isControlled || !value) {
      return;
    }

    const nextSignature = JSON.stringify({
      value,
      libraryBehaviors: libraryBehaviors || null,
    });

    if (lastControlledValueSignatureRef.current === null) {
      lastControlledValueSignatureRef.current = nextSignature;
      return;
    }

    if (lastControlledValueSignatureRef.current === nextSignature) {
      return;
    }

    lastControlledValueSignatureRef.current = nextSignature;
    const nextState = buildEditorStateFromExternalValue(value, {
      libraryBehaviors,
    });
    changeSourceRef.current = "external-sync";
    resetHistoryWith(nextState);
    gestureBaseSnapshotRef.current = null;
    setSelectedElement(null);
    setIsModalOpen(false);
  }, [isControlled, libraryBehaviors, resetHistoryWith, value]);

  useEffect(() => {
    if (isControlled || !libraryBehaviors) {
      return;
    }

    const nextSignature = JSON.stringify(libraryBehaviors);
    if (lastLibraryBehaviorsSignatureRef.current === null) {
      lastLibraryBehaviorsSignatureRef.current = nextSignature;
      if (JSON.stringify(registry) === nextSignature) {
        return;
      }
    }

    if (lastLibraryBehaviorsSignatureRef.current === nextSignature) {
      return;
    }

    lastLibraryBehaviorsSignatureRef.current = nextSignature;
    changeSourceRef.current = "external-sync";
    setRegistry(structuredClone(libraryBehaviors), {
      mode: "silent",
      markDirtyFromImport: false,
    });
  }, [isControlled, libraryBehaviors, registry]);

  const captureGestureBaseSnapshot = useCallback(() => {
    gestureBaseSnapshotRef.current = createHistorySnapshot(editorState);
  }, [editorState]);

  const commitGestureHistoryStep = useCallback(() => {
    const baseSnapshot = gestureBaseSnapshotRef.current;
    gestureBaseSnapshotRef.current = null;
    if (!baseSnapshot) {
      return;
    }

    dispatchHistory({
      type: "commit-snapshot",
      payload: baseSnapshot,
    });
  }, []);
  const contentBounds = useMemo(() => getContentBounds(nodes, nodeSizes), [nodeSizes, nodes]);
  const canvasSize = useMemo(
    () => computeCanvasSize(contentBounds, MIN_CANVAS_SIZE, CANVAS_EDGE_PADDING),
    [contentBounds],
  );
  const searchResults = useMemo(
    () =>
      searchCanvasNodes(nodes, searchQuery, {
        filters: searchFilters,
      }),
    [nodes, searchFilters, searchQuery],
  );

  const getViewportSize = () => {
    const container = containerRef.current;
    if (!container) {
      return null;
    }
    return {
      width: container.clientWidth,
      height: container.clientHeight,
    };
  };

  const getTransformState = () => {
    const transformState = transformRef.current?.instance.transformState;
    if (!transformState) {
      return {
        scale: 1,
        positionX: 0,
        positionY: 0,
      };
    }
    return {
      scale: transformState.scale,
      positionX: transformState.positionX,
      positionY: transformState.positionY,
    };
  };

  const centerAt = (
    centerX: number,
    centerY: number,
    targetZoom: number,
    behavior: ScrollBehavior,
  ) => {
    const transformApi = transformRef.current;
    const viewport = getViewportSize();
    if (!transformApi || !viewport) {
      return;
    }

    const nextPositionX = viewport.width / 2 - centerX * targetZoom;
    const nextPositionY = viewport.height / 2 - centerY * targetZoom;
    const animationTime = behavior === "smooth" ? VIEWPORT_ANIMATION_MS : 0;
    transformApi.setTransform(
      nextPositionX,
      nextPositionY,
      targetZoom,
      animationTime,
      VIEWPORT_ANIMATION_TYPE,
    );
  };

  const focusNodeInCanvas = useCallback(
    (nodeId: string, behavior: ScrollBehavior = "smooth") => {
      const node = nodes.find(
        (candidate): candidate is Extract<EditorNode, { type: "universe" | "reality" }> =>
          (candidate.type === "universe" || candidate.type === "reality") && candidate.id === nodeId,
      );
      if (!node) {
        return;
      }

      if (node.type === "universe") {
        centerAt(node.x + node.w / 2, node.y + node.h / 2, getTransformState().scale, behavior);
      } else {
        const width = nodeSizes[node.id]?.w || DEFAULT_REALITY_WIDTH;
        const height = nodeSizes[node.id]?.h || DEFAULT_REALITY_HEIGHT;
        centerAt(node.x + width / 2, node.y + height / 2, getTransformState().scale, behavior);
      }

      setSelectedElement(node);
      setSearchPulseNodeId(node.id);
      setSearchPulseTick((previous) => previous + 1);
    },
    [nodeSizes, nodes],
  );

  const moveSearchSelection = useCallback(
    (direction: "up" | "down") => {
      if (searchResults.length === 0) {
        return;
      }

      let nextIndex = 0;
      if (searchActiveIndex < 0) {
        nextIndex = direction === "down" ? 0 : searchResults.length - 1;
      } else {
        nextIndex =
          direction === "down"
            ? (searchActiveIndex + 1) % searchResults.length
            : (searchActiveIndex - 1 + searchResults.length) % searchResults.length;
      }

      setSearchActiveIndex(nextIndex);
      const nextResult = searchResults[nextIndex];
      if (nextResult) {
        focusNodeInCanvas(nextResult.nodeId, "smooth");
      }
    },
    [focusNodeInCanvas, searchActiveIndex, searchResults],
  );

  const selectSearchResult = useCallback(
    (index: number) => {
      if (index < 0 || index >= searchResults.length) {
        return;
      }
      setSearchActiveIndex(index);
      focusNodeInCanvas(searchResults[index].nodeId, "smooth");
    },
    [focusNodeInCanvas, searchResults],
  );

  const handleCenterContent = useCallback(
    (behavior: ScrollBehavior = "smooth") => {
      const centerX = contentBounds ? contentBounds.centerX : canvasSize.width / 2;
      const centerY = contentBounds ? contentBounds.centerY : canvasSize.height / 2;
      const currentScale = getTransformState().scale;
      centerAt(centerX, centerY, currentScale, behavior);
    },
    [canvasSize.height, canvasSize.width, contentBounds],
  );

  const handleFitToContent = useCallback(
    (behavior: ScrollBehavior = "smooth") => {
      const viewport = getViewportSize();
      if (!viewport) {
        return;
      }

      if (!contentBounds) {
        const fallbackZoom = 1;
        centerAt(canvasSize.width / 2, canvasSize.height / 2, fallbackZoom, behavior);
        return;
      }

      const nextZoom = computeFitZoom(
        contentBounds,
        viewport.width,
        viewport.height,
        FIT_PADDING,
        MIN_ZOOM,
        MAX_ZOOM,
      );
      centerAt(contentBounds.centerX, contentBounds.centerY, nextZoom, behavior);
    },
    [canvasSize.height, canvasSize.width, contentBounds],
  );
  const handleFitToContentRef = useRef(handleFitToContent);
  handleFitToContentRef.current = handleFitToContent;

  const beginAutoLayoutRun = useCallback((): number => {
    const runId = autoLayoutRunIdRef.current + 1;
    autoLayoutRunIdRef.current = runId;
    setIsAutoLayouting(true);
    return runId;
  }, []);

  const endAutoLayoutRun = useCallback((runId: number) => {
    if (!isMountedRef.current) {
      return;
    }

    if (autoLayoutRunIdRef.current === runId) {
      setIsAutoLayouting(false);
    }
  }, []);

  const handleZoomIn = useCallback(() => {
    const transformApi = transformRef.current;
    if (!transformApi) {
      return;
    }
    transformApi.zoomIn(ZOOM_STEP, VIEWPORT_ANIMATION_MS, VIEWPORT_ANIMATION_TYPE);
  }, []);

  const handleZoomOut = useCallback(() => {
    const transformApi = transformRef.current;
    if (!transformApi) {
      return;
    }
    transformApi.zoomOut(ZOOM_STEP, VIEWPORT_ANIMATION_MS, VIEWPORT_ANIMATION_TYPE);
  }, []);

  const scheduleAutoFit = useCallback(
    (behavior: ScrollBehavior = "auto") => {
      let attempt = 0;
      const maxAttempts = 8;

      const run = () => {
        const container = containerRef.current;
        if (!container) {
          return;
        }

        // Wait until viewport and transform instance are initialized before fitting.
        if (
          (container.clientWidth <= 0 ||
            container.clientHeight <= 0 ||
            !transformRef.current) &&
          attempt < maxAttempts
        ) {
          attempt += 1;
          requestAnimationFrame(run);
          return;
        }

        handleFitToContent(behavior);
      };

      requestAnimationFrame(run);
    },
    [handleFitToContent],
  );

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    if (didInitialAutoLayoutRef.current) {
      return;
    }

    didInitialAutoLayoutRef.current = true;
    const baseState = latestEditorStateRef.current;
    const runId = beginAutoLayoutRun();

    const applyInitialAutoLayout = async () => {
      try {
        const nextNodes = await computeAutoLayout(
          baseState.nodes,
          baseState.transitions,
          baseState.nodeSizes,
        );
        if (!isMountedRef.current || autoLayoutRunIdRef.current !== runId) {
          return;
        }

        const latestState = latestEditorStateRef.current;
        const canApply =
          latestState.nodes === baseState.nodes &&
          latestState.transitions === baseState.transitions &&
          latestState.lastImportedMachine === baseState.lastImportedMachine;
        if (!canApply) {
          return;
        }

        resetHistoryWithRef.current({
          ...latestState,
          nodes: nextNodes,
        });

        requestAnimationFrame(() => {
          handleFitToContentRef.current("auto");
        });
      } catch (error) {
        console.error("Failed to compute initial auto-layout", error);
      } finally {
        endAutoLayoutRun(runId);
      }
    };

    void applyInitialAutoLayout();
  }, [beginAutoLayoutRun, endAutoLayoutRun]);

  useEffect(() => {
    if (didInitialAutoFitRef.current) {
      return;
    }
    didInitialAutoFitRef.current = true;
    scheduleAutoFit("auto");
  }, [scheduleAutoFit]);

  useEffect(() => {
    if (!lastImportedMachine) {
      return;
    }

    if (lastAutoFitImportRef.current === lastImportedMachine) {
      return;
    }

    lastAutoFitImportRef.current = lastImportedMachine;
    scheduleAutoFit("auto");
  }, [lastImportedMachine, scheduleAutoFit]);

  useEffect(() => {
    if (!searchQuery.trim() || searchResults.length === 0) {
      if (searchActiveIndex !== -1) {
        setSearchActiveIndex(-1);
      }
      return;
    }

    if (searchActiveIndex >= searchResults.length || searchActiveIndex < 0) {
      setSearchActiveIndex(0);
    }
  }, [searchActiveIndex, searchQuery, searchResults]);

  useEffect(() => {
    if (!searchPulseNodeId) {
      return;
    }

    const timer = window.setTimeout(() => {
      setSearchPulseNodeId(null);
    }, SEARCH_PULSE_DURATION_MS);

    return () => {
      window.clearTimeout(timer);
    };
  }, [searchPulseNodeId, searchPulseTick]);

  useEffect(() => {
    const isEditableTarget = (target: EventTarget | null): boolean => {
      if (!(target instanceof HTMLElement)) {
        return false;
      }

      if (target.isContentEditable) {
        return true;
      }

      const tag = target.tagName.toLowerCase();
      return tag === "input" || tag === "textarea" || tag === "select";
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (!(event.metaKey || event.ctrlKey) || isEditableTarget(event.target)) {
        return;
      }

      const key = event.key.toLowerCase();

      if (key === "z") {
        if (event.shiftKey) {
          if (!canRedo) {
            return;
          }
          event.preventDefault();
          redo();
          return;
        }

        if (!canUndo) {
          return;
        }
        event.preventDefault();
        undo();
        return;
      }

      if (key === "y" && event.ctrlKey) {
        if (!canRedo) {
          return;
        }
        event.preventDefault();
        redo();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [canRedo, canUndo, redo, undo]);

  const transitionLegs = useMemo(() => buildTransitionLegs(transitions, nodes), [transitions, nodes]);

  const transitionsById = useMemo(
    () => new Map(transitions.map((transition) => [transition.id, transition])),
    [transitions],
  );

  const transitionOrderSummaryById = useMemo(() => {
    const groupedTransitions = new Map<string, EditorTransition[]>();
    transitions.forEach((transition) => {
      const groupKey = getTransitionGroupKey(transition);
      const currentGroup = groupedTransitions.get(groupKey) || [];
      currentGroup.push(transition);
      groupedTransitions.set(groupKey, currentGroup);
    });

    const orderSummaryById = new Map<string, { position: number; total: number }>();
    groupedTransitions.forEach((groupTransitions) => {
      const sortedByOrder = [...groupTransitions].sort((a, b) => a.order - b.order);
      const total = sortedByOrder.length || 1;
      sortedByOrder.forEach((transition, index) => {
        orderSummaryById.set(transition.id, {
          position: index + 1,
          total,
        });
      });
    });

    return orderSummaryById;
  }, [transitions]);

  const transitionLegsByTransitionId = useMemo(() => {
    const grouped = new Map<string, TransitionLeg[]>();
    transitionLegs.forEach((leg) => {
      const current = grouped.get(leg.transitionId) || [];
      current.push(leg);
      grouped.set(leg.transitionId, current);
    });
    return grouped;
  }, [transitionLegs]);

  const transitionRouteGeometryByTransitionId = useMemo(() => {
    const geometryByTransitionId = new Map<string, ReturnType<typeof getTransitionRouteGeometry>>();

    transitions.forEach((transition) => {
      const legs = transitionLegsByTransitionId.get(transition.id) || [];
      geometryByTransitionId.set(
        transition.id,
        getTransitionRouteGeometry(transition, legs, nodes, nodeSizes),
      );
    });

    return geometryByTransitionId;
  }, [nodeSizes, nodes, transitionLegsByTransitionId, transitions]);

  const transitionBadgeAnchors = useMemo(() => {
    const anchors = new Map<string, { x: number; y: number }>();

    transitions.forEach((transition) => {
      const routeGeometry = transitionRouteGeometryByTransitionId.get(transition.id);
      if (routeGeometry) {
        anchors.set(transition.id, routeGeometry.hubCenter ?? routeGeometry.anchor);
        return;
      }

      const sourceNode = nodes.find(
        (node): node is Extract<EditorNode, { type: "reality" }> =>
          node.type === "reality" && node.id === transition.sourceRealityId,
      );
      if (!sourceNode) {
        return;
      }
      const sourceSize = nodeSizes[sourceNode.id] || { w: 192, h: 80 };
      const transitionOffset = transition.visualOffset || { x: 0, y: 0 };
      anchors.set(transition.id, {
        x: sourceNode.x + sourceSize.w + 80 + transitionOffset.x,
        y: sourceNode.y + sourceSize.h / 2 + transitionOffset.y,
      });
    });

    return anchors;
  }, [nodeSizes, nodes, transitionRouteGeometryByTransitionId, transitions]);

  const resolveRegistryBehaviorUsage = (src: string) =>
    collectBehaviorUsages(src, {
      machineConfig,
      nodes,
      transitions,
    });

  const handleDeleteRegistryBehavior = (src: string) => {
    if (!features.library.behaviors.manage) {
      return;
    }

    markDirtyFromImport();
    const { nextNodes, nextTransitions, nextMachineConfig } = removeBehaviorReferences(src, {
      machineConfig,
      nodes,
      transitions,
    });
    const historyGroup = `delete-registry-behavior:${src}`;

    setNodes(nextNodes, { mode: "coalesce", group: historyGroup });
    setTransitions(nextTransitions, { mode: "coalesce", group: historyGroup });
    setMachineConfig(nextMachineConfig, { mode: "coalesce", group: historyGroup });
    setRegistry((prev) => prev.filter((entry) => entry.src !== src), {
      mode: "coalesce",
      group: historyGroup,
    });

    let nextSelectedElement: SelectedEditorElement | null = null;
    if (selectedElement) {
      if (selectedElement.type === "transition") {
        const transition = nextTransitions.find((candidate) => candidate.id === selectedElement.id);
        nextSelectedElement = transition
          ? {
              type: "transition",
              id: transition.id,
              data: transition,
            }
          : null;
      } else {
        nextSelectedElement = nextNodes.find((candidate) => candidate.id === selectedElement.id) || null;
      }
    }

    setSelectedElement(nextSelectedElement);
    if (!nextSelectedElement) {
      setIsModalOpen(false);
    }
  };

  const handleDeleteMetadataPack = (packId: string) => {
    markDirtyFromImport();
    setMetadataPackBindings((previous) => removePackBindingsByPackId(previous, packId));
  };

  const handleRenameMetadataPack = (previousPackId: string, nextPackId: string) => {
    if (!previousPackId || !nextPackId || previousPackId === nextPackId) {
      return;
    }
    markDirtyFromImport();
    setMetadataPackBindings((previous) =>
      replacePackIdInBindings(previous, previousPackId, nextPackId),
    );
  };

  const getCanvasCoords = (e: React.MouseEvent<Element>) => {
    const viewportRect = containerRef.current?.getBoundingClientRect();
    const transformState = transformRef.current?.instance.transformState;
    const canvasPoint = viewportPointToCanvasPoint(
      { x: e.clientX, y: e.clientY },
      viewportRect,
      transformState
        ? {
            scale: transformState.scale,
            positionX: transformState.positionX,
            positionY: transformState.positionY,
          }
        : null,
    );

    return canvasPoint || { x: 0, y: 0 };
  };

  const isCanvasBackgroundTarget = (target: EventTarget | null): target is Element => {
    if (!(target instanceof Element)) {
      return false;
    }

    return target === canvasRef.current || target.tagName.toLowerCase() === "svg";
  };

  const resolveUniversePlacement = (width: number, height: number) => {
    const viewport = getViewportSize();
    const transform = getTransformState();
    const centerX = viewport
      ? (viewport.width / 2 - transform.positionX) / transform.scale - width / 2
      : 1000;
    const centerY = viewport
      ? (viewport.height / 2 - transform.positionY) / transform.scale - height / 2
      : 1000;

    let newX = centerX;
    let newY = centerY;
    let overlaps = true;

    while (overlaps) {
      overlaps = false;
      nodes.forEach((node) => {
        if (node.type !== "universe") {
          return;
        }

        const overlapsV = !(newY + height + COLLISION_MARGIN <= node.y || newY >= node.y + node.h + COLLISION_MARGIN);
        const overlapsH = !(newX + width + COLLISION_MARGIN <= node.x || newX >= node.x + node.w + COLLISION_MARGIN);

        if (overlapsV && overlapsH) {
          overlaps = true;
          newY = node.y + node.h + 30;
        }
      });
    }

    return {
      x: newX,
      y: newY,
    };
  };

  const addUniverse = () => {
    markDirtyFromImport();
    const newId = `universe-${Date.now()}`;
    const placement = resolveUniversePlacement(400, 300);

    const universeCount = nodes.filter((n) => n.type === "universe").length + 1;
    const formattedName = cleanIdentifier(`universe-${universeCount}`);
    const newUniverse: Extract<EditorNode, { type: "universe" }> = {
      id: newId,
      type: "universe",
      x: placement.x,
      y: placement.y,
      w: 400,
      h: 300,
      data: {
        id: formattedName,
        name: formattedName,
        canonicalName: formattedName,
        version: "1.0.0",
        description: "",
        tags: [],
        metadata: "{}",
        universalConstants: {
          entryActions: [],
          exitActions: [],
          entryInvokes: [],
          exitInvokes: [],
          actionsOnTransition: [],
          invokesOnTransition: [],
        },
      },
    };

    setNodes((prev) => [...prev, newUniverse]);
    setSelectedElement(newUniverse);
  };

  const addUniverseFromTemplate = (templateId: string) => {
    const template = universeTemplates.find((entry) => entry.id === templateId);
    if (!template) {
      return;
    }

    const templateMachine: StateProMachine = {
      id: `template-machine-${Date.now()}`,
      canonicalName: `template-machine-${Date.now()}`,
      version: "1.0.0",
      initials: [`U:${template.universe.id}`],
      universes: {
        [template.universe.id]: structuredClone(template.universe),
      },
    };

    const templateState = deserializeStatePro(templateMachine);
    const templateUniverse = templateState.nodes.find(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );

    if (!templateUniverse) {
      return;
    }

    markDirtyFromImport();

    const usedUniverseIdentifiers = new Set(
      nodes
        .filter(
          (node): node is Extract<EditorNode, { type: "universe" }> =>
            node.type === "universe",
        )
        .flatMap((node) => [node.data.id, node.data.canonicalName || node.data.id]),
    );
    const universeBaseId =
      cleanIdentifier(templateUniverse.data.id || template.id || "template-universe") ||
      "template-universe";
    const nextUniverseDataId = ensureUniqueIdentifier(
      universeBaseId,
      usedUniverseIdentifiers,
      "template-universe",
    );
    const nextUniverseNodeId = `universe-${Date.now()}-${Math.random()
      .toString(36)
      .slice(2, 7)}`;
    const placement = resolveUniversePlacement(templateUniverse.w || 400, templateUniverse.h || 300);
    const offsetX = placement.x - templateUniverse.x;
    const offsetY = placement.y - templateUniverse.y;

    const realityNodes = templateState.nodes.filter(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.type === "reality" && node.data.universeId === templateUniverse.id,
    );

    const realityNodeIdMap = new Map<string, string>();
    const clonedRealities: Extract<EditorNode, { type: "reality" }>[] = realityNodes.map((node) => {
      const nextRealityNodeId = `reality-${Date.now()}-${Math.random()
        .toString(36)
        .slice(2, 7)}`;
      realityNodeIdMap.set(node.id, nextRealityNodeId);
      return {
        ...node,
        id: nextRealityNodeId,
        x: node.x + offsetX,
        y: node.y + offsetY,
        data: {
          ...node.data,
          universeId: nextUniverseNodeId,
        },
      };
    });

    const sourceUniverseIdentifier = templateUniverse.data.id;
    const clonedTransitions = templateState.transitions
      .filter((transition) => realityNodeIdMap.has(transition.sourceRealityId))
      .map((transition) => ({
        ...transition,
        id: `tr-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
        sourceRealityId: realityNodeIdMap.get(transition.sourceRealityId) || transition.sourceRealityId,
        targets: transition.targets.map((target) => {
          if (!target.startsWith("U:")) {
            return target;
          }

          if (target === `U:${sourceUniverseIdentifier}`) {
            return `U:${nextUniverseDataId}`;
          }

          if (target.startsWith(`U:${sourceUniverseIdentifier}:`)) {
            const suffix = target.replace(`U:${sourceUniverseIdentifier}:`, "");
            return `U:${nextUniverseDataId}:${suffix}`;
          }

          return target;
        }),
      }));

    const nextUniverseNode: Extract<EditorNode, { type: "universe" }> = {
      ...templateUniverse,
      id: nextUniverseNodeId,
      x: placement.x,
      y: placement.y,
      data: {
        ...templateUniverse.data,
        id: nextUniverseDataId,
        name: nextUniverseDataId,
        canonicalName: nextUniverseDataId,
      },
    };

    const historyGroup = `template-universe:${template.id}:${nextUniverseNodeId}`;
    setNodes((previous) => [...previous, nextUniverseNode, ...clonedRealities], {
      mode: "coalesce",
      group: historyGroup,
    });
    setTransitions((previous) => [...previous, ...clonedTransitions], {
      mode: "coalesce",
      group: historyGroup,
    });
    setSelectedElement(nextUniverseNode);
  };

  const addGlobalNote = () => {
    markDirtyFromImport();
    const viewport = getViewportSize();
    const transform = getTransformState();
    const centerX = viewport
      ? (viewport.width / 2 - transform.positionX) / transform.scale - DEFAULT_GLOBAL_NOTE_WIDTH / 2
      : 1000;
    const centerY = viewport
      ? (viewport.height / 2 - transform.positionY) / transform.scale - 100
      : 1000;

    const newNote: Extract<EditorNode, { type: "note" }> = {
      id: `note-${Date.now()}`,
      type: "note",
      x: centerX,
      y: centerY,
      data: {
        text: "",
        colorIndex: 0,
        isCollapsed: false,
      },
    };

    setNodes((prev) => [...prev, newNote], {
      markDirtyFromImport: false,
    });
    setSelectedElement(newNote);
  };

  const addRealityToUniverse = (universeId: string, typeKey: keyof typeof REALITY_TYPES) => {
    markDirtyFromImport();
    const u = nodes.find(
      (n): n is Extract<EditorNode, { type: "universe" }> =>
        n.id === universeId && n.type === "universe",
    );
    if (!u) return;

    const children = nodes.filter(
      (n): n is Extract<EditorNode, { type: "reality" }> =>
        n.type === "reality" && n.data.universeId === universeId,
    );

    let newX = u.x + 20;
    let newY = u.y + 40;
    let overlaps = true;
    let attempts = 0;

    while (overlaps && attempts < 50) {
      overlaps = false;
      for (const child of children) {
        const childW = nodeSizes[child.id]?.w || DEFAULT_REALITY_WIDTH;
        const childH = nodeSizes[child.id]?.h || DEFAULT_REALITY_HEIGHT;
        const overlapsV = !(
          newY + DEFAULT_REALITY_HEIGHT + COLLISION_MARGIN <= child.y ||
          newY >= child.y + childH + COLLISION_MARGIN
        );
        const overlapsH = !(
          newX + DEFAULT_REALITY_WIDTH + COLLISION_MARGIN <= child.x ||
          newX >= child.x + childW + COLLISION_MARGIN
        );

        if (overlapsV && overlapsH) {
          overlaps = true;
          newX += 210;
          if (newX + DEFAULT_REALITY_WIDTH > u.x + Math.max(u.w, 600)) {
            newX = u.x + 20;
            newY += 170;
          }
          break;
        }
      }
      attempts += 1;
    }

    const newId = `reality-${Date.now()}`;
    const realityName = cleanIdentifier(`reality-${nodes.filter((n) => n.type === "reality").length + 1}`);

    const newReality: Extract<EditorNode, { type: "reality" }> = {
      id: newId,
      type: "reality",
      x: newX,
      y: newY,
      data: {
        id: realityName,
        name: realityName,
        universeId,
        isInitial: children.length === 0,
        realityType: typeKey,
        description: "",
        entryActions: [],
        exitActions: [],
        observers: [],
        entryInvokes: [],
        exitInvokes: [],
      },
    };

    setNodes((prev) => {
      const expandedW = Math.max(u.w, newX + DEFAULT_REALITY_WIDTH + 20 - u.x);
      const expandedH = Math.max(u.h, newY + DEFAULT_REALITY_HEIGHT + 20 - u.y);
      return [
        ...prev.map((node) =>
          node.id === u.id && node.type === "universe"
            ? { ...node, w: expandedW, h: expandedH }
            : node,
        ),
        newReality,
      ];
    });

    setSelectedElement(newReality);
  };

  const handleNodeMouseDown = (e: React.MouseEvent<HTMLDivElement>, node: EditorNode) => {
    e.stopPropagation();
    captureGestureBaseSnapshot();
    setSelectedElement(node);
    setDraggingNode(node.id);

    if (node.type === "note") {
      setDragStartInfo({
        mouseX: e.clientX,
        mouseY: e.clientY,
        nodeStartX: node.x,
        nodeStartY: node.y,
        children: [],
        parentUniverse: null,
        isUniverse: false,
        isNote: true,
      });
      return;
    }

    const childrenNodes: Array<{ id: string; startX: number; startY: number }> = [];
    let parentUniverse: Extract<EditorNode, { type: "universe" }> | null = null;

    if (node.type === "universe") {
      nodes.forEach((n) => {
        if (n.type === "reality" && n.data.universeId === node.id) {
          childrenNodes.push({ id: n.id, startX: n.x, startY: n.y });
        }
      });
    } else if (node.type === "reality") {
      parentUniverse =
        nodes.find((u): u is Extract<EditorNode, { type: "universe" }> => u.type === "universe" && u.id === node.data.universeId) || null;
    }

    setDragStartInfo({
      mouseX: e.clientX,
      mouseY: e.clientY,
      nodeStartX: node.x,
      nodeStartY: node.y,
      children: childrenNodes,
      parentUniverse,
      isUniverse: node.type === "universe",
    });
  };

  const handleTransitionMouseDown = (
    event: React.MouseEvent<Element>,
    transition: EditorTransition,
  ) => {
    captureGestureBaseSnapshot();
    setSelectedElement({
      type: "transition",
      id: transition.id,
      data: transition,
    });
    setDraggingTransition({
      transitionId: transition.id,
      mouseX: event.clientX,
      mouseY: event.clientY,
      startOffsetX: transition.visualOffset?.x || 0,
      startOffsetY: transition.visualOffset?.y || 0,
    });
  };

  const handleResizeMouseDown = (e: React.MouseEvent<HTMLDivElement>, node: Extract<EditorNode, { type: "universe" }>) => {
    e.stopPropagation();
    captureGestureBaseSnapshot();
    setSelectedElement(node);
    setResizingStart({
      id: node.id,
      mouseX: e.clientX,
      mouseY: e.clientY,
      startW: node.w,
      startH: node.h,
      nodeX: node.x,
      nodeY: node.y,
    });
  };

  const startConnectionFromRealitySource = (
    e: React.MouseEvent<HTMLDivElement>,
    sourceRealityId: string,
  ) => {
    e.preventDefault();
    e.stopPropagation();

    const sourceSize = nodeSizes[sourceRealityId] || { w: 192, h: 80 };
    const sourceNode = nodes.find((n) => n.id === sourceRealityId);
    if (!sourceNode) {
      return;
    }

    setConnectingStart({
      kind: "reality",
      sourceRealityId,
      startX: sourceNode.x + sourceSize.w + 6,
      startY: sourceNode.y + sourceSize.h / 2,
    });
    setMousePos(getCanvasCoords(e));
  };

  const startConnectionFromTransitionOutput = (
    event: React.MouseEvent<Element>,
    transitionId: string,
  ) => {
    event.preventDefault();
    event.stopPropagation();

    const transition = transitionsById.get(transitionId);
    if (!transition) {
      return;
    }

    const routeGeometry = transitionRouteGeometryByTransitionId.get(transitionId);
    let startX: number;
    let startY: number;

    if (routeGeometry) {
      startX = routeGeometry.rightPort.x;
      startY = routeGeometry.rightPort.y;
    } else {
      const badgeAnchor = transitionBadgeAnchors.get(transitionId);
      if (!badgeAnchor) {
        return;
      }
      startX = badgeAnchor.x + BADGE_WIDTH / 2;
      startY = badgeAnchor.y;
    }

    setConnectingStart({
      kind: "transition",
      transitionId,
      sourceRealityId: transition.sourceRealityId,
      startX,
      startY,
    });
    setMousePos(getCanvasCoords(event));
  };

  const completeConnectionOnTargetPort = (
    e: React.MouseEvent<HTMLDivElement>,
    nodeId: string,
  ) => {
    e.stopPropagation();

    if (!connectingStart || connectingStart.sourceRealityId === nodeId) {
      setConnectingStart(null);
      return;
    }

    const targetRef = buildTargetReferenceFromNodes(
      connectingStart.sourceRealityId,
      nodeId,
      nodes,
    );
    if (!targetRef) {
      setConnectingStart(null);
      return;
    }

    if (connectingStart.kind === "transition") {
      const sourceTransition = transitionsById.get(connectingStart.transitionId);
      if (!sourceTransition) {
        setConnectingStart(null);
        return;
      }

      if (sourceTransition.targets.includes(targetRef)) {
        setConnectingStart(null);
        return;
      }

      markDirtyFromImport();

      const updatedTransition: EditorTransition = {
        ...sourceTransition,
        targets: [...sourceTransition.targets, targetRef],
      };

      setTransitions((prev) =>
        prev.map((transition) =>
          transition.id === sourceTransition.id
            ? { ...transition, targets: [...transition.targets, targetRef] }
            : transition,
        ),
      );
      setSelectedElement({
        type: "transition",
        id: updatedTransition.id,
        data: updatedTransition,
      });
      setConnectingStart(null);
      return;
    }

    const selectedTransition =
      selectedElement?.type === "transition" ? transitionsById.get(selectedElement.id) : null;

    if (
      selectedTransition &&
      selectedTransition.sourceRealityId === connectingStart.sourceRealityId
    ) {
      if (selectedTransition.targets.includes(targetRef)) {
        setConnectingStart(null);
        return;
      }

      markDirtyFromImport();

      const updatedTransition: EditorTransition = {
        ...selectedTransition,
        targets: [...selectedTransition.targets, targetRef],
      };

      setTransitions((prev) =>
        prev.map((transition) =>
          transition.id === selectedTransition.id
            ? { ...transition, targets: [...transition.targets, targetRef] }
            : transition,
        ),
      );
      setSelectedElement({
        type: "transition",
        id: updatedTransition.id,
        data: updatedTransition,
      });
      setConnectingStart(null);
      return;
    }

    const sourceTransitionsWithSameTrigger = transitions.filter(
      (transition) =>
        transition.sourceRealityId === connectingStart.sourceRealityId &&
        transition.triggerKind === "on" &&
        transition.eventName === "NEW_EVENT",
    );

    markDirtyFromImport();

    const newTransition: EditorTransition = {
      id: `tr-${Date.now()}`,
      sourceRealityId: connectingStart.sourceRealityId,
      triggerKind: "on",
      eventName: "NEW_EVENT",
      type: "default",
      condition: undefined,
      conditions: [],
      actions: [],
      invokes: [],
      description: "",
      metadata: "",
      targets: [targetRef],
      order: sourceTransitionsWithSameTrigger.length,
    };

    setTransitions((prev) => [...prev, newTransition]);
    setSelectedElement({ type: "transition", id: newTransition.id, data: newTransition });
    setConnectingStart(null);
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const currentScale = getTransformState().scale;

    if (draggingTransition) {
      const dx = (e.clientX - draggingTransition.mouseX) / currentScale;
      const dy = (e.clientY - draggingTransition.mouseY) / currentScale;
      const nextOffsetX = draggingTransition.startOffsetX + dx;
      const nextOffsetY = draggingTransition.startOffsetY + dy;

      setTransitions(
        (currentTransitions) =>
          currentTransitions.map((transition) =>
            transition.id === draggingTransition.transitionId
              ? {
                  ...transition,
                  visualOffset: {
                    x: nextOffsetX,
                    y: nextOffsetY,
                  },
                }
              : transition,
          ),
        {
          mode: "silent",
          markDirtyFromImport: false,
        },
      );
    } else if (draggingNode && dragStartInfo) {
      const dx = (e.clientX - dragStartInfo.mouseX) / currentScale;
      const dy = (e.clientY - dragStartInfo.mouseY) / currentScale;

      let targetX = dragStartInfo.nodeStartX + dx;
      let targetY = dragStartInfo.nodeStartY + dy;

      if (dragStartInfo.isNote) {
        setNodes(
          (currentNodes) =>
            currentNodes.map((node) =>
              node.id === draggingNode ? { ...node, x: targetX, y: targetY } : node,
            ),
          {
            mode: "silent",
            markDirtyFromImport: false,
          },
        );
        return;
      }

      let expandUniverse: { id: string; w: number; h: number } | null = null;

      if (dragStartInfo.isUniverse) {
        const currentUniverse = nodes.find(
          (node): node is Extract<EditorNode, { type: "universe" }> =>
            node.id === draggingNode && node.type === "universe",
        );
        if (!currentUniverse) {
          return;
        }

        nodes.forEach((node) => {
          if (node.type !== "universe" || node.id === draggingNode) {
            return;
          }

          const overlapsV = !(
            targetY + currentUniverse.h + COLLISION_MARGIN <= node.y ||
            targetY >= node.y + node.h + COLLISION_MARGIN
          );
          const overlapsH = !(
            targetX + currentUniverse.w + COLLISION_MARGIN <= node.x ||
            targetX >= node.x + node.w + COLLISION_MARGIN
          );

          if (!overlapsV || !overlapsH) {
            return;
          }

          const wasAbove = currentUniverse.y + currentUniverse.h + COLLISION_MARGIN <= node.y;
          const wasBelow = currentUniverse.y >= node.y + node.h + COLLISION_MARGIN;
          const wasLeft = currentUniverse.x + currentUniverse.w + COLLISION_MARGIN <= node.x;
          const wasRight = currentUniverse.x >= node.x + node.w + COLLISION_MARGIN;

          if (wasAbove) targetY = node.y - currentUniverse.h - COLLISION_MARGIN;
          else if (wasBelow) targetY = node.y + node.h + COLLISION_MARGIN;
          else if (wasLeft) targetX = node.x - currentUniverse.w - COLLISION_MARGIN;
          else if (wasRight) targetX = node.x + node.w + COLLISION_MARGIN;
          else {
            const pushRight = targetX + currentUniverse.w + COLLISION_MARGIN - node.x;
            const pushLeft = node.x + node.w + COLLISION_MARGIN - targetX;
            const pushDown = targetY + currentUniverse.h + COLLISION_MARGIN - node.y;
            const pushUp = node.y + node.h + COLLISION_MARGIN - targetY;
            const minPush = Math.min(pushRight, pushLeft, pushDown, pushUp);

            if (minPush === pushRight) targetX = node.x - currentUniverse.w - COLLISION_MARGIN;
            else if (minPush === pushLeft) targetX = node.x + node.w + COLLISION_MARGIN;
            else if (minPush === pushDown) targetY = node.y - currentUniverse.h - COLLISION_MARGIN;
            else targetY = node.y + node.h + COLLISION_MARGIN;
          }
        });
      } else {
        const currentReality = nodes.find(
          (node): node is Extract<EditorNode, { type: "reality" }> =>
            node.id === draggingNode && node.type === "reality",
        );
        if (!currentReality) {
          return;
        }

        const realityW = nodeSizes[draggingNode]?.w || DEFAULT_REALITY_WIDTH;
        const realityH = nodeSizes[draggingNode]?.h || DEFAULT_REALITY_HEIGHT;

        if (dragStartInfo.parentUniverse) {
          targetX = Math.max(dragStartInfo.parentUniverse.x + 20, targetX);
          targetY = Math.max(dragStartInfo.parentUniverse.y + 40, targetY);
        }

        nodes.forEach((node) => {
          if (node.type !== "reality" || node.id === draggingNode) {
            return;
          }

          const nodeW = nodeSizes[node.id]?.w || DEFAULT_REALITY_WIDTH;
          const nodeH = nodeSizes[node.id]?.h || DEFAULT_REALITY_HEIGHT;

          const overlapsV = !(
            targetY + realityH + COLLISION_MARGIN <= node.y ||
            targetY >= node.y + nodeH + COLLISION_MARGIN
          );
          const overlapsH = !(
            targetX + realityW + COLLISION_MARGIN <= node.x ||
            targetX >= node.x + nodeW + COLLISION_MARGIN
          );

          if (!overlapsV || !overlapsH) {
            return;
          }

          const wasAbove = currentReality.y + realityH + COLLISION_MARGIN <= node.y;
          const wasBelow = currentReality.y >= node.y + nodeH + COLLISION_MARGIN;
          const wasLeft = currentReality.x + realityW + COLLISION_MARGIN <= node.x;
          const wasRight = currentReality.x >= node.x + nodeW + COLLISION_MARGIN;

          if (wasAbove) targetY = node.y - realityH - COLLISION_MARGIN;
          else if (wasBelow) targetY = node.y + nodeH + COLLISION_MARGIN;
          else if (wasLeft) targetX = node.x - realityW - COLLISION_MARGIN;
          else if (wasRight) targetX = node.x + nodeW + COLLISION_MARGIN;
          else {
            const pushRight = targetX + realityW + COLLISION_MARGIN - node.x;
            const pushLeft = node.x + nodeW + COLLISION_MARGIN - targetX;
            const pushDown = targetY + realityH + COLLISION_MARGIN - node.y;
            const pushUp = node.y + nodeH + COLLISION_MARGIN - targetY;
            const minPush = Math.min(pushRight, pushLeft, pushDown, pushUp);

            if (minPush === pushRight) targetX = node.x - realityW - COLLISION_MARGIN;
            else if (minPush === pushLeft) targetX = node.x + nodeW + COLLISION_MARGIN;
            else if (minPush === pushDown) targetY = node.y - realityH - COLLISION_MARGIN;
            else targetY = node.y + nodeH + COLLISION_MARGIN;
          }
        });

        if (dragStartInfo.parentUniverse) {
          const parentUniverse = dragStartInfo.parentUniverse;
          targetX = Math.max(parentUniverse.x + 20, targetX);
          targetY = Math.max(parentUniverse.y + 40, targetY);

          const neededRight = targetX + realityW + 20;
          const neededBottom = targetY + realityH + 20;

          let proposedW = Math.max(parentUniverse.w, neededRight - parentUniverse.x);
          let proposedH = Math.max(parentUniverse.h, neededBottom - parentUniverse.y);

          nodes.forEach((node) => {
            if (node.type !== "universe" || node.id === parentUniverse.id) {
              return;
            }

            const overlapsV = !(
              parentUniverse.y + proposedH + COLLISION_MARGIN <= node.y ||
              parentUniverse.y >= node.y + node.h + COLLISION_MARGIN
            );
            const overlapsH = !(
              parentUniverse.x + proposedW + COLLISION_MARGIN <= node.x ||
              parentUniverse.x >= node.x + node.w + COLLISION_MARGIN
            );

            if (!overlapsV || !overlapsH) {
              return;
            }

            if (parentUniverse.x + parentUniverse.w + COLLISION_MARGIN <= node.x) {
              proposedW = node.x - COLLISION_MARGIN - parentUniverse.x;
              targetX = Math.min(targetX, parentUniverse.x + proposedW - 20 - realityW);
            }
            if (parentUniverse.y + parentUniverse.h + COLLISION_MARGIN <= node.y) {
              proposedH = node.y - COLLISION_MARGIN - parentUniverse.y;
              targetY = Math.min(targetY, parentUniverse.y + proposedH - 20 - realityH);
            }
          });

          expandUniverse = { id: parentUniverse.id, w: proposedW, h: proposedH };
        }
      }

      setNodes(
        (currentNodes) =>
          currentNodes.map((node) => {
            if (node.id === draggingNode) {
              return { ...node, x: targetX, y: targetY } as EditorNode;
            }

            if (dragStartInfo.children.some((child) => child.id === node.id)) {
              const child = dragStartInfo.children.find((entry) => entry.id === node.id);
              if (!child) {
                return node;
              }

              const finalDx = targetX - dragStartInfo.nodeStartX;
              const finalDy = targetY - dragStartInfo.nodeStartY;
              return { ...node, x: child.startX + finalDx, y: child.startY + finalDy } as EditorNode;
            }

            if (expandUniverse && node.id === expandUniverse.id && node.type === "universe") {
              return { ...node, w: expandUniverse.w, h: expandUniverse.h };
            }

            return node;
          }),
        {
          mode: "silent",
          markDirtyFromImport: false,
        },
      );
    } else if (resizingStart) {
      const dx = (e.clientX - resizingStart.mouseX) / currentScale;
      const dy = (e.clientY - resizingStart.mouseY) / currentScale;
      let newW = Math.max(200, resizingStart.startW + dx);
      let newH = Math.max(150, resizingStart.startH + dy);

      const realitiesInside = nodes.filter(
        (node): node is Extract<EditorNode, { type: "reality" }> =>
          node.type === "reality" && node.data.universeId === resizingStart.id,
      );

      realitiesInside.forEach((reality) => {
        const realityW = nodeSizes[reality.id]?.w || DEFAULT_REALITY_WIDTH;
        const realityH = nodeSizes[reality.id]?.h || DEFAULT_REALITY_HEIGHT;
        newW = Math.max(newW, reality.x + realityW + 20 - resizingStart.nodeX);
        newH = Math.max(newH, reality.y + realityH + 20 - resizingStart.nodeY);
      });

      nodes.forEach((node) => {
        if (node.type !== "universe" || node.id === resizingStart.id) {
          return;
        }

        const overlapsV = !(
          resizingStart.nodeY + newH + COLLISION_MARGIN <= node.y ||
          resizingStart.nodeY >= node.y + node.h + COLLISION_MARGIN
        );
        const overlapsH = !(
          resizingStart.nodeX + newW + COLLISION_MARGIN <= node.x ||
          resizingStart.nodeX >= node.x + node.w + COLLISION_MARGIN
        );

        if (!overlapsV || !overlapsH) {
          return;
        }

        if (resizingStart.nodeX + resizingStart.startW + COLLISION_MARGIN <= node.x) {
          newW = Math.min(newW, node.x - COLLISION_MARGIN - resizingStart.nodeX);
        }
        if (resizingStart.nodeY + resizingStart.startH + COLLISION_MARGIN <= node.y) {
          newH = Math.min(newH, node.y - COLLISION_MARGIN - resizingStart.nodeY);
        }
      });

      setNodes(
        (currentNodes) =>
          currentNodes.map((node) =>
            node.id === resizingStart.id && node.type === "universe"
              ? { ...node, w: newW, h: newH }
              : node,
          ),
        {
          mode: "silent",
          markDirtyFromImport: false,
        },
      );
    } else if (connectingStart) {
      setMousePos(getCanvasCoords(e));
    }
  };

  const handleMouseUp = () => {
    commitGestureHistoryStep();
    setDraggingTransition(null);
    setDraggingNode(null);
    setDragStartInfo(null);
    setResizingStart(null);
    setConnectingStart(null);
  };

  const handleCanvasClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!isCanvasBackgroundTarget(e.target)) {
      return;
    }

    if (panStartedRef.current || panMovedRef.current) {
      panStartedRef.current = false;
      panMovedRef.current = false;
      return;
    }

    setSelectedElement(null);
    setIsModalOpen(false);
  };

  const handleCanvasMouseDownCapture = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!isCanvasBackgroundTarget(e.target)) {
      return;
    }

    panStartedRef.current = false;
    panMovedRef.current = false;
  };

  const commitUniverseIdRename = useCallback(
    (universeNodeId: string, nextUniverseIdDraft: string): string => {
      const universeNode = nodes.find(
        (node): node is Extract<EditorNode, { type: "universe" }> =>
          node.type === "universe" && node.id === universeNodeId,
      );
      if (!universeNode) {
        return "";
      }

      const usedUniverseIds = new Set(
        nodes
          .filter(
            (node): node is Extract<EditorNode, { type: "universe" }> =>
              node.type === "universe" && node.id !== universeNodeId,
          )
          .map((node) => node.data.id)
          .filter(Boolean),
      );

      const cleanUniverseId = cleanIdentifier(nextUniverseIdDraft) || "universe";
      const resolvedUniverseId = ensureUniqueIdentifier(
        cleanUniverseId,
        usedUniverseIds,
        "universe",
      );
      const shouldRename = universeNode.data.id !== resolvedUniverseId;
      const shouldSyncName = universeNode.data.name !== resolvedUniverseId;

      if (!shouldRename && !shouldSyncName) {
        return resolvedUniverseId;
      }

      markDirtyFromImport();
      const historyGroup = `rename-universe-id:${universeNodeId}`;

      if (shouldRename) {
        const remapped = renameUniverseId({
          nodes,
          transitions,
          machineConfig,
          metadataPackBindings,
          universeNodeId,
          nextUniverseId: resolvedUniverseId,
        });

        setNodes(remapped.nodes, {
          mode: "coalesce",
          group: historyGroup,
        });
        setTransitions(remapped.transitions, {
          mode: "coalesce",
          group: historyGroup,
        });
        setMachineConfig(remapped.machineConfig, {
          mode: "coalesce",
          group: historyGroup,
        });
        setMetadataPackBindings(remapped.metadataPackBindings, {
          mode: "coalesce",
          group: historyGroup,
        });
        setSelectedElement((previous) => {
          if (
            !previous ||
            previous.type !== "universe" ||
            previous.id !== universeNodeId
          ) {
            return previous;
          }

          return (
            remapped.nodes.find(
              (node): node is InspectableEditorNode =>
                node.id === universeNodeId && node.type !== "note",
            ) || previous
          );
        });
        return resolvedUniverseId;
      }

      setNodes(
        (previous) =>
          previous.map((node) =>
            node.type === "universe" && node.id === universeNodeId
              ? {
                  ...node,
                  data: {
                    ...node.data,
                    name: resolvedUniverseId,
                  },
                }
              : node,
          ),
        {
          mode: "coalesce",
          group: historyGroup,
        },
      );
      setSelectedElement((previous) => {
        if (
          !previous ||
          previous.type !== "universe" ||
          previous.id !== universeNodeId
        ) {
          return previous;
        }

        return {
          ...previous,
          data: {
            ...previous.data,
            name: resolvedUniverseId,
          },
        };
      });
      return resolvedUniverseId;
    },
    [machineConfig, metadataPackBindings, nodes, transitions],
  );

  const commitUniverseCanonicalRename = useCallback(
    (
      universeNodeId: string,
      nextCanonicalDraft: string,
      options: { syncId: boolean } = { syncId: false },
    ): { id: string; canonicalName: string } => {
      const universeNode = nodes.find(
        (node): node is Extract<EditorNode, { type: "universe" }> =>
          node.type === "universe" && node.id === universeNodeId,
      );
      if (!universeNode) {
        return { id: "", canonicalName: "" };
      }

      const usedUniverseIdentifiers = new Set(
        nodes
          .filter(
            (node): node is Extract<EditorNode, { type: "universe" }> =>
              node.type === "universe" && node.id !== universeNodeId,
          )
          .flatMap((node) => [node.data.id, node.data.canonicalName || node.data.id])
          .filter(Boolean),
      );

      const cleanCanonical = cleanIdentifier(nextCanonicalDraft) || "universe";
      const resolvedCanonical = ensureUniqueIdentifier(
        cleanCanonical,
        usedUniverseIdentifiers,
        "universe",
      );

      const currentUniverseId = universeNode.data.id;
      const currentCanonical = universeNode.data.canonicalName || currentUniverseId;

      if (options.syncId) {
        const nextUniverseId = resolvedCanonical;
        const shouldRenameUniverse = currentUniverseId !== nextUniverseId;
        const shouldUpdateCanonical = currentCanonical !== nextUniverseId;
        const shouldSyncName = universeNode.data.name !== nextUniverseId;

        if (!shouldRenameUniverse && !shouldUpdateCanonical && !shouldSyncName) {
          return {
            id: nextUniverseId,
            canonicalName: nextUniverseId,
          };
        }

        markDirtyFromImport();
        const historyGroup = `rename-universe-canonical:${universeNodeId}`;

        if (shouldRenameUniverse) {
          const remapped = renameUniverseId({
            nodes,
            transitions,
            machineConfig,
            metadataPackBindings,
            universeNodeId,
            nextUniverseId,
          });
          const nextNodes = remapped.nodes.map((node) =>
            node.type === "universe" && node.id === universeNodeId
              ? {
                  ...node,
                  data: {
                    ...node.data,
                    canonicalName: nextUniverseId,
                    name: nextUniverseId,
                  },
                }
              : node,
          );

          setNodes(nextNodes, {
            mode: "coalesce",
            group: historyGroup,
          });
          setTransitions(remapped.transitions, {
            mode: "coalesce",
            group: historyGroup,
          });
          setMachineConfig(remapped.machineConfig, {
            mode: "coalesce",
            group: historyGroup,
          });
          setMetadataPackBindings(remapped.metadataPackBindings, {
            mode: "coalesce",
            group: historyGroup,
          });
          setSelectedElement((previous) => {
            if (
              !previous ||
              previous.type !== "universe" ||
              previous.id !== universeNodeId
            ) {
              return previous;
            }

            return (
              nextNodes.find(
                (node): node is InspectableEditorNode =>
                  node.id === universeNodeId && node.type !== "note",
              ) || previous
            );
          });
          return {
            id: nextUniverseId,
            canonicalName: nextUniverseId,
          };
        }

        setNodes(
          (previous) =>
            previous.map((node) =>
              node.type === "universe" && node.id === universeNodeId
                ? {
                    ...node,
                    data: {
                      ...node.data,
                      canonicalName: nextUniverseId,
                      name: nextUniverseId,
                    },
                  }
                : node,
            ),
          {
            mode: "coalesce",
            group: historyGroup,
          },
        );
        setSelectedElement((previous) => {
          if (
            !previous ||
            previous.type !== "universe" ||
            previous.id !== universeNodeId
          ) {
            return previous;
          }

          return {
            ...previous,
            data: {
              ...previous.data,
              canonicalName: nextUniverseId,
              name: nextUniverseId,
            },
          };
        });
        return {
          id: nextUniverseId,
          canonicalName: nextUniverseId,
        };
      }

      if (currentCanonical === resolvedCanonical) {
        return {
          id: currentUniverseId,
          canonicalName: resolvedCanonical,
        };
      }

      markDirtyFromImport();
      const historyGroup = `rename-universe-canonical:${universeNodeId}`;
      setNodes(
        (previous) =>
          previous.map((node) =>
            node.type === "universe" && node.id === universeNodeId
              ? {
                  ...node,
                  data: {
                    ...node.data,
                    canonicalName: resolvedCanonical,
                  },
                }
              : node,
          ),
        {
          mode: "coalesce",
          group: historyGroup,
        },
      );
      setSelectedElement((previous) => {
        if (
          !previous ||
          previous.type !== "universe" ||
          previous.id !== universeNodeId
        ) {
          return previous;
        }

        return {
          ...previous,
          data: {
            ...previous.data,
            canonicalName: resolvedCanonical,
          },
        };
      });
      return {
        id: currentUniverseId,
        canonicalName: resolvedCanonical,
      };
    },
    [machineConfig, metadataPackBindings, nodes, transitions],
  );

  const commitRealityIdRename = useCallback(
    (realityNodeId: string, nextRealityIdDraft: string): string => {
      const realityNode = nodes.find(
        (node): node is Extract<EditorNode, { type: "reality" }> =>
          node.type === "reality" && node.id === realityNodeId,
      );
      if (!realityNode) {
        return "";
      }

      const usedRealityIds = new Set(
        nodes
          .filter(
            (node): node is Extract<EditorNode, { type: "reality" }> =>
              node.type === "reality" &&
              node.data.universeId === realityNode.data.universeId &&
              node.id !== realityNodeId,
          )
          .map((node) => node.data.id)
          .filter(Boolean),
      );

      const cleanRealityId = cleanIdentifier(nextRealityIdDraft) || "reality";
      const resolvedRealityId = ensureUniqueIdentifier(
        cleanRealityId,
        usedRealityIds,
        "reality",
      );
      const shouldRename = realityNode.data.id !== resolvedRealityId;
      const shouldSyncName = realityNode.data.name !== resolvedRealityId;

      if (!shouldRename && !shouldSyncName) {
        return resolvedRealityId;
      }

      markDirtyFromImport();
      const historyGroup = `rename-reality-id:${realityNodeId}`;

      if (shouldRename) {
        const remapped = renameRealityId({
          nodes,
          transitions,
          machineConfig,
          metadataPackBindings,
          realityNodeId,
          nextRealityId: resolvedRealityId,
        });

        setNodes(remapped.nodes, {
          mode: "coalesce",
          group: historyGroup,
        });
        setTransitions(remapped.transitions, {
          mode: "coalesce",
          group: historyGroup,
        });
        setMachineConfig(remapped.machineConfig, {
          mode: "coalesce",
          group: historyGroup,
        });
        setMetadataPackBindings(remapped.metadataPackBindings, {
          mode: "coalesce",
          group: historyGroup,
        });
        setSelectedElement((previous) => {
          if (
            !previous ||
            previous.type !== "reality" ||
            previous.id !== realityNodeId
          ) {
            return previous;
          }

          return (
            remapped.nodes.find(
              (node): node is InspectableEditorNode =>
                node.id === realityNodeId && node.type !== "note",
            ) || previous
          );
        });
        return resolvedRealityId;
      }

      setNodes(
        (previous) =>
          previous.map((node) =>
            node.type === "reality" && node.id === realityNodeId
              ? {
                  ...node,
                  data: {
                    ...node.data,
                    name: resolvedRealityId,
                  },
                }
              : node,
          ),
        {
          mode: "coalesce",
          group: historyGroup,
        },
      );
      setSelectedElement((previous) => {
        if (
          !previous ||
          previous.type !== "reality" ||
          previous.id !== realityNodeId
        ) {
          return previous;
        }

        return {
          ...previous,
          data: {
            ...previous.data,
            name: resolvedRealityId,
          },
        };
      });
      return resolvedRealityId;
    },
    [machineConfig, metadataPackBindings, nodes, transitions],
  );

  const updateModalNodeData = (field: string, value: unknown) => {
    if (!isInspectableEditorNode(selectedElement)) return;

    markDirtyFromImport();

    setNodes(
      (nds) =>
        nds.map((n) =>
          n.id === selectedElement.id && n.type !== "note"
            ? ({ ...n, data: { ...n.data, [field]: value } } as EditorNode)
            : n,
        ),
      {
        mode: "coalesce",
        group: `modal-node:${selectedElement.id}:${field}`,
      },
    );

    setSelectedElement((prev) => {
      if (!isInspectableEditorNode(prev)) {
        return prev;
      }

      return {
        ...prev,
        data: { ...prev.data, [field]: value },
      } as InspectableEditorNode;
    });
  };

  const updateModalTransitionPatch = (
    patch: Partial<EditorTransition>,
    groupKey = "patch",
  ) => {
    if (!selectedElement || selectedElement.type !== "transition") return;

    markDirtyFromImport();

    const transitionId = selectedElement.id;

    setTransitions(
      (prevTransitions) => {
        const nextTransitions = prevTransitions.map((transition) => {
          if (transition.id !== transitionId) {
            return transition;
          }

          return {
            ...transition,
            ...patch,
          } as EditorTransition;
        });

        return nextTransitions;
      },
      {
        mode: "coalesce",
        group: `modal-transition:${transitionId}:${groupKey}`,
      },
    );

    setSelectedElement((previous) => {
      if (!previous || previous.type !== "transition" || previous.id !== transitionId) {
        return previous;
      }

      return {
        type: "transition",
        id: transitionId,
        data: {
          ...previous.data,
          ...patch,
        } as EditorTransition,
      };
    });
  };

  const updateModalTransitionData = (field: string, value: unknown) => {
    updateModalTransitionPatch(
      { [field]: value } as Partial<EditorTransition>,
      field,
    );
  };

  const moveSelectedTransition = (direction: "up" | "down") => {
    if (!selectedElement || selectedElement.type !== "transition") {
      return;
    }

    markDirtyFromImport();

    setTransitions((prevTransitions) =>
      moveTransitionInsideGroup(prevTransitions, selectedElement.id, direction),
    );
  };

  const createAnchoredNote = () => ({
    text: "",
    colorIndex: 0,
  });

  const updateNodeAnchoredNote = (
    nodeId: string,
    note: AnchoredNoteData | null,
  ) => {
    markDirtyFromImport();
    setNodes(
      (previous) =>
        previous.map((node) => {
          if (node.id !== nodeId) {
            return node;
          }

          if (node.type === "universe") {
            return {
              ...node,
              data: {
                ...node.data,
                note,
              },
            };
          }

          if (node.type === "reality") {
            return {
              ...node,
              data: {
                ...node.data,
                note,
              },
            };
          }

          return node;
        }),
      {
        mode: "coalesce",
        group: `node-note:${nodeId}`,
        markDirtyFromImport: false,
      },
    );

    setSelectedElement((previous) => {
      if (!isInspectableEditorNode(previous) || previous.id !== nodeId) {
        return previous;
      }

      return {
        ...previous,
        data: {
          ...previous.data,
          note,
        },
      } as InspectableEditorNode;
    });
  };

  const updateTransitionNote = (
    transitionId: string,
    note: AnchoredNoteData | null,
  ) => {
    markDirtyFromImport();
    setTransitions(
      (previous) =>
        previous.map((transition) =>
          transition.id === transitionId
            ? {
                ...transition,
                note,
              }
            : transition,
        ),
      {
        mode: "coalesce",
        group: `transition-note:${transitionId}`,
        markDirtyFromImport: false,
      },
    );

    setSelectedElement((previous) => {
      if (!previous || previous.type !== "transition" || previous.id !== transitionId) {
        return previous;
      }

      return {
        ...previous,
        data: {
          ...previous.data,
          note,
        },
      };
    });
  };

  const updateGlobalNote = (noteId: string, data: GlobalNoteData) => {
    markDirtyFromImport();
    setNodes(
      (previous) =>
        previous.map((node) =>
          node.id === noteId && node.type === "note"
            ? {
                ...node,
                data,
              }
            : node,
        ),
      {
        mode: "coalesce",
        group: `global-note:${noteId}`,
        markDirtyFromImport: false,
      },
    );

    setSelectedElement((previous) => {
      if (!previous || previous.type !== "note" || previous.id !== noteId) {
        return previous;
      }

      return {
        ...previous,
        data,
      };
    });
  };

  const deleteElement = (elementToDelete: any = selectedElement) => {
    if (!elementToDelete) return;

    if (elementToDelete.type === "transition") {
      markDirtyFromImport();
      const transitionId = elementToDelete.id;
      setTransitions((prev) => prev.filter((transition) => transition.id !== transitionId));
      if (selectedElement?.id === transitionId) {
        setSelectedElement(null);
        setIsModalOpen(false);
      }
      return;
    }

    const nodeToDelete = elementToDelete as EditorNode;
    if (
      nodeToDelete.type === "universe" ||
      nodeToDelete.type === "reality" ||
      nodeToDelete.type === "note"
    ) {
      markDirtyFromImport();
      const historyGroup = `delete-node:${nodeToDelete.id}`;
      const markDirtyForDelete = nodeToDelete.type !== "note";
      let nodesToDelete = [nodeToDelete.id];
      if (nodeToDelete.type === "universe") {
        const children = nodes.filter(
          (n) => n.type === "reality" && n.data.universeId === nodeToDelete.id,
        );
        nodesToDelete = [...nodesToDelete, ...children.map((child) => child.id)];
      }

      setNodes((prev) => prev.filter((node) => !nodesToDelete.includes(node.id)), {
        mode: "coalesce",
        group: historyGroup,
        markDirtyFromImport: markDirtyForDelete,
      });
      setTransitions(
        (prev) => removeTransitionsReferencingDeletedNodes(prev, nodes, nodesToDelete),
        {
          mode: "coalesce",
          group: historyGroup,
          markDirtyFromImport: markDirtyForDelete,
        },
      );

      if (
        selectedElement &&
        selectedElement.type !== "transition" &&
        nodesToDelete.includes(selectedElement.id)
      ) {
        setSelectedElement(null);
        setIsModalOpen(false);
      }
    }
  };

  useEffect(() => {
    const isEditableTarget = (target: EventTarget | null): boolean => {
      if (!(target instanceof HTMLElement)) {
        return false;
      }

      if (target.isContentEditable) {
        return true;
      }

      const tag = target.tagName.toLowerCase();
      return tag === "input" || tag === "textarea" || tag === "select";
    };

    const handleDeleteKeyDown = (event: KeyboardEvent) => {
      if (isEditableTarget(event.target) || !selectedElement) {
        return;
      }

      const key = event.key.toLowerCase();
      if (key !== "delete" && key !== "backspace") {
        return;
      }

      event.preventDefault();
      deleteElement(selectedElement);
    };

    window.addEventListener("keydown", handleDeleteKeyDown);
    return () => {
      window.removeEventListener("keydown", handleDeleteKeyDown);
    };
  }, [deleteElement, selectedElement]);

  const cloneReality = (realityId: string) => {
    markDirtyFromImport();
    const r = nodes.find(
      (n): n is Extract<EditorNode, { type: "reality" }> => n.id === realityId && n.type === "reality",
    );
    if (!r) return;

    const newId = `reality-${Date.now()}`;
    let newX = r.x;
    let newY = r.y + 170;
    let overlaps = true;

    while (overlaps) {
      overlaps = false;
      nodes.forEach((node) => {
        if (node.type !== "reality" || node.id === r.id) {
          return;
        }

        const nodeW = nodeSizes[node.id]?.w || DEFAULT_REALITY_WIDTH;
        const nodeH = nodeSizes[node.id]?.h || DEFAULT_REALITY_HEIGHT;
        const overlapsV = !(
          newY + DEFAULT_REALITY_HEIGHT + COLLISION_MARGIN <= node.y ||
          newY >= node.y + nodeH + COLLISION_MARGIN
        );
        const overlapsH = !(
          newX + DEFAULT_REALITY_WIDTH + COLLISION_MARGIN <= node.x ||
          newX >= node.x + nodeW + COLLISION_MARGIN
        );

        if (overlapsV && overlapsH) {
          overlaps = true;
          newY = node.y + nodeH + COLLISION_MARGIN;
        }
      });
    }

    const newDataId = cleanIdentifier(`${r.data.id}-copy`);
    const historyGroup = `clone-reality:${realityId}:${newId}`;
    const newReality: Extract<EditorNode, { type: "reality" }> = {
      ...r,
      id: newId,
      x: newX,
      y: newY,
      data: {
        ...r.data,
        id: newDataId,
        name: newDataId,
        isInitial: false,
      },
    };

    setNodes(
      (prev) => {
        let nextNodes = [...prev, newReality];
        const parentUniverse = prev.find(
          (node): node is Extract<EditorNode, { type: "universe" }> =>
            node.type === "universe" && node.id === r.data.universeId,
        );

        if (parentUniverse) {
          const proposedW = Math.max(parentUniverse.w, newX + DEFAULT_REALITY_WIDTH + 20 - parentUniverse.x);
          const proposedH = Math.max(parentUniverse.h, newY + DEFAULT_REALITY_HEIGHT + 20 - parentUniverse.y);
          nextNodes = nextNodes.map((node) =>
            node.id === parentUniverse.id ? { ...node, w: proposedW, h: proposedH } : node,
          );
        }

        return nextNodes;
      },
      {
        mode: "coalesce",
        group: historyGroup,
      },
    );

    const sourceTransitions = transitions.filter((transition) => transition.sourceRealityId === realityId);
    if (sourceTransitions.length > 0) {
      const clonedTransitions = sourceTransitions.map((transition) => ({
        ...transition,
        id: `tr-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
        sourceRealityId: newId,
        visualOffset: undefined,
      }));
      setTransitions((prev) => [...prev, ...clonedTransitions], {
        mode: "coalesce",
        group: historyGroup,
      });
    }

    setSelectedElement(newReality);
  };

  const cloneUniverse = (universeId: string) => {
    markDirtyFromImport();
    const universe = nodes.find(
      (n): n is Extract<EditorNode, { type: "universe" }> => n.id === universeId && n.type === "universe",
    );
    if (!universe) return;

    const children = nodes.filter(
      (n): n is Extract<EditorNode, { type: "reality" }> =>
        n.type === "reality" && n.data.universeId === universe.id,
    );

    const newUniverseId = `universe-${Date.now()}`;
    let newX = universe.x;
    let newY = universe.y + universe.h + 30;
    let overlaps = true;

    while (overlaps) {
      overlaps = false;
      nodes.forEach((node) => {
        if (node.type !== "universe" || node.id === universe.id) {
          return;
        }

        const overlapsV = !(
          newY + universe.h + COLLISION_MARGIN <= node.y ||
          newY >= node.y + node.h + COLLISION_MARGIN
        );
        const overlapsH = !(
          newX + universe.w + COLLISION_MARGIN <= node.x ||
          newX >= node.x + node.w + COLLISION_MARGIN
        );

        if (overlapsV && overlapsH) {
          overlaps = true;
          newY = node.y + node.h + 30;
        }
      });
    }

    const universeNodes = nodes.filter(
      (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
    );
    const usedUniverseIdentifiers = new Set(
      universeNodes.flatMap((node) => [node.data.id, node.data.canonicalName || node.data.id]),
    );
    const copyBase = cleanIdentifier(
      `${universe.data.canonicalName || universe.data.id}-copy`,
    ) || "universe-copy";
    const newUniverseIdentifier = ensureUniqueIdentifier(copyBase, usedUniverseIdentifiers, "universe-copy");

    const historyGroup = `clone-universe:${universeId}:${newUniverseId}`;
    const newUniverse: Extract<EditorNode, { type: "universe" }> = {
      ...universe,
      id: newUniverseId,
      x: newX,
      y: newY,
      data: {
        ...universe.data,
        id: newUniverseIdentifier,
        name: newUniverseIdentifier,
        canonicalName: newUniverseIdentifier,
      },
    };

    const offsetX = newX - universe.x;
    const offsetY = newY - universe.y;

    const realityNodeIdMap = new Map<string, string>();
    const realityDataIdMap = new Map<string, string>();

    const clonedRealities: Extract<EditorNode, { type: "reality" }>[] = children.map((child) => {
      const nextId = `reality-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`;
      const nextDataId = cleanIdentifier(`${child.data.id}-copy`);
      realityNodeIdMap.set(child.id, nextId);
      realityDataIdMap.set(child.data.id, nextDataId);

      return {
        ...child,
        id: nextId,
        x: child.x + offsetX,
        y: child.y + offsetY,
        data: {
          ...child.data,
          universeId: newUniverseId,
          id: nextDataId,
          name: nextDataId,
        },
      };
    });

    const clonedTransitions = transitions
      .filter((transition) => realityNodeIdMap.has(transition.sourceRealityId))
      .map((transition) => {
        const mappedSource = realityNodeIdMap.get(transition.sourceRealityId) || transition.sourceRealityId;
        const mappedTargets = transition.targets.map((targetRef) => {
          if (!targetRef.startsWith("U:")) {
            return realityDataIdMap.get(targetRef) || targetRef;
          }

          if (targetRef === `U:${universe.data.id}`) {
            return `U:${newUniverseIdentifier}`;
          }

          if (targetRef.startsWith(`U:${universe.data.id}:`)) {
            const suffix = targetRef.replace(`U:${universe.data.id}:`, "");
            const mappedReality = realityDataIdMap.get(suffix) || suffix;
            return `U:${newUniverseIdentifier}:${mappedReality}`;
          }

          return targetRef;
        });

        return {
          ...transition,
          id: `tr-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
          sourceRealityId: mappedSource,
          targets: mappedTargets,
          visualOffset: undefined,
        };
      });

    setNodes((prev) => [...prev, newUniverse, ...clonedRealities], {
      mode: "coalesce",
      group: historyGroup,
    });
    setTransitions((prev) => [...prev, ...clonedTransitions], {
      mode: "coalesce",
      group: historyGroup,
    });
    setSelectedElement(newUniverse);
  };

  const fitUniverseContent = (universeId: string) => {
    setNodes((nds) => {
      const universe = nds.find((n) => n.id === universeId && n.type === "universe");
      if (!universe || universe.type !== "universe") return nds;

      const children = nds.filter(
        (n): n is Extract<EditorNode, { type: "reality" }> =>
          n.type === "reality" && n.data.universeId === universe.id,
      );

      if (children.length === 0) {
        return nds.map((n) => (n.id === universeId ? { ...n, w: 240, h: 220 } : n));
      }

      let maxRight = universe.x + 240;
      let maxBottom = universe.y + 220;

      children.forEach((child) => {
        const childW = nodeSizes[child.id]?.w || 192;
        const childH = nodeSizes[child.id]?.h || 150;
        maxRight = Math.max(maxRight, child.x + childW + 20);
        maxBottom = Math.max(maxBottom, child.y + childH + 20);
      });

      return nds.map((n) =>
        n.id === universeId
          ? {
              ...n,
              w: maxRight - universe.x,
              h: maxBottom - universe.y,
            }
          : n,
      );
    });
  };

  const handleAutoLayout = useCallback(async () => {
    if (isAutoLayouting) {
      return;
    }

    const runId = beginAutoLayoutRun();
    let applied = false;

    try {
      const nextNodes = await computeAutoLayout(nodes, transitions, nodeSizes);
      const latestState = latestEditorStateRef.current;
      const canApply = latestState.nodes === nodes && latestState.transitions === transitions;
      if (!canApply) {
        return;
      }

      const historyGroup = "auto-layout";

      setNodes(nextNodes, {
        mode: "coalesce",
        group: historyGroup,
        markDirtyFromImport: false,
      });

      const transitionsWithResetOffset = transitions.map((transition) =>
        transition.visualOffset
          ? {
              ...transition,
              visualOffset: undefined,
            }
          : transition,
      );
      const hasTransitionOffsetReset = transitionsWithResetOffset.some(
        (transition, index) => transition !== transitions[index],
      );

      if (hasTransitionOffsetReset) {
        setTransitions(transitionsWithResetOffset, {
          mode: "coalesce",
          group: historyGroup,
          markDirtyFromImport: false,
        });
      }
      applied = true;
    } catch (error) {
      console.error("Failed to compute auto-layout", error);
    } finally {
      endAutoLayoutRun(runId);
      if (applied) {
        requestAnimationFrame(() => {
          handleFitToContentRef.current("smooth");
        });
      }
    }
  }, [
    beginAutoLayoutRun,
    endAutoLayoutRun,
    isAutoLayouting,
    nodeSizes,
    nodes,
    setNodes,
    setTransitions,
    transitions,
  ]);

  const handleImportMachine = async (machine: StateProMachine) => {
    if (!features.json.import) {
      return;
    }

    const importedState = deserializeStatePro(machine);
    let nextState = importedState;
    const runId = beginAutoLayoutRun();

    try {
      const autoLayoutNodes = await computeAutoLayout(
        importedState.nodes,
        importedState.transitions,
        importedState.nodeSizes,
      );
      nextState = {
        ...importedState,
        nodes: autoLayoutNodes,
      };
    } catch (error) {
      console.error("Failed to compute auto-layout for imported model", error);
    } finally {
      endAutoLayoutRun(runId);
    }

    if (!isMountedRef.current || autoLayoutRunIdRef.current !== runId) {
      return;
    }

    changeSourceRef.current = "user";
    resetHistoryWithRef.current(nextState);
    gestureBaseSnapshotRef.current = null;
    setSelectedElement(null);
    setIsModalOpen(false);
    setShowJsonModal(false);
  };

  const handleImportLayout = (
    layout: StudioLayoutDocument,
    _parseIssues: StudioLayoutIssue[],
  ) => {
    if (!features.json.import) {
      return;
    }

    const applied = applyStudioLayoutDocument(editorState, layout);
    changeSourceRef.current = "user";
    resetHistoryWith(applied.state);
    gestureBaseSnapshotRef.current = null;
    setSelectedElement(null);
    setIsModalOpen(false);
    setShowJsonModal(false);
  };

  const serialized = useMemo(() => serializeStatePro(editorState), [editorState]);
  const serializedLayout = useMemo(() => serializeStudioLayout(editorState), [editorState]);
  const issueIndex = useMemo(
    () => buildIssueIndex(serialized.issues, nodes, transitions),
    [serialized.issues, nodes, transitions],
  );
  const selectedElementIssues = useMemo(() => {
    if (!selectedElement) {
      return [];
    }
    if (selectedElement.type === "transition") {
      return issueIndex.transitions.get(selectedElement.id) || [];
    }
    if (selectedElement.type === "universe") {
      return issueIndex.universes.get(selectedElement.id) || [];
    }
    if (selectedElement.type === "note") {
      return [];
    }
    return issueIndex.realities.get(selectedElement.id) || [];
  }, [issueIndex.realities, issueIndex.transitions, issueIndex.universes, selectedElement]);
  const propertiesModalElement =
    selectedElement && selectedElement.type !== "note" ? selectedElement : null;
  const generatedJson = useMemo(() => JSON.stringify(serialized.machine, null, 2), [serialized.machine]);
  const generatedLayoutJson = useMemo(
    () => JSON.stringify(serializedLayout, null, 2),
    [serializedLayout],
  );
  const canOpenJsonModal = features.json.import || features.json.export;

  useEffect(() => {
    if (!onChange) {
      return;
    }

    const source = changeSourceRef.current;
    const timerId = window.setTimeout(() => {
      onChange({
        machine: structuredClone(serialized.machine),
        layout: structuredClone(serializedLayout),
        issues: structuredClone(serialized.issues),
        canExport: serialized.canExport,
        source,
        at: new Date().toISOString(),
      });

      if (source === "external-sync") {
        changeSourceRef.current = "user";
      }
    }, Math.max(0, changeDebounceMs));

    return () => {
      window.clearTimeout(timerId);
    };
  }, [changeDebounceMs, onChange, serialized, serializedLayout]);

  return (
    <div className="flex flex-col h-screen bg-slate-950 text-slate-200 font-sans overflow-hidden relative">
      <header className="absolute top-0 w-full h-14 bg-slate-900/80 backdrop-blur border-b border-slate-800 flex items-center justify-between px-6 z-[60]">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-blue-600 rounded flex items-center justify-center font-bold text-white shadow-lg shadow-blue-500/20">
            SP
          </div>
          <h1 className="font-semibold text-lg tracking-wide">
            {t("editor.header.title")}
          </h1>
        </div>

        <div className="flex items-center gap-2">
          {showLocaleSwitcher && (
            <button
              type="button"
              onClick={() => setLocale(locale === "es" ? "en" : "es")}
              aria-label={t("editor.header.languageAria")}
              title={
                locale === "es"
                  ? t("editor.header.switchToEnglish")
                  : t("editor.header.switchToSpanish")
              }
              className="relative flex items-center bg-slate-900 p-0.5 rounded-md border border-slate-700 cursor-pointer shadow-inner"
            >
              <span
                className={`absolute h-6 w-8 bg-slate-700 rounded shadow-sm transition-transform duration-300 ease-out ${
                  locale === "en" ? "translate-x-8" : "translate-x-0"
                }`}
              />
              <span
                className={`relative z-10 flex items-center justify-center w-8 h-6 text-[10px] font-bold transition-colors duration-300 select-none ${
                  locale === "es" ? "text-white" : "text-slate-500"
                }`}
              >
                ES
              </span>
              <span
                className={`relative z-10 flex items-center justify-center w-8 h-6 text-[10px] font-bold transition-colors duration-300 select-none ${
                  locale === "en" ? "text-white" : "text-slate-500"
                }`}
              >
                EN
              </span>
            </button>
          )}
          {showLocaleSwitcher && <div className="w-px h-5 bg-slate-700" />}
          <button
            onClick={() => setIsLibraryOpen(true)}
            className="flex items-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-md text-sm font-medium border border-slate-700 transition-colors"
          >
            <BookOpen size={16} /> {t("editor.header.library")}
          </button>
          {canOpenJsonModal && (
            <button
              onClick={() => setShowJsonModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-md text-sm font-medium shadow-md shadow-blue-900/20 transition-all"
            >
              <Code2 size={16} /> {t("editor.header.importJson")} / {t("editor.header.exportJson")}
            </button>
          )}
        </div>
      </header>

      <MachineGlobalPanel
        config={machineConfig}
        setConfig={(value) => {
          markDirtyFromImport();
          setMachineConfig(value);
        }}
        nodes={nodes}
        openBehaviorModal={setBehaviorModal}
        registry={registry}
        metadataPackRegistry={metadataPackRegistry}
        metadataPackBindings={metadataPackBindings}
        setMetadataPackBindings={(value) => {
          markDirtyFromImport();
          setMetadataPackBindings(value);
        }}
        machineIssues={issueIndex.machine}
      />

      <main
        ref={containerRef}
        className="w-full h-full relative bg-slate-950 overflow-hidden"
      >
        <TransformWrapper
          ref={transformRef}
          minScale={MIN_ZOOM}
          maxScale={MAX_ZOOM}
          centerOnInit={false}
          limitToBounds={false}
          smooth
          wheel={{ activationKeys: ["Control", "Meta"] }}
          panning={{
            disabled: false,
            wheelPanning: false,
            allowLeftClickPan: true,
            allowMiddleClickPan: true,
            allowRightClickPan: false,
            activationKeys: [],
            excluded: ["canvas-interactive"],
          }}
          pinch={{ disabled: true }}
          doubleClick={{ disabled: true }}
          onPanningStart={() => {
            panStartedRef.current = true;
            panMovedRef.current = false;
            setIsCanvasPanning(true);
          }}
          onPanning={() => {
            if (panStartedRef.current) {
              panMovedRef.current = true;
            }
          }}
          onPanningStop={() => {
            panStartedRef.current = false;
            setIsCanvasPanning(false);
          }}
          onInit={(ref) => setZoom(ref.state.scale)}
          onTransformed={(_ref, state) => setZoom(state.scale)}
        >
          <TransformComponent
            wrapperStyle={{ width: "100%", height: "100%" }}
            contentStyle={{ width: `${canvasSize.width}px`, height: `${canvasSize.height}px` }}
          >
            <div
              ref={canvasRef}
              data-testid="editor-canvas"
              className={`relative ${
                isCanvasPanning || draggingTransition
                  ? "cursor-grabbing"
                  : connectingStart
                    ? "cursor-alias"
                    : ""
              }`}
              style={{
                width: `${canvasSize.width}px`,
                height: `${canvasSize.height}px`,
                backgroundImage: "radial-gradient(circle, #334155 1px, transparent 1px)",
                backgroundSize: "20px 20px",
              }}
              onMouseDownCapture={handleCanvasMouseDownCapture}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onClick={handleCanvasClick}
            >
          <svg
            className="absolute top-0 left-0 w-full h-full pointer-events-none z-10"
          >
            <defs>
              <marker
                id="arrowhead"
                markerWidth="10"
                markerHeight="10"
                refX="9"
                refY="5"
                orient="auto"
                markerUnits="userSpaceOnUse"
              >
                <path d="M 0 0 L 10 5 L 0 10 z" fill="#94a3b8" />
              </marker>
              <marker
                id="arrowhead-selected"
                markerWidth="10"
                markerHeight="10"
                refX="9"
                refY="5"
                orient="auto"
                markerUnits="userSpaceOnUse"
              >
                <path d="M 0 0 L 10 5 L 0 10 z" fill="#3b82f6" />
              </marker>
            </defs>
            {transitions.map((transition) => {
              const legs = transitionLegsByTransitionId.get(transition.id) || [];
              const routeGeometry = transitionRouteGeometryByTransitionId.get(transition.id) || null;
              return (
                <TransitionRoute
                  key={`route-${transition.id}`}
                  transition={transition}
                  legs={legs}
                  nodes={nodes}
                  nodeSizes={nodeSizes}
                  routeGeometry={routeGeometry}
                  selected={selectedElement?.type === "transition" && selectedElement.id === transition.id}
                  invalidNotify={isInvalidNotifyTransition(transition, nodes)}
                  onSelect={() =>
                    setSelectedElement({
                      type: "transition",
                      id: transition.id,
                      data: transition,
                    })
                  }
                  onHover={(isHovered) =>
                    setHoveredTransitionId((prev) =>
                      isHovered ? transition.id : prev === transition.id ? null : prev,
                    )
                  }
                  onOutputPortMouseDown={(event) =>
                    startConnectionFromTransitionOutput(event, transition.id)
                  }
                />
              );
            })}

            {transitions.map((transition) => {
              const anchor = transitionBadgeAnchors.get(transition.id);
              if (!anchor) {
                return null;
              }

              const transitionIssueCount = issueIndex.transitions.get(transition.id)?.length || 0;
              const transitionOrderSummary = transitionOrderSummaryById.get(transition.id);

              return (
                <TransitionBadge
                  key={`badge-${transition.id}`}
                  x={anchor.x}
                  y={anchor.y}
                  transition={transition}
                  selected={selectedElement?.type === "transition" && selectedElement.id === transition.id}
                  alwaysOrderSummary={
                    transition.triggerKind === "always" && transitionOrderSummary
                      ? `${transitionOrderSummary.position} / ${transitionOrderSummary.total}`
                      : undefined
                  }
                  invalidNotify={isInvalidNotifyTransition(transition, nodes)}
                  issueCount={transitionIssueCount}
                  onSelect={() =>
                    setSelectedElement({
                      type: "transition",
                      id: transition.id,
                      data: transition,
                    })
                  }
                  onEdit={() => setIsModalOpen(true)}
                  onMouseDown={(event) => handleTransitionMouseDown(event, transition)}
                  onOutputPortMouseDown={(event) =>
                    startConnectionFromTransitionOutput(event, transition.id)
                  }
                  onHover={(isHovered) =>
                    setHoveredTransitionId((prev) =>
                      isHovered ? transition.id : prev === transition.id ? null : prev,
                    )
                  }
                />
              );
            })}

            {connectingStart &&
              (() => {
                const mX = mousePos.x;
                const mY = mousePos.y;
                const dist = Math.sqrt(
                  Math.pow(mX - connectingStart.startX, 2) +
                    Math.pow(mY - connectingStart.startY, 2),
                );
                const hLen = Math.max(dist / 2.5, 50);

                return (
                  <path
                    d={`M ${connectingStart.startX} ${connectingStart.startY} C ${connectingStart.startX + hLen} ${connectingStart.startY}, ${mX - hLen} ${mY}, ${mX} ${mY}`}
                    fill="none"
                    stroke="#94a3b8"
                    strokeWidth="2"
                    strokeDasharray="5,5"
                  />
                );
              })()}
          </svg>

          {nodes.map((node) => {
            if (node.type === "universe") {
              const universeIssueCount = issueIndex.universes.get(node.id)?.length || 0;
              const hasUniverseErrors = universeIssueCount > 0;

              return (
                <div
                  key={node.id}
                  data-testid={`universe-node-${node.id}`}
                  style={{ left: node.x, top: node.y, width: node.w, height: node.h }}
                  className={`canvas-interactive absolute border-2 border-dashed rounded-xl p-4 ${
                    selectedElement?.type === "universe" && selectedElement.id === node.id
                      ? `${hasUniverseErrors ? "border-red-500" : "border-blue-500"} bg-slate-900/40 z-0`
                      : `${hasUniverseErrors ? "border-red-700" : "border-slate-600"} bg-slate-900/20 z-0`
                  } transition-colors ${
                    searchPulseNodeId === node.id
                      ? searchPulseTick % 2 === 0
                        ? "studio-search-pulse-a"
                        : "studio-search-pulse-b"
                      : ""
                  }`}
                  onMouseDown={(e) => handleNodeMouseDown(e, node)}
                  onDoubleClick={(event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    setSelectedElement(node);
                    setIsModalOpen(true);
                  }}
                >
                  {hasUniverseErrors && (
                    <div className="absolute -top-2 -right-2 min-w-5 h-5 px-1 rounded-full bg-red-600 border-2 border-slate-900 flex items-center justify-center shadow-lg shadow-red-900/40 z-50 pointer-events-none">
                      <span className="text-[10px] font-bold text-white">{universeIssueCount}</span>
                    </div>
                  )}

                  <div className="absolute top-0 left-0 bg-slate-800 text-slate-300 text-xs px-3 py-1 rounded-br-lg rounded-tl-lg border-b border-r border-slate-600 flex items-center gap-2 pointer-events-none select-none">
                    <STUDIO_ICONS.entity.universe
                      size={14}
                      className={STUDIO_ICON_REGISTRY.entity.universe.colors.base}
                    />
                    <span className="font-mono font-semibold select-none" title={node.data.id || node.data.name}>
                      {node.data.id || node.data.name}
                    </span>
                  </div>

                  <div
                    data-testid={`universe-target-port-${node.id}`}
                    className="canvas-interactive group/port absolute top-1/2 -left-5 -translate-y-1/2 w-8 h-8 rounded-full cursor-alias active:cursor-alias z-40"
                    onMouseDown={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                    }}
                    onMouseUp={(e) => completeConnectionOnTargetPort(e, node.id)}
                  >
                    <div className="absolute inset-0 rounded-full bg-sky-400/0 border border-sky-300/0 group-hover/port:bg-sky-400/15 group-hover/port:border-sky-300/70 transition-colors pointer-events-none" />
                    <div className="absolute top-1/2 left-1/2 w-4 h-4 -translate-x-1/2 -translate-y-1/2 bg-slate-400 border-2 border-slate-800 rounded-full group-hover/port:bg-slate-300 transition-colors pointer-events-none" />
                  </div>

                  <div
                    className="absolute bottom-0 right-0 w-6 h-6 cursor-nwse-resize z-20 flex items-end justify-end p-1.5"
                    onMouseDown={(e) => handleResizeMouseDown(e, node)}
                  >
                    <div className="w-2.5 h-2.5 border-b-2 border-r-2 border-slate-400/50 rounded-sm hover:border-blue-500 transition-colors" />
                  </div>
                </div>
              );
            }

            if (node.type === "note") {
              return (
                <GlobalNoteNode
                  key={node.id}
                  node={node}
                  selected={selectedElement?.type === "note" && selectedElement.id === node.id}
                  onMouseDown={handleNodeMouseDown}
                  onUpdate={updateGlobalNote}
                  onDelete={(noteId) =>
                    deleteElement({
                      type: "note",
                      id: noteId,
                    })
                  }
                />
              );
            }

            return (
              <div
                key={node.id}
                data-testid={`reality-node-wrapper-${node.id}`}
                onDoubleClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  setSelectedElement(node);
                  setIsModalOpen(true);
                }}
              >
                <RealityNode
                  node={node}
                  selected={selectedElement?.type === "reality" && selectedElement.id === node.id}
                  searchPulseClassName={
                    searchPulseNodeId === node.id
                      ? searchPulseTick % 2 === 0
                        ? "studio-search-pulse-a"
                        : "studio-search-pulse-b"
                      : undefined
                  }
                  issueCount={issueIndex.realities.get(node.id)?.length || 0}
                  onMouseDown={handleNodeMouseDown}
                  onPortMouseDown={(e, nodeId) => startConnectionFromRealitySource(e, nodeId)}
                  onPortMouseUp={(e, nodeId) => completeConnectionOnTargetPort(e, nodeId)}
                  setNodeSizes={setNodeSizes}
                  onTypeChange={(id, type) => {
                    markDirtyFromImport();
                    setNodes((prev) =>
                      prev.map((n) =>
                        n.id === id && n.type === "reality"
                          ? { ...n, data: { ...n.data, realityType: type } }
                          : n,
                      ),
                    );
                  }}
                  onEdit={() => setIsModalOpen(true)}
                  onClone={cloneReality}
                  hasNote={Boolean(node.data.note)}
                  onCreateNote={(realityId) => updateNodeAnchoredNote(realityId, createAnchoredNote())}
                  onDelete={() => deleteElement(node)}
                />
              </div>
            );
          })}

          {nodes.map((node) => {
            if (
              node.type === "universe" &&
              selectedElement?.type === "universe" &&
              selectedElement.id === node.id
            ) {
              return (
                <div
                  key={`menu-${node.id}`}
                  style={{ left: node.x, top: node.y, width: node.w, height: node.h }}
                  className="absolute pointer-events-none z-[60]"
                >
                  <div className="canvas-interactive absolute -top-12 right-0 bg-slate-800 border border-slate-700 rounded-lg shadow-xl flex items-center p-1 gap-1 animate-in slide-in-from-bottom-2 fade-in duration-200 pointer-events-auto">
                    <div className="relative group flex items-center h-full">
                      <div className="px-3 py-1.5 hover:bg-green-500/20 rounded text-green-400 hover:text-green-300 transition-colors flex items-center gap-1.5 text-xs font-bold border border-green-500/30 cursor-default">
                        <PlusCircle size={14} /> {t("editor.context.addReality")}
                      </div>
                      <div className="absolute top-full left-0 pt-1 hidden group-hover:block z-[70]">
                        <div className="bg-slate-800 border border-slate-700 rounded shadow-xl p-1 gap-1 flex flex-col w-28">
                          {Object.entries(REALITY_TYPES).map(([typeKey, configType]) => {
                            const Icon = configType.icon;
                            return (
                              <button
                                key={typeKey}
                                onMouseDown={(e) => {
                                  e.stopPropagation();
                                  addRealityToUniverse(node.id, typeKey as keyof typeof REALITY_TYPES);
                                }}
                                className="flex items-center gap-2 p-1.5 hover:bg-slate-700 rounded transition-colors text-xs font-medium"
                              >
                                <Icon size={14} className={configType.color} />
                                <span className={configType.color}>
                                  {t(
                                    REALITY_TYPE_LABEL_KEYS[
                                      typeKey as keyof typeof REALITY_TYPE_LABEL_KEYS
                                    ] as string,
                                    undefined,
                                    configType.label,
                                  )}
                                </span>
                              </button>
                            );
                          })}
                        </div>
                      </div>
                    </div>

                    <div className="w-px h-4 bg-slate-700 mx-1"></div>

                    {!node.data.note && (
                      <>
                        <button
                          onMouseDown={(event) => {
                            event.preventDefault();
                            event.stopPropagation();
                            updateNodeAnchoredNote(node.id, createAnchoredNote());
                          }}
                          className="p-1.5 hover:bg-yellow-900/30 rounded text-yellow-500 hover:text-yellow-400 transition-colors"
                          title={t("editor.context.addNote")}
                        >
                          <StickyNote size={14} />
                        </button>
                        <div className="w-px h-4 bg-slate-700 mx-1"></div>
                      </>
                    )}

                    <button
                      onMouseDown={(e) => {
                        e.stopPropagation();
                        setIsModalOpen(true);
                      }}
                      className="p-1.5 hover:bg-slate-700 rounded text-slate-300 hover:text-white transition-colors"
                      title={t("editor.context.editProperties")}
                    >
                      <Settings2 size={14} />
                    </button>
                    <div className="w-px h-4 bg-slate-700 mx-1"></div>
                    <button
                      onMouseDown={(e) => {
                        e.stopPropagation();
                        fitUniverseContent(node.id);
                      }}
                      className="p-1.5 hover:bg-slate-700 rounded text-slate-300 hover:text-white transition-colors"
                      aria-label={t("editor.context.fitContent")}
                      title={t("editor.context.fitContent")}
                    >
                      <Maximize size={14} />
                    </button>
                    <div className="w-px h-4 bg-slate-700 mx-1"></div>
                    <button
                      onMouseDown={(e) => {
                        e.stopPropagation();
                        cloneUniverse(node.id);
                      }}
                      className="p-1.5 hover:bg-slate-700 rounded text-slate-300 hover:text-white transition-colors"
                      aria-label={t("editor.context.cloneUniverse")}
                      title={t("editor.context.cloneUniverse")}
                    >
                      <Copy size={14} />
                    </button>
                    <div className="w-px h-4 bg-slate-700 mx-1"></div>
                    <button
                      onMouseDown={(e) => {
                        e.stopPropagation();
                        deleteElement(node);
                      }}
                      className="p-1.5 hover:bg-red-900/50 rounded text-red-400 hover:text-red-300 transition-colors"
                      title={t("editor.context.deleteUniverse")}
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              );
            }
            return null;
          })}

          {selectedElement?.type === "transition" &&
            (() => {
              const anchor = transitionBadgeAnchors.get(selectedElement.id);
              if (!anchor) return null;

              return (
                <div
                  key={`menu-transition-${selectedElement.id}`}
                  className="absolute z-[60] pointer-events-none"
                  style={{ left: anchor.x, top: anchor.y - 30 }}
                >
                  <div className="canvas-interactive absolute -top-12 left-1/2 -translate-x-1/2 bg-slate-800 border border-slate-700 rounded-lg shadow-xl flex items-center p-1 gap-1 animate-in slide-in-from-bottom-2 fade-in duration-200 pointer-events-auto">
                    {!selectedElement.data.note && (
                      <>
                        <button
                          onMouseDown={(event) => {
                            event.preventDefault();
                            event.stopPropagation();
                            updateTransitionNote(selectedElement.id, createAnchoredNote());
                          }}
                          className="p-1.5 hover:bg-yellow-900/30 rounded text-yellow-500 hover:text-yellow-400 transition-colors"
                          title={t("editor.context.addNote")}
                        >
                          <StickyNote size={14} />
                        </button>
                        <div className="w-px h-4 bg-slate-700 mx-1"></div>
                      </>
                    )}

                    <button
                      onMouseDown={(e) => {
                        e.stopPropagation();
                        setIsModalOpen(true);
                      }}
                      className="p-1.5 hover:bg-slate-700 rounded text-slate-300 hover:text-white transition-colors"
                      title={t("editor.context.editTransition")}
                    >
                      <Settings2 size={14} />
                    </button>
                    <div className="w-px h-4 bg-slate-700 mx-1"></div>
                    <button
                      onMouseDown={(e) => {
                        e.stopPropagation();
                        deleteElement(selectedElement);
                      }}
                      className="p-1.5 hover:bg-red-900/50 rounded text-red-400 hover:text-red-300 transition-colors"
                      title={t("editor.context.deleteTransition")}
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              );
            })()}

          {nodes.map((node) => {
            if (node.type === "note" || !node.data.note) {
              return null;
            }

            const isFocused =
              selectedElement?.type !== "transition" && selectedElement?.id === node.id;

            let x = node.x;
            let y = node.y;

            if (node.type === "universe") {
              x = node.x + (node.w || 400) - 40;
              y = node.y - 12;
            } else {
              const realityWidth = nodeSizes[node.id]?.w || DEFAULT_REALITY_WIDTH;
              x = node.x + realityWidth - 12;
              y = node.y - 12;
            }

            return (
              <div
                key={`note-node-${node.id}`}
                className="absolute z-[80] pointer-events-auto"
                style={{ left: x, top: y }}
              >
                <AnchoredNote
                  note={node.data.note}
                  isFocused={Boolean(isFocused)}
                  onRequestFocus={() => setSelectedElement(node)}
                  onUpdate={(nextNote) => updateNodeAnchoredNote(node.id, nextNote)}
                  onDelete={() => updateNodeAnchoredNote(node.id, null)}
                />
              </div>
            );
          })}

          {transitions.map((transition) => {
            if (!transition.note) {
              return null;
            }

            const anchor = transitionBadgeAnchors.get(transition.id);
            if (!anchor) {
              return null;
            }

            const isFocused =
              selectedElement?.type === "transition" && selectedElement.id === transition.id;

            return (
              <div
                key={`note-transition-${transition.id}`}
                className="absolute z-[80] pointer-events-auto"
                style={{ left: anchor.x + 60, top: anchor.y - 20 }}
              >
                <AnchoredNote
                  note={transition.note}
                  isFocused={isFocused}
                  onRequestFocus={() =>
                    setSelectedElement({
                      type: "transition",
                      id: transition.id,
                      data: transition,
                    })
                  }
                  onUpdate={(nextNote) => updateTransitionNote(transition.id, nextNote)}
                  onDelete={() => updateTransitionNote(transition.id, null)}
                />
              </div>
            );
          })}

          {hoveredTransitionId && (() => {
            const transition = transitionsById.get(hoveredTransitionId);
            const anchor = transition ? transitionBadgeAnchors.get(transition.id) : null;
            if (!transition || !anchor) return null;

            const isAlways = transition.triggerKind === "always";
            const triggerName = isAlways
              ? t("editor.transition.trigger.always")
              : transition.eventName || t("editor.transition.trigger.newEvent");
            const invalidNotify = isInvalidNotifyTransition(transition, nodes);
            const effectiveType = invalidNotify ? "default" : transition.type;
            const isNotify = effectiveType === "notify";
            const hasEffects = transition.actions.length > 0 || transition.invokes.length > 0;
            const hasConditions = transition.conditions.length > 0;
            const transitionOrderSummary = transitionOrderSummaryById.get(transition.id);
            const transitionOrderLabel = transitionOrderSummary
              ? `${transitionOrderSummary.position} / ${transitionOrderSummary.total}`
              : "- / 1";
            const TriggerIcon = isAlways
              ? STUDIO_ICONS.transition.trigger.always
              : STUDIO_ICONS.transition.trigger.on;
            const TypeIcon = isNotify
              ? STUDIO_ICONS.transition.type.notify
              : STUDIO_ICONS.transition.type.default;
            const ConditionIcon = STUDIO_ICONS.behavior.condition;
            const ActionIcon = STUDIO_ICONS.behavior.action;
            const InvokeIcon = STUDIO_ICONS.behavior.invoke;
            const triggerOnColors = STUDIO_ICON_REGISTRY.transition.trigger.on.colors;
            const triggerAlwaysColors = STUDIO_ICON_REGISTRY.transition.trigger.always.colors;
            const transitionDefaultColors = STUDIO_ICON_REGISTRY.transition.type.default.colors;
            const transitionNotifyColors = STUDIO_ICON_REGISTRY.transition.type.notify.colors;
            const conditionColors = STUDIO_ICON_REGISTRY.behavior.condition.colors;
            const actionColors = STUDIO_ICON_REGISTRY.behavior.action.colors;
            const invokeColors = STUDIO_ICON_REGISTRY.behavior.invoke.colors;

            return (
              <div
                key={`tooltip-${transition.id}`}
                className="absolute z-[70] pointer-events-none"
                style={{ left: anchor.x, top: anchor.y + 18 }}
              >
                <div className="absolute top-0 left-1/2 -translate-x-1/2 bg-slate-900 border border-slate-700 rounded-lg shadow-2xl p-3 w-56 flex flex-col gap-2 animate-in slide-in-from-top-2 fade-in duration-200">
                  <div className="flex items-center gap-1.5 border-b border-slate-800 pb-2 mb-1">
                    <Info size={14} className="text-blue-400" />
                    <span className="text-xs font-bold text-slate-200">
                      {t("editor.transition.details")}
                    </span>
                  </div>

                  <div className="grid grid-cols-2 gap-x-2 gap-y-1.5 text-[10px]">
                    <div className="text-slate-400 flex items-center gap-1">
                      <TriggerIcon
                        size={10}
                        className={isAlways ? triggerAlwaysColors.base : (triggerOnColors.accent ?? triggerOnColors.base)}
                      />
                      {t("editor.transition.details.trigger")}:
                    </div>
                    <div
                      className={`font-mono font-bold truncate text-right ${isAlways ? triggerAlwaysColors.base : (triggerOnColors.emphasis ?? triggerOnColors.base)}`}
                    >
                      {triggerName}
                    </div>

                    <div className="text-slate-400 flex items-center gap-1">
                      <TypeIcon
                        size={10}
                        className={
                          isNotify
                            ? transitionNotifyColors.base
                            : (transitionDefaultColors.emphasis ?? transitionDefaultColors.base)
                        }
                      />
                      {t("editor.transition.details.type")}:
                    </div>
                    <div
                      className={`font-bold text-right uppercase ${isNotify ? transitionNotifyColors.base : (transitionDefaultColors.emphasis ?? transitionDefaultColors.base)}`}
                    >
                      {effectiveType}
                    </div>

                    {isAlways && (
                      <>
                        <div className="text-slate-400">
                          {t("properties.transition.priority")}:
                        </div>
                        <div className="font-mono font-bold text-right text-slate-200">
                          {transitionOrderLabel}
                        </div>
                      </>
                    )}
                  </div>

                  {(hasConditions || hasEffects) && <div className="w-full h-px bg-slate-800 my-0.5"></div>}

                  <div className="flex justify-between items-center text-[10px] text-slate-400">
                    <span className="flex items-center gap-1">
                      <ConditionIcon size={10} className={conditionColors.base} />{" "}
                      {t("editor.transition.details.conditions")}:
                    </span>
                    <span className="text-slate-200 font-mono font-bold">
                      {transition.conditions.length}
                    </span>
                  </div>
                  <div className="flex justify-between items-center text-[10px] text-slate-400">
                    <span className="flex items-center gap-1">
                      <ActionIcon size={10} className={actionColors.base} /> {t("editor.transition.details.actions")}:
                    </span>
                    <span className="text-slate-200 font-mono font-bold">{transition.actions.length}</span>
                  </div>
                  <div className="flex justify-between items-center text-[10px] text-slate-400">
                    <span className="flex items-center gap-1">
                      <InvokeIcon size={10} className={invokeColors.base} /> {t("editor.transition.details.invokes")}:
                    </span>
                    <span className="text-slate-200 font-mono font-bold">{transition.invokes.length}</span>
                  </div>

                  {transition.description && (
                    <div className="mt-1 pt-2 border-t border-slate-800 text-[10px] text-slate-400 italic leading-relaxed">
                      "{transition.description}"
                    </div>
                  )}
                  <div className="mt-1 text-[8px] text-slate-500 text-center font-medium">
                    {t("editor.transition.details.doubleClickHint")}
                  </div>
                </div>
              </div>
            );
          })()}
            </div>
          </TransformComponent>
        </TransformWrapper>
      </main>

      <CanvasToolbar
        onAddUniverse={addUniverse}
        onAddUniverseFromTemplate={addUniverseFromTemplate}
        universeTemplates={universeTemplates}
        onAddGlobalNote={addGlobalNote}
        onAutoLayout={handleAutoLayout}
        isAutoLayouting={isAutoLayouting}
        onUndo={undo}
        onRedo={redo}
        canUndo={canUndo}
        canRedo={canRedo}
        onZoomIn={handleZoomIn}
        onZoomOut={handleZoomOut}
        onFit={() => handleFitToContent("smooth")}
        onCenter={() => handleCenterContent("smooth")}
        zoom={zoom}
        searchQuery={searchQuery}
        searchResults={searchResults}
        searchFilters={searchFilters}
        searchActiveIndex={searchActiveIndex}
        onSearchQueryChange={(query) => {
          setSearchQuery(query);
          if (!query.trim()) {
            setSearchActiveIndex(-1);
          }
        }}
        onSearchFiltersChange={setSearchFilters}
        onSearchMoveSelection={moveSearchSelection}
        onSearchSelect={selectSearchResult}
      />

      {isModalOpen && propertiesModalElement && (
        <PropertiesModal
          element={propertiesModalElement}
          nodes={nodes}
          transitions={transitions}
          onClose={() => setIsModalOpen(false)}
          updateNodeData={updateModalNodeData}
          commitUniverseIdRename={commitUniverseIdRename}
          commitUniverseCanonicalRename={commitUniverseCanonicalRename}
          commitRealityIdRename={commitRealityIdRename}
          updateTransitionData={updateModalTransitionData}
          updateTransitionPatch={updateModalTransitionPatch}
          moveTransition={moveSelectedTransition}
          openBehaviorModal={setBehaviorModal}
          registry={registry}
          metadataPackRegistry={metadataPackRegistry}
          metadataPackBindings={metadataPackBindings}
          setMetadataPackBindings={(value) => {
            markDirtyFromImport();
            setMetadataPackBindings(value);
          }}
          issues={selectedElementIssues}
        />
      )}

      <BehaviorModal
        isOpen={behaviorModal.isOpen}
        type={behaviorModal.type}
        initialData={behaviorModal.initialData}
        onSave={behaviorModal.onSave}
        onClose={() =>
          setBehaviorModal((previous) => ({
            ...previous,
            isOpen: false,
          }))
        }
        registry={registry}
      />

      <LibraryModal
        isOpen={isLibraryOpen}
        onClose={() => setIsLibraryOpen(false)}
        canManageBehaviors={features.library.behaviors.manage}
        canCreateMetadataPacks={features.library.metadataPacks.create}
        registry={registry}
        setRegistry={setRegistry}
        resolveUsage={resolveRegistryBehaviorUsage}
        onDeleteBehavior={handleDeleteRegistryBehavior}
        metadataPackRegistry={metadataPackRegistry}
        setMetadataPackRegistry={(value) => {
          markDirtyFromImport();
          setMetadataPackRegistry(value);
        }}
        onDeleteMetadataPack={handleDeleteMetadataPack}
        onRenameMetadataPack={handleRenameMetadataPack}
      />

      <JsonIOModal
        isOpen={showJsonModal}
        onClose={() => setShowJsonModal(false)}
        allowImport={features.json.import}
        allowExport={features.json.export}
        modelJson={generatedJson}
        layoutJson={generatedLayoutJson}
        modelIssues={serialized.issues}
        canExportModel={serialized.canExport}
        onImportModel={handleImportMachine}
        onImportLayout={handleImportLayout}
      />
    </div>
  );
}
