SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory

USERNAME=$(shell whoami)
HOME=$(shell echo ~$(USERNAME))
NOOP=
SPACE = $(NOOP) $(NOOP)
COMMA=,
EDITOR ?= vim

# This pretty much anchors us in the project root. We want all commands to run from the root directory.
PWD=$(shell git rev-parse --show-toplevel)

###############################################################################
#
#  HOST INFORMATION
#
###############################################################################

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

###############################################################################
#
#  DEPENDENCIES
#
###############################################################################


BUILD_DIR=$(PWD)/build
TMP=$(BUILD_DIR)/tmp
TOOLS=$(BUILD_DIR)/tools
TOOLS_BIN=$(TOOLS)/bin


$(BUILD_DIR): # If the build directory does not exist, create it.
	mkdir $@

$(TOOLS): $(BUILD_DIR)
	mkdir $@

$(TOOLS_BIN): $(TOOLS) # If the tools bin directory does not exist, create it.
	mkdir $@

$(TMP): $(BUILD_DIR) # If the temp dir doesnt exist, create it.
	mkdir $@

NODE_VERSION=18.16.0
ifeq ($(ARCH),amd64)
NODE_ARCH=x64
else
NODE_ARCH=$(ARCH)
endif
NODE_NAME = node-v$(NODE_VERSION)-$(OS)-$(NODE_ARCH)
NODE_TOOLCHAIN=$(TOOLS)/$(NODE_NAME)
$(NODE_TOOLCHAIN): NODE_BINARY_URL = "https://nodejs.org/dist/v$(NODE_VERSION)/node-v$(NODE_VERSION)-$(OS)-$(NODE_ARCH).tar.gz"
$(NODE_TOOLCHAIN): NODE_TAR = $(TMP)/$(NODE_NAME).tar.gz
$(NODE_TOOLCHAIN): | $(TOOLS) $(TMP)
	-rm -rf $(NODE_TAR)
	curl -L $(NODE_BINARY_URL) --output $(NODE_TAR)
	tar -xzf $(NODE_TAR) -C $(TOOLS)
	-rm -rf $(NODE_TAR)

export PATH := $(TOOLS_BIN):$(NODE_TOOLCHAIN)/bin:$(PATH)

NODE=$(NODE_TOOLCHAIN)/bin/node
$(NODE): $(NODE_TOOLCHAIN)

NPX=$(NODE_TOOLCHAIN)/bin/npx
$(NPX): $(NODE_TOOLCHAIN)

NPM=$(NODE_TOOLCHAIN)/bin/npm
$(NPM): $(NODE_TOOLCHAIN)

YARN=$(TOOLS_BIN)/yarn
$(YARN): $(NPM) # Install yarn in the tools directory.
	$(NPM) install --global --prefix $(TOOLS) yarn

export PATH := $(TOOLS_BIN):$(PATH)

NODE_MODULES=$(PWD)/node_modules

$(NODE_MODULES): $(YARN)
ifneq ($(OS),linux)
	$(YARN) install --ignore-platform
else
	$(YARN) install
endif

deps: $(NODE_MODULES)
