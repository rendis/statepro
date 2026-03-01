# StatePro Studio

StatePro Studio is a visual editor for building and maintaining StatePro machine definitions.

It can be used in three main ways:

- As a local app for day-to-day authoring (`studio/app`).
- As an embeddable editor (`editor-core`) inside your own product.
- As a framework-agnostic custom element (`studio-web-component`) for React/Vue/other frameworks.

## What It Is

StatePro Studio provides a graph-based UI to edit:

- Machine definition JSON (`StateProMachine`).
- Visual layout data (`StudioLayoutDocument`).
- Metadata pack registry and bindings.

On every edit, Studio serializes and validates the model, then emits a change payload so host apps can persist continuously.

## Quick Start (Local Development)

```bash
cd studio
pnpm install
pnpm dev
```

Additional workspace checks:

```bash
cd studio
pnpm typecheck
pnpm test
pnpm build
```

## Studio Packages and Integration Modes

| Package                | Purpose                                                     | Typical Consumer                                     |
| ---------------------- | ----------------------------------------------------------- | ---------------------------------------------------- |
| `studio/app`           | Local development shell for Studio                          | Studio contributors                                  |
| `editor-core`          | React component package exposing `StateProEditor` and types | React applications                                   |
| `studio-web-component` | Custom element wrapper around `StateProEditor`              | Framework-agnostic embedding (Vue, vanilla JS, etc.) |

## Data Contracts (Definition, Layout, Metadata Packs)

### `StudioExternalValue`

`StateProEditor` receives external data through `value` or `defaultValue` using this shape:

```ts
interface StudioExternalValue {
  definition: StateProMachine;
  layout?: StudioLayoutDocument;
  metadataPacks?: {
    registry: MetadataPackRegistry;
    bindings: MetadataPackBindingMap;
  };
}
```

Contract notes:

- `definition` is required.
- `layout` is optional but recommended if you want deterministic node positions.
- `metadataPacks` is optional and can be provided externally.

### `StudioLayoutDocument`

`StudioLayoutDocument` stores visual state and metadata-pack snapshots.

Main sections:

- `machineRef`: machine identity for layout compatibility checks.
- `nodes`: visual snapshots for universes, realities, and global notes.
- `transitions`: visual offsets and notes per serialized transition reference.
- `packs`: `{ packRegistry, bindings }` snapshot associated with layout.

### `StudioChangePayload`

Studio notifies every change through:

```ts
interface StudioChangePayload {
  machine: StateProMachine;
  layout: StudioLayoutDocument;
  issues: SerializeIssue[];
  canExport: boolean;
  source: "user" | "external-sync";
  at: string; // ISO timestamp
}
```

Field meaning:

- `machine`: latest serialized model.
- `layout`: latest serialized layout + pack snapshot.
- `issues`: validation output (warnings/errors).
- `canExport`: `false` when blocking validation errors exist.
- `source`: whether change came from in-editor user actions or from controlled external synchronization.
- `at`: event timestamp.

## React Integration (editor-core)

### Install

```bash
pnpm add editor-core react react-dom
```

### Minimal Controlled Example (`value` + `onChange`)

```tsx
import { useState } from "react";
import {
  StateProEditor,
  type StudioChangePayload,
  type StudioExternalValue,
} from "editor-core";
import "editor-core/styles.css";

const initialValue: StudioExternalValue = {
  definition: {
    id: "machine-id",
    canonicalName: "machine",
    version: "1.0.0",
    universes: {},
  },
};

export function EmbeddedStudio() {
  const [value, setValue] = useState<StudioExternalValue>(initialValue);

  const handleChange = (payload: StudioChangePayload) => {
    if (payload.source === "external-sync") {
      return;
    }

    setValue({
      definition: payload.machine,
      layout: payload.layout,
      metadataPacks: {
        registry: payload.layout.packs.packRegistry,
        bindings: payload.layout.packs.bindings,
      },
    });
  };

  return <StateProEditor value={value} onChange={handleChange} />;
}
```

### Full Example (`universeTemplates`, `libraryBehaviors`, `features`, `locale`)

