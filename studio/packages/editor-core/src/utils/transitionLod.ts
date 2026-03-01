type TransitionLike = {
  id: string;
};

type AnchorPoint = {
  x: number;
  y: number;
};

export interface SkeletonTransitionSelectionOptions {
  transitions: TransitionLike[];
  alwaysVisibleTransitionIds: Iterable<string>;
  transitionBadgeAnchors: Map<string, AnchorPoint>;
  viewportCenter: AnchorPoint;
  limit: number;
}

export const selectSkeletonTransitionIds = ({
  transitions,
  alwaysVisibleTransitionIds,
  transitionBadgeAnchors,
  viewportCenter,
  limit,
}: SkeletonTransitionSelectionOptions): Set<string> => {
  const transitionIds = new Set(transitions.map((transition) => transition.id));
  const forcedIds = new Set<string>();

  for (const id of alwaysVisibleTransitionIds) {
    if (transitionIds.has(id)) {
      forcedIds.add(id);
    }
  }

  if (transitions.length === 0) {
    return forcedIds;
  }

  const normalizedLimit = Number.isFinite(limit) ? Math.max(0, Math.floor(limit)) : 0;
  if (normalizedLimit === 0 || forcedIds.size >= normalizedLimit) {
    return forcedIds;
  }

  if (transitions.length <= normalizedLimit) {
    return new Set(transitions.map((transition) => transition.id));
  }

  const candidates = transitions
    .map((transition, index) => {
      const anchor = transitionBadgeAnchors.get(transition.id);
      const distanceSq = anchor
        ? (anchor.x - viewportCenter.x) * (anchor.x - viewportCenter.x) +
          (anchor.y - viewportCenter.y) * (anchor.y - viewportCenter.y)
        : Number.POSITIVE_INFINITY;

      return {
        id: transition.id,
        index,
        distanceSq,
        forced: forcedIds.has(transition.id),
      };
    })
    .filter((candidate) => !candidate.forced)
    .sort((left, right) => {
      if (left.distanceSq !== right.distanceSq) {
        return left.distanceSq - right.distanceSq;
      }
      return left.index - right.index;
    });

  const selected = new Set(forcedIds);
  for (const candidate of candidates) {
    if (selected.size >= normalizedLimit) {
      break;
    }
    selected.add(candidate.id);
  }

  return selected;
};
