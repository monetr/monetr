PWD=$(shell git rev-parse --show-toplevel)
LOCAL_TMP = $(PWD)/tmp
LOCAL_BIN = $(PWD)/bin
VENDOR_DIR = $(PWD)/vendor
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
RELEASE_REVISION=$(shell git rev-parse HEAD)
RELEASE_VERSION ?= $(shell git describe --tags `git rev-list --tags --max-count=1`)
MONETR_CLI_PACKAGE = github.com/monetr/monetr/pkg/cmd
COVERAGE_TXT = $(PWD)/coverage.txt

KUBERNETES_VERSION=1.18.5

ARCH ?= amd64
OS ?= $(shell uname -s | tr A-Z a-z)

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
ALL_GO_FILES=$(shell find $(GO_SRC_DIR) -type f -name '*.go')
APP_GO_FILES=$(filter-out *_test.go, $(ALL_GO_FILES))
TEST_GO_FILES=$(shell find $(GO_SRC_DIR) -type f -name '*_test.go')

UI_SRC_DIR=$(PWD)/ui
ALL_UI_FILES=$(shell find $(UI_SRC_DIR) -type f)
APP_UI_FILES=$(filter-out *.spec.*, $(ALL_UI_FILES))
TEST_UI_FILES=$(shell find $(UI_SRC_DIR) -type f -name '*.spec.*')

GO_DEPS=$(PWD)/go.mod $(PWD)/go.sum
UI_DEPS=$(PWD)/package.json $(PWD)/yarn.lock

