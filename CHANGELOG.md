# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Studio: metadata pack schemas are compiled with one shared Ajv instance instead of one instance per schema (128 distinct schemas: about 1.4 s to 60 ms). Edited schemas are still recompiled.
- Studio: `@rendis/statepro-studio-react` no longer depends on `fast-deep-equal`.

### Fixed

- Runtime: unlocking the machine mutex when it is not locked panics, like `sync.Mutex`, instead of blocking forever.
- Studio: an edit whose patch only sets absent fields to `undefined` no longer records an undo step that changes nothing.

## [3.4.0] - 2026-10-05

### Added

- Runtime: `statepro.NewQuantumMachineWithOptions` and `experimental.NewExQuantumMachineWithOptions` accept `instrumentation.RuntimeOptions`: a shared `MaxConcurrentInvokes` budget, `CancelInvokesOnExit`, per-universe `MaxAccumulatedEvents` and `MaxTrackingEntries`, and `StrictObservers`. Existing constructors keep the previous unlimited, permissive behavior.
- Runtime: invoke capacity is reserved for a whole transition step (transition, exit, and target entry invokes) before its first callback, so a `ResourceLimitError` never leaves a transition half-applied.
- Runtime: context-aware execution. Operations cancel while waiting for the machine lock and are rejected if the context is already done at admission; once admitted they run to completion. Synchronous callbacks that reenter their owning machine with the callback context get `ErrReentrantCall` instead of deadlocking.
- Runtime: optional `instrumentation.RuntimeLifecycle` (`Close`, `WaitInvokes`) and `instrumentation.ContextSnapshotProvider` (`GetSnapshotContext`, `LoadSnapshotContext`) capabilities, plus the `instrumentation.GetSnapshotContext` and `instrumentation.LoadSnapshotContext` helpers that pick the best capability of any machine.
- Runtime: checked snapshot capture through the optional `instrumentation.SnapshotProvider` capability and the `instrumentation.GetSnapshotWithError` helper, usable from synchronous action args without reacquiring the machine lock.
- Debugger bot: cancellable event processing, checked snapshot errors, and an optional history retention limit (`bot.WithHistoryLimit`).
- Studio: optional `autoLayoutWorkerUrl` prop (`auto-layout-worker-url` attribute on the Web Component) to run ELK layout in a Web Worker. One worker is reused across layouts and released after 60 seconds idle; if it cannot load or start (missing asset, CSP, cross-origin URL), layout falls back to the bundled engine on the main thread. The standalone app enables it.
- Docs: compiled `ExampleNewQuantumMachineWithOptions` showing budgets, callback-safe capture, a cooperative invoke, and bounded shutdown.

### Changed

- Runtime: snapshot capture, `json.Unmarshal` into `MachineSnapshot`, and `LoadSnapshot` keep `json.Number` for metadata and event data numbers that would change through `float64` (large integers, long decimals). Values that round-trip exactly stay `float64`. Consumers of untyped metadata must accept both.
- Runtime: the legacy `GetSnapshot()` logs capture failures (for example unsupported JSON values in metadata) and returns `nil` instead of a partial snapshot.
- Studio: `@rendis/statepro-studio-react/styles.css` no longer contains `@tailwind` directives. The host app's Tailwind setup generates the utilities, as documented; under Tailwind 4 the directives emitted extra utilities outside any cascade layer.
- Studio: ELK is loaded on demand instead of in the main bundle (main JavaScript of the standalone app drops from about 2.3 MB to 0.8 MB minified).
- Studio: faster undo history. Unchanged sections are compared by identity before deep comparison instead of serializing the graph, transitions are grouped by source once during serialization, and committing the editor's own drag checkpoint no longer clones the graph twice. Checkpoints from the public `createHistorySnapshot` remain caller-owned and are copied on `commit-snapshot`.
- Studio: the development workspace requires Node 22 or later; the standalone app uses Tailwind 4's Vite plugin; tests run on Vitest 4; ESLint enforces React hook ordering. A `Quality` workflow runs Go build, vet, race tests and fuzz smoke, and Studio lint, typecheck, tests, Stryker dry-run, build and dependency audit.
- Studio: `@rendis/statepro-studio-react` bumped to `0.2.0` and `@rendis/statepro-studio-web-component` to `0.2.0`.

