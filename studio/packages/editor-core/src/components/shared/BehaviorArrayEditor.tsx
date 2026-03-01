import { Plus, X } from "lucide-react";
import type { Dispatch, SetStateAction } from "react";

import { BEHAVIOR_TYPE_LABEL_KEYS, BEHAVIOR_TYPES } from "../../constants";
import { useI18n } from "../../i18n";
import type { BehaviorRef, BehaviorType, BehaviorRegistryItem, BehaviorModalState } from "../../types";

interface BehaviorArrayEditorProps {
  items?: BehaviorRef[];
  type: BehaviorType;
  registry?: BehaviorRegistryItem[];
  onChange: (items: BehaviorRef[]) => void;
  openBehaviorModal: Dispatch<SetStateAction<BehaviorModalState>>;
}

export const BehaviorArrayEditor = ({
  items = [],
  onChange,
  type,
  openBehaviorModal,
  registry = [],
}: BehaviorArrayEditorProps) => {
  const { t } = useI18n();
  const config = BEHAVIOR_TYPES[type] || BEHAVIOR_TYPES.action;
  const Icon = config.icon;

  const handleAddNew = () => {
    openBehaviorModal({
      isOpen: true,
      type,
      initialData: null,
      onSave: (newItem) => {
        onChange([...items, newItem]);
        openBehaviorModal({ isOpen: false, type, initialData: null, onSave: null });
      },
    });
  };

  const handleEdit = (index: number) => {
    openBehaviorModal({
      isOpen: true,
      type,
      initialData: items[index],
      onSave: (updatedItem) => {
        const nextItems = [...items];
        nextItems[index] = updatedItem;
        onChange(nextItems);
        openBehaviorModal({ isOpen: false, type, initialData: null, onSave: null });
      },
    });
  };

  const handleRemove = (event: React.MouseEvent<HTMLButtonElement>, index: number) => {
    event.stopPropagation();
    onChange(items.filter((_, itemIndex) => itemIndex !== index));
  };

  return (
    <div className="flex flex-wrap gap-1.5 items-center">
      {items.map((item, index) => {
        const libraryDef = registry.find((entry) => entry.src === item.src);

        return (
          <span
            key={`${item.src}-${index}`}
            onClick={() => handleEdit(index)}
            className={`flex items-center gap-1.5 ${config.bg} ${config.border} border ${config.color} text-[10px] px-2 py-1 rounded font-mono shadow-sm cursor-pointer hover:brightness-125 transition-all`}
            title={libraryDef?.description || item.src}
          >
            <Icon size={10} /> {item.src}
            {item.args && Object.keys(item.args).length > 0 && (
              <span
                className="w-1.5 h-1.5 bg-green-500 rounded-full"
                title={t("behaviorArray.hasArgs")}
              />
            )}
            <button
              onClick={(event) => handleRemove(event, index)}
              className="hover:text-white transition-colors ml-1"
            >
              <X size={10} />
            </button>
          </span>
        );
      })}
      <button
        onClick={handleAddNew}
        className="flex items-center gap-1 text-[10px] px-2 py-1 rounded border border-dashed border-slate-600 text-slate-400 hover:text-slate-200 hover:border-slate-400 transition-colors"
      >
        <Plus size={10} />{" "}
        {t(BEHAVIOR_TYPE_LABEL_KEYS[type], undefined, type)}
      </button>
    </div>
  );
};
