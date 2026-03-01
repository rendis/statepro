import { ChevronDown, ChevronLeft, ChevronRight, Settings2, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { Dispatch, SetStateAction } from "react";

import { BehaviorArrayEditor, MetadataPackBindingsEditor } from "../../components/shared";
import { STUDIO_DS, STUDIO_ICON_REGISTRY, STUDIO_ICONS, studioTabClass } from "../../constants";
import { resolveSerializeIssueMessage, useI18n } from "../../i18n";
import { buildMachineEntityRef, cleanIdentifier, formatIdentifier, isValidIdentifier } from "../../utils";
import type {
  BehaviorModalState,
  EditorNode,
  MachineConfig,
  MetadataPackBindingMap,
  MetadataPackRegistry,
  SerializeIssue,
  UniversalConstants,
  BehaviorRegistryItem,
} from "../../types";

interface MachineGlobalPanelProps {
  config: MachineConfig;
  nodes: EditorNode[];
  registry: BehaviorRegistryItem[];
  metadataPackRegistry: MetadataPackRegistry;
  metadataPackBindings: MetadataPackBindingMap;
  setConfig: Dispatch<SetStateAction<MachineConfig>>;
  setMetadataPackBindings: Dispatch<SetStateAction<MetadataPackBindingMap>>;
  openBehaviorModal: Dispatch<SetStateAction<BehaviorModalState>>;
  machineIssues?: SerializeIssue[];
}

type MachinePanelTab = "general" | "constants" | "metadataPacks" | "advanced";

export const MachineGlobalPanel = ({
  config,
  setConfig,
  nodes,
  openBehaviorModal,
  registry,
  metadataPackRegistry,
  metadataPackBindings,
  setMetadataPackBindings,
  machineIssues = [],
}: MachineGlobalPanelProps) => {
  const { t } = useI18n();
  const [expanded, setExpanded] = useState(false);
  const [activeTab, setActiveTab] = useState<MachinePanelTab>("general");
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [canonicalNameDraft, setCanonicalNameDraft] = useState(config.canonicalName);
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const tabsScrollRef = useRef<HTMLDivElement | null>(null);
  const [canScrollTabsLeft, setCanScrollTabsLeft] = useState(false);
  const [canScrollTabsRight, setCanScrollTabsRight] = useState(false);

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      const container = dropdownRef.current;
      if (!container) {
        return;
      }

      const path = event.composedPath();
      if (!path.includes(container)) {
        setDropdownOpen(false);
      }
    };

    document.addEventListener("pointerdown", handlePointerDown, true);
    return () => document.removeEventListener("pointerdown", handlePointerDown, true);
  }, []);

  useEffect(() => {
    if (!dropdownOpen) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setDropdownOpen(false);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [dropdownOpen]);

  useEffect(() => {
    if (activeTab !== "general") {
      setDropdownOpen(false);
    }
  }, [activeTab]);

  useEffect(() => {
    setCanonicalNameDraft(config.canonicalName);
  }, [config.canonicalName]);

  useEffect(() => {
    if (!expanded) {
      setCanScrollTabsLeft(false);
      setCanScrollTabsRight(false);
      return;
    }

    const tabsContainer = tabsScrollRef.current;
    if (!tabsContainer) {
      return;
    }

    const updateScrollIndicators = () => {
      const maxScrollLeft = Math.max(0, tabsContainer.scrollWidth - tabsContainer.clientWidth);
      const hasOverflow = maxScrollLeft > 1;
      const scrollLeft = tabsContainer.scrollLeft;

      setCanScrollTabsLeft(hasOverflow && scrollLeft > 1);
      setCanScrollTabsRight(hasOverflow && scrollLeft < maxScrollLeft - 1);
    };

    const rafId = window.requestAnimationFrame(updateScrollIndicators);
    tabsContainer.addEventListener("scroll", updateScrollIndicators, { passive: true });
    window.addEventListener("resize", updateScrollIndicators);

    const resizeObserver =
      typeof ResizeObserver !== "undefined"
        ? new ResizeObserver(updateScrollIndicators)
        : null;
    resizeObserver?.observe(tabsContainer);

    return () => {
      window.cancelAnimationFrame(rafId);
      tabsContainer.removeEventListener("scroll", updateScrollIndicators);
      window.removeEventListener("resize", updateScrollIndicators);
      resizeObserver?.disconnect();
    };
  }, [expanded]);

  const universes = nodes.filter((node) => node.type === "universe");
  const getRealitiesForUniverse = (universeId: string) =>
    nodes.filter(
      (node): node is Extract<EditorNode, { type: "reality" }> =>
        node.type === "reality" && node.data.universeId === universeId,
    );

  const addInitial = (value: string) => {
    if (!config.initials.includes(value)) {
      setConfig((previous) => ({
        ...previous,
        initials: [...previous.initials, value],
      }));
    }
    setDropdownOpen(false);
  };

  const removeInitial = (value: string) => {
    setConfig((previous) => ({
      ...previous,
      initials: previous.initials.filter((initial) => initial !== value),
    }));
  };

  const updateUC = (field: keyof UniversalConstants, value: UniversalConstants[keyof UniversalConstants]) => {
    setConfig((previous) => ({
      ...previous,
      universalConstants: {
        ...previous.universalConstants,
        [field]: value,
      },
    }));
  };

  const uc = config.universalConstants;
  const UniverseIcon = STUDIO_ICONS.entity.universe;
  const RealityIcon = STUDIO_ICONS.entity.reality;
  const EntryPhaseIcon = STUDIO_ICONS.phase.entry;
  const ExitPhaseIcon = STUDIO_ICONS.phase.exit;
  const OnTransitionPhaseIcon = STUDIO_ICONS.phase.onTransition;
  const phaseLabelColorClass = {
    entry: STUDIO_ICON_REGISTRY.phase.entry.colors.base,
    exit: STUDIO_ICON_REGISTRY.phase.exit.colors.base,
    onTransition: STUDIO_ICON_REGISTRY.phase.onTransition.colors.base,
  } as const;
  const tabs: ReadonlyArray<{ id: MachinePanelTab; label: string }> = [
    { id: "general", label: t("machine.tabs.general") },
    { id: "constants", label: t("machine.tabs.constants") },
    { id: "metadataPacks", label: t("machine.tabs.metadataPacks") },
    { id: "advanced", label: t("machine.tabs.advanced") },
  ];
  const machineErrorCount = machineIssues.filter((issue) => issue.severity === "error").length;

  return (
    <div className="absolute top-20 left-6 z-[60] flex flex-col gap-2 w-96">
      <div
        onClick={() => setExpanded((previous) => !previous)}
        className="bg-slate-900/80 backdrop-blur-md border border-slate-700 hover:border-slate-500 rounded-xl p-3 shadow-xl cursor-pointer flex justify-between items-center transition-colors group"
      >
        <div>
          <div className="text-[10px] text-blue-400 font-bold uppercase tracking-wider flex items-center gap-1.5 mb-0.5">
            <Settings2 size={12} /> {t("machine.title")}
          </div>
          <div className="text-sm font-semibold text-slate-200 truncate w-72 flex items-center gap-2">
            <span>{config.id || t("machine.unnamed")}</span>
            {machineErrorCount > 0 && (
              <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-red-600 text-white font-mono">
                {machineErrorCount}
              </span>
            )}
          </div>
        </div>
        <div className="bg-slate-800 p-1 rounded-md group-hover:bg-slate-700 transition-colors">
          <ChevronDown
            className={`text-slate-400 transition-transform duration-300 ${expanded ? "rotate-180" : ""}`}
            size={16}
          />
        </div>
      </div>

      {expanded && (
        <div className="bg-slate-900/95 backdrop-blur-xl border border-slate-700 rounded-xl shadow-2xl p-4 space-y-4 animate-in slide-in-from-top-2 fade-in duration-200 max-h-[calc(100vh-140px)] overflow-y-auto custom-scrollbar">
          <div className="relative">
            {canScrollTabsLeft && (
              <div className="pointer-events-none absolute left-0 top-0 z-10 flex h-full w-10 items-center bg-gradient-to-r from-slate-900 to-transparent">
                <ChevronLeft size={14} className="text-slate-400" />
              </div>
            )}
            <div
              ref={tabsScrollRef}
              className={`${STUDIO_DS.tabsBar} gap-2 px-2 pt-1 overflow-x-auto overflow-y-hidden whitespace-nowrap [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden`}
            >
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`${studioTabClass(activeTab === tab.id)} shrink-0 whitespace-nowrap`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
            {canScrollTabsRight && (
              <div className="pointer-events-none absolute right-0 top-0 z-10 flex h-full w-10 items-center justify-end bg-gradient-to-l from-slate-900 to-transparent">
                <ChevronRight size={14} className="text-slate-400" />
              </div>
            )}
          </div>

          {activeTab === "general" && (
            <div className="space-y-4 animate-in fade-in">
              {machineErrorCount > 0 && (
                <div className="bg-red-950/30 border border-red-800/60 rounded-lg p-3 text-xs text-red-200">
                  <div className="font-semibold mb-1">{t("machine.errors.title")}</div>
                  <ul className="space-y-1 max-h-24 overflow-y-auto custom-scrollbar">
                    {machineIssues.slice(0, 8).map((issue, index) => (
                      <li key={`${issue.field}-${index}`}>
                        <span className="font-mono">{issue.field}</span>: {resolveSerializeIssueMessage(issue, t)}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              <div className="space-y-3">
                <div>
                  <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider flex justify-between">
                    <span>{t("machine.fields.id")}</span>
                    <span className="text-slate-500 normal-case">{t("common.readOnly")}</span>
                  </label>
                  <input
                    type="text"
                    value={config.id}
                    readOnly
                    className={`${STUDIO_DS.inputSm} text-slate-400 cursor-not-allowed`}
                  />
                </div>
                <div className="flex gap-2">
                  <div className="flex-1">
                    <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                      <span>{t("machine.fields.canonicalName")}</span>
                    </label>
                    <input
                      type="text"
                      value={canonicalNameDraft}
                      onChange={(event) => {
                        const formatted = formatIdentifier(event.target.value);
                        setCanonicalNameDraft(formatted);
                        if (isValidIdentifier(formatted)) {
                          setConfig((previous) => ({
                            ...previous,
                            canonicalName: formatted,
                          }));
                        }
                      }}
                      onBlur={() => {
                        const cleanName = cleanIdentifier(canonicalNameDraft) || config.canonicalName;
                        setCanonicalNameDraft(cleanName);
                        if (cleanName !== config.canonicalName) {
                          setConfig((previous) => ({
                            ...previous,
                            canonicalName: cleanName,
                          }));
                        }
                      }}
                      className={STUDIO_DS.inputSm}
                    />
                  </div>
                  <div className="w-24">
                    <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                      {t("machine.fields.version")}
                    </label>
                    <input
                      type="text"
                      value={config.version}
                      onChange={(event) =>
                        setConfig((previous) => ({ ...previous, version: event.target.value }))
                      }
                      className={STUDIO_DS.inputSm}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                    {t("machine.fields.description")}
                  </label>
                  <textarea
                    value={config.description || ""}
                    onChange={(event) =>
                      setConfig((previous) => ({ ...previous, description: event.target.value }))
                    }
                    className={`${STUDIO_DS.textarea} h-16 px-3 py-1.5`}
                  />
                </div>
              </div>

              <div className="pt-2 border-t border-slate-800 relative" ref={dropdownRef}>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("machine.fields.initials")}
                </label>
                <div
                  className="w-full bg-slate-950 border border-slate-700 rounded px-2 py-1.5 flex flex-wrap gap-1.5 min-h-[34px] cursor-text focus-within:border-blue-500 transition-colors"
                  onClick={() => setDropdownOpen(true)}
                >
                  {config.initials.length === 0 && !dropdownOpen && (
                    <span className="text-[10px] text-slate-600 self-center ml-1 font-mono">
                      {t("machine.initials.placeholder")}
                    </span>
                  )}
                  {config.initials.map((initial) => (
                    <span
                      key={initial}
                      className="flex items-center gap-1 bg-blue-900/30 border border-blue-500/30 text-blue-300 text-[10px] px-2 py-0.5 rounded font-mono shadow-sm"
                    >
                      {initial}
                      <button
                        onClick={(event) => {
                          event.stopPropagation();
                          removeInitial(initial);
                        }}
                        className="hover:text-blue-100 transition-colors"
                      >
                        <X size={10} />
                      </button>
                    </span>
                  ))}
                </div>

                {dropdownOpen && (
                  <div className="absolute bottom-full mb-0 left-0 w-full bg-slate-800 border border-slate-700 rounded-lg shadow-2xl z-[70] max-h-48 overflow-y-auto py-1">
                    {universes.length === 0 && (
                      <div className="p-3 text-xs text-slate-500 text-center">{t("machine.noUniverses")}</div>
                    )}
                    {universes.map((universe) => {
                      const universeCanonicalId = universe.data.id || universe.data.name;
                      const realities = getRealitiesForUniverse(universe.id);

                      return (
                        <div key={`opt-${universe.id}`} className="flex flex-col">
                          <button
                            onClick={() => addInitial(`U:${universeCanonicalId}`)}
                            disabled={config.initials.includes(`U:${universeCanonicalId}`)}
                            className="w-full text-left px-3 py-2 hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2 group transition-colors border-l-2 border-transparent hover:border-blue-500"
                          >
                            <UniverseIcon size={12} className={STUDIO_ICON_REGISTRY.entity.universe.colors.base} />
                            <span className="text-[11px] text-white font-bold font-mono">U:{universeCanonicalId}</span>
                          </button>

                          {realities.map((reality) => {
                            const realityCanonicalId = reality.data.id || reality.data.name;
                            const value = `U:${universeCanonicalId}:${realityCanonicalId}`;
                            return (
                              <button
                                key={`opt-${reality.id}`}
                                onClick={() => addInitial(value)}
                                disabled={config.initials.includes(value)}
                                className="w-full text-left pl-8 pr-3 py-1.5 hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2 group transition-colors border-l-2 border-transparent hover:border-green-500"
                              >
                                <RealityIcon
                                  size={10}
                                  className={`${STUDIO_ICON_REGISTRY.entity.reality.colors.base} z-10`}
                                />
                                <span className="text-[10px] text-slate-300 font-mono">{realityCanonicalId}</span>
                              </button>
                            );
                          })}
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === "constants" && (
            <div className="space-y-4 animate-in fade-in">
              <div className="flex items-center gap-2 mb-1">
                <label className="block text-[10px] text-slate-400 font-mono uppercase tracking-wider">
                  {t("machine.constants.title")}
                </label>
                <div className="flex-1 h-px bg-slate-800" />
              </div>

              <div className="space-y-4">
                <div className="bg-slate-950/50 p-2.5 rounded-lg border border-slate-800/50">
                  <div
                    className={`text-[10px] font-semibold uppercase tracking-wider mb-2 flex items-center gap-1 ${phaseLabelColorClass.entry}`}
                  >
                    <EntryPhaseIcon size={10} /> {t("machine.constants.entryPhase")}
                  </div>
                  <div className="space-y-2">
                    <BehaviorArrayEditor
                      items={uc.entryActions}
                      onChange={(items) => updateUC("entryActions", items)}
                      type="action"
                      openBehaviorModal={openBehaviorModal}
                      registry={registry}
                    />
                    <BehaviorArrayEditor
                      items={uc.entryInvokes}
                      onChange={(items) => updateUC("entryInvokes", items)}
                      type="invoke"
                      openBehaviorModal={openBehaviorModal}
                      registry={registry}
                    />
                  </div>
                </div>

                <div className="bg-slate-950/50 p-2.5 rounded-lg border border-slate-800/50">
                  <div
                    className={`text-[10px] font-semibold uppercase tracking-wider mb-2 flex items-center gap-1 ${phaseLabelColorClass.exit}`}
                  >
                    <ExitPhaseIcon size={10} /> {t("machine.constants.exitPhase")}
                  </div>
                  <div className="space-y-2">
                    <BehaviorArrayEditor
                      items={uc.exitActions}
                      onChange={(items) => updateUC("exitActions", items)}
                      type="action"
                      openBehaviorModal={openBehaviorModal}
                      registry={registry}
                    />
                    <BehaviorArrayEditor
                      items={uc.exitInvokes}
                      onChange={(items) => updateUC("exitInvokes", items)}
                      type="invoke"
                      openBehaviorModal={openBehaviorModal}
                      registry={registry}
                    />
                  </div>
                </div>

                <div className="bg-slate-950/50 p-2.5 rounded-lg border border-slate-800/50">
                  <div
                    className={`text-[10px] font-semibold uppercase tracking-wider mb-2 flex items-center gap-1 ${phaseLabelColorClass.onTransition}`}
                  >
                    <OnTransitionPhaseIcon size={10} /> {t("machine.constants.onTransition")}
                  </div>
                  <div className="space-y-2">
                    <BehaviorArrayEditor
                      items={uc.actionsOnTransition}
                      onChange={(items) => updateUC("actionsOnTransition", items)}
                      type="action"
                      openBehaviorModal={openBehaviorModal}
                      registry={registry}
                    />
                    <BehaviorArrayEditor
                      items={uc.invokesOnTransition}
                      onChange={(items) => updateUC("invokesOnTransition", items)}
                      type="invoke"
                      openBehaviorModal={openBehaviorModal}
                      registry={registry}
                    />
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === "advanced" && (
            <div className="space-y-3 animate-in fade-in">
              <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                {t("machine.advanced.metadata")}
              </label>
              <textarea
                value={config.metadata}
                onChange={(event) =>
                  setConfig((previous) => ({ ...previous, metadata: event.target.value }))
                }
                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-yellow-400 focus:outline-none focus:border-blue-500 font-mono text-[10px] h-40 resize-y leading-tight"
              />
            </div>
          )}

          {activeTab === "metadataPacks" && (
            <MetadataPackBindingsEditor
              scope="machine"
              entityRef={buildMachineEntityRef()}
              packRegistry={metadataPackRegistry}
              bindings={metadataPackBindings}
              metadataRaw={config.metadata}
              onChangeBindings={(next) => setMetadataPackBindings(next)}
            />
          )}
        </div>
      )}
    </div>
  );
};
