import type { ReactNode } from "react";

interface CanvasProps {
  containerRef: React.RefObject<HTMLDivElement | null>;
  canvasRef: React.RefObject<HTMLDivElement | null>;
  onMouseMove: (event: React.MouseEvent<HTMLDivElement>) => void;
  onMouseUp: () => void;
  onClick: (event: React.MouseEvent<HTMLDivElement>) => void;
  svgChildren?: ReactNode;
  children: ReactNode;
}

export const Canvas = ({
  containerRef,
  canvasRef,
  onMouseMove,
  onMouseUp,
  onClick,
  svgChildren,
  children,
}: CanvasProps) => {
  return (
    <main
      ref={containerRef}
      className="w-full h-full relative bg-slate-950 overflow-auto scroll-smooth"
    >
      <div
        ref={canvasRef}
        className="relative w-[3000px] h-[3000px] origin-top-left"
        style={{
          backgroundImage: "radial-gradient(circle, #334155 1px, transparent 1px)",
          backgroundSize: "20px 20px",
        }}
        onMouseMove={onMouseMove}
        onMouseUp={onMouseUp}
        onClick={onClick}
      >
        <svg className="absolute top-0 left-0 w-full h-full pointer-events-none z-10">
          <defs>
            <marker
              id="arrowhead"
              markerWidth="10"
              markerHeight="10"
              refX="9"
              refY="5"
              orient="auto"
              markerUnits="userSpaceOnUse"
            >
              <path d="M 0 0 L 10 5 L 0 10 z" fill="#94a3b8" />
            </marker>
            <marker
              id="arrowhead-selected"
              markerWidth="10"
              markerHeight="10"
              refX="9"
              refY="5"
              orient="auto"
              markerUnits="userSpaceOnUse"
            >
              <path d="M 0 0 L 10 5 L 0 10 z" fill="#3b82f6" />
            </marker>
          </defs>
          {svgChildren}
        </svg>

        {children}
      </div>
    </main>
  );
};
