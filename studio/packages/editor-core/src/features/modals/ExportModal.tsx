import { AlertTriangle, Code2, Download } from "lucide-react";

import { resolveSerializeIssueMessage, useI18n } from "../../i18n";
import type { SerializeIssue } from "../../types";

interface ExportModalProps {
  isOpen: boolean;
  json: string;
  issues: SerializeIssue[];
  canExport: boolean;
  onClose: () => void;
  title?: string;
  copyLabel?: string;
}

export const ExportModal = ({
  isOpen,
  json,
  issues,
  canExport,
  onClose,
  title,
  copyLabel,
}: ExportModalProps) => {
  const { t } = useI18n();
  if (!isOpen) return null;
  const resolvedTitle = title || t("exportModal.title");
  const resolvedCopyLabel = copyLabel || t("exportModal.copy");

  const errorIssues = issues.filter((issue) => issue.severity === "error");
  const warningIssues = issues.filter((issue) => issue.severity === "warning");

  return (
    <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-[200] flex items-center justify-center p-6">
      <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-4xl max-h-[85vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800 bg-slate-800/50">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <Code2 className="text-blue-400" /> {resolvedTitle}
          </h2>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
            {t("common.close")}
          </button>
        </div>

        {errorIssues.length > 0 && (
          <div className="px-6 py-3 border-b border-red-700/40 bg-red-900/20 text-red-200 text-xs">
            <div className="font-semibold flex items-center gap-2 mb-2">
              <AlertTriangle size={14} /> {t("exportModal.blocked", { count: errorIssues.length })}
            </div>
            <ul className="space-y-1 max-h-24 overflow-y-auto custom-scrollbar">
              {errorIssues.slice(0, 20).map((issue, index) => (
                <li key={`${issue.field}-${index}`}>
                  <span className="font-mono">{issue.field}</span>: {resolveSerializeIssueMessage(issue, t)}
                </li>
              ))}
            </ul>
          </div>
        )}

        {warningIssues.length > 0 && (
          <div className="px-6 py-3 border-b border-amber-700/40 bg-amber-900/20 text-amber-300 text-xs">
            <div className="font-semibold mb-2">{t("exportModal.warnings", { count: warningIssues.length })}</div>
            <ul className="space-y-1 max-h-24 overflow-y-auto custom-scrollbar">
              {warningIssues.slice(0, 20).map((issue, index) => (
                <li key={`${issue.field}-${index}`}>
                  <span className="font-mono">{issue.field}</span>: {resolveSerializeIssueMessage(issue, t)}
                </li>
              ))}
            </ul>
          </div>
        )}

        <div className="p-6 overflow-y-auto flex-1 bg-slate-950">
          <pre className="text-sm font-mono text-green-400 whitespace-pre-wrap">{json}</pre>
        </div>

        <div className="p-4 border-t border-slate-800 bg-slate-900 flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-white rounded-md text-sm font-medium transition-colors"
          >
            {t("common.close")}
          </button>
          <button
            onClick={() => navigator.clipboard.writeText(json)}
            disabled={!canExport}
            className="flex items-center gap-2 px-4 py-2 bg-blue-700 hover:bg-blue-600 disabled:bg-slate-700 disabled:text-slate-500 border border-blue-500/30 disabled:border-slate-700 text-white rounded-md text-sm font-medium transition-colors"
          >
            <Download size={16} /> {resolvedCopyLabel}
          </button>
        </div>
      </div>
    </div>
  );
};
