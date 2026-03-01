import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const studioRoot = path.resolve(scriptDir, "..");
const appRoot = path.resolve(studioRoot, "app");
const editorCoreRoot = path.resolve(studioRoot, "packages/editor-core");
const editorCoreSrcEntry = path.resolve(editorCoreRoot, "src/index.ts");
const editorCoreDistEntry = path.resolve(editorCoreRoot, "dist/index.js");
const packageSymlink = path.resolve(
  appRoot,
  "node_modules/@rendis/statepro-studio-react",
);

const mode = String(process.env.MODE || "development").toLowerCase();
const sourceFlagRaw = String(process.env.STUDIO_USE_EDITOR_CORE_SRC ?? "true").toLowerCase();
const sourceAliasEnabled =
  mode === "development" && !["0", "false", "off", "no"].includes(sourceFlagRaw);

let symlinkTarget = "(missing)";
let symlinkResolved = "(missing)";

if (fs.existsSync(packageSymlink)) {
  try {
    symlinkTarget = fs.readlinkSync(packageSymlink);
  } catch {
    symlinkTarget = "(not-a-symlink)";
  }
  try {
    symlinkResolved = fs.realpathSync(packageSymlink);
  } catch {
    symlinkResolved = "(unresolved)";
  }
}

const effectiveEntry = sourceAliasEnabled ? editorCoreSrcEntry : editorCoreDistEntry;
const effectiveKind = sourceAliasEnabled ? "src" : "dist";

console.log("Studio Dev Doctor");
console.log(`- mode: ${mode}`);
console.log(`- STUDIO_USE_EDITOR_CORE_SRC: ${sourceFlagRaw}`);
console.log(`- source alias enabled: ${sourceAliasEnabled}`);
console.log(`- app root: ${appRoot}`);
console.log(`- workspace package symlink: ${packageSymlink}`);
console.log(`- symlink target: ${symlinkTarget}`);
console.log(`- symlink resolved: ${symlinkResolved}`);
console.log(`- effective editor entry (${effectiveKind}): ${effectiveEntry}`);
console.log(`- src exists: ${fs.existsSync(editorCoreSrcEntry)}`);
console.log(`- dist exists: ${fs.existsSync(editorCoreDistEntry)}`);
