import { defineConfig } from "vitest/config";

/** Narrow Vitest config for Stryker — avoids full editor suite hangs under mutation. */
export default defineConfig({
  test: {
    environment: "jsdom",
    globals: true,
    include: [
      "src/__tests__/validateStatePro.test.ts",
      "src/__tests__/identifiers.test.ts",
      "src/__tests__/transitionRules.test.ts",
      "src/__tests__/issueMapping.test.ts",
    ],
    pool: "forks",
    poolOptions: {
      forks: {
        singleFork: true,
      },
    },
    fileParallelism: false,
  },
});
