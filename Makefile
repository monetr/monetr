LOCAL_TMP = $(PWD)/tmp
LOCAL_BIN = $(PWD)/bin
NODE_MODULES_DIR = $(PWD)/node_modules
NODE_MODULES_BIN = $(NODE_MODULES_DIR)/.bin
VENDOR_DIR = $(PWD)/vendor
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
RELEASE_REVISION=$(shell git rev-parse HEAD)
MONETR_CLI_PACKAGE = github.com/monetr/rest-api/pkg/cmd
COVERAGE_TXT = $(PWD)/coverage.txt

ifndef ENVIRONMENT
ENVIRONMENT = Staging
endif

ENV_LOWER = $(shell echo $(ENVIRONMENT) | tr A-Z a-z)

PATH+=\b:$(GOPATH)/bin:$(LOCAL_BIN):$(NODE_MODULES_BIN)

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
define infoMsg
	@echo "\033[0;32m[$@] $(1)\033[0m"
endef

define warningMsg
	@echo "\033[1;33m[$@] $(1)\033[0m"
endef

GO_SRC_DIR=$(PWD)/pkg
ALL_GO_FILES=$(wildcard $(GO_SRC_DIR)/**/*.go)
APP_GO_FILES=$(filter-out $(GO_SRC_DIR)/**/*_test.go, $(ALL_GO_FILES))
TEST_GO_FILES=$(wildcard $(GO_SRC_DIR)/**/*_test.go)

GO_DEPS=go.mod go.sum

include $(PWD)/scripts/*.mk

default: build

dependencies: $(GO_DEPS)
	$(call infoMsg,Installing dependencies for monetrs rest-api)
	go get ./...

build: dependencies $(APP_GO_FILES)
	$(call infoMsg,Building rest-api binary)
	go build -o $(LOCAL_BIN)/monetr $(MONETR_CLI_PACKAGE)

test: dependencies $(ALL_GO_FILES)
	$(call infoMsg,Running go tests for monetr rest-api)
ifndef CI
	go run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST)
endif
	go test -race -v -coverprofile=$(COVERAGE_TXT) -covermode=atomic ./...
	go tool cover -func=$(COVERAGE_TXT)

clean:
	-rm -rf $(LOCAL_BIN)
	-rm -rf $(COVERAGE_TXT)
	-rm -rf $(NODE_MODULES_DIR)
	-rm -rf $(VENDOR_DIR)
	-rm -rf $(LOCAL_TMP)
	-rm -rf $(PWD)/generated
	-rm -rf $(PWD)/docs
	-rm -rf $(PWD)/Notes.md

docs: $(SWAG) $(APP_GO_FILES)
	$(SWAG) init -d $(GO_SRC_DIR)/controller -g controller.go \
		--parseDependency \
		--parseDepth 5 \
		--parseInternal \
		--output $(PWD)/docs

docs-local: docs
	redoc-cli serve $(PWD)/docs/swagger.yaml

docker: Dockerfile $(APP_GO_FILES)
	docker build \
		--build-arg REVISION=$(RELEASE_REVISION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t monetr-rest-api -f Dockerfile .

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

generate: OUTPUT_DIR = $(PWD)/generated/$(ENV_LOWER)
generate: IMAGE_TAG=$(shell git rev-parse HEAD)
generate: VALUES_FILE=$(PWD)/values.$(ENV_LOWER).yaml
generate: $(HELM) $(SPLIT_YAML) $(VALUES_FILE) $(wildcard $(PWD)/templates/*)
	$(call infoMsg,Generating Kubernetes yaml using Helm output to:  $(OUTPUT_DIR))
	$(call infoMsg,Environment:                                      $(ENVIRONMENT))
	$(call infoMsg,Using values file:                                $(VALUES_FILE))
	-rm -rfd $(OUTPUT_DIR) # Clean up the output dir beforehand.
	$(HELM) template rest-api $(PWD) \
		--dry-run \
		--set image.tag="$(IMAGE_TAG)" \
		--set podAnnotations."monetr\.dev/date"="$(BUILD_TIME)" \
		--set podAnnotations."monetr\.dev/sha"="$(IMAGE_TAG)" \
		--values=values.$(ENV_LOWER).yaml | $(SPLIT_YAML) --outdir $(OUTPUT_DIR) -

ifdef GITLAB_CI
include Makefile.gitlab-ci
include Makefile.deploy
endif

ifdef GITHUB_ACTION
include Makefile.github-actions
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

include Makefile.release
include Makefile.docker

ifndef CI
include Makefile.tinker
include Makefile.local
endif

ifndef POSTGRES_PORT
POSTGRES_PORT=5432
endif

migrate:
	@go run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST) -P $(POSTGRES_PORT) -W $(POSTGRES_PASSWORD)

beta-code: migrate
	@go run $(MONETR_CLI_PACKAGE) beta new-code -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST) -P $(POSTGRES_PORT) -W $(POSTGRES_PASSWORD)

