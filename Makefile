DIST_DIR ?= dist/
GOOS ?= $(shell uname -s | tr "[:upper:]" "[:lower:]")
ARCH ?= $(shell uname -m)
ifeq ($(ARCH),x86_64)
GOARCH             := amd64
else
GOARCH             := $(ARCH)  # no idea if this works for other platforms....
endif
BUILDINFOSDET ?=
PROGRAM_ARGS ?=

PROJECT_VERSION           := 0.0.4
DOCKER_REPO               := synfinatic
PROJECT_NAME              := rss-tool
PROJECT_TAG               := $(shell git describe --tags 2>/dev/null $(git rev-list --tags --max-count=1))
ifeq ($(PROJECT_TAG),)
PROJECT_TAG               := NO-TAG
endif
PROJECT_COMMIT            := $(shell git rev-parse HEAD)
ifeq ($(PROJECT_COMMIT),)
PROJECT_COMMIT            := NO-CommitID
endif
PROJECT_DELTA             := $(shell DELTA_LINES=$$(git diff | wc -l); if [ $${DELTA_LINES} -ne 0 ]; then echo $${DELTA_LINES} ; else echo "''" ; fi)
VERSION_PKG               := $(shell echo $(PROJECT_VERSION) | sed 's/^v//g')
LICENSE                   := GPLv3
URL                       := https://github.com/$(DOCKER_REPO)/$(PROJECT_NAME)
DESCRIPTION               := RSS Download Tool
BUILDINFOS                := $(shell date +%FT%T%z)$(BUILDINFOSDET)
HOSTNAME                  := $(shell hostname)
OUTPUT_NAME               := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-$(GOOS)-$(GOARCH)  # default for current platform
OUTPUT_NAME               := $(DIST_DIR)$(PROJECT_NAME)  # default for current platform
# supported platforms for `make release`
WINDOWS_BIN               := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-windows-amd64.exe
WINDOWS32_BIN             := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-windows-386.exe
LINUX_BIN                 := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-linux-amd64
LINUXARM64_BIN            := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-linux-arm64
LINUXARM32_BIN            := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-linux-arm32
DARWIN_BIN                := $(DIST_DIR)$(PROJECT_NAME)-$(PROJECT_VERSION)-darwin-amd64



ALL: $(OUTPUT_NAME) ## Build binary.  Needs to be a supported plaform as defined above

include help.mk  # place after ALL target and before all other targets

release: build-release
	cd dist && shasum -a 256 * | gpg --clear-sign >release.sig

# windows & windows32 not supported currently
build-release: clean linux linux-arm32 linux-arm64 darwin ## Build all our release binaries

.PHONY: run
run: ./cmd/rss-tool/*.go  ## build and run cria using $PROGRAM_ARGS
	go run ./cmd/rss-tool/... $(PROGRAM_ARGS)

clean-all: clean ## clean _everything_

clean: ## Remove all binaries in dist
	rm -f dist/*

clean-go: ## Clean Go cache
	go clean -i -r -cache -modcache

go-get:  ## Get our go modules
	go get -v all

.PHONY: build-race
build-race: .prepare ## Build race detection binary
	go build -race -ldflags='$(LDFLAGS)' -o $(OUTPUT_NAME) ./cmd/rss-tool/...

debug: .prepare ## Run debug in dlv
	dlv debug ./cmd/rss-tool/...

.PHONY: unittest
unittest: ## Run go unit tests
	go test ./...

.PHONY: test-race
test-race: ## Run `go test -race` on the code
	@echo checking code for races...
	go test -race ./...

.PHONY: vet
vet: ## Run `go vet` on the code
	@echo checking code is vetted...
	go vet $(shell go list ./...)

test: vet unittest ## Run all tests

.prepare: $(DIST_DIR)

$(DIST_DIR):
	@if test -d $(DIST_DIR); then mkdir -p $(DIST_DIR) ; fi

.PHONY: fmt
fmt: ## Format Go code
	@go fmt ./cmd

.PHONY: test-fmt
test-fmt: fmt ## Test to make sure code if formatted correctly
	@if test `git diff ./cmd | wc -l` -gt 0; then \
	    echo "Code changes detected when running 'go fmt':" ; \
	    git diff -Xfiles ; \
	    exit -1 ; \
	fi

.PHONY: test-tidy
test-tidy:  ## Test to make sure go.mod is tidy
	@go mod tidy
	@if test `git diff go.mod | wc -l` -gt 0; then \
	    echo "Need to run 'go mod tidy' to clean up go.mod" ; \
	    exit -1 ; \
	fi

precheck: test test-fmt test-tidy lint ## Run all tests that happen in a PR 

lint:  ## Run golangci-lint
	golangci-lint run

# Build targets for our supported plaforms
windows: $(WINDOWS_BIN)  ## Build 64bit Windows binary

$(WINDOWS_BIN): $(wildcard */*.go) .prepare
	GOARCH=amd64 GOOS=windows go build -ldflags='$(LDFLAGS)' -o $(WINDOWS_BIN) ./cmd/rss-tool/...
	@echo "Created: $(WINDOWS_BIN)"

windows32: $(WINDOWS32_BIN)  ## Build 32bit Windows binary

$(WINDOWS32_BIN): $(wildcard */*.go) .prepare
	GOARCH=386 GOOS=windows go build -ldflags='$(LDFLAGS)' -o $(WINDOWS32_BIN) ./cmd/rss-tool/...
	@echo "Created: $(WINDOWS32_BIN)"

linux: $(LINUX_BIN)  ## Build Linux/x86_64 binary

$(LINUX_BIN): $(wildcard */*.go) .prepare
	GOARCH=amd64 GOOS=linux go build -ldflags='$(LDFLAGS)' -o $(LINUX_BIN) ./cmd/rss-tool/...
	@echo "Created: $(LINUX_BIN)"

linux-arm64: $(LINUXARM64_BIN)  ## Build Linux/arm64 binary

$(LINUXARM64_BIN): $(wildcard */*.go) .prepare
	GOARCH=arm64 GOOS=linux go build -ldflags='$(LDFLAGS)' -o $(LINUXARM64_BIN) ./cmd/rss-tool/...
	@echo "Created: $(LINUXARM64_BIN)"

linux-arm32: $(LINUXARM32_BIN)  ## Build Linux/arm64 binary

$(LINUXARM32_BIN): $(wildcard */*.go) .prepare
	GOARCH=arm GOOS=linux go build -ldflags='$(LDFLAGS)' -o $(LINUXARM32_BIN) ./cmd/rss-tool/...
	@echo "Created: $(LINUXARM32_BIN)"

darwin: $(DARWIN_BIN)  ## Build MacOS/x86_64 binary

$(DARWIN_BIN): $(wildcard */*.go) .prepare
	GOARCH=amd64 GOOS=darwin go build -ldflags='$(LDFLAGS)' -o $(DARWIN_BIN) ./cmd/rss-tool/...
	@echo "Created: $(DARWIN_BIN)"

$(OUTPUT_NAME): $(wildcard */*.go) .prepare
	go build -ldflags='$(LDFLAGS)' -o $(OUTPUT_NAME) ./cmd/rss-tool/...
	@echo "Created: $(OUTPUT_NAME)"
