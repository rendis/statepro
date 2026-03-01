import { ChevronDown, ChevronUp, StickyNote, X } from "lucide-react";
import { memo } from "react";
import type { MouseEvent } from "react";
import type { CSSProperties } from "react";

import type { EditorNode, GlobalNoteData } from "../../../types";
import { useI18n } from "../../../i18n";
import { TooltipIconButton } from "../../shared";
import { clampNoteColorIndex, NOTE_COLORS } from "./noteColors";

interface GlobalNoteNodeProps {
  node: Extract<EditorNode, { type: "note" }>;
  selected: boolean;
  onMouseDown: (
    event: MouseEvent<HTMLDivElement>,
    node: Extract<EditorNode, { type: "note" }>,
  ) => void;
  onUpdate: (id: string, data: GlobalNoteData) => void;
  onDelete: (id: string) => void;
}

const GlobalNoteNodeComponent = ({
  node,
  selected,
  onMouseDown,
  onUpdate,
  onDelete,
}: GlobalNoteNodeProps) => {
  const { t } = useI18n();
  const color = NOTE_COLORS[clampNoteColorIndex(node.data.colorIndex)];
  const isCollapsed = Boolean(node.data.isCollapsed);
  const scrollbarStyle = {
    "--note-scrollbar-thumb": color.scrollbarThumb,
  } as CSSProperties;

  return (
    <div
      data-testid={`global-note-node-${node.id}`}
      style={{ left: node.x, top: node.y }}
      className={`canvas-interactive absolute rounded-lg shadow-xl border flex flex-col overflow-hidden transition-all ${isCollapsed ? "w-56" : "min-w-[224px]"} ${color.bg} ${color.border} ${selected ? "shadow-2xl ring-2 ring-blue-500 z-50 scale-[1.02]" : "z-30"}`}
      onMouseDown={(event) => onMouseDown(event, node)}
    >
      <div
        className="flex items-center justify-between px-3 py-2 border-b border-black/10 cursor-move bg-black/5"
        onDoubleClick={(event) => event.stopPropagation()}
      >
        <div className="flex gap-1.5 pointer-events-auto">
          {NOTE_COLORS.map((token, index) => (
            <TooltipIconButton
              key={token.bg}
              onMouseDown={(event) => event.stopPropagation()}
              onClick={() =>
                onUpdate(node.id, {
                  ...node.data,
                  colorIndex: index,
                })
              }
              tooltip={t("note.color", { index: index + 1 })}
              className={`w-3.5 h-3.5 rounded-full border border-black/20 hover:scale-110 ${token.bg} ${clampNoteColorIndex(node.data.colorIndex) === index ? "ring-2 ring-black/50 ring-offset-1 ring-offset-transparent" : ""}`}
            />
          ))}
        </div>

        <div className="flex items-center gap-1 pointer-events-auto">
          <TooltipIconButton
            onMouseDown={(event) => event.stopPropagation()}
            onClick={() =>
              onUpdate(node.id, {
                ...node.data,
                isCollapsed: !isCollapsed,
              })
            }
            tooltip={isCollapsed ? t("note.expand") : t("note.collapse")}
            className="p-1 hover:bg-black/10 rounded text-black/50 hover:text-black transition-colors"
          >
            {isCollapsed ? <ChevronDown size={14} /> : <ChevronUp size={14} />}
          </TooltipIconButton>
          <TooltipIconButton
            onMouseDown={(event) => event.stopPropagation()}
            onClick={() => onDelete(node.id)}
            tooltip={t("note.delete")}
            className="p-1 hover:bg-black/10 rounded text-black/50 hover:text-red-600 transition-colors"
          >
            <X size={14} />
          </TooltipIconButton>
        </div>
      </div>

      {!isCollapsed && (
        <textarea
          value={node.data.text || ""}
          onMouseDown={(event) => event.stopPropagation()}
          onChange={(event) =>
            onUpdate(node.id, {
              ...node.data,
              text: event.target.value,
            })
          }
          placeholder={t("note.globalPlaceholder")}
          style={scrollbarStyle}
          className={`note-scrollbar w-full min-h-[120px] p-3 bg-transparent resize focus:outline-none text-sm font-medium leading-relaxed pointer-events-auto ${color.text} ${color.placeholder}`}
        />
      )}

      {isCollapsed && node.data.text && (
        <div className={`px-3 py-2 text-xs truncate opacity-70 font-medium pointer-events-none ${color.text}`}>
          <StickyNote size={12} className="inline mr-1" />
          {node.data.text}
        </div>
      )}
    </div>
  );
};

export const GlobalNoteNode = memo(GlobalNoteNodeComponent);
