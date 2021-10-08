PWD=$(shell git rev-parse --show-toplevel)
LOCAL_TMP = $(PWD)/tmp
LOCAL_BIN = $(PWD)/bin
VENDOR_DIR = $(PWD)/vendor
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
RELEASE_REVISION=$(shell git rev-parse HEAD)
MONETR_CLI_PACKAGE = github.com/monetr/monetr/pkg/cmd
COVERAGE_TXT = $(PWD)/coverage.txt


ARCH=amd64
OS=$(shell uname -s | tr A-Z a-z)

ENVIRONMENT ?= $(shell echo $${BUIlDKITE_GITHUB_DEPLOYMENT_ENVIRONMENT:-Local})
ENV_LOWER = $(shell echo $(ENVIRONMENT) | tr A-Z a-z)

GENERATED_YAML=$(PWD)/generated/$(ENV_LOWER)

ifeq ($(NO_CACHE),true)
DOCKER_CACHE=--no-cache
endif

DOCKER_OPTIONS=

ifeq ($(DEBUG),true)
DOCKER_OPTIONS += --debug
endif

ifndef POSTGRES_DB
POSTGRES_DB=postgres
endif

ifndef POSTGRES_USER
POSTGRES_USER=postgres
endif

ifndef POSTGRES_HOST
POSTGRES_HOST=localhost
endif

