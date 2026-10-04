import { configDefaults, defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "jsdom",
    globals: true,
    // Mutation sandboxes contain copies of the suite, sometimes instrumented.
    exclude: [...configDefaults.exclude, ".stryker-tmp/**"],
  },
});
