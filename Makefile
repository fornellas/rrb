BINDIR ?= /usr/local/bin

GO ?= go

GOIMPORTS ?= goimports
GOIMPORTS_VERSION ?= latest
GOIMPORTS_LOCAL ?= github.com/fornellas/

GOCYCLO ?= gocyclo
GOCYCLO_VERSION = latest
GOCYCLO_OVER ?= 10

GOLANGCI_LINT ?= golangci-lint
GOLANGCI_LINT_VERSION ?= latest
GOLANGCI_LINT_ARGS ?= --timeout 10m

GO_TEST ?= gotest
GOTEST_VERSION ?= latest
GO_TEST_FLAGS ?= -v -race -cover -count=1

##
## Help
##

.PHONY: help
help:

##
## Clean
##

.PHONY: clean
clean:
clean-help:
	@echo 'clean: clean all files'
help: clean-help

##
## Install Deps
##

.PHONY: install-deps-help
install-deps-help:
	@echo 'install-deps: install dependencies required by the build at BINDIR=$(BINDIR)'
help: install-deps-help

.PHONY: install-deps
install-deps:

.PHONY: uninstall-deps-help
uninstall-deps-help:
	@echo 'uninstall-deps: uninstall dependencies required by the build'
help: uninstall-deps-help

.PHONY: uninstall-deps
uninstall-deps:

##
## Lint
##

# lint

.PHONY: lint-help
lint-help:
	@echo 'lint: runs all linters'
help: lint-help

.PHONY: lint
lint:

# Generate

.PHONY: go-generate
go-generate:
	$(GO) generate ./...

# goimports

.PHONY: install-deps-goimports
install-deps-goimports:
	GOBIN=$(BINDIR) $(GO) install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
install-deps: install-deps-goimports

.PHONY: uninstall-deps-goimports
uninstall-deps-goimports:
	rm -f $(BINDIR)/goimports
uninstall-deps: uninstall-deps-goimports

.PHONY: goimports
goimports:
	$(GOIMPORTS) -w -local $(GOIMPORTS_LOCAL) .
lint: goimports

# go mod tidy

.PHONY: go-mod-tidy
go-mod-tidy: go-generate goimports
	$(GO) mod tidy
lint: go-mod-tidy

# gocyclo

.PHONY: install-deps-gocyclo
install-deps-gocyclo:
	GOBIN=$(BINDIR) $(GO) install github.com/fzipp/gocyclo/cmd/gocyclo@$(GOCYCLO_VERSION)
install-deps: install-deps-gocyclo

.PHONY: uninstall-deps-gocyclo
uninstall-deps-gocyclo:
	rm -f $(BINDIR)/gocyclo
uninstall-deps: uninstall-deps-gocyclo

.PHONY: gocyclo
gocyclo: go-generate go-mod-tidy
	$(GOCYCLO) -over $(GOCYCLO_OVER) -avg .
lint: gocyclo

# golangci-lint

.PHONY: install-deps-golangci-lint
install-deps-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BINDIR) $(GOLANGCI_LINT_VERSION)
install-deps: install-deps-golangci-lint

.PHONY: uninstall-deps-golangci-lint
uninstall-deps-golangci-lint:
	rm -f $(BINDIR)/golangci-lint
uninstall-deps: uninstall-deps-golangci-lint

.PHONY: golangci-lint
golangci-lint: go-mod-tidy go-generate
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_ARGS)
lint: golangci-lint

.PHONY: clean-golangci-lint
clean-golangci-lint:
	$(GOLANGCI_LINT) cache clean
clean: clean-golangci-lint

# go vet

.PHONY: go-vet
go-vet: go-mod-tidy go-generate
	$(GO) vet ./...
lint: go-vet

##
## Test
##

# test

.PHONY: test-help
test-help:
	@echo 'test: runs all tests'
help: test-help

.PHONY: test
test:

# gotest

.PHONY: install-deps-gotest
install-deps-gotest:
	GOBIN=$(BINDIR) $(GO) install github.com/rakyll/gotest@$(GOTEST_VERSION)
install-deps: install-deps-gotest

.PHONY: uninstall-deps-gotest
uninstall-deps-gotest:
	rm -f $(BINDIR)/gotest
uninstall-deps: uninstall-deps-gotest

.PHONY: test
gotest: go-generate
	$(GO_TEST) ./... $(GO_TEST_FLAGS)
test: gotest

.PHONY: clean-gotest
clean-gotest:
	$(GO) clean -r -testcache
clean: clean-gotest

##
## Build
##

.PHONY: build-help
build-help:
	@echo 'build: build everything'
help: build-help

.PHONY: build
build: go-generate
	$(GO) build .

.PHONY: clean-build
clean-build:
	$(GO) clean -r -cache -modcache ./...
clean: clean-build