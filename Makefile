default: build

.PHONY: build migrate test

build:
	@bazel build monetr --output_filter="^((?!dependency checking of directories is unsound).)\*$$" --verbose_failures

migrate:
	@bazel run monetr -- database migrate

test: migrate
	@bazel test //... --cache_test_results=no --test_output=errors --output_filter="^((?!dependency checking of directories is unsound).)\*$$"

bump:
	@bazel run //:gazelle-update-repos

clean:
	-rm -rf $(PWD)/.aspect
	-rm -rf $(PWD)/node_modules
	-rm -rf $(PWD)/generated
	-rm -rf $(PWD)/coverage
	-rm -rf $(PWD)/pkg/icons/sources/simple-icons
	-git clean -f -X $(PWD)/pkg/ui/static
	-bazel clean --expunge
	-bazel shutdown
