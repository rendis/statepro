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
make test-mutation-cli         # includes CLI startup, errors, and history workflows
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

Complete campaign measurements on 2026-10-04:

| Target | Tool | Efficacy / mutation score | Mutant coverage / covered score |
| --- | --- | --- | --- |
| `builtin/` | Gremlins | 92.00% | 100.00% |
| `debugger/bot/` | Gremlins | 96.00% | 100.00% |
| root package and covered dependencies | Gremlins | 83.11% | 85.43% |
| `experimental/` and covered dependencies | Gremlins | 87.31% | 96.50% |
| editor-core configured three source files | Stryker | 58.11% | 65.12% |

Stryker ran the expanded 59-test suite against 845 mutants: 489 killed, 2 timed out,
263 survived, and 91 uncovered, with no runner errors. The prior 16-test suite scored 48.88%
on 804 mutants; both test scope and guarded source changed, so the totals differ.
Gremlins efficacy excludes timeout results; Stryker includes timeouts in detected mutants.
Covered scores exclude uncovered mutants and are not comparable to statement coverage.

On 2026-10-05, the same 845 Studio mutants scored **85.56% total / 88.17% covered** after
adding diagnostic cases: 721 killed, 2 timed out, 97 survived, and 25 uncovered. Validator score
increased from 45.65% to 86.86%. CLI mutation execution is now available through
`make test-mutation-cli`: 69 killed, zero survivors among executed mutants, and 60 uncovered
(100% efficacy / 53.49% mutant coverage). See the [follow-up review](reviews/2026-10-05-validation-api-performance.md)
for scope, instrumentation results, regression evidence, and remaining limits.

The normal Vitest configuration excludes `.stryker-tmp` to avoid discovering duplicated or
instrumented test files. Babel overrides retain the major version required by the Stryker
instrumenter. See the [follow-up report](reviews/2026-10-04-runtime-controls.md) for counts,
reproduction methods, performance measurements, and remaining limits.
