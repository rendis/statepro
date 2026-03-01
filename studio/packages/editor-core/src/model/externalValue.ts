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
