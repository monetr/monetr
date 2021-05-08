LOCAL_BIN_DIR = "$(PWD)/bin"
NODE_MODULES_DIR = "$(PWD)/node_modules"
VENDOR_DIR = "$(PWD)/vendor"
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
RELEASE_REVISION=$(shell git rev-parse HEAD)
MONETR_CLI_PACKAGE = "github.com/monetrapp/rest-api/pkg/cmd"
COVERAGE_TXT = "$(PWD)/coverage.txt"

PATH += "$(GOPATH):$(LOCAL_BIN_DIR)"

ifndef POSTGRES_DB
POSTGRES_DB=postgres
endif

ifndef POSTGRES_USER
POSTGRES_USER=postgres
endif

ifndef POSTGRES_HOST
POSTGRES_HOST=localhost
endif

default: dependencies build test

dependencies:
	go get ./...

build:
	go build -o $(LOCAL_BIN_DIR)/monetr $(MONETR_CLI_PACKAGE)

test:
ifndef CI
	go run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST)
endif
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
	docker build \
		--build-arg REVISION=$(RELEASE_REVISION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t harder-rest-api -f Dockerfile .

docker-work-web-ui:
	docker build -t workwebui -f Dockerfile.work .

clean-development:
	docker-compose -f ./docker-compose.development.yaml rm --stop --force || true

compose-development: docker docker-work-web-ui
	docker-compose  -f ./docker-compose.development.yaml up

compose-development-lite:
	docker-compose  -f ./docker-compose.development.yaml up

ifdef GITLAB_CI
include Makefile.gitlab-ci
endif

ifdef GITHUB_ACTION
include Makefile.github-actions
endif

include Makefile.release
include Makefile.tinker
include Makefile.deploy
include Makefile.docker

ifndef CI
include Makefile.local
endif

ifndef POSTGRES_PORT
POSTGRES_PORT=5432
endif

migrate:
	@go run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST) -P $(POSTGRES_PORT) -W $(POSTGRES_PASSWORD)

beta-code: migrate
	@go run $(MONETR_CLI_PACKAGE) beta new-code -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST) -P $(POSTGRES_PORT) -W $(POSTGRES_PASSWORD)