### Deprecated

- Runtime: the context-free machine `GetSnapshot` and `LoadSnapshot` methods. Use `instrumentation.GetSnapshotContext` and `instrumentation.LoadSnapshotContext` outside callbacks, or `instrumentation.GetSnapshotWithError(args)` inside a synchronous action. Checked capture with `GetSnapshotWithError` remains supported.

### Security

- Studio: metadata pointer writes and object merges use own data properties, preserving JSON keys without traversing inherited properties.
- Studio: update compatible dependency overrides, including `fast-uri` to `^3.1.8`, and remove the Tailwind 3 watcher chain that pulled in a vulnerable `braces`.

### Fixed

- Runtime: validate all included universe snapshots before applying them; restore metadata and tracking instead of retaining newer entries, synchronize metadata restoration with invokes, and reconstruct empty superposition accumulators and final-state flags.
- Runtime: synchronize custom executor registration and lookup; accept custom implementations of the public `Event` interface in accumulators.
- Runtime: return errors for nil machine/universe models and nil events.
- Runtime: preserve bounded tracking during failed entry rollback.
- CLI: avoid reflection panics when copying contexts with private fields or typed nil pointers; handle empty lists and history independently of saved checkpoints; show event, restore and capture errors in the interface instead of printing them.
- Studio: reject malformed JSON without throwing and resolve definition references only from own properties, while allowing explicitly declared names such as `constructor`.
- Studio: a malformed `value`/`defaultValue` definition no longer crashes the editor. The initial render falls back to an empty machine, a controlled update keeps the current graph, and both report the error through `console.error`.
- Studio: preserve condition order, arguments, and repeated executors during import/export and validation, matching runtime semantics.
- Studio: prevent unchanged node measurements from creating render cycles; stabilize the measurement callback and preserve reducer identity for no-op updates.
- Studio: recompile changed metadata schemas even when pack IDs are reused, isolate schema IDs, and bound the validator cache to 128 entries.
- Studio: fix hook order in `PropertiesModal` and `JsonIOModal`, which could change across renders.
- Studio: fit header, machine panel, and search controls within mobile viewports.
- Web Component: support registering multiple tag names and cover attributes, properties, lifecycle, and DOM events with integration tests.
- Studio tooling: include SVG declarations when typechecking the Web Component against editor-core source; exclude Stryker sandboxes from normal Vitest discovery; keep Babel 7 overrides scoped to Babel 7 so the Stryker 10 instrumenter can use Babel 8.

## [3.3.1] - 2026-08-25

### Security

