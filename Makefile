.PHONY: schema

test:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

special-tests:
	cd tests && make assert-clean-generated

schema:
	go run github.com/harderthanitneedstobe/rest-api/v0/cmd/schemagen > schema/00000000_Initial.up.sql

apply-schema-ci:
	go run github.com/harderthanitneedstobe/rest-api/v0/cmd/schemagen \
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
