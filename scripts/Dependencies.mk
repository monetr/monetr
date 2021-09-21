$(LOCAL_BIN):
	@if [ ! -f "$(LOCAL_BIN)" ]; then mkdir -p $(LOCAL_BIN); fi

$(LOCAL_TMP):
	@if [ ! -f "$(LOCAL_TMP)" ]; then mkdir -p $(LOCAL_TMP); fi

LICENSE=$(LOCAL_BIN)/golicense
$(LICENSE):
	@if [ ! -f "$(LICENSE)" ]; then $(MAKE) install-$(LICENSE); fi

install-$(LICENSE): LICENSE_REPO = "https://github.com/mitchellh/golicense.git"
install-$(LICENSE): LICENSE_TMP=$(LOCAL_TMP)/golicense
install-$(LICENSE): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing golicense to $(LICENSE))
	rm -rf $(LICENSE_TMP) || true
	git clone $(LICENSE_REPO) $(LICENSE_TMP)
	cd $(LICENSE_TMP) && go build -o $(LICENSE) .
	rm -rf $(LICENSE_TMP) || true

HELM_VERSION=3.5.4
HELM=$(LOCAL_BIN)/helm
$(HELM):
	@if [ ! -f "$(HELM)" ]; then $(MAKE) install-$(HELM); fi

install-$(HELM): HELM_DIR=$(LOCAL_TMP)/helm
install-$(HELM): HELM_TAR=$(HELM_DIR)/helm.tar.gz
install-$(HELM): HELM_BIN_NAME=$(OS)-$(ARCH)
install-$(HELM): HELM_URL = "https://get.helm.sh/helm-v$(HELM_VERSION)-$(HELM_BIN_NAME).tar.gz"
install-$(HELM): $(LOCAL_BIN) $(LOCAL_TMP)
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
install-$(SPLIT_YAML): $(LOCAL_TMP) $(LOCAL_BIN)
	$(call infoMsg,Installing kubernetes-split-yaml from $(SPLIT_YAML_REPO) at $(SPLIT_YAML))
	-rm -rfd $(SPLIT_YAML_DIR)
	git clone $(SPLIT_YAML_REPO) $(SPLIT_YAML_DIR)
	cd $(SPLIT_YAML_DIR) && go build -o $(SPLIT_YAML) ./...
	rm -rfd $(SPLIT_YAML_DIR)

SWAG=$(LOCAL_BIN)/swag
$(SWAG):
	@if [ ! -f "$(SWAG)" ]; then $(MAKE) install-$(SWAG); fi

install-$(SWAG): SWAG_REPO = "https://github.com/swaggo/swag.git"
install-$(SWAG): SWAG_DIR=$(LOCAL_TMP)/swag
install-$(SWAG): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing swag from $(SWAG_REPO) at $(SWAG_DIR))
	-rm -rfd $(SWAG_DIR)
	git clone $(SWAG_REPO) $(SWAG_DIR)
	cd $(SWAG_DIR) && go build -o $(SWAG) github.com/swaggo/swag/cmd/swag
	chmod +x $(SWAG)
	rm -rfd $(SWAG_DIR)

HOSTESS=$(LOCAL_BIN)/hostess
$(HOSTESS):
	@if [ ! -f "$(HOSTESS)" ]; then $(MAKE) install-$(HOSTESS); fi

install-$(HOSTESS): HOSTESS_REPO = "https://github.com/cbednarski/hostess.git"
install-$(HOSTESS): HOSTESS_DIR=$(LOCAL_TMP)/hostess
install-$(HOSTESS): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing hostess to $(HOSTESS))
	-rm -rf $(HOSTESS_DIR)
	git clone $(HOSTESS_REPO) $(HOSTESS_DIR)
	cd $(HOSTESS_DIR) && go build -o $(HOSTESS) .
	rm -rf $(HOSTESS_DIR)

MKCERT=$(LOCAL_BIN)/mkcert
$(MKCERT):
	@if [ ! -f "$(MKCERT)" ]; then $(MAKE) install-$(MKCERT); fi

install-$(MKCERT): MKCERT_REPO = "https://github.com/FiloSottile/mkcert.git"
install-$(MKCERT): MKCERT_DIR=$(LOCAL_TMP)/mkcert
install-$(MKCERT): $(LOCAL_BIN) $(LOCAL_TMP)
	$(call infoMsg,Installing mkcert to $(MKCERT))
	-rm -rf $(MKCERT_DIR)
	git clone $(MKCERT_REPO) $(MKCERT_DIR)
	cd $(MKCERT_DIR) && go build -o $(MKCERT) .
	rm -rf $(MKCERT_DIR)
