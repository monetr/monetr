PATH := "$(PATH):$(GOPATH):$(PWD)/bin"

default: dependencies build test

dependencies:
	go get ./...

build: dependencies
	go build -o bin/monetr github.com/monetrapp/rest-api/cmd/monetr

test:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

docs-dependencies:
	go get ./...
	(PATH=$$PATH:./bin/swag which swag) || (go get github.com/swaggo/swag/cmd/swag && go build -o ./bin/swag github.com/swaggo/swag/cmd/swag)

docs: docs-dependencies
	PATH=$$PATH:./bin/swag swag init -d pkg/controller -g controller.go --parseDependency --parseDepth 5 --parseInternal

schema:
	go run github.com/monetrapp/rest-api/tools/schemagen > schema/00000000_Initial.up.sql
	(which yarn && yarn sql-formatter -l postgresql -u --lines-between-queries 2 -i 4 \
		schema/00000000_Initial.up.sql -o schema/00000000_Initial.up.sql) || true


docker:
	docker build -t harder-rest-api -f Dockerfile .

docker-work-web-ui:
	docker build -t workwebui -f Dockerfile.work .

clean-development:
	docker-compose -f ./docker-compose.development.yaml rm --stop --force || true

compose-development: docker docker-work-web-ui
	docker-compose  -f ./docker-compose.development.yaml up

compose-development-lite:
	docker-compose  -f ./docker-compose.development.yaml up

helm-configure:
	which kubernetes-split-yaml || make helm-deps

helm-deps:
	git clone https://github.com/mogensen/kubernetes-split-yaml.git
	cd kubernetes-split-yaml.git && go build ./...
	cp kubernetes-split-yaml/kubernetes-split-yaml $(PWD)/bin
	rm -rfd kubernetes-split-yaml

helm-generate: helm-configure
	helm template rest-api ./ --dry-run --values=values.mayview.yaml | kubernetes-split-yaml -

staging-dry:
	helm template rest-api ./ --dry-run \
		--set api.jwt.loginJwtSecret=$$(vault kv get --field=jwt_secret pipelines/harderthanitneedstobe.com/staging/primary) \
		--set api.jwt.registrationJwtSecret=$$(vault kv get --field=register_jwt_secret pipelines/harderthanitneedstobe.com/staging/primary) \
		--set api.postgreSql.password=$$(vault kv get --field=pg_password pipelines/harderthanitneedstobe.com/staging/primary) \
		--values=values.staging.yaml | kubectl apply -n harder-staging --dry-run=server -f -


generate_schema:
	$(eval TARGET_FILE := $(shell echo "$(TARGET_DIRECTORY)/0_initial.up.sql"))
	$(info "Generating current schema into file $(TARGET_FILE)")
	go run github.com/monetrapp/rest-api/tools/schemagen > $(TARGET_FILE)
	yarn sql-formatter -l postgresql -u --lines-between-queries 2 $(TARGET_FILE) -o $(TARGET_FILE)

migrations:
	$(eval CURRENT_TMP := $(shell mktemp -d))
	$(eval BASE_TMP := $(shell mktemp -d))
	$(info "Generating schema migrations for the current schema in $(CURRENT_TMP)")
	make generate_schema TARGET_DIRECTORY=$(CURRENT_TMP)
	$(info "Cleaning up temp directories")
	rm -rf $(CURRENT_TMP)

ifdef CI
include Makefile.ci
endif

