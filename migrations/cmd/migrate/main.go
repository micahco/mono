package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/micahco/mono/migrations"
)

func main() {
	var dropFlag, upFlag bool

	flag.BoolVar(&dropFlag, "drop", false, "reset database")
	flag.BoolVar(&upFlag, "up", false, "update database")
	flag.Parse()

	// No flag provided, show help message
	if !upFlag && !dropFlag {
		flag.Usage()
		os.Exit(0)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("missing env: DATABASE_URL")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("pgx: %v", err)
	}
	defer db.Close()

	m, err := migrations.NewMigrator(db)
	if err != nil {
		log.Fatal(err)
	}

	if dropFlag {
		if err := m.Reset(); err != nil {
			log.Fatalf("drop: %v", err)
		}
	}

	if upFlag {
		if err := m.Up(); err != nil {
			log.Fatalf("up: %v", err)
		}
	}
}