```tsx
import { useMemo, useState } from "react";
import {
  StateProEditor,
  type BehaviorRegistryItem,
  type StudioChangePayload,
  type StudioExternalValue,
  type StudioFeatureFlags,
  type StudioLocale,
  type StudioUniverseTemplate,
} from "editor-core";
import "editor-core/styles.css";

const templates: StudioUniverseTemplate[] = [
  {
    id: "support-template",
    label: "Support Flow",
    universe: {
      id: "support",
      canonicalName: "support",
      version: "1.0.0",
      realities: {
        waiting: { id: "waiting", type: "transition" },
      },
    },
  },
];

const libraryBehaviors: BehaviorRegistryItem[] = [
  {
    src: "builtin:action:logArgs",
    type: "action",
    description: "Logs args",
    simScript: "console.log(args);\\nreturn true;",
  },
];

const initialValue: StudioExternalValue = {
  definition: {
    id: "machine-id",
    canonicalName: "machine",
    version: "1.0.0",
    universes: {},
  },
};

export function FullEmbeddedStudio() {
  const [value, setValue] = useState(initialValue);
  const [locale, setLocale] = useState<StudioLocale>("en");

  const features = useMemo<StudioFeatureFlags>(
    () => ({
      json: { import: true, export: true },
      library: {
        behaviors: { manage: false },
        metadataPacks: { create: false },
      },
    }),
    [],
  );

  const handleChange = (payload: StudioChangePayload) => {
    if (payload.source === "external-sync") {
      return;
    }

    setValue({
      definition: payload.machine,
      layout: payload.layout,
      metadataPacks: {
        registry: payload.layout.packs.packRegistry,
        bindings: payload.layout.packs.bindings,
      },
    });
  };

  return (
    <StateProEditor
      value={value}
      onChange={handleChange}
      changeDebounceMs={250}
      universeTemplates={templates}
      libraryBehaviors={libraryBehaviors}
      features={features}
      locale={locale}
      onLocaleChange={setLocale}
      defaultLocale="en"
      persistLocale
      showLocaleSwitcher
    />
  );
}
```

### CSS + Tailwind Requirement

`editor-core/styles.css` contains Tailwind directives (`@tailwind base/components/utilities`).
Your host app must process Tailwind/PostCSS and include Studio sources in Tailwind content scanning.

Example `tailwind.config.ts`:

```ts
import type { Config } from "tailwindcss";

export default {
  content: [
    "./index.html",
    "./src/**/*.{ts,tsx}",
    "./node_modules/editor-core/**/*.{js,ts,tsx}",
    "../packages/editor-core/src/**/*.{ts,tsx}", // monorepo/workspace usage
  ],
  theme: { extend: {} },
  plugins: [],
} satisfies Config;
```

Use the path(s) that match your setup (`node_modules` or workspace source path).

## Web Component Integration (studio-web-component)

### Register and Use

```ts
import {
  defineStateProStudioElement,
  STUDIO_CHANGE_EVENT,
  STUDIO_LOCALE_EVENT,
} from "studio-web-component";
import type { StudioChangePayload, StudioExternalValue } from "editor-core";
import "editor-core/styles.css";

defineStateProStudioElement();

const el = document.querySelector("statepro-studio") as HTMLElement & {
  value?: StudioExternalValue;
  features?: unknown;
  onChange?: (payload: StudioChangePayload) => void;
};

el.value = {
  definition: {
    id: "machine-id",
    canonicalName: "machine",
    version: "1.0.0",
    universes: {},
  },
};

el.features = {
  json: { import: true, export: true },
  library: { behaviors: { manage: false }, metadataPacks: { create: false } },
};

el.addEventListener(STUDIO_CHANGE_EVENT, (event) => {
  const detail = (event as CustomEvent<StudioChangePayload>).detail;
  console.log("studio-change", detail);
});

el.addEventListener(STUDIO_LOCALE_EVENT, (event) => {
  const detail = (event as CustomEvent<{ locale: "en" | "es" }>).detail;
  console.log("studio-locale-change", detail.locale);
});
```

### Supported HTML Attributes

```html
<statepro-studio
  locale="en"
  default-locale="en"
  show-locale-switcher="true"
  persist-locale="true"
  change-debounce-ms="250"
></statepro-studio>
```

| Attribute              | Type                | Notes                                          |
| ---------------------- | ------------------- | ---------------------------------------------- | ----------------- |
| `locale`               | `"en"               | "es"`                                          | Controlled locale |
| `default-locale`       | `"en"               | "es"`                                          | Fallback locale   |
| `show-locale-switcher` | boolean-like string | Show/hide language toggle in header            |
| `persist-locale`       | boolean-like string | Enable/disable localStorage locale persistence |
| `change-debounce-ms`   | number-like string  | Debounce for emitted change payloads           |

### Supported JS Properties

