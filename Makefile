include ./Makefile.Common

RUN_CONFIG?=local/config.yaml
CMD?=
OTEL_VERSION=main
OTEL_STABLE_VERSION=main

VERSION=$(shell git describe --always --match "v[0-9]*" HEAD)

MOD_NAME=github.com/kyma-project/opentelemetry-collector-components

GROUP ?= all
FOR_GROUP_TARGET=for-$(GROUP)-target

FIND_MOD_ARGS=-type f -name "go.mod"
TO_MOD_DIR=dirname {} \; | sort | grep -E '^./'
EX_COMPONENTS=-not -path "./receiver/*" -not -path "./processor/*" -not -path "./exporter/*" -not -path "./extension/*" -not -path "./connector/*"
EX_INTERNAL=-not -path "./internal/*"
EX_CMD=-not -path "./cmd/*"

# NONROOT_MODS includes ./* dirs (excludes . dir)
NONROOT_MODS := $(shell find . $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )

RECEIVER_MODS := $(shell find ./receiver/* $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )
CMD_MODS := $(shell find ./cmd/* $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )
OTHER_MODS := $(shell find . $(EX_COMPONENTS) $(EX_INTERNAL) $(EX_CMD) $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) ) $(PWD)
ALL_MODS := $(RECEIVER_MODS) $(CMD_MODS) $(OTHER_MODS)


.DEFAULT_GOAL := all

all-modules:
	@echo $(NONROOT_MODS) | tr ' ' '\n' | sort

all-groups:
	@echo "receiver: $(RECEIVER_MODS)"
	@echo "cmd: $(CMD_MODS)"
	@echo "other: $(OTHER_MODS)"

.PHONY: all
all: install-tools all-common gotest

.PHONY: all-common
all-common:
	@$(MAKE) $(FOR_GROUP_TARGET) TARGET="common"

.PHONY: gotidy
gotidy:
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="tidy"

.PHONY: gomoddownload
gomoddownload:
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="moddownload"

.PHONY: gotest
gotest:
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="test"

.PHONY: gotest-with-cover
gotest-with-cover:
	@$(MAKE) $(FOR_GROUP_TARGET) TARGET="test-with-cover"
	$(GOCMD) tool covdata textfmt -i=./coverage/unit -o ./$(GROUP)-coverage.txt

.PHONY: gofmt
gofmt:
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="fmt"

.PHONY: golint
golint:
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="lint"

.PHONY: golintfix
golintfix:
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="lintfix"

.PHONY: for-all
for-all:
	@echo "running $${CMD} in root"
	@$${CMD}
	@set -e; for dir in $(NONROOT_MODS); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

# Define a delegation target for each module
.PHONY: $(ALL_MODS)
$(ALL_MODS):
	@echo "Running target '$(TARGET)' in module '$@' as part of group '$(GROUP)'"
	$(MAKE) --no-print-directory -C $@ $(TARGET)

# Trigger each module's delegation target
.PHONY: for-all-target
for-all-target: $(ALL_MODS)

.PHONY: for-receiver-target
for-receiver-target: $(RECEIVER_MODS)

.PHONY: for-cmd-target
for-cmd-target: $(CMD_MODS)

.PHONY: for-other-target
for-other-target: $(OTHER_MODS)

# Debugging target, which helps to quickly determine whether for-all-target is working or not.
.PHONY: all-pwd
all-pwd:
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="pwd"

.PHONY: run
run:
	cd ./cmd/otelkymacol && GO111MODULE=on $(GOCMD) run --race . --config ../../${RUN_CONFIG} ${RUN_ARGS}

.PHONY: generate
generate: install-tools
	cd ./internal/tools && go install go.opentelemetry.io/collector/cmd/mdatagen
	$(MAKE) for-all CMD="$(GOCMD) generate ./..."
	$(MAKE) gofmt

.PHONY: genotelkymacol
genotelkymacol: $(BUILDER)
	$(BUILDER) --skip-compilation --config cmd/otelkymacol/builder-config.yaml
	$(MAKE) --no-print-directory -C cmd/otelkymacol fmt

# Build the Collector executable.
.PHONY: otelkymacol
otelkymacol:
	cd ./cmd/otelkymacol && GO111MODULE=on CGO_ENABLED=0 $(GOCMD) build -trimpath -o ../../bin/otelkymacol_$(GOOS)_$(GOARCH)$(EXTENSION) -tags $(GO_BUILD_TAGS) .

.PHONY: crosslink
crosslink: $(CROSSLINK)
	@echo "Executing crosslink"
	$(CROSSLINK) --root=$(shell pwd) --prune

.PHONY: clean
clean:
	@echo "Removing coverage files"
	find . -type f -name 'coverage.txt' -delete
	find . -type f -name 'coverage.html' -delete
	find . -type f -name 'coverage.out' -delete

.PHONY: checks
checks:
	$(MAKE) crosslink
	$(MAKE) -j4 gotidy
	git diff --exit-code || (echo 'Some files need committing' &&  git status && exit 1)

.PHONY: check-coverage
check-coverage: $(GO_TEST_COVERAGE) gotest-with-cover
	$(GO_TEST_COVERAGE) --config=./.testcoverage.yml
	  # ( git tag -a $$($$dir | sed s/^.\\///)/$(OCC_VERSION) -m "Release $(OCC_VERSION)" ); \

.PHONY: create-and-push-tags
create-and-push-tags:
	@if [ -z "$(OCC_VERSION)" ]; then \
	  echo "OCC_VERSION is not set"; \
	  exit 1; \
	fi
	@if [ -z "$(REMOTE)" ]; then \
	  echo "REMOTE is not set"; \
	  exit 1; \
	fi
	# execute git tag for every item in $(RECEIVER_MODS)
	@set -e; for dir in $(RECEIVER_MODS); do \
  	  clean_dir=$$(echo $$dir | sed s/^.\\///); \
  	  git tag -a $$clean_dir/v$(OCC_VERSION) -m "Release $(OCC_VERSION)"; \
  	  git push $(REMOTE) $$clean_dir/v$(OCC_VERSION); \
	done
