package postgres

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/micahco/mono/migrations"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"github.com/stretchr/testify/assert"
)

func testDB(t *testing.T) *PostgresDB {
	t.Helper()
	dbconf := pgtestdb.Config{
		DriverName: "pgx",
		User:       "postgres",
		Password:   "password",
		Host:       "localhost",
		Port:       "5433", // non-default testing port
		Options:    "sslmode=disable",
	}
	m := golangmigrator.New(".", golangmigrator.WithFS(migrations.Files))
	c := pgtestdb.Custom(t, dbconf, m)
	assert.NotEqual(t, dbconf, *c)

	pg, err := NewPostgresDB(c.URL())
	assert.NoError(t, err)

	return pg
}
