.SUFFIXES:
MAKEFLAGS += --no-print-directory
MAKEFLAGS += --no-builtin-rules

default: build

GIT_REPOSITORY=https://github.com/monetr/monetr.git

# These variables are set first as they are not folder or environment specific.
USERNAME=$(shell whoami)
HOME=$(shell echo ~$(USERNAME))
NOOP=
SPACE = $(NOOP) $(NOOP)
COMMA=,
EDITOR ?= vim

# This stuff is used for versioning monetr when doing a release or developing locally.
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_HOST=$(shell hostname)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD || echo 'branch-unknown')
RELEASE_REVISION=$(shell git rev-parse HEAD || echo 'rev-unknown')
RELEASE_VERSION ?= $(shell git describe --tags --dirty || echo 'dev-unknown')

# Containers should not have the `v` prefix. So we take the release version variable and trim the `v` at the beginning
# if it is there.
CONTAINER_VERSION ?= $(RELEASE_VERSION:v%=%)

# We want ALL of our paths to be relative to the repository path on the computer we are on. Never relative to anything
# else.
PWD=$(shell git rev-parse --show-toplevel || echo '.')


# Then include the colors file to make a lot of the printing prettier.
include $(PWD)/scripts/Colors.mk

MONETR_DIR=$(HOME)/.monetr
$(MONETR_DIR):
	if [ ! -d "$(MONETR_DIR)" ]; then mkdir -p $(MONETR_DIR); fi

# If the developer has a development.env file, then include that in the makefile variables.
MONETR_ENV=$(MONETR_DIR)/development.env
settings: $(MONETR_DIR)
	if [ ! -f "$(MONETR_ENV)" ]; then cat $(PWD)/compose/development.example.env > $(MONETR_ENV); fi
	$(EDITOR) $(MONETR_ENV)

ifneq (,$(wildcard $(MONETR_ENV)))
include $(MONETR_ENV)
endif

# These are some working directories we need for local development.
LOCAL_TMP = $(PWD)/tmp
LOCAL_BIN = $(PWD)/bin
BUILD_DIR = $(PWD)/build

