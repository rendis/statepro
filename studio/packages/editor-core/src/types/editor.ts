import type { Dispatch, SetStateAction } from "react";

export type RealityType = "normal" | "success" | "error";
export type BehaviorType = "action" | "invoke" | "condition" | "observer";
export type TransitionType = "default" | "notify";
export type TransitionTriggerKind = "always" | "on";

export type JsonValue =
  | string
  | number
  | boolean
  | null
  | { [key: string]: JsonValue }
  | JsonValue[];

export type JsonObject = Record<string, JsonValue>;

export type MetadataScope = "machine" | "universe" | "reality" | "transition";

export type MetadataPackWidget =
  | "text"
  | "textarea"
  | "number"
  | "boolean"
  | "select"
  | "array"
  | "json";

export interface MetadataPackSelectOption {
  value: string;
  label: string;
}

export interface MetadataPackUiOverride {
  widget?: MetadataPackWidget;
  options?: MetadataPackSelectOption[];
  allowCustomValue?: boolean;
  constant?: boolean;
  constantValue?: JsonValue;
  placeholder?: string;
  help?: string;
  order?: number;
  title?: string;
}

export type MetadataPackUiMap = Record<string, MetadataPackUiOverride>;

export interface MetadataPackDefinition {
  id: string;
  label: string;
  description?: string;
  scopes: MetadataScope[];
  schema: JsonObject;
  ui?: MetadataPackUiMap;
}

export type MetadataPackRegistry = MetadataPackDefinition[];

export interface MetadataPackBinding {
  id: string;
  packId: string;
  scope: MetadataScope;
  entityRef: string;
  values: JsonObject;
}

export type MetadataPackBindingMap = Record<MetadataScope, MetadataPackBinding[]>;

export interface StudioMetadataEnvelope {
  version: 1;
  packRegistry: MetadataPackRegistry;
  bindings: MetadataPackBindingMap;
}

export interface BehaviorRef {
  src: string;
  args?: JsonObject;
  description?: string;
  metadata?: JsonObject;
}

export interface BehaviorRegistryItem {
  src: string;
  type: BehaviorType;
  description?: string;
  simScript: string;
}

export interface UniversalConstants {
  entryActions: BehaviorRef[];
  exitActions: BehaviorRef[];
  entryInvokes: BehaviorRef[];
  exitInvokes: BehaviorRef[];
  actionsOnTransition: BehaviorRef[];
  invokesOnTransition: BehaviorRef[];
}

export interface AnchoredNoteData {
  text: string;
  colorIndex: number;
}

export interface GlobalNoteData extends AnchoredNoteData {
  isCollapsed?: boolean;
}

export interface UniverseNodeData {
  id: string;
  name: string;
  canonicalName: string;
  version: string;
  description?: string;
  tags?: string[];
  metadata?: string;
  note?: AnchoredNoteData | null;
  universalConstants?: UniversalConstants;
}

export interface RealityNodeData {
  id: string;
  name: string;
  universeId: string;
  isInitial: boolean;
  realityType: RealityType;
  description?: string;
  metadata?: string;
  note?: AnchoredNoteData | null;
  observers?: BehaviorRef[];
  entryActions?: BehaviorRef[];
  exitActions?: BehaviorRef[];
  entryInvokes?: BehaviorRef[];
  exitInvokes?: BehaviorRef[];
}

export interface UniverseNode {
  id: string;
  type: "universe";
  x: number;
  y: number;
  w: number;
  h: number;
  data: UniverseNodeData;
}

export interface RealityNode {
  id: string;
  type: "reality";
  x: number;
  y: number;
  data: RealityNodeData;
}

export interface NoteNode {
  id: string;
  type: "note";
  x: number;
  y: number;
  data: GlobalNoteData;
}

export type EditorNode = UniverseNode | RealityNode | NoteNode;

