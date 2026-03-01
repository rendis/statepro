import type { EditorNode, EditorState, StudioLayoutDocument } from "../types";
import {
  buildRealityEntityRef,
  buildTransitionEntityRef,
  buildUniverseEntityRef,
  normalizeTransitionsOrder,
} from "../utils";

export const serializeStudioLayout = (state: EditorState): StudioLayoutDocument => {
  const universeNodes = state.nodes.filter(
    (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
  );
  const realityNodes = state.nodes.filter(
    (node): node is Extract<EditorNode, { type: "reality" }> => node.type === "reality",
  );
  const noteNodes = state.nodes.filter(
    (node): node is Extract<EditorNode, { type: "note" }> => node.type === "note",
  );
  const universeByNodeId = new Map(universeNodes.map((node) => [node.id, node]));
  const realityByNodeId = new Map(realityNodes.map((node) => [node.id, node]));
  const orderedTransitions = normalizeTransitionsOrder(state.transitions);

  return {
    version: 1,
    machineRef: {
      id: state.machineConfig.id,
      canonicalName: state.machineConfig.canonicalName,
      version: state.machineConfig.version,
    },
    nodes: {
      universes: universeNodes.map((node) => ({
        entityRef: buildUniverseEntityRef(node.data.id),
        x: node.x,
        y: node.y,
        w: node.w,
        h: node.h,
        note: node.data.note ? structuredClone(node.data.note) : null,
      })),
      realities: realityNodes
        .map((node) => {
          const parentUniverse = universeByNodeId.get(node.data.universeId);
          if (!parentUniverse) {
            return null;
          }

          return {
            entityRef: buildRealityEntityRef(parentUniverse.data.id, node.data.id),
            x: node.x,
            y: node.y,
            note: node.data.note ? structuredClone(node.data.note) : null,
          };
        })
        .filter(
          (
            snapshot,
          ): snapshot is StudioLayoutDocument["nodes"]["realities"][number] => Boolean(snapshot),
        ),
      globalNotes: noteNodes.map((node) => ({
        x: node.x,
        y: node.y,
        data: structuredClone(node.data),
      })),
    },
    transitions: orderedTransitions.reduce<StudioLayoutDocument["transitions"]>(
      (accumulator, transition) => {
        const sourceReality = realityByNodeId.get(transition.sourceRealityId);
        if (!sourceReality) {
          return accumulator;
        }
        const sourceUniverse = universeByNodeId.get(sourceReality.data.universeId);
        if (!sourceUniverse) {
          return accumulator;
        }

        accumulator.push({
          entityRef: buildTransitionEntityRef(
            sourceUniverse.data.id,
            sourceReality.data.id,
            transition.triggerKind,
            transition.eventName,
            transition.order,
          ),
          visualOffset: transition.visualOffset
            ? {
                x: transition.visualOffset.x,
                y: transition.visualOffset.y,
              }
            : undefined,
          note: transition.note ? structuredClone(transition.note) : null,
        });
        return accumulator;
      },
      [],
    ),
    packs: {
      packRegistry: structuredClone(state.metadataPackRegistry),
      bindings: structuredClone(state.metadataPackBindings),
    },
  };
};
