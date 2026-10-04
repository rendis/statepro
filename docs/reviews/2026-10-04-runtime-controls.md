# Runtime controls and follow-up validation

Reviewed against merged `main` at `4d4959e6020ee32033cb69c270a3a74742ec1cba`.
This review covers the runtime, debugger, Studio validation, Web Component, mobile controls,
and the remaining layout/history performance costs.

## Confirmed problems and changes

| Area | Reproduction / finding | Result |
| --- | --- | --- |
| Invokes | Machine constants and reality invokes used separate unbounded goroutine launch paths. Leaving a reality did not cancel its work. | A shared manager covers machine constants, universe constants, and reality/transition invokes. Optional capacity rejects admission before spawning; exit cancellation and shutdown are cooperative. Completion and panic release capacity. |
| Accumulation and tracking | Persistent superposition and repeated transitions could grow retained state indefinitely. | Optional entry budgets reserve event fan-out before appending, validate restored accumulators, and retain recent tracking. Failed entry rollback retains the previous bounded history. Bot history has an independent optional limit. |
| Callbacks | Reentering the owning execution lock from a synchronous callback could deadlock; waiting callers could not cancel promptly. | Context-aware lock waiting and callback scope detection reject synchronous reentry, including nested calls across machines. Async invoke contexts retain cancellation and user values without inheriting the synchronous scope. |
| Unknown observers | An unregistered observer approved by default. | `StrictObservers` rejects missing registrations before event accumulation. Zero-value options preserve the existing permissive contract. |
| Bot | Cancellation was ignored between provider iterations and snapshot capture errors were hidden. | Check cancellation around provider/event processing, use optional contextual capture/restoration, propagate errors, and reject nil captures. A real-machine integration test cancels restoration while another callback holds its lock. |
| CLI | Copying a context with private fields panicked through reflection; empty list selection could panic; history navigation depended on saved checkpoints. | Copy the whole struct value, handle typed nil pointers and empty selections, and use actual history availability. Integration tests exercise a real machine, event history, rollback, checkpoint loading, and precise JSON display. |
| Studio validation | Twelve new cases failed before the fix: malformed JSON threw exceptions, and inherited `constructor` / `toString` values resolved as definition entries. | Type guards return structured errors; own-property lookup requires explicitly declared universes and realities. Valid own entries named `constructor` remain supported. |
| Web Component | Registering the same constructor under a second tag caused `NotSupportedError`. | Alternate tags receive subclasses. Tests cover native registration, React mounting, attributes/properties, disconnection/reconnection, and event/callback delivery. |
| Mobile UI | Header controls crowded each other; the machine panel exceeded the viewport; search overlapped it and expanded beyond the screen. | Responsive spacing, compact labeled controls, bounded panels, mobile search placement, and wrapping/stacked bottom toolbars. Browser assertions cover 320, 390, 768, and 1440 px. |
| Mutation tooling | A global Babel 7 override replaced Babel 8 required by Stryker 10, crashing instrumentation of TypeScript generics. | Scope the Babel 7 override to Babel 7 requests. Keep Stryker's Babel 8 dependency tree intact; frozen installation and dependency audit validate the resulting lockfile. |
| Test discovery | Vitest picked up Stryker sandbox copies, doubling the normal suite and treating instrumented files as application tests. | Exclude mutation sandboxes while retaining Vitest's default exclusions. Repeat the normal suite with the configured scope. |

