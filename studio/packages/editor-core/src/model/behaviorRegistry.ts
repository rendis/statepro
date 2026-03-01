import {
  STUDIO_DEFAULT_LOCALE,
  createStudioTranslator,
  type StudioLocale,
} from "../i18n";
import type { BehaviorRegistryItem, BehaviorType } from "../types";
import { getBehaviorScriptContract } from "../utils";
import { BUILTIN_BEHAVIOR_CATALOG } from "./generatedBuiltinBehaviorCatalog";

const VALID_BEHAVIOR_TYPES: ReadonlySet<BehaviorType> = new Set([
  "action",
  "invoke",
  "condition",
  "observer",
]);

const BUILTIN_BEHAVIOR_SOURCE_SET: Set<string> = new Set(
  BUILTIN_BEHAVIOR_CATALOG.map((item) => item.src),
);

export type BehaviorRegistrySource = "builtin" | "external" | "user";

interface ComposeBehaviorRegistryOptions {
  locale?: StudioLocale;
  currentRegistry?: BehaviorRegistryItem[];
  externalRegistry?: BehaviorRegistryItem[];
  previousExternalSources?: Iterable<string>;
  preferExternalForExternalSources?: boolean;
}

const isBehaviorType = (value: unknown): value is BehaviorType =>
  typeof value === "string" &&
  VALID_BEHAVIOR_TYPES.has(value as BehaviorType);

const buildDefaultScriptsByType = (
  locale: StudioLocale,
): Record<BehaviorType, string> => {
  const t = createStudioTranslator(locale);

  return {
    action: getBehaviorScriptContract("action", t).defaultScript,
    invoke: getBehaviorScriptContract("invoke", t).defaultScript,
    condition: getBehaviorScriptContract("condition", t).defaultScript,
    observer: getBehaviorScriptContract("observer", t).defaultScript,
  };
};

const sanitizeRegistryItem = (
  item: BehaviorRegistryItem,
  defaultScriptsByType: Record<BehaviorType, string>,
): BehaviorRegistryItem | null => {
  if (!item || typeof item.src !== "string") {
    return null;
  }
  const src = item.src.trim();
  if (!src || !isBehaviorType(item.type)) {
    return null;
  }

  return {
    src,
    type: item.type,
    description: typeof item.description === "string" ? item.description : undefined,
    simScript:
      typeof item.simScript === "string" && item.simScript.trim()
        ? item.simScript
        : defaultScriptsByType[item.type],
  };
};

const toRegistryMap = (
  items: BehaviorRegistryItem[] | undefined,
  defaultScriptsByType: Record<BehaviorType, string>,
): Map<string, BehaviorRegistryItem> => {
  const map = new Map<string, BehaviorRegistryItem>();
  (items || []).forEach((item) => {
    const normalized = sanitizeRegistryItem(item, defaultScriptsByType);
    if (!normalized) {
      return;
    }
    map.set(normalized.src, normalized);
  });
  return map;
};

export const isBuiltinBehaviorSource = (src: string): boolean =>
  BUILTIN_BEHAVIOR_SOURCE_SET.has(src);

export const getBuiltinBehaviorSources = (): string[] =>
  Array.from(BUILTIN_BEHAVIOR_SOURCE_SET);

export const buildBuiltinBehaviorRegistry = (
  locale: StudioLocale = STUDIO_DEFAULT_LOCALE,
): BehaviorRegistryItem[] => {
  const t = createStudioTranslator(locale);
  const defaultScriptsByType = buildDefaultScriptsByType(locale);

  return BUILTIN_BEHAVIOR_CATALOG.map((entry) => ({
    src: entry.src,
    type: entry.type,
    description: t(entry.descriptionKey, undefined, entry.descriptionFallback),
    simScript: defaultScriptsByType[entry.type],
  }));
};

export const composeBehaviorRegistry = ({
  locale = STUDIO_DEFAULT_LOCALE,
  currentRegistry = [],
  externalRegistry = [],
  previousExternalSources,
  preferExternalForExternalSources = false,
}: ComposeBehaviorRegistryOptions): BehaviorRegistryItem[] => {
  const defaultScriptsByType = buildDefaultScriptsByType(locale);
  const currentBySrc = toRegistryMap(currentRegistry, defaultScriptsByType);
  const externalBySrc = toRegistryMap(externalRegistry, defaultScriptsByType);
  const builtinRegistry = buildBuiltinBehaviorRegistry(locale);
  const previousExternalSrcSet = new Set(previousExternalSources || []);
  const nextRegistry: BehaviorRegistryItem[] = [];
  const seen = new Set<string>();

  // Builtins are always present and keep official type/description.
  builtinRegistry.forEach((builtinItem) => {
    const currentItem = currentBySrc.get(builtinItem.src);
    const externalItem = externalBySrc.get(builtinItem.src);
    const simScript =
      currentItem?.simScript ||
      externalItem?.simScript ||
      builtinItem.simScript;

    nextRegistry.push({
      src: builtinItem.src,
      type: builtinItem.type,
      description: builtinItem.description,
      simScript,
    });
    seen.add(builtinItem.src);
  });

  externalBySrc.forEach((externalItem) => {
    if (isBuiltinBehaviorSource(externalItem.src)) {
      return;
    }
    const currentItem = currentBySrc.get(externalItem.src);
    const selected =
      preferExternalForExternalSources || !currentItem
        ? externalItem
        : currentItem;
    if (seen.has(selected.src)) {
      return;
    }
    nextRegistry.push(selected);
    seen.add(selected.src);
  });

  currentRegistry.forEach((currentItem) => {
    const normalized = sanitizeRegistryItem(currentItem, defaultScriptsByType);
    if (!normalized) {
      return;
    }
    if (seen.has(normalized.src) || isBuiltinBehaviorSource(normalized.src)) {
      return;
    }
    if (externalBySrc.has(normalized.src)) {
      return;
    }
    if (previousExternalSrcSet.has(normalized.src)) {
      return;
    }
    nextRegistry.push(normalized);
    seen.add(normalized.src);
  });

  return nextRegistry;
};

export const normalizeBehaviorRegistry = (
  registry: BehaviorRegistryItem[],
  options: Omit<
    ComposeBehaviorRegistryOptions,
    "currentRegistry" | "preferExternalForExternalSources"
  > = {},
): BehaviorRegistryItem[] =>
  composeBehaviorRegistry({
    ...options,
    currentRegistry: registry,
    preferExternalForExternalSources: false,
  });

export const mergeBehaviorRegistryWithExternal = (
  registry: BehaviorRegistryItem[],
  options: Omit<ComposeBehaviorRegistryOptions, "currentRegistry"> = {},
): BehaviorRegistryItem[] =>
  composeBehaviorRegistry({
    ...options,
    currentRegistry: registry,
    preferExternalForExternalSources: true,
  });

export const buildBehaviorSourceIndex = (
  registry: BehaviorRegistryItem[],
  externalRegistry: BehaviorRegistryItem[] = [],
): Record<string, BehaviorRegistrySource> => {
  const externalSources = new Set(
    externalRegistry
      .filter((entry) => entry && typeof entry.src === "string")
      .map((entry) => entry.src),
  );

  const sourceIndex: Record<string, BehaviorRegistrySource> = {};

  registry.forEach((entry) => {
    if (isBuiltinBehaviorSource(entry.src)) {
      sourceIndex[entry.src] = "builtin";
      return;
    }

    if (externalSources.has(entry.src)) {
      sourceIndex[entry.src] = "external";
      return;
    }

    sourceIndex[entry.src] = "user";
  });

  return sourceIndex;
};
