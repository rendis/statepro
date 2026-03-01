import {
  AlertTriangle,
  BookOpen,
  Boxes,
  Eye,
  Plus,
  Save,
  Search,
  Terminal,
  Trash2,
  X,
} from "lucide-react";
import { useMemo, useState } from "react";

import {
  BEHAVIOR_TYPE_LABEL_KEYS,
  BEHAVIOR_TYPES,
  STUDIO_ICON_REGISTRY,
  STUDIO_ICONS,
} from "../../constants";
import { TooltipIconButton, TruncatedWithTooltip } from "../../components/shared";
import { useI18n } from "../../i18n";
import type { StudioTranslate } from "../../i18n";
import type {
  BehaviorRegistryItem,
  BehaviorType,
  JsonObject,
  JsonValue,
  MetadataPackDefinition,
  MetadataPackRegistry,
  MetadataPackSelectOption,
  MetadataScope,
} from "../../types";
import {
  cleanIdentifier,
  getBehaviorScriptContract,
  findFieldPathCollisionsDetailed,
  invertPointerCollisionRelation,
} from "../../utils";
import type { UsageSummary } from "../../utils";

interface LibraryModalProps {
  isOpen: boolean;
  onClose: () => void;
  canManageBehaviors?: boolean;
  canCreateMetadataPacks?: boolean;
  registry: BehaviorRegistryItem[];
  setRegistry: (registry: BehaviorRegistryItem[]) => void;
  resolveUsage: (src: string) => UsageSummary;
  onDeleteBehavior: (src: string) => void;
  metadataPackRegistry?: MetadataPackRegistry;
  setMetadataPackRegistry?: (registry: MetadataPackRegistry) => void;
  onDeleteMetadataPack?: (packId: string) => void;
  onRenameMetadataPack?: (previousPackId: string, nextPackId: string) => void;
}

interface BehaviorLibraryForm {
  src: string;
  type: BehaviorType;
  description: string;
  simScript: string;
  simScriptCustomized: boolean;
  isNew: boolean;
  originalSrc?: string;
}

interface MetadataPackLibraryForm {
  label: string;
  description: string;
  fields: VisualPackField[];
  scopes: Record<MetadataScope, boolean>;
  isNew: boolean;
  originalId?: string;
}

type VisualFieldKind =
  | "text"
  | "textarea"
  | "number"
  | "boolean"
  | "select"
  | "array"
  | "json";

type VisualArrayItemType = "text" | "number" | "boolean" | "json";

interface VisualPackField {
  id: string;
  path: string;
  label: string;
  kind: VisualFieldKind;
  required: boolean;
  placeholder: string;
  help: string;
  allowCustomValue: boolean;
  options: MetadataPackSelectOption[];
  arrayItemType: VisualArrayItemType;
  isConstant: boolean;
  constantValueText: string;
  constantValueBoolean: boolean;
}

const emptyScopes = (): Record<MetadataScope, boolean> => ({
  machine: true,
  universe: false,
  reality: false,
  transition: false,
});

const METADATA_SCOPE_ORDER: MetadataScope[] = [
  "machine",
  "universe",
  "reality",
  "transition",
];

const resolveScopeVisual = (scope: MetadataScope) => {
  if (scope === "machine") {
    return {
      icon: STUDIO_ICONS.entity.machine,
      colorClass: STUDIO_ICON_REGISTRY.entity.machine.colors.base,
    };
  }
  if (scope === "universe") {
    return {
      icon: STUDIO_ICONS.entity.universe,
      colorClass: STUDIO_ICON_REGISTRY.entity.universe.colors.base,
    };
  }
  if (scope === "reality") {
    return {
      icon: STUDIO_ICONS.entity.reality,
      colorClass: STUDIO_ICON_REGISTRY.entity.reality.colors.base,
    };
  }
  return {
    icon: STUDIO_ICONS.transition.type.default,
    colorClass:
      STUDIO_ICON_REGISTRY.transition.type.default.colors.accent ||
      STUDIO_ICON_REGISTRY.transition.type.default.colors.base,
  };
};

const FIELD_KINDS: Array<{ id: VisualFieldKind; label: string }> = [
  { id: "text", label: "text" },
  { id: "textarea", label: "textarea" },
  { id: "number", label: "number" },
  { id: "boolean", label: "boolean" },
  { id: "select", label: "select" },
  { id: "array", label: "array" },
  { id: "json", label: "json" },
];

const ARRAY_ITEM_TYPES: Array<{ id: VisualArrayItemType; label: string }> = [
  { id: "text", label: "text" },
  { id: "number", label: "number" },
  { id: "boolean", label: "boolean" },
  { id: "json", label: "json" },
];

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const createVisualField = (partial?: Partial<VisualPackField>): VisualPackField => ({
  id: partial?.id || `field-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  path: partial?.path || "",
  label: partial?.label || "",
  kind: partial?.kind || "text",
  required: partial?.required || false,
  placeholder: partial?.placeholder || "",
  help: partial?.help || "",
  allowCustomValue: partial?.allowCustomValue || false,
  options: partial?.options || [],
  arrayItemType: partial?.arrayItemType || "text",
  isConstant: partial?.isConstant || false,
  constantValueText: partial?.constantValueText || "",
  constantValueBoolean: partial?.constantValueBoolean || false,
});

const pathToSegments = (path: string): string[] =>
  path
    .split(".")
    .map((segment) => segment.trim())
    .filter(Boolean);

const normalizeFieldPath = (path: string): string => pathToSegments(path).join(".");

const normalizePackLabel = (label: string): string =>
  label
    .trim()
    .replace(/\s+/g, " ")
    .toLowerCase();

const translatePointerCollisionRelation = (
  t: StudioTranslate,
  relation: "exact" | "ancestor" | "descendant",
): string => t(`library.pack.relation.${relation}`, undefined, relation);

const translateFieldKind = (
  t: StudioTranslate,
  kind: VisualFieldKind | VisualArrayItemType,
): string => t(`library.pack.kind.${kind}`, undefined, kind);

interface ParsedFieldConstantValue {
  ok: boolean;
  value?: JsonValue;
  error?: string;
}

const formatConstantValueForField = (
  kind: VisualFieldKind,
  value: unknown,
): { constantValueText: string; constantValueBoolean: boolean } => {
  if (kind === "boolean") {
    return {
      constantValueText: "",
      constantValueBoolean: Boolean(value),
    };
  }

  if (kind === "array") {
    return {
      constantValueText: Array.isArray(value) ? JSON.stringify(value, null, 2) : "[]",
      constantValueBoolean: false,
    };
  }

  if (kind === "json") {
    return {
      constantValueText:
        value && typeof value === "object" && !Array.isArray(value)
          ? JSON.stringify(value, null, 2)
          : "{}",
      constantValueBoolean: false,
    };
  }

  if (kind === "number") {
    return {
      constantValueText:
        typeof value === "number" && Number.isFinite(value) ? String(value) : "",
      constantValueBoolean: false,
    };
  }

  return {
    constantValueText:
      typeof value === "string"
        ? value
        : value === undefined || value === null
          ? ""
          : String(value),
    constantValueBoolean: false,
  };
};

const parseVisualFieldConstantValue = (
  field: VisualPackField,
  t: StudioTranslate,
): ParsedFieldConstantValue => {
  if (!field.isConstant) {
    return { ok: true };
  }

  if (field.kind === "select") {
    return {
      ok: false,
      error: t("library.pack.validation.selectConstantUnsupported"),
    };
  }

  if (field.kind === "boolean") {
    return { ok: true, value: field.constantValueBoolean };
  }

  if (field.kind === "number") {
    const raw = field.constantValueText.trim();
    if (!raw) {
      return {
        ok: false,
        error: t("library.pack.validation.constantNumberRequired"),
      };
    }
    const parsed = Number(raw);
    if (!Number.isFinite(parsed)) {
      return {
        ok: false,
        error: t("library.pack.validation.constantNumberInvalid"),
      };
    }
    return { ok: true, value: parsed };
  }

  if (field.kind === "array") {
    const raw = field.constantValueText.trim();
    if (!raw) {
      return {
        ok: false,
        error: t("library.pack.validation.constantArrayInvalid"),
      };
    }
    try {
      const parsed = JSON.parse(raw) as JsonValue;
      if (!Array.isArray(parsed)) {
        return {
          ok: false,
          error: t("library.pack.validation.constantArrayMustBeArray"),
        };
      }
      return { ok: true, value: parsed };
    } catch {
      return {
        ok: false,
        error: t("library.pack.validation.constantArrayInvalid"),
      };
    }
  }

  if (field.kind === "json") {
    const raw = field.constantValueText.trim();
    if (!raw) {
      return {
        ok: false,
        error: t("library.pack.validation.constantJsonRequired"),
      };
    }
    try {
      const parsed = JSON.parse(raw) as JsonValue;
      if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
        return {
          ok: false,
          error: t("library.pack.validation.constantJsonMustBeObject"),
        };
      }
      return { ok: true, value: parsed };
    } catch {
      return {
        ok: false,
        error: t("library.pack.validation.constantJsonInvalid"),
      };
    }
  }

  return { ok: true, value: field.constantValueText };
};

interface VisualFieldValidationResult {
  fieldErrors: Map<string, string[]>;
  hasBlockingErrors: boolean;
}

const pushFieldError = (
  fieldErrors: Map<string, string[]>,
  fieldId: string,
  message: string,
): void => {
  const current = fieldErrors.get(fieldId) || [];
  if (current.includes(message)) {
    return;
  }
  fieldErrors.set(fieldId, [...current, message]);
};

const buildVisualFieldValidation = (
  fields: VisualPackField[],
  t: StudioTranslate,
): VisualFieldValidationResult => {
  const fieldErrors = new Map<string, string[]>();
  const normalizedPaths = fields.map((field) => normalizeFieldPath(field.path.trim()));

  fields.forEach((field, index) => {
    const rawPath = field.path.trim();
    if (!rawPath) {
      const message = t("library.pack.validation.pathRequired", { index: index + 1 });
      pushFieldError(fieldErrors, field.id, message);
      return;
    }

    if (!normalizedPaths[index]) {
      const message = t("library.pack.validation.pathInvalid", { path: rawPath });
      pushFieldError(fieldErrors, field.id, message);
    }

    const constantValidation = parseVisualFieldConstantValue(field, t);
    if (!constantValidation.ok) {
      const pathValue = rawPath || `#${index + 1}`;
      const message = t("library.pack.validation.constantInvalid", {
        path: pathValue,
        error:
          constantValidation.error ||
          t("library.pack.validation.constantInvalidShort"),
      });
      pushFieldError(fieldErrors, field.id, message);
    }
  });

  const duplicatePaths = new Map<string, number[]>();
  normalizedPaths.forEach((normalizedPath, index) => {
    if (!normalizedPath) {
      return;
    }
    const current = duplicatePaths.get(normalizedPath) || [];
    current.push(index);
    duplicatePaths.set(normalizedPath, current);
  });

  duplicatePaths.forEach((indexes, normalizedPath) => {
    if (indexes.length <= 1) {
      return;
    }
    const message = t("library.pack.validation.duplicatePath", {
      path: normalizedPath,
    });
    indexes.forEach((entryIndex) => {
      const field = fields[entryIndex];
      if (!field) {
        return;
      }
      pushFieldError(fieldErrors, field.id, message);
    });
  });

  findFieldPathCollisionsDetailed(normalizedPaths)
    .filter((collision) => collision.relation !== "exact")
    .forEach((collision) => {
      const leftField = fields[collision.leftIndex];
      const rightField = fields[collision.rightIndex];
      if (!leftField || !rightField) {
        return;
      }

      const leftRelation = translatePointerCollisionRelation(t, collision.relation);
      const rightRelation = translatePointerCollisionRelation(
        t,
        invertPointerCollisionRelation(collision.relation),
      );

      pushFieldError(
        fieldErrors,
        leftField.id,
        t("library.pack.validation.pathCollision", {
          leftPath: collision.leftPath,
          rightPath: collision.rightPath,
          relation: leftRelation,
        }),
      );
      pushFieldError(
        fieldErrors,
        rightField.id,
        t("library.pack.validation.pathCollision", {
          leftPath: collision.rightPath,
          rightPath: collision.leftPath,
          relation: rightRelation,
        }),
      );
    });

  return {
    fieldErrors,
    hasBlockingErrors: fieldErrors.size > 0,
  };
};

