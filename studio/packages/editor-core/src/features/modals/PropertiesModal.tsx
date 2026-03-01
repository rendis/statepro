import { AlertTriangle, ArrowDown, ArrowUp, Info, X } from "lucide-react";
import { useEffect, useState } from "react";
import type { Dispatch, SetStateAction } from "react";

import {
  BehaviorArrayEditor,
  EventSelector,
  MetadataPackBindingsEditor,
  TagEditor,
  TransitionTargetSelector,
} from "../../components/shared";
import { STUDIO_DS, STUDIO_ICON_REGISTRY, STUDIO_ICONS, studioTabClass } from "../../constants";
import { resolveSerializeIssueMessage, useI18n } from "../../i18n";
import {
  buildRealityEntityRef,
  buildTransitionEntityRef,
  buildUniverseEntityRef,
  formatIdentifier,
  hasInternalTargets,
} from "../../utils";
import type {
  BehaviorModalState,
  BehaviorRegistryItem,
  BehaviorRef,
  EditorNode,
  EditorTransition,
  MetadataPackBindingMap,
  MetadataPackRegistry,
  SerializeIssue,
  TransitionTriggerKind,
} from "../../types";

interface TransitionSelection {
  id: string;
  type: "transition";
  data: EditorTransition;
}

type InspectableNode = Extract<EditorNode, { type: "universe" | "reality" }>;

interface PropertiesModalProps {
  element: InspectableNode | TransitionSelection | null;
  nodes: EditorNode[];
  transitions: EditorTransition[];
  onClose: () => void;
  updateNodeData: (field: string, value: unknown) => void;
  commitUniverseIdRename: (universeNodeId: string, nextUniverseIdDraft: string) => string;
  commitUniverseCanonicalRename: (
    universeNodeId: string,
    nextCanonicalDraft: string,
    options: { syncId: boolean },
  ) => { id: string; canonicalName: string };
  commitRealityIdRename: (realityNodeId: string, nextRealityIdDraft: string) => string;
  updateTransitionData: (field: string, value: unknown) => void;
  moveTransition: (direction: "up" | "down") => void;
  openBehaviorModal: Dispatch<SetStateAction<BehaviorModalState>>;
  registry: BehaviorRegistryItem[];
  metadataPackRegistry: MetadataPackRegistry;
  metadataPackBindings: MetadataPackBindingMap;
  setMetadataPackBindings: Dispatch<SetStateAction<MetadataPackBindingMap>>;
  issues?: SerializeIssue[];
}

type UniverseConstantsField =
  | "entryActions"
  | "entryInvokes"
  | "exitActions"
  | "exitInvokes"
  | "actionsOnTransition"
  | "invokesOnTransition";

