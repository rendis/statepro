# Studio Visual Editor Reference

## Table of Contents

1. [Monorepo Structure](#monorepo-structure)
2. [Setup & Commands](#setup--commands)
3. [React Integration](#react-integration)
4. [Web Component Integration](#web-component-integration)
5. [Data Contracts](#data-contracts)
6. [Feature Flags](#feature-flags)
7. [Tailwind CSS Setup](#tailwind-css-setup)
8. [i18n](#i18n)
9. [Builtin Behavior Catalog](#builtin-behavior-catalog)

## Monorepo Structure

```
studio/
├── app/                           # Standalone Vite dev shell (port 5173, private)
├── packages/
│   ├── editor-core/               # @rendis/statepro-studio-react (npm public)
│   └── web-component/             # @rendis/statepro-studio-web-component (npm public)
├── scripts/
│   ├── generate-builtin-behaviors.mjs  # YAML → TS catalog
│   └── dev-doctor.mjs                  # Environment diagnostics
├── package.json                   # Root scripts
├── pnpm-workspace.yaml
└── tsconfig.base.json
```

**Dev shell** (`studio/app`): Simple wrapper around `StateProEditor`. Uses `STUDIO_USE_EDITOR_CORE_SRC=true` to import editor-core from source (hot reload).

**Editor-core** (`@rendis/statepro-studio-react`): Full editor with canvas, modals, reducers, auto-layout (elkjs), i18n. Published to npm.

**Web component** (`@rendis/statepro-studio-web-component`): Custom element `<statepro-studio>` wrapping editor-core for framework-agnostic use.

## Setup & Commands

```bash
# Install
pnpm -C studio install

# Dev (source mode — editor-core from src/ via Vite alias)
pnpm -C studio dev

# Dev (dist mode — editor-core from node_modules dist)
pnpm -C studio dev:dist

# Build all packages
pnpm -C studio build

# Test all
pnpm -C studio test

# TypeScript check
pnpm -C studio typecheck

# Environment diagnostics
pnpm -C studio dev:doctor

# Editor-core specific
pnpm -C studio/packages/editor-core build
pnpm -C studio/packages/editor-core test
pnpm -C studio/packages/editor-core generate:builtin-catalog
```

**Environment variable**: `STUDIO_USE_EDITOR_CORE_SRC=true` (default in `pnpm dev`) makes Vite alias `@rendis/statepro-studio-react` to source `src/index.ts`.

**Test conventions**: Test descriptions are written in **Spanish**.

## React Integration

### Installation

```bash
npm install @rendis/statepro-studio-react
```

### Usage (Controlled)

```tsx
import { StateProEditor, type StudioChangePayload, type StudioExternalValue } from "@rendis/statepro-studio-react";
import "@rendis/statepro-studio-react/styles.css";

const [value, setValue] = useState<StudioExternalValue>({
  definition: machineJSON,
  layout: savedLayout,          // optional but recommended
  metadataPacks: {              // optional
    registry: packRegistry,
    bindings: packBindings,
  },
});

<StateProEditor
  value={value}
  onChange={(payload: StudioChangePayload) => {
    if (payload.source === "user") {  // skip external-sync echoes
      setValue({
        definition: payload.machine,
        layout: payload.layout,
        metadataPacks: {
          registry: payload.layout.packs.packRegistry,
          bindings: payload.layout.packs.bindings,
        },
      });
    }
  }}
  changeDebounceMs={250}           // default
  features={{ json: { import: true, export: true } }}
  locale="en"
/>
```

### Usage (Uncontrolled)

```tsx
<StateProEditor
  defaultValue={{ definition: machineJSON }}
  onChange={(payload) => saveToBackend(payload)}
/>
```

### Props

| Prop | Type | Description |
|---|---|---|
| `value` | `StudioExternalValue` | Controlled mode: editor state |
| `defaultValue` | `StudioExternalValue` | Uncontrolled mode: initial state |
| `onChange` | `(payload: StudioChangePayload) => void` | Debounced change callback |
| `changeDebounceMs` | `number` | Debounce delay (default: 250) |
| `universeTemplates` | `StudioUniverseTemplate[]` | Pre-built universe templates |
| `libraryBehaviors` | `BehaviorRegistryItem[]` | External behavior definitions |
| `features` | `StudioFeatureFlags` | Enable/disable features |
| `locale` | `"en" \| "es"` | Controlled locale |
| `defaultLocale` | `"en" \| "es"` | Fallback locale |
| `onLocaleChange` | `(locale) => void` | Locale change callback |
| `persistLocale` | `boolean` | Save locale to localStorage |
| `showLocaleSwitcher` | `boolean` | Show locale toggle |
| `logoSrc` | `string` | Custom logo URL |

## Web Component Integration

### Installation

```bash
npm install @rendis/statepro-studio-web-component @rendis/statepro-studio-react
```

### Usage

```ts
import { defineStateProStudioElement, STUDIO_CHANGE_EVENT, STUDIO_LOCALE_CHANGE_EVENT } from "@rendis/statepro-studio-web-component";
import "@rendis/statepro-studio-react/styles.css";  // styles still from editor-core

// Register <statepro-studio> custom element
defineStateProStudioElement();

const el = document.querySelector("statepro-studio");

// Set value (JS property, not attribute)
el.value = { definition: machineJSON };

// Listen for changes
el.addEventListener(STUDIO_CHANGE_EVENT, (e) => {
  console.log(e.detail); // StudioChangePayload
});
el.addEventListener(STUDIO_LOCALE_CHANGE_EVENT, (e) => {
  console.log(e.detail); // { locale: "en" | "es" }
});
```

### HTML Attributes

| Attribute | Type | Description |
|---|---|---|
| `locale` | string | Controlled locale |
| `default-locale` | string | Fallback locale |
| `show-locale-switcher` | boolean | Show locale toggle |
| `persist-locale` | boolean | Save to localStorage |
| `change-debounce-ms` | number | Debounce delay |

### JS Properties

`value`, `defaultValue`, `universeTemplates`, `libraryBehaviors`, `features`, `onChange`, `onLocaleChange` — same types as React props.

## Data Contracts

### StudioExternalValue (Input)

```typescript
{
  definition: StateProMachine;              // required — the machine JSON definition
  layout?: StudioLayoutDocument;            // optional — canvas visual state
  metadataPacks?: {
    registry: MetadataPackRegistry;
    bindings: MetadataPackBindingMap;
  };
}
```

### StudioChangePayload (Output)

```typescript
{
  machine: StateProMachine;                 // serialized definition
  layout: StudioLayoutDocument;             // visual state + pack snapshot
  issues: SerializeIssue[];                 // validation output
  canExport: boolean;                       // false if blocking errors exist
  source: "user" | "external-sync";         // who triggered the change
  at: string;                               // ISO timestamp
}
```

### StudioLayoutDocument

```typescript
{
  machineRef: { ... };                      // identity for compatibility check
  nodes: { ... };                           // visual snapshots per universe/reality
  transitions: { ... };                     // visual offsets per transition
  packs: {                                  // metadata pack snapshot
    packRegistry: MetadataPackRegistry;
    bindings: MetadataPackBindingMap;
  };
}
```

## Feature Flags

```typescript
interface StudioFeatureFlags {
  json?: {
    import?: boolean;   // enable JSON import modal
    export?: boolean;   // enable JSON export modal
  };
  library?: {
    behaviors?: { manage?: boolean; };       // behavior library management
    metadataPacks?: { create?: boolean; };   // metadata pack creation
  };
  performance?: {
    mode?: "auto" | "off" | "aggressive";   // default: "auto"
    staticPressureThreshold?: number;        // default: 1200
    onEmaMs?: number;                        // default: 18
    offEmaMs?: number;                       // default: 14
    onMissRatio?: number;                    // default: 0.25
    offMissRatio?: number;                   // default: 0.1
  };
}
```

**Performance modes**:
- `auto`: monitors frame time and miss ratio, culls offscreen elements when under pressure
- `off`: no culling (full render fidelity)
- `aggressive`: always cull offscreen, reduce overlays

## Tailwind CSS Setup

Studio hardcodes Tailwind default color palette classes (slate, blue, yellow, green, red, cyan, orange, purple, sky). Host apps **must** provide these.

### Tailwind v3

```ts
// tailwind.config.ts
export default {
  content: [
    "./src/**/*.{ts,tsx}",
    "./node_modules/@rendis/statepro-studio-react/dist/**/*.{js,mjs}",
    // OR for source mode:
    "../packages/editor-core/src/**/*.{ts,tsx}",
  ],
};
```

### Tailwind v4

```css
/* In CSS entry point */
@source "../node_modules/@rendis/statepro-studio-react/dist/**/*.js";

/* Preserve default palette with inline theme */
@theme inline {
  --color-primary: hsl(var(--primary));
}
```

### Host App CSS Interference

If host Tailwind config resets border styles (common with `@tailwind base`), wrap the editor in an isolation container:

```css
.studio-wrapper * {
  border-color: currentColor;
}
```

## i18n

Supported locales: `"en"` (English), `"es"` (Spanish).

**Persistence**: localStorage key `statepro.studio.locale`.

**Resolution order**: controlled `locale` prop → localStorage → `defaultLocale` → `"en"`.

Use `persistLocale={true}` and `showLocaleSwitcher={true}` to let users choose.

## Builtin Behavior Catalog

Auto-generated from `studio/packages/editor-core/model/builtin-behaviors.yaml`.

**Generation command**: `pnpm -C studio/packages/editor-core generate:builtin-catalog`

This runs automatically before `build` and `test` scripts. Output: `model/generatedBuiltinBehaviorCatalog.ts`.

Builtin behaviors are read-only in the editor. External behaviors are provided via the `libraryBehaviors` prop and are fully editable.
