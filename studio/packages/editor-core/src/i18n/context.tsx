import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";

import {
  STUDIO_DEFAULT_LOCALE,
  STUDIO_LOCALE_STORAGE_KEY,
  createStudioTranslator,
  normalizeStudioLocale,
} from "./translate";
import type {
  StudioLocale,
  StudioTranslate,
  StudioTranslationKey,
  StudioTranslationParams,
} from "./types";

export interface StudioI18nContextValue {
  locale: StudioLocale;
  setLocale: (locale: StudioLocale) => void;
  t: StudioTranslate;
}

export interface StudioI18nProviderProps {
  children: ReactNode;
  locale?: StudioLocale;
  defaultLocale?: StudioLocale;
  onLocaleChange?: (locale: StudioLocale) => void;
  persistLocale?: boolean;
}

const readStoredLocale = (): StudioLocale | null => {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    return normalizeStudioLocale(window.localStorage.getItem(STUDIO_LOCALE_STORAGE_KEY));
  } catch {
    return null;
  }
};

const writeStoredLocale = (locale: StudioLocale): void => {
  if (typeof window === "undefined") {
    return;
  }

  try {
    window.localStorage.setItem(STUDIO_LOCALE_STORAGE_KEY, locale);
  } catch {
    // noop: do not crash if storage is unavailable.
  }
};

export interface ResolveInitialLocaleOptions {
  locale?: StudioLocale;
  defaultLocale?: StudioLocale;
  persistLocale?: boolean;
}

export const resolveInitialStudioLocale = ({
  locale,
  defaultLocale = STUDIO_DEFAULT_LOCALE,
  persistLocale = true,
}: ResolveInitialLocaleOptions): StudioLocale => {
  if (locale) {
    return normalizeStudioLocale(locale);
  }

  if (persistLocale) {
    const stored = readStoredLocale();
    if (stored) {
      return stored;
    }
  }

  return normalizeStudioLocale(defaultLocale);
};

const DEFAULT_CONTEXT: StudioI18nContextValue = {
  locale: STUDIO_DEFAULT_LOCALE,
  setLocale: () => undefined,
  t: (key, params, fallback) => createStudioTranslator(STUDIO_DEFAULT_LOCALE)(key, params, fallback),
};

const StudioI18nContext = createContext<StudioI18nContextValue>(DEFAULT_CONTEXT);

export const StudioI18nProvider = ({
  children,
  locale: controlledLocale,
  defaultLocale = STUDIO_DEFAULT_LOCALE,
  onLocaleChange,
  persistLocale = true,
}: StudioI18nProviderProps) => {
  const [internalLocale, setInternalLocale] = useState<StudioLocale>(() =>
    resolveInitialStudioLocale({
      locale: controlledLocale,
      defaultLocale,
      persistLocale,
    }),
  );

  const isControlled = controlledLocale !== undefined;
  const locale = normalizeStudioLocale(isControlled ? controlledLocale : internalLocale);

  useEffect(() => {
    if (!persistLocale) {
      return;
    }
    writeStoredLocale(locale);
  }, [locale, persistLocale]);

  const setLocale = useCallback(
    (nextLocale: StudioLocale) => {
      const normalized = normalizeStudioLocale(nextLocale);
      if (!isControlled) {
        setInternalLocale(normalized);
      }
      onLocaleChange?.(normalized);
    },
    [isControlled, onLocaleChange],
  );

  const t = useMemo<StudioTranslate>(() => createStudioTranslator(locale), [locale]);

  const value = useMemo(
    () => ({
      locale,
      setLocale,
      t,
    }),
    [locale, setLocale, t],
  );

  return <StudioI18nContext.Provider value={value}>{children}</StudioI18nContext.Provider>;
};

export const useI18n = (): StudioI18nContextValue => {
  return useContext(StudioI18nContext);
};

export const tLabel = (
  t: StudioTranslate,
  key: StudioTranslationKey,
  params?: StudioTranslationParams,
  fallback?: string,
): string => t(key, params, fallback);
