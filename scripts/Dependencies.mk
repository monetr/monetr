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
ifndef CI
GOROOT=$(LOCAL_BIN)/go
GOVERSION=go1.18.0
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
else
$(GO):
	$(error You must have golang installed to perform this task)
endif
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
	cd $(SPLIT_YAML_DIR) && go build -o $(SPLIT_YAML) ./...
	rm -rfd $(SPLIT_YAML_DIR)

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

install-$(HOSTESS): $(LOCAL_BIN) $(GO)
	$(call infoMsg,Installing hostess to $(HOSTESS))
	GOBIN=$(LOCAL_BIN) $(GO) install github.com/cbednarski/hostess@latest

MKCERT=$(LOCAL_BIN)/mkcert
$(MKCERT):
	@if [ ! -f "$(MKCERT)" ]; then $(MAKE) install-$(MKCERT); fi

install-$(MKCERT): $(LOCAL_BIN) $(GO)
	$(call infoMsg,Installing mkcert to $(MKCERT))
	GOBIN=$(LOCAL_BIN) $(GO) install filippo.io/mkcert@latest

KUBEVAL=$(LOCAL_BIN)/kubeval
$(KUBEVAL):
	@if [ ! -f "$(KUBEVAL)" ]; then $(MAKE) install-$(KUBEVAL); fi

install-$(KUBEVAL): KUBEVAL_VERSION=latest
install-$(KUBEVAL): KUBEVAL_URL = "https://github.com/instrumenta/kubeval/releases/$(KUBEVAL_VERSION)/download/kubeval-$(OS)-$(ARCH).tar.gz"
install-$(KUBEVAL): KUBEVAL_DIR=$(LOCAL_TMP)/kubeval
install-$(KUBEVAL): KUBEVAL_TAR=$(KUBEVAL_DIR).tar.gz
install-$(KUBEVAL): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing kubeval to $(KUBEVAL))
	-rm -rf $(KUBEVAL_DIR)
	mkdir -p $(KUBEVAL_DIR)
	curl -SsL $(KUBEVAL_URL) --output $(KUBEVAL_TAR)
	tar -xzf $(KUBEVAL_TAR) -C $(KUBEVAL_DIR)
	cp $(KUBEVAL_DIR)/kubeval $(KUBEVAL)
	rm -rf $(KUBEVAL_DIR)
	rm -rf $(KUBEVAL_TAR)

KUBELINT=$(LOCAL_BIN)/kube-linter
$(KUBELINT):
	@if [ ! -f "$(KUBELINT)" ]; then $(MAKE) install-$(KUBELINT); fi

install-$(KUBELINT): KUBELINT_REPO = "https://github.com/stackrox/kube-linter.git"
install-$(KUBELINT): KUBELINT_DIR=$(LOCAL_TMP)/kube-lint
install-$(KUBELINT): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing kube-lint to $(KUBELINT))
	-rm -rf $(KUBELINT_DIR)
	git clone $(KUBELINT_REPO) $(KUBELINT_DIR)
	cd $(KUBELINT_DIR) && go build -o $(KUBELINT) golang.stackrox.io/kube-linter/cmd/kube-linter
	rm -rf $(KUBELINT_DIR)

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

GOTESTSUM=$(LOCAL_BIN)/gotestsum
$(GOTESTSUM):
	@if [ ! -f "$(GOTESTSUM)" ]; then $(MAKE) install-$(GOTESTSUM); fi

install-$(GOTESTSUM): GOTESTSUM_VERSION=1.7.0
install-$(GOTESTSUM): GOTESTSUM_URL="https://github.com/gotestyourself/gotestsum/releases/download/v$(GOTESTSUM_VERSION)/gotestsum_$(GOTESTSUM_VERSION)_$(OS)_$(ARCH).tar.gz"
install-$(GOTESTSUM): GOTESTSUM_DIR=$(LOCAL_TMP)/gotestsum
install-$(GOTESTSUM): GOTESTSUM_TAR=$(GOTESTSUM_DIR).tar.gz
install-$(GOTESTSUM): $(LOCAL_BIN) $(LOCAL_TMP) $(CURL)
	$(call infoMsg,Installing gotestsum to $(GOTESTSUM))
	-rm -rf $(GOTESTSUM_DIR)
	mkdir -p $(GOTESTSUM_DIR)
	curl -SsL $(GOTESTSUM_URL) --output $(GOTESTSUM_TAR)
	tar -xzf $(GOTESTSUM_TAR) -C $(GOTESTSUM_DIR)
	cp $(GOTESTSUM_DIR)/gotestsum $(GOTESTSUM)
	rm -rf $(GOTESTSUM_DIR)
	rm -rf $(GOTESTSUM_TAR)

JQ=$(LOCAL_BIN)/jq
$(JQ):
	@if [ ! -f "$(JQ)" ]; then $(MAKE) install-$(JQ); fi

ifeq ($(OS),darwin)
install-$(JQ): JQ_OS=osx
else
install-$(JQ): JQ_OS=linux
endif
install-$(JQ): JQ_VERSION=1.6
install-$(JQ): JQ_URL = "https://github.com/stedolan/jq/releases/download/jq-$(JQ_VERSION)/jq-$(JQ_OS)-$(ARCH)"
install-$(JQ): $(LOCAL_BIN)
	$(call infoMsg,Installing jq to $(JQ))
	curl -L $(JQ_URL) -o $(JQ)
	sudo chmod +x $(JQ)

YQ=$(LOCAL_BIN)/yq
$(YQ):
	@if [ ! -f "$(YQ)" ]; then make install-$(YQ); fi

install-$(YQ): YQ_VERSION=v4.7.1
install-$(YQ): YQ_BINARY=yq_$(OS)_$(ARCH)
install-$(YQ): YQ_URL = "https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/$(YQ_BINARY).tar.gz"
install-$(YQ): YQ_DIR=$(LOCAL_TMP)/yq
install-$(YQ): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing yq to $(YQ))
	-rm -rf $(YQ_DIR)
	mkdir -p $(YQ_DIR)
	curl -L $(YQ_URL) -o $(YQ_DIR).tar.gz
	tar -xzf $(YQ_DIR).tar.gz -C $(YQ_DIR)
	mv $(YQ_DIR)/yq_$(OS)_$(ARCH) $(YQ)
	-rm -rf $(YQ_DIR)
	-rm -rf $(YQ_DIR).tar.gz

TERRAFORM=$(LOCAL_BIN)/terraform
$(TERRAFORM):
	@if [ ! -f "$(TERRAFORM)" ]; then make install-$(TERRAFORM); fi

install-$(TERRAFORM): TERRAFORM_VERSION=1.0.10
install-$(TERRAFORM): TERRAFORM_URL = "https://releases.hashicorp.com/terraform/$(TERRAFORM_VERSION)/terraform_$(TERRAFORM_VERSION)_$(OS)_$(ARCH).zip"
install-$(TERRAFORM): TERRAFORM_ZIP=$(LOCAL_TMP)/terraform.zip
install-$(TERRAFORM): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing terraform to $(TERRAFORM))
	-rm -rf $(TERRAFORM_ZIP)
	curl -L $(TERRAFORM_URL) -o $(TERRAFORM_ZIP)
	unzip $(TERRAFORM_ZIP) -d $(LOCAL_BIN)
	-rm -rf $(TERRAFORM_ZIP)

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
