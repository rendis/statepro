import { ChevronDown, ChevronRight } from "lucide-react";
import { useMemo, useState } from "react";

import { STUDIO_DS } from "../../constants";
import { useI18n } from "../../i18n";
import type {
  MetadataPackBinding,
  MetadataPackBindingMap,
  MetadataPackRegistry,
  MetadataScope,
} from "../../types";
import {
  applyPackConstantsToValues,
  buildDefaultBinding,
  buildMetadataPackRegistryIndex,
  collectOwnedPointersFromPack,
  findBindingOwnershipCollisionsDetailed,
  invertPointerCollisionRelation,
  getBindingsForEntity,
  listManualConflictingPointers,
  setBindingsForEntity,
} from "../../utils";
import { tryParseJsonObject } from "../../utils/behavior";
import { MetadataPackForm } from "./MetadataPackForm";

interface MetadataPackBindingsEditorProps {
  scope: MetadataScope;
  entityRef: string;
  packRegistry: MetadataPackRegistry;
  bindings: MetadataPackBindingMap;
  metadataRaw: string | undefined;
  onChangeBindings: (next: MetadataPackBindingMap) => void;
}

export const MetadataPackBindingsEditor = ({
  scope,
  entityRef,
  packRegistry,
  bindings,
  metadataRaw,
  onChangeBindings,
}: MetadataPackBindingsEditorProps) => {
  const { t } = useI18n();
  const [selectedPackId, setSelectedPackId] = useState("");
  const [collapsedByBindingId, setCollapsedByBindingId] = useState<Record<string, boolean>>({});
  const registryIndex = useMemo(
    () => buildMetadataPackRegistryIndex(packRegistry),
    [packRegistry],
  );
  const entityBindings = useMemo(
    () => getBindingsForEntity(bindings, scope, entityRef),
    [bindings, entityRef, scope],
  );
  const activeEntityCollisions = useMemo(
    () => findBindingOwnershipCollisionsDetailed(entityBindings, registryIndex),
    [entityBindings, registryIndex],
  );
  const parsedManualMetadata = tryParseJsonObject(metadataRaw).value;

  const updateEntityBindings = (nextEntityBindings: MetadataPackBinding[]) => {
    onChangeBindings(setBindingsForEntity(bindings, scope, entityRef, nextEntityBindings));
  };

  const packDisplayName = (packId: string): string => registryIndex.get(packId)?.id || packId;
  const relationLabel = (relation: "exact" | "ancestor" | "descendant"): string =>
    t(`metadataBindings.reason.relation.${relation}`, undefined, relation);

  const describeEntityCollision = (collision: {
    leftPackId: string;
    rightPackId: string;
    leftPointer: string;
    rightPointer: string;
    relation: "exact" | "ancestor" | "descendant";
  }): string =>
    t(
      "metadataBindings.reason.collisionSummary",
      {
        leftPack: packDisplayName(collision.leftPackId),
        leftPointer: collision.leftPointer,
        relation: relationLabel(collision.relation),
        rightPack: packDisplayName(collision.rightPackId),
        rightPointer: collision.rightPointer,
      },
      `Pack "${packDisplayName(collision.leftPackId)}": ${collision.leftPointer} (${relationLabel(collision.relation)}) vs pack "${packDisplayName(collision.rightPackId)}": ${collision.rightPointer}.`,
    );

  const canAttachPack = (packId: string): { ok: boolean; reason?: string } => {
    const pack = registryIndex.get(packId);
    if (!pack) {
      return { ok: false, reason: t("metadataBindings.reason.packMissing") };
    }
    if (!pack.scopes.includes(scope)) {
      return {
        ok: false,
        reason: t("metadataBindings.reason.scopeNotEnabled", { scope }),
      };
    }
    if (entityBindings.some((binding) => binding.packId === packId)) {
      return { ok: false, reason: t("metadataBindings.reason.alreadyAttached") };
    }

    if (activeEntityCollisions.length > 0) {
      const firstCollision = activeEntityCollisions[0];
      if (!firstCollision) {
        return {
          ok: false,
          reason: t("metadataBindings.reason.activeCollisions"),
        };
      }
      return {
        ok: false,
        reason: `${t("metadataBindings.reason.activeCollisions")} ${describeEntityCollision(firstCollision)}`,
      };
    }

    const candidateBinding = buildDefaultBinding(pack.id, scope, entityRef);
    const hypothetical = [
      ...entityBindings,
      candidateBinding,
    ];
    const collisions = findBindingOwnershipCollisionsDetailed(hypothetical, registryIndex);
    const candidateCollisions = collisions.filter(
      (collision) =>
        collision.leftBindingId === candidateBinding.id ||
        collision.rightBindingId === candidateBinding.id,
    );
    const candidateCollision =
      candidateCollisions.find((collision) => collision.relation !== "exact") ||
      candidateCollisions[0];
    if (candidateCollision) {
      const candidateIsLeft = candidateCollision.leftBindingId === candidateBinding.id;
      const candidatePointer = candidateIsLeft
        ? candidateCollision.leftPointer
        : candidateCollision.rightPointer;
      const otherPointer = candidateIsLeft
        ? candidateCollision.rightPointer
        : candidateCollision.leftPointer;
      const otherPackId = candidateIsLeft
        ? candidateCollision.rightPackId
        : candidateCollision.leftPackId;
      const candidateRelation = candidateIsLeft
        ? candidateCollision.relation
        : invertPointerCollisionRelation(candidateCollision.relation);
      return {
        ok: false,
        reason: t(
          "metadataBindings.reason.collidesWithPack",
          {
            otherPack: packDisplayName(otherPackId),
            candidatePointer,
            relation: relationLabel(candidateRelation),
            otherPointer,
          },
          `Collides with pack "${packDisplayName(otherPackId)}": ${candidatePointer} (${relationLabel(candidateRelation)}) vs ${otherPointer}.`,
        ),
      };
    }

    return { ok: true };
  };

  const packAttachValidationById = useMemo(() => {
    const result = new Map<string, { ok: boolean; reason?: string }>();
    packRegistry.forEach((pack) => {
      result.set(pack.id, canAttachPack(pack.id));
    });
    return result;
  }, [activeEntityCollisions, entityBindings, entityRef, packRegistry, registryIndex, scope]);

  const compatiblePacks = packRegistry.filter((pack) => {
    const result = packAttachValidationById.get(pack.id) || { ok: false };
    return result.ok;
  });

  const selectedPackValidation = selectedPackId
    ? packAttachValidationById.get(selectedPackId) || { ok: false }
    : { ok: false };

  const blockedPackSummaries = useMemo(
    () =>
      packRegistry
        .map((pack) => ({
          packId: pack.id,
          reason: packAttachValidationById.get(pack.id)?.reason,
          ok: packAttachValidationById.get(pack.id)?.ok ?? false,
        }))
        .filter((entry) => !entry.ok && Boolean(entry.reason)),
    [packAttachValidationById, packRegistry],
  );

  return (
    <div className="space-y-4 animate-in fade-in">
      {activeEntityCollisions.length > 0 && (
        <div className="text-[10px] text-red-200 bg-red-950/25 border border-red-900/60 rounded p-3 space-y-1">
          <div className="font-semibold uppercase tracking-wider">
            {t("metadataBindings.collisionsTitle")}
          </div>
          <div>
            {t("metadataBindings.collisionsImported")}
          </div>
          {activeEntityCollisions.slice(0, 4).map((collision, index) => (
            <div key={`${collision.leftBindingId}-${collision.rightBindingId}-${index}`}>
              {describeEntityCollision(collision)}
            </div>
          ))}
          {activeEntityCollisions.length > 4 && (
            <div>
              {t("metadataBindings.moreCollisions", {
                count: activeEntityCollisions.length - 4,
              })}
            </div>
          )}
        </div>
      )}

      <div className="bg-slate-950/50 border border-slate-800/60 rounded-lg p-3 space-y-3">
        <div className="flex flex-col gap-2">
          <label className={STUDIO_DS.labelMono}>{t("metadataBindings.attachLabel")}</label>
          <p className="text-[10px] text-slate-500">
            {t("metadataBindings.attachHelp")}
          </p>
          <div className="flex gap-2">
            <select
              value={selectedPackId}
              onChange={(event) => setSelectedPackId(event.target.value)}
              className={`${STUDIO_DS.input} cursor-pointer`}
            >
              <option value="">{t("metadataBindings.selectPack")}</option>
              {packRegistry.map((pack) => {
                const validation = packAttachValidationById.get(pack.id) || { ok: false };
                return (
                  <option key={pack.id} value={pack.id} disabled={!validation.ok}>
                    {pack.label} ({pack.id}){!validation.ok ? ` - ${t("metadataBindings.blockedShort")}` : ""}
                  </option>
                );
              })}
            </select>
            <button
              type="button"
              disabled={!selectedPackValidation.ok || !selectedPackId}
              onClick={() => {
                if (!selectedPackId) {
                  return;
                }
                const pack = registryIndex.get(selectedPackId);
                if (!pack) {
                  return;
                }
                const nextBinding = buildDefaultBinding(pack.id, scope, entityRef);
                setCollapsedByBindingId((previous) => ({
                  ...previous,
                  [nextBinding.id]: false,
                }));
                updateEntityBindings([
                  ...entityBindings,
                  {
                    ...nextBinding,
                    values: applyPackConstantsToValues(pack, nextBinding.values),
                  },
                ]);
                setSelectedPackId("");
              }}
              className="px-4 py-2 rounded border border-blue-700/60 text-blue-300 hover:bg-blue-900/30 disabled:border-slate-700 disabled:text-slate-500 disabled:hover:bg-transparent text-xs font-semibold"
            >
              {t("common.attach")}
            </button>
          </div>
          {selectedPackId && !selectedPackValidation.ok && (
            <p className="text-[10px] text-amber-400">{selectedPackValidation.reason}</p>
          )}
          {!selectedPackId && blockedPackSummaries.length > 0 && (
            <div className="text-[10px] text-slate-500 space-y-1">
              {blockedPackSummaries.slice(0, 3).map((entry) => (
                <div key={entry.packId}>
                  {entry.packId}: {entry.reason}
                </div>
              ))}
              {blockedPackSummaries.length > 3 && (
                <div>
                  {t("metadataBindings.moreBlocked", {
                    count: blockedPackSummaries.length - 3,
                  })}
                </div>
              )}
            </div>
          )}
          {compatiblePacks.length === 0 && (
            <p className="text-[10px] text-slate-500">
              {t("metadataBindings.noCompatible")}
            </p>
          )}
        </div>
      </div>

      {entityBindings.length === 0 && (
        <div className="text-xs text-slate-500 border border-slate-800 rounded-lg p-3 bg-slate-950/40">
          {t("metadataBindings.noneAttached")}
        </div>
      )}

      {entityBindings.map((binding) => {
        const pack = registryIndex.get(binding.packId);
        if (!pack) {
          return (
            <div
              key={binding.id}
              className="border border-red-900/60 rounded-lg p-3 bg-red-950/20 text-xs text-red-200"
            >
              {t("metadataBindings.missingPack")}: <span className="font-mono">{binding.packId}</span>
            </div>
          );
        }

        const ownedPointers = collectOwnedPointersFromPack(pack);
        const manualConflicts = listManualConflictingPointers(parsedManualMetadata, ownedPointers);
        const isCollapsed = Boolean(collapsedByBindingId[binding.id]);

        return (
          <div
            key={binding.id}
            className="relative space-y-3 border border-slate-800 rounded-lg bg-slate-950/50 p-3"
          >
            <button
              type="button"
              data-testid={`toggle-pack-${binding.id}`}
              aria-label={t("metadataBindings.togglePack", { packId: pack.id })}
              onClick={() =>
                setCollapsedByBindingId((previous) => ({
                  ...previous,
                  [binding.id]: !Boolean(previous[binding.id]),
                }))
              }
              className="absolute top-2 right-2 inline-flex items-center justify-center w-5 h-5 rounded border border-slate-700 text-slate-300 hover:bg-slate-900/50"
            >
              {isCollapsed ? <ChevronRight size={12} /> : <ChevronDown size={12} />}
            </button>

            <div className="flex items-center justify-between gap-2 pr-8">
              <div>
                <div className="text-xs font-semibold text-slate-200">{pack.label}</div>
                <div className="text-[10px] text-slate-500 font-mono">{pack.id}</div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => {
                    setCollapsedByBindingId((previous) => {
                      if (!(binding.id in previous)) {
                        return previous;
                      }
                      const next = { ...previous };
                      delete next[binding.id];
                      return next;
                    });
                    updateEntityBindings(entityBindings.filter((entry) => entry.id !== binding.id));
                  }}
                  className="px-2 py-1 rounded border border-red-700/60 text-red-300 hover:bg-red-900/30 text-[10px]"
                >
                  {t("common.remove")}
                </button>
              </div>
            </div>

            {isCollapsed ? (
              <div
                data-testid={`collapsed-pack-indicator-${binding.id}`}
                className="mt-1 space-y-1"
              >
                <div className="h-1 rounded-full bg-slate-800/70">
                  <div className="h-full w-2/3 rounded-full bg-slate-600/60" />
                </div>
                <div className="h-1 rounded-full bg-slate-800/70">
                  <div className="h-full w-1/3 rounded-full bg-slate-600/50" />
                </div>
              </div>
            ) : (
              <>
                {manualConflicts.length > 0 && (
                  <div className="text-[10px] text-amber-300 bg-amber-950/20 border border-amber-900/50 rounded p-2">
                    {t("metadataBindings.manualOverride")}
                    <div className="font-mono mt-1">
                      {manualConflicts.slice(0, 5).join(", ")}
                      {manualConflicts.length > 5 ? " ..." : ""}
                    </div>
                  </div>
                )}

                <MetadataPackForm
                  pack={pack}
                  value={binding.values}
                  onChange={(nextValues) => {
                    const nextEntityBindings = entityBindings.map((entry) =>
                      entry.id === binding.id
                        ? {
                            ...entry,
                            values: nextValues,
                          }
                        : entry,
                    );
                    updateEntityBindings(nextEntityBindings);
                  }}
                />
              </>
            )}
          </div>
        );
      })}
    </div>
  );
};
