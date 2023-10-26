.SUFFIXES:
MAKEFLAGS += --no-print-directory
MAKEFLAGS += --no-builtin-rules

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PWD := $(dir $(MAKEFILE_PATH))

RELEASE ?= $(shell git describe --tag --dirty)
REVISION ?= $(shell git rev-parse HEAD)
CONTAINER_RELEASE ?= $(RELEASE:v%=%)

ifeq ($(OS),Windows_NT)
    # This block will be executed if OS is Windows
    TIME := $(shell powershell -command "[DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ssZ')")
else
    # This block will be executed if OS is not Windows
    TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
endif

ifndef CI
	CMAKE_OPTIONS ?= -DBUILD_THIRD_PARTY_NOTICE=OFF -DCMAKE_BUILD_TYPE=Debug -DBUILD_TESTING=OFF
	CONCURRENCY ?= 8
else
	CMAKE_OPTIONS ?= -DBUILD_THIRD_PARTY_NOTICE=ON -DCMAKE_BUILD_TYPE=Release
	CONCURRENCY := 4
endif

GENERATOR ?= "Unix Makefiles"

ifdef DEBUG
	CMAKE_ARGS += --debug-output
	BUILD_ARGS += -v
	CONCURRENCY = 1
endif

export RELEASE_VERSION=$(RELEASE)
export CONTAINER_VERSION=$(CONTAINER_RELEASE)
export RELEASE_REVISION=$(REVISION)
export BUILD_TIME=$(TIME)

default: monetr

CMAKE_CONFIGURATION_DIRECTORY=build

$(CMAKE_CONFIGURATION_DIRECTORY): CMakeLists.txt
	cmake -S . -B $(CMAKE_CONFIGURATION_DIRECTORY) -G $(GENERATOR) $(CMAKE_OPTIONS) $(CMAKE_ARGS)

clean:
	-@$(MAKE) shutdown CMAKE_OPTIONS="$(CMAKE_OPTIONS) -DBUILD_TESTING=OFF"
	-cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t clean $(BUILD_ARGS)
	-cmake -E remove_directory $(CMAKE_CONFIGURATION_DIRECTORY) $(BUILD_ARGS)
	-git clean -f -X server/ui/static
	-git submodule deinit -f server/icons/sources/simple-icons

dependencies: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t fetch.dependencies -j $(CONCURRENCY) $(BUILD_ARGS)

deps: dependencies

monetr: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.monetr -j $(CONCURRENCY) $(BUILD_ARGS)

monetr-release:
	$(MAKE) monetr -B CMAKE_OPTIONS=-DCMAKE_BUILD_TYPE=Release

.PHONY: docs
docs: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.docs $(BUILD_ARGS)

email: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.email $(BUILD_ARGS)

migrate: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.migrate $(BUILD_ARGS)

test:
	cmake -S . -B $(CMAKE_CONFIGURATION_DIRECTORY) -G $(GENERATOR) -DBUILD_TESTING=ON $(CMAKE_ARGS)
	ctest --test-dir $(CMAKE_CONFIGURATION_DIRECTORY) --no-tests=error --output-on-failure -j $(CONCURRENCY)

develop: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.monetr.up $(BUILD_ARGS)

develop-docs: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.documentation.up $(BUILD_ARGS)

develop-email: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.email $(BUILD_ARGS)

logs: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.logs $(BUILD_ARGS)

restart: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.restart $(BUILD_ARGS)

shell: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.shell $(BUILD_ARGS)

sql-shell: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.shell.sql $(BUILD_ARGS)

redis-shell: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.shell.redis $(BUILD_ARGS)

shutdown: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.down $(BUILD_ARGS)

container: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.container.docker $(BUILD_ARGS)

container-push: $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.container.docker.push $(BUILD_ARGS)

###################################

generate:
	cmake --preset deploy
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.yaml

dry:
	cmake --preset deploy
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t deploy.dry

deploy:
	cmake --preset deploy
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t deploy.apply
