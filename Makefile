export VERSION := 0.1.0
OWNER ?= elastic
REPO ?= go-licenser
TEST_UNIT_FLAGS ?= -timeout 10s -p 4 -race -cover
TEST_UNIT_PACKAGE ?= ./...
LINT_FOLDERS ?= $(shell go list ./... | sed 's|github.com/$(OWNER)/$(REPO)/||' | grep -v github.com/$(OWNER)/$(REPO))
GOLINT_PRESENT := $(shell command -v golint 2> /dev/null)
GOIMPORTS_PRESENT := $(shell command -v goimports 2> /dev/null)

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

.PHONY: install
install:
	@ go build -o bin/$(REPO)

.PHONY: lint
lint: deps install
	@ golint -set_exit_status $(shell go list ./...)
	@ gofmt -d -e -s $(LINT_FOLDERS)
	@ ./bin/go-licenser -d

.PHONY: format
format: deps install
	@ gofmt -e -w -s $(LINT_FOLDERS)
	@ goimports -w $(LINT_FOLDERS)
	@ ./bin/go-licenser

