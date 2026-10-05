# Snapshot migration, validation coverage, and worker layout

Reviewed against `main` at `fb3d56d6da129cbebdefa8a940c21beda16be6ea`.

## Changes

- Add contextual snapshot helpers that select the optional contextual capability, then checked
  capture, then a compatible legacy fallback. Preserve provider errors and reject nil captures.
  Deprecate context-free owner methods without changing mandatory interfaces. Migrate debugger
  internals, the Go example, and consumer documentation. Checked action arguments remain the safe
  capture path inside synchronous callbacks; reentry tests cover the new helpers.
- The CLI now reports capture/restoration failures in its view and does not create checkpoints
  from failed captures. A successfully handled event remains marked as sent even if capture fails.
  Reject null records in event/snapshot files before dereferencing them. Test actual container
  initialization, malformed files, checkpoint loading, failed restoration, and history rollback.
- Add 26 validator diagnostic cases covering precise fields, messages, localization parameters,
  JSON Pointer decoding, seed deduplication, and warning/error export behavior. Include them in
  the configured Stryker campaign. Add a reproducible CLI mutation target.
- Run the standalone app's initial/manual/import layout in a native worker. React and Web Component
  consumers can opt in using a worker asset URL. Use only the lightweight ELK API on the UI thread,
  terminate each owned worker on completion/error/timeout, and retain the existing bundled fallback.
- Compare optional checkpoint fields in both snapshots so an omitted imported model is cleared
  on restoration. Consolidate restoration into one detached clone and retain explicit restored
  fields and selection validation. Test undo/redo isolation and the omitted-field regression.
- Provide a compiled runtime-policy example with registered observers, callback-safe capture,
  cooperative invokes, illustrative finite budgets, and bounded shutdown. Application owners
  still choose budgets and enable strict policies for their workloads.

## Validation

- Go 1.25: build, vet, and race-enabled tests with coverage across all 11 packages. CLI coverage
  increased from 51.0% in the previous review to 62.3%. The compiled example and helper/reentry
  regressions pass. Three native fuzz smoke targets ran for five seconds each.
- Node 24.19 / pnpm 10.11: frozen installation, lint, typecheck, builds, and dependency audit
  including development dependencies. Audit found zero advisories. Editor-core: 313 tests in
  49 files; Web Component: 3 tests. No new dependency version was introduced; the app declares
  its existing `elkjs` version directly to import the worker asset.
- One initial concurrent suite run hit a five-second timeout in an existing visual-filter test.
  That file passed independently, and the complete suite passed with two Vitest workers.
  A failing new checkpoint test exposed the omitted-field bug; the corrected regression and
  complete suite passed. The final restoration adjustment was checked with the history,
  editor history, and auto-layout integration suites.
- Chromium production app at `http://127.0.0.1:5188/`, 390×900 and 1440×900: initial layout,
  manual layout, adding a universe, undo/redo, and JSON import passed. Each layout used a native
  worker and terminated it. Network inspection confirmed no `elk.bundled` request. Page identity,
  meaningful canvas content, absence of framework overlays, clean console, and no horizontal
  page overflow passed; screenshots were captured outside the repository.
- A real-browser worker-download failure rejected promptly and released the layout control.
  Removing the failing route and retrying succeeded. Unit tests additionally cover message errors,
  blocked-worker timeouts, constructor failures, concurrent requests, and cleanup.

## Completed mutation campaigns

| Target | Killed | Survived | Uncovered | Timed out | Score / efficacy | Covered score / mutant coverage |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Stryker: three configured Studio sources | 721 | 97 | 25 | 2 | 85.56% | 88.17% |
| Gremlins: CLI | 69 | 0 | 60 | 0 | 100.00% | 53.49% |
| Gremlins: instrumentation | 15 | 0 | 8 | 0 | 100.00% | 65.22% |

Stryker executed the 85-test mutation suite against the same 845 source mutants as the previous
review, with no runner errors. Overall score increased from 58.11% to 85.56%; the validator
increased from 45.65% to 86.86%. Identifiers and transition-rule scores remained 82.04% and 84.35%.
Gremlins efficacy excludes timeouts; Stryker counts them as detected. The CLI campaign also
passed the configured efficacy and coverage gates. Instrumentation was measured separately.

Remaining validator survivors include defensive AJV branches, schema keywords not emitted by
the current schema, and semantic guard differentiation on malformed input. They have not all
been classified as equivalent. CLI uncovered mutants include terminal UI and process-exit paths.
These scoped campaigns are not exhaustive repository mutation coverage or proof of no defects.

## Performance and compatibility limits

Worker execution preserves ELK's layered algorithm and moves engine work off the UI thread.
The worker is approximately 1.61 MB raw / 467 kB gzip in the local build; its API is about
4.8 kB raw / 2 kB gzip. This reduces UI-thread engine parsing/execution, not download size.
The compatibility bundle remains available. Workers require a correctly served trusted asset
and an appropriate `worker-src` CSP. Individual requests have a 30-second ceiling.

Checkpoint storage still makes detached copies and scales with graph size. The benchmark accepts
a baseline entry path for reproducible comparisons. Runs competing with mutation/test processes
were noisy and do not establish a speedup from clone consolidation. An incremental representation
needs owned patches or private immutable storage plus detached public views; it is not implemented
in this patch. See [Studio performance](../studio-performance.md) for the contract and design criteria.

A subsequent local run after mutation/test processes finished measured 20 undo/redo pairs at
22.24 ms for 100 realities, 205.01 ms for 1,000, and 1,069.71 ms for 5,000. The baseline measured
28.93 / 249.32 / 1,453.71 ms respectively, but unchanged checkpoint capture also varied between
those runs, so the comparison does not isolate the consolidation's effect. The current 5,000-state
fixture still averages about 53.5 ms per pair; large-history copying remains a performance limit.

Legacy machine implementations retain their existing blocking behavior. Context-free owner methods
can still deadlock inside their own synchronous callbacks; they cannot identify the caller under
the existing interface. Pass callback args for capture and retain the supplied context for other
operations. Arbitrary user code must cooperate with cancellation. A mandatory replacement for
the legacy snapshot contract would require a major-version API change. Cross-browser and
large-graph browser load testing were not performed.