const pointerFromSegments = (segments: string[]): string =>
  `/${segments.map((segment) => segment.replace(/~/g, "~0").replace(/\//g, "~1")).join("/")}`;

const normalizeRequired = (node: Record<string, unknown>): string[] => {
  const raw = node.required;
  if (!Array.isArray(raw)) {
    return [];
  }
  return raw.filter((entry): entry is string => typeof entry === "string");
};

const resolveArrayItemSchema = (
  itemType: VisualArrayItemType,
): Record<string, unknown> => {
  if (itemType === "number") {
    return { type: "number" };
  }
  if (itemType === "boolean") {
    return { type: "boolean" };
  }
  if (itemType === "json") {
    return { type: "object", additionalProperties: true };
  }
  return { type: "string" };
};

const kindFromSchema = (
  schemaNode: Record<string, unknown>,
  widget: unknown,
): VisualFieldKind => {
  if (typeof widget === "string") {
    if (
      widget === "text" ||
      widget === "textarea" ||
      widget === "number" ||
      widget === "boolean" ||
      widget === "select" ||
      widget === "array" ||
      widget === "json"
    ) {
      return widget;
    }
  }

  const type = schemaNode.type;
  if (type === "array") {
    return "array";
  }
  if (type === "boolean") {
    return "boolean";
  }
  if (type === "number" || type === "integer") {
    return "number";
  }
  if (Array.isArray(schemaNode.enum)) {
    return "select";
  }
  if (type === "object") {
    return "json";
  }
  return "text";
};

const ensureObjectNode = (
  root: Record<string, unknown>,
  segments: string[],
): Record<string, unknown> => {
  let cursor = root;
  segments.forEach((segment) => {
    if (!isRecord(cursor.properties)) {
      cursor.properties = {};
    }

    const properties = cursor.properties as Record<string, unknown>;
    if (!isRecord(properties[segment])) {
      properties[segment] = {
        type: "object",
        properties: {},
        additionalProperties: false,
      };
    }

    const currentNode = properties[segment];
    if (!isRecord(currentNode)) {
      properties[segment] = {
        type: "object",
        properties: {},
        additionalProperties: false,
      };
    }

    cursor = properties[segment] as Record<string, unknown>;
    if (cursor.type !== "object") {
      cursor.type = "object";
    }
    if (!isRecord(cursor.properties)) {
      cursor.properties = {};
    }
    if (!("additionalProperties" in cursor)) {
      cursor.additionalProperties = false;
    }
  });
  return cursor;
};

const buildSchemaAndUiFromFields = (
  fields: VisualPackField[],
  t: StudioTranslate,
): { schema: JsonObject; ui: JsonObject; errors: string[] } => {
  const schema: Record<string, unknown> = {
    type: "object",
    properties: {},
    additionalProperties: false,
  };
  const ui: Record<string, unknown> = {};
  const errors: string[] = [];
  const errorSet = new Set<string>();
  const pushError = (message: string): void => {
    if (errorSet.has(message)) {
      return;
    }
    errors.push(message);
    errorSet.add(message);
  };
  const seenPaths = new Set<string>();
  const normalizedPaths = fields.map((field) => normalizeFieldPath(field.path.trim()));

  findFieldPathCollisionsDetailed(normalizedPaths)
    .filter((collision) => collision.relation !== "exact")
    .forEach((collision) => {
      const relation = translatePointerCollisionRelation(t, collision.relation);
      const inverseRelation = translatePointerCollisionRelation(
        t,
        invertPointerCollisionRelation(collision.relation),
      );
      pushError(
        t("library.pack.validation.hierarchicalCollision", {
          leftPath: collision.leftPath,
          leftRelation: relation,
          rightPath: collision.rightPath,
          rightRelation: inverseRelation,
        }),
      );
    });

  fields.forEach((field, index) => {
    const rawPath = field.path.trim();
    if (!rawPath) {
      pushError(t("library.pack.validation.pathRequired", { index: index + 1 }));
      return;
    }

    const normalizedPath = normalizeFieldPath(rawPath);
    if (!normalizedPath) {
      pushError(t("library.pack.validation.pathInvalid", { path: rawPath }));
      return;
    }

    if (seenPaths.has(normalizedPath)) {
      pushError(t("library.pack.validation.duplicatePath", { path: normalizedPath }));
      return;
    }
    seenPaths.add(normalizedPath);

    const segments = normalizedPath.split(".");
    const parentSegments = segments.slice(0, -1);
    const leaf = segments[segments.length - 1];
    if (!leaf) {
      pushError(t("library.pack.validation.pathInvalid", { path: rawPath }));
      return;
    }

    const parentNode = ensureObjectNode(schema, parentSegments);
    const properties = parentNode.properties as Record<string, unknown>;
    const schemaNode: Record<string, unknown> = {};

    if (field.kind === "number") {
      schemaNode.type = "number";
    } else if (field.kind === "boolean") {
      schemaNode.type = "boolean";
    } else if (field.kind === "json") {
      schemaNode.type = "object";
      schemaNode.additionalProperties = true;
    } else if (field.kind === "array") {
      schemaNode.type = "array";
      schemaNode.items = resolveArrayItemSchema(field.arrayItemType);
    } else {
      schemaNode.type = "string";
      if (field.kind === "select") {
        const optionValues = field.options
          .map((option) => option.value.trim())
          .filter(Boolean);
        if (optionValues.length > 0) {
          schemaNode.enum = optionValues;
        }
      }
    }

    const normalizedLabel = field.label.trim();
    schemaNode.title = normalizedLabel || leaf;
    if (field.help.trim()) {
      schemaNode.description = field.help.trim();
    }

    properties[leaf] = schemaNode;

    if (field.required) {
      const required = new Set(normalizeRequired(parentNode));
      required.add(leaf);
      parentNode.required = Array.from(required);
    }

    const pointer = pointerFromSegments(segments);
    const uiNode: Record<string, unknown> = {};
    if (field.kind !== "text") {
      uiNode.widget = field.kind;
    }
    if (field.placeholder.trim()) {
      uiNode.placeholder = field.placeholder.trim();
    }
    if (field.help.trim()) {
      uiNode.help = field.help.trim();
    }
    uiNode.title = normalizedLabel || leaf;
    if (field.kind === "select") {
      if (field.options.length > 0) {
        uiNode.options = field.options
          .filter((option) => option.value.trim())
          .map((option) => ({
            value: option.value.trim(),
            label: option.label.trim() || option.value.trim(),
          }));
      }
      if (field.allowCustomValue) {
        uiNode.allowCustomValue = true;
      }
    }

    const constantValidation = parseVisualFieldConstantValue(field, t);
    if (!constantValidation.ok) {
      pushError(
        t("library.pack.validation.constantInvalid", {
          path: normalizedPath,
          error:
            constantValidation.error ||
            t("library.pack.validation.constantInvalidShort"),
        }),
      );
    } else if (field.isConstant) {
      uiNode.constant = true;
      uiNode.constantValue = constantValidation.value;
    }

    if (Object.keys(uiNode).length > 0) {
      ui[pointer] = uiNode;
    }
  });

  return {
    schema: schema as JsonObject,
    ui: ui as JsonObject,
    errors,
  };
};

