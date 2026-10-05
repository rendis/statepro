import {
  createInitialMetadataPackBindingMap,
} from "./defaults";
import { composeBehaviorRegistry } from "./behaviorRegistry";
import { applyStudioLayoutDocument } from "./studioLayout";
import { deserializeStatePro } from "./deserializeStatePro";
import type { StudioLocale } from "../i18n";
import type {
  BehaviorRegistryItem,
  EditorState,
  StudioExternalValue,
} from "../types";

type BuildEditorStateFromExternalValueOptions = {
  libraryBehaviors?: BehaviorRegistryItem[];
  locale?: StudioLocale;
};

export const buildEditorStateFromExternalValue = (
  value: StudioExternalValue,
  options: BuildEditorStateFromExternalValueOptions = {},
): EditorState => {
  const locale = options.locale;
  let state = deserializeStatePro(value.definition);

  if (value.layout) {
    const appliedLayout = applyStudioLayoutDocument(state, value.layout);
    state = appliedLayout.state;
  }

  if (value.metadataPacks) {
    state = {
      ...state,
      metadataPackRegistry: structuredClone(value.metadataPacks.registry || []),
      metadataPackBindings: structuredClone(
        value.metadataPacks.bindings || createInitialMetadataPackBindingMap(),
      ),
    };
  }

  if (options.libraryBehaviors) {
    state = {
      ...state,
      registry: composeBehaviorRegistry({
        locale,
        currentRegistry: state.registry,
        externalRegistry: structuredClone(options.libraryBehaviors),
        preferExternalForExternalSources: true,
      }),
    };
  }

  return state;
};

export type ExternalValueBuildResult = {
  /** Null when the value was rejected; `error` then explains why. */
  state: EditorState | null;
  error: unknown;
};

const isPlainObject = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

/**
 * Host-provided values are untrusted. A definition that is not a machine object, or
 * that cannot be deserialized, is reported instead of crashing the editor; callers
 * decide which state to keep. Semantically invalid machines still load so they can
 * be fixed in the editor.
 */
export const tryBuildEditorStateFromExternalValue = (
  value: StudioExternalValue,
  options: BuildEditorStateFromExternalValueOptions = {},
): ExternalValueBuildResult => {
  const definition: unknown = isPlainObject(value) ? value.definition : undefined;
  if (!isPlainObject(definition) || !isPlainObject(definition.universes)) {
    return { state: null, error: new TypeError("definition must be an object with a universes map") };
  }

  try {
    return { state: buildEditorStateFromExternalValue(value, options), error: null };
  } catch (error) {
    return { state: null, error };
  }
};
