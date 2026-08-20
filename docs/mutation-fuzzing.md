# Mutation testing & fuzzing

This repo uses:

| Layer | Tool | Command |
|---|---|---|
| Go mutation | [Gremlins](https://gremlins.dev) v0.6 | `make tools && make test-mutation` |
| Go fuzz | Native `testing.F` (Go 1.18+) | `make test-fuzz-smoke` / `make test-fuzz` |
| Studio mutation | [Stryker](https://stryker-mutator.io) + Vitest runner | `make test-mutation-studio` |

## Go (Gremlins)

Config: [`.gremlins.yaml`](../.gremlins.yaml)

```bash
make tools                     # installs gremlins
make test-mutation-builtin     # fast gate (~5s)
make test-mutation-bot
make test-mutation-root        # serde + validators (+ deps)
make test-mutation-experimental
make test-mutation             # builtin + bot + root
make test-mutation-dry         # discover mutants only
```

Gremlins accepts **one package path** per invocation. Packages without tests used to break coverage gathering; stub `package_test.go` files keep `go test -cover ./...` healthy.

## Go fuzz

```bash
make test-fuzz-smoke   # ~5s per target
make test-fuzz         # FUZZTIME=15s by default
```

Targets:

- `experimental.FuzzProcessReference`
- `experimental.FuzzEventBuilder`
- `statepro.FuzzValidateDefinitionBinary`
- `statepro.FuzzDeserializeQuantumMachine`
- `builtin.FuzzBuiltinObserverArgs`

## Studio (Stryker)

Config: [`studio/packages/editor-core/stryker.config.json`](../studio/packages/editor-core/stryker.config.json)  
Narrow Vitest suite: [`vitest.mutation.config.ts`](../studio/packages/editor-core/vitest.mutation.config.ts) (avoids hanging the full editor suite under mutation).

```bash
make test-mutation-studio
# reports: studio/packages/editor-core/reports/mutation/
```

Mutates `validateStatePro`, `identifiers`, and `transitionRules` by default. Expand the `mutate` glob as the suite hardens.

## Interpreting scores

Aim to kill **observable** survivors (wrong branch, wrong sentinel, off-by-one on a documented boundary). Do **not** chase arithmetic on buffer capacity, loop-control on uniquely named maps, or sort comparator `<` vs `<=` when IDs are unique — those are usually equivalent mutants.

Approximate baselines after survivor hardening:

| Package | Tool | Score (ballpark) |
|---------|------|------------------|
| `builtin/` | Gremlins | ~92% (remaining LIVED: capacity arithmetic) |
| `experimental/` | Gremlins | ~89%+ |
| editor-core validators | Stryker | ~45% (narrow suite; raise by expanding mutate + tests) |
