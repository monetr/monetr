.PHONY: schema

docs-dependencies:
	stat /bin/swag || (go get github.com/swaggo/swag/cmd/swag && go build -o /bin/swag github.com/swaggo/swag/cmd/swag)

docs: docs-dependencies
	/bin/swag init -d pkg/controller -g controller.go --parseDependency --parseDepth 5 --parseInternal

test:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

special-tests:
	cd tests && make assert-clean-generated

schema:
	go run github.com/harderthanitneedstobe/rest-api/v0/tools/schemagen > schema/00000000_Initial.up.sql
	yarn sql-formatter -l postgresql -u --lines-between-queries 2 -i 4 \
		schema/00000000_Initial.up.sql -o schema/00000000_Initial.up.sql

apply-schema-ci:
	go run github.com/harderthanitneedstobe/rest-api/v0/tools/schemagen \
		--address=$$POSTGRES_HOST \
		--port=5432 \
		--user=$$POSTGRES_USER \
		--db=$$POSTGRES_DB \
		--dry-run=false \
		--drop=true \
		--print=false

docker:
	docker build -t harder-rest-api -f Dockerfile .

docker-work-web-ui:
	docker build -t workwebui -f Dockerfile.work .

clean-development:
	docker-compose -f ./docker-compose.development.yaml rm --stop --force || true

compose-development: schema docker-work-web-ui
	docker-compose  -f ./docker-compose.development.yaml up
