import { useEffect, useMemo, useRef, useState } from "react";

import type { StudioPerformanceFeatureFlags, StudioPerformanceMode } from "../types";

const FRAME_BUDGET_MS = 16.7;
const LONG_FRAME_MS = 50;
const EMA_ALPHA = 0.2;
const FRAME_WINDOW = 30;
const OFF_STABLE_FRAMES = 24;

export interface PerformanceControllerMetrics {
  emaFrameMs: number;
  missRatio: number;
  renderPressure: number;
  loafDetected: boolean;
}

export interface PerformanceControllerResult {
  isPerformanceMode: boolean;
  mode: StudioPerformanceMode;
  metrics: PerformanceControllerMetrics;
}

type PerformanceControllerOptions = {
  featureFlags: StudioPerformanceFeatureFlags | undefined;
  renderPressure: number;
  isInteracting: boolean;
};

const LOAF_COOLDOWN_MS = 2000;

const defaultFlags = {
  mode: "auto" as StudioPerformanceMode,
  staticPressureThreshold: 1200,
  onEmaMs: 18,
  offEmaMs: 14,
  onMissRatio: 0.25,
  offMissRatio: 0.1,
};

export const computeRenderPressure = (
  nodeCount: number,
  routeSegmentCount: number,
  transitionCount: number,
): number => {
  return nodeCount + routeSegmentCount * 2 + transitionCount * 8;
};

export const usePerformanceController = ({
  featureFlags,
  renderPressure,
  isInteracting,
}: PerformanceControllerOptions): PerformanceControllerResult => {
  const config = useMemo(
    () => ({
      mode: featureFlags?.mode ?? defaultFlags.mode,
      staticPressureThreshold:
        featureFlags?.staticPressureThreshold ?? defaultFlags.staticPressureThreshold,
      onEmaMs: featureFlags?.onEmaMs ?? defaultFlags.onEmaMs,
      offEmaMs: featureFlags?.offEmaMs ?? defaultFlags.offEmaMs,
      onMissRatio: featureFlags?.onMissRatio ?? defaultFlags.onMissRatio,
      offMissRatio: featureFlags?.offMissRatio ?? defaultFlags.offMissRatio,
    }),
    [featureFlags],
  );

  const [isPerformanceMode, setIsPerformanceMode] = useState(false);
  const metricsRef = useRef<PerformanceControllerMetrics>({
    emaFrameMs: FRAME_BUDGET_MS,
    missRatio: 0,
    renderPressure,
    loafDetected: false,
  });

  const configRef = useRef(config);
  const renderPressureRef = useRef(renderPressure);
  const isInteractingRef = useRef(isInteracting);
  const loopGenerationRef = useRef(0);
  const modeRef = useRef(false);
  const emaRef = useRef(FRAME_BUDGET_MS);
  const missWindowRef = useRef<boolean[]>([]);
  const lastFrameAtRef = useRef<number | null>(null);
  const stableFramesRef = useRef(0);
  const loafUntilTsRef = useRef(0);

  useEffect(() => {
    configRef.current = config;
    if (config.mode !== "off") {
      return;
    }

    stableFramesRef.current = 0;
    modeRef.current = false;
    setIsPerformanceMode((previous) => (previous ? false : previous));
  }, [config]);

  useEffect(() => {
    renderPressureRef.current = renderPressure;
    metricsRef.current.renderPressure = renderPressure;
  }, [renderPressure]);

  useEffect(() => {
    isInteractingRef.current = isInteracting;
  }, [isInteracting]);

  useEffect(() => {
    let rafId: number | null = null;
    const generation = loopGenerationRef.current + 1;
    loopGenerationRef.current = generation;

    const scheduleNext = () => {
      rafId = window.requestAnimationFrame(tick);
    };

    const applyMode = (nextMode: boolean) => {
      if (modeRef.current === nextMode) {
        return;
      }
      modeRef.current = nextMode;
      setIsPerformanceMode((previous) => (previous === nextMode ? previous : nextMode));
    };

    const tick = (ts: number) => {
      if (loopGenerationRef.current !== generation) {
        return;
      }

      const previous = lastFrameAtRef.current;
      lastFrameAtRef.current = ts;

      if (previous === null) {
        scheduleNext();
        return;
      }

      const deltaMs = Math.max(0, ts - previous);
      const currentConfig = configRef.current;
      const currentRenderPressure = renderPressureRef.current;
      const currentlyInteracting = isInteractingRef.current;
      const previousMode = modeRef.current;

      const nextEma = emaRef.current * (1 - EMA_ALPHA) + deltaMs * EMA_ALPHA;
      emaRef.current = nextEma;

      const nextMissWindow = missWindowRef.current;
      nextMissWindow.push(deltaMs > FRAME_BUDGET_MS);
      if (nextMissWindow.length > FRAME_WINDOW) {
        nextMissWindow.shift();
      }
      const misses = nextMissWindow.filter(Boolean).length;
      const nextMissRatio = nextMissWindow.length > 0 ? misses / nextMissWindow.length : 0;

      if (deltaMs > LONG_FRAME_MS) {
        loafUntilTsRef.current = Math.max(loafUntilTsRef.current, ts + LOAF_COOLDOWN_MS);
      }
      const loafDetected = ts < loafUntilTsRef.current;

      metricsRef.current.emaFrameMs = nextEma;
      metricsRef.current.missRatio = nextMissRatio;
      metricsRef.current.renderPressure = currentRenderPressure;
      metricsRef.current.loafDetected = loafDetected;

      if (currentConfig.mode === "off") {
        stableFramesRef.current = 0;
        applyMode(false);
        scheduleNext();
        return;
      }

      if (currentConfig.mode === "aggressive") {
        const aggressiveThreshold = currentConfig.staticPressureThreshold * 0.75;
        const shouldEnable =
          currentlyInteracting ||
          currentRenderPressure >= aggressiveThreshold ||
          nextEma > currentConfig.offEmaMs;
        stableFramesRef.current = 0;
        applyMode(shouldEnable);
        scheduleNext();
        return;
      }

      if (!previousMode) {
        const shouldEnable =
          currentlyInteracting &&
          (currentRenderPressure >= currentConfig.staticPressureThreshold ||
            nextEma > currentConfig.onEmaMs ||
            nextMissRatio > currentConfig.onMissRatio ||
            loafDetected);

        if (shouldEnable) {
          stableFramesRef.current = 0;
          applyMode(true);
        }
      } else {
        const belowOffThresholds =
          !currentlyInteracting &&
          nextEma < currentConfig.offEmaMs &&
          nextMissRatio < currentConfig.offMissRatio &&
          !loafDetected;

        if (belowOffThresholds) {
          stableFramesRef.current += 1;
        } else {
          stableFramesRef.current = 0;
        }

        if (stableFramesRef.current >= OFF_STABLE_FRAMES) {
          stableFramesRef.current = 0;
          applyMode(false);
        }
      }

      scheduleNext();
    };

    scheduleNext();

    return () => {
      if (loopGenerationRef.current === generation) {
        loopGenerationRef.current += 1;
      }
      if (rafId !== null) {
        window.cancelAnimationFrame(rafId);
      }
    };
  }, []);

  return {
    isPerformanceMode: config.mode === "off" ? false : isPerformanceMode,
    mode: config.mode,
    metrics: metricsRef.current,
  };
};
