KUBERNETES_SCHEMA_REPOS=https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master

lint: kubeval kubelint vet

kubeval: $(KUBEVAL) $(GENERATED_YAML)
	-$(KUBEVAL) -v "$(KUBERNETES_VERSION)" --strict \
		--additional-schema-locations="$(KUBERNETES_SCHEMA_REPOS)" \
		$(GENERATED_YAML)/*.yaml

kubelint: $(KUBELINT) $(GENERATED_YAML)
	-$(KUBELINT) lint $(GENERATED_YAML)

vet:
	-go vet ./...