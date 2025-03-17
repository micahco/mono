package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var Files embed.FS

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) (*Migrator, error) {
	goose.SetBaseFS(Files)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if _, err := goose.EnsureDBVersion(db); err != nil {
		return nil, err
	}

	return &Migrator{db}, nil
}

func (m *Migrator) Up() error {
	return goose.Up(m.db, "sql")
}

func (m *Migrator) Reset() error {
	return goose.Reset(m.db, "sql")
}
