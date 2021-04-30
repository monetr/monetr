
# Make it so we can run commands from our dependencies directly.
PATH += :$(PWD)/node_modules/.bin
BUILD_DIR = $(PWD)/build
PUBLIC_DIR = $(PWD)/public

RELEASE_REVISION=$(shell git rev-parse HEAD)
ifndef ENVIRONMENT
	ENVIRONMENT = Staging
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

build: dependencies clean
	RELEASE_REVISION=$(RELEASE_REVISION) MONETR_ENV=$(ENV_LOWER) yarn build:production
	cp $(PUBLIC_DIR)/favicon.ico $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/logo*.png $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/manifest.json $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/robots.txt $(BUILD_DIR)/

include Makefile.deploy
include Makefile.docker
include Makefile.release
