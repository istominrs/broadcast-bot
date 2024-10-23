package app

import (
	"context"
	"log/slog"
	"os"
	"telegram-bot/internal/bot/telegram"
	"telegram-bot/internal/config"
	"telegram-bot/internal/handlers/client"
	"telegram-bot/internal/repository/postgres"
	"telegram-bot/internal/service"
)

func Run() {
	cfg := config.MustLoad()

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)

	repository, err := postgres.New(cfg.DSN)
	if err != nil {
		panic(err)
	}

	c, err := client.New(
		client.WithLogger(log),
		client.WithClient(),
	)
	if err != nil {
		panic(err)
	}

	svc, err := service.New(
		service.WithLogger(log),
		service.WithClient(c),
		service.WithRepository(repository),
	)
	if err != nil {
		panic(err)
	}

	bot, err := telegram.New(
		telegram.WithToken(cfg.Token),
		telegram.WithLogger(log),
		telegram.WithService(svc),
		telegram.WithChatID(cfg.ChannelID),
	)
	if err != nil {
		panic(err)
	}

	if err := bot.Start(context.Background()); err != nil {
		panic(err)
	}
}