| Property            | Type                                       |
| ------------------- | ------------------------------------------ |
| `value`             | `StateProEditorProps["value"]`             |
| `defaultValue`      | `StateProEditorProps["defaultValue"]`      |
| `universeTemplates` | `StateProEditorProps["universeTemplates"]` |
| `libraryBehaviors`  | `StateProEditorProps["libraryBehaviors"]`  |
| `features`          | `StateProEditorProps["features"]`          |
| `onChange`          | `StateProEditorProps["onChange"]`          |
| `onLocaleChange`    | `StateProEditorProps["onLocaleChange"]`    |

### Emitted Events

| Event                  | Detail                     |
| ---------------------- | -------------------------- |
| `studio-change`        | `StudioChangePayload`      |
| `studio-locale-change` | `{ locale: StudioLocale }` |

## Vue Wrapper Example (using the custom element)

```vue
<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from "vue";
import {
  defineStateProStudioElement,
  STUDIO_CHANGE_EVENT,
  STUDIO_LOCALE_EVENT,
} from "studio-web-component";
import type { StudioChangePayload, StudioExternalValue } from "editor-core";
import "editor-core/styles.css";

defineStateProStudioElement();

const studioEl = ref<HTMLElement | null>(null);

const initialValue: StudioExternalValue = {
  definition: {
    id: "machine-id",
    canonicalName: "machine",
    version: "1.0.0",
    universes: {},
  },
};

const handleChange = (event: Event) => {
  const payload = (event as CustomEvent<StudioChangePayload>).detail;
  console.log("studio-change", payload);
};

const handleLocale = (event: Event) => {
  const detail = (event as CustomEvent<{ locale: "en" | "es" }>).detail;
  console.log("studio-locale-change", detail.locale);
};

onMounted(() => {
  const el = studioEl.value as
    | (HTMLElement & {
        value?: StudioExternalValue;
        features?: unknown;
      })
    | null;

  if (!el) return;

  el.value = initialValue;
  el.features = {
    json: { import: true, export: true },
    library: {
      behaviors: { manage: false },
      metadataPacks: { create: false },
    },
  };

  el.addEventListener(STUDIO_CHANGE_EVENT, handleChange as EventListener);
  el.addEventListener(STUDIO_LOCALE_EVENT, handleLocale as EventListener);
});

onBeforeUnmount(() => {
  const el = studioEl.value;
  if (!el) return;

  el.removeEventListener(STUDIO_CHANGE_EVENT, handleChange as EventListener);
  el.removeEventListener(STUDIO_LOCALE_EVENT, handleLocale as EventListener);
});
</script>

<template>
  <statepro-studio
    ref="studioEl"
    style="display: block; width: 100%; height: 100vh"
  />
</template>
```

If your Vue build complains about unknown custom elements, configure compiler custom element handling for `statepro-studio`.

## Props Reference (StateProEditorProps)

| Prop                 | Type                                     | Default                    | Behavior                                                 |
| -------------------- | ---------------------------------------- | -------------------------- | -------------------------------------------------------- | ------------------------ |
| `value`              | `StudioExternalValue`                    | `undefined`                | Controlled mode source of truth.                         |
| `defaultValue`       | `StudioExternalValue`                    | `undefined`                | Initial value for uncontrolled mode.                     |
| `onChange`           | `(payload: StudioChangePayload) => void` | `undefined`                | Called after debounced serialization/validation updates. |
| `changeDebounceMs`   | `number`                                 | `250`                      | Debounce window for `onChange`.                          |
| `universeTemplates`  | `StudioUniverseTemplate[]`               | `[]`                       | Enables template-based universe creation in toolbar.     |
| `libraryBehaviors`   | `BehaviorRegistryItem[]`                 | `undefined`                | External behavior registry input.                        |
| `features`           | `StudioFeatureFlags`                     | all supported flags `true` | Enables/disables JSON/library capabilities.              |
| `locale`             | `"en"                                    | "es"`                      | `undefined`                                              | Controlled locale value. |
| `defaultLocale`      | `"en"                                    | "es"`                      | `"en"`                                                   | Initial/fallback locale. |
| `onLocaleChange`     | `(locale: StudioLocale) => void`         | `undefined`                | Notified when locale changes in UI/provider.             |
| `persistLocale`      | `boolean`                                | `true`                     | Writes/reads locale from localStorage.                   |
| `showLocaleSwitcher` | `boolean`                                | `true`                     | Shows/hides header language toggle button.               |

## Feature Flags (Enable/Disable Capabilities)

