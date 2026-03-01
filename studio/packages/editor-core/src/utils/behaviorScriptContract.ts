import type { StudioTranslate } from "../i18n";
import type { BehaviorType } from "../types";

export interface BehaviorScriptContract {
  signature: string;
  defaultScript: string;
  hint: string;
}

const BEHAVIOR_SCRIPT_SIGNATURES: Record<BehaviorType, string> = {
  action: "function executeAction(args, context) {",
  invoke: "function executeInvoke(args, context, resolve, reject) {",
  condition: "function executeCondition(args, context) {",
  observer: "function executeObserver(args, context) {",
};

const BEHAVIOR_SCRIPT_HINT_KEYS: Record<BehaviorType, string> = {
  action: "library.behavior.scriptHint.action",
  invoke: "library.behavior.scriptHint.invoke",
  condition: "library.behavior.scriptHint.condition",
  observer: "library.behavior.scriptHint.observer",
};

const BEHAVIOR_SCRIPT_DEFAULT_KEYS: Record<BehaviorType, string> = {
  action: "library.behavior.defaultScript.action",
  invoke: "library.behavior.defaultScript.invoke",
  condition: "library.behavior.defaultScript.condition",
  observer: "library.behavior.defaultScript.observer",
};

const BEHAVIOR_SCRIPT_HINT_FALLBACKS: Record<BehaviorType, string> = {
  action: "Injected: args, context. Return value is ignored.",
  invoke: "Injected: args, context, resolve, reject. Call resolve() or reject(error).",
  condition: "Injected: args, context. Must return boolean.",
  observer: "Injected: args, context. Must return boolean.",
};

const BEHAVIOR_SCRIPT_DEFAULT_FALLBACKS: Record<BehaviorType, string> = {
  action: [
    "// Injected: args, context",
    'console.log("Running action...", args);',
    "// Return value is ignored for actions.",
  ].join("\n"),
  invoke: [
    "// Injected: args, context, resolve, reject",
    'console.log("Running invoke...", args);',
    "setTimeout(() => resolve(), 1000);",
  ].join("\n"),
  condition: [
    "// Injected: args, context",
    "// Must return true/false.",
    "return true;",
  ].join("\n"),
  observer: [
    "// Injected: args, context",
    "// Must return true/false.",
    "return true;",
  ].join("\n"),
};

export const getBehaviorScriptContract = (
  type: BehaviorType,
  t: StudioTranslate,
): BehaviorScriptContract => {
  return {
    signature: BEHAVIOR_SCRIPT_SIGNATURES[type],
    hint: t(
      BEHAVIOR_SCRIPT_HINT_KEYS[type],
      undefined,
      BEHAVIOR_SCRIPT_HINT_FALLBACKS[type],
    ),
    defaultScript: t(
      BEHAVIOR_SCRIPT_DEFAULT_KEYS[type],
      undefined,
      BEHAVIOR_SCRIPT_DEFAULT_FALLBACKS[type],
    ),
  };
};
