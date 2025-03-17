package data_test

import (
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/micahco/mono/lib/data/postgres"
	"github.com/micahco/mono/migrations"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/goosemigrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create an ephemeral postgres db for testing
func newPostgresDB(t *testing.T) *postgres.PostgresDB {
	t.Helper()

	port := os.Getenv("TESTDB_PORT")
	require.NotEmpty(t, port, "missing env: PGTESTDB_PORT")

	dbconf := pgtestdb.Config{
		DriverName: "pgx",
		User:       "postgres",
		Password:   "password",
		Host:       "127.0.0.1",
		Port:       port,
		Options:    "sslmode=disable",
	}
	m := goosemigrator.New("sql", goosemigrator.WithFS(migrations.Files))
	c := pgtestdb.Custom(t, dbconf, m)
	assert.NotEqual(t, dbconf, *c)

	pg, err := postgres.NewPostgresDB(c.URL())
	assert.NoError(t, err)

	return pg
}

func TestPostgresUserRepository(t *testing.T) {
	t.Parallel()

	pg := newPostgresDB(t)
	defer pg.Close()

	runUserRepositoryTests(t, pg.DB)
}

func TestPostgresAuthenticationTokenRepository(t *testing.T) {
	t.Parallel()

	pg := newPostgresDB(t)
	defer pg.Close()

	runAuthenticationTokenRepositoryTests(t, pg.DB)
}

func TestPostgresVerificationTokenRepository(t *testing.T) {
	t.Parallel()

	pg := newPostgresDB(t)
	defer pg.Close()

	runVerificationTokenRepositoryTests(t, pg.DB)
}
