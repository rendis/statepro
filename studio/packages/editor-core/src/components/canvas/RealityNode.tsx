import { Copy, Settings2, StickyNote, Trash2 } from "lucide-react";
import { useEffect, useRef } from "react";
import type { Dispatch, MouseEvent, SetStateAction } from "react";

import {
  REALITY_TYPE_LABEL_KEYS,
  REALITY_TYPES,
  STUDIO_ICON_REGISTRY,
  STUDIO_ICONS,
} from "../../constants";
import { useI18n } from "../../i18n";
import type { NodeSizeMap, RealityType } from "../../types";

interface RealityNodeProps {
  node: Extract<import("../../types").EditorNode, { type: "reality" }>;
  selected: boolean;
  searchPulseClassName?: string;
  issueCount?: number;
  onMouseDown: (
    event: MouseEvent<HTMLDivElement>,
    node: Extract<import("../../types").EditorNode, { type: "reality" }>,
  ) => void;
  onPortMouseDown: (
    event: MouseEvent<HTMLDivElement>,
    nodeId: string,
    type: "source" | "target",
  ) => void;
  onPortMouseUp: (
    event: MouseEvent<HTMLDivElement>,
    nodeId: string,
    type: "source" | "target",
  ) => void;
  setNodeSizes: Dispatch<SetStateAction<NodeSizeMap>>;
  onTypeChange: (id: string, type: RealityType) => void;
  onClone: (realityId: string) => void;
  hasNote?: boolean;
  onCreateNote?: (realityId: string) => void;
  onDelete: () => void;
  onEdit: () => void;
}

