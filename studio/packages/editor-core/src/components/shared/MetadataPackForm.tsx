import { useEffect, useMemo, useState } from "react";

import { STUDIO_DS } from "../../constants";
import type {
  JsonObject,
  JsonValue,
  MetadataPackDefinition,
  MetadataPackUiOverride,
} from "../../types";
import { useI18n } from "../../i18n";
import { getValueAtPointer, setValueAtPointer } from "../../utils";

interface MetadataPackFormProps {
  pack: MetadataPackDefinition;
  value: JsonObject;
  onChange: (next: JsonObject) => void;
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const normalizeType = (schemaNode: Record<string, unknown>): string => {
  const type = schemaNode.type;
  if (typeof type === "string") {
    return type;
  }
  if (Array.isArray(type)) {
    const preferred = type.find((candidate) => candidate !== "null");
    if (typeof preferred === "string") {
      return preferred;
    }
  }
  if (Array.isArray(schemaNode.enum)) {
    return "string";
  }
  if (isRecord(schemaNode.properties)) {
    return "object";
  }
  return "string";
};

const schemaProperties = (
  schemaNode: Record<string, unknown>,
): Array<{ key: string; schema: Record<string, unknown> }> => {
  const properties = schemaNode.properties;
  if (!isRecord(properties)) {
    return [];
  }

  return Object.entries(properties)
    .map(([key, value]) => ({
      key,
      schema: isRecord(value) ? value : {},
    }))
    .sort((left, right) => left.key.localeCompare(right.key));
};

const pointerFor = (parent: string, key: string): string => `${parent}/${key.replace(/\//g, "~1")}`;

const toTextValue = (value: JsonValue | undefined): string =>
  value === undefined || value === null ? "" : String(value);

const ensureObject = (value: JsonValue | undefined): JsonObject =>
  value && typeof value === "object" && !Array.isArray(value) ? (value as JsonObject) : {};

const isJsonValue = (value: unknown): value is JsonValue => {
  if (
    value === null ||
    typeof value === "string" ||
    typeof value === "number" ||
    typeof value === "boolean"
  ) {
    return true;
  }
  if (Array.isArray(value)) {
    return value.every((entry) => isJsonValue(entry));
  }
  if (!value || typeof value !== "object") {
    return false;
  }
  return Object.values(value as Record<string, unknown>).every((entry) => isJsonValue(entry));
};

const jsonEquals = (left: JsonValue | undefined, right: JsonValue | undefined): boolean => {
  try {
    return JSON.stringify(left) === JSON.stringify(right);
  } catch {
    return false;
  }
};

export const MetadataPackForm = ({ pack, value, onChange }: MetadataPackFormProps) => {
  const { t } = useI18n();
  const [jsonErrors, setJsonErrors] = useState<Record<string, string>>({});

  const uiByPointer = useMemo(() => pack.ui || {}, [pack.ui]);
  const constantEntries = useMemo(
    () =>
      Object.entries(uiByPointer)
        .filter(
          ([, ui]) =>
            Boolean(ui?.constant) && isJsonValue((ui as MetadataPackUiOverride).constantValue),
        )
        .map(([pointer, ui]) => ({
          pointer,
          value: (ui as MetadataPackUiOverride).constantValue as JsonValue,
        })),
    [uiByPointer],
  );

  useEffect(() => {
    if (constantEntries.length === 0) {
      return;
    }

    let nextValue = value;
    let changed = false;
    constantEntries.forEach((entry) => {
      const currentAtPointer = getValueAtPointer(nextValue, entry.pointer);
      if (jsonEquals(currentAtPointer, entry.value)) {
        return;
      }
      nextValue = setValueAtPointer(nextValue, entry.pointer, entry.value);
      changed = true;
    });

    if (changed) {
      onChange(nextValue);
    }
  }, [constantEntries, onChange, value]);

  const renderField = (
    key: string,
    schemaNode: Record<string, unknown>,
    pointer: string,
    depth: number,
  ) => {
    const currentValue = getValueAtPointer(value, pointer);
    const ui = (uiByPointer[pointer] || {}) as MetadataPackUiOverride;
    const fieldType = normalizeType(schemaNode);
    const title =
      ui.title || (typeof schemaNode.title === "string" ? schemaNode.title : key);
    const help = ui.help || (typeof schemaNode.description === "string" ? schemaNode.description : "");
    const isConstant = Boolean(ui.constant);
    const constantValue = isJsonValue(ui.constantValue) ? (ui.constantValue as JsonValue) : undefined;
    const effectiveCurrentValue = isConstant ? constantValue : currentValue;

    const commonLabel = (
      <div className="flex items-center gap-2">
        <label className={`${STUDIO_DS.labelMono} mb-1`}>{title}</label>
        {isConstant && (
          <span className="text-[9px] px-1.5 py-0.5 rounded border border-slate-700 text-slate-300 bg-slate-900">
            {t("metadataPackForm.constant")}
          </span>
        )}
      </div>
    );

    const indentClass = depth > 0 ? "ml-3 pl-3 border-l border-slate-800" : "";
    const wrapperClass = `${indentClass} space-y-2`;

    if (isConstant) {
      const renderConstantControl = () => {
        if (typeof constantValue === "boolean") {
          return (
            <label className="flex items-center gap-2 text-xs text-slate-300">
              <input type="checkbox" checked={constantValue} disabled />
              {constantValue ? "true" : "false"}
            </label>
          );
        }

        if (
          Array.isArray(constantValue) ||
          (constantValue && typeof constantValue === "object")
        ) {
          return (
            <textarea
              readOnly
              value={JSON.stringify(constantValue, null, 2)}
              className={`${STUDIO_DS.textarea} h-24 font-mono text-[11px] text-slate-200`}
            />
          );
        }

        return (
          <input
            readOnly
            value={toTextValue(effectiveCurrentValue)}
            className={STUDIO_DS.input}
          />
        );
      };

      return (
        <div key={pointer} className={wrapperClass}>
          {commonLabel}
          {renderConstantControl()}
          {help && <p className="text-[10px] text-slate-500">{help}</p>}
        </div>
      );
    }

    if (fieldType === "object" && ui.widget !== "json") {
      const nested = schemaProperties(schemaNode);
      return (
        <div key={pointer} className={wrapperClass}>
          <div className="text-[10px] uppercase tracking-wider text-slate-300 font-semibold">
            {title}
          </div>
          {help && <p className="text-[10px] text-slate-500">{help}</p>}
          <div className="space-y-3">
            {nested.map((entry) =>
              renderField(
                entry.key,
                entry.schema,
                pointerFor(pointer, entry.key),
                depth + 1,
              ),
            )}
          </div>
        </div>
      );
    }

    if (fieldType === "array" || ui.widget === "array") {
      const itemSchema = isRecord(schemaNode.items) ? schemaNode.items : {};
      const itemType = normalizeType(itemSchema);
      const currentArray = Array.isArray(effectiveCurrentValue) ? [...effectiveCurrentValue] : [];

      const updateArray = (nextArray: JsonValue[]) => {
        const next = setValueAtPointer(value, pointer, nextArray);
        onChange(next);
      };

      return (
        <div key={pointer} className={wrapperClass}>
          {commonLabel}
          {help && <p className="text-[10px] text-slate-500 -mt-1">{help}</p>}
          <div className="space-y-2">
            {currentArray.map((item, index) => (
              <div key={`${pointer}-${index}`} className="flex items-center gap-2">
                {itemType === "number" || itemType === "integer" ? (
                  <input
                    type="number"
                    value={item === null || item === undefined ? "" : String(item)}
                    onChange={(event) => {
                      const raw = event.target.value;
                      const parsed = raw === "" ? null : Number(raw);
                      const nextArray = [...currentArray];
                      nextArray[index] = Number.isFinite(parsed) ? (parsed as JsonValue) : null;
                      updateArray(nextArray);
                    }}
                    className={STUDIO_DS.input}
                  />
                ) : itemType === "boolean" ? (
                  <label className="flex items-center gap-2 text-xs text-slate-300">
                    <input
                      type="checkbox"
                      checked={Boolean(item)}
                      onChange={(event) => {
                        const nextArray = [...currentArray];
                        nextArray[index] = event.target.checked;
                        updateArray(nextArray);
                      }}
                    />
                    {String(item)}
                  </label>
                ) : (
                  <input
                    type="text"
                    value={item === null || item === undefined ? "" : String(item)}
                    onChange={(event) => {
                      const nextArray = [...currentArray];
                      nextArray[index] = event.target.value;
                      updateArray(nextArray);
                    }}
                    className={STUDIO_DS.input}
                  />
                )}

                <button
                  type="button"
                  onClick={() => {
                    const nextArray = currentArray.filter((_, itemIndex) => itemIndex !== index);
                    updateArray(nextArray);
                  }}
                  className="px-2 py-1 text-[10px] rounded border border-red-700/70 text-red-300 hover:bg-red-900/30"
                >
                  {t("metadataPackForm.remove")}
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() => {
                let defaultValue: JsonValue = "";
                if (itemType === "number" || itemType === "integer") {
                  defaultValue = 0;
                } else if (itemType === "boolean") {
                  defaultValue = false;
                } else if (itemType === "object") {
                  defaultValue = {};
                } else if (itemType === "array") {
                  defaultValue = [];
                }
                updateArray([...currentArray, defaultValue]);
              }}
              className="px-3 py-1 text-[10px] rounded border border-blue-700/60 text-blue-300 hover:bg-blue-900/30"
            >
              {t("metadataPackForm.addItem")}
            </button>
          </div>
        </div>
      );
    }

    if (fieldType === "boolean" || ui.widget === "boolean") {
      return (
        <div key={pointer} className={wrapperClass}>
          <label className="flex items-center gap-2 text-xs text-slate-300">
            <input
              type="checkbox"
              checked={Boolean(effectiveCurrentValue)}
              onChange={(event) => onChange(setValueAtPointer(value, pointer, event.target.checked))}
            />
            {title}
          </label>
          {help && <p className="text-[10px] text-slate-500">{help}</p>}
        </div>
      );
    }

    if (ui.widget === "select" || Array.isArray(schemaNode.enum)) {
      const enumValues = Array.isArray(schemaNode.enum)
        ? schemaNode.enum.filter((candidate): candidate is string => typeof candidate === "string")
        : [];
      const uiOptions = Array.isArray(ui.options) ? ui.options : [];
      const options =
        uiOptions.length > 0
          ? uiOptions
          : enumValues.map((entry) => ({ value: entry, label: entry }));
      const current = toTextValue(effectiveCurrentValue);

      return (
        <div key={pointer} className={wrapperClass}>
          {commonLabel}
          <select
            value={current}
            onChange={(event) => onChange(setValueAtPointer(value, pointer, event.target.value))}
            className={`${STUDIO_DS.input} cursor-pointer`}
          >
            <option value="">{t("metadataPackForm.selectOption")}</option>
            {options.map((option) => (
              <option key={`${pointer}-${option.value}`} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
          {ui.allowCustomValue && (
            <input
              type="text"
              placeholder={ui.placeholder || t("metadataPackForm.customValue")}
              value={current}
              onChange={(event) => onChange(setValueAtPointer(value, pointer, event.target.value))}
              className={STUDIO_DS.input}
            />
          )}
          {help && <p className="text-[10px] text-slate-500">{help}</p>}
        </div>
      );
    }

    if (ui.widget === "json") {
      const textValue =
        effectiveCurrentValue !== undefined ? JSON.stringify(effectiveCurrentValue, null, 2) : "";
      const parseError = jsonErrors[pointer];

      return (
        <div key={pointer} className={wrapperClass}>
          {commonLabel}
          <textarea
            value={textValue}
            onChange={(event) => {
              const raw = event.target.value;
              if (!raw.trim()) {
                const { [pointer]: _removed, ...rest } = jsonErrors;
                setJsonErrors(rest);
                onChange(setValueAtPointer(value, pointer, {}));
                return;
              }
              try {
                const parsed = JSON.parse(raw) as JsonValue;
                const { [pointer]: _removed, ...rest } = jsonErrors;
                setJsonErrors(rest);
                onChange(setValueAtPointer(value, pointer, parsed));
              } catch {
                setJsonErrors((previous) => ({
                  ...previous,
                  [pointer]: t("metadataPackForm.invalidJson"),
                }));
              }
            }}
            className={`${STUDIO_DS.textarea} h-28 font-mono text-[11px] text-yellow-300`}
          />
          {parseError && <p className="text-[10px] text-red-400">{parseError}</p>}
          {help && <p className="text-[10px] text-slate-500">{help}</p>}
        </div>
      );
    }

    const inputType = fieldType === "number" || fieldType === "integer" || ui.widget === "number"
      ? "number"
      : "text";

    return (
      <div key={pointer} className={wrapperClass}>
        {commonLabel}
        {(ui.widget === "textarea" || fieldType === "string" && (schemaNode.maxLength as number | undefined) && (schemaNode.maxLength as number) > 120) ? (
          <textarea
            value={toTextValue(effectiveCurrentValue)}
            onChange={(event) => onChange(setValueAtPointer(value, pointer, event.target.value))}
            placeholder={ui.placeholder}
            className={`${STUDIO_DS.textarea} h-20`}
          />
        ) : (
          <input
            type={inputType}
            value={toTextValue(effectiveCurrentValue)}
            onChange={(event) => {
              if (inputType === "number") {
                const raw = event.target.value;
                const parsed = raw === "" ? null : Number(raw);
                onChange(
                  setValueAtPointer(
                    value,
                    pointer,
                    Number.isFinite(parsed) ? (parsed as JsonValue) : null,
                  ),
                );
                return;
              }
              onChange(setValueAtPointer(value, pointer, event.target.value));
            }}
            placeholder={ui.placeholder}
            className={STUDIO_DS.input}
          />
        )}
        {help && <p className="text-[10px] text-slate-500">{help}</p>}
      </div>
    );
  };

  const rootSchema = isRecord(pack.schema) ? pack.schema : {};
  const rootValue = ensureObject(value as JsonValue);
  const rootFields = schemaProperties(rootSchema);

  return (
    <div className="space-y-4">
      {rootFields.length === 0 ? (
        <p className="text-xs text-slate-400">
          {t("metadataPackForm.emptySchema")}
        </p>
      ) : (
        rootFields.map((entry) =>
          renderField(entry.key, entry.schema, pointerFor("", entry.key), 0),
        )
      )}

      {rootFields.length === 0 && (
        <textarea
          value={JSON.stringify(rootValue, null, 2)}
          onChange={(event) => {
            try {
              const parsed = JSON.parse(event.target.value) as JsonObject;
              onChange(parsed);
            } catch {
              // Keep last valid value when JSON is invalid.
            }
          }}
          className={`${STUDIO_DS.textarea} h-32 font-mono text-[11px] text-yellow-300`}
        />
      )}
    </div>
  );
};
