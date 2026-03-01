import { X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";

import { STUDIO_ICON_REGISTRY, STUDIO_ICONS } from "../../constants";
import { useI18n } from "../../i18n";
import {
  buildTransitionTargetDisplayItems,
  buildTransitionTargetOptions,
  type TransitionTargetOption,
} from "../../utils";
import type { EditorNode } from "../../types";
import { TooltipIconButton } from "./tooltip";

interface TransitionTargetSelectorProps {
  sourceRealityNodeId: string;
  targets: string[];
  nodes: EditorNode[];
  onChange: (targets: string[]) => void;
}

export const TransitionTargetSelector = ({
  sourceRealityNodeId,
  targets,
  nodes,
  onChange,
}: TransitionTargetSelectorProps) => {
  const { t } = useI18n();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement | null>(null);
  const displayItems = useMemo(
    () => buildTransitionTargetDisplayItems(targets, sourceRealityNodeId, nodes),
    [nodes, sourceRealityNodeId, targets],
  );
  const options = useMemo(
    () => buildTransitionTargetOptions(sourceRealityNodeId, nodes),
    [nodes, sourceRealityNodeId],
  );
  const selected = useMemo(() => new Set(targets), [targets]);

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

  const addTarget = (value: string) => {
    if (selected.has(value)) {
      setDropdownOpen(false);
      return;
    }

    onChange([...targets, value]);
    setDropdownOpen(false);
  };

  const removeTarget = (value: string, index: number) => {
    onChange(targets.filter((target, targetIndex) => !(target === value && targetIndex === index)));
  };

  const groupedOptions = useMemo(() => {
    const groups: Record<"internal" | "universe" | "universeReality", TransitionTargetOption[]> = {
      internal: [],
      universe: [],
      universeReality: [],
    };

    options.forEach((option) => {
      if (option.kind === "internalReality") {
        groups.internal.push(option);
        return;
      }
      if (option.kind === "externalUniverse") {
        groups.universe.push(option);
        return;
      }
      groups.universeReality.push(option);
    });

    return groups;
  }, [options]);

  const UniverseIcon = STUDIO_ICONS.entity.universe;
  const RealityIcon = STUDIO_ICONS.entity.reality;
  const WarningIcon = STUDIO_ICONS.status.warning;
  const universeColors = STUDIO_ICON_REGISTRY.entity.universe.colors;
  const realityColors = STUDIO_ICON_REGISTRY.entity.reality.colors;
  const warningColors = STUDIO_ICON_REGISTRY.status.warning.colors;

  return (
    <div className="relative" ref={dropdownRef}>
      <div
        data-testid="transition-target-selector"
        className="w-full bg-slate-950 border border-slate-700 rounded px-2 py-1.5 flex flex-wrap gap-1.5 min-h-[36px] cursor-text focus-within:border-blue-500 transition-colors"
        onClick={() => setDropdownOpen(true)}
      >
        {displayItems.length === 0 && !dropdownOpen && (
          <span className="text-[10px] text-slate-600 self-center ml-1 font-mono">
            {t("transitionTargetSelector.placeholder")}
          </span>
        )}

        {displayItems.map((item, index) => (
          <span
            key={`${item.value}-${index}`}
            data-testid={`transition-target-chip-${index}`}
            className={`flex items-center gap-1.5 text-[10px] px-2 py-0.5 rounded font-mono shadow-sm ${
              item.valid
                ? "bg-blue-900/30 border border-blue-500/30 text-blue-300"
                : "bg-red-900/20 border border-red-500/30 text-red-200"
            }`}
          >
            {item.kind === "internalReality" || item.kind === "externalReality" ? (
              <RealityIcon size={10} className={realityColors.base} />
            ) : item.kind === "externalUniverse" ? (
              <UniverseIcon size={10} className={universeColors.base} />
            ) : (
              <WarningIcon size={10} className={warningColors.base} />
            )}

            <span>{item.label}</span>
            {!item.valid && (
              <span
                data-testid={`transition-target-chip-invalid-${index}`}
                className={`uppercase tracking-wide ${warningColors.base}`}
              >
                {t("transitionTargetSelector.invalid")}
              </span>
            )}

            <TooltipIconButton
              onClick={(event) => {
                event.stopPropagation();
                removeTarget(item.value, index);
              }}
              tooltip={t("properties.transition.deleteTarget")}
              className="hover:text-white transition-colors"
            >
              <X size={10} />
            </TooltipIconButton>
          </span>
        ))}
      </div>

      {dropdownOpen && (
        <div className="absolute top-full mt-1 left-0 w-full bg-slate-800 border border-slate-700 rounded-lg shadow-2xl z-[70] max-h-56 overflow-y-auto py-1">
          {options.length === 0 ? (
            <div className="p-3 text-xs text-slate-500 text-center">
              {t("transitionTargetSelector.noOptions")}
            </div>
          ) : (
            <>
              {groupedOptions.internal.length > 0 && (
                <div className="pb-1">
                  <div className="px-3 py-1 text-[10px] text-slate-500 uppercase tracking-wider font-semibold">
                    {t("transitionTargetSelector.internalRealities")}
                  </div>
                  {groupedOptions.internal.map((option) => (
                    <button
                      key={option.value}
                      onClick={() => addTarget(option.value)}
                      disabled={selected.has(option.value)}
                      className="w-full text-left pl-3 pr-3 py-1.5 hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2 transition-colors border-l-2 border-transparent hover:border-green-500"
                    >
                      <RealityIcon size={10} className={realityColors.base} />
                      <span className="text-[10px] text-slate-200 font-mono">{option.label}</span>
                    </button>
                  ))}
                </div>
              )}

              {groupedOptions.universe.length > 0 && (
                <div className="pb-1">
                  <div className="px-3 py-1 text-[10px] text-slate-500 uppercase tracking-wider font-semibold">
                    {t("transitionTargetSelector.universes")}
                  </div>
                  {groupedOptions.universe.map((option) => (
                    <button
                      key={option.value}
                      onClick={() => addTarget(option.value)}
                      disabled={selected.has(option.value)}
                      className="w-full text-left px-3 py-2 hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2 transition-colors border-l-2 border-transparent hover:border-blue-500"
                    >
                      <UniverseIcon size={12} className={universeColors.base} />
                      <span className="text-[11px] text-white font-bold font-mono">{option.label}</span>
                    </button>
                  ))}
                </div>
              )}

              {groupedOptions.universeReality.length > 0 && (
                <div>
                  <div className="px-3 py-1 text-[10px] text-slate-500 uppercase tracking-wider font-semibold">
                    {t("transitionTargetSelector.universeRealities")}
                  </div>
                  {groupedOptions.universeReality.map((option) => (
                    <button
                      key={option.value}
                      onClick={() => addTarget(option.value)}
                      disabled={selected.has(option.value)}
                      className="w-full text-left pl-8 pr-3 py-1.5 hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed flex items-center gap-2 transition-colors border-l-2 border-transparent hover:border-green-500"
                    >
                      <RealityIcon size={10} className={realityColors.base} />
                      <span className="text-[10px] text-slate-300 font-mono">{option.label}</span>
                    </button>
                  ))}
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
};
