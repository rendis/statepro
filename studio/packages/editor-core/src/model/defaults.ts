import {
  STUDIO_DEFAULT_LOCALE,
  normalizeStudioLocale,
  translateStudioMessage,
  type StudioLocale,
} from "../i18n";
import { buildBuiltinBehaviorRegistry } from "./behaviorRegistry";
import type {
  EditorState,
  MachineConfig,
  EditorNode,
  BehaviorRegistryItem,
  UniversalConstants,
  EditorTransition,
  MetadataPackBindingMap,
} from "../types";

export const createEmptyUniversalConstants = (): UniversalConstants => ({
  entryActions: [],
  exitActions: [],
  entryInvokes: [],
  exitInvokes: [],
  actionsOnTransition: [],
  invokesOnTransition: [],
});

const tDefaults = (locale: StudioLocale, key: string, fallback: string): string =>
  translateStudioMessage(locale, key, undefined, fallback);

export const buildInitialNodes = (locale: StudioLocale = STUDIO_DEFAULT_LOCALE): EditorNode[] => [
  {
    id: "univ-1",
    type: "universe",
    x: 1050,
    y: 1050,
    w: 500,
    h: 400,
    data: {
      id: "main-universe",
      name: "main-universe",
      canonicalName: "main-universe",
      version: "1.0.0",
      description: "",
      tags: [],
      metadata: "{}",
      universalConstants: createEmptyUniversalConstants(),
    },
  },
  {
    id: "real-1",
    type: "reality",
    x: 1100,
    y: 1120,
    data: {
      id: "idle",
      name: "idle",
      universeId: "univ-1",
      isInitial: true,
      realityType: "normal",
      description: "",
      entryActions: [],
      exitActions: [],
      observers: [],
      entryInvokes: [],
      exitInvokes: [],
    },
  },
  {
    id: "real-2",
    type: "reality",
    x: 1300,
    y: 1250,
    data: {
      id: "processing",
      name: "processing",
      universeId: "univ-1",
      isInitial: false,
      realityType: "success",
      description: "",
      entryActions: [{ src: "builtin:action:logArgs" }],
      exitActions: [],
      observers: [],
      entryInvokes: [],
      exitInvokes: [],
    },
  },
];

export const initialNodes: EditorNode[] = buildInitialNodes();

export const initialTransitions: EditorTransition[] = [
  {
    id: "tr-1",
    sourceRealityId: "real-1",
    triggerKind: "on",
    eventName: "START_PROCESS",
    type: "default",
    conditions: [],
    actions: [],
    invokes: [],
    description: "",
    metadata: "",
    targets: ["processing"],
    order: 0,
  },
];

export const buildDefaultRegistry = (
  locale: StudioLocale = STUDIO_DEFAULT_LOCALE,
): BehaviorRegistryItem[] => buildBuiltinBehaviorRegistry(locale);

export const defaultRegistry: BehaviorRegistryItem[] = buildDefaultRegistry();

export const buildDefaultMachineConfig = (
  locale: StudioLocale = STUDIO_DEFAULT_LOCALE,
): MachineConfig => ({
  id: "admission-process-machine",
  canonicalName: "admission-process",
  version: "1.0.0",
  description: tDefaults(
    locale,
    "defaults.machine.description",
    "Main orchestrator for user admissions.",
  ),
  initials: ["U:main-universe"],
  universalConstants: createEmptyUniversalConstants(),
  metadata: '{\n  "author": "Rendis",\n  "environment": "production"\n}',
});

export const defaultMachineConfig: MachineConfig = buildDefaultMachineConfig();

export const createInitialMetadataPackBindingMap = (): MetadataPackBindingMap => ({
  machine: [],
  universe: [],
  reality: [],
  transition: [],
});

export const createInitialEditorState = (locale?: StudioLocale): EditorState => {
  const resolvedLocale = normalizeStudioLocale(locale);

  return {
    nodes: structuredClone(buildInitialNodes(resolvedLocale)),
    transitions: structuredClone(initialTransitions),
    nodeSizes: {},
    selectedElement: null,
    machineConfig: structuredClone(buildDefaultMachineConfig(resolvedLocale)),
    registry: structuredClone(buildDefaultRegistry(resolvedLocale)),
    metadataPackRegistry: [],
    metadataPackBindings: createInitialMetadataPackBindingMap(),
    lastImportedMachine: undefined,
    isDirtyFromImport: false,
  };
};
