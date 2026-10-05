import type { ELK as ElkInstance, ELKConstructorArguments, ElkExtendedEdge, ElkNode } from "elkjs/lib/elk-api";

import type { EditorNode, EditorTransition, NodeSizeMap } from "../types";
import {
  buildNodeReferenceIndex,
  resolveTargetReferenceToNodeId,
  type NodeReferenceIndex,
} from "./references";

const DEFAULT_REALITY_WIDTH = 192;
const DEFAULT_REALITY_HEIGHT = 150;
const MIN_UNIVERSE_WIDTH = 240;
const MIN_UNIVERSE_HEIGHT = 220;
const TRANSITION_BADGE_WIDTH = 280;
const TRANSITION_LABEL_HORIZONTAL_PADDING = 40;
const MIN_LABEL_CHANNEL_WIDTH = TRANSITION_BADGE_WIDTH + TRANSITION_LABEL_HORIZONTAL_PADDING;
const INTERNAL_LAYER_GAP = Math.max(120, MIN_LABEL_CHANNEL_WIDTH);
const INTERNAL_NODE_VERTICAL_GAP = 110;
const EXTERNAL_LAYER_GAP = 360;
const EXTERNAL_NODE_VERTICAL_GAP = 200;

const INNER_PADDING_LEFT = 20;
const INNER_PADDING_RIGHT = 20;
const INNER_PADDING_TOP = 52;
const INNER_PADDING_BOTTOM = 20;

const DISCONNECTED_COMPONENT_GAP = 180;

type ElkConstructorLike = new (args?: ELKConstructorArguments) => ElkInstance;

const resolveElkConstructor = (bundle: unknown): ElkConstructorLike => {
  const moduleShape = bundle as {
    default?: ElkConstructorLike;
    ELK?: ElkConstructorLike;
  };
  const globalShape = globalThis as typeof globalThis & {
    ELK?: ElkConstructorLike;
  };

  const elkCtor = moduleShape.default || moduleShape.ELK || globalShape.ELK;
  if (!elkCtor) {
    throw new Error("StatePro Studio: failed to resolve ELK constructor from bundled module.");
  }

  return elkCtor;
};

let elkPromise: Promise<ElkInstance> | undefined;
let workerConstructorPromise: Promise<ElkConstructorLike> | undefined;

const getWorkerConstructor = (): Promise<ElkConstructorLike> => {
  if (!workerConstructorPromise) {
    workerConstructorPromise = import("elkjs/lib/elk-api.js")
      .then(resolveElkConstructor)
      .catch(error => {
        workerConstructorPromise = undefined;
        throw error;
      });
  }
  return workerConstructorPromise;
};

const getElk = (): Promise<ElkInstance> => {
  if (!elkPromise) {
    elkPromise = import("elkjs/lib/elk.bundled.js")
      .then((bundle) => new (resolveElkConstructor(bundle))())
      .catch((error) => {
        elkPromise = undefined;
        throw error;
      });
  }
  return elkPromise;
};

const LAYOUT_WORKER_TIMEOUT_MS = 30_000;
const LAYOUT_WORKER_IDLE_MS = 60_000;

/** The worker could not run layouts; the request falls back to the bundled engine. */
class LayoutWorkerUnavailableError extends Error {}
/** The worker was released on purpose (idle or URL change); its URL is still usable. */
class LayoutWorkerReleasedError extends LayoutWorkerUnavailableError {}
class LayoutWorkerTimeoutError extends Error {}

interface WorkerEngine {
  worker: Worker;
  elk: ElkInstance;
  /** Rejects once the worker is discarded, so in-flight requests stop waiting on it. */
  discarded: Promise<never>;
  discard: (error: Error) => void;
  pending: number;
  /** Set after the first completed layout; later failures are crashes, not load failures. */
  succeeded: boolean;
  idleTimer?: ReturnType<typeof setTimeout>;
}

interface WorkerEngineEntry {
  url: string;
  promise: Promise<WorkerEngine>;
  engine?: WorkerEngine;
}