export const RealityNode = ({
  node,
  selected,
  searchPulseClassName,
  issueCount = 0,
  onMouseDown,
  onPortMouseDown,
  onPortMouseUp,
  setNodeSizes,
  onTypeChange,
  onClone,
  hasNote = false,
  onCreateNote,
  onDelete,
  onEdit,
}: RealityNodeProps) => {
  const { t } = useI18n();
  const ref = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!ref.current) return;

    const observer = new ResizeObserver(() => {
      setNodeSizes((previous) => {
        if (!ref.current) return previous;

        if (
          previous[node.id]?.w === ref.current.offsetWidth &&
          previous[node.id]?.h === ref.current.offsetHeight
        ) {
          return previous;
        }

        return {
          ...previous,
          [node.id]: {
            w: ref.current.offsetWidth,
            h: ref.current.offsetHeight,
          },
        };
      });
    });

    observer.observe(ref.current);
    return () => observer.disconnect();
  }, [node.id, setNodeSizes]);

  const typeConfig = REALITY_TYPES[node.data.realityType || "normal"];
  const IconComponent = typeConfig.icon;
  const InitialIcon = STUDIO_ICONS.reality.initial;
  const ObserverIcon = STUDIO_ICONS.behavior.observer;
  const ActionIcon = STUDIO_ICONS.behavior.action;
  const ExitPhaseIcon = STUDIO_ICONS.phase.exit;
  const initialColorClass = STUDIO_ICON_REGISTRY.reality.initial.colors.base;
  const observerColorClass = STUDIO_ICON_REGISTRY.behavior.observer.colors.base;
  const actionColorClass = STUDIO_ICON_REGISTRY.behavior.action.colors.base;
  const exitPhaseColorClass = STUDIO_ICON_REGISTRY.phase.exit.colors.base;
  const hasLifecycleSummary =
    (node.data.observers?.length || 0) > 0 ||
    (node.data.entryActions?.length || 0) > 0 ||
    (node.data.exitActions?.length || 0) > 0;
  const hasErrors = issueCount > 0;

  return (
    <div
      ref={ref}
      style={{ left: node.x, top: node.y }}
      className={`canvas-interactive absolute bg-slate-800 border-2 rounded-lg shadow-xl w-48 transition-colors ${
        selected
          ? `${hasErrors ? "border-red-500" : typeConfig.border} z-40`
          : `${hasErrors ? "border-red-700" : "border-slate-600"} z-20`
      } ${node.data.isInitial ? "ring-2 ring-blue-500 ring-offset-2 ring-offset-slate-900" : ""} ${searchPulseClassName || ""}`}
      onMouseDown={(event) => onMouseDown(event, node)}
    >
      {hasErrors && (
        <div className="absolute -top-2 -right-2 min-w-5 h-5 px-1 rounded-full bg-red-600 border-2 border-slate-900 flex items-center justify-center shadow-lg shadow-red-900/40 z-50 pointer-events-none">
          <span className="text-[10px] font-bold text-white">{issueCount}</span>
        </div>
      )}

      {node.data.isInitial && (
        <div className="absolute -top-2 -left-2 w-6 h-6 rounded-full bg-blue-600 border-2 border-slate-900 flex items-center justify-center shadow-lg shadow-blue-900/40 z-50 pointer-events-none">
          <InitialIcon size={12} className={initialColorClass} />
        </div>
      )}

      {selected && (
        <div className="absolute -top-12 left-0 bg-slate-800 border border-slate-700 rounded-lg shadow-xl flex items-center p-1 gap-1 z-50 animate-in slide-in-from-bottom-2 fade-in duration-200">
          <button
            onMouseDown={(event) => {
              event.stopPropagation();
              onEdit();
            }}
            className="p-1.5 hover:bg-slate-700 rounded text-slate-300 hover:text-white transition-colors"
            title={t("reality.editProperties")}
          >
            <Settings2 size={14} />
          </button>
          <div className="w-px h-4 bg-slate-700 mx-1" />
          <div className="relative group flex items-center h-full">
            <button
              className={`p-1.5 hover:bg-slate-700 rounded transition-colors ${typeConfig.color}`}
              title={t("reality.changeType")}
            >
              <IconComponent size={14} />
            </button>
            <div className="absolute top-full left-0 pt-1 hidden group-hover:block z-50">
              <div className="bg-slate-800 border border-slate-700 rounded shadow-xl p-1 gap-1 flex">
                {(Object.entries(REALITY_TYPES) as Array<[RealityType, (typeof REALITY_TYPES)[RealityType]]>).map(
                  ([typeKey, configType]) => {
                    const SubIcon = configType.icon;
                    return (
                      <button
                        key={typeKey}
                        onMouseDown={(event) => {
                          event.stopPropagation();
                          onTypeChange(node.id, typeKey);
                        }}
                        className={`p-1.5 hover:bg-slate-700 rounded ${configType.color}`}
                        title={t(
                          REALITY_TYPE_LABEL_KEYS[
                            typeKey as keyof typeof REALITY_TYPE_LABEL_KEYS
                          ] as string,
                          undefined,
                          configType.label,
                        )}
                      >
                        <SubIcon size={14} />
                      </button>
                    );
                  },
                )}
              </div>
            </div>
          </div>
          <div className="w-px h-4 bg-slate-700 mx-1" />
          {!hasNote && onCreateNote && (
            <>
              <button
                onMouseDown={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                  onCreateNote?.(node.id);
                }}
                className="p-1.5 hover:bg-yellow-900/30 rounded text-yellow-500 hover:text-yellow-400 transition-colors"
                title={t("reality.addNote")}
              >
                <StickyNote size={14} />
              </button>
              <div className="w-px h-4 bg-slate-700 mx-1" />
            </>
          )}
          <button
            onMouseDown={(event) => {
              event.stopPropagation();
              onClone(node.id);
            }}
            className="p-1.5 hover:bg-slate-700 rounded text-slate-300 hover:text-white transition-colors"
            title={t("reality.clone")}
          >
            <Copy size={14} />
          </button>
          <div className="w-px h-4 bg-slate-700 mx-1" />
          <button
            onMouseDown={(event) => {
              event.stopPropagation();
              onDelete();
            }}
            className="p-1.5 hover:bg-red-900/50 rounded text-red-400 hover:text-red-300 transition-colors"
            title={t("reality.delete")}
          >
            <Trash2 size={14} />
          </button>
        </div>
      )}

      <div
        data-testid={`reality-target-port-${node.id}`}
        className="canvas-interactive group/port absolute top-1/2 -left-5 -translate-y-1/2 w-8 h-8 rounded-full cursor-alias active:cursor-alias z-40"
        onMouseDown={(event) => {
          event.preventDefault();
          event.stopPropagation();
        }}
        onMouseUp={(event) => onPortMouseUp(event, node.id, "target")}
      >
        <div className="absolute inset-0 rounded-full bg-sky-400/0 border border-sky-300/0 group-hover/port:bg-sky-400/15 group-hover/port:border-sky-300/70 transition-colors pointer-events-none" />
        <div className="absolute top-1/2 left-1/2 w-4 h-4 -translate-x-1/2 -translate-y-1/2 bg-slate-400 border-2 border-slate-800 rounded-full group-hover/port:bg-slate-300 transition-colors pointer-events-none" />
      </div>

      <div className={`p-3 pointer-events-none select-none ${hasLifecycleSummary ? "" : "min-h-[56px] flex items-center"}`}>
        <div className={hasLifecycleSummary ? "" : "w-full"}>
          <div className={`flex items-center gap-2 ${hasLifecycleSummary ? "mb-2" : ""}`}>
            <IconComponent size={16} className={typeConfig.color} />
            <div className="font-mono font-bold text-slate-100 truncate select-none" title={node.data.id || node.data.name}>
              {node.data.id || node.data.name}
            </div>
          </div>

          {hasLifecycleSummary && (
            <div className="flex flex-col gap-1 mt-3">
              {node.data.observers && node.data.observers.length > 0 && (
                <div className="text-[10px] text-slate-400 flex items-center gap-1 bg-slate-900 px-2 py-1 rounded">
                  <ObserverIcon size={10} className={observerColorClass} />{" "}
                  {t("reality.lifecycle.observers")}: {node.data.observers.length}
                </div>
              )}
              {node.data.entryActions && node.data.entryActions.length > 0 && (
                <div className="text-[10px] text-slate-400 flex items-center gap-1 bg-slate-900 px-2 py-1 rounded">
                  <ActionIcon size={10} className={actionColorClass} /> {t("reality.lifecycle.entry")}:{" "}
                  {node.data.entryActions.length}
                </div>
              )}
              {node.data.exitActions && node.data.exitActions.length > 0 && (
                <div className="text-[10px] text-slate-400 flex items-center gap-1 bg-slate-900 px-2 py-1 rounded">
                  <ExitPhaseIcon size={10} className={exitPhaseColorClass} /> {t("reality.lifecycle.exit")}:{" "}
                  {node.data.exitActions.length}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      <div
        data-testid={`reality-source-port-${node.id}`}
        className="canvas-interactive group/port absolute top-1/2 -right-5 -translate-y-1/2 w-8 h-8 rounded-full cursor-alias active:cursor-alias z-40"
        onMouseDown={(event) => onPortMouseDown(event, node.id, "source")}
      >
        <div className="absolute inset-0 rounded-full bg-sky-400/0 border border-sky-300/0 group-hover/port:bg-sky-400/15 group-hover/port:border-sky-300/70 transition-colors pointer-events-none" />
        <div
          className={`absolute top-1/2 left-1/2 w-4 h-4 -translate-x-1/2 -translate-y-1/2 ${typeConfig.bg} border-2 border-slate-800 rounded-full group-hover/port:brightness-125 transition pointer-events-none`}
        />
      </div>
    </div>
  );
};
