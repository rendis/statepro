import { AlertTriangle, Code2, Download, Upload, X } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { parseStudioLayout, validateStateProMachine } from "../../model";
import { studioTabClass } from "../../constants";
import { resolveSerializeIssueMessage, useI18n } from "../../i18n";
import type {
  SerializeIssue,
  StateProMachine,
  StudioLayoutDocument,
  StudioLayoutIssue,
} from "../../types";

type ModalTab = "import" | "export";
type JsonDocumentType = "model" | "layout";

interface JsonIOModalProps {
  isOpen: boolean;
  onClose: () => void;
  allowImport?: boolean;
  allowExport?: boolean;
  modelJson: string;
  layoutJson: string;
  modelIssues: SerializeIssue[];
  canExportModel: boolean;
  onImportModel: (machine: StateProMachine) => void;
  onImportLayout: (layout: StudioLayoutDocument, issues: StudioLayoutIssue[]) => void;
}

type ImportValidationResult = {
  canImport: boolean;
  issues: Array<SerializeIssue | StudioLayoutIssue>;
  model: StateProMachine | null;
  layout: StudioLayoutDocument | null;
};

const DOCUMENT_TYPE_OPTIONS: Array<{ id: JsonDocumentType; labelKey: string }> = [
  { id: "model", labelKey: "jsonModal.document.model" },
  { id: "layout", labelKey: "jsonModal.document.layout" },
];

