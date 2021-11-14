PWD=$(shell git rev-parse --show-toplevel)
LOCAL_TMP = $(PWD)/tmp
LOCAL_BIN = $(PWD)/bin
BUILD_DIR = $(PWD)/build
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
RELEASE_REVISION=$(shell git rev-parse HEAD)
RELEASE_VERSION ?= $(shell git describe --tags `git rev-list --tags --max-count=1`)
CONTAINER_VERSION = $(subst v,,$(RELEASE_VERSION))
MONETR_CLI_PACKAGE = github.com/monetr/monetr/pkg/cmd
COVERAGE_TXT = $(PWD)/coverage.txt

KUBERNETES_VERSION=1.18.5

ifeq ($(OS),Windows_NT)
	OS=windows
	# I'm not sure how arm64 will show up on windows. I also have no idea how this makefile would even work on windows.
	# It probably wouldn't.
	ARCH ?= amd64
else
	OS ?= $(shell uname -s | tr A-Z a-z)
    UNAME_P := $(shell uname -p)
    ifeq ($(UNAME_P),x86_64)
		ARCH=amd64
    endif
    ifneq ($(filter %86,$(UNAME_P)),)
    	# This can happen on macOS with Intel CPUs, we get an i386 arch.
		ARCH=amd64
    endif
    ifneq ($(filter arm%,$(UNAME_P)),)
        ARCH=arm64
    endif
endif
# If we still didn't figure out the architecture, then just default to amd64
ARCH ?= amd64

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

SOURCE_MAP_DIR=$(BUILD_DIR)/source_maps
$(SOURCE_MAP_DIR):
	mkdir -p $(SOURCE_MAP_DIR)

NODE_MODULES=$(PWD)/node_modules
$(NODE_MODULES): $(UI_DEPS)
	yarn install
	touch -a -m $(NODE_MODULES) # Dumb hack to make sure the node modules directory timestamp gets bumpbed for make.

WEBPACK=$(word 1,$(wildcard $(YARN_BIN)/webpack) $(NODE_MODULES)/.bin/webpack)
$(WEBPACK): $(NODE_MODULES)

