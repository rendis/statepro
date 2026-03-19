# AGENTS.md

Instructions for AI coding agents working in this repository.

## Build & Test Commands

### Go (core library)

- Build: `go build ./...`
- Test all: `go test ./...`
- Test single package: `go test ./experimental/...`
- Test single function: `go test -run TestFuncName ./...`

### Studio (visual editor — pnpm workspace)

All studio commands run from `studio/`:

- Install: `pnpm -C studio install`
- Dev (source mode): `pnpm -C studio dev` (sets `STUDIO_USE_EDITOR_CORE_SRC=true`)
- Dev (dist mode): `pnpm -C studio dev:dist`
- Build all packages: `pnpm -C studio build`
- Test: `pnpm -C studio test`
- Typecheck: `pnpm -C studio typecheck`
- Diagnose dev env: `pnpm -C studio dev:doctor`

Editor-core package specifically:

- Test: `pnpm -C studio/packages/editor-core test`
- Build: `pnpm -C studio/packages/editor-core build`
- Generate builtin catalog (runs before build/test automatically): `pnpm -C studio/packages/editor-core generate:builtin-catalog`

## Architecture

### Go Core (`github.com/rendis/statepro/v3`)

Quantum state machine library with multi-universe superposition. Three-layer architecture:

1. **`theoretical/`** — Pure data models mapping to JSON schema. `QuantumMachineModel` → `UniverseModel` → `RealityModel` → `TransitionModel`. No execution logic.

2. **`instrumentation/`** — Public interfaces. `QuantumMachine` (Init, SendEvent, GetSnapshot, LoadSnapshot), executor function types (`ObserverFn`, `ActionFn`, `InvokeFn`, `ConditionFn`), `Event`, `Accumulator`, `MachineSnapshot`.

3. **`experimental/`** — Runtime implementation. `ExQuantumMachine` orchestrates `ExUniverse` instances. Handles event routing, superposition/collapse, and emitted event cascading (FIFO, max depth 10).

Supporting packages:

- **`builtin/`** — Registry of built-in executors. Custom executors registered via `builtin.RegisterObserver()`, `RegisterAction()`, `RegisterInvoke()`, `RegisterCondition()`.
- **`statepro.go`** — Public entry: `NewQuantumMachine()`, event builders.
- **`serde.go`** — JSON serialization/deserialization of machine definitions.
- **`schema/`** — JSON schema files for machine definition validation.
- **`debugger/`** — CLI debugging tools (Bubble Tea TUI) and automation bot.

### Key runtime concepts

- **Universe**: an independent state machine with its own current reality
- **Reality**: a state within a universe
- **Superposition**: universe has no concrete reality; events accumulate in an `Accumulator` until an observer collapses it
- **Observers**: guard functions that watch accumulated events and trigger superposition collapse
- **Actions**: synchronous operations executed on transitions or entry
- **Invokes**: asynchronous operations
- **Conditions**: guard functions on transitions

### Studio (`studio/`)

pnpm monorepo with three packages:

- **`studio/app`** — Standalone Vite React app (port 5173). Local development UI.
- **`studio/packages/editor-core`** (`@rendis/statepro-studio-react`) — Core editor React library. Published to npm. Contains canvas, modals, reducers, auto-layout (elkjs), and all editor logic.
- **`studio/packages/web-component`** (`@rendis/statepro-studio-web-component`) — Framework-agnostic Custom Element wrapping editor-core.

`STUDIO_USE_EDITOR_CORE_SRC=true` makes the app import editor-core from source (via Vite alias) instead of dist.

### Testing conventions

- Go: standard `testing` package, table-driven tests
- Studio: Vitest + @testing-library/react + jsdom
- Test descriptions are written in **Spanish**
- Editor-core tests live in `src/__tests__/` directories alongside source
