.PHONY: schema

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
		--drop=true

docker:
	GOOS=linux go build -o ./bin/rest-api github.com/harderthanitneedstobe/rest-api/v0/cmd/api
	docker build -t harder-rest-api -f Dockerfile .

clean-development:
	docker-compose -f ./docker-compose.development.yaml rm --stop --force || true

compose-development: schema
	docker-compose  -f ./docker-compose.development.yaml up
