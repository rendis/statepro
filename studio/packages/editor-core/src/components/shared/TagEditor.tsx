import { Tag, Plus, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { useI18n } from "../../i18n";
import { cleanIdentifier, formatIdentifier } from "../../utils";

interface TagEditorProps {
  tags?: string[];
  availableTags?: string[];
  onChange: (tags: string[]) => void;
}

export const TagEditor = ({ tags = [], onChange, availableTags = [] }: TagEditorProps) => {
  const { t } = useI18n();
  const [inputValue, setInputValue] = useState("");
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleAdd = (tag: string) => {
    const cleanTag = cleanIdentifier(tag);
    if (cleanTag && !tags.includes(cleanTag)) {
      onChange([...tags, cleanTag]);
    }
    setInputValue("");
    setDropdownOpen(false);
  };

  const handleRemove = (tagToRemove: string) => {
    onChange(tags.filter((tag) => tag !== tagToRemove));
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.preventDefault();
      handleAdd(inputValue);
    }
  };

  const filteredAvailable = availableTags.filter(
    (candidate) => !tags.includes(candidate) && candidate.includes(cleanIdentifier(inputValue)),
  );
  const pendingTag = cleanIdentifier(inputValue);

  return (
    <div className="relative" ref={dropdownRef}>
      <div
        className="w-full bg-slate-950 border border-slate-700 rounded px-2 py-1.5 flex flex-wrap gap-1.5 min-h-[36px] cursor-text focus-within:border-blue-500 transition-colors"
        onClick={() => setDropdownOpen(true)}
      >
        {tags.map((tag) => (
          <span
            key={tag}
            className="flex items-center gap-1 bg-slate-800 border border-slate-600 text-slate-300 text-[10px] px-2 py-0.5 rounded-full font-mono shadow-sm"
          >
            <Tag size={8} className="text-slate-500" />
            {tag}
            <button
              onClick={(event) => {
                event.stopPropagation();
                handleRemove(tag);
              }}
              className="hover:text-red-400 transition-colors ml-1"
            >
              <X size={10} />
            </button>
          </span>
        ))}
        <input
          type="text"
          value={inputValue}
          onChange={(event) => setInputValue(formatIdentifier(event.target.value))}
          onKeyDown={handleKeyDown}
          placeholder={tags.length === 0 ? t("tagEditor.placeholder") : ""}
          className="bg-transparent border-none focus:outline-none text-[11px] font-mono text-slate-200 flex-1 min-w-[120px]"
        />
      </div>

      {dropdownOpen && (filteredAvailable.length > 0 || pendingTag) && (
        <div className="absolute top-full mt-1 left-0 w-full bg-slate-800 border border-slate-700 rounded-lg shadow-2xl z-[70] max-h-48 overflow-y-auto py-1">
          {pendingTag &&
            !tags.includes(pendingTag) &&
            !filteredAvailable.includes(pendingTag) && (
              <button
                onMouseDown={(event) => {
                  event.preventDefault();
                  handleAdd(pendingTag);
                }}
                className="w-full text-left px-3 py-2 hover:bg-slate-700 flex items-center gap-2 group transition-colors border-l-2 border-transparent hover:border-blue-500"
              >
                <Plus size={12} className="text-blue-400" />
                <span className="text-[11px] text-white font-mono">
                  {t("tagEditor.create", { value: pendingTag })}
                </span>
              </button>
            )}
          {filteredAvailable.map((tag) => (
            <button
              key={tag}
              onMouseDown={(event) => {
                event.preventDefault();
                handleAdd(tag);
              }}
              className="w-full text-left px-3 py-2 hover:bg-slate-700 flex items-center gap-2 group transition-colors border-l-2 border-transparent hover:border-slate-500"
            >
              <Tag size={12} className="text-slate-400" />
              <span className="text-[11px] text-slate-300 font-mono">{tag}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
