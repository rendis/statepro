import { AlertTriangle, Save, X } from "lucide-react";
import { useEffect, useState } from "react";

import { BEHAVIOR_TYPES } from "../../constants";
import { BEHAVIOR_TYPE_LABEL_KEYS } from "../../constants";
import { useI18n } from "../../i18n";
import type { BehaviorRef, BehaviorRegistryItem, BehaviorType, JsonObject } from "../../types";

interface BehaviorModalProps {
  isOpen: boolean;
  type: BehaviorType;
  initialData: BehaviorRef | null;
  onSave: ((item: BehaviorRef) => void) | null;
  onClose: () => void;
  registry: BehaviorRegistryItem[];
}

export const BehaviorModal = ({
  isOpen,
  onClose,
  onSave,
  initialData,
  type,
  registry,
}: BehaviorModalProps) => {
  const { t } = useI18n();
  const [src, setSrc] = useState("");
  const [argsStr, setArgsStr] = useState("");
  const [metaStr, setMetaStr] = useState("");
  const [error, setError] = useState("");

  const availableBehaviors = registry.filter((entry) => entry.type === type);
  const config = BEHAVIOR_TYPES[type] || BEHAVIOR_TYPES.action;
  const Icon = config.icon;

  useEffect(() => {
    if (!isOpen) return;

    setSrc(initialData?.src || "");
    setArgsStr(
      initialData?.args && Object.keys(initialData.args).length > 0
        ? JSON.stringify(initialData.args, null, 2)
        : "",
    );
    setMetaStr(
      initialData?.metadata && Object.keys(initialData.metadata).length > 0
        ? JSON.stringify(initialData.metadata, null, 2)
        : "",
    );
    setError("");
  }, [isOpen, initialData]);

  if (!isOpen) return null;

  const handleSave = () => {
    if (!src.trim()) {
      setError(t("behaviorModal.validation.sourceRequired"));
      return;
    }

    let parsedArgs: JsonObject = {};
    let parsedMetadata: JsonObject = {};

    try {
      if (argsStr.trim()) {
        const raw = JSON.parse(argsStr) as unknown;
        if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
          setError(t("behaviorModal.validation.argsObject"));
          return;
        }
        parsedArgs = raw as JsonObject;
      }
    } catch {
      setError(t("behaviorModal.validation.argsInvalid"));
      return;
    }

    try {
      if (metaStr.trim()) {
        const raw = JSON.parse(metaStr) as unknown;
        if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
          setError(t("behaviorModal.validation.metadataObject"));
          return;
        }
        parsedMetadata = raw as JsonObject;
      }
    } catch {
      setError(t("behaviorModal.validation.metadataInvalid"));
      return;
    }

    onSave?.({
      src: src.trim(),
      args: Object.keys(parsedArgs).length > 0 ? parsedArgs : undefined,
      metadata: Object.keys(parsedMetadata).length > 0 ? parsedMetadata : undefined,
    });
  };

  return (
    <div
      className="fixed inset-0 z-[300] flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm"
      onMouseDown={(event) => event.stopPropagation()}
    >
      <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-md flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between px-5 py-4 border-b border-slate-800 bg-slate-800/50">
          <h3 className={`text-sm font-bold flex items-center gap-2 ${config.color} uppercase tracking-wider`}>
            <Icon size={16} />{" "}
            {t("behaviorModal.title.instantiate", {
              type: t(BEHAVIOR_TYPE_LABEL_KEYS[type], undefined, config.label),
            })}
          </h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
            <X size={18} />
          </button>
        </div>

        <div className="p-5 space-y-4 max-h-[60vh] overflow-y-auto custom-scrollbar">
          {error && (
            <div className="flex items-center gap-2 bg-red-950/50 text-red-400 p-3 rounded-lg border border-red-900/50 text-xs">
              <AlertTriangle size={14} /> {error}
            </div>
          )}

          <div>
            <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
              {t("behaviorModal.source")}
            </label>
            <select
              value={src}
              onChange={(event) => setSrc(event.target.value)}
              className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 font-mono text-sm cursor-pointer"
            >
              <option value="" disabled>
                {t("behaviorModal.selectSource")}
              </option>
              {availableBehaviors.map((behavior) => (
                <option key={behavior.src} value={behavior.src}>
                  {behavior.src}
                </option>
              ))}
            </select>
            {availableBehaviors.length === 0 && (
              <p className="text-red-400 text-[10px] mt-1">
                {t("behaviorModal.noBehaviors", {
                  type: t(BEHAVIOR_TYPE_LABEL_KEYS[type], undefined, type).toLowerCase(),
                })}
              </p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                {t("behaviorModal.args")}
              </label>
              <textarea
                value={argsStr}
                onChange={(event) => setArgsStr(event.target.value)}
                placeholder={'{\n  "param": "val"\n}'}
                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-green-400 focus:outline-none focus:border-blue-500 font-mono text-[10px] h-24 resize-y leading-tight custom-scrollbar"
              />
            </div>
            <div>
              <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                {t("behaviorModal.metadata")}
              </label>
              <textarea
                value={metaStr}
                onChange={(event) => setMetaStr(event.target.value)}
                placeholder="{}"
                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-yellow-400 focus:outline-none focus:border-blue-500 font-mono text-[10px] h-24 resize-y leading-tight custom-scrollbar"
              />
            </div>
          </div>
        </div>

        <div className="p-4 border-t border-slate-800 bg-slate-900 flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2 text-slate-400 hover:text-white text-xs font-medium transition-colors"
          >
            {t("common.cancel")}
          </button>
          <button
            onClick={handleSave}
            disabled={availableBehaviors.length === 0}
            className="flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 disabled:text-slate-500 text-white rounded-md text-xs font-medium shadow-md transition-colors"
          >
            <Save size={14} /> {t("behaviorModal.assign")}
          </button>
        </div>
      </div>
    </div>
  );
};
