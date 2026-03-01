import type { StudioEnMessageKey } from "./messages/en";

export type StudioLocale = "en" | "es";

export type StudioTranslationKey = StudioEnMessageKey;

export type StudioTranslationParams = Record<string, string | number>;

export type StudioMessageCatalog = Record<StudioTranslationKey, string>;

export type StudioTranslate = (
  key: StudioTranslationKey | (string & {}),
  params?: StudioTranslationParams,
  fallback?: string,
) => string;