const readFieldFromSchema = (
  key: string,
  schemaNode: Record<string, unknown>,
  required: Set<string>,
  path: string[],
  uiMap: Record<string, unknown>,
  fields: VisualPackField[],
): void => {
  const pointer = pointerFromSegments(path);
  const uiNode = isRecord(uiMap[pointer]) ? (uiMap[pointer] as Record<string, unknown>) : {};
  const properties = isRecord(schemaNode.properties)
    ? (schemaNode.properties as Record<string, unknown>)
    : null;

  if (schemaNode.type === "object" && properties && Object.keys(properties).length > 0) {
    Object.entries(properties).forEach(([childKey, childValue]) => {
      if (!isRecord(childValue)) {
        return;
      }
      const nextRequired = new Set(
        normalizeRequired(schemaNode),
      );
      readFieldFromSchema(
        childKey,
        childValue,
        nextRequired,
        [...path, childKey],
        uiMap,
        fields,
      );
    });
    return;
  }

  const kind = kindFromSchema(schemaNode, uiNode.widget);
  const arrayItems = isRecord(schemaNode.items) ? (schemaNode.items as Record<string, unknown>) : null;
  const arrayItemType: VisualArrayItemType =
    kind === "array"
      ? arrayItems?.type === "number"
        ? "number"
        : arrayItems?.type === "boolean"
          ? "boolean"
          : arrayItems?.type === "object"
            ? "json"
            : "text"
      : "text";

  const optionsFromUi = Array.isArray(uiNode.options)
    ? (uiNode.options as unknown[])
        .filter((entry) => isRecord(entry) && typeof entry.value === "string")
        .map((entry) => ({
          value: String((entry as Record<string, unknown>).value),
          label: String(
            (entry as Record<string, unknown>).label ||
              (entry as Record<string, unknown>).value,
          ),
        }))
    : [];

  const optionsFromEnum = Array.isArray(schemaNode.enum)
    ? (schemaNode.enum as unknown[])
        .filter((entry): entry is string => typeof entry === "string")
        .map((entry) => ({ value: entry, label: entry }))
    : [];

  const constantRawValue = uiNode.constantValue;
  const constantVisualState = formatConstantValueForField(kind, constantRawValue);

  fields.push(
    createVisualField({
      path: path.join("."),
      label:
        (typeof uiNode.title === "string" && uiNode.title) ||
        (typeof schemaNode.title === "string" && schemaNode.title) ||
        key,
      kind,
      required: required.has(key),
      placeholder:
        typeof uiNode.placeholder === "string"
          ? uiNode.placeholder
          : "",
      help:
        (typeof uiNode.help === "string" && uiNode.help) ||
        (typeof schemaNode.description === "string" ? schemaNode.description : ""),
      allowCustomValue: Boolean(uiNode.allowCustomValue),
      options: optionsFromUi.length > 0 ? optionsFromUi : optionsFromEnum,
      arrayItemType,
      isConstant: Boolean(uiNode.constant),
      constantValueText: constantVisualState.constantValueText,
      constantValueBoolean: constantVisualState.constantValueBoolean,
    }),
  );
};

const visualFieldsFromPack = (
  pack: MetadataPackDefinition,
  t: StudioTranslate,
): VisualPackField[] => {
  const rootSchema = isRecord(pack.schema) ? (pack.schema as Record<string, unknown>) : {};
  const rootProperties = isRecord(rootSchema.properties)
    ? (rootSchema.properties as Record<string, unknown>)
    : {};
  const rootRequired = new Set(normalizeRequired(rootSchema));
  const uiMap = isRecord(pack.ui) ? (pack.ui as Record<string, unknown>) : {};
  const fields: VisualPackField[] = [];

  Object.entries(rootProperties).forEach(([key, value]) => {
    if (!isRecord(value)) {
      return;
    }
    readFieldFromSchema(key, value, rootRequired, [key], uiMap, fields);
  });

  if (fields.length === 0) {
    return [
      createVisualField({
        path: "field",
        label: t("library.pack.defaultFieldLabel"),
        kind: "text",
      }),
    ];
  }
  return fields;
};

