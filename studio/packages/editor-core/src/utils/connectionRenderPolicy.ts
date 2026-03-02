export type ConnectionRenderMode = "full" | "skeleton";

export const resolveConnectionRenderMode = ({
  isNavigating,
  transitionCount,
  navigatingFullTransitionThreshold,
}: {
  isNavigating: boolean;
  transitionCount: number;
  navigatingFullTransitionThreshold: number;
}): ConnectionRenderMode => {
  if (isNavigating && transitionCount > navigatingFullTransitionThreshold) {
    return "skeleton";
  }
  return "full";
};