```ts
features?: {
  json?: {
    import?: boolean;
    export?: boolean;
  };
  library?: {
    behaviors?: {
      manage?: boolean;
    };
    metadataPacks?: {
      create?: boolean;
    };
  };
}
```

| Flag                           | `true`                                                   | `false`                                 |
| ------------------------------ | -------------------------------------------------------- | --------------------------------------- |
| `json.import`                  | Enables model/layout import tab/actions in JSON modal    | Import actions disabled/hidden          |
| `json.export`                  | Enables model/layout export tab/actions in JSON modal    | Export actions disabled/hidden          |
| `library.behaviors.manage`     | Allows creating/editing/deleting behavior registry items | Behavior management UI/actions disabled |
| `library.metadataPacks.create` | Allows creating new metadata packs                       | New pack creation is disabled           |

Important nuance:

- `library.metadataPacks.create = false` disables creating new packs.
- Existing metadata packs can still be opened/edited by current implementation.

## Change Notifications and Persistence Pattern

`onChange` is emitted with debounce (`changeDebounceMs`, default `250`).

Source semantics:

- `source: "user"`: user-driven edits inside Studio (graph, properties, library, import, etc.).
- `source: "external-sync"`: controlled re-synchronization after parent updates `value` (or external behavior registry sync in uncontrolled behavior-list updates).

Recommended controlled persistence pattern:

```ts
const handleChange = (payload: StudioChangePayload) => {
  if (payload.source === "external-sync") {
    return; // prevents persistence loops
  }

  saveToServer(payload.machine, payload.layout);
  setValue({
    definition: payload.machine,
    layout: payload.layout,
    metadataPacks: {
      registry: payload.layout.packs.packRegistry,
      bindings: payload.layout.packs.bindings,
    },
  });
};
```

## i18n and Locale Behavior

Supported locales:

- `en`
- `es`

Locale-related props:

- `locale`: controlled locale.
- `defaultLocale`: fallback locale.
- `onLocaleChange`: callback on locale update.
- `persistLocale`: enables storage persistence.
- `showLocaleSwitcher`: toggles language switch UI.

Persistence behavior:

- When `persistLocale` is enabled, Studio stores locale in `localStorage` using key `statepro.studio.locale`.
- On startup, if `locale` is not controlled, Studio resolves initial locale from storage (when available) and falls back to `defaultLocale`.

## Controlled vs Uncontrolled Usage

| Mode         | How to Use                                      | Notes                                           |
| ------------ | ----------------------------------------------- | ----------------------------------------------- |
| Controlled   | Provide `value` and update it from `onChange`   | Parent owns full source of truth                |
| Uncontrolled | Omit `value`, optionally provide `defaultValue` | Studio owns internal state after initialization |

Guidelines:

- Use controlled mode if you need autosave, collaborative sync, or authoritative external persistence.
- Use uncontrolled mode for simpler embedded editing where parent does not need full state synchronization on every change.

## Validation, Import/Export, and Issues

Validation model:

- Studio serializes editor state to `StateProMachine` continuously.
- Validation output is exposed in `onChange.payload.issues`.
- `canExport` indicates whether blocking errors exist.

Import/Export behavior:

- JSON model import validates machine schema/semantics before applying.
- Layout import validates layout structure before applying.
- Feature flags (`json.import` / `json.export`) can hide/disable tabs and actions.

Issue interpretation:

- `severity: "error"`: blocks export/import actions that require valid payloads.
- `severity: "warning"`: non-blocking but should be reviewed.

## Troubleshooting

### 1) Studio appears unstyled or partially styled

Cause:

- `editor-core/styles.css` not imported.
- Tailwind pipeline not active in host app.
- Tailwind content globs missing Studio classes.

Fix:

- Import `editor-core/styles.css`.
- Ensure PostCSS + Tailwind are configured.
- Add correct `content` globs (`node_modules/editor-core` or workspace source paths).

### 2) Controlled mode creates update loops

Cause:

- Parent writes back `value` for every callback, including `source: "external-sync"` payloads.

Fix:

- Ignore `external-sync` updates in persistence and state-rewrite logic.

### 3) Import fails even with valid-looking JSON

Cause:

- Schema or semantic validation errors in model/layout.

Fix:

- Inspect `issues` details (`field`, `message`, `severity`).
- Resolve all `error` entries first.
- Re-try import after fixing blocking fields.

## Build/Test Commands

```bash
cd studio
pnpm typecheck
pnpm test
pnpm build
```
