$(LOCAL_BIN):
	@if [ ! -f "$(LOCAL_BIN)" ]; then mkdir -p $(LOCAL_BIN); fi

$(LOCAL_TMP):
	@if [ ! -f "$(LOCAL_TMP)" ]; then mkdir -p $(LOCAL_TMP); fi

CURL=$(shell which curl)

# If curl is not installed then we have some basic stuff to install it.
ifeq ("$(strip $(CURL))","")
# If we are on linux.
ifeq ($(OS),linux)
# If the flavor of linux we are on is debian.
ifeq ("$(shell cat /etc/os-release | grep 'ID_LIKE=')","ID_LIKE=debian")
# Install CURL for debian.define
CURL=/usr/bin/curl
$(CURL):
	@if [ ! -f "$(CURL)" ]; then $(MAKE) install-$(CURL); fi

install-$(CURL):
	$(call warningMsg,curl needs to be installed in order to continue)
	apt-get update -y
	apt-get install -y curl
else
# If we are not on debian then we need a different script. I'm not sure what it would need to be.
#$(error I dont know how to install curl on your OS)
endif
else
# If we are not on linux then I have no idea what I'd need to do to get curl.
$(error Cannot install curl on something other than linux)
endif
endif

GO_BINARY=$(shell which go)

# Check to see if the user has golang installed. If they don't then install it locally just for this project.
ifeq ("$(strip $(GO_BINARY))","")
GOROOT=$(LOCAL_BIN)/go
GOVERSION=go1.17.1
GO_BINARY=$(GOROOT)/bin/go

GO=$(GO_BINARY)
$(GO):
	$(call infoMsg,Go needs to be installed)
	@if [ ! -f "$(GO)" ]; then $(MAKE) install-$(GO); fi

install-$(GO): GO_URL = "https://golang.org/dl/$(GOVERSION).$(OS)-$(ARCH).tar.gz"
install-$(GO): GO_TAR=$(LOCAL_TMP)/$(GOVERSION).$(OS)-$(ARCH).tar.gz
install-$(GO): $(LOCAL_BIN) $(LOCAL_TMP) $(CURL)
	$(call infoMsg,Installing $(GOVERSION) to $(GOROOT))
	-rm -rf $(GO_TAR)
	curl -L $(GO_URL) --output $(GO_TAR)
	tar -xzf $(GO_TAR) -C $(LOCAL_BIN)
	rm -rf $(GO_TAR)

GO111MODULE=on
GOROOT=$(LOCAL_BIN)/go
endif

GO=$(GO_BINARY)

LICENSE=$(LOCAL_BIN)/golicense
$(LICENSE):
	@if [ ! -f "$(LICENSE)" ]; then $(MAKE) install-$(LICENSE); fi

install-$(LICENSE): LICENSE_REPO = "https://github.com/mitchellh/golicense.git"
install-$(LICENSE): LICENSE_TMP=$(LOCAL_TMP)/golicense
install-$(LICENSE): $(LOCAL_BIN) $(LOCAL_TMP) $(GO)
	$(call infoMsg,Installing golicense to $(LICENSE))
	rm -rf $(LICENSE_TMP) || true
	git clone $(LICENSE_REPO) $(LICENSE_TMP)
	cd $(LICENSE_TMP) && $(GO) build -o $(LICENSE) .
	rm -rf $(LICENSE_TMP) || true

HELM_VERSION=3.5.4
HELM=$(LOCAL_BIN)/helm
$(HELM):
	@if [ ! -f "$(HELM)" ]; then $(MAKE) install-$(HELM); fi

install-$(HELM): HELM_DIR=$(LOCAL_TMP)/helm
install-$(HELM): HELM_TAR=$(HELM_DIR)/helm.tar.gz
install-$(HELM): HELM_BIN_NAME=$(OS)-$(ARCH)
install-$(HELM): HELM_URL = "https://get.helm.sh/helm-v$(HELM_VERSION)-$(HELM_BIN_NAME).tar.gz"
install-$(HELM): $(LOCAL_BIN) $(LOCAL_TMP) $(CURL)
	$(call infoMsg,Installing helm v$(HELM_VERSION) at $(HELM))
	-rm -rf $(HELM_DIR)
	mkdir -p $(HELM_DIR)
	curl -SsL $(HELM_URL) --output $(HELM_TAR)
	tar -xzf $(HELM_TAR) -C $(HELM_DIR)
	cp $(HELM_DIR)/$(HELM_BIN_NAME)/helm $(HELM)
	-rm -rf $(HELM_DIR)
	-rm -rf $(HELM_TAR)

SPLIT_YAML=$(LOCAL_BIN)/kubernetes-split-yaml
$(SPLIT_YAML):
	@if [ ! -f "$(SPLIT_YAML)" ]; then $(MAKE) install-$(SPLIT_YAML); fi

install-$(SPLIT_YAML): SPLIT_YAML_REPO = "https://github.com/mogensen/kubernetes-split-yaml.git"
install-$(SPLIT_YAML): SPLIT_YAML_DIR=$(LOCAL_TMP)/kubernetes-split-yaml
install-$(SPLIT_YAML): $(LOCAL_TMP) $(LOCAL_BIN) $(GO)
	$(call infoMsg,Installing kubernetes-split-yaml from $(SPLIT_YAML_REPO) at $(SPLIT_YAML))
	-rm -rf $(SPLIT_YAML_DIR)
	git clone $(SPLIT_YAML_REPO) $(SPLIT_YAML_DIR)
	cd $(SPLIT_YAML_DIR) && $(GO) build -o $(SPLIT_YAML) ./...
	rm -rf $(SPLIT_YAML_DIR)