export const LibraryModal = ({
  isOpen,
  onClose,
  canManageBehaviors = true,
  canCreateMetadataPacks = true,
  registry,
  setRegistry,
  resolveUsage,
  onDeleteBehavior,
  metadataPackRegistry = [],
  setMetadataPackRegistry = () => undefined,
  onDeleteMetadataPack,
  onRenameMetadataPack,
}: LibraryModalProps) => {
  const { t } = useI18n();
  const [activeTab, setActiveTab] = useState<"behaviors" | "metadata-packs">("behaviors");

  const [form, setForm] = useState<BehaviorLibraryForm | null>(null);
  const [error, setError] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [pendingDeleteSrc, setPendingDeleteSrc] = useState<string | null>(null);
  const [activeFilters, setActiveFilters] = useState<Record<BehaviorType, boolean>>({
    action: true,
    invoke: true,
    condition: true,
    observer: true,
  });

  const [packForm, setPackForm] = useState<MetadataPackLibraryForm | null>(null);
  const [packError, setPackError] = useState("");
  const [packSearchQuery, setPackSearchQuery] = useState("");
  const [pendingDeletePackId, setPendingDeletePackId] = useState<string | null>(null);
  const [showPackGeneratedPreview, setShowPackGeneratedPreview] = useState(false);
  const packFieldValidation = useMemo(
    () => buildVisualFieldValidation(packForm?.fields || [], t),
    [packForm?.fields, t],
  );
  const packGeneratedPreview = useMemo(
    () => buildSchemaAndUiFromFields(packForm?.fields || [], t),
    [packForm?.fields, t],
  );
  const normalizedPackId = useMemo(
    () => cleanIdentifier(packForm?.label || ""),
    [packForm?.label],
  );
  const conflictingPackByLabel = useMemo(() => {
    if (!packForm) {
      return null;
    }
    const currentLabel = normalizePackLabel(packForm.label);
    if (!currentLabel) {
      return null;
    }
    return (
      metadataPackRegistry.find((pack) => {
        if (!packForm.isNew && pack.id === packForm.originalId) {
          return false;
        }
        return normalizePackLabel(pack.label) === currentLabel;
      }) || null
    );
  }, [metadataPackRegistry, packForm]);
  const conflictingPackById = useMemo(() => {
    if (!packForm || !normalizedPackId) {
      return null;
    }
    return (
      metadataPackRegistry.find((pack) => {
        if (!packForm.isNew && pack.id === packForm.originalId) {
          return false;
        }
        return pack.id === normalizedPackId;
      }) || null
    );
  }, [metadataPackRegistry, normalizedPackId, packForm]);
  const packLabelError = useMemo(() => {
    if (!packForm) {
      return "";
    }

    if (!packForm.label.trim()) {
      return t("library.pack.validation.labelRequired");
    }
    if (!normalizedPackId) {
      return t("library.pack.validation.invalidGeneratedId");
    }
    if (conflictingPackByLabel) {
      return t("library.pack.validation.duplicateLabel", {
        label: conflictingPackByLabel.label,
      });
    }
    if (conflictingPackById) {
      return t("library.pack.validation.duplicateId", { id: normalizedPackId });
    }
    return "";
  }, [conflictingPackById, conflictingPackByLabel, normalizedPackId, packForm, t]);
  const packCanSave = useMemo(() => {
    if (!packForm) {
      return false;
    }
    const hasEnabledScope = Object.values(packForm.scopes).some(Boolean);
    return (
      !packLabelError &&
      hasEnabledScope &&
      !packFieldValidation.hasBlockingErrors &&
      packGeneratedPreview.errors.length === 0
    );
  }, [
    packFieldValidation.hasBlockingErrors,
    packForm,
    packGeneratedPreview.errors.length,
    packLabelError,
  ]);
  const activeBehaviorScriptContract = useMemo(() => {
    if (!form) {
      return null;
    }
    return getBehaviorScriptContract(form.type, t);
  }, [form, t]);

  const handleCreateNewBehavior = () => {
    if (!canManageBehaviors) {
      return;
    }

    const actionContract = getBehaviorScriptContract("action", t);
    setPendingDeleteSrc(null);
    setForm({
      src: "",
      type: "action",
      description: "",
      simScript: actionContract.defaultScript,
      simScriptCustomized: false,
      isNew: true,
    });
    setError("");
  };

  const handleEditBehavior = (item: BehaviorRegistryItem) => {
    if (!canManageBehaviors) {
      return;
    }

    setPendingDeleteSrc(null);
    setForm({
      ...item,
      description: item.description || "",
      simScriptCustomized: true,
      originalSrc: item.src,
      isNew: false,
    });
    setError("");
  };

  const handleBehaviorTypeChange = (type: BehaviorType) => {
    if (!canManageBehaviors) {
      return;
    }

    setForm((previous) => {
      if (!previous || previous.type === type) {
        return previous;
      }

      const nextContract = getBehaviorScriptContract(type, t);
      const shouldSyncTemplate = previous.isNew && !previous.simScriptCustomized;

      return {
        ...previous,
        type,
        simScript: shouldSyncTemplate ? nextContract.defaultScript : previous.simScript,
      };
    });
  };

  const handleBehaviorScriptChange = (simScript: string) => {
    if (!canManageBehaviors) {
      return;
    }

    setForm((previous) => {
      if (!previous) {
        return previous;
      }
      return {
        ...previous,
        simScript,
        simScriptCustomized: true,
      };
    });
  };

  const handleSaveBehavior = () => {
    if (!canManageBehaviors) {
      return;
    }

    if (!form) return;

    if (!form.src.trim()) {
      setError(t("library.behavior.validation.sourceRequired"));
      return;
    }

    if (form.isNew) {
      if (registry.some((entry) => entry.src === form.src)) {
        setError(t("library.behavior.validation.duplicateSource"));
        return;
      }

      setRegistry([
        ...registry,
        {
          src: form.src.trim(),
          type: form.type,
          description: form.description,
          simScript: form.simScript,
        },
      ]);
    } else {
      if (form.src !== form.originalSrc && registry.some((entry) => entry.src === form.src)) {
        setError(t("library.behavior.validation.duplicateNewSource"));
        return;
      }

      setRegistry(
        registry.map((entry) =>
          entry.src === form.originalSrc
            ? {
                src: form.src.trim(),
                type: form.type,
                description: form.description,
                simScript: form.simScript,
              }
            : entry,
        ),
      );
    }

    setForm(null);
  };

  const handleCreateNewPack = () => {
    if (!canCreateMetadataPacks) {
      return;
    }

    setPendingDeletePackId(null);
    setShowPackGeneratedPreview(false);
    setPackForm({
      label: "",
      description: "",
      fields: [
        createVisualField({
          path: "category",
          label: t("library.pack.sample.categoryLabel"),
          kind: "select",
          options: [
            { value: "a", label: t("library.pack.sample.optionA") },
            { value: "b", label: t("library.pack.sample.optionB") },
          ],
          allowCustomValue: true,
          placeholder: t("library.pack.sample.categoryPlaceholder"),
        }),
      ],
      scopes: emptyScopes(),
      isNew: true,
    });
    setPackError("");
  };

  const handleEditPack = (pack: MetadataPackDefinition) => {
    setPendingDeletePackId(null);
    setShowPackGeneratedPreview(false);
    const scopes = emptyScopes();
    pack.scopes.forEach((scope) => {
      scopes[scope] = true;
    });
    setPackForm({
      label: pack.label,
      description: pack.description || "",
      fields: visualFieldsFromPack(pack, t),
      scopes,
      isNew: false,
      originalId: pack.id,
    });
    setPackError("");
  };

  const handleSavePack = () => {
    if (!packForm) return;

    const id = normalizedPackId;
    const label = packForm.label.trim();
    const scopes = (Object.entries(packForm.scopes) as Array<[MetadataScope, boolean]>)
      .filter(([, enabled]) => enabled)
      .map(([scope]) => scope);
    const { schema, ui, errors } = packGeneratedPreview;

    if (packLabelError) {
      setPackError("");
      return;
    }
    if (packFieldValidation.hasBlockingErrors) {
      setPackError("");
      return;
    }
    if (errors.length > 0) {
      setPackError(errors[0] || t("library.pack.validation.invalidFields"));
      return;
    }
    if (scopes.length === 0) {
      setPackError(t("library.pack.validation.scopeRequired"));
      return;
    }

    const nextPack: MetadataPackDefinition = {
      id,
      label,
      description: packForm.description.trim() || undefined,
      schema,
      ui: ui as MetadataPackDefinition["ui"],
      scopes,
    };

    if (packForm.isNew) {
      setMetadataPackRegistry([...metadataPackRegistry, nextPack]);
    } else {
      setMetadataPackRegistry(
        metadataPackRegistry.map((pack) =>
          pack.id === packForm.originalId ? nextPack : pack,
        ),
      );
      if (packForm.originalId && id !== packForm.originalId) {
        onRenameMetadataPack?.(packForm.originalId, id);
      }
    }

    setPackForm(null);
    setPackError("");
  };

  const toggleFilter = (typeKey: BehaviorType) => {
    setActiveFilters((previous) => ({ ...previous, [typeKey]: !previous[typeKey] }));
  };

  const filteredRegistry = registry.filter((entry) => {
    if (!activeFilters[entry.type]) return false;
    if (!searchQuery.trim()) return true;

    const normalized = searchQuery.toLowerCase();
    return (
      entry.src.toLowerCase().includes(normalized) ||
      (entry.description && entry.description.toLowerCase().includes(normalized))
    );
  });

  const filteredPacks = useMemo(() => {
    if (!packSearchQuery.trim()) {
      return metadataPackRegistry;
    }
    const normalized = packSearchQuery.toLowerCase();
    return metadataPackRegistry.filter(
      (pack) =>
        pack.id.toLowerCase().includes(normalized) ||
        pack.label.toLowerCase().includes(normalized) ||
        (pack.description && pack.description.toLowerCase().includes(normalized)),
    );
  }, [metadataPackRegistry, packSearchQuery]);

  if (!isOpen) return null;

  const behaviorPane = (
    <div className="flex flex-1 overflow-hidden">
      <div className="w-1/3 border-r border-slate-800 bg-slate-950/50 flex flex-col">
        <div className="p-4 border-b border-slate-800 flex flex-col gap-3 shrink-0">
          <button
            onClick={handleCreateNewBehavior}
            disabled={!canManageBehaviors}
            className="w-full flex items-center justify-center gap-2 bg-blue-600/20 hover:bg-blue-600/30 disabled:bg-slate-800 disabled:text-slate-500 disabled:border-slate-700 text-blue-400 border border-blue-500/30 py-2 rounded-lg font-medium text-xs transition-colors"
          >
            <Plus size={14} /> {t("library.behavior.new")}
          </button>

          <div className="relative mt-1">
            <Search size={14} className="absolute left-3 top-2.5 text-slate-500" />
            <input
              type="text"
              placeholder={t("library.behavior.searchPlaceholder")}
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg pl-8 pr-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-blue-500 transition-colors placeholder:text-slate-600"
            />
          </div>

          <div className="flex items-center gap-0 justify-between px-6">
            {(Object.entries(BEHAVIOR_TYPES) as Array<[BehaviorType, (typeof BEHAVIOR_TYPES)[BehaviorType]]>).map(
              ([key, value]) => {
                const IconComponent = value.icon;
                const label = t(BEHAVIOR_TYPE_LABEL_KEYS[key], undefined, value.label);
                return (
                  <TooltipIconButton
                    key={key}
                    onClick={() => toggleFilter(key)}
                    tooltip={t("library.behavior.filter", { label })}
                    aria-label={t("library.behavior.filter", { label })}
                    className={`p-1.5 rounded-md border transition-all flex-1 flex justify-center items-center ${activeFilters[key] ? `${value.bg} ${value.border} ${value.color}` : "bg-slate-900 border-slate-800 text-slate-600 hover:border-slate-700"}`}
                  >
                    <IconComponent size={14} />
                  </TooltipIconButton>
                );
              },
            )}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
          {filteredRegistry.length === 0 && (
            <div className="text-center text-xs text-slate-500 mt-6 px-4">
              {t("library.behavior.empty")}
            </div>
          )}
          {filteredRegistry.map((item) => {
            const config = BEHAVIOR_TYPES[item.type] || BEHAVIOR_TYPES.action;
            const Icon = config.icon;
            const usage = pendingDeleteSrc === item.src ? resolveUsage(item.src) : null;
            const isPendingDelete = pendingDeleteSrc === item.src;

            return (
              <div
                key={item.src}
                onClick={() => {
                  if (!canManageBehaviors) {
                    return;
                  }
                  handleEditBehavior(item);
                }}
                className={`p-3 rounded-lg border hover:border-slate-600 cursor-pointer group transition-colors ${form?.originalSrc === item.src ? `ring-1 ${config.border} bg-slate-800 border-transparent` : "border-slate-800 bg-slate-900"} ${isPendingDelete ? "border-red-900/80 bg-red-950/20" : ""}`}
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="overflow-hidden flex-1 min-w-0">
                    <div className={`flex items-center gap-1.5 text-[10px] font-bold ${config.color} mb-1 uppercase tracking-wider`}>
                      <Icon size={12} />{" "}
                      {t(BEHAVIOR_TYPE_LABEL_KEYS[item.type], undefined, config.label)}
                    </div>
                    <TruncatedWithTooltip
                      text={item.src}
                      className="text-xs font-mono text-slate-200 truncate"
                    />
                  </div>
                  <button
                    onClick={(event) => {
                      event.stopPropagation();
                      if (!canManageBehaviors) {
                        return;
                      }
                      setPendingDeleteSrc((previous) => (previous === item.src ? null : item.src));
                    }}
                    disabled={!canManageBehaviors}
                    aria-label={t("library.item.deleteAria", { label: item.src })}
                    className={`p-1 self-center shrink-0 transition-colors ${isPendingDelete ? "text-red-400" : "text-slate-400 hover:text-red-400"}`}
                  >
                    <Trash2 size={14} />
                  </button>
                </div>

                {usage && (
                  <div
                    onClick={(event) => event.stopPropagation()}
                    className="mt-3 rounded-lg border border-red-900/60 bg-slate-950/70 p-2.5 text-[10px] text-slate-300 space-y-2"
                  >
                    <div className="font-semibold text-red-300 uppercase tracking-wider">
                      {t("library.behavior.deleteConfirm")}
                    </div>

                  {usage.total > 0 ? (
                    <p className="leading-snug">
                        {t("library.behavior.deleteInUse", { count: usage.total })}
                    </p>
                  ) : (
                    <p className="leading-snug text-slate-400">
                        {t("library.behavior.deleteNoRefs")}
                    </p>
                  )}

                    {usage.total > 0 && (
                      <div className="max-h-24 overflow-y-auto custom-scrollbar space-y-1 border border-slate-800 rounded p-2 bg-slate-900/80">
                        {usage.locations.map((location, locationIndex) => (
                          <div key={`${location.label}-${locationIndex}`} className="font-mono text-[10px] text-slate-300">
                            {location.label}
                          </div>
                        ))}
                      </div>
                    )}

                    <div className="flex justify-end gap-2 pt-1">
                      <button
                        onClick={() => setPendingDeleteSrc(null)}
                        className="px-2.5 py-1 text-slate-400 hover:text-white transition-colors"
                      >
                        {t("common.cancel")}
                      </button>
                      <button
                        onClick={() => {
                          if (!canManageBehaviors) {
                            return;
                          }
                          onDeleteBehavior(item.src);
                          if (form && form.originalSrc === item.src) {
                            setForm(null);
                          }
                          setPendingDeleteSrc(null);
                        }}
                        className="px-2.5 py-1 rounded bg-red-600/90 hover:bg-red-500 text-white font-medium transition-colors"
                      >
                        {usage.total > 0
                          ? t("library.behavior.deleteAndClean", {
                              count: usage.total,
                            })
                          : t("common.delete")}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      <div className="w-2/3 bg-slate-900 flex flex-col overflow-y-auto custom-scrollbar">
        {!form ? (
          <div className="flex-1 flex flex-col items-center justify-center text-slate-500 p-8 text-center">
            <Terminal size={48} className="mb-4 opacity-20" />
            <p>
              {t("library.behavior.selectHint")}
            </p>
          </div>
        ) : (
          <div className="p-6 space-y-5 animate-in fade-in duration-200">
            {error && (
              <div className="bg-red-950/50 text-red-400 p-3 rounded border border-red-900/50 text-xs flex items-center gap-2">
                <AlertTriangle size={14} /> {error}
              </div>
            )}

            <div className="space-y-4">
              <div>
                <label className="block text-[10px] text-slate-400 mb-2 font-mono uppercase tracking-wider">
                  {t("library.behavior.type")}
                </label>
                <div className="flex gap-2">
                  {(Object.entries(BEHAVIOR_TYPES) as Array<[BehaviorType, (typeof BEHAVIOR_TYPES)[BehaviorType]]>).map(
                    ([typeKey, config]) => {
                      const isSelected = form.type === typeKey;
                      const Icon = config.icon;

                      return (
                        <button
                          key={typeKey}
                          onClick={() => handleBehaviorTypeChange(typeKey)}
                          disabled={!canManageBehaviors}
                          className={`flex-1 flex items-center justify-center gap-1.5 py-2 rounded-lg border text-xs font-medium transition-all ${isSelected ? `${config.bg} ${config.border} ${config.color}` : "bg-slate-950 border-slate-800 text-slate-500 hover:bg-slate-800 hover:text-slate-300"}`}
                        >
                          <Icon size={14} />{" "}
                          {t(BEHAVIOR_TYPE_LABEL_KEYS[typeKey], undefined, config.label)}
                        </button>
                      );
                    },
                  )}
                </div>
              </div>

              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("library.behavior.source")}
                </label>
                <input
                  type="text"
                  value={form.src}
                  disabled={!canManageBehaviors}
                  onChange={(event) => setForm({ ...form, src: event.target.value })}
                  placeholder={t("library.behavior.sourcePlaceholder")}
                  className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 font-mono text-sm"
                />
              </div>
            </div>

            <div>
              <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                {t("library.behavior.description")}
              </label>
              <input
                type="text"
                value={form.description}
                disabled={!canManageBehaviors}
                onChange={(event) => setForm({ ...form, description: event.target.value })}
                placeholder={t("library.behavior.descriptionPlaceholder")}
                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-sm"
              />
            </div>

            <div className="pt-4 border-t border-slate-800 flex-1 flex flex-col">
              <label className="flex items-center justify-between text-[10px] text-slate-400 mb-2 font-mono uppercase tracking-wider">
                <span>{t("library.behavior.simScript")}</span>
                <span className="text-[9px] bg-slate-800 px-1.5 py-0.5 rounded text-blue-300">{t("library.behavior.debuggerV2")}</span>
              </label>
              {activeBehaviorScriptContract && (
                <p className="text-[10px] text-slate-400 mb-2 font-mono">
                  {activeBehaviorScriptContract.hint}
                </p>
              )}
              <div className="bg-[#1e1e1e] rounded-lg border border-slate-700 flex flex-col flex-1 min-h-[200px]">
                <div className="bg-[#2d2d2d] px-3 py-1.5 text-[10px] text-slate-400 font-mono border-b border-[#404040] rounded-t-lg">
                  {activeBehaviorScriptContract?.signature || "function executeAction(args, context) {"}
                </div>
                <textarea
                  value={form.simScript}
                  disabled={!canManageBehaviors}
                  onChange={(event) => handleBehaviorScriptChange(event.target.value)}
                  spellCheck="false"
                  className="w-full flex-1 bg-transparent text-green-400 p-3 font-mono text-[12px] focus:outline-none resize-none leading-relaxed custom-scrollbar"
                />
                <div className="bg-[#2d2d2d] px-3 py-1.5 text-[10px] text-slate-400 font-mono border-t border-[#404040] rounded-b-lg">
                  {"}"}
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-4">
              <button
                onClick={() => setForm(null)}
                className="px-4 py-2 text-slate-400 hover:text-white text-xs font-medium transition-colors"
              >
                {t("common.cancel")}
              </button>
              <button
                onClick={handleSaveBehavior}
                disabled={!canManageBehaviors}
                className="flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-md text-xs font-medium shadow-md transition-colors"
              >
                <Save size={14} /> {t("library.behavior.save")}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );

  const metadataPacksPane = (
    <div className="flex flex-1 overflow-hidden">
      <div className="w-1/3 border-r border-slate-800 bg-slate-950/50 flex flex-col">
        <div className="p-4 border-b border-slate-800 flex flex-col gap-3 shrink-0">
          <button
            onClick={handleCreateNewPack}
            disabled={!canCreateMetadataPacks}
            className="w-full flex items-center justify-center gap-2 bg-blue-600/20 hover:bg-blue-600/30 disabled:bg-slate-800 disabled:text-slate-500 disabled:border-slate-700 text-blue-400 border border-blue-500/30 py-2 rounded-lg font-medium text-xs transition-colors"
          >
            <Plus size={14} /> {t("library.pack.new")}
          </button>

          <div className="relative mt-1">
            <Search size={14} className="absolute left-3 top-2.5 text-slate-500" />
            <input
              type="text"
              placeholder={t("library.pack.searchPlaceholder")}
              value={packSearchQuery}
              onChange={(event) => setPackSearchQuery(event.target.value)}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg pl-8 pr-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-blue-500 transition-colors placeholder:text-slate-600"
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
          {filteredPacks.length === 0 && (
            <div className="text-center text-xs text-slate-500 mt-6 px-4">
              {t("library.pack.empty")}
            </div>
          )}
          {filteredPacks.map((pack) => (
            <div
              key={pack.id}
              onClick={() => handleEditPack(pack)}
              className={`p-3 rounded-lg border hover:border-slate-600 cursor-pointer group transition-colors ${packForm?.originalId === pack.id ? "ring-1 border-blue-500 bg-slate-800" : "border-slate-800 bg-slate-900"}`}
            >
              <div className="flex items-center justify-between gap-2">
                <div className="overflow-hidden flex-1 min-w-0">
                  <div className="text-[10px] font-bold text-emerald-300 mb-1 uppercase tracking-wider">
                    {t("library.pack.tag")}
                  </div>
                  <TruncatedWithTooltip
                    text={pack.id}
                    className="text-xs font-mono text-slate-200 truncate"
                  />
                  <TruncatedWithTooltip
                    text={pack.label}
                    className="text-[10px] text-slate-400 truncate"
                  />
                  <div className="mt-2 flex flex-wrap gap-1.5">
                    {METADATA_SCOPE_ORDER.filter((scope) => pack.scopes.includes(scope)).map(
                      (scope) => {
                        const { icon: ScopeIcon, colorClass } = resolveScopeVisual(scope);
                        return (
                          <span
                            key={`${pack.id}-${scope}`}
                            className="inline-flex items-center gap-1 rounded border border-slate-700 bg-slate-950 px-1.5 py-0.5 text-[9px] text-slate-300 font-mono uppercase"
                          >
                            <ScopeIcon size={10} className={colorClass} />
                            {scope}
                          </span>
                        );
                      },
                    )}
                  </div>
                </div>
                <button
                  onClick={(event) => {
                    event.stopPropagation();
                    setPendingDeletePackId((previous) => (previous === pack.id ? null : pack.id));
                  }}
                  aria-label={t("library.item.deleteAria", { label: pack.id })}
                  className={`p-1 self-center shrink-0 transition-colors ${pendingDeletePackId === pack.id ? "text-red-400" : "text-slate-400 hover:text-red-400"}`}
                >
                  <Trash2 size={14} />
                </button>
              </div>

              {pendingDeletePackId === pack.id && (
                <div
                  onClick={(event) => event.stopPropagation()}
                  className="mt-3 rounded-lg border border-red-900/60 bg-slate-950/70 p-2.5 text-[10px] text-slate-300 space-y-2"
                >
                  <div className="font-semibold text-red-300 uppercase tracking-wider">
                    {t("library.pack.deleteConfirm")}
                  </div>
                  <p>{t("library.pack.deleteHelp")}</p>
                  <div className="flex justify-end gap-2 pt-1">
                    <button
                      onClick={() => setPendingDeletePackId(null)}
                      className="px-2.5 py-1 text-slate-400 hover:text-white transition-colors"
                    >
                      {t("common.cancel")}
                    </button>
                    <button
                      onClick={() => {
                        setMetadataPackRegistry(
                          metadataPackRegistry.filter((item) => item.id !== pack.id),
                        );
                        onDeleteMetadataPack?.(pack.id);
                        if (packForm?.originalId === pack.id) {
                          setPackForm(null);
                        }
                        setPendingDeletePackId(null);
                      }}
                      className="px-2.5 py-1 rounded bg-red-600/90 hover:bg-red-500 text-white font-medium transition-colors"
                    >
                      {t("common.delete")}
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      <div className="w-2/3 bg-slate-900 flex flex-col overflow-y-auto custom-scrollbar">
        {!packForm ? (
          <div className="flex-1 flex flex-col items-center justify-center text-slate-500 p-8 text-center">
            <Boxes size={48} className="mb-4 opacity-20" />
            <p>
              {t("library.pack.selectHint")}
            </p>
          </div>
        ) : (
          <div className="p-6 space-y-4 animate-in fade-in duration-200">
            {packError && (
              <div className="bg-red-950/50 text-red-400 p-3 rounded border border-red-900/50 text-xs flex items-center gap-2">
                <AlertTriangle size={14} /> {packError}
              </div>
            )}

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("library.pack.label")}
                </label>
                <input
                  aria-label={t("library.pack.fieldLabel")}
                  type="text"
                  value={packForm.label}
                  onChange={(event) => {
                    setPackForm({ ...packForm, label: event.target.value });
                    setPackError("");
                  }}
                  className={`w-full bg-slate-950 border rounded px-3 py-2 text-slate-200 focus:outline-none text-sm ${packLabelError ? "border-red-700 focus:border-red-500" : "border-slate-700 focus:border-blue-500"}`}
                />
                {packLabelError && (
                  <p className="mt-1 text-[10px] text-red-300">{packLabelError}</p>
                )}
              </div>
              <div>
                <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                  {t("library.pack.idAuto")}
                </label>
                <input
                  aria-label={t("library.pack.idAuto")}
                  type="text"
                  value={normalizedPackId}
                  readOnly
                  className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-300 focus:outline-none font-mono text-sm"
                />
              </div>
            </div>

            <div>
              <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                {t("library.pack.description")}
              </label>
              <input
                aria-label={t("library.pack.description")}
                type="text"
                value={packForm.description}
                onChange={(event) => setPackForm({ ...packForm, description: event.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-sm"
              />
            </div>

            <div>
              <label className="block text-[10px] text-slate-400 mb-2 font-mono uppercase tracking-wider">
                {t("library.pack.scopes")}
              </label>
              <div className="flex gap-2 flex-wrap">
                {METADATA_SCOPE_ORDER.map((scope) => {
                  const { icon: ScopeIcon, colorClass: iconColorClass } = resolveScopeVisual(scope);

                  return (
                    <label
                      key={scope}
                      className={`flex items-center gap-2 px-2 py-1 rounded border text-xs bg-slate-950 transition-colors ${
                        packForm.scopes[scope]
                          ? "border-slate-500 text-slate-100"
                          : "border-slate-700 text-slate-300"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={packForm.scopes[scope]}
                        onChange={(event) =>
                          setPackForm({
                            ...packForm,
                            scopes: {
                              ...packForm.scopes,
                              [scope]: event.target.checked,
                            },
                          })
                        }
                      />
                      <ScopeIcon size={12} className={iconColorClass} />
                      <span className="font-mono uppercase">{scope}</span>
                    </label>
                  );
                })}
              </div>
            </div>

            <div className="space-y-3 rounded-lg border border-slate-800/70 bg-slate-950/40 p-3">
              <div className="flex items-center justify-between">
                <label className="block text-[10px] text-slate-400 font-mono uppercase tracking-wider">
                  {t("library.pack.builder")}
                </label>
                <button
                  type="button"
                  onClick={() => setShowPackGeneratedPreview((previous) => !previous)}
                  className="text-[10px] flex items-center gap-1 px-2 py-1 rounded border border-slate-700 text-slate-300 hover:text-white hover:border-slate-600 transition-colors"
                >
                  <Eye size={12} />
                  {showPackGeneratedPreview ? t("library.pack.previewHide") : t("library.pack.previewShow")}
                </button>
              </div>

              <div className="space-y-3">
                {packForm.fields.map((field, fieldIndex) => {
                  const fieldErrors = packFieldValidation.fieldErrors.get(field.id) || [];
                  const hasFieldErrors = fieldErrors.length > 0;

                  return (
                    <div
                      key={field.id}
                      className={`rounded border p-3 space-y-2 ${hasFieldErrors ? "border-red-700/70 bg-red-950/15" : "border-slate-800 bg-slate-900/80"}`}
                    >
                    <div className="flex items-center justify-between">
                      <div className="text-[10px] uppercase tracking-wider text-slate-500 font-semibold">
                        {t("library.pack.field", { index: fieldIndex + 1 })}
                      </div>
                      <button
                        type="button"
                        onClick={() =>
                          setPackForm({
                            ...packForm,
                            fields: packForm.fields.filter((entry) => entry.id !== field.id),
                          })
                        }
                        className="text-[10px] px-2 py-1 rounded border border-red-800/60 text-red-300 hover:bg-red-900/30"
                      >
                        {t("common.remove")}
                      </button>
                    </div>

                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                          {t("library.pack.path")}
                        </label>
                        <input
                          type="text"
                          value={field.path}
                          onChange={(event) =>
                            setPackForm({
                              ...packForm,
                              fields: packForm.fields.map((entry) =>
                                entry.id === field.id
                                  ? { ...entry, path: event.target.value }
                                  : entry,
                              ),
                            })
                          }
                          placeholder={t("library.pack.placeholderPath")}
                          className={`w-full bg-slate-950 border rounded px-3 py-2 text-slate-200 focus:outline-none font-mono text-xs ${hasFieldErrors ? "border-red-700 focus:border-red-500" : "border-slate-700 focus:border-blue-500"}`}
                        />
                      </div>
                      <div>
                        <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                          {t("library.pack.fieldLabel")}
                        </label>
                        <input
                          type="text"
                          value={field.label}
                          onChange={(event) =>
                            setPackForm({
                              ...packForm,
                              fields: packForm.fields.map((entry) =>
                                entry.id === field.id
                                  ? { ...entry, label: event.target.value }
                                  : entry,
                              ),
                            })
                          }
                          placeholder={t("library.pack.placeholderLabel")}
                          className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-xs"
                        />
                      </div>
                    </div>

                    {fieldErrors.length > 0 && (
                      <div className="rounded border border-red-900/50 bg-red-950/30 px-2 py-1.5 text-[10px] text-red-200 space-y-1">
                        {fieldErrors.slice(0, 3).map((fieldError, errorIndex) => (
                          <div key={`${field.id}-error-${errorIndex}`}>{fieldError}</div>
                        ))}
                      </div>
                    )}

                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                          {t("library.pack.type")}
                        </label>
                        <select
                          value={field.kind}
                          onChange={(event) =>
                            setPackForm({
                              ...packForm,
                              fields: packForm.fields.map((entry) => {
                                if (entry.id !== field.id) {
                                  return entry;
                                }
                                const nextKind = event.target.value as VisualFieldKind;
                                const constantState =
                                  nextKind === "array"
                                    ? { constantValueText: "[]", constantValueBoolean: false }
                                    : nextKind === "json"
                                      ? { constantValueText: "{}", constantValueBoolean: false }
                                      : nextKind === "boolean"
                                        ? { constantValueText: "", constantValueBoolean: false }
                                        : { constantValueText: "", constantValueBoolean: false };
                                return {
                                  ...entry,
                                  kind: nextKind,
                                  isConstant: nextKind === "select" ? false : entry.isConstant,
                                  ...constantState,
                                };
                              }),
                            })
                          }
                          className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-xs cursor-pointer"
                        >
                          {FIELD_KINDS.map((kind) => (
                            <option key={kind.id} value={kind.id}>
                              {translateFieldKind(t, kind.id)}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div>
                        <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                          {t("library.pack.placeholder")}
                        </label>
                        <input
                          type="text"
                          value={field.placeholder}
                          onChange={(event) =>
                            setPackForm({
                              ...packForm,
                              fields: packForm.fields.map((entry) =>
                                entry.id === field.id
                                  ? { ...entry, placeholder: event.target.value }
                                  : entry,
                              ),
                            })
                          }
                          placeholder={t("library.pack.placeholderPlaceholder")}
                          className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-xs"
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                        {t("library.pack.helpText")}
                      </label>
                      <input
                        type="text"
                        value={field.help}
                        onChange={(event) =>
                          setPackForm({
                            ...packForm,
                            fields: packForm.fields.map((entry) =>
                              entry.id === field.id
                                ? { ...entry, help: event.target.value }
                                : entry,
                            ),
                          })
                        }
                        placeholder={t("library.pack.helpPlaceholder")}
                        className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-xs"
                      />
                    </div>

                    <div className="flex items-center gap-4 text-xs text-slate-300">
                      <label className="flex items-center gap-2">
                        <input
                          type="checkbox"
                          checked={field.required}
                          onChange={(event) =>
                            setPackForm({
                              ...packForm,
                              fields: packForm.fields.map((entry) =>
                                entry.id === field.id
                                  ? { ...entry, required: event.target.checked }
                                  : entry,
                              ),
                            })
                          }
                        />
                        {t("library.pack.required")}
                      </label>

                      {field.kind !== "select" && (
                        <label className="flex items-center gap-2">
                          <input
                            type="checkbox"
                            checked={field.isConstant}
                            onChange={(event) =>
                              setPackForm({
                                ...packForm,
                                fields: packForm.fields.map((entry) => {
                                  if (entry.id !== field.id) {
                                    return entry;
                                  }
                                  const shouldEnable = event.target.checked;
                                  return {
                                    ...entry,
                                    isConstant: shouldEnable,
                                    constantValueText: shouldEnable
                                      ? entry.kind === "array"
                                        ? entry.constantValueText || "[]"
                                        : entry.kind === "json"
                                          ? entry.constantValueText || "{}"
                                          : entry.kind === "number"
                                            ? entry.constantValueText || "0"
                                            : entry.constantValueText
                                      : entry.constantValueText,
                                  };
                                }),
                              })
                            }
                          />
                          {t("library.pack.constant")}
                        </label>
                      )}
                    </div>

                    {field.isConstant && field.kind !== "select" && (
                      <div>
                        <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                          {t("library.pack.constantValue")}
                        </label>
                        {field.kind === "boolean" ? (
                          <label className="flex items-center gap-2 text-xs text-slate-300">
                            <input
                              type="checkbox"
                              checked={field.constantValueBoolean}
                              onChange={(event) =>
                                setPackForm({
                                  ...packForm,
                                  fields: packForm.fields.map((entry) =>
                                    entry.id === field.id
                                      ? {
                                          ...entry,
                                          constantValueBoolean: event.target.checked,
                                        }
                                      : entry,
                                  ),
                                })
                              }
                            />
                            {field.constantValueBoolean ? "true" : "false"}
                          </label>
                        ) : field.kind === "array" || field.kind === "json" ? (
                          <textarea
                            value={field.constantValueText}
                            onChange={(event) =>
                              setPackForm({
                                ...packForm,
                                fields: packForm.fields.map((entry) =>
                                  entry.id === field.id
                                    ? {
                                        ...entry,
                                        constantValueText: event.target.value,
                                      }
                                    : entry,
                                ),
                              })
                            }
                            placeholder={field.kind === "array" ? "[]" : "{}"}
                            className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-yellow-300 focus:outline-none focus:border-blue-500 text-xs font-mono h-20 resize-y"
                          />
                        ) : field.kind === "textarea" ? (
                          <textarea
                            value={field.constantValueText}
                            onChange={(event) =>
                              setPackForm({
                                ...packForm,
                                fields: packForm.fields.map((entry) =>
                                  entry.id === field.id
                                    ? {
                                        ...entry,
                                        constantValueText: event.target.value,
                                      }
                                    : entry,
                                ),
                              })
                            }
                            placeholder={t("library.pack.constantValue")}
                            className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-xs h-20 resize-y"
                          />
                        ) : (
                          <input
                            type={field.kind === "number" ? "number" : "text"}
                            value={field.constantValueText}
                            onChange={(event) =>
                              setPackForm({
                                ...packForm,
                                fields: packForm.fields.map((entry) =>
                                  entry.id === field.id
                                    ? {
                                        ...entry,
                                        constantValueText: event.target.value,
                                      }
                                    : entry,
                                ),
                              })
                            }
                            placeholder={t("library.pack.constantValue")}
                            className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-xs font-mono"
                          />
                        )}
                      </div>
                    )}

                    {field.kind === "select" && (
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <label className="block text-[10px] text-slate-400 font-mono uppercase tracking-wider">
                            {t("library.pack.selectOptions")}
                          </label>
                          <button
                            type="button"
                            onClick={() =>
                              setPackForm({
                                ...packForm,
                                fields: packForm.fields.map((entry) =>
                                  entry.id === field.id
                                    ? {
                                        ...entry,
                                        options: [
                                          ...entry.options,
                                          { value: "", label: "" },
                                        ],
                                      }
                                    : entry,
                                ),
                              })
                            }
                            className="text-[10px] px-2 py-1 rounded border border-blue-700/60 text-blue-300 hover:bg-blue-900/30"
                          >
                            {t("library.pack.addOption")}
                          </button>
                        </div>

                        {field.options.map((option, optionIndex) => (
                          <div key={`${field.id}-option-${optionIndex}`} className="grid grid-cols-[1fr_1fr_auto] gap-2">
                            <input
                              type="text"
                              value={option.value}
                              onChange={(event) =>
                                setPackForm({
                                  ...packForm,
                                  fields: packForm.fields.map((entry) => {
                                    if (entry.id !== field.id) return entry;
                                    return {
                                      ...entry,
                                      options: entry.options.map((entryOption, entryIndex) =>
                                        entryIndex === optionIndex
                                          ? { ...entryOption, value: event.target.value }
                                          : entryOption,
                                      ),
                                    };
                                  }),
                                })
                              }
                              placeholder={t("library.pack.optionValue")}
                              className="w-full bg-slate-950 border border-slate-700 rounded px-2 py-1.5 text-slate-200 focus:outline-none focus:border-blue-500 text-xs font-mono"
                            />
                            <input
                              type="text"
                              value={option.label}
                              onChange={(event) =>
                                setPackForm({
                                  ...packForm,
                                  fields: packForm.fields.map((entry) => {
                                    if (entry.id !== field.id) return entry;
                                    return {
                                      ...entry,
                                      options: entry.options.map((entryOption, entryIndex) =>
                                        entryIndex === optionIndex
                                          ? { ...entryOption, label: event.target.value }
                                          : entryOption,
                                      ),
                                    };
                                  }),
                                })
                              }
                              placeholder={t("library.pack.optionLabel")}
                              className="w-full bg-slate-950 border border-slate-700 rounded px-2 py-1.5 text-slate-200 focus:outline-none focus:border-blue-500 text-xs"
                            />
                            <button
                              type="button"
                              onClick={() =>
                                setPackForm({
                                  ...packForm,
                                  fields: packForm.fields.map((entry) =>
                                    entry.id === field.id
                                      ? {
                                          ...entry,
                                          options: entry.options.filter(
                                            (_, entryIndex) => entryIndex !== optionIndex,
                                          ),
                                        }
                                      : entry,
                                  ),
                                })
                              }
                              className="px-2 py-1 rounded border border-red-800/60 text-red-300 hover:bg-red-900/30 text-[10px]"
                            >
                              x
                            </button>
                          </div>
                        ))}

                        <label className="flex items-center gap-2 text-xs text-slate-300">
                          <input
                            type="checkbox"
                            checked={field.allowCustomValue}
                            onChange={(event) =>
                              setPackForm({
                                ...packForm,
                                fields: packForm.fields.map((entry) =>
                                  entry.id === field.id
                                    ? { ...entry, allowCustomValue: event.target.checked }
                                    : entry,
                                ),
                              })
                            }
                          />
                          {t("library.pack.allowCustomValue")}
                        </label>
                      </div>
                    )}

                    {field.kind === "array" && (
                      <div>
                        <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                          {t("library.pack.arrayItemType")}
                        </label>
                        <select
                          value={field.arrayItemType}
                          onChange={(event) =>
                            setPackForm({
                              ...packForm,
                              fields: packForm.fields.map((entry) =>
                                entry.id === field.id
                                  ? {
                                      ...entry,
                                      arrayItemType: event.target.value as VisualArrayItemType,
                                    }
                                  : entry,
                              ),
                            })
                          }
                          className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-slate-200 focus:outline-none focus:border-blue-500 text-xs cursor-pointer"
                        >
                          {ARRAY_ITEM_TYPES.map((item) => (
                            <option key={item.id} value={item.id}>
                              {translateFieldKind(t, item.id)}
                            </option>
                          ))}
                        </select>
                      </div>
                    )}
                  </div>
                  );
                })}

                <button
                  type="button"
                  onClick={() =>
                    setPackForm({
                      ...packForm,
                      fields: [
                        ...packForm.fields,
                        createVisualField({
                          path: `field-${packForm.fields.length + 1}`,
                          label: t("library.pack.field", {
                            index: packForm.fields.length + 1,
                          }),
                        }),
                      ],
                    })
                  }
                  className="w-full px-3 py-2 rounded border border-blue-700/60 text-blue-300 hover:bg-blue-900/30 text-xs font-semibold"
                >
                  {t("library.pack.addField")}
                </button>
              </div>

              {showPackGeneratedPreview && (
                <div className="space-y-2">
                  <div
                    className={`text-[10px] rounded px-2 py-1 border ${packFieldValidation.hasBlockingErrors || packGeneratedPreview.errors.length > 0 ? "border-red-900/60 bg-red-950/20 text-red-200" : "border-emerald-900/60 bg-emerald-950/20 text-emerald-200"}`}
                  >
                    {packFieldValidation.hasBlockingErrors || packGeneratedPreview.errors.length > 0
                      ? t("library.pack.stateInvalid")
                      : t("library.pack.stateValid")}
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                      {t("library.pack.generatedSchema")}
                    </label>
                    <textarea
                      readOnly
                      value={JSON.stringify(packGeneratedPreview.schema, null, 2)}
                      className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-green-300 focus:outline-none font-mono text-[11px] h-36 resize-y"
                    />
                  </div>
                  <div>
                    <label className="block text-[10px] text-slate-400 mb-1 font-mono uppercase tracking-wider">
                      {t("library.pack.generatedUi")}
                    </label>
                    <textarea
                      readOnly
                      value={JSON.stringify(packGeneratedPreview.ui, null, 2)}
                      className="w-full bg-slate-950 border border-slate-700 rounded px-3 py-2 text-yellow-300 focus:outline-none font-mono text-[11px] h-36 resize-y"
                    />
                  </div>
                </div>
                </div>
              )}
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <button
                onClick={() => setPackForm(null)}
                className="px-4 py-2 text-slate-400 hover:text-white text-xs font-medium transition-colors"
              >
                {t("common.cancel")}
              </button>
              <button
                onClick={handleSavePack}
                disabled={!packCanSave}
                className="flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 disabled:text-slate-400 disabled:cursor-not-allowed text-white rounded-md text-xs font-medium shadow-md transition-colors"
              >
                <Save size={14} /> {t("library.pack.save")}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );

  return (
    <div className="fixed inset-0 z-[400] flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm">
      <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-5xl h-[85vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800 bg-slate-800/50 shrink-0">
          <h2 className="text-lg font-bold flex items-center gap-2 text-slate-200">
            <BookOpen className="text-blue-400" /> {t("library.title")}
          </h2>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
            <X size={20} />
          </button>
        </div>

        <div className="px-6 pt-3 border-b border-slate-800 bg-slate-900/80 flex items-center gap-3">
          <button
            onClick={() => setActiveTab("behaviors")}
            className={`pb-2 text-xs font-semibold uppercase tracking-wider border-b-2 transition-colors ${activeTab === "behaviors" ? "border-blue-500 text-blue-400" : "border-transparent text-slate-500 hover:text-slate-300"}`}
          >
            {t("library.tabs.behaviors")}
          </button>
          <button
            onClick={() => setActiveTab("metadata-packs")}
            className={`pb-2 text-xs font-semibold uppercase tracking-wider border-b-2 transition-colors ${activeTab === "metadata-packs" ? "border-blue-500 text-blue-400" : "border-transparent text-slate-500 hover:text-slate-300"}`}
          >
            {t("library.tabs.metadataPacks")}
          </button>
        </div>

        {activeTab === "behaviors" ? behaviorPane : metadataPacksPane}
      </div>
    </div>
  );
};
