## help: print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" |  sed -e "s/^/ /"

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	docker compose exec db psql ${DATABASE_URL}

## migrations/new label=$1: create a new database migration
.PHONY: migrations/new
migrations/new:
	@echo "Creating migration files for ${label}..."
	migrate create -seq -ext=.sql -dir=./migrations ${label}

## migrations/up: apply all up database migrations
.PHONY: migrations/up
migrations/up:
	@echo "Running up migrations..."
	docker-compose run --rm migrate -path ./migrations -database ${DATABASE_URL} up

## migrations/drop: drop the entire databse schema
.PHONY: migrations/drop
migrations/drop:
	@echo "Dropping the entire database schema..."
	docker-compose run --rm migrate -path ./migrations -database ${DATABASE_URL} drop

## test: project-wide test suite
.PHONY: test
test:
	go test ./api/...
	go test ./shared/crypto/...
	go test ./shared/data/...
	go test ./shared/mailer/...