MONETR_CLI_PACKAGE = github.com/monetr/monetr/pkg/cmd
COVERAGE_TXT = $(PWD)/coverage.txt

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
UI_CONFIG_FILES=$(PWD)/tsconfig.json $(wildcard $(PWD)/*.config.js) $(wildcard $(PWD)/*.config.cjs)

GO_DEPS=$(PWD)/go.mod $(PWD)/go.sum
UI_DEPS=$(PWD)/package.json $(PWD)/yarn.lock

include $(PWD)/scripts/Dependencies.mk
include $(PWD)/scripts/Deployment.mk
include $(PWD)/scripts/Lint.mk


SOURCE_MAP_DIR=$(BUILD_DIR)/source_maps
$(SOURCE_MAP_DIR):
	mkdir -p $(SOURCE_MAP_DIR)

YARN=$(shell which yarn)

NODE_MODULES=$(PWD)/node_modules
$(NODE_MODULES): $(UI_DEPS)
ifneq ($(OS),linux)
	$(YARN) install --ignore-platform
else
	$(YARN) install
endif
	touch -a -m $(NODE_MODULES) # Dumb hack to make sure the node modules directory timestamp gets bumpbed for make.

STATIC_DIR=$(GO_SRC_DIR)/ui/static
$(STATIC_DIR): $(APP_UI_FILES) $(NODE_MODULES) $(PUBLIC_FILES) $(UI_CONFIG_FILES) $(SOURCE_MAP_DIR)
	$(call infoMsg,Building UI files)
	rm -rf $(SOURCE_MAP_DIR)/*.js.map # Removing old map files
	git clean -f -X $(STATIC_DIR)
	RELEASE_VERSION=$(RELEASE_VERSION) RELEASE_REVISION=$(RELEASE_REVISION) $(YARN) build --mode production
	cp $(STATIC_DIR)/assets/scripts/*.js.map $(SOURCE_MAP_DIR)

GOMODULES=$(GOPATH)/pkg/mod
$(GOMODULES): $(GO) $(GO_DEPS)
	$(call infoMsg,Installing dependencies for monetr)
	$(GO) get -t $(GO_SRC_DIR)/...
ifndef CI
	touch -a -m $(GOMODULES)
endif

go-dependencies: $(GOMODULES)

ui-dependencies: $(NODE_MODULES)

dependencies: $(GOMODULES) $(NODE_MODULES)

deps: dependencies

build-ui: $(STATIC_DIR)

SIMPLE_ICONS=$(PWD)/pkg/icons/sources/simple-icons/README.md
$(SIMPLE_ICONS):
	git submodule update --init pkg/icons/sources/simple-icons

LICENSED_IMAGE=$(BUILD_DIR)/licensed-$(RELEASE_VERSION).image
$(LICENSED_IMAGE): $(PWD)/scripts/Licensed.containerfile
$(LICENSED_IMAGE) |: $(BUILD_DIR)
$(LICENSED_IMAGE): IMAGE=licensed:$(RELEASE_VERSION)
$(LICENSED_IMAGE):
	$(DOCKER) build -f $< $(dir $<) -t $(IMAGE)
	echo $(IMAGE) > $@

LICENSED_CONFIG=$(PWD)/.licensed.yaml
LICENSED_CACHE=$(PWD)/.licenses
$(LICENSED_CACHE): $(LICENSED_IMAGE) $(GO_DEPS) $(NODE_MODULES) $(LICENSED_CONFIG) $(SIMPLE_ICONS)
$(LICENSED_CACHE): IMAGE=$(shell cat $(LICENSED_IMAGE))
$(LICENSED_CACHE):
	$(DOCKER) run -v "$(PWD):/workspace" $(IMAGE) "licensed cache --force"
ifndef CI
	touch -a -m $(LICENSED_CACHE) # Dumb hack to make sure the licenses directory timestamp gets bumped for make.
endif

.PHONY: license
license: $(LICENSED_IMAGE) $(LICENSED_CACHE) $(LICENSED_CONFIG)
license: IMAGE=$(shell cat $(LICENSED_IMAGE))
license:
	$(DOCKER) run -v "$(PWD):/workspace" $(IMAGE) "licensed status"

NOTICES=$(LICENSED_CACHE)/monetr-API/NOTICE $(LICENSED_CACHE)/monetr-UI/NOTICE
$(NOTICES): $(LICENSED_IMAGE) $(LICENSED_CACHE) $(LICENSED_CONFIG) $(NODE_MODULES) $(SIMPLE_ICONS)
$(NOTICES): IMAGE=$(shell cat $(LICENSED_IMAGE))
$(NOTICES):
	$(DOCKER) run -v "$(PWD):/workspace" $(IMAGE) "licensed notices"

NOTICE=$(GO_SRC_DIR)/build/NOTICE.md
$(NOTICE): $(NOTICES)
	cat $(NOTICES) > $@

notice: $(NOTICE)

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
TAGS ?= icons,simple_icons
ifdef TAGS
	TAGS_FLAG=-tags=$(TAGS)
else
	TAGS_FLAG=-tags=
endif
$(BINARY): $(GO) $(APP_GO_FILES)
ifndef CI
$(BINARY): $(BUILD_DIR) $(STATIC_DIR) $(GOMODULES)
ifndef LITE
$(BINARY): $(NOTICE)
endif
endif
ifneq (,$(findstring simple_icons,$(TAGS))) # If our icon packs include simple_icons then make sure the dir exists.
$(BINARY): $(SIMPLE_ICONS)
endif
	$(GO) build $(TAGS_FLAG) -ldflags "-s -w -X main.buildHost=$(BUILD_HOST) -X main.buildTime=$(BUILD_TIME) -X main.buildRevision=$(RELEASE_REVISION) -X main.release=$(RELEASE_VERSION)" -o $(BINARY) $(MONETR_CLI_PACKAGE)
	$(call infoMsg,Built monetr binary for: $(GOOS)/$(GOARCH))
	$(call infoMsg,          Build Version: $(RELEASE_VERSION))

build: $(BINARY)

lite:
	$(MAKE) $(BINARY) LITE=true

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

TEST_FLAGS=-race -v -parallel 8
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

ifdef GITPOD_WORKSPACE_ID
LOCAL_DOMAIN:=80-$(GITPOD_WORKSPACE_ID).$(GITPOD_WORKSPACE_CLUSTER_HOST)
LOCAL_PROTOCOL:=https
CLOUD_MAGIC:=magic
else
ifdef CODESPACE_NAME
LOCAL_DOMAIN:=$(CODESPACE_NAME)-80.githubpreview.dev
LOCAL_PROTOCOL:=https
CLOUD_MAGIC:=magic
else
LOCAL_DOMAIN ?= monetr.local
LOCAL_PROTOCOL=https
CLOUD_MAGIC=false
endif
endif

clean: shutdown $(HOSTESS)
	-rm -rf $(NOTICE)
	-rm -rf $(LOCAL_BIN)
	-rm -rf $(COVERAGE_TXT)
	-rm -rf $(LICENSED_CACHE)
	-rm -rf $(NODE_MODULES)
	-rm -rf $(LOCAL_TMP)
	-rm -rf $(SOURCE_MAP_DIR)
	-rm -rf $(PWD)/generated
	-rm -rf $(PWD)/coverage
	-rm -rf $(BUILD_DIR)
	-git clean -f -X $(STATIC_DIR)
	-git submodule deinit -f pkg/icons/sources/simple-icons

LOCAL_CERTIFICATE_DIR=$(MONETR_DIR)/certs/$(LOCAL_DOMAIN)
$(LOCAL_CERTIFICATE_DIR): $(MONETR_DIR)
	mkdir -p $(LOCAL_CERTIFICATE_DIR)

LOCAL_CERTIFICATE_KEY=$(LOCAL_CERTIFICATE_DIR)/key.pem
LOCAL_CERTIFICATE_CERT=$(LOCAL_CERTIFICATE_DIR)/cert.pem
LOCAL_CERTIFICATE=$(LOCAL_CERTIFICATE_KEY) $(LOCAL_CERTIFICATE_CERT)
$(LOCAL_CERTIFICATE) &: $(LOCAL_CERTIFICATE_DIR) $(MKCERT)
	$(call infoMsg,Setting up local development TLS certificate; This is required for OAuth2)
	sudo $(MKCERT) -install
	$(MKCERT) -key-file $(LOCAL_CERTIFICATE_KEY) -cert-file $(LOCAL_CERTIFICATE_CERT) $(LOCAL_DOMAIN)

DOCKER=$(shell which docker || echo 'docker')
DEVELOPMENT_ENV_FILE=$(MONETR_ENV)
COMPOSE_FILE=$(PWD)/docker-compose.yaml
ifneq ("$(wildcard $(DEVELOPMENT_ENV_FILE))","")
	COMPOSE=$(DOCKER) compose --env-file=$(DEVELOPMENT_ENV_FILE) -f $(COMPOSE_FILE)
else
	COMPOSE=$(DOCKER) compose -f $(COMPOSE_FILE)
endif
.EXPORT_ALL_VARIABLES: develop
develop: $(NODE_MODULES) $(SIMPLE_ICONS)
ifndef GITPOD_WORKSPACE_ID
ifndef CODESPACE_NAME
ifneq ($(LOCAL_DOMAIN),localhost)
develop: $(HOSTESS) $(GOOGLE_KMS_AUTH) $(LOCAL_CERTIFICATE)
develop: export LOCAL_CERTIFICATE_DIRECTORY=$(LOCAL_CERTIFICATE_DIR)
develop:
	$(call infoMsg,Setting up $(LOCAL_DOMAIN) domain with your /etc/hosts file)
	$(call infoMsg,If you would prefer to not use this; add)
	$(call infoMsg,	LOCAL_DOMAIN=localhost)
	$(call infoMsg,to your $(DEVELOPMENT_ENV_FILE) file)
	sudo $(HOSTESS) add $(LOCAL_DOMAIN) 127.0.0.1
endif
else
	$(call infoMsg,Detected cloud environment; nginx will not use builtin TLS.)
	$(eval export NGINX_CONFIG_NAME=nginx-cloud.conf)
	$(eval export NGINX_PORT=80)
endif
else
	$(call infoMsg,Detected cloud environment; nginx will not use builtin TLS.)
	$(eval export NGINX_CONFIG_NAME=nginx-cloud.conf)
	$(eval export NGINX_PORT=80)
endif
ifdef MKDOCS_IMAGE
	$(call infoMsg,Using custom MKDocs container image; $(MKDOCS_IMAGE))
endif
	$(COMPOSE) up --wait --remove-orphans
ifdef NGROK_AUTH # If the developer has an NGROK_AUTH token specified, then bring up webhooks right away too.
	$(MAKE) webhooks
endif
ifdef GOOGLE_KMS_AUTH
	$(call infoMsg,Google credentials have been provided by (GOOGLE_KMS_AUTH) they will be copied to your development environment.)
	$(COMPOSE) cp $(GOOGLE_KMS_AUTH) monetr:/etc/monetr/google-service-account.json
	$(COMPOSE) restart monetr
else
	$(call infoMsg,No Google credentials were found. Instead monetr will use a KMS service within its development environment.)
	$(call infoMsg,  If you want to develop using Google Cloud KMS then specify a path to a service account JSON file)
	$(call infoMsg,  using the environment variable GOOGLE_KMS_AUTH. The file specified at this path will be loaded)
	$(call infoMsg,  into the development environment at startup.)
	$(call infoMsg,Note: This is not at all required. There is an AWS compliant KMS built into the development)
	$(call infoMsg,  environment by default.)
endif
	$(MAKE) development-info

development-info:
	$(call infoMsg,=====================================================================================================)
	$(call infoMsg,Local environment is setup.)
	$(call infoMsg,You should be able to access monetr at:       $(LOCAL_PROTOCOL)://$(LOCAL_DOMAIN))
	$(call infoMsg,)
	$(call infoMsg,Other services are run alongside monetr locally; you can access them at the following URLs:)
	$(call infoMsg,    Email:                                    $(LOCAL_PROTOCOL)://$(LOCAL_DOMAIN)/mail)
	$(call infoMsg,    Documentation:                            $(LOCAL_PROTOCOL)://$(LOCAL_DOMAIN)/documentation)
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

storybook: $(NODE_MODULES)
	$(YARN) storybook

up:
ifndef CONTAINER
	$(error Must provide a CONTAINER to up)
else
	$(COMPOSE) up $(CONTAINER) -d
endif

logs: # Tail logs for the current development environment. Provide NAME to limit to a single process.
ifdef CONTAINER
	$(COMPOSE) logs -f $(CONTAINER)
else
	$(COMPOSE) logs -f
endif

webhooks:
	$(COMPOSE) up ngrok -d
	$(COMPOSE) restart monetr

sql-shell:
	$(COMPOSE) exec postgres psql -U postgres

redis-shell:
	$(COMPOSE) exec redis redis-cli

shell:
ifdef CONTAINER
	$(COMPOSE) exec $(CONTAINER) /bin/sh
else
	$(error Must specify a CONTAINER to shell into)
endif

stop:
	$(COMPOSE) stop

start:
	$(COMPOSE) start

restart:
ifndef CONTAINER
	$(COMPOSE) restart
else
	$(COMPOSE) restart $(CONTAINER)
endif

shutdown:
	-$(COMPOSE) exec monetr monetr development clean:plaid
	-$(COMPOSE) down --remove-orphans -v

MKDOCS_IMAGE ?= squidfunk/mkdocs-material:8.2.8
MKDOCS_YAML=$(PWD)/mkdocs.yaml
DOCS_FILES=$(shell find $(PWD)/docs -type f)
DOCS_SITE=$(PWD)/build/site/index.html
$(DOCS_SITE): $(MKDOCS_YAML) $(DOCS_FILES)
	$(DOCKER) run -v $(PWD):/work -e SITE_URL=https://monetr.app/ -w /work --rm --entrypoint sh $(MKDOCS_IMAGE) /work/scripts/docs-build.sh

mkdocs: $(DOCS_SITE)

docs: mkdocs

ifdef GITHUB_TOKEN
license-old: $(LICENSE) $(BINARY)
	$(call infoMsg,Checking dependencies for open source licenses)
	-$(LICENSE) $(PWD)/licenses.hcl $(BINARY)
else
.PHONY: license-old
license-old:
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

ifndef POSTGRES_PORT
POSTGRES_PORT=5432
endif

MIGRATE_FLAGS=$(NOOP)
ifdef POSTGRES_DB
MIGRATE_FLAGS += -d $(POSTGRES_DB)
endif
ifdef POSTGRES_USER
MIGRATE_FLAGS += -U $(POSTGRES_USER)
endif
ifdef POSTGRES_HOST
MIGRATE_FLAGS += -H $(POSTGRES_HOST)
endif
ifdef POSTGRES_PORT
MIGRATE_FLAGS += -P $(POSTGRES_PORT)
endif
ifdef POSTGRES_PASSWORD
MIGRATE_FLAGS += -W $(POSTGRES_PASSWORD)
endif


migrate: $(GO)
	@$(GO) run $(MONETR_CLI_PACKAGE) database migrate $(MIGRATE_FLAGS)

beta-code: $(GO)
	@$(GO) run $(MONETR_CLI_PACKAGE) beta new-code -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST) -P $(POSTGRES_PORT) -W $(POSTGRES_PASSWORD)

all: build test generate lint

include $(PWD)/scripts/Container.mk
