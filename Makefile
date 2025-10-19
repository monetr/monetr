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

ifneq ("$(wildcard $(PWD)/.cmakepreset)","")
	CMAKE_PRESET ?= $(shell cat $(PWD)/.cmakepreset)
endif

ifneq ("$(wildcard $(HOME)/.monetr/development.env)","")
	include $(HOME)/.monetr/development.env
	export
endif

ifndef CI
	CMAKE_PRESET ?= default
	CONCURRENCY ?= 8
else
	CMAKE_PRESET ?= release
	CONCURRENCY ?= 4
endif

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
# cmake -S . -B $(CMAKE_CONFIGURATION_DIRECTORY) -G $(GENERATOR) $(CMAKE_OPTIONS) $(CMAKE_ARGS)

.PHONY: $(CMAKE_CONFIGURATION_DIRECTORY)
$(CMAKE_CONFIGURATION_DIRECTORY): CMakeLists.txt CMakePresets.json
	cmake --preset $(CMAKE_PRESET) $(CMAKE_OPTIONS)

clean:
	-@$(MAKE) shutdown CMAKE_OPTIONS="$(CMAKE_OPTIONS) -DBUILD_TESTING=OFF"
	-cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t clean $(BUILD_ARGS)
	-cmake -E remove_directory $(CMAKE_CONFIGURATION_DIRECTORY) $(BUILD_ARGS)

dependencies: | $(CMAKE_CONFIGURATION_DIRECTORY)
	+cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t dependencies $(BUILD_ARGS)

deps: dependencies

monetr: | $(CMAKE_CONFIGURATION_DIRECTORY)
	+cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.monetr $(BUILD_ARGS)

monetr-release:
	+$(MAKE) monetr -B CMAKE_PRESET=release

release: monetr-release

interface: $(CMAKE_CONFIGURATION_DIRECTORY)
	+cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.interface $(BUILD_ARGS)

.PHONY: docs
docs: | $(CMAKE_CONFIGURATION_DIRECTORY)
	+cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.docs $(BUILD_ARGS)

email: | $(CMAKE_CONFIGURATION_DIRECTORY)
	+cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t build.email $(BUILD_ARGS)

migrate: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.migrate $(BUILD_ARGS)

# If the user provides a pattern, then pass that through to CTest
ifdef PATTERN
PATTERN_ARG=-R $(PATTERN)
endif
test:
	cmake --preset testing
	ctest --test-dir $(CMAKE_CONFIGURATION_DIRECTORY) --no-tests=error --output-on-failure --output-junit $(PWD)$(CMAKE_CONFIGURATION_DIRECTORY)/junit.xml -j $(CONCURRENCY) $(PATTERN_ARG)

lint: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t lint $(BUILD_ARGS)

develop: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.monetr.up $(BUILD_ARGS)

develop-lite: | $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t development.lite $(BUILD_ARGS)

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

images: $(CMAKE_CONFIGURATION_DIRECTORY)
	cmake --build $(CMAKE_CONFIGURATION_DIRECTORY) -t images $(BUILD_ARGS)
