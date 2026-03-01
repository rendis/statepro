import { AlertTriangle, Upload, X } from "lucide-react";
import { useMemo, useState } from "react";

import { resolveSerializeIssueMessage, useI18n } from "../../i18n";
import type { SerializeIssue, StateProMachine } from "../../types";
import { validateStateProMachine } from "../../model";

interface ImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImport: (machine: StateProMachine) => void;
}

export const ImportModal = ({ isOpen, onClose, onImport }: ImportModalProps) => {
  const { t } = useI18n();
  const [jsonText, setJsonText] = useState("");
  const [parseError, setParseError] = useState<string | null>(null);

  const validation = useMemo(() => {
    if (!jsonText.trim()) {
      return { issues: [] as SerializeIssue[], canImport: false };
    }

    try {
      const parsed = JSON.parse(jsonText) as StateProMachine;
      const result = validateStateProMachine(parsed);
      return {
        issues: result.issues,
        canImport: result.canExport,
      };
    } catch {
      return {
        issues: [],
        canImport: false,
      };
    }
  }, [jsonText]);

  if (!isOpen) return null;

  const handleImport = () => {
    try {
      const parsed = JSON.parse(jsonText) as StateProMachine;
      const result = validateStateProMachine(parsed);
      if (!result.canExport) {
        setParseError(t("importModal.validationFailed"));
        return;
      }

      onImport(parsed);
      setParseError(null);
      setJsonText("");
    } catch {
      setParseError(t("importModal.invalidJson"));
    }
  };

  const errorIssues = validation.issues.filter((issue) => issue.severity === "error");

  return (
    <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-[220] flex items-center justify-center p-6">
      <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-4xl max-h-[85vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800 bg-slate-800/50">
          <h2 className="text-lg font-semibold flex items-center gap-2 text-white">
            <Upload className="text-blue-400" /> {t("importModal.title")}
          </h2>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
            <X size={18} />
          </button>
        </div>

        {parseError && (
          <div className="px-6 py-3 border-b border-red-700/40 bg-red-900/20 text-red-200 text-xs">
            {parseError}
          </div>
        )}

        {errorIssues.length > 0 && (
          <div className="px-6 py-3 border-b border-red-700/40 bg-red-900/20 text-red-200 text-xs">
            <div className="font-semibold flex items-center gap-2 mb-1">
              <AlertTriangle size={14} /> {t("importModal.validationTitle", { count: errorIssues.length })}
            </div>
            <ul className="max-h-24 overflow-y-auto custom-scrollbar space-y-1">
              {errorIssues.slice(0, 20).map((issue, index) => (
                <li key={`${issue.field}-${index}`}>
                  <span className="font-mono">{issue.field}</span>: {resolveSerializeIssueMessage(issue, t)}
                </li>
              ))}
            </ul>
          </div>
        )}

        <div className="p-6 flex-1 overflow-y-auto bg-slate-950">
          <textarea
            value={jsonText}
            onChange={(event) => {
              setJsonText(event.target.value);
              setParseError(null);
            }}
            className="w-full h-[420px] bg-slate-900 border border-slate-700 rounded p-3 font-mono text-xs text-green-300 focus:outline-none focus:border-blue-500"
            placeholder={t("importModal.placeholder")}
          />
        </div>

        <div className="p-4 border-t border-slate-800 bg-slate-900 flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-white rounded-md text-sm"
          >
            {t("common.cancel")}
          </button>
          <button
            onClick={handleImport}
            disabled={!validation.canImport}
            className="px-4 py-2 bg-blue-700 hover:bg-blue-600 disabled:bg-slate-700 disabled:text-slate-500 border border-blue-500/30 disabled:border-slate-700 text-white rounded-md text-sm"
          >
            {t("importModal.import")}
          </button>
        </div>
      </div>
    </div>
  );
};
