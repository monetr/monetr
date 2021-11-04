ifndef ENVIRONMENT
dry:
	$(error ENVIRONMENT is not specified)

deploy:
	$(error ENVIRONMENT is not specified)
else

ifeq ($(ENV_LOWER),local)
DEPLOY_NAMESPACE=default
else
DEPLOY_NAMESPACE=monetr
endif

dry: $(KUBECTL) $(GENERATED_YAML)
	$(call infoMsg,Dry running deployment of monetr to $(DEPLOY_NAMESPACE))
	$(KUBECTL) apply -f $(GENERATED_YAML) -n $(DEPLOY_NAMESPACE) --dry-run=server

deploy: $(KUBECTL) $(GENERATED_YAML)
	$(call infoMsg,Deploying monetr to $(DEPLOY_NAMESPACE))
	$(KUBECTL) apply -f $(GENERATED_YAML) -n $(DEPLOY_NAMESPACE)
	$(KUBECTL) rollout status deploy/monetr -n $(DEPLOY_NAMESPACE) --timeout=120s
endif