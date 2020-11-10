PROJECT_ROOT:=${PWD}/../../
LD_FLAGS="-w -s"
VERSION=$(shell git describe --tags --match=v* --always --dirty)


SERVER_OUT := "bin/networkservice"

PKG := "${PROJECT_ROOT}"
SERVER_PKG_BUILD := "${PKG}/cmd/networkservice"
PKG_LIST := $(shell go list ${PKG}/...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go | grep -v .pb)

.PHONY: all
all: build test vet lint fmt ## Run build, test, vet, lint and fmt

.PHONY: build
# build: clean server client	## Build server
build: clean server	## Build both api, server (maybe clients)

.PHONY: test
test:	## Run unit tests
	$(info Running unit tests...)
	@go test -cover -short ${PKG_LIST}

.PHONY: vet
vet:
	$(info Running go vet...)
	@go vet -all ${PKG_LIST} | tee -a govet-report.out

.PHONY: lint
lint:	# Lint the files
	$(info Running golint...)
	@golint ${PKG_LIST} | tee -a golint-report.out

.PHONY: fmt
fmt:
	$(info Running go fmt...)
	@test -z $$(go fmt ${PKG_LIST})

race: dep ## Run data race detector
	$(info Running data race detector...)
	@go test -race -short ${PKG_LIST}

msan: dep ## Run memory sanitizer
	$(info Running memory sanitizer...)
	@go test -msan -short ${PKG_LIST}

coverage: ## Generate global code coverage report
	$(info Generating global code coverage report...)
	./coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	$(info Generating global code coverage report in HTML...)
	./coverage.sh html;

dep: ## Get the dependencies
	$(info Get the dependencies...)
	@go get -v -d ./...
	@go get -v -u golang.org/x/lint/golint

server: dep ## Build the binary file for server
	$(info Building server...)
	@go build -i -v -ldflags=$(LD_FLAGS) -o $(SERVER_OUT) $(SERVER_PKG_BUILD)
	$(info Built server at $(SERVER_OUT) )

clean: ## Remove previous builds
	$(info Running clean up, removing previous builds...)
	@rm -rf $(SERVER_OUT)


.PHONY: deb
deb:	## Build Debian package (deb)
	@goreleaser -f "${PROJECT_ROOT}build/package/.goreleaser.yml"  --snapshot --skip-publish --rm-dist


# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.DEFAULT_GOAL := help
