## help: print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" |  sed -e "s/^/ /"

# confirmation dialog helper
.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

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
migrations/up: confirm
	@echo "Running up migrations..."
	docker-compose run --rm migrate -path ./migrations -database ${DATABASE_URL} up

## migrations/drop: drop the entire databse schema
.PHONY: migrations/drop
migrations/drop: confirm
	@echo "Dropping the entire database schema..."
	docker-compose run --rm migrate -path ./migrations -database ${DATABASE_URL} drop
