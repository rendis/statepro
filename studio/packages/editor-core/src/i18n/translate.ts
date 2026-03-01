import type { SerializeIssue } from "../types";
import { STUDIO_I18N_EN } from "./messages/en";
import { STUDIO_I18N_ES } from "./messages/es";
import type { StudioLocale, StudioTranslationKey, StudioTranslationParams, StudioTranslate } from "./types";

export const STUDIO_LOCALE_STORAGE_KEY = "statepro.studio.locale";

export const STUDIO_DEFAULT_LOCALE: StudioLocale = "en";

export const isStudioLocale = (value: unknown): value is StudioLocale =>
  value === "en" || value === "es";

export const normalizeStudioLocale = (value: unknown): StudioLocale =>
  isStudioLocale(value) ? value : STUDIO_DEFAULT_LOCALE;

const STUDIO_I18N_BUNDLES: Record<StudioLocale, Record<string, string>> = {
  en: STUDIO_I18N_EN,
  es: STUDIO_I18N_ES,
};

const interpolateTemplate = (template: string, params?: StudioTranslationParams): string => {
  if (!params) {
    return template;
  }

  return Object.entries(params).reduce((acc, [key, value]) => {
    return acc.replaceAll(`{{${key}}}`, String(value));
  }, template);
};

export const translateStudioMessage = (
  locale: StudioLocale,
  key: StudioTranslationKey | (string & {}),
  params?: StudioTranslationParams,
  fallback?: string,
): string => {
  const currentBundle = STUDIO_I18N_BUNDLES[locale];
  const fallbackBundle = STUDIO_I18N_BUNDLES[STUDIO_DEFAULT_LOCALE];

  const template = currentBundle[key] || fallbackBundle[key] || fallback || key;
  return interpolateTemplate(template, params);
};

export const createStudioTranslator = (locale: StudioLocale): StudioTranslate => {
  return (key, params, fallback) => translateStudioMessage(locale, key, params, fallback);
};

export const resolveSerializeIssueMessage = (issue: SerializeIssue, t: StudioTranslate): string => {
  if (issue.messageKey) {
    return t(issue.messageKey as StudioTranslationKey, issue.messageParams, issue.message);
  }
  return issue.message;
};