All resource policies are opt-in. The mandatory machine/executor interfaces remain unchanged.
See [runtime policies](../runtime.md#runtime-policies-and-resource-limits) for budgets, error behavior,
callback context requirements, and shutdown semantics.

## Validation

- Go 1.25: build, vet, race-enabled tests with coverage across all 11 packages.
- CLI statement coverage increased from 1.3% to 51.0%; experimental runtime coverage is 90.1%.
- Three native fuzz smoke targets, five seconds each: references, binary definition validation,
  and builtin observer arguments.
- Node 24.19 / pnpm 10.11: frozen install, lint, typecheck, 283 tests (280 editor-core and 3 wrapper tests), builds, and full
  dependency audit (including development dependencies): zero advisories.
- Chromium: canvas, library, JSON export, bounded mobile header and search, production layout,
  and real Web Component locale/event flow. Wrapper unit tests mock only the editor boundary;
  browser tests render the actual editor.
- A regression test confirms that consumer mutations cannot change captured checkpoints,
  stored undo snapshots, or the restored graph's source snapshot.

### Complete configured mutation campaigns

These campaigns executed mutants, rather than stopping after an instrumented dry-run. Scope is
the configured targets below, not every file in the repository. Gremlins also mutates covered
module dependencies; target results overlap and must not be summed as unique mutants.

| Target | Killed | Survived | Uncovered | Timed out | Efficacy / total score | Mutant coverage / covered score |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Gremlins: builtin | 23 | 2 | 0 | 0 | 92.00% | 100.00% |
| Gremlins: bot | 24 | 1 | 0 | 2 | 96.00% | 100.00% |
| Gremlins: root and covered dependencies | 492 | 100 | 101 | 8 | 83.11% | 85.43% |
| Gremlins: experimental and covered dependencies | 289 | 42 | 12 | 49 | 87.31% | 96.50% |
| Stryker: configured Studio sources | 489 | 263 | 91 | 2 | 58.11% | 65.12% |

Gremlins efficacy excludes timeouts; Stryker counts them as detected. The root campaign preceded
one final private async-context adjustment; the completed experimental campaign and final Go
race tests cover that resulting runtime. All four Go campaigns exceeded their configured floors.
The final Studio campaign executed 845 mutants with 59 tests and no runner errors. The prior
16-test suite scored 48.88% on 804 mutants; test scope and guarded source changed together.
`transitionRules` improved from 45.22% to 84.35%; `validateStatePro` remains 45.65%, with much
of its issue-path mapping and diagnostic differentiation still weakly tested.

Stryker's expanded suite includes serialization, condition ordering, issue mapping, regression,
malformed-input, and transition-reference tests. Neither tool's score is a proof that no bugs remain.
Surviving and uncovered mutants remain a coverage backlog; timeouts and equivalent mutants must
be reviewed before treating them as meaningful defects.

## ELK size and checkpoint cost

The production build retains a separate ELK chunk of approximately 1,458 kB (445 kB gzip), with
an approximately 803 kB main JavaScript entry (225 kB gzip). Dynamic import removes ELK from the
main chunk, but the application's automatic initial layout still requests it during startup.
Manual layout reuses that loaded module. This does not eliminate the initial graph's layout cost.

`elkjs` distributes the full generated worker implementation. Selecting only `layered` in the
constructor controls registered algorithms; it does not remove the other algorithms' shipped
bytes. A smaller generated engine or a separately hosted worker would require a packaging and
hosting design change, not a configuration-only size fix. This change retains the existing
layered routing behavior and deployment contract.

History comparisons and coalescing already avoid unnecessary full-graph serialization. New
checkpoints and undo/redo still clone graph state so externally mutable data cannot corrupt history.
Removing those copies would require immutable ownership guarantees or a patch-based history
design. The regression test covers the isolation that must be preserved.

Reproduce the local benchmark after building editor-core:

```bash
cd studio
pnpm --filter @rendis/statepro-studio-react build
node scripts/benchmark-history.mjs
```

The fixture has one universe and 100, 1,000, or 5,000 realities, each with a 256-byte metadata payload
and no edges. For each operation, discard one warm-up, take five samples, and report median elapsed
milliseconds for 20 operations; undo/redo measures 20 pairs. These are Node timings on a local
machine, not browser frame timings or a universal performance threshold.

| Realities | Capture 20 checkpoints (ms) | Record 20 edits (ms) | Coalesce 20 edits (ms) | 20 undo/redo pairs (ms) |
| --- | ---: | ---: | ---: | ---: |
| 100 | 6.05 | 4.89 | 0.43 | 26.62 |
| 1000 | 62.33 | 98.34 | 4.67 | 239.83 |
| 5000 | 313.87 | 286.76 | 11.87 | 1059.54 |

At 5,000 realities, coalescing makes twenty edits about 24 times cheaper than recording each edit
in this fixture. Undo/redo still averages about 53 ms per pair, exceeding a 16 ms frame budget.
These measurements justify preserving coalescing and identifying clone-heavy large histories as
an architectural optimization opportunity, rather than silently weakening snapshot isolation.

## Practical limits

- Callbacks and invokes must cooperate with cancellation. Existing context-free snapshot methods
  can still deadlock when called on their owner inside a synchronous callback; use callback args
  or the new contextual capability. Replacing the supplied context bypasses scope detection.
- Limit/cancellation errors do not roll back prior side effects, other universes, or already admitted
  invokes. `Close()` requests shutdown; `WaitInvokes` needs a deadline for uncooperative user code.
- Accumulator budgets count entries, not payload bytes; history bounds do not bound individual
  snapshot size. The bot's legacy provider cannot be forcibly interrupted and runs are not concurrent.
- Strict observers and finite budgets require application configuration. Existing defaults remain
  compatible. ELK download size and full checkpoint copy costs remain architectural tradeoffs.
- CLI coverage is broader but not complete. Browser validation used Chromium; no cross-browser
  or distributed-service load campaign is claimed.
