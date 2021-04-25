
# Make it so we can run commands from our dependencies directly.
PATH += :$(PWD)/node_modules/.bin
BUILD_DIR = $(PWD)/build
PUBLIC_DIR = $(PWD)/public

dependencies:
	yarn install

clean:
	rm -rf $(BUILD_DIR)/* || true

big-clean: clean
	rm -rf $(PWD)/node_modules || true

build: dependencies clean
	yarn build-prod
	cp $(PUBLIC_DIR)/favicon.ico $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/logo*.png $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/manifest.json $(BUILD_DIR)/
	cp $(PUBLIC_DIR)/robots.txt $(BUILD_DIR)/

include Makefile.deploy