export const JsonIOModal = ({
  isOpen,
  onClose,
  allowImport = true,
  allowExport = true,
  modelJson,
  layoutJson,
  modelIssues,
  canExportModel,
  onImportModel,
  onImportLayout,
}: JsonIOModalProps) => {
  const { t } = useI18n();
  const defaultTab: ModalTab = allowImport ? "import" : "export";
  const [activeTab, setActiveTab] = useState<ModalTab>(defaultTab);
  const [documentType, setDocumentType] = useState<JsonDocumentType>("model");
  const [jsonText, setJsonText] = useState("");
  const [parseError, setParseError] = useState<string | null>(null);

  const visibleTabs: ModalTab[] = useMemo(() => {
    const tabs: ModalTab[] = [];
    if (allowImport) {
      tabs.push("import");
    }
    if (allowExport) {
      tabs.push("export");
    }
    return tabs;
  }, [allowExport, allowImport]);

  if (visibleTabs.length === 0) {
    return null;
  }

  const importValidation = useMemo<ImportValidationResult>(() => {
    if (!jsonText.trim()) {
      return {
        canImport: false,
        issues: [],
        model: null,
        layout: null,
      };
    }

    if (documentType === "model") {
      try {
        const parsed = JSON.parse(jsonText) as StateProMachine;
        const result = validateStateProMachine(parsed);
        return {
          canImport: result.canExport,
          issues: result.issues,
          model: parsed,
          layout: null,
        };
      } catch {
        return {
          canImport: false,
          issues: [],
          model: null,
          layout: null,
        };
      }
    }

    const parsedLayout = parseStudioLayout(jsonText);
    return {
      canImport: parsedLayout.canImport,
      issues: parsedLayout.issues,
      model: null,
      layout: parsedLayout.document,
    };
  }, [documentType, jsonText]);

  useEffect(() => {
    if (!visibleTabs.includes(activeTab)) {
      setActiveTab(visibleTabs[0]);
    }
  }, [activeTab, visibleTabs]);

  if (!isOpen) return null;

  const exportIssues = documentType === "model" ? modelIssues : [];
  const exportErrorIssues = exportIssues.filter((issue) => issue.severity === "error");
  const exportWarningIssues = exportIssues.filter((issue) => issue.severity === "warning");
  const exportCanCopy = documentType === "model" ? canExportModel : true;
  const exportJson = documentType === "model" ? modelJson : layoutJson;

  const importErrorIssues = importValidation.issues.filter((issue) => issue.severity === "error");
  const importWarningIssues = importValidation.issues.filter((issue) => issue.severity === "warning");

  const importPlaceholder =
    documentType === "model"
      ? t("importModal.placeholder")
      : t("jsonModal.import.placeholderLayout");
  const importActionLabel =
    documentType === "model"
      ? t("jsonModal.import.model")
      : t("jsonModal.import.layout");

  const resetImportEditor = () => {
    setJsonText("");
    setParseError(null);
  };

  const handleImport = () => {
    if (!allowImport) {
      setParseError(t("importModal.validationFailed"));
      return;
    }

    if (documentType === "model") {
      if (!importValidation.model || !importValidation.canImport) {
        setParseError(t("importModal.validationFailed"));
        return;
      }
      onImportModel(importValidation.model);
      resetImportEditor();
      return;
    }

    if (!importValidation.layout || !importValidation.canImport) {
      setParseError(t("jsonModal.import.layoutInvalid"));
      return;
    }

    onImportLayout(importValidation.layout, importValidation.issues as StudioLayoutIssue[]);
    resetImportEditor();
  };

  return (
    <div className="fixed inset-0 bg-slate-950/80 backdrop-blur-sm z-[220] flex items-center justify-center p-6">
      <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-4xl max-h-[85vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800 bg-slate-800/50">
          <h2 className="text-lg font-semibold flex items-center gap-2 text-white">
            <Code2 className="text-blue-400" /> {t("jsonModal.title")}
          </h2>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
            <X size={18} />
          </button>
        </div>

        <div className="flex items-center justify-between px-6 py-3 border-b border-slate-800 bg-slate-900/80">
          <div className="flex items-center gap-5">
            {allowImport && (
              <button onClick={() => setActiveTab("import")} className={studioTabClass(activeTab === "import")}>
                {t("jsonModal.tab.import")}
              </button>
            )}
            {allowExport && (
              <button onClick={() => setActiveTab("export")} className={studioTabClass(activeTab === "export")}>
                {t("jsonModal.tab.export")}
              </button>
            )}
          </div>
          <div className="flex items-center rounded-md border border-slate-700 bg-slate-900 p-1">
            {DOCUMENT_TYPE_OPTIONS.map((option) => {
              const isActive = documentType === option.id;
              return (
                <button
                  key={option.id}
                  onClick={() => {
                    setDocumentType(option.id);
                    setParseError(null);
                  }}
                  className={`px-3 py-1 text-xs font-semibold rounded transition-colors ${
                    isActive ? "bg-blue-600 text-white" : "text-slate-400 hover:text-white"
                  }`}
                >
                  {t(option.labelKey)}
                </button>
              );
            })}
          </div>
        </div>

        {activeTab === "import" ? (
          <>
            {parseError && (
              <div className="px-6 py-3 border-b border-red-700/40 bg-red-900/20 text-red-200 text-xs">
                {parseError}
              </div>
            )}

            {importErrorIssues.length > 0 && (
              <div className="px-6 py-3 border-b border-red-700/40 bg-red-900/20 text-red-200 text-xs">
                <div className="font-semibold flex items-center gap-2 mb-1">
                  <AlertTriangle size={14} />{" "}
                  {t("importModal.validationTitle", { count: importErrorIssues.length })}
                </div>
                <ul className="max-h-24 overflow-y-auto custom-scrollbar space-y-1">
                  {importErrorIssues.slice(0, 20).map((issue, index) => (
                    <li key={`${issue.field}-${index}`}>
                      <span className="font-mono">{issue.field}</span>:{" "}
                      {resolveSerializeIssueMessage(issue as SerializeIssue, t)}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {importWarningIssues.length > 0 && (
              <div className="px-6 py-3 border-b border-amber-700/40 bg-amber-900/20 text-amber-300 text-xs">
                <div className="font-semibold mb-1">
                  {t("exportModal.warnings", { count: importWarningIssues.length })}
                </div>
                <ul className="max-h-24 overflow-y-auto custom-scrollbar space-y-1">
                  {importWarningIssues.slice(0, 20).map((issue, index) => (
                    <li key={`${issue.field}-${index}`}>
                      <span className="font-mono">{issue.field}</span>:{" "}
                      {resolveSerializeIssueMessage(issue as SerializeIssue, t)}
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
                placeholder={importPlaceholder}
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
                disabled={!importValidation.canImport}
                className="flex items-center gap-2 px-4 py-2 bg-blue-700 hover:bg-blue-600 disabled:bg-slate-700 disabled:text-slate-500 border border-blue-500/30 disabled:border-slate-700 text-white rounded-md text-sm"
              >
                <Upload size={16} /> {importActionLabel}
              </button>
            </div>
          </>
        ) : (
          <>
            {exportErrorIssues.length > 0 && (
              <div className="px-6 py-3 border-b border-red-700/40 bg-red-900/20 text-red-200 text-xs">
                <div className="font-semibold flex items-center gap-2 mb-2">
                  <AlertTriangle size={14} /> {t("exportModal.blocked", { count: exportErrorIssues.length })}
                </div>
                <ul className="space-y-1 max-h-24 overflow-y-auto custom-scrollbar">
                  {exportErrorIssues.slice(0, 20).map((issue, index) => (
                    <li key={`${issue.field}-${index}`}>
                      <span className="font-mono">{issue.field}</span>: {resolveSerializeIssueMessage(issue, t)}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {exportWarningIssues.length > 0 && (
              <div className="px-6 py-3 border-b border-amber-700/40 bg-amber-900/20 text-amber-300 text-xs">
                <div className="font-semibold mb-1">
                  {t("exportModal.warnings", { count: exportWarningIssues.length })}
                </div>
                <ul className="max-h-24 overflow-y-auto custom-scrollbar space-y-1">
                  {exportWarningIssues.slice(0, 20).map((issue, index) => (
                    <li key={`${issue.field}-${index}`}>
                      <span className="font-mono">{issue.field}</span>:{" "}
                      {resolveSerializeIssueMessage(issue, t)}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            <div className="p-6 overflow-y-auto flex-1 bg-slate-950">
              <pre className="text-sm font-mono text-green-400 whitespace-pre-wrap">{exportJson}</pre>
            </div>

            <div className="p-4 border-t border-slate-800 bg-slate-900 flex justify-end gap-2">
              <button
                onClick={onClose}
                className="px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                {t("common.close")}
              </button>
              <button
                onClick={() => navigator.clipboard.writeText(exportJson)}
                disabled={!exportCanCopy}
                className="flex items-center gap-2 px-4 py-2 bg-blue-700 hover:bg-blue-600 disabled:bg-slate-700 disabled:text-slate-500 border border-blue-500/30 disabled:border-slate-700 text-white rounded-md text-sm font-medium transition-colors"
              >
                <Download size={16} /> {t("exportModal.copy")}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};
