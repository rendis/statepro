import { StickyNote, Trash2, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { CSSProperties } from "react";

import type { AnchoredNoteData } from "../../../types";
import { useI18n } from "../../../i18n";
import { clampNoteColorIndex, NOTE_COLORS } from "./noteColors";

interface AnchoredNoteProps {
  note: AnchoredNoteData;
  isFocused: boolean;
  onRequestFocus?: () => void;
  onUpdate: (next: AnchoredNoteData) => void;
  onDelete: () => void;
  styleClass?: string;
}

export const AnchoredNote = ({
  note,
  isFocused,
  onRequestFocus,
  onUpdate,
  onDelete,
  styleClass,
}: AnchoredNoteProps) => {
  const { t } = useI18n();
  const [isExpanded, setIsExpanded] = useState(() => !note.text);
  const ref = useRef<HTMLDivElement | null>(null);
  const color = NOTE_COLORS[clampNoteColorIndex(note.colorIndex)];
  const scrollbarStyle = {
    "--note-scrollbar-thumb": color.scrollbarThumb,
  } as CSSProperties;

  useEffect(() => {
    if (!isFocused) {
      setIsExpanded(false);
    }
  }, [isFocused]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (!ref.current || ref.current.contains(event.target as Node)) {
        return;
      }
      setIsExpanded(false);
    };

    if (isExpanded) {
      document.addEventListener("mousedown", handleClickOutside, true);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside, true);
    };
  }, [isExpanded]);

  if (!isExpanded) {
    return (
      <div
        onMouseDown={(event) => {
          event.stopPropagation();
          onRequestFocus?.();
        }}
        onClick={(event) => {
          event.stopPropagation();
          onRequestFocus?.();
          setIsExpanded(true);
        }}
        className={`w-6 h-6 rounded shadow-md cursor-pointer flex items-center justify-center border-b border-r hover:scale-110 transition-transform ${color.bg} ${color.border} ${styleClass || ""}`}
        title={t("note.view")}
      >
        <StickyNote size={12} className={color.text} />
        {note.text && (
          <div className="absolute -top-1 -right-1 w-2 h-2 bg-red-500 rounded-full border border-slate-900" />
        )}
      </div>
    );
  }

  return (
    <div
      ref={ref}
      onMouseDown={(event) => {
        event.stopPropagation();
        onRequestFocus?.();
      }}
      className={`w-48 rounded-lg shadow-2xl border flex flex-col overflow-hidden animate-in zoom-in-95 duration-150 ${color.bg} ${color.border} ${styleClass || ""}`}
    >
      <div className="flex items-center justify-between px-2 py-1.5 border-b border-black/10">
        <div className="flex gap-1">
          {NOTE_COLORS.map((token, index) => (
            <button
              key={token.bg}
              onClick={() =>
                onUpdate({
                  ...note,
                  colorIndex: index,
                })
              }
              className={`w-3 h-3 rounded-full border border-black/20 hover:scale-110 ${token.bg} ${clampNoteColorIndex(note.colorIndex) === index ? "ring-1 ring-black ring-offset-1 ring-offset-transparent" : ""}`}
              title={t("note.color", { index: index + 1 })}
            />
          ))}
        </div>

        <div className="flex items-center gap-1">
          <button
            onClick={onDelete}
            className="p-0.5 hover:bg-black/10 rounded text-black/50 hover:text-red-600 transition-colors"
            title={t("note.deleteAnchored")}
          >
            <Trash2 size={12} />
          </button>
          <button
            onClick={() => setIsExpanded(false)}
            className="p-0.5 hover:bg-black/10 rounded text-black/50 hover:text-black transition-colors"
            title={t("note.collapseAnchored")}
          >
            <X size={12} />
          </button>
        </div>
      </div>

      <textarea
        autoFocus
        value={note.text || ""}
        onChange={(event) =>
          onUpdate({
            ...note,
            text: event.target.value,
          })
        }
        placeholder={t("note.anchoredPlaceholder")}
        style={scrollbarStyle}
        className={`note-scrollbar w-full h-32 p-2 bg-transparent resize-none focus:outline-none text-xs font-medium leading-relaxed ${color.text} ${color.placeholder}`}
      />
    </div>
  );
};
