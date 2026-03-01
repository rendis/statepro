import type { EditorNode } from "../types";

export type CanvasSearchMatchedField = "id" | "canonicalName" | "name" | "tag";
export type CanvasSearchScope = "universe" | "tag" | "reality";
export type CanvasSearchFilters = Record<CanvasSearchScope, boolean>;

export const DEFAULT_CANVAS_SEARCH_FILTERS: CanvasSearchFilters = {
  universe: true,
  tag: true,
  reality: true,
};

export interface CanvasSearchResult {
  nodeId: string;
  nodeType: "universe" | "reality";
  label: string;
  contextLabel?: string;
  matchedField: CanvasSearchMatchedField;
  matchedTag?: string;
  score: number;
}

const FIELD_PRIORITY: Record<CanvasSearchMatchedField, number> = {
  id: 0,
  canonicalName: 1,
  name: 2,
  tag: 3,
};

type MatchKind = "exact" | "prefix" | "contains";

const MATCH_KIND_PRIORITY: Record<MatchKind, number> = {
  exact: 0,
  prefix: 1,
  contains: 2,
};

const normalize = (value: string): string => value.trim().toLowerCase();

const resolveMatchKind = (value: string, query: string): MatchKind | null => {
  const normalizedValue = normalize(value);
  if (!normalizedValue) {
    return null;
  }
  if (normalizedValue === query) {
    return "exact";
  }
  if (normalizedValue.startsWith(query)) {
    return "prefix";
  }
  if (normalizedValue.includes(query)) {
    return "contains";
  }
  return null;
};

const buildScore = (matchKind: MatchKind, field: CanvasSearchMatchedField): number => {
  return MATCH_KIND_PRIORITY[matchKind] * 10 + FIELD_PRIORITY[field];
};

interface MatchCandidate {
  value: string;
  field: CanvasSearchMatchedField;
  matchedTag?: string;
}

interface BestNodeMatch {
  matchedField: CanvasSearchMatchedField;
  matchedTag?: string;
  score: number;
}

export interface CanvasSearchOptions {
  limit?: number;
  filters?: CanvasSearchFilters;
}

const resolveBestMatch = (candidates: MatchCandidate[], query: string): BestNodeMatch | null => {
  let best: BestNodeMatch | null = null;

  candidates.forEach((candidate) => {
    const matchKind = resolveMatchKind(candidate.value, query);
    if (!matchKind) {
      return;
    }

    const score = buildScore(matchKind, candidate.field);
    if (!best || score < best.score) {
      best = {
        matchedField: candidate.field,
        matchedTag: candidate.matchedTag,
        score,
      };
    }
  });

  return best;
};

export const searchCanvasNodes = (
  nodes: EditorNode[],
  rawQuery: string,
  limitOrOptions: number | CanvasSearchOptions = 20,
): CanvasSearchResult[] => {
  const resolvedLimit =
    typeof limitOrOptions === "number" ? limitOrOptions : (limitOrOptions.limit ?? 20);
  const resolvedFilters: CanvasSearchFilters =
    typeof limitOrOptions === "number"
      ? DEFAULT_CANVAS_SEARCH_FILTERS
      : (limitOrOptions.filters ?? DEFAULT_CANVAS_SEARCH_FILTERS);
  const query = normalize(rawQuery);
  if (!query) {
    return [];
  }
  if (!resolvedFilters.universe && !resolvedFilters.tag && !resolvedFilters.reality) {
    return [];
  }

  const universeNodes = nodes.filter(
    (node): node is Extract<EditorNode, { type: "universe" }> => node.type === "universe",
  );
  const universeLabelByNodeId = new Map(universeNodes.map((node) => [node.id, node.data.id || node.data.name]));

  const results: CanvasSearchResult[] = [];

  nodes.forEach((node) => {
    if (node.type === "note") {
      return;
    }

    if (node.type === "universe") {
      const candidates: MatchCandidate[] = [];
      if (resolvedFilters.universe) {
        candidates.push({ field: "id", value: node.data.id || node.data.name || "" });
        candidates.push({ field: "canonicalName", value: node.data.canonicalName || "" });
      }

      if (resolvedFilters.tag) {
        (node.data.tags || []).forEach((tag) => {
          candidates.push({
            field: "tag",
            value: tag,
            matchedTag: tag,
          });
        });
      }

      const best = resolveBestMatch(candidates, query);
      if (!best) {
        return;
      }

      results.push({
        nodeId: node.id,
        nodeType: node.type,
        label: node.data.id || node.data.name || node.id,
        contextLabel: node.data.canonicalName || undefined,
        matchedField: best.matchedField,
        matchedTag: best.matchedTag,
        score: best.score,
      });
      return;
    }

    if (!resolvedFilters.reality) {
      return;
    }

    const best = resolveBestMatch(
      [
        { field: "id", value: node.data.id || "" },
        { field: "name", value: node.data.name || "" },
      ],
      query,
    );
    if (!best) {
      return;
    }

    results.push({
      nodeId: node.id,
      nodeType: node.type,
      label: node.data.id || node.data.name || node.id,
      contextLabel: universeLabelByNodeId.get(node.data.universeId),
      matchedField: best.matchedField,
      score: best.score,
    });
  });

  return results
    .sort((left, right) => {
      if (left.score !== right.score) {
        return left.score - right.score;
      }
      const labelOrder = left.label.localeCompare(right.label);
      if (labelOrder !== 0) {
        return labelOrder;
      }
      return left.nodeId.localeCompare(right.nodeId);
    })
    .slice(0, Math.max(resolvedLimit, 0));
};
