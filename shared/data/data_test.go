package data_test

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/micahco/mono/shared/data/postgres"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"github.com/stretchr/testify/assert"
)

// Create an ephemeral postgres db for testing
func newPostgresDB(t *testing.T) *postgres.PostgresDB {
	t.Helper()
	dbconf := pgtestdb.Config{
		DriverName: "pgx",
		User:       "postgres",
		Password:   "password",
		Host:       "127.0.0.1",
		Port:       "5433", // non-default testing port
		Options:    "sslmode=disable",
	}
	m := golangmigrator.New("../../migrations")
	c := pgtestdb.Custom(t, dbconf, m)
	assert.NotEqual(t, dbconf, *c)

	pg, err := postgres.NewPostgresDB(c.URL())
	assert.NoError(t, err)

	return pg
}

// Test postgres implementation
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
