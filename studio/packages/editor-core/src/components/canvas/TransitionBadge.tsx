import { memo } from "react";

import { STUDIO_ICON_REGISTRY, STUDIO_ICONS } from "../../constants";
import { useI18n } from "../../i18n";
import type { EditorTransition } from "../../types";
import {
  BADGE_HEIGHT,
  BADGE_PORT_RADIUS,
  BADGE_RADIUS,
  BADGE_WIDTH,
  truncateTriggerLabel,
} from "../../utils";

interface TransitionBadgeProps {
  x: number;
  y: number;
  transition: EditorTransition;
  selected: boolean;
  alwaysOrderSummary?: string;
  invalidNotify?: boolean;
  issueCount?: number;
  onSelect: () => void;
  onEdit?: () => void;
  onMouseDown?: (event: React.MouseEvent<SVGGElement>) => void;
  onOutputPortMouseDown?: (event: React.MouseEvent<SVGCircleElement>) => void;
  onHover?: (isHovered: boolean) => void;
}

const TransitionBadgeComponent = ({
  x,
  y,
  transition,
  selected,
  alwaysOrderSummary,
  invalidNotify = false,
  issueCount = 0,
  onSelect,
  onEdit,
  onMouseDown,
  onOutputPortMouseDown,
  onHover,
}: TransitionBadgeProps) => {
  const { t } = useI18n();
  const isAlways = transition.triggerKind === "always";
  const isNotify = transition.type === "notify";
  const hasConditions = transition.conditions.length > 0;
  const hasEffects = transition.actions.length > 0 || transition.invokes.length > 0;
  const hasError = invalidNotify || issueCount > 0;
  const hasRightIndicators = hasError || hasConditions || hasEffects;

  const triggerName = isAlways
    ? t("editor.transition.trigger.always")
    : transition.eventName || t("editor.transition.trigger.newEvent");
  const triggerLabel = truncateTriggerLabel(
    isAlways && alwaysOrderSummary
      ? `${triggerName} ${alwaysOrderSummary}`
      : triggerName,
  );

  const TypeIcon = isNotify ? STUDIO_ICONS.transition.type.notify : STUDIO_ICONS.transition.type.default;
  const TriggerIcon = isAlways ? STUDIO_ICONS.transition.trigger.always : STUDIO_ICONS.transition.trigger.on;
  const ConditionIcon = STUDIO_ICONS.behavior.condition;
  const WarningIcon = STUDIO_ICONS.status.warning;

  const triggerOnColors = STUDIO_ICON_REGISTRY.transition.trigger.on.colors;
  const triggerAlwaysColors = STUDIO_ICON_REGISTRY.transition.trigger.always.colors;
  const transitionDefaultColors = STUDIO_ICON_REGISTRY.transition.type.default.colors;
  const transitionNotifyColors = STUDIO_ICON_REGISTRY.transition.type.notify.colors;
  const conditionColors = STUDIO_ICON_REGISTRY.behavior.condition.colors;
  const warningColors = STUDIO_ICON_REGISTRY.status.warning.colors;

  const fillColor = hasError ? "#450a0a" : selected ? "#1e3a8a" : "#1e293b";
  const strokeColor = hasError ? "#b91c1c" : selected ? "#3b82f6" : "#475569";
  const textColor = selected ? "#dbeafe" : "#e2e8f0";
  const ringColor = hasError ? "#fca5a5" : selected ? "#bfdbfe" : "#cbd5e1";
  const coreColor = hasError ? "#f87171" : selected ? "#60a5fa" : "#38bdf8";
  const leftX = x - BADGE_WIDTH / 2;
  const topY = y - BADGE_HEIGHT / 2;
  const portY = BADGE_HEIGHT / 2;

  return (
    <g
      data-testid={`transition-badge-${transition.id}`}
      className="canvas-interactive select-none cursor-grab active:cursor-grabbing"
      onMouseEnter={() => onHover?.(true)}
      onMouseLeave={() => onHover?.(false)}
      onMouseDown={(event) => {
        event.preventDefault();
        event.stopPropagation();
        onMouseDown?.(event);
      }}
      onClick={(event) => {
        event.stopPropagation();
        onSelect();
      }}
      onDoubleClick={(event) => {
        event.preventDefault();
        event.stopPropagation();
        onSelect();
        onEdit?.();
      }}
      transform={`translate(${leftX} ${topY})`}
      pointerEvents="all"
    >
      <rect
        x={0}
        y={0}
        width={BADGE_WIDTH}
        height={BADGE_HEIGHT}
        rx={BADGE_RADIUS}
        ry={BADGE_RADIUS}
        fill={fillColor}
        stroke={strokeColor}
        strokeWidth={selected ? 1.8 : 1.4}
      />

      <circle
        data-testid={`transition-badge-port-left-${transition.id}`}
        cx={0}
        cy={portY}
        r={BADGE_PORT_RADIUS}
        fill="#020617"
        stroke={ringColor}
        strokeWidth={1.8}
        className="pointer-events-none"
      />
      <circle cx={0} cy={portY} r={2} fill={coreColor} className="pointer-events-none" />

      <circle
        data-testid={`transition-badge-port-right-${transition.id}`}
        cx={BADGE_WIDTH}
        cy={portY}
        r={14}
        fill="transparent"
        stroke="transparent"
        strokeWidth={1.2}
        className="canvas-interactive cursor-alias active:cursor-alias hover:fill-sky-400/15 hover:stroke-sky-300/70 transition-colors"
        onMouseDown={(event) => {
          event.preventDefault();
          event.stopPropagation();
          onOutputPortMouseDown?.(event);
        }}
        onClick={(event) => {
          event.preventDefault();
          event.stopPropagation();
        }}
      />
      <circle
        cx={BADGE_WIDTH}
        cy={portY}
        r={BADGE_PORT_RADIUS}
        fill="#020617"
        stroke={ringColor}
        strokeWidth={1.8}
        className="pointer-events-none"
      />
      <circle cx={BADGE_WIDTH} cy={portY} r={2} fill={coreColor} className="pointer-events-none" />

      {hasRightIndicators && (
        <line
          x1={BADGE_WIDTH - 66}
          y1={7}
          x2={BADGE_WIDTH - 66}
          y2={BADGE_HEIGHT - 7}
          stroke={selected ? "#334155" : "#475569"}
          strokeWidth={1}
        />
      )}

      <g transform={`translate(10 ${portY - 6})`}>
        <TypeIcon
          size={12}
          className={
            hasError
              ? warningColors.base
              : isNotify
                ? transitionNotifyColors.base
                : transitionDefaultColors.base
          }
        />
      </g>

      <g transform={`translate(27 ${portY - 5})`}>
        <TriggerIcon
          size={10}
          className={isAlways ? triggerAlwaysColors.base : (triggerOnColors.accent ?? triggerOnColors.base)}
        />
      </g>

      <text
        x={40}
        y={portY + 3.4}
        fill={textColor}
        fontSize={10}
        fontWeight={700}
        fontFamily="ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, Courier New, monospace"
      >
        {triggerLabel}
      </text>

      {hasError && (
        <g transform={`translate(${BADGE_WIDTH - 60} ${portY - 5})`}>
          <WarningIcon size={10} className={warningColors.base} />
        </g>
      )}

      {hasConditions && (
        <g transform={`translate(${BADGE_WIDTH - 42} ${portY - 5})`}>
          <ConditionIcon size={10} className={conditionColors.base} />
          <text
            x={12}
            y={8.2}
            fill={selected ? "#fdba74" : "#fb923c"}
            fontSize={8.5}
            fontWeight={700}
            fontFamily="ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, Liberation Mono, Courier New, monospace"
          >
            {transition.conditions.length}
          </text>
        </g>
      )}

      {hasEffects && (
        <circle
          cx={BADGE_WIDTH - 16}
          cy={portY}
          r={3}
          fill="#22c55e"
          stroke="#15803d"
          strokeWidth={0.8}
        />
      )}
    </g>
  );
};

export const TransitionBadge = memo(TransitionBadgeComponent);