// Canonical transition representation (source of truth)
export interface EditorTransition {
  id: string;
  sourceRealityId: string;
  triggerKind: TransitionTriggerKind;
  eventName?: string;
  type: TransitionType;
  condition?: BehaviorRef;
  conditions: BehaviorRef[];
  actions: BehaviorRef[];
  invokes: BehaviorRef[];
  description?: string;
  metadata?: string;
  note?: AnchoredNoteData | null;
  targets: string[];
  order: number;
  visualOffset?: {
    x: number;
    y: number;
  };
}

// Derived visual leg for each transition target.
export interface TransitionLeg {
  id: string;
  transitionId: string;
  source: string;
  target: string;
  targetRef: string;
}

export interface MachineConfig {
  id: string;
  canonicalName: string;
  version: string;
  description?: string;
  initials: string[];
  universalConstants: UniversalConstants;
  metadata: string;
}

export interface NodeSize {
  w: number;
  h: number;
}

export type NodeSizeMap = Record<string, NodeSize>;

export type SelectedElementRef =
  | { kind: "node"; id: string }
  | { kind: "transition"; id: string };

export interface EditorState {
  nodes: EditorNode[];
  transitions: EditorTransition[];
  nodeSizes: NodeSizeMap;
  selectedElement: SelectedElementRef | null;
  machineConfig: MachineConfig;
  registry: BehaviorRegistryItem[];
  metadataPackRegistry: MetadataPackRegistry;
  metadataPackBindings: MetadataPackBindingMap;
  lastImportedMachine?: StateProMachine;
  isDirtyFromImport: boolean;
}

export interface BehaviorModalState {
  isOpen: boolean;
  type: BehaviorType;
  initialData: BehaviorRef | null;
  onSave: ((item: BehaviorRef) => void) | null;
}

export interface DragStartInfo {
  mouseX: number;
  mouseY: number;
  nodeStartX: number;
  nodeStartY: number;
  nodeStartW?: number;
  nodeStartH?: number;
  children: Array<{ id: string; startX: number; startY: number }>;
  childStartById?: Record<string, { startX: number; startY: number }>;
  parentUniverse: UniverseNode | null;
  otherUniverseBounds?: Array<{ id: string; x: number; y: number; w: number; h: number }>;
  otherRealityBounds?: Array<{ id: string; x: number; y: number; w: number; h: number }>;
  nodeSizeById?: Record<string, { w: number; h: number }>;
  isUniverse: boolean;
  isNote?: boolean;
}

export interface ResizingStart {
  id: string;
  mouseX: number;
  mouseY: number;
  startW: number;
  startH: number;
  nodeX: number;
  nodeY: number;
  realityBounds?: Array<{ id: string; x: number; y: number; w: number; h: number }>;
  otherUniverseBounds?: Array<{ id: string; x: number; y: number; w: number; h: number }>;
}

export type ConnectingStart =
  | {
      kind: "reality";
      sourceRealityId: string;
      startX: number;
      startY: number;
    }
  | {
      kind: "transition";
      transitionId: string;
      sourceRealityId: string;
      startX: number;
      startY: number;
    };

export type SetNodeSizes = Dispatch<SetStateAction<NodeSizeMap>>;

export type RealityNodeTypeConfig = {
  icon: React.ComponentType<{ size?: string | number; className?: string }>;
  color: string;
  border: string;
  bg: string;
  label: string;
};

export type BehaviorTypeConfig = {
  icon: React.ComponentType<{ size?: string | number; className?: string }>;
  color: string;
  border: string;
  bg: string;
  label: string;
};

export interface SerializeIssue {
  code: "INVALID_JSON" | "SCHEMA_ERROR" | "SEMANTIC_ERROR";
  severity: "error" | "warning";
  field: string;
  message: string;
  messageKey?: string;
  messageParams?: Record<string, string | number>;
}

export interface SerializeResult {
  machine: StateProMachine;
  issues: SerializeIssue[];
  canExport: boolean;
}

