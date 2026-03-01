import { describe, expect, it } from "vitest";

import { BEHAVIOR_TYPES, REALITY_TYPES, STUDIO_ICON_REGISTRY, STUDIO_ICONS } from "../constants";

describe("iconography", () => {
  it("mantiene iconos únicos para conceptos de dominio", () => {
    const strictDomainIcons = [
      STUDIO_ICONS.reality.normal,
      STUDIO_ICONS.reality.success,
      STUDIO_ICONS.reality.error,
      STUDIO_ICONS.reality.initial,
      STUDIO_ICONS.behavior.action,
      STUDIO_ICONS.behavior.invoke,
      STUDIO_ICONS.behavior.condition,
      STUDIO_ICONS.behavior.observer,
      STUDIO_ICONS.transition.trigger.on,
      STUDIO_ICONS.transition.trigger.always,
      STUDIO_ICONS.transition.type.default,
      STUDIO_ICONS.transition.type.notify,
      STUDIO_ICONS.phase.entry,
      STUDIO_ICONS.phase.exit,
      STUDIO_ICONS.phase.onTransition,
    ];

    expect(new Set(strictDomainIcons).size).toBe(strictDomainIcons.length);
  });

  it("config.ts usa STUDIO_ICON_REGISTRY como fuente de verdad", () => {
    expect(REALITY_TYPES.normal.icon).toBe(STUDIO_ICON_REGISTRY.reality.normal.icon);
    expect(REALITY_TYPES.normal.color).toBe(STUDIO_ICON_REGISTRY.reality.normal.colors.base);
    expect(REALITY_TYPES.normal.border).toBe(STUDIO_ICON_REGISTRY.reality.normal.border);
    expect(REALITY_TYPES.normal.bg).toBe(STUDIO_ICON_REGISTRY.reality.normal.bg);

    expect(REALITY_TYPES.success.icon).toBe(STUDIO_ICON_REGISTRY.reality.success.icon);
    expect(REALITY_TYPES.error.icon).toBe(STUDIO_ICON_REGISTRY.reality.error.icon);

    expect(BEHAVIOR_TYPES.action.icon).toBe(STUDIO_ICON_REGISTRY.behavior.action.icon);
    expect(BEHAVIOR_TYPES.action.color).toBe(STUDIO_ICON_REGISTRY.behavior.action.colors.base);
    expect(BEHAVIOR_TYPES.action.border).toBe(STUDIO_ICON_REGISTRY.behavior.action.border);
    expect(BEHAVIOR_TYPES.action.bg).toBe(STUDIO_ICON_REGISTRY.behavior.action.bg);

    expect(BEHAVIOR_TYPES.invoke.icon).toBe(STUDIO_ICON_REGISTRY.behavior.invoke.icon);
    expect(BEHAVIOR_TYPES.condition.icon).toBe(STUDIO_ICON_REGISTRY.behavior.condition.icon);
    expect(BEHAVIOR_TYPES.observer.icon).toBe(STUDIO_ICON_REGISTRY.behavior.observer.icon);
  });

  it("STUDIO_ICONS se deriva del registro central", () => {
    expect(STUDIO_ICONS.reality.normal).toBe(STUDIO_ICON_REGISTRY.reality.normal.icon);
    expect(STUDIO_ICONS.behavior.action).toBe(STUDIO_ICON_REGISTRY.behavior.action.icon);
    expect(STUDIO_ICONS.transition.trigger.on).toBe(STUDIO_ICON_REGISTRY.transition.trigger.on.icon);
    expect(STUDIO_ICONS.transition.type.default).toBe(STUDIO_ICON_REGISTRY.transition.type.default.icon);
  });

  it("expone variantes de color clave para transiciones y estado", () => {
    expect(STUDIO_ICON_REGISTRY.transition.trigger.on.colors).toEqual({
      base: "text-yellow-400",
      muted: "text-slate-400",
      accent: "text-blue-400",
      emphasis: "text-blue-300",
    });
    expect(STUDIO_ICON_REGISTRY.transition.type.default.colors).toEqual({
      base: "text-slate-400",
      accent: "text-blue-500",
      emphasis: "text-slate-300",
    });
    expect(STUDIO_ICON_REGISTRY.status.warning.colors).toEqual({
      base: "text-red-300",
      muted: "text-red-200",
    });
  });

  it("evita la colisión action vs trigger.on", () => {
    expect(STUDIO_ICONS.behavior.action).not.toBe(STUDIO_ICONS.transition.trigger.on);
  });
});
