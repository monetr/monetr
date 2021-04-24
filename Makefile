LOCAL_BIN_DIR = "$(PWD)/bin"
NODE_MODULES_DIR = "$(PWD)/node_modules"
VENDOR_DIR = "$(PWD)/vendor"

MONETR_CLI_PACKAGE = "github.com/monetrapp/rest-api/pkg/cmd"
COVERAGE_TXT = "$(PWD)/coverage.txt"

PATH += "$(GOPATH):$(LOCAL_BIN_DIR)"

default: dependencies build test

dependencies:
	go get ./...

build:
	go build -o $(LOCAL_BIN_DIR)/monetr $(MONETR_CLI_PACKAGE)

test:
	go test -race -v -coverprofile=$(COVERAGE_TXT) -covermode=atomic ./...
	go tool cover -func=$(COVERAGE_TXT)

clean:
	rm -rf $(LOCAL_BIN_DIR) || true
	rm -rf $(COVERAGE_TXT) || true
	rm -rf $(NODE_MODULES_DIR) || true
	rm -rf $(VENDOR_DIR) || true

.PHONY: docs
docs:
	swag init -d pkg/controller -g controller.go --parseDependency --parseDepth 5 --parseInternal

docker:
	docker build -t harder-rest-api -f Dockerfile .

docker-work-web-ui:
	docker build -t workwebui -f Dockerfile.work .

clean-development:
	docker compose -f ./docker-compose.development.yaml rm --stop --force || true

compose-development: docker docker-work-web-ui
	docker compose  -f ./docker-compose.development.yaml up

compose-development-lite:
	docker compose  -f ./docker-compose.development.yaml up

generate_schema:
	$(eval TARGET_FILE := $(shell echo "$(TARGET_DIRECTORY)/0_initial.up.sql"))
	$(info "Generating current schema into file $(TARGET_FILE)")
	go run github.com/monetrapp/rest-api/tools/schemagen > $(TARGET_FILE)
	yarn sql-formatter -l postgresql -u --lines-between-queries 2 $(TARGET_FILE) -o $(TARGET_FILE)

migrations:
	$(eval CURRENT_TMP := $(shell mktemp -d))
	$(eval BASE_TMP := $(shell mktemp -d))
	$(info "Generating schema migrations for the current schema in $(CURRENT_TMP)")
	make generate_schema TARGET_DIRECTORY=$(CURRENT_TMP)
	$(info "Cleaning up temp directories")
	rm -rf $(CURRENT_TMP)

ifdef GITLAB_CI
include Makefile.gitlab-ci
endif

ifdef GITHUB_ACTION
include Makefile.github-actions
endif

include Makefile.tinker
include Makefile.deploy
include Makefile.docker

