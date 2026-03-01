import type { BehaviorRef, JsonObject } from "../types";

export const cleanBehaviorRef = (behavior: BehaviorRef): BehaviorRef => {
  const cleaned: BehaviorRef = { src: behavior.src };

  if (behavior.args && Object.keys(behavior.args).length > 0) {
    cleaned.args = behavior.args;
  }

  if (behavior.metadata && Object.keys(behavior.metadata).length > 0) {
    cleaned.metadata = behavior.metadata;
  }

  return cleaned;
};

export const cleanBehaviorRefs = (behaviors: BehaviorRef[] = []): BehaviorRef[] => {
  return behaviors.map(cleanBehaviorRef);
};

interface ParseJsonObjectError {
  message: string;
  messageKey: string;
  messageParams?: Record<string, string | number>;
}

interface ParseJsonObjectResult {
  value: JsonObject;
  ok: boolean;
  error?: ParseJsonObjectError;
}

export const tryParseJsonObject = (
  raw: string | undefined,
): ParseJsonObjectResult => {
  if (!raw || !raw.trim()) {
    return { value: {}, ok: true };
  }

  try {
    const parsed = JSON.parse(raw) as unknown;
    if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
      return { value: parsed as JsonObject, ok: true };
    }
    return {
      value: {},
      ok: false,
      error: {
        message: "Metadata must be a JSON object",
        messageKey: "issue.invalidJsonObject",
      },
    };
  } catch {
    return {
      value: {},
      ok: false,
      error: {
        message: "Invalid JSON",
        messageKey: "issue.invalidJson",
      },
    };
  }
};
