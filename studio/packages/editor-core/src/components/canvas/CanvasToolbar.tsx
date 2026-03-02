import {
  ChevronUp,
  Eye,
  EyeOff,
  Info,
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
import {
  useEffect,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from "react";
import visualizationDimUnrelated from "../../assets/visualization_dim_unrelated.svg";
import visualizationHideUnrelated from "../../assets/visualization_hide_unrelated.svg";

import { STUDIO_ICON_REGISTRY, STUDIO_ICONS } from "../../constants";
import { useI18n } from "../../i18n";
import type { StudioUniverseTemplate } from "../../types";
import type { CanvasSearchFilters, CanvasSearchResult } from "../../utils";
import { StudioTooltip } from "../shared";

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
  selectedNodeCount?: number;
  visualizationMode?: "off" | "hide" | "dim";
  onVisualizationModeChange?: (mode: "off" | "hide" | "dim") => void;
}

interface VisualModePreviewInfoProps {
  imageSrc: string;
  alt: string;
}

const VISUALIZATION_HIDE_PREVIEW_SRC = visualizationHideUnrelated;
const VISUALIZATION_DIM_PREVIEW_SRC = visualizationDimUnrelated;

const VisualModePreviewInfo = ({ imageSrc, alt }: VisualModePreviewInfoProps) => {
  return (
    <StudioTooltip
      label={<img src={imageSrc} alt={alt} className="block" style={{ width: 288, height: "auto" }} />}
      side="right"
      portal={true}
      bubbleClassName="p-0 overflow-hidden rounded-md border border-slate-700 bg-slate-950 shadow-2xl"
    >
      <button
        type="button"
        className="h-6 w-6 inline-flex items-center justify-center rounded-md text-slate-400 hover:text-slate-200 hover:bg-slate-800 transition-colors"
        onClick={(event) => {
          event.preventDefault();
          event.stopPropagation();
        }}
        aria-label={alt}
      >
        <Info size={12} />
      </button>
    </StudioTooltip>
  );
};

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
  selectedNodeCount = 0,
  visualizationMode = "off",
  onVisualizationModeChange,
}: CanvasToolbarProps) => {
  const { t } = useI18n();
  const [templatesMenuOpen, setTemplatesMenuOpen] = useState(false);
  const [visualMenuOpen, setVisualMenuOpen] = useState(false);
  const [isSearchFocusWithin, setIsSearchFocusWithin] = useState(false);
  const [isSearchExpanded, setIsSearchExpanded] = useState(false);
  const templateMenuRef = useRef<HTMLDivElement | null>(null);
  const visualMenuRef = useRef<HTMLDivElement | null>(null);
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
    const handlePointerDown = (event: PointerEvent) => {
      const container = visualMenuRef.current;
      if (!container) {
        return;
      }

      const path = event.composedPath();
      if (!path.includes(container)) {
        setVisualMenuOpen(false);
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

  const setVisualizationMode = (mode: "off" | "hide" | "dim") => {
    onVisualizationModeChange?.(mode);
    setVisualMenuOpen(false);
  };
  const centerToolbarIconButtonBase =
    "h-9 w-9 inline-flex items-center justify-center rounded-lg transition-colors";

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
          className={`h-11 overflow-visible border rounded-full backdrop-blur transition-[padding,background-color,box-shadow] duration-300 ease-[cubic-bezier(0.22,1,0.36,1)] ${
            showExpandedSearch
              ? "bg-slate-900/95 border-slate-700 shadow-2xl px-3 py-2"
              : "bg-slate-900/90 border-slate-700/80 shadow-lg p-0"
          }`}
        >
          {!showExpandedSearch ? (
            <StudioTooltip label={t("toolbar.searchAria")} side="bottom">
              <button
                type="button"
                data-testid="toolbar-search-toggle"
                onMouseDown={(event) => {
                  event.preventDefault();
                }}
                onClick={openSearchInput}
                className="w-11 h-11 flex items-center justify-center text-slate-300 hover:text-slate-100 transition-colors"
                aria-label={t("toolbar.searchAria")}
              >
                <Search size={14} />
              </button>
            </StudioTooltip>
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
                <StudioTooltip label={t("toolbar.searchFilter.universe")} side="bottom">
                  <button
                    type="button"
                    data-testid="toolbar-search-filter-universe"
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
                </StudioTooltip>
                <StudioTooltip label={t("toolbar.searchFilter.tag")} side="bottom">
                  <button
                    type="button"
                    data-testid="toolbar-search-filter-tag"
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
                </StudioTooltip>
                <StudioTooltip label={t("toolbar.searchFilter.reality")} side="bottom">
                  <button
                    type="button"
                    data-testid="toolbar-search-filter-reality"
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
                </StudioTooltip>
                {searchQuery && (
                  <StudioTooltip label={t("toolbar.searchClear")} side="bottom">
                    <button
                      type="button"
                      data-testid="toolbar-search-clear"
                      onClick={() => onSearchQueryChange?.("")}
                      className="p-1 text-slate-400 hover:text-slate-200 rounded transition-colors"
                      aria-label={t("toolbar.searchClear")}
                    >
                      <X size={14} />
                    </button>
                  </StudioTooltip>
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

      <div className="fixed bottom-6 left-1/2 -translate-x-1/2 bg-slate-900/95 backdrop-blur border border-slate-700 rounded-xl p-1.5 shadow-2xl flex items-center gap-1 z-[60] animate-in slide-in-from-bottom-5">
        <StudioTooltip label={t("toolbar.addBlankUniverse")}>
          <button
            onClick={onAddUniverse}
            className={`${centerToolbarIconButtonBase} text-slate-200 hover:text-white hover:bg-slate-700`}
            aria-label={t("toolbar.addBlankUniverse")}
          >
            <span className="relative inline-flex items-center justify-center">
              <STUDIO_ICONS.entity.universe
                size={16}
                className={STUDIO_ICON_REGISTRY.entity.universe.colors.base}
              />
              <Plus
                size={13}
                className="absolute -right-2 -bottom-2 text-emerald-300 drop-shadow-[0_0_4px_rgba(16,185,129,0.45)]"
              />
            </span>
          </button>
        </StudioTooltip>
        {onAddGlobalNote && (
          <>
            <div className="w-px h-5 bg-slate-700 mx-0.5" />
            <StudioTooltip label={t("toolbar.addGlobalNote")}>
              <button
                onClick={onAddGlobalNote}
                className={`${centerToolbarIconButtonBase} text-yellow-500 hover:text-yellow-400 hover:bg-yellow-900/30`}
                aria-label={t("toolbar.addGlobalNote")}
              >
                <StickyNote size={16} />
              </button>
            </StudioTooltip>
          </>
        )}
        <div className="w-px h-5 bg-slate-700 mx-0.5" />
        <div className="relative" ref={templateMenuRef}>
          <StudioTooltip label={t("toolbar.prebuiltUniverses")}>
            <button
              type="button"
              onClick={() => {
                if (!canUseTemplates) {
                  return;
                }
                setTemplatesMenuOpen((previous) => !previous);
              }}
              disabled={!canUseTemplates}
              className={`${centerToolbarIconButtonBase} ${
                canUseTemplates
                  ? "text-slate-300 hover:text-white hover:bg-slate-700"
                  : "text-slate-400 cursor-not-allowed"
              }`}
              aria-label={t("toolbar.prebuiltUniverses")}
            >
              <ChevronUp size={16} />
            </button>
          </StudioTooltip>

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

        <div className="w-px h-5 bg-slate-700 mx-0.5" />
        <div className="relative" ref={visualMenuRef}>
          <StudioTooltip label={t("toolbar.visual.mode")}>
            <button
              type="button"
              data-testid="toolbar-visual-mode-trigger"
              onClick={() => {
                setVisualMenuOpen((previous) => !previous);
              }}
              aria-label={t("toolbar.visual.mode")}
              aria-haspopup="menu"
              aria-expanded={visualMenuOpen}
              className={`${centerToolbarIconButtonBase} ${
                visualizationMode === "hide"
                  ? "text-rose-200 bg-rose-500/20 hover:bg-rose-500/30"
                  : visualizationMode === "dim"
                    ? "text-amber-100 bg-amber-500/20 hover:bg-amber-500/30"
                    : "text-slate-300 hover:text-white hover:bg-slate-700"
              }`}
            >
              {visualizationMode === "hide" ? (
                <EyeOff size={16} />
              ) : visualizationMode === "dim" ? (
                <Target size={16} />
              ) : (
                <Eye size={16} />
              )}
            </button>
          </StudioTooltip>

          {visualMenuOpen && (
            <div
              data-testid="toolbar-visual-mode-menu"
              role="menu"
              className="absolute bottom-14 left-1/2 -translate-x-1/2 w-52 bg-slate-900 border border-slate-700 rounded-lg shadow-2xl p-2 z-[80]"
            >
              <button
                type="button"
                role="menuitemradio"
                data-testid="toolbar-visual-mode-off"
                aria-checked={visualizationMode === "off"}
                onClick={() => setVisualizationMode("off")}
                className={`w-full text-left rounded-md px-2.5 py-2 transition-colors flex items-center gap-2 ${
                  visualizationMode === "off"
                    ? "bg-slate-700/80 text-slate-100"
                    : "text-slate-300 hover:bg-slate-800"
                }`}
              >
                <Eye size={14} />
                <span className="text-xs">{t("toolbar.visual.off")}</span>
              </button>
              <div className="w-full flex items-center gap-1">
                <button
                  type="button"
                  role="menuitemradio"
                  data-testid="toolbar-visual-mode-hide"
                  aria-checked={visualizationMode === "hide"}
                  onClick={() => setVisualizationMode("hide")}
                  className={`flex-1 text-left rounded-md px-2.5 py-2 transition-colors flex items-center gap-2 ${
                    visualizationMode === "hide"
                      ? "bg-rose-500/20 text-rose-100"
                      : "text-slate-300 hover:bg-slate-800"
                  }`}
                >
                  <EyeOff size={14} />
                  <span className="text-xs">{t("toolbar.visual.hide")}</span>
                </button>
                <VisualModePreviewInfo
                  imageSrc={VISUALIZATION_HIDE_PREVIEW_SRC}
                  alt={t("toolbar.visual.hidePreview")}
                />
              </div>
              <div className="w-full flex items-center gap-1">
                <button
                  type="button"
                  role="menuitemradio"
                  data-testid="toolbar-visual-mode-dim"
                  aria-checked={visualizationMode === "dim"}
                  onClick={() => setVisualizationMode("dim")}
                  className={`flex-1 text-left rounded-md px-2.5 py-2 transition-colors flex items-center gap-2 ${
                    visualizationMode === "dim"
                      ? "bg-amber-500/20 text-amber-100"
                      : "text-slate-300 hover:bg-slate-800"
                  }`}
                >
                  <Target size={14} />
                  <span className="text-xs">{t("toolbar.visual.dim")}</span>
                </button>
                <VisualModePreviewInfo
                  imageSrc={VISUALIZATION_DIM_PREVIEW_SRC}
                  alt={t("toolbar.visual.dimPreview")}
                />
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="fixed bottom-6 right-8 bg-slate-900/95 backdrop-blur border border-slate-700 rounded-xl p-1.5 shadow-2xl flex items-center gap-1 z-[60] animate-in slide-in-from-bottom-4">
        <StudioTooltip
          label={isAutoLayouting ? t("toolbar.autoLayoutRunning") : t("toolbar.autoLayout")}
        >
          <button
            onClick={onAutoLayout}
            disabled={!onAutoLayout || isAutoLayouting}
            className="px-3 py-2 text-xs font-semibold text-emerald-300 hover:text-emerald-200 hover:bg-emerald-500/20 border border-emerald-500/30 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            aria-label={t("toolbar.autoLayout")}
          >
            {isAutoLayouting ? t("toolbar.autoLayoutRunning") : t("toolbar.autoLayout")}
          </button>
        </StudioTooltip>

        <div className="w-px h-5 bg-slate-700 mx-0.5" />

        <StudioTooltip label={t("toolbar.undo")}>
          <button
            onClick={onUndo}
            disabled={!onUndo || !canUndo}
            className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            aria-label={t("toolbar.undo")}
          >
            <Undo2 size={16} />
          </button>
        </StudioTooltip>

        <StudioTooltip label={t("toolbar.redo")}>
          <button
            onClick={onRedo}
            disabled={!onRedo || !canRedo}
            className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            aria-label={t("toolbar.redo")}
          >
            <Redo2 size={16} />
          </button>
        </StudioTooltip>

        <div className="w-px h-5 bg-slate-700 mx-0.5" />

        <StudioTooltip label={t("toolbar.zoomOut")}>
          <button
            onClick={onZoomOut}
            className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
            aria-label={t("toolbar.zoomOut")}
          >
            <Minus size={16} />
          </button>
        </StudioTooltip>

        <StudioTooltip label={t("toolbar.zoomIn")}>
          <button
            onClick={onZoomIn}
            className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
            aria-label={t("toolbar.zoomIn")}
          >
            <Plus size={16} />
          </button>
        </StudioTooltip>

        <span className="px-2 py-1 text-[11px] font-mono font-semibold text-slate-300 min-w-[52px] text-center border border-slate-700 rounded-md bg-slate-950/70">
          {Math.round(zoom * 100)}%
        </span>

        <div className="w-px h-5 bg-slate-700 mx-0.5" />

        <StudioTooltip label={t("toolbar.fitContent")}>
          <button
            onClick={onFit}
            className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
            aria-label={t("toolbar.fitContent")}
          >
            <Maximize size={16} />
          </button>
        </StudioTooltip>

        <StudioTooltip label={t("toolbar.centerContent")}>
          <button
            onClick={onCenter}
            className="p-2 text-slate-300 hover:text-white hover:bg-slate-700 rounded-lg transition-colors"
            aria-label={t("toolbar.centerContent")}
          >
            <Target size={16} />
          </button>
        </StudioTooltip>
      </div>
    </>
  );
};
