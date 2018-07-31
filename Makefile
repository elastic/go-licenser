export VERSION := 0.1.0
OWNER ?= elastic
REPO ?= go-licenser
TEST_UNIT_FLAGS ?= -timeout 10s -p 4 -race -cover
TEST_UNIT_PACKAGE ?= ./...
LINT_FOLDERS ?= $(shell go list ./... | sed 's|github.com/$(OWNER)/$(REPO)/||' | grep -v github.com/$(OWNER)/$(REPO))
GOLINT_PRESENT := $(shell command -v golint 2> /dev/null)
GOIMPORTS_PRESENT := $(shell command -v goimports 2> /dev/null)

define HELP
/////////////////////////////////////////
/\t$(REPO) Makefile \t\t/
/////////////////////////////////////////

## Build target

- build:                  It will build $(REPO) for the current architecture in bin/$(REPO).
- install:                It will install $(REPO) in the current system (by default in $(GOPATH)/bin/$(REPO)).

## Development targets

- deps:                   It will install the dependencies required to run developemtn targets.
- unit:                   Runs the unit tests.
- lint:                   Runs the linters.
- format:                 Formats the source files according to gofmt, goimports and go-licenser.
- update-golden-files:    Updates the test golden files.

endef
export HELP

.DEFAULT: help
.PHONY: help
help:
	@ echo "$$HELP"

.PHONY: deps
deps:
ifndef GOLINT_PRESENT
	@ go get -u golang.org/x/lint/golint
endif
ifndef GOIMPORTS_PRESENT
	@ go get -u golang.org/x/tools/cmd/goimports
endif

.PHONY: update-golden-files
update-golden-files:
	$(eval GOLDEN_FILE_PACKAGES := "github.com/$(OWNER)/$(REPO)")
	@ go test $(GOLDEN_FILE_PACKAGES) -update

.PHONY: unit
unit:
	@ go test $(TEST_UNIT_FLAGS) $(TEST_UNIT_PACKAGE)

.PHONY: build
build:
	@ go build -o bin/$(REPO)

.PHONY: install
install:
	@ go install

.PHONY: lint
lint: deps build
	@ golint -set_exit_status $(shell go list ./...)
	@ gofmt -d -e -s $(LINT_FOLDERS)
	@ ./bin/go-licenser -d -exclude golden

.PHONY: format
format: deps build
	@ gofmt -e -w -s $(LINT_FOLDERS)
	@ goimports -w $(LINT_FOLDERS)
	@ ./bin/go-licenser

