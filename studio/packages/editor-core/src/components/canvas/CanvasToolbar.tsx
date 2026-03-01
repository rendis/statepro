import {
  ChevronUp,
  Maximize,
  Minus,
  Plus,
  Redo2,
  Search,
  StickyNote,
  Tag,
  Target,
  Undo2,
  X,
} from "lucide-react";
import { useEffect, useRef, useState, type KeyboardEvent as ReactKeyboardEvent } from "react";

import { STUDIO_ICON_REGISTRY, STUDIO_ICONS } from "../../constants";
import { useI18n } from "../../i18n";
import type { StudioUniverseTemplate } from "../../types";
import type { CanvasSearchFilters, CanvasSearchResult } from "../../utils";

export interface CanvasToolbarProps {
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
  searchQuery?: string;
  searchResults?: CanvasSearchResult[];
  searchFilters?: CanvasSearchFilters;
  searchActiveIndex?: number;
  onSearchQueryChange?: (query: string) => void;
  onSearchFiltersChange?: (next: CanvasSearchFilters) => void;
  onSearchMoveSelection?: (direction: "up" | "down") => void;
  onSearchSelect?: (index: number) => void;
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
  searchQuery = "",
  searchResults = [],
  searchFilters = {
    universe: true,
    tag: true,
    reality: true,
  },
  searchActiveIndex = -1,
  onSearchQueryChange,
  onSearchFiltersChange,
  onSearchMoveSelection,
  onSearchSelect,
}: CanvasToolbarProps) => {
  const { t } = useI18n();
  const [templatesMenuOpen, setTemplatesMenuOpen] = useState(false);
  const [isSearchFocusWithin, setIsSearchFocusWithin] = useState(false);
  const [isSearchExpanded, setIsSearchExpanded] = useState(false);
  const templateMenuRef = useRef<HTMLDivElement | null>(null);
  const searchRootRef = useRef<HTMLDivElement | null>(null);
  const searchInputRef = useRef<HTMLInputElement | null>(null);
  const resultsViewportRef = useRef<HTMLDivElement | null>(null);
  const resultButtonRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const canUseTemplates = universeTemplates.length > 0 && Boolean(onAddUniverseFromTemplate);
  const hasActiveFilters = searchFilters.universe || searchFilters.tag || searchFilters.reality;
  const showSearchResults = isSearchFocusWithin && searchQuery.trim().length > 0;
  const showExpandedSearch = isSearchExpanded || isSearchFocusWithin;

  const resolveMatchedFieldLabel = (result: CanvasSearchResult): string => {
    if (result.matchedField === "tag" && result.matchedTag) {
      return t("toolbar.searchMatchTag", { tag: result.matchedTag });
    }

    const key = `toolbar.searchMatch.${result.matchedField}`;
    return t(key, undefined, result.matchedField);
  };

  const toggleSearchFilter = (key: keyof CanvasSearchFilters) => {
    onSearchFiltersChange?.({
      ...searchFilters,
      [key]: !searchFilters[key],
    });

    requestAnimationFrame(() => {
      searchInputRef.current?.focus();
    });
  };

  const handleSearchKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (!showExpandedSearch) {
      return;
    }

    if (event.key === "ArrowDown") {
      event.preventDefault();
      onSearchMoveSelection?.("down");
      return;
    }

    if (event.key === "ArrowUp") {
      event.preventDefault();
      onSearchMoveSelection?.("up");
      return;
    }

    if (event.key === "Enter" && searchActiveIndex >= 0) {
      event.preventDefault();
      onSearchSelect?.(searchActiveIndex);
      return;
    }

    if (event.key === "Escape") {
      event.preventDefault();
      onSearchQueryChange?.("");
      setIsSearchFocusWithin(false);
      setIsSearchExpanded(false);

      const target = event.target;
      if (target instanceof HTMLElement) {
        target.blur();
      }
    }
  };

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

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      const container = searchRootRef.current;
      if (!container) {
        return;
      }

      const path = event.composedPath();
      if (path.includes(container)) {
        return;
      }

      setIsSearchFocusWithin(false);
      setIsSearchExpanded(false);
    };

    document.addEventListener("pointerdown", handlePointerDown, true);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown, true);
    };
  }, []);

  useEffect(() => {
    if (!showSearchResults || searchActiveIndex < 0) {
      return;
    }

    const activeResult = resultButtonRefs.current[searchActiveIndex];
    activeResult?.scrollIntoView({
      block: "nearest",
      behavior: "smooth",
    });
  }, [searchActiveIndex, searchResults, showSearchResults]);

  useEffect(() => {
    if (!showExpandedSearch) {
      return;
    }

    const frameId = window.requestAnimationFrame(() => {
      searchInputRef.current?.focus({ preventScroll: true });
    });

    return () => {
      window.cancelAnimationFrame(frameId);
    };
  }, [showExpandedSearch]);

  const openSearchInput = () => {
    setIsSearchExpanded(true);
  };

  const closeSearchPanel = () => {
    setIsSearchFocusWithin(false);
    setIsSearchExpanded(false);

    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
  };

  return (
    <>
      <div
        ref={searchRootRef}
        className={`fixed top-20 right-8 z-[70] transition-[width] duration-300 ease-[cubic-bezier(0.22,1,0.36,1)] ${
          showExpandedSearch ? "w-[440px]" : "w-11"
        }`}
        onKeyDownCapture={handleSearchKeyDown}
        onFocusCapture={() => setIsSearchFocusWithin(true)}
        onBlurCapture={(event) => {
          const nextTarget = event.relatedTarget;
          if (
            searchRootRef.current &&
            nextTarget instanceof Node &&
            searchRootRef.current.contains(nextTarget)
          ) {
            return;
          }
          setIsSearchFocusWithin(false);
          setIsSearchExpanded(false);
        }}
      >
        <div
          className={`h-11 overflow-hidden border rounded-full backdrop-blur transition-[padding,background-color,box-shadow] duration-300 ease-[cubic-bezier(0.22,1,0.36,1)] ${
            showExpandedSearch
              ? "bg-slate-900/95 border-slate-700 shadow-2xl px-3 py-2"
              : "bg-slate-900/90 border-slate-700/80 shadow-lg p-0"
          }`}
        >
          {!showExpandedSearch ? (
            <button
              type="button"
              data-testid="toolbar-search-toggle"
              onMouseDown={(event) => {
                event.preventDefault();
              }}
              onClick={openSearchInput}
              className="w-11 h-11 flex items-center justify-center text-slate-300 hover:text-slate-100 transition-colors"
              title={t("toolbar.searchAria")}
              aria-label={t("toolbar.searchAria")}
            >
              <Search size={14} />
            </button>
          ) : (
            <div className="flex items-center gap-2">
              <Search size={14} className="text-slate-400 shrink-0" />
              <input
                ref={searchInputRef}
                autoFocus
                data-testid="toolbar-search-input"
                value={searchQuery}
                onChange={(event) => onSearchQueryChange?.(event.target.value)}
                className="min-w-0 flex-1 bg-transparent text-sm text-slate-200 placeholder:text-slate-500 focus:outline-none"
                placeholder={t("toolbar.searchPlaceholder")}
                aria-label={t("toolbar.searchAria")}
              />

              <div className="flex items-center gap-1 pl-2 border-l border-slate-700">
                <button
                  type="button"
                  data-testid="toolbar-search-filter-universe"
                  title={t("toolbar.searchFilter.universe")}
                  aria-label={t("toolbar.searchFilter.universe")}
                  onClick={() => toggleSearchFilter("universe")}
                  className={`p-1.5 rounded transition-colors border ${
                    searchFilters.universe
                      ? "text-blue-300 border-blue-500/60 bg-blue-500/20"
                      : "text-slate-500 border-slate-700 hover:border-slate-600"
                  }`}
                >
                  <STUDIO_ICONS.entity.universe size={13} />
                </button>
                <button
                  type="button"
                  data-testid="toolbar-search-filter-tag"
                  title={t("toolbar.searchFilter.tag")}
                  aria-label={t("toolbar.searchFilter.tag")}
                  onClick={() => toggleSearchFilter("tag")}
                  className={`p-1.5 rounded transition-colors border ${
                    searchFilters.tag
                      ? "text-cyan-300 border-cyan-500/60 bg-cyan-500/20"
                      : "text-slate-500 border-slate-700 hover:border-slate-600"
                  }`}
                >
                  <Tag size={13} />
                </button>
                <button
                  type="button"
                  data-testid="toolbar-search-filter-reality"
                  title={t("toolbar.searchFilter.reality")}
                  aria-label={t("toolbar.searchFilter.reality")}
                  onClick={() => toggleSearchFilter("reality")}
                  className={`p-1.5 rounded transition-colors border ${
                    searchFilters.reality
                      ? "text-green-300 border-green-500/60 bg-green-500/20"
                      : "text-slate-500 border-slate-700 hover:border-slate-600"
                  }`}
                >
                  <STUDIO_ICONS.entity.reality size={13} />
                </button>
                {searchQuery && (
                  <button
                    type="button"
                    data-testid="toolbar-search-clear"
                    onClick={() => onSearchQueryChange?.("")}
                    className="p-1 text-slate-400 hover:text-slate-200 rounded transition-colors"
                    title={t("toolbar.searchClear")}
                    aria-label={t("toolbar.searchClear")}
                  >
                    <X size={14} />
                  </button>
                )}
              </div>
            </div>
          )}
        </div>
        {showSearchResults && (
          <div
            ref={resultsViewportRef}
            className="mt-2 max-h-72 overflow-y-auto custom-scrollbar bg-slate-900/95 backdrop-blur border border-slate-700 rounded-xl shadow-2xl"
          >
            {!hasActiveFilters ? (
              <div className="px-3 py-3 text-xs text-slate-500">{t("toolbar.searchNoFilters")}</div>
            ) : searchResults.length === 0 ? (
              <div className="px-3 py-3 text-xs text-slate-500">{t("toolbar.searchNoResults")}</div>
            ) : (
              searchResults.map((result, index) => {
                const EntityIcon =
                  result.nodeType === "universe" ? STUDIO_ICONS.entity.universe : STUDIO_ICONS.entity.reality;
                const iconColor =
                  result.nodeType === "universe"
                    ? STUDIO_ICON_REGISTRY.entity.universe.colors.base
                    : STUDIO_ICON_REGISTRY.entity.reality.colors.base;
                const isActive = searchActiveIndex === index;

                return (
                  <button
                    key={result.nodeId}
                    type="button"
                    data-testid={`toolbar-search-result-${result.nodeId}`}
                    ref={(element) => {
                      resultButtonRefs.current[index] = element;
                    }}
                    onMouseDown={(event) => {
                      event.preventDefault();
                      onSearchSelect?.(index);
                    }}
                    onDoubleClick={(event) => {
                      event.preventDefault();
                      onSearchSelect?.(index);
                      closeSearchPanel();
                    }}
                    className={`w-full text-left px-3 py-2.5 border-b border-slate-800/70 last:border-b-0 transition-colors ${
                      isActive ? "bg-blue-600/20" : "hover:bg-slate-800/70"
                    }`}
                  >
                    <div className="flex items-start gap-2">
                      <EntityIcon
                        size={14}
                        data-testid={`toolbar-search-icon-${result.nodeType}`}
                        className={`${iconColor} mt-0.5 shrink-0`}
                      />
                      <div className="min-w-0 flex-1">
                        <div className="text-sm text-slate-100 font-mono truncate">{result.label}</div>
                        <div className="text-[11px] text-slate-400 mt-0.5 flex items-center gap-2">
                          <span>{resolveMatchedFieldLabel(result)}</span>
                          {result.nodeType === "reality" && result.contextLabel && (
                            <span>
                              {t("toolbar.searchContextReality", {
                                universe: result.contextLabel,
                              })}
                            </span>
                          )}
                          {result.nodeType === "universe" && result.contextLabel && (
                            <span className="truncate">
                              {t("toolbar.searchContextUniverse", {
                                canonicalName: result.contextLabel,
                              })}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  </button>
                );
              })
            )}
          </div>
        )}
      </div>

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
