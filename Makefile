GIT_REPOSITORY=https://github.com/monetr/monetr.git

# These variables are set first as they are not folder or environment specific.
USERNAME=$(shell whoami)
HOME=$(shell echo ~$(USERNAME))
NOOP=
SPACE = $(NOOP) $(NOOP)
COMMA=,

# This stuff is used for versioning monetr when doing a release or developing locally.
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_HOST=$(shell hostname)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
RELEASE_REVISION=$(shell git rev-parse HEAD)
RELEASE_VERSION ?= $(shell git describe --tags --dirty)

# Containers should not have the `v` prefix. So we take the release version variable and trim the `v` at the beginning
# if it is there.
CONTAINER_VERSION ?= $(RELEASE_VERSION:v%=%)

# We want ALL of our paths to be relative to the repository path on the computer we are on. Never relative to anything
# else.
PWD=$(shell git rev-parse --show-toplevel)


# Then include the colors file to make a lot of the printing prettier.
include $(PWD)/scripts/Colors.mk

# These are some working directories we need for local development.
LOCAL_TMP = $(PWD)/tmp
LOCAL_BIN = $(PWD)/bin
BUILD_DIR = $(PWD)/build

MONETR_CLI_PACKAGE = github.com/monetr/monetr/pkg/cmd
COVERAGE_TXT = $(PWD)/coverage.txt

MONETR_DIR=$(HOME)/.monetr
$(MONETR_DIR):
	if [ ! -d "$(MONETR_DIR)" ]; then mkdir -p $(MONETR_DIR); fi

KUBERNETES_VERSION=1.20.1

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

ifndef POSTGRES_DB
POSTGRES_DB=postgres
endif

ifndef POSTGRES_USER
POSTGRES_USER=postgres
endif

ifndef POSTGRES_HOST
POSTGRES_HOST=localhost
endif

# Just a shorthand to print some colored text, makes it easier to read and tell the developer what all the makefile is
# doing since its doing a ton.
ifndef BUILDKITE
define infoMsg
	@echo "$(GREEN)[$@]$(WHITE) $(1)$(NC)"
endef

define warningMsg
	@echo "$(YELLOW)[$@]$(WHITE) $(1)$(NC)"
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
TEST_GO_FILES=$(shell find $(GO_SRC_DIR) -type f -name '*_test.go')
ALL_SQL_FILES=$(shell find $(GO_SRC_DIR)/migrations/schema -type f -name '*.sql')
# Include the SQL files in this variable, this way when new migrations are added then the app will trigger a rebuild.
APP_GO_FILES=$(filter-out $(TEST_GO_FILES),$(ALL_GO_FILES)) $(ALL_SQL_FILES)

