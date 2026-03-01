export const BADGE_WIDTH = 188;
export const BADGE_HEIGHT = 34;
export const BADGE_RADIUS = 17;
export const BADGE_PORT_RADIUS = 5;
export const TRIGGER_MAX_CHARS = 18;

export const truncateTriggerLabel = (label: string, maxChars = TRIGGER_MAX_CHARS): string => {
  if (label.length <= maxChars) {
    return label;
  }

  return `${label.slice(0, Math.max(maxChars - 3, 1))}...`;
};
