import type { ButtonHTMLAttributes } from "react";

import { StudioTooltip, type TooltipSide, type TooltipWidth } from "./StudioTooltip";

export interface TooltipIconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  tooltip: string;
  side?: TooltipSide;
  tooltipWidth?: TooltipWidth;
}

export const TooltipIconButton = ({
  tooltip,
  side,
  tooltipWidth,
  type = "button",
  className,
  ["aria-label"]: ariaLabel,
  ...buttonProps
}: TooltipIconButtonProps) => {
  return (
    <StudioTooltip label={tooltip} side={side} width={tooltipWidth}>
      <button type={type} className={className} aria-label={ariaLabel || tooltip} {...buttonProps} />
    </StudioTooltip>
  );
};
