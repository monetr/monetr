ROOT_DIR=$(shell git rev-parse --show-toplevel)

ifndef CI
# Now you're thinking with portals
%:
	$(MAKE) -C $(ROOT_DIR) $@
else
%:
	@echo "Skipping 'make $@' due to CI=true";
endif
