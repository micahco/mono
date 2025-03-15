module github.com/micahco/mono/shared/data

go 1.24.1

require (
	github.com/alexedwards/argon2id v1.0.0
	github.com/gofrs/uuid/v5 v5.3.1
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438
	github.com/jackc/pgx-gofrs-uuid v0.0.0-20230224015001-1d428863c2e2
	github.com/jackc/pgx/v5 v5.7.2
	github.com/micahco/mono/migrations v0.0.0
	github.com/peterldowns/pgtestdb v0.1.1
	github.com/peterldowns/pgtestdb/migrators/golangmigrator v0.1.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-migrate/migrate/v4 v4.16.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/micahco/mono/migrations => ../../migrations
