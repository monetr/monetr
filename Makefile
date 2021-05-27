
# Make it so we can run commands from our dependencies directly.
PATH:=$(PATH):$(PWD)/node_modules/.bin
BUILD_DIR = $(PWD)/build
PUBLIC_DIR = $(PWD)/public

RELEASE_REVISION=$(shell git rev-parse HEAD)
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
ifndef ENVIRONMENT
	ENVIRONMENT = Local
endif
ENV_LOWER = $(shell echo $(ENVIRONMENT) | tr A-Z a-z)

dependencies:
	yarn install

clean:
	rm -rf $(BUILD_DIR)/* || true

big-clean: clean
	rm -rf $(PWD)/node_modules || true

start: dependencies
	RELEASE_REVISION=$(RELEASE_REVISION) MONETR_ENV=local yarn start

build:
	RELEASE_REVISION=$(RELEASE_REVISION) MONETR_ENV=$(ENV_LOWER) yarn build:production
	cp $(PUBLIC_DIR)/favicon.ico $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/logo*.png $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/manifest.json $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/robots.txt $(BUILD_DIR)/

include Makefile.deploy
include Makefile.docker
include Makefile.release
include Makefile.local

# This is something to help debug CI issues locally. It will run a container and mount the current directory
# locally. Its the same container used in the pipelines so it should be pretty close to the same for debugging.
debug-ci:
	docker run \
		-w /build \
		-v $(PWD):/build \
		-it containers.monetr.dev/node:15.14.0-buster \
		/bin/bash