// One worker per URL is reused across layouts: elk-api tags each request with an id,
// so concurrent layouts share it safely and ELK is initialized only once.
let workerEngineEntry: WorkerEngineEntry | undefined;
// URLs whose worker never completed a layout (404, CSP, cross-origin). Later layouts
// go straight to the bundled engine instead of failing or retrying the same URL.
const unavailableWorkerUrls = new Set<string>();

const createWorkerEngine = async (entry: WorkerEngineEntry): Promise<WorkerEngine> => {
  let ElkConstructor: ElkConstructorLike;
  let worker: Worker;
  try {
    ElkConstructor = await getWorkerConstructor();
    worker = new Worker(entry.url);
  } catch (error) {
    throw new LayoutWorkerUnavailableError(`StatePro Studio: layout worker could not start: ${String(error)}`);
  }

  let rejectDiscarded!: (error: Error) => void;
  const discarded = new Promise<never>((_, reject) => { rejectDiscarded = reject; });
  discarded.catch(() => {});
  const onError = () => engine.discard(
    new LayoutWorkerUnavailableError("StatePro Studio: layout worker failed to load or execute."),
  );
  const engine: WorkerEngine = {
    worker,
    elk: undefined as unknown as ElkInstance,
    discarded,
    pending: 0,
    succeeded: false,
    discard: (error) => {
      if (workerEngineEntry?.engine === engine) workerEngineEntry = undefined;
      clearTimeout(engine.idleTimer);
      worker.removeEventListener("error", onError);
      worker.removeEventListener("messageerror", onError);
      worker.terminate();
      rejectDiscarded(error);
    },
  };
  entry.engine = engine;
  worker.addEventListener("error", onError);
  worker.addEventListener("messageerror", onError);
  try {
    engine.elk = new ElkConstructor({ workerFactory: () => worker, algorithms: ["layered"] });
  } catch (error) {
    const unavailable = new LayoutWorkerUnavailableError(
      `StatePro Studio: layout worker could not start: ${String(error)}`,
    );
    engine.discard(unavailable);
    throw unavailable;
  }
  return engine;
};

const getWorkerEngine = (url: string): Promise<WorkerEngine> => {
  if (workerEngineEntry?.url !== url) {
    workerEngineEntry?.engine?.discard(
      new LayoutWorkerReleasedError("StatePro Studio: layout worker URL changed."),
    );
    const entry: WorkerEngineEntry = { url, promise: undefined as unknown as Promise<WorkerEngine> };
    entry.promise = createWorkerEngine(entry);
    entry.promise.catch(() => {
      if (workerEngineEntry === entry) workerEngineEntry = undefined;
    });
    workerEngineEntry = entry;
  }
  return workerEngineEntry.promise;
};

const runOnBundledEngine = async <T>(
  run: (elk: ElkInstance) => Promise<T>,
  reason?: unknown,
): Promise<T> => {
  if (reason) {
    console.warn("StatePro Studio: falling back to main-thread layout.", reason);
  }
  return run(await getElk());
};

