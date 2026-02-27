GO ?= go
SRC ?= .
BUILDDIR ?= build
BINARY ?= clio
COVERAGE_OUT ?= coverage.out
COVERAGE_MIN ?= 90

.DEFAULT_GOAL := help

.PHONY: build run install uninstall tidy fmt vet lint test cover ci clean help

build: ## Build the binary into ./build/
	@mkdir -p $(BUILDDIR)
	$(GO) -C $(SRC) build -o $(CURDIR)/$(BUILDDIR)/$(BINARY) ./cmd/clio

run: build ## Build and run the TUI
	$(CURDIR)/$(BUILDDIR)/$(BINARY)

install: build ## Install the binary to ~/.local/bin
	@mkdir -p $$HOME/.local/bin
	@install -m 0755 $(CURDIR)/$(BUILDDIR)/$(BINARY) $$HOME/.local/bin/$(BINARY)

uninstall: ## Remove the binary from ~/.local/bin
	@rm -f $$HOME/.local/bin/$(BINARY)

tidy: ## Tidy go.mod and go.sum
	$(GO) -C $(SRC) mod tidy

fmt: ## Format all Go source files
	$(GO) -C $(SRC) fmt ./...

vet: ## Run go vet
	$(GO) -C $(SRC) vet ./...

lint: fmt vet ## Run fmt and vet

test: ## Run tests with race detector
	$(GO) -C $(SRC) test -race -count=1 ./...

cover: ## Run tests and generate coverage report
	$(GO) -C $(SRC) test -race -count=1 -coverprofile=$(CURDIR)/$(COVERAGE_OUT) -covermode=atomic ./...
	$(GO) -C $(SRC) tool cover -func=$(CURDIR)/$(COVERAGE_OUT) | tail -5

ci: ## Format check, vet, tests and coverage >= COVERAGE_MIN%
	@echo "  → Format check..."
	@test -z "$$(gofmt -l ./$(SRC))" && echo "  ✓ Format OK" || { echo "  ✗ Run 'make fmt' to fix"; exit 1; }
	@echo "  → Vet..."
	@$(GO) -C $(SRC) vet ./... && echo "  ✓ Vet OK"
	@echo "  → Tests..."
	@$(GO) -C $(SRC) test -race -count=1 -coverprofile=$(CURDIR)/$(COVERAGE_OUT) -covermode=atomic ./...
	@echo "  → Coverage (min $(COVERAGE_MIN)%)..."
	@$(GO) -C $(SRC) tool cover -func=$(CURDIR)/$(COVERAGE_OUT) | awk \
	  '/^total:/{gsub(/%/,"",$$3); pct=$$3+0; \
	   if(pct<$(COVERAGE_MIN)){printf "  ✗ %.1f%% < $(COVERAGE_MIN)%%\n",pct; exit 1} \
	   else {printf "  ✓ %.1f%%\n",pct}}'
	@echo "  ✓ CI passed"

clean: ## Remove ./build/ and ./dist/
	rm -rf $(CURDIR)/$(BUILDDIR) $(CURDIR)/dist/

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
