package service

import (
	"context"
	"fmt"
	"telegram-bot/internal/entity"
	"telegram-bot/internal/handlers/client"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type store interface {
	Server(context.Context) ([]entity.Server, error)
}

type Service struct {
	bot       *tgbotapi.BotAPI
	channelID int64

	client *client.Client
	store  store
}

func New(token string,
	channelID int64,
	client *client.Client,
	store store,
) (*Service, error) {
	const op = "service.New"

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Service{
		bot:       bot,
		channelID: channelID,
		client:    client,
		store:     store,
	}, nil
}

// StartBroadcast starts telegram channel broadcast.
func (s *Service) StartBroadcast(ctx context.Context) error {
	const op = "service.StartBroadcast"

	servers, err := s.store.Server(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	accessURLs, err := s.client.CreateAccessURLs(servers)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()

	go func() {
		for _, u := range accessURLs {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				msg := tgbotapi.NewMessage(s.channelID, u.Url)
				if _, err := s.bot.Send(msg); err != nil {
					log.Printf("%s: %s", op, err)
				}
			}
		}
	}()

	time.Sleep(48 * time.Hour)
	if err := s.client.RemoveAccessURLs(accessURLs); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