GREEN=\033[0;32m
YELLOW=\033[1;33m
RESET=\033[0m

# Just a shorthand to print some colored text, makes it easier to read and tell the developer what all the makefile is
# doing since its doing a ton.
ifndef BUILDKITE
define infoMsg
	@echo "$(GREEN)[$@] $(1)$(RESET)"
endef

define warningMsg
	@echo "$(YELLOW)[$@] $(1)$(RESET)"
endef
else
define infoMsg
	@echo "INFO [$@] $(1)"
endef

define warningMsg
	@echo "WARN [$@] $(1)"
endef
endif

GO_SRC_DIR=$(PWD)/pkg
ALL_GO_FILES=$(filter-out $(GO_SRC_DIR)/ui/static, $(wildcard $(GO_SRC_DIR)/**/*.go))
APP_GO_FILES=$(filter-out $(GO_SRC_DIR)/**/*_test.go, $(ALL_GO_FILES))
TEST_GO_FILES=$(wildcard $(GO_SRC_DIR)/**/*_test.go)

UI_SRC_DIR=$(PWD)/ui
ALL_UI_FILES=$(shell find $(UI_SRC_DIR) -type f)
APP_UI_FILES=$(filter-out *.spec.*, $(ALL_UI_FILES))
TEST_UI_FILES=$(shell find $(UI_SRC_DIR) -type f -name '*.spec.*')

GO_DEPS=$(PWD)/go.mod $(PWD)/go.sum
UI_DEPS=$(PWD)/package.json

include $(PWD)/scripts/*.mk

default: build

HASH_DIR=$(PWD)/tmp/hashes
$(HASH_DIR):
	mkdir -p $(HASH_DIR)


$(NODE_MODULES)-install:
	yarn install -d

dependencies: $(GO) $(GO_DEPS)
	$(call infoMsg,Installing dependencies for monetrs rest-api)
	$(GO) get $(GO_SRC_DIR)/...

PATH+=\b:$(NODE_MODULES)/.bin


define hash
md5sum $1 | cut -d " " -f 1
endef


NODE_MODULES=$(PWD)/node_modules
$(NODE_MODULES): $(HASH_DIR)
$(NODE_MODULES): NODE_MODULES_HASH=$(HASH_DIR)/$(shell md5sum ** $(NODE_MODULES)/**/** 2>/dev/null | $(call hash,-))
$(NODE_MODULES):
	@echo "Node Modules Hash: $(STATIC_HASH)"
	@if [ ! -f "$(NODE_MODULES_HASH)" ]; then make $(NODE_MODULES)-install && touch $(NODE_MODULES_HASH); fi

$(NODE_MODULES)-install:
	yarn install -d

STATIC_DIR=$(GO_SRC_DIR)/ui/static
$(STATIC_DIR): $(NODE_MODULES) $(HASH_DIR)
$(STATIC_DIR): UI_HASH=$(shell md5sum ** $(UI_SRC_DIR)/**/** 2>/dev/null | $(call hash,-))
$(STATIC_DIR): NODE_MODULES_HASH=$(shell md5sum ** $(NODE_MODULES)/**/** 2>/dev/null | $(call hash,-))
$(STATIC_DIR): STATIC_HASH=$(HASH_DIR)/$(shell echo "$(UI_HASH)|$(NODE_MODULES_HASH)" | $(call hash,-))
$(STATIC_DIR):
	@echo "Static Files Hash: $(STATIC_HASH)"
	@if [ ! -f "$(STATIC_HASH)" ]; then make $(STATIC_DIR)-build && touch $(STATIC_HASH); fi

$(STATIC_DIR)-build:
	-rm -rf $(GO_SRC_DIR)/ui/static
	RELEASE_REVISION=$(RELEASE_REVISION) yarn build-dev

build-ui: $(STATIC_DIR)

build: $(GO) $(STATIC_DIR) dependencies $(APP_GO_FILES)
	$(call infoMsg,Generating anything needed for binary)
	$(GO) install golang.org/x/tools/cmd/stringer
	$(call infoMsg,Building monetr binary)
	$(GO) build -o $(LOCAL_BIN)/monetr $(MONETR_CLI_PACKAGE)

test: $(GO) dependencies $(ALL_GO_FILES) $(GOTESTSUM)
	$(call infoMsg,Running go tests for monetr rest-api)
ifndef CI
	$(GO) run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST)
endif
	$(GOTESTSUM) --junitfile $(PWD)/rest-api.xml --format testname -- -race -v \
		-coverprofile=$(COVERAGE_TXT) \
		-covermode=atomic $(GO_SRC_DIR)/...
	$(GO) tool cover -func=$(COVERAGE_TXT)

clean:
	-rm -rf $(LOCAL_BIN)
	-rm -rf $(COVERAGE_TXT)
	-rm -rf $(NODE_MODULES)
	-rm -rf $(VENDOR_DIR)
	-rm -rf $(LOCAL_TMP)
	-rm -rf $(PWD)/generated
	-rm -rf $(PWD)/docs
	-rm -rf $(PWD)/Notes.md
	-rm -rf $(GO_SRC_DIR)/ui/static

docs: $(SWAG) $(APP_GO_FILES)
	$(SWAG) init -d $(GO_SRC_DIR)/controller -g controller.go \
		--parseDependency \
		--parseDepth 5 \
		--parseInternal \
		--output $(PWD)/docs

docs-local: docs
	$(PWD)/node_modules/.bin/redoc-cli serve $(PWD)/docs/swagger.yaml

docker: Dockerfile.monolith build
	docker $(DOCKER_OPTIONS) build $(DOCKER_CACHE) \
		--build-arg REVISION=$(RELEASE_REVISION) \
		-t monetr -f $(PWD)/Dockerfile.monetr $(PWD)

docker-work-web-ui:
	docker build -t workwebui -f Dockerfile.work .

ifdef GITHUB_TOKEN
license: $(LICENSE) build
	$(call infoMsg,Checking dependencies for open source licenses)
	-$(LICENSE) $(PWD)/licenses.hcl $(LOCAL_BIN)/monetr
else
.PHONY: license
license:
	$(call warningMsg,GITHUB_TOKEN is required to check licenses)
endif

VALUES_FILE=$(PWD)/values.$(ENV_LOWER).yaml
VALUES_FILES=$(PWD)/values.yaml $(VALUES_FILE)

TEMPLATE_FILES=$(PWD)/templates/*

$(GENERATED_YAML): $(VALUES_FILES) $(TEMPLATE_FILES)
$(GENERATED_YAML): IMAGE_TAG=$(shell git rev-parse HEAD)
$(GENERATED_YAML): $(HELM) $(SPLIT_YAML)
	$(call infoMsg,Generating Kubernetes yaml using Helm output to:  $(GENERATED_YAML))
	$(call infoMsg,Environment:                                      $(ENVIRONMENT))
	$(call infoMsg,Using values file:                                $(VALUES_FILE))
	-rm -rf $(GENERATED_YAML)
	-mkdir -p $(GENERATED_YAML)
	$(HELM) template rest-api $(PWD) \
		--dry-run \
		--set image.tag="$(IMAGE_TAG)" \
		--set podAnnotations."monetr\.dev/date"="$(BUILD_TIME)" \
		--set podAnnotations."monetr\.dev/sha"="$(IMAGE_TAG)" \
		--values=values.$(ENV_LOWER).yaml | $(SPLIT_YAML) --outdir $(GENERATED_YAML) -

generate: $(GENERATED_YAML)

ifdef GITLAB_CI
include $(PWD)/Makefile.gitlab-ci
include $(PWD)/Makefile.deploy
endif

ifdef GITHUB_ACTION
include $(PWD)/Makefile.github-actions
endif

# PostgreSQL tests currently only work in CI pipelines.
ifdef CI
PG_TEST_EXTENSION_QUERY = "CREATE EXTENSION pgtap;"
JUNIT_OUTPUT_FILE=/junit.xml
pg_test:
	@for FILE in $(PWD)/schema/*.up.sql; do \
		echo "Applying $$FILE"; \
  		psql -q -d $(POSTGRES_DB) -U $(POSTGRES_USER) -h $(POSTGRES_HOST) -f $$FILE || exit 1; \
  	done;
	psql -q -d $(POSTGRES_DB) -U $(POSTGRES_USER) -h $(POSTGRES_HOST) -c $(PG_TEST_EXTENSION_QUERY)
	-JUNIT_OUTPUT_FILE=$(JUNIT_OUTPUT_FILE) pg_prove -h $(POSTGRES_HOST) -U $(POSTGRES_USER) -d $(POSTGRES_DB) -f -c $(PWD)/tests/pg/*.sql --verbose --harness TAP::Harness::JUnit
endif

include $(PWD)/Makefile.release
include $(PWD)/Makefile.docker

ifndef CI
include $(PWD)/Makefile.tinker

ifeq ($(ENV_LOWER),local)
include $(PWD)/Makefile.local
endif

endif

ifndef POSTGRES_PORT
POSTGRES_PORT=5432
endif

migrate: $(GO)
	@$(GO) run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST) -P $(POSTGRES_PORT) -W $(POSTGRES_PASSWORD)

beta-code: $(GO) migrate
	@$(GO) run $(MONETR_CLI_PACKAGE) beta new-code -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST) -P $(POSTGRES_PORT) -W $(POSTGRES_PASSWORD)

all: build test generate lint
