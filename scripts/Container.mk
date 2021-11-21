
PODMAN_MACHINE=monetr
PODMAN_CPUS=4
PODMAN_DISK_SIZE=4 # GB of disk for podman
PODMAN_MEMORY=4096 # MB of memory for podman

PODMAN=$(word 1, $(wildcard $(shell which podman)) $(LOCAL_BIN)/podman)

.PHONY: $(PODMAN)
$(PODMAN):
	@if [ ! -f "$(PODMAN)" ]; then $(MAKE) $(PODMAN)-install && exit 1; fi
	$(MAKE) $(PODMAN)-status

$(PODMAN)-install:
	$(error Podman is not installed; you must install podman to continue)

ifdef CI
$(PODMAN)-status:
	@exit 0; # no op
else
$(PODMAN)-status:
	($(PODMAN) machine list | grep $(PODMAN_MACHINE)) > /dev/null 2>&1 || $(MAKE) $(PODMAN)-machine
	@$(PODMAN) info > /dev/null 2>&1 || $(MAKE) $(PODMAN)-machine

$(PODMAN)-machine:
	-$(PODMAN) machine stop $(PODMAN_MACHINE) > /dev/null 2>&1
	-$(PODMAN) machine rm --force $(PODMAN_MACHINE) > /dev/null 2>&1
	$(PODMAN) machine init \
 		--cpus=$(PODMAN_CPUS) \
 		--disk-size=$(PODMAN_DISK_SIZE) \
 		--memory=$(PODMAN_MEMORY) \
 		$(PODMAN_MACHINE)
	$(call infoMsg,Starting monetr podman machine)
	$(PODMAN) machine start $(PODMAN_MACHINE)
endif

DOCKERFILE=$(PWD)/Dockerfile
DOCKER_IGNORE=$(PWD)/.dockerignore
CONTAINER_REPOS=ghcr.io/monetr/monetr docker.io/monetr/monetr
CONTAINER_VERSIONS=latest $(CONTAINER_VERSION)
CONTAINER_TAGS=$(foreach CONTAINER_REPO,$(CONTAINER_REPOS),$(foreach C_VERSION,$(CONTAINER_VERSIONS),$(CONTAINER_REPO):$(C_VERSION)))
CONTAINER_TAG_ARGS=$(foreach TAG,$(CONTAINER_TAGS),-t $(TAG))
CONTAINER_VARS = GOFLAGS="" REVISION="$(RELEASE_REVISION)" RELEASE="$(RELEASE_VERSION)"
CONTAINER_VAR_ARGS=$(foreach ARG,$(CONTAINER_VARS),--build-arg $(ARG))
ifdef CI
CONTAINER_PLATFORMS=linux/amd64 linux/arm64
else # Eventually we can add arm64 back for local builds
CONTAINER_PLATFORMS=linux/amd64
endif
CONTAINER_PLATFORM_ARGS=$(foreach PLATFORM,$(CONTAINER_PLATFORMS),--platform $(PLATFORM))

CONTAINER_MANIFEST=$(word 1,$(CONTAINER_REPOS)):$(RELEASE_REVISION)
ifeq ($(word 1,$(CONTAINER_PLATFORMS)),$(CONTAINER_PLATFORMS)) # If platforms[0] == platforms then there is only one platform.
CONTAINER_EXTRA_ARGS=$(CONTAINER_TAG_ARGS)
else # When we are working with more than one platform, then we need to use manifest instead of tags.
CONTAINER_EXTRA_ARGS=--manifest $(CONTAINER_MANIFEST)
endif

container: $(BUILD_DIR) $(DOCKERFILE) $(DOCKER_IGNORE) $(APP_GO_FILES)
ifdef CI # When we are in CI we don't want to run the static dir targets, these files are provided via artifacts.
container: BUILDAH=$(shell which buildah)
container:
	$(call infoMsg,Building monetr container for; $(subst $(SPACE),$(COMMA)$(SPACE),$(CONTAINER_PLATFORMS)))
	$(foreach PLATFORM,$(CONTAINER_PLATFORMS),$(BUILDAH) bud \
		$(CONTAINER_VAR_ARGS) \
		--ignorefile=$(DOCKER_IGNORE) \
		--platform="$(PLATFORM)" \
		$(CONTAINER_EXTRA_ARGS) \
		-f $(DOCKERFILE) \
		$(PWD) &&) exit 0;
	$(BUILDAH) manifest inspect $(CONTAINER_MANIFEST)
else
container: $(PODMAN) $(STATIC_DIR)
	$(call infoMsg,Building monetr container for; $(subst $(SPACE),$(COMMA)$(SPACE),$(CONTAINER_PLATFORMS)))
	$(call infoMsg,Tagging container with versions; $(subst $(SPACE),$(COMMA)$(SPACE),$(CONTAINER_VERSIONS)))
	$(PODMAN) build \
		$(CONTAINER_VAR_ARGS) \
		--ignorefile=$(DOCKER_IGNORE) \
		$(CONTAINER_PLATFORM_ARGS) \
		$(CONTAINER_EXTRA_ARGS) \
		-f $(DOCKERFILE) \
		$(PWD)
endif

ifdef CI
container-push: BUILDAH=$(shell which buildah)
container-push:
	$(call infoMsg,Tagging container with versions; $(subst $(SPACE),$(COMMA)$(SPACE),$(CONTAINER_VERSIONS)))
	($(foreach TAG,$(CONTAINER_TAGS),$(BUILDAH) manifest push --all $(CONTAINER_MANIFEST) docker://$(TAG) &&) exit 0)
endif

docker: container

