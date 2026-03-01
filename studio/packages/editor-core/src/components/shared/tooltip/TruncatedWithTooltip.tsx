import { StudioTooltip, type TooltipSide } from "./StudioTooltip";

export interface TruncatedWithTooltipProps {
  text: string;
  className?: string;
  side?: TooltipSide;
}

export const TruncatedWithTooltip = ({
  text,
  className,
  side,
}: TruncatedWithTooltipProps) => {
  if (!text) {
    return <div className={className}>{text}</div>;
  }

  return (
    <StudioTooltip label={text} side={side} width="wrap">
      <div className={className}>{text}</div>
    </StudioTooltip>
  );
};
