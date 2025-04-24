# Global Variables (Immutable)
PROJECT_ROOT := $(CURDIR)/../..
VERSION := $(shell git describe --tags --match='v*' --always --dirty)
GO_MODULE := $(shell go list -m)

# Output Directories
BIN_DIR := bin
SERVER_OUT := $(BIN_DIR)/networkservice

# Build Flags
GO_LDFLAGS := -w -s
GO_TAGS :=
GO_TEST_FLAGS := -gcflags=all=-l -v -cover -short
GO_RACE_FLAGS := -race
GO_MSAN_FLAGS := -msan

# Packages
PKG := $(PROJECT_ROOT)
SERVER_PKG_BUILD := $(PKG)/cmd/networkservice
PKG_LIST := $(shell go list $(PKG)/...)

# Tools
GOLANGCI_LINT_VERSION := latest
GOSEC_VERSION := latest
GORELEASER_CONFIG := $(PROJECT_ROOT)/build/package/.goreleaser.yml

# Phony Targets
.PHONY: all build test govet golint golangcilint gosec race msan coverage coverhtml dep clean deb help

##@ General
all: build test govet golint golangcilint gosec ## Build, test, govet, golint, golangcilint and gosec  (default workflow)

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Build
build: dep ## Build server binary
	@mkdir -p $(BIN_DIR)
	@echo "⇒ Building server..."
	@go build -trimpath -tags='$(GO_TAGS)' -ldflags='$(GO_LDFLAGS)' -o $(SERVER_OUT) $(SERVER_PKG_BUILD)
	@echo "✓ Server built: $(SERVER_OUT)"

##@ Testing & Validation
test: dep ## Run unit tests
	@echo "⇒ Running unit tests..."
	@go test $(GO_TEST_FLAGS) $(PKG_LIST)

govet: dep ## Run static analysis with go vet
	@echo "⇒ Running govet..."
	@go vet -json $(PKG_LIST) 2>&1 | tee -a govet-report.out

golint: dep ## Run golint linter
	@echo "⇒ Running golint..."
	@go install golang.org/x/lint/golint@latest
	@golint ${PKG_LIST} | tee -a golint-report.out

golangcilint: dep ## Run golangci-lint linter
	@echo "⇒ Running golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@${GOPATH}/bin/golangci-lint run --out-format=checkstyle  $(PROJECT_ROOT)/... | tee -a golangcilint-report.xml

gosec: dep ## Run gosec security scanner
	@echo "⇒ Running gosec..."
	@go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	@${GOPATH}/bin/gosec -exclude-dir=.cache -fmt=sonarqube $(PROJECT_ROOT)/... | tee -a gosec-report.json

race: dep ## Run tests with race detector
	@echo "⇒ Running race detector..."
	@go test $(GO_TEST_FLAGS) $(GO_RACE_FLAGS) $(PKG_LIST)

msan: dep ## Run memory sanitizer
	@echo "⇒ Running memory sanitizer..."
	@go test $(GO_TEST_FLAGS) $(GO_MSAN_FLAGS) $(PKG_LIST)

##@ Coverage
coverage: dep ## Generate global code coverage report
	@echo "⇒ Generating global code coverage report..."
	@./coverage.sh

coverhtml: coverage ## Generate global code coverage report in HTML
	@echo "⇒ Generating global code coverage report in HTML..."
	@./coverage.sh html;

##@ Dependencies
dep: ## Install dependencies
	@echo "⇒ Installing tools..."
	@go mod tidy

##@ Cleanup
clean: ## Remove build artifacts
	@echo "⇒ Cleaning build artifacts..."
	@rm -rf $(BIN_DIR) dist/ *-report.xml *-report.json *-report.out
	@find . -name "coverage*" ! -name "coverage.sh" -exec rm -rf {} +

##@ Packaging
deb: dep ## Build Debian package
	@echo "⇒ Building Debian package..."
	@goreleaser --config $(GORELEASER_CONFIG) --snapshot --skip=publish --clean

.DEFAULT_GOAL := help
