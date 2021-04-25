
# Make it so we can run commands from our dependencies directly.
PATH += :$(PWD)/node_modules/.bin

dependencies:
	yarn install

build: dependencies
	yarn build-prod