// Layout runs in the host-provided worker when available. If the worker cannot load
// or crashes, the request falls back to the bundled engine on the main thread. A
// timeout rejects instead: running a stuck graph on the main thread would freeze the UI.
const withLayoutEngine = async <T>(
  workerUrl: string | undefined,
  run: (elk: ElkInstance) => Promise<T>,
): Promise<T> => {
  if (!workerUrl || typeof Worker === "undefined" || unavailableWorkerUrls.has(workerUrl)) {
    return runOnBundledEngine(run);
  }

  let engine: WorkerEngine;
  try {
    engine = await getWorkerEngine(workerUrl);
  } catch (error) {
    unavailableWorkerUrls.add(workerUrl);
    return runOnBundledEngine(run, error);
  }

  clearTimeout(engine.idleTimer);
  engine.pending += 1;
  let timer: ReturnType<typeof setTimeout> | undefined;
  const timeout = new Promise<never>((_, reject) => {
    timer = setTimeout(
      () => reject(new LayoutWorkerTimeoutError(
        `StatePro Studio: layout worker timed out after ${LAYOUT_WORKER_TIMEOUT_MS / 1000} seconds.`,
      )),
      LAYOUT_WORKER_TIMEOUT_MS,
    );
  });
  try {
    const result = await Promise.race([run(engine.elk), engine.discarded, timeout]);
    engine.succeeded = true;
    return result;
  } catch (error) {
    if (error instanceof LayoutWorkerUnavailableError) {
      if (!engine.succeeded && !(error instanceof LayoutWorkerReleasedError)) {
        unavailableWorkerUrls.add(workerUrl);
      }
      return runOnBundledEngine(run, error);
    }
    if (error instanceof LayoutWorkerTimeoutError) {
      // A stuck worker is never reused; the next layout starts a fresh one.
      engine.discard(error);
    }
    throw error;
  } finally {
    clearTimeout(timer);
    engine.pending -= 1;
    if (engine.pending === 0 && workerEngineEntry?.engine === engine) {
      // Release the worker (and ELK's memory) once the editor stops requesting layouts.
      engine.idleTimer = setTimeout(
        () => engine.discard(new LayoutWorkerReleasedError("StatePro Studio: idle layout worker released.")),
        LAYOUT_WORKER_IDLE_MS,
      );
    }
  }
};

type UniverseNode = Extract<EditorNode, { type: "universe" }>;
type RealityNode = Extract<EditorNode, { type: "reality" }>;

