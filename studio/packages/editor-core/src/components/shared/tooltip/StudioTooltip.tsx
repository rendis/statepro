import {
  autoUpdate,
  flip,
  FloatingPortal,
  offset,
  shift,
  useFloating,
  type Placement,
} from "@floating-ui/react";
import { useState, type FocusEvent, type ReactNode } from "react";

export type TooltipSide = "top" | "bottom" | "left" | "right";
export type TooltipWidth = "fit" | "wrap";

export interface StudioTooltipProps {
  label: ReactNode;
  side?: TooltipSide;
  width?: TooltipWidth;
  portal?: boolean;
  containerClassName?: string;
  bubbleClassName?: string;
  children: ReactNode;
}

const TOOLTIP_SIDE_TO_PLACEMENT: Record<TooltipSide, Placement> = {
  top: "top",
  bottom: "bottom",
  left: "left",
  right: "right",
};

const TOOLTIP_SIDE_CLASSES: Record<TooltipSide, string> = {
  top: "bottom-full left-1/2 -translate-x-1/2 mb-2",
  bottom: "top-full left-1/2 -translate-x-1/2 mt-2",
  left: "right-full top-1/2 -translate-y-1/2 mr-2",
  right: "left-full top-1/2 -translate-y-1/2 ml-2",
};

const TOOLTIP_WIDTH_CLASSES: Record<TooltipWidth, string> = {
  fit: "px-2 py-1 whitespace-nowrap",
  wrap: "px-2 py-1 max-w-96 whitespace-normal break-words leading-4",
};

const TOOLTIP_Z_INDEX = 450;

export const StudioTooltip = ({
  label,
  side = "top",
  width = "fit",
  portal = false,
  containerClassName,
  bubbleClassName,
  children,
}: StudioTooltipProps) => {
  const [isVisible, setIsVisible] = useState(false);

  const { refs, floatingStyles, context } = useFloating({
    open: isVisible,
    onOpenChange: setIsVisible,
    placement: TOOLTIP_SIDE_TO_PLACEMENT[side],
    strategy: "fixed",
    whileElementsMounted: autoUpdate,
    middleware: [offset(8), flip({ padding: 8 }), shift({ padding: 8 })],
  });

  const handleBlurCapture = (event: FocusEvent<HTMLDivElement>) => {
    const nextTarget = event.relatedTarget;
    if (nextTarget instanceof Node && event.currentTarget.contains(nextTarget)) {
      return;
    }
    setIsVisible(false);
  };

  const tooltipBubble = (
    <div
      ref={refs.setFloating}
      data-side={side}
      role="tooltip"
      className="pointer-events-none z-[450]"
      style={{ ...floatingStyles, zIndex: TOOLTIP_Z_INDEX }}
    >
      <div
        className={`${TOOLTIP_WIDTH_CLASSES[width]} rounded-md text-[11px] font-medium text-slate-100 bg-slate-900 border border-slate-700 shadow-xl ${bubbleClassName || ""}`}
      >
        {label}
      </div>
    </div>
  );

  const inlineBubble = (
    <div
      role="tooltip"
      data-side={side}
      className={`pointer-events-none absolute z-[450] ${TOOLTIP_SIDE_CLASSES[side]} animate-in fade-in duration-150`}
      style={{ zIndex: TOOLTIP_Z_INDEX }}
    >
      <div
        className={`${TOOLTIP_WIDTH_CLASSES[width]} rounded-md text-[11px] font-medium text-slate-100 bg-slate-900 border border-slate-700 shadow-xl ${bubbleClassName || ""}`}
      >
        {label}
      </div>
    </div>
  );

  return (
    <div
      ref={portal ? refs.setReference : undefined}
      className={`relative inline-flex ${containerClassName || ""}`}
      onMouseEnter={() => setIsVisible(true)}
      onMouseLeave={() => setIsVisible(false)}
      onFocusCapture={() => setIsVisible(true)}
      onBlurCapture={handleBlurCapture}
    >
      {children}
      {isVisible && (portal ? <FloatingPortal>{tooltipBubble}</FloatingPortal> : inlineBubble)}
    </div>
  );
};
