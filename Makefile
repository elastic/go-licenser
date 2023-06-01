export VERSION := v0.4.0
export GOBIN = $(shell pwd)/bin
OWNER ?= elastic
REPO ?= go-licenser
TEST_UNIT_FLAGS ?= -timeout 10s -p 4 -race -cover
TEST_UNIT_PACKAGE ?= ./...
RELEASED = $(shell git tag -l $(VERSION))
DEFAULT_LDFLAGS ?= -X main.version=$(VERSION)-dev -X main.commit=$(shell git rev-parse HEAD)
VERSION_STATICCHECK = 2023.1.3
VERSION_GOIMPORT = v0.1.12
VERSION_GORELEASER:=v0.184.0

define HELP
/////////////////////////////////////////
/\t$(REPO) Makefile \t\t/
/////////////////////////////////////////

## Build target

- build:                  It will build $(REPO) for the current architecture in bin/$(REPO).
- install:                It will install $(REPO) in the current system (by default in $(GOPATH)/bin/$(REPO)).

## Development targets

- unit:                   Runs the unit tests.
- lint:                   Runs the linters.
- format:                 Formats the source files according to gofmt, goimports and go-licenser.
- update-golden-files:    Updates the test golden files.

## Release targets

- release:                Creates and publishes a new release matching the VERSION variable.
- snapshot:               Creates a snapshot locally in the dist/ folder.

endef
export HELP

.DEFAULT: help
.PHONY: help
help:
	@ echo "$$HELP"

.PHONY: update-golden-files
update-golden-files:
	$(eval GOLDEN_FILE_PACKAGES := "github.com/$(OWNER)/$(REPO)")
	@ go test $(GOLDEN_FILE_PACKAGES) -update

.PHONY: unit
unit:
	@ go test $(TEST_UNIT_FLAGS) $(TEST_UNIT_PACKAGE)

.PHONY: build
build:
	@ go build -o bin/$(REPO) -ldflags="$(DEFAULT_LDFLAGS)"

.PHONY: install
install:
	@ go install

.PHONY: lint
lint: build
	@ go run honnef.co/go/tools/cmd/staticcheck@$(VERSION_STATICCHECK)
	@ gofmt -d -e -s .
	@ $(GOBIN)/go-licenser -d -exclude golden

.PHONY: format
format: build
	@ gofmt -e -w -s .
	@ go run golang.org/x/tools/cmd/goimports@$(VERSION_GOIMPORT) -w .
	@ $(GOBIN)/go-licenser -exclude golden

.PHONY: release
release:
	@ echo "-> Releasing $(REPO) $(VERSION)..."
	@ git fetch upstream
ifeq ($(strip $(RELEASED)),)
	@ echo "-> Creating and pushing a new tag $(VERSION)..."
	@ git tag $(VERSION)
	@ git push upstream $(VERSION)
	@ go run github.com/goreleaser/goreleaser@$(VERSION_GORELEASER) release --skip-validate --rm-dist
else
	@ echo "-> git tag $(VERSION) already present, skipping release..."
endif

.PHONY: snapshot
snapshot:
	@ echo "-> Snapshotting $(REPO) $(VERSION)..."
	@ go run github.com/goreleaser/goreleaser@$(VERSION_GORELEASER) release --snapshot --rm-dist
