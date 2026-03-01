import {
  ChevronUp,
  Maximize,
  Minus,
  Plus,
  Redo2,
  StickyNote,
  Target,
  Undo2,
} from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { useI18n } from "../../i18n";
import type { StudioUniverseTemplate } from "../../types";

interface CanvasToolbarProps {
  onAddUniverse: () => void;
  onAddUniverseFromTemplate?: (templateId: string) => void;
  universeTemplates?: StudioUniverseTemplate[];
  onAddGlobalNote?: () => void;
  onAutoLayout?: () => void;
  isAutoLayouting?: boolean;
  onUndo?: () => void;
  onRedo?: () => void;
  canUndo?: boolean;
  canRedo?: boolean;
  onZoomIn: () => void;
  onZoomOut: () => void;
  onFit: () => void;
  onCenter: () => void;
  zoom: number;
}

export const CanvasToolbar = ({
  onAddUniverse,
  onAddUniverseFromTemplate,
  universeTemplates = [],
  onAddGlobalNote,
  onAutoLayout,
  isAutoLayouting = false,
  onUndo,
  onRedo,
  canUndo = false,
  canRedo = false,
  onZoomIn,
  onZoomOut,
  onFit,
  onCenter,
  zoom,
}: CanvasToolbarProps) => {
  const { t } = useI18n();
  const [templatesMenuOpen, setTemplatesMenuOpen] = useState(false);
  const templateMenuRef = useRef<HTMLDivElement | null>(null);
  const canUseTemplates = universeTemplates.length > 0 && Boolean(onAddUniverseFromTemplate);

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      const container = templateMenuRef.current;
      if (!container) {
        return;
      }

      const path = event.composedPath();
      if (!path.includes(container)) {
        setTemplatesMenuOpen(false);
      }
    };

    document.addEventListener("pointerdown", handlePointerDown, true);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown, true);
    };
  }, []);

  useEffect(() => {
    if (!canUseTemplates && templatesMenuOpen) {
      setTemplatesMenuOpen(false);
    }
  }, [canUseTemplates, templatesMenuOpen]);

  return (
    <>
      <div className="fixed bottom-6 left-1/2 -translate-x-1/2 bg-slate-800 border border-slate-700 rounded-full p-1.5 shadow-2xl flex items-center gap-1 z-[60] animate-in slide-in-from-bottom-5">
        <button
          onClick={onAddUniverse}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-full text-white font-medium text-sm transition-colors shadow-inner"
        >
          <Plus size={16} /> {t("toolbar.addBlankUniverse")}
        </button>
        {onAddGlobalNote && (
          <>
            <div className="w-px h-6 bg-slate-700 mx-1" />
            <button
              onClick={onAddGlobalNote}
              className="flex items-center gap-2 px-4 py-2 bg-yellow-600/20 text-yellow-500 hover:bg-yellow-600/30 border border-yellow-500/30 hover:border-yellow-500/50 rounded-full font-medium text-sm transition-colors"
              title={t("toolbar.addGlobalNote")}
            >
              <StickyNote size={16} /> {t("editor.context.addNote")}
            </button>
          </>
        )}
        <div className="relative" ref={templateMenuRef}>
          <button
            type="button"
            onClick={() => {
              if (!canUseTemplates) {
                return;
              }
              setTemplatesMenuOpen((previous) => !previous);
            }}
            disabled={!canUseTemplates}
            className={`p-2 rounded-full transition-colors ${
              canUseTemplates
                ? "text-slate-300 hover:text-white hover:bg-slate-700"
                : "text-slate-400 cursor-not-allowed"
            }`}
            title={t("toolbar.prebuiltUniverses")}
          >
            <ChevronUp size={16} />
          </button>

          {templatesMenuOpen && canUseTemplates && (
            <div className="absolute bottom-14 left-1/2 -translate-x-1/2 w-72 bg-slate-900 border border-slate-700 rounded-lg shadow-2xl p-2 z-[80] max-h-72 overflow-y-auto custom-scrollbar">
              <div className="text-[10px] text-slate-400 font-mono uppercase tracking-wider px-2 py-1">
                {t("toolbar.prebuiltUniverses")}
              </div>
              {universeTemplates.map((template) => (
                <button
                  key={template.id}
                  type="button"
                  onClick={() => {
                    onAddUniverseFromTemplate?.(template.id);
                    setTemplatesMenuOpen(false);
                  }}
                  className="w-full text-left rounded-md px-3 py-2 hover:bg-slate-800 transition-colors"
                >
                  <div className="text-sm font-semibold text-slate-100">{template.label}</div>
                  {template.description && (
                    <div className="text-[11px] text-slate-400 mt-0.5 line-clamp-2">
                      {template.description}
                    </div>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="fixed bottom-6 right-8 bg-slate-900/95 backdrop-blur border border-slate-700 rounded-xl p-1.5 shadow-2xl flex items-center gap-1 z-[60] animate-in slide-in-from-bottom-4">
        <button
          onClick={onAutoLayout}
          disabled={!onAutoLayout || isAutoLayouting}
          className="px-3 py-2 text-xs font-semibold text-emerald-300 hover:text-emerald-200 hover:bg-emerald-500/20 border border-emerald-500/30 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          title={t("toolbar.autoLayout")}
          aria-label={t("toolbar.autoLayout")}
        >
          {isAutoLayouting ? t("toolbar.autoLayoutRunning") : t("toolbar.autoLayout")}
        </button>

        <div className="w-px h-5 bg-slate-700 mx-0.5" />

        <button
          onClick={onUndo}
          disabled={!onUndo || !canUndo}
          className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          title={t("toolbar.undo")}
          aria-label={t("toolbar.undo")}
        >
          <Undo2 size={16} />
        </button>

        <button
          onClick={onRedo}
          disabled={!onRedo || !canRedo}
          className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          title={t("toolbar.redo")}
          aria-label={t("toolbar.redo")}
        >
          <Redo2 size={16} />
        </button>

        <div className="w-px h-5 bg-slate-700 mx-0.5" />

        <button
          onClick={onZoomOut}
          className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
          title={t("toolbar.zoomOut")}
          aria-label={t("toolbar.zoomOut")}
        >
          <Minus size={16} />
        </button>

        <button
          onClick={onZoomIn}
          className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
          title={t("toolbar.zoomIn")}
          aria-label={t("toolbar.zoomIn")}
        >
          <Plus size={16} />
        </button>

        <span className="px-2 py-1 text-[11px] font-mono font-semibold text-slate-300 min-w-[52px] text-center border border-slate-700 rounded-md bg-slate-950/70">
          {Math.round(zoom * 100)}%
        </span>

        <div className="w-px h-5 bg-slate-700 mx-0.5" />

        <button
          onClick={onFit}
          className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
          title={t("toolbar.fitContent")}
          aria-label={t("toolbar.fitContent")}
        >
          <Maximize size={16} />
        </button>

        <button
          onClick={onCenter}
          className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
          title={t("toolbar.centerContent")}
          aria-label={t("toolbar.centerContent")}
        >
          <Target size={16} />
        </button>
      </div>
    </>
  );
};