- Studio: bumped `vite` to `^6.4.3` and `vitest` to `^3.2.6` to address Dependabot alerts (critical Vitest UI RCE, high Vite `server.fs.deny` bypass on Windows). (#19)
- Studio: bumped `postcss` to `^8.5.26` and pinned transitive `nanoid`, `ws`, `fast-uri`, `qs`, and `@babel/core` via `pnpm.overrides`. (#19)
- Studio: pinned tsup's `esbuild@^0.27.0` to `0.28.2` (Dependabot #23 — arbitrary file read on the Windows esbuild dev server). Vite remains on `esbuild@0.25.12`, which is outside the vulnerable range. (#20)
- Studio: replaced quantified regular expressions in identifier normalization and JSON Pointer trimming with linear scans (CodeQL polynomial ReDoS). (#20)

### Changed

- Studio: `@rendis/statepro-studio-react` bumped to `0.1.4`.

## [3.3.0] - 2026-08-20

### Fixed

- Experimental runtime: `LoadSnapshot` now restores per-universe tracking history.
- Experimental runtime: `GetSnapshot` copies tracking slices and serializes under the machine mutex (actions use an unlocked snapshot path to avoid deadlock).
- Experimental runtime: machine-level constant actions report the correct `ActionType` (`entry` / `exit` / `transition`).
- Experimental runtime: final realities ignore `On` handlers; `ReplayOnEntry` skips finalized universes.
- Experimental runtime: failed entry actions roll back the previous reality instead of leaving a half-applied state.
- Experimental runtime: observer errors are ignored when another observer approves (matches documented semantics).
- Experimental runtime: superposition collapse iterates realities in sorted-id order (deterministic).
- Experimental runtime: always-transition cycles and notify/external-target cascades fail with an error instead of hanging.
- Experimental runtime: invalid external targets return an error instead of panicking; single-character IDs match the JSON schema.
- Experimental runtime: `SendEvent` accumulates into superposition universes when no concrete `On` handler exists, without double-delivering events that are already routed as external targets.
- Experimental runtime: invoke/action metadata mutations are synchronized so concurrent `GetSnapshot` is race-free.

### Added

- Experimental runtime: universe-level `UniversalConstants` now execute between machine-level constants and reality-level actions/invokes.
- Adversarial / hostile test suite across runtime, definition validation, builtin observers, debugger bot, and Studio validation (races, cascade depth limits, invoke panic isolation, malformed payloads, fuzz for `processReference`).
- Mutation testing tooling: Gremlins (Go) via `.gremlins.yaml` + `make test-mutation`, and Stryker+Vitest (Studio editor-core) via `make test-mutation-studio`.
- Expanded native Go fuzz targets (`FuzzValidateDefinitionBinary`, `FuzzDeserializeQuantumMachine`, `FuzzBuiltinObserverArgs`, `FuzzEventBuilder`) with `make test-fuzz` / `make test-fuzz-smoke`.
- Mutation-survivor tests: exact observer boundaries (`GreaterThanEqualCounter`, `TotalEventsBetweenLimits`), experimental helpers/routing/emit-depth/snapshot/metadata cases, and additional Studio `validateStatePro` assertions that kill clear Stryker survivors.

### Changed

- Studio: validation no longer warns that the runtime ignores `universe.universalConstants`; those constants are executed by the experimental runtime.
- Experimental runtime: invoke goroutines recover from panics so a crashing invoke cannot take down the process.
- Studio: `@rendis/statepro-studio-react` bumped to `0.1.3`.


## [3.2.2] - 2026-04-25

### Changed

- **BREAKING**: Replaced `github.com/rendis/abslog/v3` with the standard library
  `log/slog` across the codebase. Log calls now use structured key-value
  attributes (e.g. `slog.WarnContext(ctx, "observer not found", "src", src)`)
  instead of printf-style formatting.
- Inlined the four helpers previously imported from
  `github.com/rendis/devtoolkit` (`ToInt`, `StructToMap`, `MapToStruct`,
  `Pair`/`NewPair`) into a new private `internal/util` package. No public API
  change. (#7)
- Bumped `debugger/cli` TUI dependencies to their latest v1 releases:
  `charmbracelet/bubbles` v0.20.0 → v1.0.0, `charmbracelet/bubbletea` v1.2.4
  → v1.3.10, `charmbracelet/lipgloss` v1.0.0 → v1.1.0. Plus minor / patch
  bumps on transitive deps (`x/ansi`, `x/term`, `x/sys`, `x/text`, `mattn/*`,
  `lucasb-eyer/go-colorful`, `muesli/termenv`). (#9)

### Removed

- **BREAKING**: Removed the public function `builtin.SetLogger(abslog.AbsLog)`.
  Consumers must now configure logging via the standard `slog.SetDefault(*slog.Logger)`.
- Direct dependency on `github.com/rendis/abslog/v3` and its transitive
  dependencies (`logrus`, `logrus-stackdriver-formatter`, `zap`, `multierr`,
  `go-stack/stack`).
- Direct dependency on `github.com/rendis/devtoolkit` and its transitive
  dependencies (`gabriel-vasile/mimetype`, `go-playground/locales`,
  `go-playground/universal-translator`, `go-playground/validator/v10`,
  `leodido/go-urn`, `golang.org/x/crypto`, `golang.org/x/exp`,
  `golang.org/x/net`, `gopkg.in/yaml.v3`). (#7)

### Security

- Studio: bumped `vite` to `^6.4.2` to address Dependabot alerts #12 (high —
  arbitrary file read via dev-server WebSocket) and #13 (medium — path
  traversal in optimized deps `.map` handling). (#9)
- Studio: bumped `postcss` to `^8.5.10` to address Dependabot alert #14
  (medium — XSS via unescaped `</style>` in CSS stringify output). (#9)
- Studio: pinned all transitive `picomatch` resolutions to `^4.0.4` via a
  `pnpm.overrides` entry in `studio/package.json` to address Dependabot alerts
  #8 / #9 (high / medium — ReDoS and method injection in versions <2.3.2)
  and #10 / #11 (high / medium — same vectors in 4.0.0–4.0.3). (#9)

### Migration

Before:

```go
import "github.com/rendis/abslog/v3"
import "github.com/rendis/statepro/v3/builtin"

builtin.SetLogger(myAbsLogger)
```

After:

```go
import "log/slog"

slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
```

## [3.2.1] - 2026-03-19

### Added

- `EmitEvent` for internal event emission from entry actions in the experimental
  runtime.
- `AGENTS.md`, `CLAUDE.md`, and a statepro skill for AI coding agents.
- Studio: scoped canvas search UX with morphing toolbar, multi-selection and
  visualization filter UX, standardized tooltips across the editor.

### Changed

- Studio: upgraded to React 19; renamed packages to the `@rendis` scope.
- Studio: low-latency rendering improvements, skeleton connections, smoother
  machine panel expansion.

### Fixed

- Experimental runtime: observer-superposition collapse, metadata persistence,
  and double-init guard.
- Editor-core: persist always-trigger updates, fix close-search on result
  double click, tune search focus and pulse highlight behavior.
- Studio: propagate universe/reality id renames safely; protect built-in
  library behaviors.

### Security

- Updated `golang.org/x/crypto` to v0.49.0 (Dependabot alerts #5 and #6).
- Studio: upgraded `vitest` to v3 to address Dependabot alert #7 (esbuild CVE).

## [3.2.0] - 2026-02-01

### Added

- `debugger/bot`: option to ignore unhandled events.

## [3.1.1] - 2025-12-10

### Fixed

- Reset reality initialized flag to ensure entry actions execute.

## [3.1.0] - 2025-10-01

### Added

- `PositionMachine` and `PositionMachineOnInitial` convenience methods on
  `QuantumMachine`.
- Canonical name positioning methods.
- Expanded godoc for the `QuantumMachine` interface and the new positioning
  APIs.

### Changed

- Refactored superposition snapshot logic into a separate method.

## [3.0.0] - 2025-09-25

Initial release of the v3 module path (`github.com/rendis/statepro/v3`).
This is a major rewrite introducing the quantum state machine model with
universes, realities, superposition, observers, and accumulators.

See the GitHub release page for the full v3.0.0 release notes.

[Unreleased]: https://github.com/rendis/statepro/compare/v3.4.0...HEAD
[3.4.0]: https://github.com/rendis/statepro/compare/v3.3.1...v3.4.0
[3.3.1]: https://github.com/rendis/statepro/compare/v3.3.0...v3.3.1
[3.3.0]: https://github.com/rendis/statepro/compare/v3.2.2...v3.3.0
[3.2.2]: https://github.com/rendis/statepro/compare/v3.2.1...v3.2.2
[3.2.1]: https://github.com/rendis/statepro/compare/v3.2.0...v3.2.1
[3.2.0]: https://github.com/rendis/statepro/compare/v3.1.1...v3.2.0
[3.1.1]: https://github.com/rendis/statepro/compare/v3.1.0...v3.1.1
[3.1.0]: https://github.com/rendis/statepro/compare/v3.0.0...v3.1.0
[3.0.0]: https://github.com/rendis/statepro/releases/tag/v3.0.0
