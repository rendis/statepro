import "@testing-library/jest-dom";
import { act, renderHook } from "@testing-library/react";
import { StrictMode, type ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { StudioPerformanceFeatureFlags } from "../types";
import { computeRenderPressure, usePerformanceController } from "../utils/performanceController";

const defaultFlags: StudioPerformanceFeatureFlags = {
  mode: "auto",
  staticPressureThreshold: 1200,
  onEmaMs: 18,
  offEmaMs: 14,
  onMissRatio: 0.25,
  offMissRatio: 0.1,
};

let frameNow = 0;
let frameId = 0;
const frameCallbacks = new Map<number, FrameRequestCallback>();

const runFrame = (deltaMs: number) => {
  frameNow += deltaMs;
  const callbacks = [...frameCallbacks.values()];
  frameCallbacks.clear();
  callbacks.forEach((callback) => callback(frameNow));
};

const runFrames = (count: number, deltaMs: number) => {
  for (let index = 0; index < count; index += 1) {
    act(() => {
      runFrame(deltaMs);
    });
  }
};

describe("performanceController", () => {
  beforeEach(() => {
    frameNow = 0;
    frameId = 0;
    frameCallbacks.clear();

    vi.spyOn(window, "requestAnimationFrame").mockImplementation((callback) => {
      frameId += 1;
      frameCallbacks.set(frameId, callback);
      return frameId;
    });
    vi.spyOn(window, "cancelAnimationFrame").mockImplementation((id) => {
      frameCallbacks.delete(id);
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calcula renderPressure con la fórmula acordada", () => {
    expect(computeRenderPressure(77, 140, 110)).toBe(1237);
  });

  it("activa perf mode por presión estática alta durante interacción", () => {
    const { result } = renderHook(() =>
      usePerformanceController({
        featureFlags: defaultFlags,
        renderPressure: 1800,
        isInteracting: true,
      }),
    );

    runFrames(4, 16.7);

    expect(result.current.isPerformanceMode).toBe(true);
    expect(result.current.metrics.renderPressure).toBe(1800);
  });

  it("activa perf mode por frame-time/miss ratio alto aunque la presión sea baja", () => {
    const { result } = renderHook(() =>
      usePerformanceController({
        featureFlags: defaultFlags,
        renderPressure: 80,
        isInteracting: true,
      }),
    );

    runFrames(12, 38);

    expect(result.current.metrics.emaFrameMs).toBeGreaterThan(defaultFlags.onEmaMs || 18);
    expect(result.current.metrics.missRatio).toBeGreaterThan(defaultFlags.onMissRatio || 0.25);
    expect(result.current.isPerformanceMode).toBe(true);
  });

  it("desactiva con histeresis cuando baja carga y termina interacción", () => {
    const { result, rerender } = renderHook(
      ({
        renderPressure,
        isInteracting,
      }: {
        renderPressure: number;
        isInteracting: boolean;
      }) =>
        usePerformanceController({
          featureFlags: defaultFlags,
          renderPressure,
          isInteracting,
        }),
      {
        initialProps: {
          renderPressure: 1800,
          isInteracting: true,
        },
      },
    );

    runFrames(8, 30);
    expect(result.current.isPerformanceMode).toBe(true);

    rerender({
      renderPressure: 120,
      isInteracting: false,
    });
    runFrames(120, 8);

    expect(result.current.isPerformanceMode).toBe(false);
  });

  it("no renderiza por frame en idle (metrics desacopladas)", () => {
    let renderCount = 0;
    const Wrapper = ({ children }: { children: ReactNode }) => {
      renderCount += 1;
      return <>{children}</>;
    };

    renderHook(
      () =>
        usePerformanceController({
          featureFlags: defaultFlags,
          renderPressure: 120,
          isInteracting: false,
        }),
      { wrapper: Wrapper },
    );

    runFrames(180, 16.7);

    expect(renderCount).toBeLessThan(40);
  });

  it("en StrictMode no dispara Maximum update depth exceeded", () => {
    const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);
    const StrictWrapper = ({ children }: { children: ReactNode }) => (
      <StrictMode>{children}</StrictMode>
    );

    const { unmount } = renderHook(
      () =>
        usePerformanceController({
          featureFlags: defaultFlags,
          renderPressure: 600,
          isInteracting: true,
        }),
      { wrapper: StrictWrapper },
    );

    runFrames(240, 16.7);

    const depthErrors = consoleErrorSpy.mock.calls.filter((call) =>
      call.some((value) =>
        String(value).includes("Maximum update depth exceeded"),
      ),
    );
    expect(depthErrors).toHaveLength(0);

    unmount();
    consoleErrorSpy.mockRestore();
  });
});
