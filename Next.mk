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

export PATH := $(TOOLS_BIN):$(PATH)

$(BUILD_DIR): # If the build directory does not exist, create it.
	mkdir $@

$(TOOLS_BIN): $(BUILD_DIR) # If the tools bin directory does not exist, create it.
	mkdir -p $@

$(TMP): $(BUILD_DIR) # If the temp dir doesnt exist, create it.
	mkdir $@

NODE=$(TOOLS_BIN)/node
NODE_VERSION=18.16.0
ifeq ($(ARCH),amd64)
NODE_ARCH=x64
else
NODE_ARCH=$(ARCH)
endif
NODE_NAME = node-v$(NODE_VERSION)-$(OS)-$(NODE_ARCH)
$(NODE): NODE_BINARY_URL = "https://nodejs.org/dist/v$(NODE_VERSION)/node-v$(NODE_VERSION)-$(OS)-$(NODE_ARCH).tar.gz"
$(NODE): NODE_TAR = $(TMP)/$(NODE_NAME).tar.gz
$(NODE): | $(TOOLS_BIN) $(TMP)
	-rm -rf $(NODE_TAR)
	curl -L $(NODE_BINARY_URL) --output $(NODE_TAR)
	tar -xzf $(NODE_TAR) -C $(TOOLS)
	ln -sf $(TOOLS)/$(NODE_NAME)/bin/node $(NODE)
	touch -a -m $(NODE)
	-rm -rf $(NODE_TAR)

NPX=$(TOOLS_BIN)/npx
$(NPX): $(NODE)
	ln -sf $(TOOLS)/$(NODE_NAME)/bin/npx $(NPX)









node_modules: $(NPX)
	$(NPX) yarn install
