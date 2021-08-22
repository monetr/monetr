LOCAL_TMP = $(PWD)/tmp
LOCAL_BIN = $(PWD)/bin
NODE_MODULES_DIR = $(PWD)/node_modules
NODE_MODULES_BIN = $(NODE_MODULES_DIR)/.bin
VENDOR_DIR = $(PWD)/vendor
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
RELEASE_REVISION=$(shell git rev-parse HEAD)
MONETR_CLI_PACKAGE = github.com/monetr/rest-api/pkg/cmd
COVERAGE_TXT = $(PWD)/coverage.txt
LICENSE=$(LOCAL_BIN)/golicense

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


default: build test

dependencies: go.mod go.sum
	$(call infoMsg,Installing dependencies for monetrs rest-api)
	go get ./...

build: dependencies $(wildcard $(PWD)/pkg/**/*.go)
	$(call infoMsg,Building rest-api binary)
	go build -o $(LOCAL_BIN)/monetr $(MONETR_CLI_PACKAGE)

test:
	$(call infoMsg,Running go tests for monetr rest-api)
ifndef CI
	go run $(MONETR_CLI_PACKAGE) database migrate -d $(POSTGRES_DB) -U $(POSTGRES_USER) -H $(POSTGRES_HOST)
endif
	go test -race -v -coverprofile=$(COVERAGE_TXT) -covermode=atomic ./...
	go tool cover -func=$(COVERAGE_TXT)

clean:
	echo $$PATH
	rm -rf $(LOCAL_BIN) || true
	rm -rf $(COVERAGE_TXT) || true
	rm -rf $(NODE_MODULES_DIR) || true
	rm -rf $(VENDOR_DIR) || true
	rm -rf $(LOCAL_TMP) || true

.PHONY: docs
docs:
	swag init -d pkg/controller -g controller.go --parseDependency --parseDepth 5 --parseInternal

docs-local: docs
	redoc-cli serve $(PWD)/docs/swagger.yaml

docker:
	docker build \
		--build-arg REVISION=$(RELEASE_REVISION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t harder-rest-api -f Dockerfile .

docker-work-web-ui:
	docker build -t workwebui -f Dockerfile.work .

$(LOCAL_BIN):
	mkdir $(LOCAL_BIN)

$(LOCAL_TMP):
	mkdir $(LOCAL_TMP)

$(LICENSE):
	@if [ ! -f "$(LICENSE)" ]; then make install-$(LICENSE); fi

LICENSE_REPO=https://github.com/mitchellh/golicense.git
LICENSE_TMP=$(LOCAL_TMP)/golicense
install-$(LICENSE): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing golicense to $(LICENSE))
	rm -rf $(LICENSE_TMP) || true
	git clone $(LICENSE_REPO) $(LICENSE_TMP)
	cd $(LICENSE_TMP) && go build -o $(LICENSE) .
	rm -rf $(LICENSE_TMP) || true

license: $(LICENSE) build
	$(call infoMsg,Checking dependencies for open source licenses)
	- $(LICENSE) $(PWD)/licenses.hcl $(LOCAL_BIN)/monetr

ifdef GITLAB_CI
include Makefile.gitlab-ci
include Makefile.deploy
endif

ifdef GITHUB_ACTION
include Makefile.github-actions
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

