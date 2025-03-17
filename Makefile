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
	goose -dir ./migrations/sql -s create ${label} sql

## migrations/up: apply all up database migrations
.PHONY: migrations/up
migrations/up:
	@echo "Running up migrations..."
	docker compose run --rm migrate -up

## migrations/drop: drop the entire database schema
.PHONY: migrations/drop
migrations/drop:
	@echo "Dropping the entire database schema..."
	docker compose run --rm migrate -drop

## migrations/reset: drop the database then apply all up migrations
.PHONY: migrations/reset
migrations/reset: migrations/drop migrations/up

## test: project-wide test suite
.PHONY: test
test:
	go test ./api/...
	go test ./lib/crypto/...
	go test ./lib/data/...
	go test ./lib/mailer/...
