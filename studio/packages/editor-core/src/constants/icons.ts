import type { ComponentType } from "react";
import {
  Activity,
  AlertCircle,
  AlertTriangle,
  ArrowRight,
  Bell,
  Box,
  CheckCircle,
  Circle,
  Eye,
  FastForward,
  LogIn,
  LogOut,
  Play,
  Radio,
  Route,
  Scale,
  Settings2,
  Wrench,
} from "lucide-react";

import type { BehaviorType, RealityType, TransitionTriggerKind, TransitionType } from "../types";

type StudioIcon = ComponentType<{ size?: string | number; className?: string }>;

export type StudioIconColorVariants = {
  base: string;
  muted?: string;
  accent?: string;
  emphasis?: string;
};

export type StudioIconToken = {
  icon: StudioIcon;
  colors: StudioIconColorVariants;
  border?: string;
  bg?: string;
};

type StudioIconSurfaceToken = StudioIconToken & {
  border: string;
  bg: string;
};

type StudioIconRegistry = {
  reality: Record<RealityType, StudioIconSurfaceToken> & { initial: StudioIconToken };
  behavior: Record<BehaviorType, StudioIconSurfaceToken>;
  transition: {
    trigger: Record<TransitionTriggerKind, StudioIconToken>;
    type: Record<TransitionType, StudioIconToken>;
  };
  phase: Record<"entry" | "exit" | "onTransition", StudioIconToken>;
  entity: Record<"machine" | "universe" | "reality", StudioIconToken>;
  status: Record<"warning", StudioIconToken>;
};

export const STUDIO_ICON_REGISTRY: StudioIconRegistry = {
  reality: {
    normal: {
      icon: Circle,
      colors: { base: "text-green-400" },
      border: "border-green-400",
      bg: "bg-green-400",
    },
    success: {
      icon: CheckCircle,
      colors: { base: "text-blue-400" },
      border: "border-blue-400",
      bg: "bg-blue-400",
    },
    error: {
      icon: AlertCircle,
      colors: { base: "text-red-400" },
      border: "border-red-400",
      bg: "bg-red-400",
    },
    initial: {
      icon: Play,
      colors: { base: "text-white" },
    },
  },
  behavior: {
    action: {
      icon: Wrench,
      colors: { base: "text-yellow-400" },
      border: "border-yellow-500/50",
      bg: "bg-yellow-900/30",
    },
    invoke: {
      icon: Activity,
      colors: { base: "text-purple-400" },
      border: "border-purple-500/50",
      bg: "bg-purple-900/30",
    },
    condition: {
      icon: Scale,
      colors: { base: "text-orange-400" },
      border: "border-orange-500/50",
      bg: "bg-orange-900/30",
    },
    observer: {
      icon: Eye,
      colors: { base: "text-cyan-400" },
      border: "border-cyan-500/50",
      bg: "bg-cyan-900/30",
    },
  },
  transition: {
    trigger: {
      on: {
        icon: Radio,
        colors: {
          base: "text-yellow-400",
          muted: "text-slate-400",
          accent: "text-blue-400",
          emphasis: "text-blue-300",
        },
      },
      always: {
        icon: FastForward,
        colors: { base: "text-purple-400" },
      },
    },
    type: {
      default: {
        icon: ArrowRight,
        colors: {
          base: "text-slate-400",
          accent: "text-blue-500",
          emphasis: "text-slate-300",
        },
      },
      notify: {
        icon: Bell,
        colors: { base: "text-yellow-400" },
      },
    },
  },
  phase: {
    entry: {
      icon: LogIn,
      colors: { base: "text-yellow-400" },
    },
    exit: {
      icon: LogOut,
      colors: { base: "text-orange-400" },
    },
    onTransition: {
      icon: Route,
      colors: { base: "text-slate-300" },
    },
  },
  entity: {
    machine: {
      icon: Settings2,
      colors: { base: "text-sky-400" },
    },
    universe: {
      icon: Box,
      colors: { base: "text-blue-400" },
    },
    reality: {
      icon: Circle,
      colors: { base: "text-green-400" },
    },
  },
  status: {
    warning: {
      icon: AlertTriangle,
      colors: {
        base: "text-red-300",
        muted: "text-red-200",
      },
    },
  },
};

export const STUDIO_ICONS = {
  reality: {
    normal: STUDIO_ICON_REGISTRY.reality.normal.icon,
    success: STUDIO_ICON_REGISTRY.reality.success.icon,
    error: STUDIO_ICON_REGISTRY.reality.error.icon,
    initial: STUDIO_ICON_REGISTRY.reality.initial.icon,
  } satisfies Record<RealityType | "initial", StudioIcon>,
  behavior: {
    action: STUDIO_ICON_REGISTRY.behavior.action.icon,
    invoke: STUDIO_ICON_REGISTRY.behavior.invoke.icon,
    condition: STUDIO_ICON_REGISTRY.behavior.condition.icon,
    observer: STUDIO_ICON_REGISTRY.behavior.observer.icon,
  } satisfies Record<BehaviorType, StudioIcon>,
  transition: {
    trigger: {
      on: STUDIO_ICON_REGISTRY.transition.trigger.on.icon,
      always: STUDIO_ICON_REGISTRY.transition.trigger.always.icon,
    },
    type: {
      default: STUDIO_ICON_REGISTRY.transition.type.default.icon,
      notify: STUDIO_ICON_REGISTRY.transition.type.notify.icon,
    },
  },
  phase: {
    entry: STUDIO_ICON_REGISTRY.phase.entry.icon,
    exit: STUDIO_ICON_REGISTRY.phase.exit.icon,
    onTransition: STUDIO_ICON_REGISTRY.phase.onTransition.icon,
  },
  entity: {
    machine: STUDIO_ICON_REGISTRY.entity.machine.icon,
    universe: STUDIO_ICON_REGISTRY.entity.universe.icon,
    reality: STUDIO_ICON_REGISTRY.entity.reality.icon,
  },
  status: {
    warning: STUDIO_ICON_REGISTRY.status.warning.icon,
  },
} as const;
