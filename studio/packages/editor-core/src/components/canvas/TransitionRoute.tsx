import { memo, useMemo, type MouseEvent } from "react";
import type { EditorNode, EditorTransition, NodeSizeMap, TransitionLeg } from "../../types";
import { getTransitionRouteGeometry, type TransitionRouteGeometry } from "../../utils";

export type TransitionRouteRenderMode = "full" | "skeleton";

interface TransitionRouteProps {
  transition: EditorTransition;
  legs: TransitionLeg[];
  nodes: EditorNode[];
  nodeSizes: NodeSizeMap;
  routeGeometry?: TransitionRouteGeometry | null;
  renderMode?: TransitionRouteRenderMode;
  selected: boolean;
  invalidNotify?: boolean;
  onSelect: () => void;
  onOutputPortMouseDown?: (event: MouseEvent<SVGCircleElement>) => void;
  onHover?: (isHovered: boolean) => void;
}

const TransitionRouteComponent = ({
  transition,
  legs,
  nodes,
  nodeSizes,
  routeGeometry,
  renderMode = "full",
  selected,
  invalidNotify = false,
  onSelect,
  onOutputPortMouseDown,
  onHover,
}: TransitionRouteProps) => {
  const geometry = useMemo(() => {
    if (routeGeometry) {
      return routeGeometry;
    }
    const nodeLookup = new Map(nodes.map((node) => [node.id, node]));
    return getTransitionRouteGeometry(transition, legs, nodes, nodeSizes, nodeLookup);
  }, [routeGeometry, transition, legs, nodes, nodeSizes]);
  if (!geometry) {
    return null;
  }

  const isSkeleton = renderMode === "skeleton";
  const stroke = invalidNotify ? "#ef4444" : selected ? "#3b82f6" : "#94a3b8";
  const strokeWidth = isSkeleton ? 1.25 : 2;

  const handleSelect = (event: MouseEvent<SVGElement>) => {
    event.stopPropagation();
    onSelect();
  };

  return (
    <g>
      {geometry.segments.map((segment) => (
        <g key={segment.id}>
          {!isSkeleton && (
            <path
              d={segment.d}
              fill="none"
              stroke="transparent"
              strokeWidth="20"
              data-segment-role={segment.role}
              data-testid={`transition-segment-hit-${segment.id}`}
              className="canvas-interactive cursor-pointer pointer-events-auto"
              onMouseEnter={() => onHover?.(true)}
              onMouseLeave={() => onHover?.(false)}
              onClick={handleSelect}
            />
          )}
          <path
            d={segment.d}
            fill="none"
            stroke={stroke}
            strokeWidth={strokeWidth}
            data-segment-role={segment.role}
            data-testid={`transition-segment-${segment.id}`}
            markerEnd={
              !isSkeleton && segment.hasArrow
                ? selected
                  ? "url(#arrowhead-selected)"
                  : "url(#arrowhead)"
                : undefined
            }
            className={`pointer-events-none ${
              isSkeleton
                ? "studio-transition-skeleton"
                : "transition-colors"
            }`}
          />
        </g>
      ))}
      {!isSkeleton && (
        <>
          <circle
            cx={geometry.leftPort.x}
            cy={geometry.leftPort.y}
            r={12}
            fill="transparent"
            className="canvas-interactive cursor-pointer pointer-events-auto"
            data-testid={`transition-port-left-hit-${transition.id}`}
            onMouseEnter={() => onHover?.(true)}
            onMouseLeave={() => onHover?.(false)}
            onClick={handleSelect}
          />
          <circle
            cx={geometry.rightPort.x}
            cy={geometry.rightPort.y}
            r={16}
            fill="transparent"
            strokeWidth={1.2}
            className="canvas-interactive cursor-alias active:cursor-alias pointer-events-auto stroke-transparent hover:fill-sky-400/15 hover:stroke-sky-300/70 transition-colors"
            data-testid={`transition-port-right-hit-${transition.id}`}
            onMouseDown={(event) => {
              event.preventDefault();
              event.stopPropagation();
              onOutputPortMouseDown?.(event);
            }}
            onMouseEnter={() => onHover?.(true)}
            onMouseLeave={() => onHover?.(false)}
            onClick={handleSelect}
          />
        </>
      )}
    </g>
  );
};

const areTransitionRoutePropsEqual = (
  previous: TransitionRouteProps,
  next: TransitionRouteProps,
): boolean => {
  if (!previous.routeGeometry || !next.routeGeometry) {
    return false;
  }

  return (
    previous.routeGeometry === next.routeGeometry &&
    previous.renderMode === next.renderMode &&
    previous.selected === next.selected &&
    previous.invalidNotify === next.invalidNotify &&
    previous.onSelect === next.onSelect &&
    previous.onOutputPortMouseDown === next.onOutputPortMouseDown &&
    previous.onHover === next.onHover &&
    previous.transition.id === next.transition.id
  );
};

export const TransitionRoute = memo(TransitionRouteComponent, areTransitionRoutePropsEqual);
