import { Edit3, Plus } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { STUDIO_ICON_REGISTRY, STUDIO_ICONS } from "../../constants";
import { useI18n } from "../../i18n";

interface EventSelectorProps {
  value?: string;
  availableEvents?: string[];
  onChange: (value: string) => void;
}

export const EventSelector = ({ value, onChange, availableEvents = [] }: EventSelectorProps) => {
  const { t } = useI18n();
  const [isEditing, setIsEditing] = useState(!value);
  const [inputValue, setInputValue] = useState(value || "");
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement | null>(null);

  const handleSelect = (nextEvent: string) => {
    const cleanedEvent = nextEvent.trim();
    if (!cleanedEvent) {
      return false;
    }

    if (cleanedEvent !== value) {
      onChange(cleanedEvent);
    }
    setInputValue(cleanedEvent);
    setIsEditing(false);
    setDropdownOpen(false);
    return true;
  };

  useEffect(() => {
    setInputValue(value || "");
  }, [value]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        const wasCommitted = handleSelect(inputValue);
        if (!wasCommitted) {
          setInputValue(value || "");
          setDropdownOpen(false);
          setIsEditing(!value);
        }
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [handleSelect, inputValue, value]);

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.preventDefault();
      handleSelect(inputValue);
    } else if (event.key === "Escape") {
      setIsEditing(false);
      setDropdownOpen(false);
      setInputValue(value || "");
    }
  };

  const filteredAvailable = availableEvents.filter(
    (candidate) => candidate !== value && candidate.toLowerCase().includes(inputValue.toLowerCase()),
  );
  const pendingEvent = inputValue.trim();
  const EventIcon = STUDIO_ICONS.transition.trigger.on;
  const triggerOnColors = STUDIO_ICON_REGISTRY.transition.trigger.on.colors;

  if (!isEditing && value) {
    return (
      <div
        onClick={() => setIsEditing(true)}
        className="group cursor-pointer flex items-center justify-between w-full bg-blue-900/20 border border-blue-500/50 hover:border-blue-500 hover:bg-blue-900/30 text-blue-300 px-3 py-2 rounded-lg font-mono text-sm transition-colors"
        title={t("eventSelector.clickToChange")}
      >
        <div className="flex items-center gap-2">
          <EventIcon size={14} className={triggerOnColors.base} />
          <span className="font-bold tracking-wider">{value}</span>
        </div>
        <Edit3
          size={14}
          className="text-blue-400 opacity-0 group-hover:opacity-100 transition-opacity"
        />
      </div>
    );
  }

  return (
    <div className="relative w-full" ref={dropdownRef}>
      <div className="flex items-center gap-2 bg-slate-950 border border-blue-500 rounded-lg px-3 py-2 transition-colors">
        <EventIcon size={14} className={triggerOnColors.base} />
        <input
          type="text"
          autoFocus
          value={inputValue}
          onChange={(event) => {
            setInputValue(event.target.value);
            setDropdownOpen(true);
          }}
          onFocus={() => setDropdownOpen(true)}
          onKeyDown={handleKeyDown}
          placeholder={t("eventSelector.placeholder")}
          className="bg-transparent border-none focus:outline-none text-sm font-bold font-mono text-slate-200 flex-1 tracking-wider placeholder:text-slate-600 placeholder:font-normal"
        />
      </div>

      {dropdownOpen && (filteredAvailable.length > 0 || (pendingEvent && pendingEvent !== value)) && (
        <div className="absolute top-full mt-2 left-0 w-full bg-slate-800 border border-slate-700 rounded-lg shadow-2xl z-[70] max-h-48 overflow-y-auto py-1 animate-in slide-in-from-top-2 duration-200">
          {pendingEvent && !availableEvents.includes(pendingEvent) && pendingEvent !== value && (
            <button
              onMouseDown={(event) => {
                event.preventDefault();
                handleSelect(pendingEvent);
              }}
            className="w-full text-left px-3 py-2.5 hover:bg-slate-700 flex items-center gap-2 group transition-colors border-l-2 border-transparent hover:border-blue-500"
          >
            <Plus size={14} className="text-blue-400" />
            <span className="text-xs text-white font-bold font-mono">
              {t("eventSelector.create", { value: pendingEvent })}
            </span>
          </button>
          )}
          {filteredAvailable.map((eventName) => (
            <button
              key={eventName}
              onMouseDown={(event) => {
                event.preventDefault();
                handleSelect(eventName);
              }}
              className="w-full text-left px-3 py-2.5 hover:bg-slate-700 flex items-center gap-2 group transition-colors border-l-2 border-transparent hover:border-slate-500"
            >
              <EventIcon
                size={14}
                className={`${triggerOnColors.muted ?? triggerOnColors.base} transition-colors`}
              />
              <span className="text-xs text-slate-300 group-hover:text-white font-mono transition-colors">
                {eventName}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
