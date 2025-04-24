# Get full Makefile directory path
# Src: https://timmurphy.org/2015/09/27/how-to-get-a-makefile-directory-path/
CUR_DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
PROJECT_ROOT := $(shell readlink -f "$(CUR_DIR)"/../..)
LD_FLAGS="-w -s"
VERSION=$(shell git describe --tags --match=v* --always --dirty)

SERVER_OUT := "bin/networkservice"

PROJECT_NAME = $(shell basename "${PROJECT_ROOT}")
PKG := "${PROJECT_ROOT}"
SERVER_PKG_BUILD := "${PKG}/cmd/${PROJECT_NAME}"
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
	@go test -gcflags=all=-l -v -cover -short ${PKG_LIST}

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
	@go test -gcflags=all=-l -race -short ${PKG_LIST}

msan: dep ## Run memory sanitizer
	$(info Running memory sanitizer...)
	@go test -gcflags=all=-l -msan -short ${PKG_LIST}

coverage: ## Generate global code coverage report
	$(info Generating global code coverage report...)
	./coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	$(info Generating global code coverage report in HTML...)
	./coverage.sh html;

dep: ## Get the dependencies
	$(info Get the dependencies...)
	@go get -v -d $(PKG)/...
	@go get -v golang.org/x/lint/golint
	@go mod tidy

server: dep ## Build the binary file for server
	$(info Building server...)
	@go build -v -ldflags=$(LD_FLAGS) -o $(SERVER_OUT) $(SERVER_PKG_BUILD)
	$(info Built server at $(SERVER_OUT) )

clean: ## Remove previous builds
	$(info Running clean up, removing previous builds...)
	@rm -rvf $(SERVER_OUT) 
	@rm -rvf govet-report.out
	@rm -rvf golint-report.out
	@rm -rvf coverage.cov
	@rm -rvf coverage.html
	@rm -rvf coverage_global.txt


.PHONY: deb
deb:	## Build Debian package (deb)
	@goreleaser -f "${PROJECT_ROOT}/build/package/.goreleaser.yml"  --snapshot --skip-publish --clean

.PHONY: vagrant-up
vagrant-up:	## Bring vagrant vm up and running
	$(info Bringing vagrant vm up for ${PROJECT_NAME}...)
	@which vagrant 1>/dev/null || echo "vagrant command not found, install vagrant first!"
	@vagrant destroy -f 1>/dev/null || true
	@vagrant up --provider libvirt --no-provision
	@vagrant provision

.PHONY: vagrant-down
vagrant-down:	## Shutting down vagrant vm
	$(info Shutting down vagrant for ${PROJECT_NAME}...)
	@which vagrant 1>/dev/null || echo "vagrant command not found, install vagrant first!"
	@vagrant halt && vagrant destroy -f

.PHONY: vagrant-testing-all
vagrant-testing-all: vagrant-testing vagrant-testing-race vagrant-testing-msan  ## Run all testing stuff within a vagrant vm

.PHONY: vagrant-testing
vagrant-testing: ## Run unit tests within a vagrant vm
	$(info Project path is "${PROJECT_ROOT}")
	$(info Project name is "${PROJECT_NAME}")
	$(info Makefile dir is "${CUR_DIR}")
	$(info Running unit tests in vagrant box for "${PROJECT_NAME}")
	@which vagrant 1>/dev/null || echo "vagrant command not found, install vagrant first!"
	@vagrant ssh -c "cd networkservice/build/package && sudo -E -H go version"
	@vagrant ssh -c "cd networkservice/build/package && sudo -E -H make dep"
	@vagrant ssh -c "cd networkservice/build/package && sudo -E -H make test"
	@vagrant ssh -c "cd networkservice/build/package && sudo -E -H make coverhtml"

.PHONY: vagrant-testing-race
vagrant-testing-race: ## Run data race detector tests within a vagrant vm
	$(info Project path is "${PROJECT_ROOT}")
	$(info Project name is "${PROJECT_NAME}")
	$(info Makefile dir is "${CUR_DIR}")
	$(info Running data race detector tests in vagrant box for "${PROJECT_NAME}")
	@which vagrant 1>/dev/null || echo "vagrant command not found, install vagrant first!"
	@vagrant ssh -c "cd networkservice/build/package && sudo -E -H make dep"
	@vagrant ssh -c "cd networkservice/build/package && sudo -E -H make race"

.PHONY: vagrant-testing-msan
vagrant-testing-msan: ## Run memory sanitizer tests within a vagrant vm
	$(info Project path is "${PROJECT_ROOT}")
	$(info Project name is "${PROJECT_NAME}")
	$(info Makefile dir is "${CUR_DIR}")
	$(info Running memory sanitizer tests in vagrant box for "${PROJECT_NAME}")
	@which vagrant 1>/dev/null || echo "vagrant command not found, install vagrant first!"
	@vagrant ssh -c "cd networkservice/build/package && sudo -E -H make msan"

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.DEFAULT_GOAL := help