interface InternalRealityLayout {
  id: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

interface InternalUniverseLayout {
  id: string;
  w: number;
  h: number;
  realities: InternalRealityLayout[];
}

interface PositionedUniverse {
  id: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

const isFiniteNumber = (value: unknown): value is number => {
  return typeof value === "number" && Number.isFinite(value);
};

const getRealitySize = (nodeId: string, nodeSizes: NodeSizeMap): { w: number; h: number } => {
  return nodeSizes[nodeId] || { w: DEFAULT_REALITY_WIDTH, h: DEFAULT_REALITY_HEIGHT };
};

const compareByVisualOrder = (
  left: { x: number; y: number; id: string },
  right: { x: number; y: number; id: string },
): number => {
  if (left.x !== right.x) {
    return left.x - right.x;
  }
  if (left.y !== right.y) {
    return left.y - right.y;
  }
  return left.id.localeCompare(right.id);
};

const toComponentList = (
  universeIds: string[],
  edges: ElkExtendedEdge[],
): string[][] => {
  const adjacency = new Map<string, Set<string>>();
  universeIds.forEach((id) => adjacency.set(id, new Set<string>()));

  edges.forEach((edge) => {
    const source = edge.sources?.[0];
    const target = edge.targets?.[0];
    if (!source || !target || source === target) {
      return;
    }
    adjacency.get(source)?.add(target);
    adjacency.get(target)?.add(source);
  });

  const visited = new Set<string>();
  const components: string[][] = [];

  universeIds.forEach((id) => {
    if (visited.has(id)) {
      return;
    }

    const stack = [id];
    visited.add(id);
    const component: string[] = [];

    while (stack.length > 0) {
      const current = stack.pop();
      if (!current) {
        continue;
      }
      component.push(current);

      const neighbors = adjacency.get(current);
      if (!neighbors) {
        continue;
      }

      neighbors.forEach((neighbor) => {
        if (visited.has(neighbor)) {
          return;
        }
        visited.add(neighbor);
        stack.push(neighbor);
      });
    }

    components.push(component);
  });

  return components;
};

const buildInternalUniverseEdges = (
  universeId: string,
  realities: RealityNode[],
  transitions: EditorTransition[],
  nodes: EditorNode[],
  index: NodeReferenceIndex,
): ElkExtendedEdge[] => {
  const realityIds = new Set(realities.map((reality) => reality.id));
  const edges: ElkExtendedEdge[] = [];
  const seenPairs = new Set<string>();

  transitions.forEach((transition) => {
    if (!realityIds.has(transition.sourceRealityId)) {
      return;
    }

    transition.targets.forEach((targetRef, targetIndex) => {
      const targetNodeId = resolveTargetReferenceToNodeId(
        transition.sourceRealityId,
        targetRef,
        nodes,
        index,
      );
      if (!targetNodeId || !realityIds.has(targetNodeId) || targetNodeId === transition.sourceRealityId) {
        return;
      }

      const pair = `${transition.sourceRealityId}->${targetNodeId}`;
      if (seenPairs.has(pair)) {
        return;
      }
      seenPairs.add(pair);

      edges.push({
        id: `${universeId}::internal::${transition.id}::${targetIndex}`,
        sources: [transition.sourceRealityId],
        targets: [targetNodeId],
      });
    });
  });

  return edges;
};

const buildInDegreeByRealityId = (
  realityIds: string[],
  edges: ElkExtendedEdge[],
): Map<string, number> => {
  const inDegreeByRealityId = new Map(realityIds.map((id) => [id, 0]));

  edges.forEach((edge) => {
    const targetId = edge.targets?.[0];
    if (!targetId || !inDegreeByRealityId.has(targetId)) {
      return;
    }

    inDegreeByRealityId.set(targetId, (inDegreeByRealityId.get(targetId) || 0) + 1);
  });

  return inDegreeByRealityId;
};

const alignAndCenterStartRealities = (
  realities: InternalRealityLayout[],
  startRealityIds: Set<string>,
  verticalGap: number,
): InternalRealityLayout[] => {
  if (startRealityIds.size === 0) {
    return realities;
  }

  const startEntries = realities.filter((entry) => startRealityIds.has(entry.id));
  if (startEntries.length === 0) {
    return realities;
  }

  const orderedStarts = [...startEntries].sort((left, right) => {
    if (left.y !== right.y) {
      return left.y - right.y;
    }
    return left.id.localeCompare(right.id);
  });

  const startColumnX = Math.min(...orderedStarts.map((entry) => entry.x));
  const contentTop = INNER_PADDING_TOP;
  const contentBottom = Math.max(...realities.map((entry) => entry.y + entry.h));
  const startBlockHeight = orderedStarts.reduce(
    (sum, entry, index) => sum + entry.h + (index > 0 ? verticalGap : 0),
    0,
  );

  const centeredStartY =
    contentTop + (contentBottom - contentTop - startBlockHeight) / 2;
  const maxStartY = contentBottom - startBlockHeight;
  const firstStartY =
    maxStartY >= contentTop
      ? Math.min(Math.max(centeredStartY, contentTop), maxStartY)
      : contentTop;

  const adjustedStartPositionById = new Map<string, { x: number; y: number }>();
  let cursorY = firstStartY;
  orderedStarts.forEach((entry) => {
    adjustedStartPositionById.set(entry.id, {
      x: startColumnX,
      y: cursorY,
    });
    cursorY += entry.h + verticalGap;
  });

  return realities.map((entry) => {
    const adjusted = adjustedStartPositionById.get(entry.id);
    if (!adjusted) {
      return entry;
    }

    return {
      ...entry,
      x: adjusted.x,
      y: adjusted.y,
    };
  });
};

const computeInternalUniverseLayout = async (
  universe: UniverseNode,
  realities: RealityNode[],
  transitions: EditorTransition[],
  nodes: EditorNode[],
  nodeSizes: NodeSizeMap,
  index: NodeReferenceIndex,
  elk: ElkInstance,
): Promise<InternalUniverseLayout> => {
  if (realities.length === 0) {
    return {
      id: universe.id,
      w: MIN_UNIVERSE_WIDTH,
      h: MIN_UNIVERSE_HEIGHT,
      realities: [],
    };
  }

  const orderedRealities = [...realities].sort(compareByVisualOrder);
  const children: ElkNode[] = orderedRealities.map((reality) => {
    const size = getRealitySize(reality.id, nodeSizes);
    return {
      id: reality.id,
      width: size.w,
      height: size.h,
    };
  });

  const edges = buildInternalUniverseEdges(
    universe.id,
    orderedRealities,
    transitions,
    nodes,
    index,
  );
  const inDegreeByRealityId = buildInDegreeByRealityId(
    orderedRealities.map((reality) => reality.id),
    edges,
  );
  const startRealityIds = orderedRealities
    .map((reality) => reality.id)
    .filter((realityId) => (inDegreeByRealityId.get(realityId) || 0) === 0);

  const graph: ElkNode = {
    id: `universe:${universe.id}`,
    layoutOptions: {
      "elk.algorithm": "layered",
      "elk.direction": "RIGHT",
      "elk.edgeRouting": "ORTHOGONAL",
      "elk.layered.considerModelOrder": "NODES_AND_EDGES",
      "elk.layered.crossingMinimization.forceNodeModelOrder": "true",
      "elk.spacing.nodeNode": String(INTERNAL_NODE_VERTICAL_GAP),
      "elk.layered.spacing.nodeNodeBetweenLayers": String(INTERNAL_LAYER_GAP),
      "elk.layered.spacing.edgeNodeBetweenLayers": "80",
      "elk.layered.spacing.edgeEdgeBetweenLayers": "40",
    },
    children,
    edges,
  };

  const result = await elk.layout(graph, { measureExecutionTime: true });
  const childById = new Map((result.children || []).map((child) => [child.id, child]));

  const rawLayout = orderedRealities.map((reality, indexPosition) => {
    const child = childById.get(reality.id);
    const size = getRealitySize(reality.id, nodeSizes);
    const fallbackX = indexPosition * (size.w + 80);

    return {
      id: reality.id,
      x: isFiniteNumber(child?.x) ? child.x : fallbackX,
      y: isFiniteNumber(child?.y) ? child.y : 0,
      w: isFiniteNumber(child?.width) ? child.width : size.w,
      h: isFiniteNumber(child?.height) ? child.height : size.h,
    };
  });

  const minX = Math.min(...rawLayout.map((entry) => entry.x));
  const minY = Math.min(...rawLayout.map((entry) => entry.y));

  const normalized = rawLayout.map((entry) => ({
    ...entry,
    x: entry.x - minX + INNER_PADDING_LEFT,
    y: entry.y - minY + INNER_PADDING_TOP,
  }));
  const adjustedRealities =
    startRealityIds.length > 0
      ? alignAndCenterStartRealities(
          normalized,
          new Set(startRealityIds),
          INTERNAL_NODE_VERTICAL_GAP,
        )
      : normalized;

  const maxRight = Math.max(...adjustedRealities.map((entry) => entry.x + entry.w));
  const maxBottom = Math.max(...adjustedRealities.map((entry) => entry.y + entry.h));

  return {
    id: universe.id,
    w: Math.max(MIN_UNIVERSE_WIDTH, maxRight + INNER_PADDING_RIGHT),
    h: Math.max(MIN_UNIVERSE_HEIGHT, maxBottom + INNER_PADDING_BOTTOM),
    realities: adjustedRealities,
  };
};

const buildCrossUniverseEdges = (
  transitions: EditorTransition[],
  nodes: EditorNode[],
  index: NodeReferenceIndex,
): ElkExtendedEdge[] => {
  const realitiesById = new Map(
    nodes
      .filter((node): node is RealityNode => node.type === "reality")
      .map((reality) => [reality.id, reality]),
  );
  const nodesById = new Map(nodes.map((node) => [node.id, node]));
  const edges: ElkExtendedEdge[] = [];
  const seenPairs = new Set<string>();

  transitions.forEach((transition) => {
    const sourceReality = realitiesById.get(transition.sourceRealityId);
    if (!sourceReality) {
      return;
    }

    const sourceUniverseId = sourceReality.data.universeId;
    transition.targets.forEach((targetRef, targetIndex) => {
      const targetNodeId = resolveTargetReferenceToNodeId(
        transition.sourceRealityId,
        targetRef,
        nodes,
        index,
      );
      if (!targetNodeId) {
        return;
      }

      const targetNode = nodesById.get(targetNodeId);
      if (!targetNode) {
        return;
      }

      const targetUniverseId =
        targetNode.type === "universe"
          ? targetNode.id
          : targetNode.type === "reality"
            ? targetNode.data.universeId
            : null;

      if (!targetUniverseId || sourceUniverseId === targetUniverseId) {
        return;
      }

      const pair = `${sourceUniverseId}->${targetUniverseId}`;
      if (seenPairs.has(pair)) {
        return;
      }
      seenPairs.add(pair);

      edges.push({
        id: `cross::${transition.id}::${targetIndex}`,
        sources: [sourceUniverseId],
        targets: [targetUniverseId],
      });
    });
  });

  return edges;
};

const reorderDisconnectedComponents = (
  positionedUniverses: PositionedUniverse[],
  universeOrder: UniverseNode[],
  edges: ElkExtendedEdge[],
): Map<string, PositionedUniverse> => {
  const byId = new Map(positionedUniverses.map((entry) => [entry.id, entry]));
  const components = toComponentList(
    universeOrder.map((universe) => universe.id),
    edges,
  );

  const indexByUniverseId = new Map(
    universeOrder.map((universe, index) => [universe.id, index]),
  );

  const orderedComponents = [...components].sort((left, right) => {
    const leftIndex = Math.min(...left.map((id) => indexByUniverseId.get(id) ?? Number.MAX_SAFE_INTEGER));
    const rightIndex = Math.min(...right.map((id) => indexByUniverseId.get(id) ?? Number.MAX_SAFE_INTEGER));
    return leftIndex - rightIndex;
  });

  const reordered = new Map<string, PositionedUniverse>();
  let cursorX = 0;

  orderedComponents.forEach((component) => {
    const layouts = component
      .map((id) => byId.get(id))
      .filter((entry): entry is PositionedUniverse => Boolean(entry));
    if (layouts.length === 0) {
      return;
    }

    const minX = Math.min(...layouts.map((entry) => entry.x));
    const minY = Math.min(...layouts.map((entry) => entry.y));
    const maxX = Math.max(...layouts.map((entry) => entry.x + entry.w));
    const componentWidth = maxX - minX;

    layouts.forEach((entry) => {
      reordered.set(entry.id, {
        ...entry,
        x: entry.x - minX + cursorX,
        y: entry.y - minY,
      });
    });

    cursorX += componentWidth + DISCONNECTED_COMPONENT_GAP;
  });

  return reordered;
};

const anchorToCurrentUniverseArea = (
  positionedById: Map<string, PositionedUniverse>,
  currentUniverses: UniverseNode[],
): Map<string, PositionedUniverse> => {
  const positioned = Array.from(positionedById.values());
  if (positioned.length === 0 || currentUniverses.length === 0) {
    return positionedById;
  }

  const currentMinX = Math.min(...currentUniverses.map((universe) => universe.x));
  const currentMinY = Math.min(...currentUniverses.map((universe) => universe.y));
  const nextMinX = Math.min(...positioned.map((universe) => universe.x));
  const nextMinY = Math.min(...positioned.map((universe) => universe.y));
  const offsetX = currentMinX - nextMinX;
  const offsetY = currentMinY - nextMinY;

  const anchored = new Map<string, PositionedUniverse>();
  positionedById.forEach((entry, id) => {
    anchored.set(id, {
      ...entry,
      x: entry.x + offsetX,
      y: entry.y + offsetY,
    });
  });
  return anchored;
};

export const computeAutoLayout = async (
  nodes: EditorNode[],
  transitions: EditorTransition[],
  nodeSizes: NodeSizeMap,
  workerUrl?: string,
): Promise<EditorNode[]> => {
  const universes = nodes
    .filter((node): node is UniverseNode => node.type === "universe")
    .sort(compareByVisualOrder);

  if (universes.length === 0) {
    return nodes;
  }

  return withLayoutEngine(workerUrl, elk => computeLayoutWithEngine(nodes, transitions, nodeSizes, universes, elk));
};

const computeLayoutWithEngine = async (
  nodes: EditorNode[],
  transitions: EditorTransition[],
  nodeSizes: NodeSizeMap,
  universes: UniverseNode[],
  elk: ElkInstance,
): Promise<EditorNode[]> => {
  const realities = nodes.filter((node): node is RealityNode => node.type === "reality");
  const realitiesByUniverse = new Map<string, RealityNode[]>();
  universes.forEach((universe) => realitiesByUniverse.set(universe.id, []));
  realities.forEach((reality) => {
    const bucket = realitiesByUniverse.get(reality.data.universeId);
    if (!bucket) {
      return;
    }
    bucket.push(reality);
  });

  const referenceIndex = buildNodeReferenceIndex(nodes);
  const internalLayouts = await Promise.all(
    universes.map((universe) =>
      computeInternalUniverseLayout(
        universe,
        realitiesByUniverse.get(universe.id) || [],
        transitions,
        nodes,
        nodeSizes,
        referenceIndex,
        elk,
      ),
    ),
  );
  const internalLayoutByUniverseId = new Map(
    internalLayouts.map((layout) => [layout.id, layout]),
  );

  const externalEdges = buildCrossUniverseEdges(transitions, nodes, referenceIndex);
  const externalGraph: ElkNode = {
    id: "external-layout",
    layoutOptions: {
      "elk.algorithm": "layered",
      "elk.direction": "RIGHT",
      "elk.edgeRouting": "ORTHOGONAL",
      "elk.layered.considerModelOrder": "NODES_AND_EDGES",
      "elk.layered.crossingMinimization.forceNodeModelOrder": "true",
      "elk.spacing.nodeNode": String(EXTERNAL_NODE_VERTICAL_GAP),
      "elk.layered.spacing.nodeNodeBetweenLayers": String(EXTERNAL_LAYER_GAP),
      "elk.layered.spacing.edgeNodeBetweenLayers": "120",
      "elk.layered.spacing.edgeEdgeBetweenLayers": "80",
    },
    children: universes.map((universe) => {
      const internal = internalLayoutByUniverseId.get(universe.id);
      return {
        id: universe.id,
        width: internal?.w || universe.w,
        height: internal?.h || universe.h,
      };
    }),
    edges: externalEdges,
  };

  const externalLayout = await elk.layout(externalGraph, { measureExecutionTime: true });
  const externalById = new Map((externalLayout.children || []).map((child) => [child.id, child]));

  const positionedUniverses: PositionedUniverse[] = universes.map((universe, index) => {
    const child = externalById.get(universe.id);
    const internal = internalLayoutByUniverseId.get(universe.id);
    const width = internal?.w || universe.w;
    const height = internal?.h || universe.h;

    return {
      id: universe.id,
      x: isFiniteNumber(child?.x) ? child.x : index * (width + 180),
      y: isFiniteNumber(child?.y) ? child.y : 0,
      w: width,
      h: height,
    };
  });

  const reorderedByUniverseId = reorderDisconnectedComponents(
    positionedUniverses,
    universes,
    externalEdges,
  );
  const anchoredByUniverseId = anchorToCurrentUniverseArea(reorderedByUniverseId, universes);

  const positionedRealities = new Map<string, { x: number; y: number }>();
  anchoredByUniverseId.forEach((positionedUniverse, universeId) => {
    const internal = internalLayoutByUniverseId.get(universeId);
    if (!internal) {
      return;
    }

    internal.realities.forEach((reality) => {
      positionedRealities.set(reality.id, {
        x: positionedUniverse.x + reality.x,
        y: positionedUniverse.y + reality.y,
      });
    });
  });

  return nodes.map((node) => {
    if (node.type === "universe") {
      const positionedUniverse = anchoredByUniverseId.get(node.id);
      if (!positionedUniverse) {
        return node;
      }

      return {
        ...node,
        x: positionedUniverse.x,
        y: positionedUniverse.y,
        w: positionedUniverse.w,
        h: positionedUniverse.h,
      };
    }

    if (node.type === "reality") {
      const positionedReality = positionedRealities.get(node.id);
      if (!positionedReality) {
        return node;
      }

      return {
        ...node,
        x: positionedReality.x,
        y: positionedReality.y,
      };
    }

    return node;
  });
};
