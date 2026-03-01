import type { EditorNode, EditorTransition, NodeSizeMap, TransitionLeg } from "../../types";
import { getTransitionLegGeometry } from "../../utils";

interface EdgeProps {
  leg: TransitionLeg;
  transition: EditorTransition;
  nodes: EditorNode[];
  nodeSizes: NodeSizeMap;
  selected: boolean;
  invalidNotify?: boolean;
  onSelect: () => void;
  onHover?: (isHovered: boolean) => void;
}

export const Edge = ({
  leg,
  transition,
  nodes,
  nodeSizes,
  selected,
  invalidNotify = false,
  onSelect,
  onHover,
}: EdgeProps) => {
  const geometry = getTransitionLegGeometry(leg, nodes, nodeSizes, transition);
  if (!geometry) {
    return null;
  }

  const { d } = geometry;
  return (
    <g>
      <path
        d={d}
        fill="none"
        stroke="transparent"
        strokeWidth="20"
        className="canvas-interactive cursor-pointer pointer-events-auto"
        onMouseEnter={() => onHover?.(true)}
        onMouseLeave={() => onHover?.(false)}
        onClick={(event) => {
          event.stopPropagation();
          onSelect();
        }}
      />
      <path
        d={d}
        fill="none"
        stroke={invalidNotify ? "#ef4444" : selected ? "#3b82f6" : "#94a3b8"}
        strokeWidth="2"
        markerEnd={selected ? "url(#arrowhead-selected)" : "url(#arrowhead)"}
        className="pointer-events-none transition-colors"
      />
    </g>
  );
};