STATIC_DIR=$(GO_SRC_DIR)/ui/static
PUBLIC_FILES=$(PWD)/public/favicon.ico $(PWD)/public/logo192.png $(PWD)/public/logo512.png $(PWD)/public/manifest.json $(PWD)/public/robots.txt
UI_CONFIG_FILES=$(PWD)/tsconfig.json $(PWD)/webpack.config.js
$(STATIC_DIR): $(APP_UI_FILES) $(NODE_MODULES) $(PUBLIC_FILES) $(UI_CONFIG_FILES) $(WEBPACK) $(SOURCE_MAP_DIR)
	$(call infoMsg,Building UI files)
	git clean -f -X $(STATIC_DIR)
	RELEASE_VERSION=$(RELEASE_VERSION) RELEASE_REVISION=$(RELEASE_REVISION) $(WEBPACK) --mode production
	cp $(PWD)/public/favicon.ico $(STATIC_DIR)/favicon.ico
	cp $(PWD)/public/logo192.png $(STATIC_DIR)/logo192.png
	cp $(PWD)/public/logo512.png $(STATIC_DIR)/logo512.png
	cp $(PWD)/public/manifest.json $(STATIC_DIR)/manifest.json
	cp $(PWD)/public/robots.txt $(STATIC_DIR)/robots.txt
	mv $(STATIC_DIR)/*.js.map $(SOURCE_MAP_DIR)

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
	$(GO) build -ldflags "-X main.buildRevision=$(RELEASE_REVISION) -X main.release=$(RELEASE_VERSION)" -o $(BINARY) $(MONETR_CLI_PACKAGE)

BUILD_DIR=$(PWD)/build
$(BUILD_DIR):
	mkdir -p $(PWD)/build

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

TEST_FLAGS=-race -v
test-go: $(GO) $(GOMODULES) $(ALL_GO_FILES) $(GOTESTSUM)
	$(call infoMsg,Running go tests for monetr REST API)
	$(GO) run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST)
	$(GOTESTSUM) --junitfile $(PWD)/rest-api-junit.xml --jsonfile $(PWD)/rest-api-tests.json --format testname -- \
		$(TEST_FLAGS) \
		-coverprofile=$(COVERAGE_TXT) \
		-covermode=atomic $(GO_SRC_DIR)/...
	$(GO) tool cover -func=$(COVERAGE_TXT)

test-ui: $(ALL_UI_FILES) $(NODE_MODULES)
	$(call infoMsg,Running go tests for monetrs UI)
	yarn test --coverage

test: test-go test-ui

clean:
	-rm -rf $(LOCAL_BIN)
	-rm -rf $(COVERAGE_TXT)
	-rm -rf $(NODE_MODULES)
	-rm -rf $(LOCAL_TMP)
	-rm -rf $(SOURCE_MAP_DIR)
	-rm -rf $(PWD)/generated
	-rm -rf $(PWD)/docs
	-rm -rf $(PWD)/build
	-rm -rf $(PWD)/Notes.md
	-git clean -f -X $(STATIC_DIR)

SWAGGER_YAML=$(PWD)/docs/swagger.yaml
$(SWAGGER_YAML): $(SWAG) $(APP_GO_FILES)
	$(SWAG) init -d $(GO_SRC_DIR)/controller -g controller.go \
		--parseDependency \
		--parseDepth 5 \
		--parseInternal \
		--output $(PWD)/docs
	sed 's/x-deprecated:/deprecated:/g' $(SWAGGER_YAML) > $(SWAGGER_YAML).new
	rm $(SWAGGER_YAML)
	mv $(SWAGGER_YAML).new $(SWAGGER_YAML)
	cp $(PWD)/public/favicon.ico $(PWD)/docs/favicon.ico
	cp $(PWD)/public/logo192.png $(PWD)/docs/logo192.png
	cp $(PWD)/public/logo512.png $(PWD)/docs/logo512.png
	cp $(PWD)/public/manifest.json $(PWD)/docs/manifest.json

docs: $(SWAGGER_YAML)

# redoc-cli is either installed globally and accessible via yarn, or is installed in the node_modules bin dir. This
# variable will check if the file exists via yarn, and if it does not it will default to the node_modules dir. If the
# resulting file path does not exist, then the $(NODE_MODULES) target will be run, which will install the redoc-cli.
REDOC_CLI=$(word 1,$(wildcard $(shell yarn bin)/redoc-cli) $(NODE_MODULES)/.bin/redoc-cli)
$(REDOC_CLI): $(NODE_MODULES)

docs-local: $(SWAGGER_YAML) $(REDOC_CLI)
	$(REDOC_CLI) serve $(SWAGGER_YAML)

DOCKERFILE=$(PWD)/Dockerfile
DOCKER_DEPS=$(DOCKERFILE) $(PWD)/.dockerignore
CONTAINER=$(BUILD_DIR)/monetr.container.tar
$(CONTAINER): $(BUILD_DIR) $(DOCKER_DEPS) $(APP_GO_FILES) $(STATIC_DIR)
	docker buildx $(DOCKER_OPTIONS) build $(DOCKER_CACHE) \
		--build-arg REVISION=$(RELEASE_REVISION) \
		--output type=tar,dest=$(CONTAINER) \
		-t monetr -f $(PWD)/Dockerfile .

docker: $(CONTAINER)

docker-work-web-ui:
	docker build -t workwebui -f Dockerfile.work .

ifdef GITHUB_TOKEN
license: $(LICENSE) $(BINARY)
	$(call infoMsg,Checking dependencies for open source licenses)
	-$(LICENSE) $(PWD)/licenses.hcl $(BINARY)
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
$(GENERATED_YAML): $(HELM) $(SPLIT_YAML)
	$(call infoMsg,Generating Kubernetes yaml using Helm output to:  $(GENERATED_YAML))
	$(call infoMsg,Environment:                                      $(ENVIRONMENT))
	$(call infoMsg,Using values file:                                $(VALUES_FILE))
	$(call infoMsg,Deploying version:                                $(RELEASE_VERSION))
	-rm -rf $(GENERATED_YAML)
	-mkdir -p $(GENERATED_YAML)
	$(HELM) template monetr $(PWD) \
		--dry-run \
		--set image.tag="$(CONTAINER_VERSION)" \
		--set podAnnotations."monetr\.dev/date"="$(BUILD_TIME)" \
		--values=values.$(ENV_LOWER).yaml | $(SPLIT_YAML) --outdir $(GENERATED_YAML) -

generate: $(GENERATED_YAML)

ifdef GITHUB_ACTION
include $(PWD)/Makefile.github-actions
endif

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

beta-code: $(GO)
	@$(GO) run $(MONETR_CLI_PACKAGE) beta new-code -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST) -P $(POSTGRES_PORT) -W $(POSTGRES_PASSWORD)

all: build test generate lint
