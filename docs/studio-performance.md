# Studio layout and history performance

## Worker-backed layout

The standalone app runs ELK's layered layout in a native Web Worker. It imports the
worker as a Vite asset, so development, production hashes, and base paths resolve
through the bundler. The UI thread imports only the small ELK API when using this path.
Initial layout, manual layout, and JSON import use the same worker configuration.

React consumers can opt in without changing the existing default deployment contract:

```tsx
import { StateProEditor } from "@rendis/statepro-studio-react";
import layoutWorkerUrl from "elkjs/lib/elk-worker.min.js?url";

export function Editor() {
  return <StateProEditor autoLayoutWorkerUrl={layoutWorkerUrl} />;
}
```

Install `elkjs` as a direct dependency of the host application when importing its asset.
For other bundlers, serve `elkjs/lib/elk-worker.min.js` from the host application's
origin and pass that URL. Use the same ELK version as editor-core. The worker is an
application asset, not part of an imported machine definition; use a trusted URL.
Deployments with Content Security Policy must permit it in `worker-src` and serve
JavaScript with the correct MIME type. This is a classic worker, not a module worker.

Web Component consumers can set the corresponding property or attribute:

```html
<statepro-studio auto-layout-worker-url="/assets/elk-worker.min.js"></statepro-studio>
```

```js
element.autoLayoutWorkerUrl = "/assets/elk-worker.min.js";
```

Without a worker URL, editor-core retains its lazily imported bundled ELK implementation.
Both paths preserve the layered layout algorithm and graph anchoring. One worker per URL is
reused across layout requests, including concurrent ones (ELK tags each request), so the
engine is parsed and initialized once instead of on every layout. It is released after
60 seconds without layout requests and recreated on demand.

If the worker cannot start or load (missing asset, `worker-src` CSP, a cross-origin URL,
or no `Worker` global), the request falls back to the bundled engine on the main thread
and logs a warning; that URL is not retried for the rest of the page session. A worker that
crashes after completing layouts is discarded, its in-flight requests fall back, and the next
request starts a new worker. A request that does not finish within 30 seconds rejects and
discards the worker without falling back, since rerunning a stuck graph on the main thread
would freeze the UI. ELK errors for an invalid graph are returned as-is.

Moving ELK off the UI thread does **not** reduce its download. The worker asset is about
1.61 MB before HTTP compression (about 467 kB gzip in the local build); the API chunk is
about 4.8 kB (2 kB gzip). The compatibility
bundle remains in the build output but is not requested by the standalone app's worker path.
Changing the generated engine or replacing the layout algorithm is a separate size optimization.

## Checkpoints and incremental history

History stores up to 100 detached snapshots. Text edits within the coalescing window already
share one undo step. Capture, recording, and undo/redo keep defensive copies because public
editor state and supplied checkpoints are mutable. Restoration now copies the snapshot in
one operation while retaining that isolation. Optional imported-model fields are compared in
both snapshots and cleared when absent from the restored checkpoint.

Full copies still scale with graph size. A patch-based history could reduce storage for sparse
edits, but sharing mutable sections between public snapshots would corrupt unrelated checkpoints.
A compatible incremental implementation needs private immutable storage plus detached public
views, or reversible owned patches whose inserted/deleted values are cloned at admission.
It must also preserve grouped drag operations, branching after undo, import resets, selection
validation, history limits, metadata, and externally supplied checkpoints. Restoring mutable
public editor state will still require materializing an isolated graph.

Do not remove defensive copies or introduce object-identity caches for caller-owned mutable data.
The current change does not claim to implement incremental history. Benchmark sparse edits and
memory retention before choosing that larger representation change.

Reproduce timings with the same build, runtime, fixture, and idle host:

```bash
pnpm -C studio --filter @rendis/statepro-studio-react build
node studio/scripts/benchmark-history.mjs
# Optional absolute path to a baseline CommonJS build:
STATEPRO_BENCHMARK_ENTRY=/path/to/baseline/index.cjs node studio/scripts/benchmark-history.mjs
```

The fixture uses one universe, 100/1,000/5,000 realities, a 256-byte metadata string per
reality, and no edges. It discards one warm-up, then reports the median and maximum of five
samples of 20 operations. Undo/redo measures 20 pairs. These are Node timings, not browser
frame measurements; avoid comparing runs made while test or mutation campaigns compete for CPU.
