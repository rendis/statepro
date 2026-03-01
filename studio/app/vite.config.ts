import path from "node:path";
import { fileURLToPath } from "node:url";

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const appRoot = path.dirname(fileURLToPath(import.meta.url));
const workspaceRoot = path.resolve(appRoot, "..");
const editorCoreRoot = path.resolve(workspaceRoot, "packages/editor-core");
const editorCoreSrcEntry = path.resolve(editorCoreRoot, "src/index.ts");
const editorCoreSrcStyles = path.resolve(editorCoreRoot, "src/styles.css");

const isSourceAliasEnabled = (mode: string): boolean => {
  if (mode !== "development") {
    return false;
  }

  const raw = String(process.env.STUDIO_USE_EDITOR_CORE_SRC ?? "true").toLowerCase();
  return !["0", "false", "off", "no"].includes(raw);
};

export default defineConfig(({ mode }) => {
  const useSourceAlias = isSourceAliasEnabled(mode);

  return {
    plugins: [react()],
    resolve: {
      alias: useSourceAlias
        ? [
            {
              find: "@rendis/statepro-studio-react/styles.css",
              replacement: editorCoreSrcStyles,
            },
            {
              find: "@rendis/statepro-studio-react",
              replacement: editorCoreSrcEntry,
            },
          ]
        : [],
    },
    server: {
      port: 5173,
      fs: {
        allow: [workspaceRoot, editorCoreRoot],
      },
    },
    optimizeDeps: {
      exclude: [
        "@rendis/statepro-studio-react",
        "elkjs",
        "elkjs/lib/elk.bundled.js",
      ],
    },
  };
});
