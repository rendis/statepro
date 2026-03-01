import Ajv2020 from "ajv/dist/2020";

import type {
  JsonObject,
  JsonValue,
  MetadataPackBinding,
  MetadataPackBindingMap,
  MetadataPackDefinition,
  MetadataPackRegistry,
  MetadataScope,
  StudioMetadataEnvelope,
  TransitionTriggerKind,
} from "../types";

export const STUDIO_METADATA_NAMESPACE = "__studio";
export const METADATA_PACK_MACHINE_ENTITY_REF = "machine";

const METADATA_SCOPES: MetadataScope[] = ["machine", "universe", "reality", "transition"];

const ajv = new Ajv2020({
  allErrors: true,
  strict: false,
  allowUnionTypes: true,
});

const validatorCache = new Map<string, ReturnType<typeof ajv.compile>>();

export type PointerCollisionRelation = "exact" | "ancestor" | "descendant";

export interface PointerCollisionDetail {
  leftBindingId: string;
  rightBindingId: string;
  leftPackId: string;
  rightPackId: string;
  leftPointer: string;
  rightPointer: string;
  relation: PointerCollisionRelation;
}

export interface FieldPathCollisionDetail {
  leftIndex: number;
  rightIndex: number;
  leftPath: string;
  rightPath: string;
  leftPointer: string;
  rightPointer: string;
  relation: PointerCollisionRelation;
}

export interface MetadataBindingValidationError {
  path: string;
  detail: string;
  message: string;
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const toJsonObject = (value: unknown): JsonObject | null => {
  if (!isRecord(value)) {
    return null;
  }
  return value as JsonObject;
};

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

  if (!isRecord(value)) {
    return false;
  }

  return Object.values(value).every((entry) => isJsonValue(entry));
};