SWAG=$(LOCAL_BIN)/swag
$(SWAG):
	@if [ ! -f "$(SWAG)" ]; then $(MAKE) install-$(SWAG); fi

install-$(SWAG): SWAG_REPO = "https://github.com/swaggo/swag.git"
install-$(SWAG): SWAG_DIR=$(LOCAL_TMP)/swag
install-$(SWAG): $(LOCAL_BIN) $(LOCAL_TMP) $(GO)
	$(call infoMsg,Installing swag from $(SWAG_REPO) at $(SWAG_DIR))
	-rm -rf $(SWAG_DIR)
	git clone $(SWAG_REPO) $(SWAG_DIR)
	cd $(SWAG_DIR) && $(GO) build -o $(SWAG) github.com/swaggo/swag/cmd/swag
	chmod +x $(SWAG)
	rm -rf $(SWAG_DIR)

HOSTESS=$(LOCAL_BIN)/hostess
$(HOSTESS):
	@if [ ! -f "$(HOSTESS)" ]; then $(MAKE) install-$(HOSTESS); fi

install-$(HOSTESS): HOSTESS_REPO = "https://github.com/cbednarski/hostess.git"
install-$(HOSTESS): HOSTESS_DIR=$(LOCAL_TMP)/hostess
install-$(HOSTESS): $(LOCAL_BIN) $(LOCAL_TMP) $(GO)
	$(call infoMsg,Installing hostess to $(HOSTESS))
	-rm -rf $(HOSTESS_DIR)
	git clone $(HOSTESS_REPO) $(HOSTESS_DIR)
	cd $(HOSTESS_DIR) && $(GO) build -o $(HOSTESS) .
	rm -rf $(HOSTESS_DIR)

MKCERT=$(LOCAL_BIN)/mkcert
$(MKCERT):
	@if [ ! -f "$(MKCERT)" ]; then $(MAKE) install-$(MKCERT); fi

install-$(MKCERT): MKCERT_REPO = "https://github.com/FiloSottile/mkcert.git"
install-$(MKCERT): MKCERT_DIR=$(LOCAL_TMP)/mkcert
install-$(MKCERT): $(LOCAL_BIN) $(LOCAL_TMP) $(GO)
	$(call infoMsg,Installing mkcert to $(MKCERT))
	-rm -rf $(MKCERT_DIR)
	git clone $(MKCERT_REPO) $(MKCERT_DIR)
	cd $(MKCERT_DIR) && $(GO) build -o $(MKCERT) .
	rm -rf $(MKCERT_DIR)

GCLOUD=$(LOCAL_BIN)/google-cloud-sdk/bin/gcloud
$(GCLOUD):
	@if [ ! -f "$(GCLOUD)" ]; then $(MAKE) install-$(GCLOUD); fi

install-$(GCLOUD): GCLOUD_VERSION=358.0.0
install-$(GCLOUD): GCLOUD_ARCH=$(shell uname -m)
install-$(GCLOUD): GCLOUD_URL = "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-$(GCLOUD_VERSION)-$(OS)-$(GCLOUD_ARCH).tar.gz"
install-$(GCLOUD): GCLOUD_TAR=$(LOCAL_TMP)/gcloud.tar.gz
install-$(GCLOUD): $(LOCAL_BIN) $(LOCAL_TMP) $(CURL)
	$(call infoMsg,Installing gcloud SDK v$(GCLOUD_VERSION) to $(GCLOUD))
	-rm -rf $(GCLOUD_TAR)
	curl -SsL $(GCLOUD_URL) --output $(GCLOUD_TAR)
	tar -xzf $(GCLOUD_TAR) -C $(LOCAL_BIN)
	rm -rf $(GCLOUD_TAR)

ifneq ($(ENV_LOWER),local)
KUBECTL=$(LOCAL_BIN)/kubectl
$(KUBECTL):
	@if [ ! -f "$(KUBECTL)" ]; then $(MAKE) install-$(KUBECTL); fi

install-$(KUBECTL): $(CURL)
install-$(KUBECTL): KUBECTL_STABLE=$(shell curl -L -s https://dl.k8s.io/release/stable.txt)
install-$(KUBECTL): KUBECTL_URL = "https://dl.k8s.io/release/$(KUBECTL_STABLE)/bin/$(OS)/$(ARCH)/kubectl"
install-$(KUBECTL): $(LOCAL_BIN)
	$(call infoMsg,Installing kubectl to $(KUBECTL))
	curl -L $(KUBECTL_URL) --output $(KUBECTL)
	chmod +x $(KUBECTL)
endif


buildkite: $(KUBECTL)
	$(KUBECTL) get po

gcloud: $(GCLOUD)
	$(GCLOUD) info

PROJECT_ID=acceptance-320515
cloud-build: $(GCLOUD)
cloud-build: NAME=$(shell basename `git rev-parse --show-toplevel`)
cloud-build:; $(GCLOUD) builds submit --project=$(PROJECT_ID) --tag=gcr.io/$(PROJECT_ID)/github.com/monetr/$(NAME):$(RELEASE_REVISION)
