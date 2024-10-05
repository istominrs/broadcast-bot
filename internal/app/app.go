package app

import (
	"context"
	"log"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/client"
	"telegram-bot/internal/service"
	"telegram-bot/internal/store"
)

func Run() {
	cfg := config.MustLoad()

	store, err := store.New("postgres://username:password@localhost:5432/postgres?sslmode=disable")
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
