module github.com/micahco/mono/lib/data

go 1.24.1

require (
	github.com/gofrs/uuid/v5 v5.3.1
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438
	github.com/jackc/pgx-gofrs-uuid v0.0.0-20230224015001-1d428863c2e2
	github.com/jackc/pgx/v5 v5.7.2
	github.com/micahco/mono/migrations v0.0.0-20250317190502-ad2fa0c0b13c
	github.com/peterldowns/pgtestdb v0.1.1
	github.com/peterldowns/pgtestdb/migrators/goosemigrator v0.1.1
	github.com/stretchr/testify v1.10.0
)

replace github.com/micahco/mono/migrations v0.0.0-20250317190502-ad2fa0c0b13c => ../../migrations

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/pressly/goose/v3 v3.24.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