export interface StudioLayoutMachineRef {
  id: string;
  canonicalName: string;
  version: string;
}

export interface StudioPacksSnapshot {
  packRegistry: MetadataPackRegistry;
  bindings: MetadataPackBindingMap;
}

export interface StudioLayoutDocument {
  version: 1;
  machineRef: StudioLayoutMachineRef;
  nodes: {
    universes: Array<{
      entityRef: string;
      x: number;
      y: number;
      w: number;
      h: number;
      note: AnchoredNoteData | null;
    }>;
    realities: Array<{
      entityRef: string;
      x: number;
      y: number;
      note: AnchoredNoteData | null;
    }>;
    globalNotes: Array<{
      x: number;
      y: number;
      data: GlobalNoteData;
    }>;
  };
  transitions: Array<{
    entityRef: string;
    visualOffset?: {
      x: number;
      y: number;
    };
    note: AnchoredNoteData | null;
  }>;
  packs: StudioPacksSnapshot;
}

export interface StudioLayoutIssue {
  severity: "error" | "warning";
  field: string;
  message: string;
}

export interface ParseStudioLayoutResult {
  document: StudioLayoutDocument | null;
  issues: StudioLayoutIssue[];
  canImport: boolean;
}

export interface ApplyStudioLayoutResult {
  state: EditorState;
  issues: StudioLayoutIssue[];
}

export interface StateProTransition {
  targets: string[];
  type?: TransitionType;
  description?: string;
  condition?: BehaviorRef;
  conditions?: BehaviorRef[];
  actions?: BehaviorRef[];
  invokes?: BehaviorRef[];
  metadata?: JsonObject;
}

export interface StateProReality {
  id: string;
  type: "transition" | "final" | "unsuccessfulFinal";
  description?: string;
  metadata?: JsonObject;
  observers?: BehaviorRef[];
  entryActions?: BehaviorRef[];
  exitActions?: BehaviorRef[];
  entryInvokes?: BehaviorRef[];
  exitInvokes?: BehaviorRef[];
  always?: StateProTransition[];
  on?: Record<string, StateProTransition[]> | null;
}

export interface StateProUniverse {
  id: string;
  canonicalName: string;
  version: string;
  description?: string;
  tags?: string[];
  metadata?: JsonObject;
  universalConstants?: Partial<UniversalConstants>;
  initial?: string;
  realities: Record<string, StateProReality>;
}

export interface StateProMachine {
  id: string;
  canonicalName: string;
  version: string;
  description?: string;
  initials?: string[];
  universalConstants?: Partial<UniversalConstants>;
  metadata?: JsonObject;
  universes: Record<string, StateProUniverse>;
}

export interface StudioExternalMetadataPacks {
  registry: MetadataPackRegistry;
  bindings: MetadataPackBindingMap;
}

export interface StudioExternalValue {
  definition: StateProMachine;
  layout?: StudioLayoutDocument;
  metadataPacks?: StudioExternalMetadataPacks;
}

export interface StudioUniverseTemplate {
  id: string;
  label: string;
  description?: string;
  universe: StateProUniverse;
}

export type StudioPerformanceMode = "auto" | "off" | "aggressive";

export interface StudioPerformanceFeatureFlags {
  mode?: StudioPerformanceMode;
  staticPressureThreshold?: number;
  onEmaMs?: number;
  offEmaMs?: number;
  onMissRatio?: number;
  offMissRatio?: number;
}

export interface StudioFeatureFlags {
  json?: {
    import?: boolean;
    export?: boolean;
  };
  library?: {
    behaviors?: {
      manage?: boolean;
    };
    metadataPacks?: {
      create?: boolean;
    };
  };
  performance?: StudioPerformanceFeatureFlags;
}

export interface StudioChangePayload {
  machine: StateProMachine;
  layout: StudioLayoutDocument;
  issues: SerializeIssue[];
  canExport: boolean;
  source: "user" | "external-sync";
  at: string;
}
