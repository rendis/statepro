# Security, quality, and concurrency review — 2026-10-03

Reviewed baseline: `601f7af032d9f126ca4f140005c736210c57c696` (`main`, version 3.3.1).
This review combines code inspection, local reproductions, regression tests, and
dependency analysis. R01–R16 identify review observations; they are not GitHub
advisory IDs, and not every observation establishes an exploitable vulnerability.

## Comparison with previous GitHub work

The public issue and PR queries found no open issues on 2026-10-03.
The following changes were already merged into the reviewed baseline:

| Previous work | Status and relationship to this review |
| --- | --- |
| [Issue #9](https://github.com/rendis/statepro/issues/9), [PR #12](https://github.com/rendis/statepro/pull/12) | Earlier dependency updates, already closed/merged. |
| [PR #18](https://github.com/rendis/statepro/pull/18) | Consolidated #14–#17: snapshots, tracking, and runtime fixes. Concurrent metadata restoration and atomic rejection of invalid snapshots require the additional fixes in this PR. |
| [PR #19](https://github.com/rendis/statepro/pull/19) | Updated dependencies and documented that the autolayout test already hung before that update. This PR fixes the render loop. |
| [PR #20](https://github.com/rendis/statepro/pull/20) | Fixed ReDoS in identifier and JSON Pointer normalization, plus an esbuild advisory. The safe JSON property writes addressed here are a separate issue. |
| [PR #21](https://github.com/rendis/statepro/pull/21) | Released 3.3.1 with #19 and #20; it did not include the new fixes from this review. |

The current state of the private Code Scanning, Dependabot, and Security Advisories
dashboards was not queried. A closed issue or merged PR does not establish that
those dashboards have no alerts. No additional public vulnerability issue was
created: the repository directs vulnerability reports to private Security Advisories.

## Finding disposition

| ID | Observation | Outcome in this PR |
| --- | --- | --- |
| R01 | Metadata property and pointer safety | Fixed. Read only own properties; write and merge data properties while preserving JSON keys. Includes direct tests and tests for merging pack constants. |
| R02 | Unknown observer fallback | Existing contract, documented and tested in #14/#18: returns `true`. The default remains unchanged. A strict mode requires a compatibility proposal and an assessment of the trust boundary for definitions. |
| R03 | Conditions with the same `src` are lost | Fixed in import, export, and validation. Arguments, order, and multiplicity are preserved, matching Go runtime execution. Tests cover edited exports and unedited imported snapshots. |
| R04 | Concurrent global executor registration | Fixed with `RWMutex`. Concurrent test covers all four executor types under `-race`. The mutex is not held while executing a registered function. |
| R05 | Metadata restoration races with invokes | Fixed using the shared metadata mutex while preserving the map identity used by existing executor arguments. Includes concurrent access and a write after restoration. |
| R06 | Invalid snapshots mutate state before returning an error | Fixed by preparing and validating every included universe before applying changes. Validation covers essential flags, reality references, and accumulator entries. Omitted empty accumulators are reconstructed. |
| R07 | Unchanged measurements cause a render loop | Fixed with reducer identity bailouts and a stable callback. All five autolayout tests pass, including StrictMode; previously these tests did not complete. |
| R08 | Callbacks reenter the owner's public methods | Requires a design change. Synchronous callbacks execute under the machine mutex; they must use `args.GetSnapshot()` and `args.EmitEvent()` rather than synchronously calling the captured owner. Arguments must not be retained for asynchronous access to unsynchronized state. |
| R09 | Event submission ignores cancellation | Fixed at admission in `SendEvent`: check the context before and after acquiring the mutex. This does not add rollback or interrupt callbacks already executing; debugger bot loops and other operations require separate evaluation. |
| R10 | Validator cache uses only an editable ID | Fixed: key by schema content, isolate compilers per schema, and cap the LRU cache at 128 entries. Tests cover schema edits and reuse of the same ID across instances, including repeated `$id` values. |
| R11 | Restoration retains newer metadata, tracking, and flags | Fixed for included universes: replace metadata under its mutex, clear absent tracking, and recompute the final-state flag from the current reality or the reality before superposition. |
| R12 | Snapshot conversion hides JSON errors and loses precision | Pending. Requires a snapshot API that returns an error and a decision on a compatible numeric representation. Existing signatures and JSON conversion remain unchanged. |
| R13 | Nil inputs and custom events cause panics | Fixed for `NewQuantumMachine` with nil machine/universe models, `NewExQuantumMachine` with a nil model, `SendEvent` with typed or untyped nil events, and accumulation of custom `Event` implementations. This does not establish tolerance of arbitrary malformed models: callers remain responsible for validating definitions; `NewExUniverse(nil)` is outside this fix. |
| R14 | Invokes, accumulators, and histories have no resource budgets | Requires a design change. Invokes that continue after leaving a reality are an explicit contract in #14/#18. Limits and cancellation on exit should be configurable; this PR does not change their lifecycle. |
| R15 | History, serialization, and ELK loading costs | Pending optimization. History still compares/stores the full graph; measurement bailouts do not resolve that cost. The entry bundle still includes ELK. |
| R16 | Broken typecheck and ineffective lint checks | Fixed the Web Component typecheck by including the SVG declaration. Lint configuration remains a placeholder, and this PR does not add a CI workflow. |

## Behavior and compatibility

`LoadSnapshot(nil, ...)` remains a no-op. Partial snapshots remain supported:
only included, known universes are restored. For those universes, metadata and
tracking represent the restored state rather than a merge with newer state.
Rejected restoration leaves state, tracking, and machine context unchanged by
the restore operation. Validation does not verify cryptographic identity or
the definition's version.

Invokes that have already started may write metadata after restoration. This fix
synchronizes access; it does not cancel tasks or make their external effects
transactional. Metadata values must be JSON-compatible for the current snapshot
API. Nested mutable values must not be shared across goroutines without their
own synchronization.

The constructor that previously panicked on a nil model now returns an error.
Callers should handle the error already included in its signature. In Studio,
repeated condition executors no longer block export: repetition is valid in the
runtime, including when arguments are identical.

## Validation and coverage

The snapshot and Studio regressions were run against the unmodified baseline and
failed before the fixes. The original race checks reproduced concurrent registry
access and `LoadSnapshot`/invoke access. Permanent regression tests assert the
corrected behavior rather than the presence of a defect.

Reproducible commands from the repository root, with Go in `PATH`:

```sh
go build ./...
go vet ./...
go test -race -cover ./...
GOMAXPROCS=2 make test-fuzz-smoke
```

From `studio/`, using the repository's pinned pnpm version:

```sh
npx --yes pnpm@10.11.0 install --frozen-lockfile
npx --yes pnpm@10.11.0 test
npx --yes pnpm@10.11.0 typecheck
npx --yes pnpm@10.11.0 build
npx --yes pnpm@10.11.0 audit --prod
npx --yes pnpm@10.11.0 audit
```

Local validation passed: Go build and vet, and `go test -race -cover ./...` across
all 11 packages (builtin and bot 100%, experimental runtime approximately 92%,
CLI 1.3%). All three five-second fuzz smoke runs passed. Studio completed 251
tests in 44 files, plus typecheck and build for all three packages. App and Web
Component have no tests of their own: their scripts allow completion with no
test files, which is not counted as integration coverage.

`pnpm audit --prod` reports zero advisories. The full audit dropped from 15
advisories in the reviewed baseline to three: two moderate advisories (`vitest`
and `@vitest/mocker`,
[GHSA-82fw-gwwq-j7x9](https://github.com/advisories/GHSA-82fw-gwwq-j7x9),
requiring >=4.1.11), and one high advisory (`braces`,
[GHSA-vfj7-8cjw-p6xm](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm),
with no patched version according to the registry response). All are in
development dependencies. Registry advisories do not establish exploitability
in the deployed application. Migrating Vitest requires reviewing its dependencies
and Stryker integration; braces requires replacing or updating its parent
dependency when a solution becomes available.

The app's JavaScript entry remains approximately 2.28 MB minified / 677 kB gzip;
the build retains its bundle-size warning. Dependency audit results and counts
depend on the date and registry queried.

The initial analysis covered 296 tracked files (approximately 30,722 lines of Go,
TS, and TSX), with targeted inspection of higher-risk paths rather than exhaustive
inspection of every UI line. Real-browser, production, and mutation tests were
not run.

In the baseline review, `govulncheck` with Go 1.26.8 found no affected symbols;
with Go 1.25.0 it identified three `net/url` advisories through jsonschema's
initialization path. Malicious input reaching that path was not demonstrated.
Results depend on the toolchain, not only `go.mod`: use a version with security
patches, such as Go 1.25.13 or a later patch in that branch. Two advisories in
`x/text` and `x/sys` modules were not reachable in the analysis performed.
These baseline results do not establish coverage of other operating systems,
build tags, or toolchains.
