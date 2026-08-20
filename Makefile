# StatePro quality gates
#
# Mutation (Go):     make test-mutation
# Mutation (Studio): make test-mutation-studio
# Fuzz (Go native):  make test-fuzz
# Fuzz short smoke:  make test-fuzz-smoke

GO ?= go
GOPATH_BIN := $(shell $(GO) env GOPATH)/bin
export PATH := $(GOPATH_BIN):$(PATH)

GREMLINS ?= gremlins
FUZZTIME ?= 15s

.PHONY: tools test test-race test-fuzz test-fuzz-smoke test-mutation test-mutation-dry test-mutation-builtin test-mutation-experimental test-mutation-root test-mutation-bot test-mutation-studio

tools: ## Install mutation tooling (gremlins)
	$(GO) install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0

test: ## Run all Go unit tests
	$(GO) test ./...

test-race: ## Run Go tests with the race detector
	$(GO) test -race ./...

test-fuzz-smoke: ## Short native fuzz smoke (CI-friendly)
	$(GO) test -fuzz=FuzzProcessReference -fuzztime=5s ./experimental/
	$(GO) test -fuzz=FuzzValidateDefinitionBinary -fuzztime=5s .
	$(GO) test -fuzz=FuzzBuiltinObserverArgs -fuzztime=5s ./builtin/

test-fuzz: ## Longer native Go fuzz pass
	$(GO) test -fuzz=FuzzProcessReference -fuzztime=$(FUZZTIME) ./experimental/
	$(GO) test -fuzz=FuzzValidateDefinitionBinary -fuzztime=$(FUZZTIME) .
	$(GO) test -fuzz=FuzzDeserializeQuantumMachine -fuzztime=$(FUZZTIME) .
	$(GO) test -fuzz=FuzzBuiltinObserverArgs -fuzztime=$(FUZZTIME) ./builtin/
	$(GO) test -fuzz=FuzzEventBuilder -fuzztime=$(FUZZTIME) ./experimental/

# Gremlins accepts a single package path. Packages without tests break coverage
# gathering on module root, so mutation targets are run per package.

test-mutation-dry: tools ## Discover mutants (per package, dry-run)
	$(GREMLINS) unleash --dry-run ./builtin
	$(GREMLINS) unleash --dry-run ./experimental
	$(GREMLINS) unleash --dry-run ./debugger/bot
	$(GREMLINS) unleash --dry-run .

test-mutation-builtin: tools ## Mutation test builtin (fast gate)
	$(GREMLINS) unleash --workers 2 --timeout-coefficient 3 --threshold-efficacy 70 ./builtin

test-mutation-experimental: tools ## Mutation test experimental runtime
	$(GREMLINS) unleash --workers 2 --timeout-coefficient 3 --threshold-efficacy 55 ./experimental

test-mutation-root: tools ## Mutation test root package (serde/validation)
	$(GREMLINS) unleash --workers 2 --timeout-coefficient 3 --threshold-efficacy 55 .

test-mutation-bot: tools ## Mutation test debugger bot
	$(GREMLINS) unleash --workers 2 --timeout-coefficient 3 --threshold-efficacy 70 ./debugger/bot

test-mutation: test-mutation-builtin test-mutation-bot test-mutation-root ## Core Go mutation gates (builtin+bot+root)
	@echo "Core mutation gates done. For runtime: make test-mutation-experimental"

test-mutation-studio: ## Mutation test editor-core (Stryker)
	pnpm -C studio/packages/editor-core test:mutation
