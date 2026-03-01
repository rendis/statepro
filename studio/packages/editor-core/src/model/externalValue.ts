import {
  createInitialMetadataPackBindingMap,
} from "./defaults";
import { applyStudioLayoutDocument } from "./studioLayout";
import { deserializeStatePro } from "./deserializeStatePro";
import type {
  BehaviorRegistryItem,
  EditorState,
  StudioExternalValue,
} from "../types";

type BuildEditorStateFromExternalValueOptions = {
  libraryBehaviors?: BehaviorRegistryItem[];
};

export const buildEditorStateFromExternalValue = (
  value: StudioExternalValue,
  options: BuildEditorStateFromExternalValueOptions = {},
): EditorState => {
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
      registry: structuredClone(options.libraryBehaviors),
    };
  }

  return state;
};