const PropertiesModal = ({
  element,
  nodes,
  transitions,
  onClose,
  updateNodeData,
  commitUniverseIdRename,
  commitUniverseCanonicalRename,
  commitRealityIdRename,
  updateTransitionData,
  moveTransition,
  openBehaviorModal,
  registry,
  metadataPackRegistry,
  metadataPackBindings,
  setMetadataPackBindings,
  issues = [],
}: PropertiesModalProps) => {
  const { t } = useI18n();
  const [activeTab, setActiveTab] = useState("general");
  const [universeIdDraft, setUniverseIdDraft] = useState("");
  const [universeCanonicalDraft, setUniverseCanonicalDraft] = useState("");
  const [universeIdCanonicalLinked, setUniverseIdCanonicalLinked] = useState(false);
  const [realityIdDraft, setRealityIdDraft] = useState("");

  useEffect(() => {
    setActiveTab("general");
  }, [element?.id]);

  if (!element) return null;

  const universeElement = element.type === "universe" ? element : null;
  const realityElement = element.type === "reality" ? element : null;

  useEffect(() => {
    if (universeElement) {
      const nextUniverseId = universeElement.data.id || "";
      const nextCanonical = universeElement.data.canonicalName || nextUniverseId;
      setUniverseIdDraft(nextUniverseId);
      setUniverseCanonicalDraft(nextCanonical);
      setUniverseIdCanonicalLinked(nextUniverseId === nextCanonical);
      return;
    }

    if (realityElement) {
      setRealityIdDraft(realityElement.data.id || realityElement.data.name || "");
    }
  }, [
    realityElement?.data.id,
    realityElement?.data.name,
    realityElement?.id,
    universeElement?.data.canonicalName,
    universeElement?.data.id,
    universeElement?.id,
  ]);

  const commitUniverseIdDraft = () => {
    if (!universeElement || universeIdCanonicalLinked) {
      return;
    }

    const resolvedId = commitUniverseIdRename(universeElement.id, universeIdDraft);
    setUniverseIdDraft(resolvedId);
  };

  const commitUniverseCanonicalDraft = () => {
    if (!universeElement) {
      return;
    }

    const committed = commitUniverseCanonicalRename(
      universeElement.id,
      universeCanonicalDraft,
      { syncId: universeIdCanonicalLinked },
    );
    setUniverseCanonicalDraft(committed.canonicalName);
    if (universeIdCanonicalLinked) {
      setUniverseIdDraft(committed.id);
    }
  };

  const commitRealityIdDraft = () => {
    if (!realityElement) {
      return;
    }

    const resolvedId = commitRealityIdRename(realityElement.id, realityIdDraft);
    setRealityIdDraft(resolvedId);
  };

  const handleUniversePairToggle = (checked: boolean) => {
    setUniverseIdCanonicalLinked(checked);
    if (!checked || !universeElement) {
      return;
    }

    const committed = commitUniverseCanonicalRename(
      universeElement.id,
      universeCanonicalDraft,
      { syncId: true },
    );
    setUniverseCanonicalDraft(committed.canonicalName);
    setUniverseIdDraft(committed.id);
  };

  const transitionElement = element.type === "transition" ? element : null;

  const sourceReality = transitionElement
    ? nodes.find(
        (node): node is Extract<EditorNode, { type: "reality" }> =>
          node.type === "reality" && node.id === transitionElement.data.sourceRealityId,
      )
    : null;
  const sourceUniverse = sourceReality
    ? nodes.find(
        (node): node is Extract<EditorNode, { type: "universe" }> =>
          node.type === "universe" && node.id === sourceReality.data.universeId,
      )
    : null;
  const selectedUniverseNode =
    element.type === "universe"
      ? element
      : element.type === "reality"
        ? nodes.find(
            (node): node is Extract<EditorNode, { type: "universe" }> =>
              node.type === "universe" && node.id === element.data.universeId,
          ) || null
        : sourceUniverse;

  const allAvailableTags = Array.from(
    new Set(
      nodes
        .filter((node) => node.type === "universe")
        .flatMap((node) => node.data.tags || []),
    ),
  );

  const allAvailableEvents = Array.from(
    new Set(
      transitions
        .filter((transition) => transition.triggerKind === "on" && transition.eventName)
        .map((transition) => transition.eventName || ""),
    ),
  ).filter(Boolean);

  const transitionSiblings = transitionElement
    ? transitions
        .filter(
          (transition) =>
            transition.sourceRealityId === transitionElement.data.sourceRealityId &&
            transition.triggerKind === transitionElement.data.triggerKind &&
            (transition.triggerKind === "always" ||
              (transition.eventName || "") === (transitionElement.data.eventName || "")),
        )
        .sort((a, b) => a.order - b.order)
    : [];

  const transitionPosition = transitionElement
    ? transitionSiblings.findIndex((transition) => transition.id === transitionElement.id)
    : -1;

  const canMoveTransitionUp = transitionPosition > 0;
  const canMoveTransitionDown =
    transitionPosition >= 0 && transitionPosition < transitionSiblings.length - 1;
  const errorIssues = issues.filter((issue) => issue.severity === "error");
  const warningIssues = issues.filter((issue) => issue.severity === "warning");
  const universeConstantsRuntimeWarning = issues.find(
    (issue) => issue.messageKey === "issue.universeConstantsRuntimeIgnored",
  );

  const getTabs = () => {
    if (element.type === "universe") {
      return [
        { id: "general", label: t("properties.tabs.general") },
        { id: "constants", label: t("properties.tabs.constants") },
        { id: "metadataPacks", label: t("properties.tabs.metadataPacks") },
        { id: "advanced", label: t("properties.tabs.advanced") },
      ];
    }

    if (element.type === "reality") {
      return [
        { id: "general", label: t("properties.tabs.general") },
        { id: "lifecycle", label: t("properties.tabs.lifecycle") },
        { id: "metadataPacks", label: t("properties.tabs.metadataPacks") },
        { id: "advanced", label: t("properties.tabs.advanced") },
      ];
    }

    return [
      { id: "general", label: t("properties.tabs.general") },
      { id: "guards", label: t("properties.tabs.guards") },
      { id: "effects", label: t("properties.tabs.effects") },
      { id: "metadataPacks", label: t("properties.tabs.metadataPacks") },
      { id: "advanced", label: t("properties.tabs.advanced") },
    ];
  };

  const tabs = getTabs();
  const metadataPackEntityRef =
    element.type === "universe"
      ? buildUniverseEntityRef(element.data.id)
      : element.type === "reality" && selectedUniverseNode
        ? buildRealityEntityRef(selectedUniverseNode.data.id, element.data.id)
        : transitionElement && sourceReality && sourceUniverse
          ? buildTransitionEntityRef(
              sourceUniverse.data.id,
              sourceReality.data.id,
              transitionElement.data.triggerKind,
              transitionElement.data.eventName,
              transitionElement.data.order,
            )
          : null;
  const metadataPackScope =
    element.type === "universe"
      ? "universe"
      : element.type === "reality"
        ? "reality"
        : "transition";

  const updateUniverseUC = (field: UniverseConstantsField, items: BehaviorRef[]) => {
    if (element.type !== "universe") return;
    const uc = element.data.universalConstants || {
      entryActions: [],
      exitActions: [],
      entryInvokes: [],
      exitInvokes: [],
      actionsOnTransition: [],
      invokesOnTransition: [],
    };

    updateNodeData("universalConstants", {
      ...uc,
      [field]: items,
    });
  };

  const updateTransitionTargets = (targets: string[]) => {
    updateTransitionData("targets", targets);
  };

  const dedupeConditionsBySource = (items: BehaviorRef[]): BehaviorRef[] => {
    const seen = new Set<string>();
    const unique: BehaviorRef[] = [];

    items.forEach((item) => {
      const src = item?.src || "";
      if (!src || seen.has(src)) {
        return;
      }
      seen.add(src);
      unique.push(item);
    });

    return unique;
  };

  const transitionHasLocalTargets = transitionElement
    ? hasInternalTargets(transitionElement.data, nodes)
    : false;

  const effectiveTransitionType = transitionElement
    ? transitionHasLocalTargets && transitionElement.data.type === "notify"
      ? "default"
      : transitionElement.data.type || "default"
    : "default";

  const TransitionDefaultIcon = STUDIO_ICONS.transition.type.default;
  const TransitionNotifyIcon = STUDIO_ICONS.transition.type.notify;
  const UniverseIcon = STUDIO_ICONS.entity.universe;
  const RealityIcon = STUDIO_ICONS.entity.reality;
  const EntryPhaseIcon = STUDIO_ICONS.phase.entry;
  const ExitPhaseIcon = STUDIO_ICONS.phase.exit;
  const OnTransitionPhaseIcon = STUDIO_ICONS.phase.onTransition;
  const transitionNotifyColorClass = STUDIO_ICON_REGISTRY.transition.type.notify.colors.base;
  const transitionDefaultColorClass = STUDIO_ICON_REGISTRY.transition.type.default.colors.base;
  const universeColorClass = STUDIO_ICON_REGISTRY.entity.universe.colors.base;
  const realityColorClass = STUDIO_ICON_REGISTRY.entity.reality.colors.base;
  const observerColorClass = STUDIO_ICON_REGISTRY.behavior.observer.colors.base;
  const actionColorClass = STUDIO_ICON_REGISTRY.behavior.action.colors.base;
  const invokeColorClass = STUDIO_ICON_REGISTRY.behavior.invoke.colors.base;
  const conditionColorClass = STUDIO_ICON_REGISTRY.behavior.condition.colors.base;
  const entryPhaseColorClass = STUDIO_ICON_REGISTRY.phase.entry.colors.base;
  const exitPhaseColorClass = STUDIO_ICON_REGISTRY.phase.exit.colors.base;
  const onTransitionPhaseColorClass = STUDIO_ICON_REGISTRY.phase.onTransition.colors.base;

  const transitionIcon =
    element.type === "transition" ? (
      transitionElement?.data.type === "notify" ? (
        <TransitionNotifyIcon className={transitionNotifyColorClass} />
      ) : (
        <TransitionDefaultIcon className={transitionDefaultColorClass} />
      )
    ) : element.type === "universe" ? (
      <UniverseIcon className={universeColorClass} />
    ) : (
      <RealityIcon className={realityColorClass} />
    );

  return (
    <div className={STUDIO_DS.modalOverlay}>
      <div className={`${STUDIO_DS.modalPanel} max-w-md`}>
        <div className={STUDIO_DS.modalHeader}>
          <h2 className={STUDIO_DS.modalTitle}>
            {transitionIcon}
            {t("properties.title", {
              entity:
                element.type === "transition"
                  ? t("properties.entity.transition")
                  : element.type === "universe"
                    ? t("properties.entity.universe")
                    : t("properties.entity.reality"),
            })}
          </h2>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
            <X size={20} />
          </button>
        </div>

        <div className={STUDIO_DS.tabsBar}>
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={studioTabClass(activeTab === tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div className={`${STUDIO_DS.panelBody} h-[450px]`}>
          {issues.length > 0 && (
            <div
              className={`mb-4 rounded-lg p-3 text-xs border ${
                errorIssues.length > 0
                  ? "bg-red-950/40 border-red-800/60 text-red-200"
                  : "bg-amber-950/40 border-amber-700/60 text-amber-200"
              }`}
            >
              <div className="font-semibold flex items-center gap-2 mb-2">
                <AlertTriangle size={14} />{" "}
                {errorIssues.length > 0
                  ? t("properties.validationErrors", { count: errorIssues.length })
                  : t("properties.validationWarnings", { count: warningIssues.length })}
              </div>
              <ul className="space-y-1 max-h-20 overflow-y-auto custom-scrollbar">
                {issues.slice(0, 8).map((issue, index) => (
                  <li key={`${issue.field}-${index}`}>
                    <span className="font-mono">{issue.field}</span>: {resolveSerializeIssueMessage(issue, t)}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {element.type === "reality" && activeTab === "general" && (
            <div className="space-y-4 animate-in fade-in">
              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("properties.reality.id")}
                </label>
                <input
                  type="text"
                  value={realityIdDraft}
                  onChange={(event) => setRealityIdDraft(formatIdentifier(event.target.value))}
                  onBlur={commitRealityIdDraft}
                  onKeyDown={(event) => {
                    if (event.key !== "Enter") {
                      return;
                    }
                    event.preventDefault();
                    event.currentTarget.blur();
                  }}
                  className={`${STUDIO_DS.input} font-mono`}
                />
              </div>
              <label className="flex items-center gap-2 cursor-pointer bg-slate-950 p-2 border border-slate-700 rounded hover:border-slate-600 transition-colors">
                <input
                  type="checkbox"
                  checked={element.data.isInitial || false}
                  onChange={(event) => updateNodeData("isInitial", event.target.checked)}
                  className="accent-blue-500 w-4 h-4"
                />
                <span className="text-slate-300 text-sm">{t("properties.reality.initial")}</span>
              </label>
              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("properties.description")}
                </label>
                <textarea
                  value={element.data.description || ""}
                  onChange={(event) => updateNodeData("description", event.target.value)}
                  className={`${STUDIO_DS.textarea} h-16`}
                />
              </div>
            </div>
          )}

          {element.type === "reality" && activeTab === "lifecycle" && (
            <div className="space-y-4 animate-in fade-in">
              <div className="bg-slate-950/50 p-3 rounded-lg border border-slate-800/50">
                <label
                  className={`block text-[10px] ${observerColorClass} mb-2 font-mono uppercase tracking-wider`}
                >
                  {t("properties.lifecycle.observers")}
                </label>
                <BehaviorArrayEditor
                  items={element.data.observers || []}
                  onChange={(items) => updateNodeData("observers", items)}
                  type="observer"
                  openBehaviorModal={openBehaviorModal}
                  registry={registry}
                />
              </div>

              <div className="bg-slate-950/50 p-3 rounded-lg border border-slate-800/50">
                <label
                  className={`block text-[10px] ${entryPhaseColorClass} mb-2 font-mono uppercase tracking-wider flex items-center gap-1`}
                >
                  <EntryPhaseIcon size={10} /> {t("machine.constants.entryPhase")}
                </label>
                <div className="space-y-2">
                  <BehaviorArrayEditor
                    items={element.data.entryActions || []}
                    onChange={(items) => updateNodeData("entryActions", items)}
                    type="action"
                    openBehaviorModal={openBehaviorModal}
                    registry={registry}
                  />
                  <BehaviorArrayEditor
                    items={element.data.entryInvokes || []}
                    onChange={(items) => updateNodeData("entryInvokes", items)}
                    type="invoke"
                    openBehaviorModal={openBehaviorModal}
                    registry={registry}
                  />
                </div>
              </div>

              <div className="bg-slate-950/50 p-3 rounded-lg border border-slate-800/50">
                <label
                  className={`block text-[10px] ${exitPhaseColorClass} mb-2 font-mono uppercase tracking-wider flex items-center gap-1`}
                >
                  <ExitPhaseIcon size={10} /> {t("machine.constants.exitPhase")}
                </label>
                <div className="space-y-2">
                  <BehaviorArrayEditor
                    items={element.data.exitActions || []}
                    onChange={(items) => updateNodeData("exitActions", items)}
                    type="action"
                    openBehaviorModal={openBehaviorModal}
                    registry={registry}
                  />
                  <BehaviorArrayEditor
                    items={element.data.exitInvokes || []}
                    onChange={(items) => updateNodeData("exitInvokes", items)}
                    type="invoke"
                    openBehaviorModal={openBehaviorModal}
                    registry={registry}
                  />
                </div>
              </div>
            </div>
          )}

          {element.type === "universe" && activeTab === "general" && (
            <div className="space-y-4 animate-in fade-in">
              <div className="space-y-3">
                <div>
                  <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider flex justify-between">
                    <span>{t("properties.universe.id")}</span>
                    {universeIdCanonicalLinked && (
                      <span className="text-slate-500 normal-case">{t("common.readOnly")}</span>
                    )}
                  </label>
                  <input
                    type="text"
                    value={universeIdDraft}
                    onChange={(event) => setUniverseIdDraft(formatIdentifier(event.target.value))}
                    onBlur={commitUniverseIdDraft}
                    onKeyDown={(event) => {
                      if (event.key !== "Enter") {
                        return;
                      }
                      event.preventDefault();
                      event.currentTarget.blur();
                    }}
                    readOnly={universeIdCanonicalLinked}
                    className={`${STUDIO_DS.inputSm} ${
                      universeIdCanonicalLinked
                        ? "text-slate-400 cursor-not-allowed"
                        : ""
                    }`}
                  />
                </div>
                <label className="flex items-center gap-2 cursor-pointer bg-slate-950 p-2 border border-slate-700 rounded hover:border-slate-600 transition-colors">
                  <input
                    type="checkbox"
                    checked={universeIdCanonicalLinked}
                    onChange={(event) => handleUniversePairToggle(event.target.checked)}
                    className="accent-blue-500 w-4 h-4"
                  />
                  <span className="text-slate-300 text-sm">{t("properties.universe.linkIdCanonical")}</span>
                </label>
                <p className="text-[10px] text-slate-500 -mt-2">{t("properties.universe.linkHelp")}</p>
                <div className="flex gap-2">
                  <div className="flex-1">
                    <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                      {t("machine.fields.canonicalName")}
                    </label>
                    <input
                      type="text"
                      value={universeCanonicalDraft}
                      onChange={(event) => setUniverseCanonicalDraft(formatIdentifier(event.target.value))}
                      onBlur={commitUniverseCanonicalDraft}
                      onKeyDown={(event) => {
                        if (event.key !== "Enter") {
                          return;
                        }
                        event.preventDefault();
                        event.currentTarget.blur();
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
                      value={element.data.version || ""}
                      onChange={(event) => updateNodeData("version", event.target.value)}
                      className={STUDIO_DS.inputSm}
                    />
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("properties.description")}
                </label>
                <textarea
                  value={element.data.description || ""}
                  onChange={(event) => updateNodeData("description", event.target.value)}
                  className={`${STUDIO_DS.textarea} h-16 px-3 py-1.5`}
                />
              </div>

              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("properties.universe.tags")}
                </label>
                <TagEditor
                  tags={element.data.tags || []}
                  onChange={(newTags) => updateNodeData("tags", newTags)}
                  availableTags={allAvailableTags}
                />
              </div>
            </div>
          )}

          {element.type === "universe" &&
            activeTab === "constants" &&
            (() => {
              const uc = element.data.universalConstants || {
                entryActions: [],
                exitActions: [],
                entryInvokes: [],
                exitInvokes: [],
                actionsOnTransition: [],
                invokesOnTransition: [],
              };

              return (
                <div className="space-y-4 animate-in fade-in">
                  {universeConstantsRuntimeWarning && (
                    <div className="text-xs bg-amber-950/30 border border-amber-700/60 text-amber-200 rounded-lg p-3">
                      {resolveSerializeIssueMessage(universeConstantsRuntimeWarning, t)}
                    </div>
                  )}
                  <div className="bg-slate-950/50 p-3 rounded-lg border border-slate-800/50">
                    <label
                      className={`block text-[10px] ${entryPhaseColorClass} mb-2 font-mono uppercase tracking-wider flex items-center gap-1`}
                    >
                      <EntryPhaseIcon size={10} /> {t("properties.constants.entryGlobal")}
                    </label>
                    <div className="space-y-2">
                      <BehaviorArrayEditor
                        items={uc.entryActions}
                        onChange={(items) => updateUniverseUC("entryActions", items)}
                        type="action"
                        openBehaviorModal={openBehaviorModal}
                        registry={registry}
                      />
                      <BehaviorArrayEditor
                        items={uc.entryInvokes}
                        onChange={(items) => updateUniverseUC("entryInvokes", items)}
                        type="invoke"
                        openBehaviorModal={openBehaviorModal}
                        registry={registry}
                      />
                    </div>
                  </div>

                  <div className="bg-slate-950/50 p-3 rounded-lg border border-slate-800/50">
                    <label
                      className={`block text-[10px] ${exitPhaseColorClass} mb-2 font-mono uppercase tracking-wider flex items-center gap-1`}
                    >
                      <ExitPhaseIcon size={10} /> {t("properties.constants.exitGlobal")}
                    </label>
                    <div className="space-y-2">
                      <BehaviorArrayEditor
                        items={uc.exitActions}
                        onChange={(items) => updateUniverseUC("exitActions", items)}
                        type="action"
                        openBehaviorModal={openBehaviorModal}
                        registry={registry}
                      />
                      <BehaviorArrayEditor
                        items={uc.exitInvokes}
                        onChange={(items) => updateUniverseUC("exitInvokes", items)}
                        type="invoke"
                        openBehaviorModal={openBehaviorModal}
                        registry={registry}
                      />
                    </div>
                  </div>

                  <div className="bg-slate-950/50 p-3 rounded-lg border border-slate-800/50">
                    <label
                      className={`block text-[10px] ${onTransitionPhaseColorClass} mb-2 font-mono uppercase tracking-wider flex items-center gap-1`}
                    >
                      <OnTransitionPhaseIcon size={10} /> {t("properties.constants.onTransitionGlobal")}
                    </label>
                    <div className="space-y-2">
                      <BehaviorArrayEditor
                        items={uc.actionsOnTransition}
                        onChange={(items) => updateUniverseUC("actionsOnTransition", items)}
                        type="action"
                        openBehaviorModal={openBehaviorModal}
                        registry={registry}
                      />
                      <BehaviorArrayEditor
                        items={uc.invokesOnTransition}
                        onChange={(items) => updateUniverseUC("invokesOnTransition", items)}
                        type="invoke"
                        openBehaviorModal={openBehaviorModal}
                        registry={registry}
                      />
                    </div>
                  </div>
                </div>
              );
            })()}

          {transitionElement && activeTab === "general" && (
            <div className="space-y-4 animate-in fade-in">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                    {t("properties.transition.trigger")}
                  </label>
                  <select
                    value={transitionElement.data.triggerKind}
                    onChange={(event) => {
                      const nextKind = event.target.value as TransitionTriggerKind;
                      updateTransitionData("triggerKind", nextKind);
                      if (nextKind === "always") {
                        updateTransitionData("eventName", undefined);
                      } else if (!transitionElement.data.eventName) {
                        updateTransitionData("eventName", "NEW_EVENT");
                      }
                    }}
                    className={STUDIO_DS.input}
                  >
                    <option value="on">{t("properties.transition.triggerOn")}</option>
                    <option value="always">{t("properties.transition.triggerAlways")}</option>
                  </select>
                </div>

                <div>
                  <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                    {t("properties.transition.mode")}
                  </label>
                  <select
                    value={effectiveTransitionType}
                    onChange={(event) => updateTransitionData("type", event.target.value as "default" | "notify")}
                    className={`${STUDIO_DS.input} disabled:opacity-50 disabled:cursor-not-allowed`}
                    disabled={transitionHasLocalTargets}
                    title={
                      transitionHasLocalTargets
                        ? t("properties.transition.notifyDisabled")
                        : ""
                    }
                  >
                    <option value="default">{t("properties.transition.default")}</option>
                    {!transitionHasLocalTargets && (
                      <option value="notify">{t("properties.transition.notify")}</option>
                    )}
                  </select>
                </div>
              </div>

              <div className="p-3 bg-slate-950/50 border border-slate-800 rounded-lg">
                <label className="block text-[10px] text-slate-400 mb-2 font-mono uppercase tracking-wider">
                  {t("properties.transition.order")}
                </label>
                <div className="flex items-center justify-between gap-3">
                  <span className="text-xs text-slate-300">
                    {t("properties.transition.priority")}{" "}
                    <span className="font-mono">
                      {transitionPosition >= 0 ? transitionPosition + 1 : "-"} /{" "}
                      {transitionSiblings.length || 1}
                    </span>
                  </span>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => moveTransition("up")}
                      disabled={!canMoveTransitionUp}
                      className="px-2 py-1 rounded border border-slate-700 text-slate-200 disabled:opacity-40 disabled:cursor-not-allowed hover:border-slate-500"
                      title={t("properties.transition.up")}
                    >
                      <ArrowUp size={12} />
                    </button>
                    <button
                      onClick={() => moveTransition("down")}
                      disabled={!canMoveTransitionDown}
                      className="px-2 py-1 rounded border border-slate-700 text-slate-200 disabled:opacity-40 disabled:cursor-not-allowed hover:border-slate-500"
                      title={t("properties.transition.down")}
                    >
                      <ArrowDown size={12} />
                    </button>
                  </div>
                </div>
              </div>

              {transitionHasLocalTargets && (
                <p className="text-[10px] text-slate-500 flex items-start gap-1 leading-tight -mt-2">
                  <Info size={12} className="shrink-0 text-blue-400" />
                  <span>
                    {t("properties.transition.notifySafeDisabled")}
                  </span>
                </p>
              )}

              {transitionElement.data.triggerKind === "on" && (
                <div className="p-4 bg-slate-950/40 border border-slate-800 rounded-xl mt-2">
                  <label className="block text-[10px] text-slate-400 mb-3 font-mono uppercase tracking-wider">
                    {t("properties.transition.eventToListen")}
                  </label>
                  <EventSelector
                    value={transitionElement.data.eventName}
                    onChange={(value) => updateTransitionData("eventName", value)}
                    availableEvents={allAvailableEvents}
                  />
                </div>
              )}

              <div className="p-4 bg-slate-950/40 border border-slate-800 rounded-xl">
                <label className="block text-[10px] text-slate-400 mb-2 font-mono uppercase tracking-wider">
                  {t("properties.transition.targets")}
                </label>
                <TransitionTargetSelector
                  sourceRealityNodeId={transitionElement.data.sourceRealityId}
                  targets={transitionElement.data.targets}
                  nodes={nodes}
                  onChange={updateTransitionTargets}
                />
                <p className="text-[10px] text-slate-500 mt-2">
                  {t("transitionTargetSelector.graphOnlyHelp")}
                </p>
              </div>

              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider mt-2">
                  {t("properties.transition.description")}
                </label>
                <textarea
                  value={transitionElement.data.description || ""}
                  onChange={(event) => updateTransitionData("description", event.target.value)}
                  placeholder={t("properties.transition.descriptionPlaceholder")}
                  className={`${STUDIO_DS.textarea} h-16`}
                />
              </div>
            </div>
          )}

          {transitionElement && activeTab === "guards" && (
            <div className="space-y-4 animate-in fade-in">
              <div className="bg-slate-950/50 p-4 rounded-lg border border-slate-800/50">
                <p className="text-xs text-slate-400 mb-4">
                  {t("properties.guards.help")}
                </p>
                <label
                  className={`block text-[10px] ${conditionColorClass} mb-2 font-mono uppercase tracking-wider`}
                >
                  {t("properties.guards.required")}
                </label>
                <BehaviorArrayEditor
                  items={transitionElement.data.conditions || []}
                  onChange={(items) => {
                    updateTransitionData("conditions", dedupeConditionsBySource(items));
                  }}
                  type="condition"
                  openBehaviorModal={openBehaviorModal}
                  registry={registry}
                />
              </div>
            </div>
          )}

          {transitionElement && activeTab === "effects" && (
            <div className="space-y-4 animate-in fade-in">
              <div className="bg-slate-950/50 p-3 rounded-lg border border-slate-800/50">
                <label
                  className={`block text-[10px] ${actionColorClass} mb-2 font-mono uppercase tracking-wider flex items-center gap-1`}
                >
                  {t("properties.effects.sync")}
                </label>
                <BehaviorArrayEditor
                  items={transitionElement.data.actions || []}
                  onChange={(items) => updateTransitionData("actions", items)}
                  type="action"
                  openBehaviorModal={openBehaviorModal}
                  registry={registry}
                />
              </div>

              <div className="bg-slate-950/50 p-3 rounded-lg border border-slate-800/50">
                <label
                  className={`block text-[10px] ${invokeColorClass} mb-2 font-mono uppercase tracking-wider flex items-center gap-1`}
                >
                  {t("properties.effects.async")}
                </label>
                <BehaviorArrayEditor
                  items={transitionElement.data.invokes || []}
                  onChange={(items) => updateTransitionData("invokes", items)}
                  type="invoke"
                  openBehaviorModal={openBehaviorModal}
                  registry={registry}
                />
              </div>
            </div>
          )}

          {activeTab === "metadataPacks" && (
            <>
              {metadataPackEntityRef ? (
                <MetadataPackBindingsEditor
                  scope={metadataPackScope}
                  entityRef={metadataPackEntityRef}
                  packRegistry={metadataPackRegistry}
                  bindings={metadataPackBindings}
                  metadataRaw={element.type === "transition" ? element.data.metadata : element.data.metadata}
                  onChangeBindings={(next) => setMetadataPackBindings(next)}
                />
              ) : (
                <div className="text-xs text-slate-500 bg-slate-950/40 border border-slate-800 rounded p-3">
                  {t("properties.metadata.identityMissing")}
                </div>
              )}
            </>
          )}

          {activeTab === "advanced" && (
            <div className="space-y-4 animate-in fade-in">
              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("properties.advanced.metadata")}
                </label>
                <textarea
                  value={element.type === "transition" ? element.data.metadata || "" : element.data.metadata || ""}
                  onChange={(event) =>
                    element.type === "transition"
                      ? updateTransitionData("metadata", event.target.value)
                      : updateNodeData("metadata", event.target.value)
                  }
                  placeholder={'{\n  "ui_color": "red"\n}'}
                  className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-yellow-400 focus:outline-none focus:border-blue-500 font-mono text-[11px] h-32 resize-y leading-tight custom-scrollbar"
                />
                <p className="text-[9px] text-slate-500 mt-2">
                  {t("properties.advanced.metadataHelp")}
                </p>
              </div>
            </div>
          )}
        </div>

        <div className="p-4 border-t border-slate-800 bg-slate-800/50 flex justify-end">
          <button onClick={onClose} className={STUDIO_DS.primaryButton}>
            {t("properties.saveInspector")}
          </button>
        </div>
      </div>
    </div>
  );
};

export { PropertiesModal };