const encodePointerSegment = (segment: string): string =>
  segment.replace(/~/g, "~0").replace(/\//g, "~1");

const decodePointerSegment = (segment: string): string =>
  segment.replace(/~1/g, "/").replace(/~0/g, "~");

const cloneJson = <T>(value: T): T => structuredClone(value);

const isFiniteIndex = (token: string): boolean => {
  if (!/^\d+$/.test(token)) {
    return false;
  }
  const parsed = Number.parseInt(token, 10);
  return Number.isFinite(parsed);
};

const normalizeScope = (scope: unknown): MetadataScope | null => {
  if (typeof scope !== "string") {
    return null;
  }
  return METADATA_SCOPES.includes(scope as MetadataScope) ? (scope as MetadataScope) : null;
};

const normalizePackDefinition = (value: unknown): MetadataPackDefinition | null => {
  if (!isRecord(value)) {
    return null;
  }

  const id = typeof value.id === "string" ? value.id.trim() : "";
  const label = typeof value.label === "string" ? value.label.trim() : "";
  const schema = toJsonObject(value.schema);
  if (!id || !label || !schema) {
    return null;
  }

  const scopes = Array.isArray(value.scopes)
    ? value.scopes
        .map((scope) => normalizeScope(scope))
        .filter((scope): scope is MetadataScope => Boolean(scope))
    : [];

  const normalizedScopes: MetadataScope[] =
    scopes.length > 0 ? Array.from(new Set(scopes)) : ["machine"];
  const description = typeof value.description === "string" ? value.description : undefined;
  const ui = toJsonObject(value.ui) ?? undefined;

  return {
    id,
    label,
    description,
    scopes: normalizedScopes,
    schema,
    ui: ui as MetadataPackDefinition["ui"],
  };
};

const normalizeBinding = (value: unknown): MetadataPackBinding | null => {
  if (!isRecord(value)) {
    return null;
  }
  const id = typeof value.id === "string" ? value.id.trim() : "";
  const packId = typeof value.packId === "string" ? value.packId.trim() : "";
  const scope = normalizeScope(value.scope);
  const entityRef = typeof value.entityRef === "string" ? value.entityRef.trim() : "";
  const values = toJsonObject(value.values);

  if (!id || !packId || !scope || !entityRef || !values) {
    return null;
  }

  return {
    id,
    packId,
    scope,
    entityRef,
    values,
  };
};

export const createEmptyMetadataPackBindingMap = (): MetadataPackBindingMap => ({
  machine: [],
  universe: [],
  reality: [],
  transition: [],
});

export const normalizeMetadataPackRegistry = (value: unknown): MetadataPackRegistry => {
  if (!Array.isArray(value)) {
    return [];
  }

  return value
    .map((item) => normalizePackDefinition(item))
    .filter((item): item is MetadataPackDefinition => Boolean(item));
};

export const normalizeMetadataPackBindingMap = (
  value: unknown,
): MetadataPackBindingMap => {
  const empty = createEmptyMetadataPackBindingMap();
  if (!isRecord(value)) {
    return empty;
  }

  METADATA_SCOPES.forEach((scope) => {
    const raw = value[scope];
    if (!Array.isArray(raw)) {
      return;
    }
    empty[scope] = raw
      .map((item) => normalizeBinding(item))
      .filter((item): item is MetadataPackBinding => Boolean(item) && item.scope === scope);
  });

  return empty;
};

export const createStudioMetadataEnvelope = (
  packRegistry: MetadataPackRegistry,
  bindings: MetadataPackBindingMap,
): StudioMetadataEnvelope => ({
  version: 1,
  packRegistry: cloneJson(packRegistry),
  bindings: cloneJson(bindings),
});

export const isStudioEnvelopeEmpty = (envelope: StudioMetadataEnvelope): boolean =>
  envelope.packRegistry.length === 0 &&
  METADATA_SCOPES.every((scope) => envelope.bindings[scope].length === 0);

export const extractStudioMetadataEnvelope = (
  metadata: JsonObject,
): { manualMetadata: JsonObject; envelope: StudioMetadataEnvelope | null } => {
  const manualMetadata = cloneJson(metadata);
  const rawEnvelope = manualMetadata[STUDIO_METADATA_NAMESPACE];
  delete manualMetadata[STUDIO_METADATA_NAMESPACE];

  const envelopeObj = toJsonObject(rawEnvelope);
  if (!envelopeObj) {
    return { manualMetadata, envelope: null };
  }

  const envelope: StudioMetadataEnvelope = {
    version: 1,
    packRegistry: normalizeMetadataPackRegistry(envelopeObj.packRegistry),
    bindings: normalizeMetadataPackBindingMap(envelopeObj.bindings),
  };

  return { manualMetadata, envelope };
};

export const injectStudioMetadataEnvelope = (
  metadata: JsonObject,
  envelope: StudioMetadataEnvelope | null,
): JsonObject => {
  const next = cloneJson(metadata);
  delete next[STUDIO_METADATA_NAMESPACE];
  if (!envelope || isStudioEnvelopeEmpty(envelope)) {
    return next;
  }

  next[STUDIO_METADATA_NAMESPACE] = envelope as unknown as JsonValue;
  return next;
};

export const buildMetadataPackRegistryIndex = (
  registry: MetadataPackRegistry,
): Map<string, MetadataPackDefinition> => {
  const map = new Map<string, MetadataPackDefinition>();
  registry.forEach((pack) => map.set(pack.id, pack));
  return map;
};

export const buildMachineEntityRef = (): string => METADATA_PACK_MACHINE_ENTITY_REF;

export const buildUniverseEntityRef = (universeId: string): string => `U:${universeId}`;

export const buildRealityEntityRef = (universeId: string, realityId: string): string =>
  `U:${universeId}:R:${realityId}`;

export const buildTransitionEntityRef = (
  universeId: string,
  realityId: string,
  triggerKind: TransitionTriggerKind,
  eventName: string | undefined,
  order: number,
): string =>
  `U:${universeId}:R:${realityId}:T:${triggerKind}:${eventName || "-" }:${order}`;

export const getBindingsForEntity = (
  bindings: MetadataPackBindingMap,
  scope: MetadataScope,
  entityRef: string,
): MetadataPackBinding[] =>
  bindings[scope]
    .filter((binding) => binding.entityRef === entityRef)
    .map((binding) => cloneJson(binding));

export const setBindingsForEntity = (
  bindings: MetadataPackBindingMap,
  scope: MetadataScope,
  entityRef: string,
  nextEntityBindings: MetadataPackBinding[],
): MetadataPackBindingMap => {
  const next = cloneJson(bindings);
  const remaining = next[scope].filter((binding) => binding.entityRef !== entityRef);
  next[scope] = [...remaining, ...nextEntityBindings];
  return next;
};

export const removePackBindingsByPackId = (
  bindings: MetadataPackBindingMap,
  packId: string,
): MetadataPackBindingMap => {
  const next = cloneJson(bindings);
  METADATA_SCOPES.forEach((scope) => {
    next[scope] = next[scope].filter((binding) => binding.packId !== packId);
  });
  return next;
};

export const replacePackIdInBindings = (
  bindings: MetadataPackBindingMap,
  previousPackId: string,
  nextPackId: string,
): MetadataPackBindingMap => {
  if (!previousPackId || !nextPackId || previousPackId === nextPackId) {
    return cloneJson(bindings);
  }

  const next = cloneJson(bindings);
  METADATA_SCOPES.forEach((scope) => {
    next[scope] = next[scope].map((binding) =>
      binding.packId === previousPackId
        ? {
            ...binding,
            packId: nextPackId,
          }
        : binding,
    );
  });
  return next;
};

export const normalizeJsonPointer = (pointer: string): string => {
  if (!pointer) {
    return "";
  }
  const prefixed = pointer.startsWith("/") ? pointer : `/${pointer}`;
  return prefixed.replace(/\/+$/, "");
};

export const pointerTokens = (pointer: string): string[] =>
  normalizeJsonPointer(pointer)
    .split("/")
    .filter(Boolean)
    .map((segment) => decodePointerSegment(segment));

export const classifyPointerCollision = (
  left: string,
  right: string,
): PointerCollisionRelation | null => {
  const a = normalizeJsonPointer(left);
  const b = normalizeJsonPointer(right);

  if (!a || !b) {
    return null;
  }

  if (a === b) {
    return "exact";
  }

  if (b.startsWith(`${a}/`)) {
    return "ancestor";
  }

  if (a.startsWith(`${b}/`)) {
    return "descendant";
  }

  return null;
};

export const invertPointerCollisionRelation = (
  relation: PointerCollisionRelation,
): PointerCollisionRelation => {
  if (relation === "ancestor") {
    return "descendant";
  }
  if (relation === "descendant") {
    return "ancestor";
  }
  return "exact";
};

export const toPointerCollisionRelationLabel = (
  relation: PointerCollisionRelation,
): "exacta" | "ancestro" | "descendiente" => {
  if (relation === "exact") {
    return "exacta";
  }
  if (relation === "ancestor") {
    return "ancestro";
  }
  return "descendiente";
};

export const isPointerCollision = (left: string, right: string): boolean => {
  return classifyPointerCollision(left, right) !== null;
};

export const collectSchemaOwnedPointers = (schema: JsonObject): string[] => {
  const pointers = new Set<string>();

  const walk = (node: JsonObject, basePointer: string): void => {
    const properties = toJsonObject(node.properties);
    if (!properties) {
      return;
    }

    Object.entries(properties).forEach(([key, childNode]) => {
      const child = toJsonObject(childNode);
      const pointer = `${basePointer}/${encodePointerSegment(key)}`;
      pointers.add(normalizeJsonPointer(pointer));
      if (child) {
        walk(child, pointer);
      }
    });
  };

  walk(schema, "");
  return Array.from(pointers).sort();
};

export const collectOwnedPointersFromPack = (pack: MetadataPackDefinition): string[] =>
  collectSchemaOwnedPointers(pack.schema);

export const collectOwnedPointersForBinding = (
  binding: MetadataPackBinding,
  registryIndex: Map<string, MetadataPackDefinition>,
): string[] => {
  const pack = registryIndex.get(binding.packId);
  if (!pack) {
    return [];
  }
  return collectOwnedPointersFromPack(pack);
};

export const findBindingOwnershipCollisionsDetailed = (
  entityBindings: MetadataPackBinding[],
  registryIndex: Map<string, MetadataPackDefinition>,
): PointerCollisionDetail[] => {
  const collisions: PointerCollisionDetail[] = [];
  const seen = new Set<string>();

  for (let i = 0; i < entityBindings.length; i += 1) {
    const left = entityBindings[i];
    if (!left) {
      continue;
    }
    const leftPointers = Array.from(
      new Set(collectOwnedPointersForBinding(left, registryIndex)),
    );
    for (let j = i + 1; j < entityBindings.length; j += 1) {
      const right = entityBindings[j];
      if (!right) {
        continue;
      }
      const rightPointers = Array.from(
        new Set(collectOwnedPointersForBinding(right, registryIndex)),
      );

      leftPointers.forEach((leftPointer) => {
        rightPointers.forEach((rightPointer) => {
          const relation = classifyPointerCollision(leftPointer, rightPointer);
          if (!relation) {
            return;
          }
          const collisionKey = `${left.id}::${right.id}::${leftPointer}::${rightPointer}`;
          if (seen.has(collisionKey)) {
            return;
          }
          seen.add(collisionKey);
          collisions.push({
            leftBindingId: left.id,
            rightBindingId: right.id,
            leftPackId: left.packId,
            rightPackId: right.packId,
            leftPointer,
            rightPointer,
            relation,
          });
        });
      });
    }
  }

  return collisions;
};

export const findBindingOwnershipCollisions = (
  entityBindings: MetadataPackBinding[],
  registryIndex: Map<string, MetadataPackDefinition>,
): Array<{ leftBindingId: string; rightBindingId: string; pointerLeft: string; pointerRight: string }> =>
  findBindingOwnershipCollisionsDetailed(entityBindings, registryIndex).map((collision) => ({
    leftBindingId: collision.leftBindingId,
    rightBindingId: collision.rightBindingId,
    pointerLeft: collision.leftPointer,
    pointerRight: collision.rightPointer,
  }));

const normalizeDotPath = (path: string): string =>
  path
    .split(".")
    .map((segment) => segment.trim())
    .filter(Boolean)
    .join(".");

const dotPathToPointer = (path: string): string => {
  const normalizedPath = normalizeDotPath(path);
  if (!normalizedPath) {
    return "";
  }
  const segments = normalizedPath
    .split(".")
    .map((segment) => encodePointerSegment(segment));
  return normalizeJsonPointer(`/${segments.join("/")}`);
};

export const findFieldPathCollisionsDetailed = (
  paths: string[],
): FieldPathCollisionDetail[] => {
  const collisions: FieldPathCollisionDetail[] = [];
  const normalizedPaths = paths.map((path) => normalizeDotPath(path));
  const pointers = normalizedPaths.map((path) => dotPathToPointer(path));
  const seen = new Set<string>();

  for (let i = 0; i < pointers.length; i += 1) {
    const leftPointer = pointers[i];
    if (!leftPointer) {
      continue;
    }
    for (let j = i + 1; j < pointers.length; j += 1) {
      const rightPointer = pointers[j];
      if (!rightPointer) {
        continue;
      }

      const relation = classifyPointerCollision(leftPointer, rightPointer);
      if (!relation) {
        continue;
      }

      const collisionKey = `${i}::${j}::${leftPointer}::${rightPointer}`;
      if (seen.has(collisionKey)) {
        continue;
      }
      seen.add(collisionKey);

      collisions.push({
        leftIndex: i,
        rightIndex: j,
        leftPath: normalizedPaths[i] || paths[i] || "",
        rightPath: normalizedPaths[j] || paths[j] || "",
        leftPointer,
        rightPointer,
        relation,
      });
    }
  }

  return collisions;
};

export const deepMergeJsonObjects = (base: JsonObject, override: JsonObject): JsonObject => {
  const result: JsonObject = cloneJson(base);

  Object.entries(override).forEach(([key, value]) => {
    const previous = result[key];
    if (
      isRecord(previous) &&
      isRecord(value) &&
      !Array.isArray(previous) &&
      !Array.isArray(value)
    ) {
      result[key] = deepMergeJsonObjects(previous as JsonObject, value as JsonObject);
      return;
    }
    result[key] = cloneJson(value);
  });

  return result;
};

export const getValueAtPointer = (
  value: JsonValue | undefined,
  pointer: string,
): JsonValue | undefined => {
  const tokens = pointerTokens(pointer);
  if (tokens.length === 0) {
    return value;
  }

  let cursor: unknown = value;
  for (const token of tokens) {
    if (Array.isArray(cursor) && isFiniteIndex(token)) {
      cursor = cursor[Number.parseInt(token, 10)];
      continue;
    }
    if (!isRecord(cursor)) {
      return undefined;
    }
    cursor = cursor[token];
  }

  return cursor as JsonValue | undefined;
};

export const setValueAtPointer = (
  target: JsonObject,
  pointer: string,
  value: JsonValue,
): JsonObject => {
  const tokens = pointerTokens(pointer);
  if (tokens.length === 0) {
    return cloneJson(target);
  }

  const root = cloneJson(target) as Record<string, unknown>;
  let cursor: Record<string, unknown> = root;

  for (let index = 0; index < tokens.length; index += 1) {
    const token = tokens[index];
    const isLast = index === tokens.length - 1;
    if (!token) {
      continue;
    }
    if (isLast) {
      cursor[token] = cloneJson(value);
      continue;
    }

    const nextValue = cursor[token];
    if (!isRecord(nextValue)) {
      cursor[token] = {};
    }
    cursor = cursor[token] as Record<string, unknown>;
  }

  return root as JsonObject;
};

export const hasManualValueForPointer = (
  manualMetadata: JsonObject,
  pointer: string,
): boolean => getValueAtPointer(manualMetadata, pointer) !== undefined;

export const listManualConflictingPointers = (
  manualMetadata: JsonObject,
  ownedPointers: string[],
): string[] =>
  ownedPointers.filter((pointer) => hasManualValueForPointer(manualMetadata, pointer));

export const collectPackConstantPointers = (
  pack: MetadataPackDefinition,
): Array<{ pointer: string; value: JsonValue }> => {
  if (!pack.ui) {
    return [];
  }

  return Object.entries(pack.ui)
    .filter(([, ui]) => Boolean(ui?.constant) && isJsonValue(ui?.constantValue))
    .map(([pointer, ui]) => ({
      pointer: normalizeJsonPointer(pointer),
      value: cloneJson(ui.constantValue as JsonValue),
    }))
    .filter((entry) => Boolean(entry.pointer));
};

export const applyPackConstantsToValues = (
  pack: MetadataPackDefinition,
  values: JsonObject,
): JsonObject => {
  const constants = collectPackConstantPointers(pack);
  if (constants.length === 0) {
    return cloneJson(values);
  }

  return constants.reduce<JsonObject>((accumulator, entry) => {
    return setValueAtPointer(accumulator, entry.pointer, entry.value);
  }, cloneJson(values));
};

export const validateBindingWithPack = (
  binding: MetadataPackBinding,
  pack: MetadataPackDefinition,
): MetadataBindingValidationError[] => {
  const cacheKey = pack.id;
  let validator = validatorCache.get(cacheKey);
  if (!validator) {
    validator = ajv.compile(pack.schema as object);
    validatorCache.set(cacheKey, validator);
  }

  const ok = validator(binding.values);
  if (ok) {
    return [];
  }

  return (validator.errors || []).map((error) => {
    const path = error.instancePath || "/";
    const detail = error.message || "invalid value";
    return {
      path,
      detail,
      message: `${path} ${detail}`.trim(),
    };
  });
};

export const mergePackBindingsToMetadata = (
  bindings: MetadataPackBinding[],
  registryIndex: Map<string, MetadataPackDefinition>,
): JsonObject =>
  bindings.reduce<JsonObject>((accumulator, binding) => {
    const pack = registryIndex.get(binding.packId);
    if (!pack) {
      return accumulator;
    }
    return deepMergeJsonObjects(accumulator, applyPackConstantsToValues(pack, binding.values));
  }, {});

export const buildDefaultBinding = (
  packId: string,
  scope: MetadataScope,
  entityRef: string,
): MetadataPackBinding => ({
  id: `${scope}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  packId,
  scope,
  entityRef,
  values: {},
});
