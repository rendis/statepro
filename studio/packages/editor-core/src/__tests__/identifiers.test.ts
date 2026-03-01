import { describe, expect, it } from "vitest";

import {
  cleanEventName,
  cleanIdentifier,
  ensureUniqueIdentifier,
  formatEventName,
  formatIdentifier,
} from "../utils";

describe("identifiers utils", () => {
  it("normaliza identifiers en kebab-case", () => {
    expect(formatIdentifier("Main Universe__V1")).toBe("main-universe-v1");
    expect(formatIdentifier("123 Main Universe__V1")).toBe("main-universe-v1");
    expect(cleanIdentifier("main---")).toBe("main");
    expect(cleanIdentifier("123---")).toBe("");
    expect(cleanIdentifier("__99Machine--")).toBe("machine");
  });

  it("normaliza eventos en SCREAMING_SNAKE_CASE", () => {
    expect(formatEventName("Start process!!")).toBe("START_PROCESS");
    expect(cleanEventName("EVENT__")).toBe("EVENT");
  });

  it("genera sufijos incrementales para ids duplicados", () => {
    const used = new Set<string>(["main-universe"]);
    const first = ensureUniqueIdentifier("main-universe-copy", used);
    used.add(first);
    const second = ensureUniqueIdentifier("main-universe-copy", used);
    used.add(second);
    const third = ensureUniqueIdentifier("main-universe-copy", used);

    expect(first).toBe("main-universe-copy");
    expect(second).toBe("main-universe-copy-2");
    expect(third).toBe("main-universe-copy-3");
  });
});