PUBLIC_DIR=$(PWD)/public
UI_SRC_DIR=$(PWD)/ui
ALL_UI_FILES=$(shell find $(UI_SRC_DIR) -type f)
TEST_UI_FILES=$(shell find $(UI_SRC_DIR) -type f -name '*.spec.*')
APP_UI_FILES=$(filter-out $(TEST_UI_FILES),$(ALL_UI_FILES))
PUBLIC_FILES=$(wildcard $(PUBLIC_DIR)/*)
# Of the public files, these are the files that should be copied to the static_dir before the go build.
COPIED_PUBLIC_FILES=$(filter-out $(PUBLIC_DIR)/index.html,$(PUBLIC_FILES))
UI_CONFIG_FILES=$(PWD)/tsconfig.json $(wildcard $(PWD)/*.config.js)

GO_DEPS=$(PWD)/go.mod $(PWD)/go.sum
UI_DEPS=$(PWD)/package.json $(PWD)/yarn.lock

include $(PWD)/scripts/Dependencies.mk
include $(PWD)/scripts/Deployment.mk
include $(PWD)/scripts/Lint.mk
include $(PWD)/scripts/Container.mk

default: build

SOURCE_MAP_DIR=$(BUILD_DIR)/source_maps
$(SOURCE_MAP_DIR):
	mkdir -p $(SOURCE_MAP_DIR)

YARN=$(shell which yarn)

NODE_MODULES=$(PWD)/node_modules
$(NODE_MODULES): $(UI_DEPS)
	$(YARN) install
	touch -a -m $(NODE_MODULES) # Dumb hack to make sure the node modules directory timestamp gets bumpbed for make.

WEBPACK=$(word 1,$(wildcard $(YARN_BIN)/webpack) $(NODE_MODULES)/.bin/webpack)
$(WEBPACK): $(NODE_MODULES)

STATIC_DIR=$(GO_SRC_DIR)/ui/static
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
	$(call infoMsg,Installing dependencies for monetr)
	$(GO) get -t $(GO_SRC_DIR)/...
	touch -a -m $(GOMODULES)

go-dependencies: $(GOMODULES)

ui-dependencies: $(NODE_MODULES)

dependencies: $(GOMODULES) $(NODE_MODULES)

deps: dependencies

build-ui: $(STATIC_DIR)

GOOS ?= $(OS)
GOARCH ?= amd64

ifeq ($(GOOS),windows)
BINARY_FILE_NAME=monetr.exe
else
BINARY_FILE_NAME=monetr
endif

BUILD_DIR=$(PWD)/build
$(BUILD_DIR):
	mkdir -p $(PWD)/build

BINARY=$(BUILD_DIR)/$(BINARY_FILE_NAME)
$(BINARY): $(GO) $(APP_GO_FILES)
ifndef CI
$(BINARY): $(BUILD_DIR) $(STATIC_DIR) $(GOMODULES)
endif
	$(GO) build -ldflags "-s -w -X main.buildHost=$(BUILD_HOST) -X main.buildTime=$(BUILD_TIME) -X main.buildRevision=$(RELEASE_REVISION) -X main.release=$(RELEASE_VERSION)" -o $(BINARY) $(MONETR_CLI_PACKAGE)
	$(call infoMsg,Built monetr binary for: $(GOOS)/$(GOARCH))
	$(call infoMsg,          Build Version: $(RELEASE_VERSION))

build: $(BINARY)

BINARY_TAR=$(BUILD_DIR)/monetr-$(RELEASE_VERSION)-$(GOOS)-$(GOARCH).tar.gz
$(BINARY_TAR): $(BINARY)
$(BINARY_TAR): TAR=$(shell which tar)
$(BINARY_TAR):
	cd $(BUILD_DIR) && $(TAR) -czf $(BINARY_TAR) $(BINARY_FILE_NAME)

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
	$(GOTESTSUM) --junitfile $(PWD)/rest-api-junit.xml \
		--jsonfile $(PWD)/rest-api-tests.json \
		--format testname -- $(TEST_FLAGS) \
		-coverprofile=$(COVERAGE_TXT) \
		-covermode=atomic $(GO_SRC_DIR)/...
	$(GO) tool cover -func=$(COVERAGE_TXT)

test-ui: $(ALL_UI_FILES) $(NODE_MODULES)
	$(call infoMsg,Running go tests for monetrs UI)
	$(YARN) test --coverage

test: test-go test-ui

clean: shutdown
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

DOCKER=$(shell which docker)
DEVELOPMENT_ENV_FILE=$(MONETR_DIR)/development.env
COMPOSE_FILE=$(PWD)/docker-compose.yaml
ifneq ("$(wildcard $(DEVELOPMENT_ENV_FILE))","")
    COMPOSE=$(DOCKER) compose --env-file=$(DEVELOPMENT_ENV_FILE) -f $(COMPOSE_FILE)
else
	COMPOSE=$(DOCKER) compose -f $(COMPOSE_FILE)
endif
develop: $(NODE_MODULES)
	$(COMPOSE) up --wait
	$(MAKE) development-info

development-info:
	$(call infoMsg,=====================================================================================================)
	$(call infoMsg,Local environment is setup.)
	$(call infoMsg,You should be able to access monetr at:       http://localhost)
	$(call infoMsg,)
	$(call infoMsg,Other services are run alongside monetr locally; you can access them at the following URLs:)
	$(call infoMsg,    Email:                                    http://localhost/mail)
	$(call infoMsg,)
	$(call infoMsg,If you want you can see the logs for all the containers using:)
	$(call infoMsg,  $$ make logs)
	$(call infoMsg,)
	$(call infoMsg,If you are working on features related to webhooks you can setup webhook development using:)
	$(call infoMsg,  $$ make webhooks)
	$(call infoMsg,This will setup an ngrok container forwarding to your API instance you dont need to have an API key.)
	$(call infoMsg,However if you dont have one then the webhooks endpoint will only work for a few hours.)
	$(call infoMsg,)
	$(call infoMsg,If you run into problems or need a clean development environment; run the following command:)
	$(call infoMsg,  $$ make shutdown)
	$(call infoMsg,This command will take down the local dev environment but wont remove any node_modules or clean anything.)
	$(call infoMsg,)
	$(call infoMsg,You can see all of these details at any time by running the following command:)
	$(call infoMsg,  $$ make development-info)
	$(call infoMsg,)
	$(call infoMsg,=====================================================================================================)

logs: # Tail logs for the current development environment. Provide NAME to limit to a single process.
ifdef NAME
	$(COMPOSE) logs -f $(NAME)
else
	$(COMPOSE) logs -f
endif

webhooks:
	$(COMPOSE) up ngrok -d
	$(COMPOSE) restart monetr

restart:
	$(COMPOSE) restart

shutdown:
	-$(COMPOSE) exec monetr monetr development clean:plaid
	-$(COMPOSE) down --remove-orphans -v

restart-monetr:
	$(COMPOSE) restart monetr

DOCS_DIR=$(BUILD_DIR)/docs
SWAGGER_YAML=$(DOCS_DIR)/swagger.yaml
$(SWAGGER_YAML): $(SWAG) $(APP_GO_FILES) $(BUILD_DIR)
	$(call infoMsg,Generating Swagger yaml from API comments)
	$(SWAG) init -d $(GO_SRC_DIR)/controller -g controller.go \
		--parseDependency \
		--parseDepth 5 \
		--parseInternal \
		--output $(DOCS_DIR)
	sed 's/x-deprecated:/deprecated:/g' $(SWAGGER_YAML) > $(SWAGGER_YAML).new
	rm $(SWAGGER_YAML)
	mv $(SWAGGER_YAML).new $(SWAGGER_YAML)
	cp $(PWD)/public/favicon.ico $(DOCS_DIR)/favicon.ico
	cp $(PWD)/public/logo192.png $(DOCS_DIR)/logo192.png
	cp $(PWD)/public/logo512.png $(DOCS_DIR)/logo512.png
	cp $(PWD)/public/manifest.json $(DOCS_DIR)/manifest.json

docs: $(SWAGGER_YAML)

# redoc-cli is either installed globally and accessible via yarn, or is installed in the node_modules bin dir. This
# variable will check if the file exists via yarn, and if it does not it will default to the node_modules dir. If the
# resulting file path does not exist, then the $(NODE_MODULES) target will be run, which will install the redoc-cli.
REDOC_CLI=$(word 1,$(wildcard $(shell yarn bin)/redoc-cli) $(NODE_MODULES)/.bin/redoc-cli)
$(REDOC_CLI): $(NODE_MODULES)

docs-local: $(SWAGGER_YAML) $(REDOC_CLI)
	$(REDOC_CLI) serve $(SWAGGER_YAML)

docs-static: $(SWAGGER_YAML) $(REDOC_CLI)
	$(call infoMsg,Building static API documentation site)
	$(REDOC_CLI) bundle $(SWAGGER_YAML) -o $(DOCS_DIR)/index.html

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
TEMPLATE_FILES=$(wildcard $(PWD)/templates/*)

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

ifndef CI
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