include $(PWD)/scripts/*.mk

default: build

HASH_DIR=$(PWD)/tmp/hashes
$(HASH_DIR):
	mkdir -p $(HASH_DIR)


$(NODE_MODULES)-install:
	yarn install -d


#PATH+=\b:$(NODE_MODULES)/.bin


define hash
md5sum $1 | cut -d " " -f 1
endef


NODE_MODULES=$(PWD)/node_modules
$(NODE_MODULES): $(UI_DEPS)
	yarn install
	touch -a -m $(NODE_MODULES) # Dumb hack to make sure the node modules directory timestamp gets bumpbed for make.

STATIC_DIR=$(GO_SRC_DIR)/ui/static
PUBLIC_FILES=$(PWD)/public/favicon.ico $(PWD)/public/logo192.png $(PWD)/public/logo512.png $(PWD)/public/manifest.json $(PWD)/public/robots.txt
$(STATIC_DIR): $(APP_UI_FILES) $(NODE_MODULES) $(PUBLIC_FILES) $(PWD)/tsconfig.json $(PWD)/webpack.config.js
$(STATIC_DIR): YARN_BIN=$(shell yarn bin)
$(STATIC_DIR):
	$(call infoMsg,Building UI files)
	git clean -f -X $(STATIC_DIR)
	RELEASE_VERSION=$(RELEASE_VERSION) RELEASE_REVISION=$(RELEASE_REVISION) $(YARN_BIN)/webpack --mode production
	cp $(PWD)/public/favicon.ico $(STATIC_DIR)/favicon.ico
	cp $(PWD)/public/logo192.png $(STATIC_DIR)/logo192.png
	cp $(PWD)/public/logo512.png $(STATIC_DIR)/logo512.png
	cp $(PWD)/public/manifest.json $(STATIC_DIR)/manifest.json
	cp $(PWD)/public/robots.txt $(STATIC_DIR)/robots.txt

GOMODULES=$(GOPATH)/pkg/mod
$(GOMODULES): $(GO) $(GO_DEPS)
	$(call infoMsg,Installing dependencies for monetrs rest-api)
	$(GO) get -t $(GO_SRC_DIR)/...
	touch -a -m $(GOMODULES)

go-dependencies: $(GOMODULES)

ui-dependencies: $(NODE_MODULES)

dependencies: $(GOMODULES) $(NODE_MODULES)

build-ui: $(STATIC_DIR)

GOOS ?= $(OS)
GOARCH ?= amd64

ifeq ($(GOOS),windows)
BINARY_FILE_NAME=monetr.exe
else
BINARY_FILE_NAME=monetr
endif

BINARY=$(LOCAL_BIN)/$(BINARY_FILE_NAME)
$(BINARY): $(GO) $(APP_GO_FILES)
ifndef CI
$(BINARY): $(STATIC_DIR) $(GOMODULES)
endif
	$(call infoMsg,Building monetr binary for: $(GOOS)/$(GOARCH))
	$(GO) build -o $(BINARY) $(MONETR_CLI_PACKAGE)

BUILD_DIR=$(PWD)/build
$(BUILD_DIR):
	mkdir -p $(PWD)/build

CONTAINER_BINARY=$(BUILD_DIR)/monetr
$(CONTAINER_BINARY): $(BUILD_DIR) $(GO) $(STATIC_DIR) $(GOMODULES) $(APP_GO_FILES)
	$(call infoMsg,Building monetr binary for container)
	GOOS=linux GOARCH=$(ARCH) $(GO) build -o $(CONTAINER_BINARY) $(MONETR_CLI_PACKAGE)

build: $(BINARY)

BINARY_TAR=$(PWD)/bin/monetr-$(RELEASE_VERSION)-$(GOOS)-$(GOARCH).tar.gz
$(BINARY_TAR): $(BINARY)
$(BINARY_TAR): TAR=$(shell which tar)
$(BINARY_TAR):
	cd $(LOCAL_BIN) && $(TAR) -czf $(BINARY_TAR) $(BINARY_FILE_NAME)

tar: $(BINARY_TAR)

ifdef GITHUB_ACTION
release-asset: $(BINARY_TAR)
release-asset: GH=$(shell which gh)
release-asset:
	$(GH) release upload $(RELEASE_VERSION) $(BINARY_TAR) --clobber
endif


test-go: $(GO) $(GOMODULES) $(ALL_GO_FILES) $(GOTESTSUM)
	$(call infoMsg,Running go tests for monetr REST API)
	$(GO) run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST)
	$(GOTESTSUM) --junitfile $(PWD)/rest-api-junit.xml --jsonfile $(PWD)/rest-api-tests.json --format testname -- \
		-race -v \
		-coverprofile=$(COVERAGE_TXT) \
		-covermode=atomic $(GO_SRC_DIR)/...
	$(GO) tool cover -func=$(COVERAGE_TXT)

test-ui: $(ALL_UI_FILES) $(NODE_MODULES)
	$(call infoMsg,Running go tests for monetrs UI)
	yarn test

test: test-go test-ui

clean:
	-rm -rf $(LOCAL_BIN)
	-rm -rf $(COVERAGE_TXT)
	-rm -rf $(NODE_MODULES)
	-rm -rf $(VENDOR_DIR)
	-rm -rf $(LOCAL_TMP)
	-rm -rf $(PWD)/generated
	-rm -rf $(PWD)/docs
	-rm -rf $(PWD)/build
	-rm -rf $(PWD)/Notes.md
	-git clean -f -X $(STATIC_DIR)


docs: $(SWAG) $(APP_GO_FILES)
	$(SWAG) init -d $(GO_SRC_DIR)/controller -g controller.go \
		--parseDependency \
		--parseDepth 5 \
		--parseInternal \
		--output $(PWD)/docs
	cp $(PWD)/public/favicon.ico $(PWD)/docs/favicon.ico
	cp $(PWD)/public/logo192.png $(PWD)/docs/logo192.png
	cp $(PWD)/public/logo512.png $(PWD)/docs/logo512.png
	cp $(PWD)/public/manifest.json $(PWD)/docs/manifest.json

docs-local: docs
	$(PWD)/node_modules/.bin/redoc-cli serve $(PWD)/docs/swagger.yaml

CONTAINER=$(BUILD_DIR)/monetr.container.tar
$(CONTAINER): $(BUILD_DIR) $(PWD)/Dockerfile $(PWD)/.dockerignore $(APP_GO_FILES) $(STATIC_DIR)
	docker buildx $(DOCKER_OPTIONS) build $(DOCKER_CACHE) \
		--build-arg REVISION=$(RELEASE_REVISION) \
		--output type=tar,dest=$(CONTAINER) \
		-t monetr -f $(PWD)/Dockerfile .

docker: $(CONTAINER)

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

CHART_FILE=$(PWD)/Chart.yaml
VALUES_FILE=$(PWD)/values.$(ENV_LOWER).yaml
VALUES_FILES=$(PWD)/values.yaml $(VALUES_FILE)

TEMPLATE_FILES=$(PWD)/templates/*

$(GENERATED_YAML): $(CHART_FILE) $(VALUES_FILES) $(TEMPLATE_FILES)
$(GENERATED_YAML): IMAGE_TAG=$(shell git rev-parse HEAD)
$(GENERATED_YAML): $(HELM) $(SPLIT_YAML)
	$(call infoMsg,Generating Kubernetes yaml using Helm output to:  $(GENERATED_YAML))
	$(call infoMsg,Environment:                                      $(ENVIRONMENT))
	$(call infoMsg,Using values file:                                $(VALUES_FILE))
	-rm -rf $(GENERATED_YAML)
	-mkdir -p $(GENERATED_YAML)
	$(HELM) template monetr $(PWD) \
		--dry-run \
		--set image.tag="$(IMAGE_TAG)" \
		--set podAnnotations."monetr\.dev/date"="$(BUILD_TIME)" \
		--set podAnnotations."monetr\.dev/sha"="$(IMAGE_TAG)" \
		--values=values.$(ENV_LOWER).yaml | $(SPLIT_YAML) --outdir $(GENERATED_YAML) -

generate: $(GENERATED_YAML)

ifdef GITHUB_ACTION
include $(PWD)/Makefile.github-actions
endif

# PostgreSQL tests currently only work in CI pipelines.
ifdef CI
PG_TEST_EXTENSION_QUERY = "CREATE EXTENSION pgtap;"
JUNIT_OUTPUT_FILE=$(PWD)/postgres-junit.xml
pg_test:
	@for FILE in $(PWD)/schema/*.up.sql; do \
		echo "Applying $$FILE"; \
  		psql -q -d $(POSTGRES_DB) -U $(POSTGRES_USER) -h $(POSTGRES_HOST) -f $$FILE || exit 1; \
  	done;
	psql -q -d $(POSTGRES_DB) -U $(POSTGRES_USER) -h $(POSTGRES_HOST) -c $(PG_TEST_EXTENSION_QUERY)
	-JUNIT_OUTPUT_FILE=$(JUNIT_OUTPUT_FILE) pg_prove \
		-h $(POSTGRES_HOST) \
		-U $(POSTGRES_USER) \
		-d $(POSTGRES_DB) -f \
		-c $(PWD)/tests/pg/*.sql --verbose --harness TAP::Harness::JUnit
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
