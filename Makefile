GO ?= go

GOIMPORTS ?= goimports
GOIMPORTS_VERSION ?= latest
GOIMPORTS_LOCAL = github.com/fornellas/rrb/

GOCYCLO ?= gocyclo
GOCYCLO_VERSION = latest
GOCYCLO_OVER ?= 10

GOLANGCI_LINT ?= golangci-lint
GOLANGCI_LINT_VERSION ?= latest
GOLANGCI_LINT_RUN_ARGS ?= --timeout 5m

GO_TEST ?= gotest
GOTEST_VERSION ?= latest
GO_TEST_FLAGS ?= -v -race -cover -count=1

##
## Help
##

# help

.PHONY: help
help:

##
## Dependencies
##

# install-deps

.PHONY: install-deps-help
install-deps-help:
	@echo 'install-deps: install dependencies required for the build'
help: install-deps-help
.PHONY: install-deps
install-deps:
	$(GO) install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GO) install github.com/fzipp/gocyclo/cmd/gocyclo@$(GOCYCLO_VERSION)
	$(GO) install github.com/rakyll/gotest@$(GOTEST_VERSION)

##
## Generate
##

# generate

.PHONY: generate-help
generate-help:
	@echo 'generate: runs `go generate`'
help: generate-help
.PHONY: generate
generate:
	$(GO) generate

##
## lint
##

# goimports

.PHONY: goimports-help
goimports-help:
	@echo 'goimports: formats all files with goimports'
help: goimports-help
.PHONY: goimports
goimports: generate
	$(GOIMPORTS) -w -local $(GOIMPORTS_LOCAL) .

# go mod tidy

.PHONY: go-mod-tidy-help
go-mod-tidy-help:
	@echo 'go-mod-tidy: runs `go mod tidy`'
help: go-mod-tidy-help
.PHONY: go-mod-tidy
go-mod-tidy: goimports
	$(GO) mod tidy

# golangci-lint

.PHONY: golangci-lint-help
golangci-lint-help:
	@echo 'golangci-lint: runs golangci-lint'
help: golangci-lint-help
.PHONY: golangci-lint
golangci-lint: go-mod-tidy generate
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_RUN_ARGS)
.PHONY: clean-golangci-lint
clean-golangci-lint:
	$(GOLANGCI_LINT) cache clean
clean: clean-golangci-lint

# gocyclo

.PHONY: gocyclo-help
gocyclo-help:
	@echo 'gocyclo: runs gocyclo'
help: gocyclo-help
.PHONY: gocyclo
gocyclo: go-mod-tidy generate
	$(GOCYCLO) -over $(GOCYCLO_OVER) -avg .

# lint

.PHONY: lint-help
lint-help:
	@echo "lint: lint all files"
help: lint-help
.PHONY: lint
lint: goimports go-mod-tidy golangci-lint gocyclo

##
## test
##

.PHONY: test-help
test-help:
	@echo 'test: runs all tests'
help: test-help
.PHONY: test
test: generate
	$(GO_TEST) ./... $(GO_TEST_FLAGS)
.PHONY: clean-test
clean-test:
	$(GO) clean -r -testcache
clean: clean-test

##
## build
##

# build

.PHONY: build-help
build-help:
	@echo 'build: builds everything'
help: build-help
.PHONY: build
build: generate
	$(GO) build
.PHONY: clean-build
clean-build:
	$(GO) clean -r -cache -modcache
clean: clean-build

##
## clean
##

# clean
.PHONY: clean
clean: