import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: [{
      find: /^@rendis\/statepro-studio-react$/,
      replacement: fileURLToPath(new URL("../editor-core/src/index.ts", import.meta.url)),
    }],
  },
  test: { environment: "jsdom" },
});
