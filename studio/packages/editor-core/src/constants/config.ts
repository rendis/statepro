import type { BehaviorType, BehaviorTypeConfig, RealityNodeTypeConfig, RealityType } from "../types";
import { STUDIO_ICON_REGISTRY } from "./icons";

export const REALITY_TYPE_LABEL_KEYS: Record<RealityType, string> = {
  normal: "realityType.normal",
  success: "realityType.success",
  error: "realityType.error",
};

export const BEHAVIOR_TYPE_LABEL_KEYS: Record<BehaviorType, string> = {
  action: "behaviorType.action",
  invoke: "behaviorType.invoke",
  condition: "behaviorType.condition",
  observer: "behaviorType.observer",
};

export const REALITY_TYPES: Record<RealityType, RealityNodeTypeConfig> = {
  normal: {
    icon: STUDIO_ICON_REGISTRY.reality.normal.icon,
    color: STUDIO_ICON_REGISTRY.reality.normal.colors.base,
    border: STUDIO_ICON_REGISTRY.reality.normal.border,
    bg: STUDIO_ICON_REGISTRY.reality.normal.bg,
    label: "Normal",
  },
  success: {
    icon: STUDIO_ICON_REGISTRY.reality.success.icon,
    color: STUDIO_ICON_REGISTRY.reality.success.colors.base,
    border: STUDIO_ICON_REGISTRY.reality.success.border,
    bg: STUDIO_ICON_REGISTRY.reality.success.bg,
    label: "Success",
  },
  error: {
    icon: STUDIO_ICON_REGISTRY.reality.error.icon,
    color: STUDIO_ICON_REGISTRY.reality.error.colors.base,
    border: STUDIO_ICON_REGISTRY.reality.error.border,
    bg: STUDIO_ICON_REGISTRY.reality.error.bg,
    label: "Error",
  },
};

export const BEHAVIOR_TYPES: Record<BehaviorType, BehaviorTypeConfig> = {
  action: {
    icon: STUDIO_ICON_REGISTRY.behavior.action.icon,
    color: STUDIO_ICON_REGISTRY.behavior.action.colors.base,
    border: STUDIO_ICON_REGISTRY.behavior.action.border,
    bg: STUDIO_ICON_REGISTRY.behavior.action.bg,
    label: "Action",
  },
  invoke: {
    icon: STUDIO_ICON_REGISTRY.behavior.invoke.icon,
    color: STUDIO_ICON_REGISTRY.behavior.invoke.colors.base,
    border: STUDIO_ICON_REGISTRY.behavior.invoke.border,
    bg: STUDIO_ICON_REGISTRY.behavior.invoke.bg,
    label: "Invoke",
  },
  condition: {
    icon: STUDIO_ICON_REGISTRY.behavior.condition.icon,
    color: STUDIO_ICON_REGISTRY.behavior.condition.colors.base,
    border: STUDIO_ICON_REGISTRY.behavior.condition.border,
    bg: STUDIO_ICON_REGISTRY.behavior.condition.bg,
    label: "Condition",
  },
  observer: {
    icon: STUDIO_ICON_REGISTRY.behavior.observer.icon,
    color: STUDIO_ICON_REGISTRY.behavior.observer.colors.base,
    border: STUDIO_ICON_REGISTRY.behavior.observer.border,
    bg: STUDIO_ICON_REGISTRY.behavior.observer.bg,
    label: "Observer",
  },
};
