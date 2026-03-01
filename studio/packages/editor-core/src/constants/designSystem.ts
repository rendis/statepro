export const STUDIO_DS = {
  modalOverlay: "fixed inset-0 bg-slate-950/70 backdrop-blur-sm z-[100] flex items-center justify-center p-4",
  modalPanel:
    "bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full animate-in zoom-in-95 duration-200 flex flex-col overflow-hidden",
  modalHeader: "flex items-center justify-between px-5 pt-4 pb-2 bg-slate-800/50",
  modalTitle: "text-base font-semibold flex items-center gap-2 text-white capitalize",
  tabsBar: "flex border-b border-slate-800 bg-slate-800/50 px-5 gap-6",
  tabButtonBase: "pb-2 text-[11px] uppercase tracking-wider font-bold border-b-2 transition-colors",
  panelBody: "p-5 space-y-4 overflow-y-auto custom-scrollbar bg-slate-900",
  panelCard: "bg-slate-950/50 p-3 rounded-lg border border-slate-800/50",
  labelMono: "block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider",
  input:
    "w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500",
  inputSm:
    "w-full bg-slate-950 border border-slate-700 rounded px-3 py-1.5 text-slate-200 focus:outline-none focus:border-blue-500 font-mono text-sm",
  textarea:
    "w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-xs resize-none",
  primaryButton:
    "px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-md text-sm font-medium shadow-md transition-colors",
} as const;

export const studioTabClass = (active: boolean): string =>
  `${STUDIO_DS.tabButtonBase} ${
    active ? "border-blue-500 text-blue-400" : "border-transparent text-slate-500 hover:text-slate-300"
  }`;
