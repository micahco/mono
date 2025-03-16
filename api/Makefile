## help: print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" |  sed -e "s/^/ /"

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy
	go mod verify
	@echo "Formatting code..."
	go fmt ./...
	@echo "Vetting code..."
	go vet ./...
	go tool staticcheck ./...
	@echo "Running tests..."
	go test -race -vet=off ./...
	
## build: build the cmd/api application
current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long;)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

.PHONY: build
build:
	@echo "Building cmd/api..."
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api

## run: run the cmd/api application
# override env vars to use localhost
database_url = 'postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable'
smtp_host = 'localhost'

.PHONY: run
run:
	go run ./cmd/api -dev \
	-db-dsn=${database_url} \
	-smtp-host=${smtp_host}
