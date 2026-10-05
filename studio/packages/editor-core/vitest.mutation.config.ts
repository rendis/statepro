import { defineConfig } from "vitest/config";

/** Narrow Vitest config for Stryker — avoids full editor suite hangs under mutation. */
export default defineConfig({
  test: {
    environment: "jsdom",
    globals: true,
    include: [
      "src/__tests__/validateStatePro.test.ts",
      "src/__tests__/validateStatePro.inputs.test.ts",
      "src/__tests__/validateStatePro.diagnostics.test.ts",
      "src/__tests__/identifiers.test.ts",
      "src/__tests__/transitionRules.test.ts",
      "src/__tests__/issueMapping.test.ts",
      "src/__tests__/serializeStatePro.test.ts",
      "src/__tests__/deserializeStatePro.conditions.test.ts",
      "src/__tests__/review.regression.test.ts",
    ],
    pool: "forks",
    maxWorkers: 1,
    fileParallelism: false,
  },
});
