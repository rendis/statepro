import type { ReactNode } from "react";

import { StudioTooltip, type TooltipSide } from "./StudioTooltip";

export interface TruncatedWithTooltipProps {
  text: string;
  tooltipLabel?: ReactNode;
  className?: string;
  side?: TooltipSide;
}

export const TruncatedWithTooltip = ({
  text,
  tooltipLabel,
  className,
  side,
}: TruncatedWithTooltipProps) => {
  if (!text) {
    return <div className={className}>{text}</div>;
  }

  return (
    <StudioTooltip label={tooltipLabel || text} side={side} width="wrap">
      <div className={className}>{text}</div>
    </StudioTooltip>
  );
};
