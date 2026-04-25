# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **BREAKING**: Replaced `github.com/rendis/abslog/v3` with the standard library
  `log/slog` across the codebase. Log calls now use structured key-value
  attributes (e.g. `slog.WarnContext(ctx, "observer not found", "src", src)`)
  instead of printf-style formatting.
- Inlined the four helpers previously imported from
  `github.com/rendis/devtoolkit` (`ToInt`, `StructToMap`, `MapToStruct`,
  `Pair`/`NewPair`) into a new private `internal/util` package. No public API
  change. (#7)

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

[Unreleased]: https://github.com/rendis/statepro/compare/v3.2.1...HEAD
[3.2.1]: https://github.com/rendis/statepro/compare/v3.2.0...v3.2.1
[3.2.0]: https://github.com/rendis/statepro/compare/v3.1.1...v3.2.0
[3.1.1]: https://github.com/rendis/statepro/compare/v3.1.0...v3.1.1
[3.1.0]: https://github.com/rendis/statepro/compare/v3.0.0...v3.1.0
[3.0.0]: https://github.com/rendis/statepro/releases/tag/v3.0.0
