package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var dsn, migrationsPath string

	flag.StringVar(&dsn, "dsn", "", "path to storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.Parse()

	if dsn == "" {
		panic("storage path is required")
	}

	if migrationsPath == "" {
		panic("migrations path is required")
	}

	migrations, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to create migrations: %s", err))
	}

	if err := migrations.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return
		}
		panic(fmt.Sprintf("failed to run migrations: %s", err))
	}
}
