package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/client"
	"telegram-bot/internal/service"
	"telegram-bot/internal/store"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run() {
	cfg := config.MustLoad()

	mustFetchMigrations(*cfg)

	store, err := store.New(cfg.DSN)
	if err != nil {
		log.Fatal("failed to init store", err)
	}

	client := client.New()

	service, err := service.New(cfg.Token, cfg.ChannelID, client, store)
	if err != nil {
		log.Fatal("failed to start telegram bot", err)
	}

	if err := service.StartBroadcast(context.Background()); err != nil {
		log.Fatal("failed to start broadcast", err)
	}
}

func mustFetchMigrations(cfg config.Config) {
	migrations, err := migrate.New(fmt.Sprintf("file://%s", cfg.MigrationsPath), cfg.DSN)
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
